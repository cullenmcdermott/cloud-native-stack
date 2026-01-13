package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/result"
	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/types"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

// Deployer defines the interface for deployment artifact generators.
type Deployer interface {
	// Generate creates deployment artifacts for the given recipe result.
	// The bundleDir is the root directory where component bundles have been generated.
	Generate(ctx context.Context, recipeResult *recipe.RecipeResult,
		bundleDir string) (*result.Artifacts, error)
}

// Factory is a function that creates a new deployer instance.
type Factory func() Deployer

var (
	globalRegistry = &Registry{
		deployers: make(map[types.DeployerType]Factory),
	}
	globalMu sync.RWMutex
)

// Register registers a deployer factory with the global registry.
// Returns an error if the deployer type is already registered.
func Register(deployerType types.DeployerType, factory Factory) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	if _, exists := globalRegistry.deployers[deployerType]; exists {
		return fmt.Errorf("deployer type %s already registered", deployerType)
	}

	globalRegistry.deployers[deployerType] = factory
	return nil
}

// MustRegister registers a deployer factory with the global registry.
// Panics if the deployer type is already registered.
func MustRegister(deployerType types.DeployerType, factory Factory) {
	if err := Register(deployerType, factory); err != nil {
		panic(err)
	}
}

// Registry manages deployer instances.
type Registry struct {
	deployers map[types.DeployerType]Factory
	mu        sync.RWMutex
}

// NewFromGlobal creates a new Registry instance using the global registry's factories.
func NewFromGlobal() *Registry {
	globalMu.RLock()
	defer globalMu.RUnlock()

	r := &Registry{
		deployers: make(map[types.DeployerType]Factory),
	}

	// Copy factories from global registry
	for deployerType, factory := range globalRegistry.deployers {
		r.deployers[deployerType] = factory
	}

	return r
}

// Get retrieves a deployer instance by type.
func (r *Registry) Get(deployerType types.DeployerType) (Deployer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.deployers[deployerType]
	if !ok {
		return nil, false
	}

	return factory(), true
}

// GetAll returns all registered deployer instances.
func (r *Registry) GetAll() map[types.DeployerType]Deployer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[types.DeployerType]Deployer)
	for deployerType, factory := range r.deployers {
		result[deployerType] = factory()
	}

	return result
}

// Types returns all registered deployer types.
func (r *Registry) Types() []types.DeployerType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]types.DeployerType, 0, len(r.deployers))
	for deployerType := range r.deployers {
		result = append(result, deployerType)
	}

	return result
}
