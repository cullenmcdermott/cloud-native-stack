// Package flux provides Flux-based deployment artifact generation.
// Generates Flux HelmRelease and Kustomization resources for GitOps deployment.
package flux

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/internal"
	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/registry"
	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/result"
	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/types"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

//go:embed templates/kustomization.yaml.tmpl
var kustomizationTemplate string

//go:embed templates/helmrelease.yaml.tmpl
var helmReleaseTemplate string

//go:embed templates/README.md.tmpl
var readmeTemplate string

func init() {
	// Self-register with the global deployer registry
	registry.MustRegister(types.DeployerTypeFlux, NewDeployer)
}

// Deployer generates Flux deployment artifacts.
type Deployer struct{}

// NewDeployer creates a new Flux deployer instance.
func NewDeployer() registry.Deployer {
	return &Deployer{}
}

// HelmReleaseData contains data for component HelmRelease template.
type HelmReleaseData struct {
	Namespace          string
	Name               string
	Source             string
	Version            string
	DependsOnName      string // Name of the HelmRelease this depends on (empty if first)
	DependsOnNamespace string // Namespace of the dependency
}

// Generate creates Flux Kustomization resources and deployment README.
func (d *Deployer) Generate(ctx context.Context, recipeResult *recipe.RecipeResult,
	bundleDir string) (*result.Artifacts, error) {

	startTime := time.Now()
	artifacts := result.New(string(types.DeployerTypeFlux))

	// Create flux directory
	fluxDir := filepath.Join(bundleDir, "flux")
	if err := os.MkdirAll(fluxDir, 0755); err != nil {
		artifacts.Success = false
		artifacts.Error = fmt.Sprintf("failed to create flux directory: %v", err)
		return artifacts, err
	}

	// Generate parent kustomization.yaml
	parentKustomization, err := internal.RenderTemplate(kustomizationTemplate, nil)
	if err != nil {
		artifacts.Success = false
		artifacts.Error = fmt.Sprintf("failed to render parent kustomization template: %v", err)
		return artifacts, err
	}

	parentKustomizationPath := filepath.Join(fluxDir, "kustomization.yaml")
	if writeErr := os.WriteFile(parentKustomizationPath, []byte(parentKustomization), 0600); writeErr != nil {
		artifacts.Success = false
		artifacts.Error = fmt.Sprintf("failed to write parent kustomization: %v", writeErr)
		return artifacts, writeErr
	}

	artifacts.Files = append(artifacts.Files, parentKustomizationPath)

	// Order components by deployment order
	orderedComponents := orderComponentsByDeployment(recipeResult.ComponentRefs, recipeResult.DeploymentOrder)

	// Generate HelmRelease for each component with dependency chain
	var previousName string
	var previousNamespace string
	for _, componentRef := range orderedComponents {
		componentData := HelmReleaseData{
			Namespace: internal.GetNamespaceForComponent(componentRef.Name),
			Name:      componentRef.Name,
			Source:    componentRef.Source,
			Version:   componentRef.Version,
		}

		// Set dependency on previous component in deployment order
		if previousName != "" {
			componentData.DependsOnName = previousName
			componentData.DependsOnNamespace = previousNamespace
		}

		helmRelease, renderErr := internal.RenderTemplate(helmReleaseTemplate, componentData)
		if renderErr != nil {
			artifacts.Success = false
			artifacts.Error = fmt.Sprintf("failed to render HelmRelease template: %v", renderErr)
			return artifacts, renderErr
		}

		componentDir := filepath.Join(bundleDir, componentRef.Name)
		helmReleasePath := filepath.Join(componentDir, "helmrelease.yaml")

		if writeErr := os.WriteFile(helmReleasePath, []byte(helmRelease), 0600); writeErr != nil {
			artifacts.Success = false
			artifacts.Error = fmt.Sprintf("failed to write HelmRelease: %v", writeErr)
			return artifacts, writeErr
		}

		artifacts.Files = append(artifacts.Files, helmReleasePath)
		previousName = componentRef.Name
		previousNamespace = componentData.Namespace
	}

	// Generate README with ordered components
	readmeData := internal.ReadmeData{
		Timestamp:  time.Now().Format(time.RFC3339),
		Components: make([]internal.ComponentInfo, 0, len(orderedComponents)),
	}

	for _, componentRef := range orderedComponents {
		readmeData.Components = append(readmeData.Components, internal.ComponentInfo{
			Name:    componentRef.Name,
			Version: componentRef.Version,
		})
	}

	readme, err := internal.RenderTemplate(readmeTemplate, readmeData)
	if err != nil {
		artifacts.Success = false
		artifacts.Error = fmt.Sprintf("failed to render README template: %v", err)
		return artifacts, err
	}

	readmePath := filepath.Join(bundleDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0600); err != nil {
		artifacts.Success = false
		artifacts.Error = fmt.Sprintf("failed to write README: %v", err)
		return artifacts, err
	}

	artifacts.Files = append(artifacts.Files, readmePath)
	artifacts.ReadmeContent = readme
	artifacts.Duration = time.Since(startTime)

	return artifacts, nil
}

// orderComponentsByDeployment returns components sorted according to deployment order.
// If deployment order is empty or a component is not in the order, it will appear after
// the ordered components in its original position.
func orderComponentsByDeployment(components []recipe.ComponentRef, order []string) []recipe.ComponentRef {
	if len(order) == 0 {
		return components
	}

	// Build position map
	posMap := make(map[string]int, len(order))
	for i, name := range order {
		posMap[name] = i
	}

	// Create result slice
	result := make([]recipe.ComponentRef, len(components))
	copy(result, components)

	// Sort by deployment order
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			posI, okI := posMap[result[i].Name]
			posJ, okJ := posMap[result[j].Name]

			// Components in order come before components not in order
			if !okI && okJ {
				result[i], result[j] = result[j], result[i]
			} else if okI && okJ && posJ < posI {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
