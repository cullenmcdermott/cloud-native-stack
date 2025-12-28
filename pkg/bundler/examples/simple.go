package examples

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/NVIDIA/cloud-native-stack/pkg/bundler"
	"github.com/NVIDIA/cloud-native-stack/pkg/errors"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

const (
	// SimpleBundlerType is the type identifier for the simple example bundler.
	SimpleBundlerType bundler.BundleType = "simple"
)

// SimpleBundler demonstrates a minimal bundler implementation.
// It creates a simple text file with recipe information.
type SimpleBundler struct {
	config *bundler.BundlerConfig
}

// init registers the simple bundler with the global registry.
func init() {
	bundler.Register(SimpleBundlerType, NewSimpleBundler())
}

// NewSimpleBundler creates a new simple bundler instance.
func NewSimpleBundler() bundler.Bundler {
	return &SimpleBundler{
		config: bundler.DefaultBundlerConfig(),
	}
}

// Configure implements the ConfigurableBundler interface.
func (b *SimpleBundler) Configure(config *bundler.BundlerConfig) error {
	if config == nil {
		return errors.New(errors.ErrCodeInvalidRequest, "config cannot be nil")
	}

	b.config.Merge(config)
	return nil
}

// Validate implements the Validator interface.
func (b *SimpleBundler) Validate(_ context.Context, r *recipe.Recipe) error {
	// Simple validation: just ensure recipe has at least one measurement
	if r == nil || len(r.Measurements) == 0 {
		return errors.New(errors.ErrCodeInvalidRequest, "recipe must have at least one measurement")
	}

	slog.Debug("simple bundler validated recipe",
		"measurements", len(r.Measurements))
	return nil
}

// Make implements the Bundler interface.
// It creates a simple text file with recipe summary information.
func (b *SimpleBundler) Make(ctx context.Context, r *recipe.Recipe, outputDir string) (*bundler.BundleResult, error) {
	slog.Info("generating simple bundle",
		"dir", outputDir,
		"measurements", len(r.Measurements))

	result := bundler.NewBundleResult(SimpleBundlerType)

	// Create bundle directory
	bundleDir := filepath.Join(outputDir, "simple")
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternal, "failed to create bundle directory", err)
	}

	// Generate summary file
	if err := b.generateSummary(r, bundleDir, result); err != nil {
		return nil, err
	}

	// Generate metadata file
	if err := b.generateMetadata(r, bundleDir, result); err != nil {
		return nil, err
	}

	result.MarkSuccess()
	slog.Info("simple bundle generated successfully",
		"files", len(result.Files),
		"size", result.Size)

	return result, nil
}

// generateSummary creates a summary.txt file with recipe overview.
func (b *SimpleBundler) generateSummary(r *recipe.Recipe, dir string, result *bundler.BundleResult) error {
	content := "Simple Bundle Summary\n"
	content += "=====================\n\n"
	content += fmt.Sprintf("Generated from recipe with %d measurements:\n\n", len(r.Measurements))

	for _, m := range r.Measurements {
		content += fmt.Sprintf("- Type: %s\n", m.Type)
		content += fmt.Sprintf("  Subtypes: %d\n", len(m.Subtypes))
	}

	path := filepath.Join(dir, "summary.txt")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return errors.Wrap(errors.ErrCodeInternal, "failed to write summary file", err)
	}

	result.AddFile(path, int64(len(content)))
	slog.Debug("generated summary file", "path", path, "size", len(content))
	return nil
}

// generateMetadata creates a metadata.txt file with configuration info.
func (b *SimpleBundler) generateMetadata(_ *recipe.Recipe, dir string, result *bundler.BundleResult) error {
	content := "Bundle Metadata\n"
	content += "===============\n\n"
	content += fmt.Sprintf("Namespace: %s\n", b.config.Namespace)
	content += fmt.Sprintf("Output Format: %s\n", b.config.OutputFormat)
	content += fmt.Sprintf("Include Scripts: %t\n", b.config.IncludeScripts)

	if len(b.config.CustomLabels) > 0 {
		content += "\nCustom Labels:\n"
		for k, v := range b.config.CustomLabels {
			content += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}

	path := filepath.Join(dir, "metadata.txt")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return errors.Wrap(errors.ErrCodeInternal, "failed to write metadata file", err)
	}

	result.AddFile(path, int64(len(content)))
	slog.Debug("generated metadata file", "path", path, "size", len(content))
	return nil
}
