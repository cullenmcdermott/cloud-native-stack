package bundler

import (
	"context"

	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

// Bundler defines the interface for creating application bundles.
// Implementations generate deployment artifacts from recipes.
type Bundler interface {
	// Make generates the bundle in the specified directory.
	// Returns a BundleResult containing information about generated files.
	Make(ctx context.Context, recipe *recipe.Recipe, dir string) (*BundleResult, error)
}

// ConfigurableBundler extends Bundler with configuration support.
type ConfigurableBundler interface {
	Bundler
	// Configure applies configuration to the bundler.
	Configure(config *BundlerConfig) error
}

// BundleType identifies different types of bundles.
type BundleType string

const (
	// BundleTypeGpuOperator generates GPU Operator bundles.
	BundleTypeGpuOperator BundleType = "gpu-operator"

	// BundleTypeNetworkOperator generates Network Operator bundles.
	BundleTypeNetworkOperator BundleType = "network-operator"
)
