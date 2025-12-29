package networkoperator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/NVIDIA/cloud-native-stack/pkg/bundler/bundle"
	"github.com/NVIDIA/cloud-native-stack/pkg/bundler/config"
	"github.com/NVIDIA/cloud-native-stack/pkg/measurement"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

const (
	manifestsDir  = "manifests"
	scriptsDir    = "scripts"
	filePerms     = 0644
	execPerms     = 0755
	configSubtype = "config"
)

// Bundler generates Network Operator deployment bundles.
type Bundler struct {
	// Config for customization
	cfg *config.Config
}

// NewBundler creates a new Network Operator bundler.
func NewBundler(cfg *config.Config) *Bundler {
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &Bundler{
		cfg: cfg,
	}
}

// Make generates a Network Operator bundle from a recipe.
func (b *Bundler) Make(ctx context.Context, r *recipe.Recipe, outputDir string) (*bundle.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	// Validate recipe has required measurements
	if err := b.validateRecipe(r); err != nil {
		return nil, fmt.Errorf("invalid recipe: %w", err)
	}

	// Create result tracker
	result := bundle.NewResult(bundle.BundleTypeNetworkOperator)

	// Create output directory structure
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create subdirectories
	for _, dir := range []string{manifestsDir, scriptsDir} {
		path := filepath.Join(outputDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s directory: %w", dir, err)
		}
	}

	// Build configuration map from recipe and bundler config
	configMap := b.buildConfigMap(r)

	// Generate all bundle components
	if err := b.generateHelmValues(ctx, r, configMap, outputDir, result); err != nil {
		return nil, fmt.Errorf("failed to generate helm values: %w", err)
	}

	if err := b.generateNicClusterPolicy(ctx, r, configMap, outputDir, result); err != nil {
		return nil, fmt.Errorf("failed to generate NicClusterPolicy: %w", err)
	}

	if err := b.generateScripts(ctx, r, configMap, outputDir, result); err != nil {
		return nil, fmt.Errorf("failed to generate scripts: %w", err)
	}

	if err := b.generateReadme(ctx, r, configMap, outputDir, result); err != nil {
		return nil, fmt.Errorf("failed to generate README: %w", err)
	}

	// Generate checksums file last
	if err := b.generateChecksums(ctx, outputDir, result); err != nil {
		return nil, fmt.Errorf("failed to generate checksums: %w", err)
	}

	// Mark as successful
	result.MarkSuccess()

	return result, nil
}

// validateRecipe checks if recipe has required measurements.
func (b *Bundler) validateRecipe(r *recipe.Recipe) error {
	if r == nil {
		return fmt.Errorf("recipe is nil")
	}

	// Check for required K8s measurements
	hasK8s := false
	for _, m := range r.Measurements {
		if m.Type == measurement.TypeK8s {
			hasK8s = true
			break
		}
	}

	if !hasK8s {
		return fmt.Errorf("recipe missing required Kubernetes measurements")
	}

	return nil
}

// buildConfigMap extracts configuration from recipe and bundler config.
func (b *Bundler) buildConfigMap(r *recipe.Recipe) map[string]string {
	configMap := make(map[string]string)

	// Add bundler config values
	if b.cfg.HelmRepository != "" {
		configMap["helm_repository"] = b.cfg.HelmRepository
	}
	if b.cfg.HelmChartVersion != "" {
		configMap["helm_chart_version"] = b.cfg.HelmChartVersion
	}
	if b.cfg.Namespace != "" {
		configMap["namespace"] = b.cfg.Namespace
	}

	// Add custom labels and annotations
	for k, v := range b.cfg.CustomLabels {
		configMap["label_"+k] = v
	}
	for k, v := range b.cfg.CustomAnnotations {
		configMap["annotation_"+k] = v
	}

	// Extract values from recipe measurements
	for _, m := range r.Measurements {
		switch m.Type {
		case measurement.TypeK8s:
			for _, st := range m.Subtypes {
				if st.Name == "image" {
					// Extract Network Operator version
					if val, ok := st.Data["network-operator"]; ok {
						if s, ok := val.Any().(string); ok {
							configMap["network_operator_version"] = s
						}
					}
					// Extract OFED driver version
					if val, ok := st.Data["ofed-driver"]; ok {
						if s, ok := val.Any().(string); ok {
							configMap["ofed_version"] = s
						}
					}
				}

				if st.Name == configSubtype {
					// Extract RDMA setting
					if val, ok := st.Data["rdma"]; ok {
						if b, ok := val.Any().(bool); ok {
							configMap["enable_rdma"] = fmt.Sprintf("%t", b)
						}
					}
					// Extract SR-IOV setting
					if val, ok := st.Data["sr-iov"]; ok {
						if b, ok := val.Any().(bool); ok {
							configMap["enable_sriov"] = fmt.Sprintf("%t", b)
						}
					}
				}
			}
		case measurement.TypeSystemD, measurement.TypeOS, measurement.TypeGPU:
			// Not used for Network Operator configuration
			continue
		}
	}

	return configMap
}

// generateHelmValues creates the Helm values.yaml file.
func (b *Bundler) generateHelmValues(ctx context.Context, r *recipe.Recipe, configMap map[string]string, outputDir string, result *bundle.Result) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	values := GenerateHelmValues(r, configMap)
	if err := values.Validate(); err != nil {
		return fmt.Errorf("invalid helm values: %w", err)
	}

	content, err := b.renderTemplate("values.yaml", values.ToMap())
	if err != nil {
		return fmt.Errorf("failed to render values template: %w", err)
	}

	path := filepath.Join(outputDir, "values.yaml")
	return b.writeFile(path, content, filePerms, result)
}

// generateNicClusterPolicy creates the NicClusterPolicy manifest.
func (b *Bundler) generateNicClusterPolicy(ctx context.Context, r *recipe.Recipe, configMap map[string]string, outputDir string, result *bundle.Result) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	data := GenerateManifestData(r, configMap)

	content, err := b.renderTemplate("nicclusterpolicy", data.ToMap())
	if err != nil {
		return fmt.Errorf("failed to render NicClusterPolicy template: %w", err)
	}

	path := filepath.Join(outputDir, manifestsDir, "nicclusterpolicy.yaml")
	return b.writeFile(path, content, filePerms, result)
}

// generateScripts creates installation and uninstallation scripts.
func (b *Bundler) generateScripts(ctx context.Context, r *recipe.Recipe, configMap map[string]string, outputDir string, result *bundle.Result) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	scriptData := GenerateScriptData(r, configMap)
	data := scriptData.ToMap()

	// Generate install script
	installContent, err := b.renderTemplate("install.sh", data)
	if err != nil {
		return fmt.Errorf("failed to render install script template: %w", err)
	}

	installPath := filepath.Join(outputDir, scriptsDir, "install.sh")
	if err = b.writeFile(installPath, installContent, execPerms, result); err != nil {
		return err
	}

	// Generate uninstall script
	uninstallContent, err := b.renderTemplate("uninstall.sh", data)
	if err != nil {
		return fmt.Errorf("failed to render uninstall script template: %w", err)
	}

	uninstallPath := filepath.Join(outputDir, scriptsDir, "uninstall.sh")
	return b.writeFile(uninstallPath, uninstallContent, execPerms, result)
}

// generateReadme creates the README.md file.
func (b *Bundler) generateReadme(ctx context.Context, r *recipe.Recipe, configMap map[string]string, outputDir string, result *bundle.Result) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	scriptData := GenerateScriptData(r, configMap)
	helmValues := GenerateHelmValues(r, configMap)

	data := map[string]interface{}{
		"Script": scriptData.ToMap(),
		"Helm":   helmValues.ToMap(),
	}

	content, err := b.renderTemplate("README.md", data)
	if err != nil {
		return fmt.Errorf("failed to render README template: %w", err)
	}

	path := filepath.Join(outputDir, "README.md")
	return b.writeFile(path, content, filePerms, result)
}

// generateChecksums creates a checksums.txt file with SHA256 hashes.
func (b *Bundler) generateChecksums(ctx context.Context, outputDir string, result *bundle.Result) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var sb strings.Builder
	for _, file := range result.Files {
		if filepath.Base(file) == "checksums.txt" {
			continue // Don't include checksums file in checksums
		}

		// Read file content
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file %s for checksum: %w", file, err)
		}

		// Calculate checksum
		checksum := bundle.ComputeChecksum(content)

		// Get relative path
		relPath, err := filepath.Rel(outputDir, file)
		if err != nil {
			relPath = filepath.Base(file)
		}
		sb.WriteString(fmt.Sprintf("%s  %s\n", checksum, relPath))
	}

	path := filepath.Join(outputDir, "checksums.txt")
	return b.writeFile(path, sb.String(), filePerms, result)
}

// renderTemplate renders a template with the given data.
func (b *Bundler) renderTemplate(name string, data map[string]interface{}) (string, error) {
	tmplContent, ok := GetTemplate(name)
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	tmpl, err := template.New(name).Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return sb.String(), nil
}

// writeFile writes content to a file and updates the result.
func (b *Bundler) writeFile(path, content string, perms os.FileMode, result *bundle.Result) error {
	// Write file
	bytes := []byte(content)
	if err := os.WriteFile(path, bytes, perms); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	// Update result with file path and size
	result.AddFile(path, int64(len(bytes)))

	return nil
}
