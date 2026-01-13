// Package script provides script-based deployment artifact generation.
// This is the default deployer that generates shell scripts and basic README
// for manual deployment of CNS components.
package script

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

//go:embed templates/README.md.tmpl
var readmeTemplate string

func init() {
	// Self-register with the global deployer registry
	registry.MustRegister(types.DeployerTypeScript, NewDeployer)
}

// Deployer generates script-based deployment artifacts.
type Deployer struct{}

// NewDeployer creates a new script deployer instance.
func NewDeployer() registry.Deployer {
	return &Deployer{}
}

// Generate creates deployment artifacts including a basic README with manual deployment instructions.
func (d *Deployer) Generate(ctx context.Context, recipeResult *recipe.RecipeResult,
	bundleDir string) (*result.Artifacts, error) {

	startTime := time.Now()
	artifacts := result.New(string(types.DeployerTypeScript))

	// Order components by deployment order
	orderedComponents := orderComponentsByDeployment(recipeResult.ComponentRefs, recipeResult.DeploymentOrder)

	// Prepare template data with ordered components
	data := internal.ReadmeData{
		Timestamp:     time.Now().Format(time.RFC3339),
		RecipeVersion: recipeResult.Metadata.Version,
		Components:    make([]internal.ComponentInfo, 0, len(orderedComponents)),
	}

	for _, componentRef := range orderedComponents {
		data.Components = append(data.Components, internal.ComponentInfo{
			Name:    componentRef.Name,
			Version: componentRef.Version,
		})
	}

	// Render README template
	readme, err := internal.RenderTemplate(readmeTemplate, data)
	if err != nil {
		artifacts.Success = false
		artifacts.Error = fmt.Sprintf("failed to render README template: %v", err)
		return artifacts, err
	}

	// Write README.md
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
