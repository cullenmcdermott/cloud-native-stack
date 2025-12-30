package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/NVIDIA/cloud-native-stack/pkg/bundler/result"
	"github.com/NVIDIA/cloud-native-stack/pkg/bundler/types"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

// Bundler defines the interface for creating application bundles.
// Implementations generate deployment artifacts from recipes.
type Bundler interface {
	Make(ctx context.Context, recipe *recipe.Recipe, dir string) (*result.Result, error)
}

// ValidatableBundler is an optional interface that bundlers can implement
// to validate recipes before processing. This provides type-safe validation
// without reflection.
type ValidatableBundler interface {
	Bundler
	Validate(ctx context.Context, recipe *recipe.Recipe) error
}

// Registry manages registered bundlers with thread-safe operations.
type Registry struct {
	bundlers map[types.BundleType]Bundler
	mu       sync.RWMutex
}

// NewRegistry creates a new empty Registry instance.
// Bundlers should be registered explicitly using Register().
func NewRegistry() *Registry {
	return &Registry{
		bundlers: make(map[types.BundleType]Bundler),
	}
}

// Register registers a bundler in this registry.
func (r *Registry) Register(bundleType types.BundleType, b Bundler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bundlers[bundleType] = b
}

// Get retrieves a bundler by type from this registry.
func (r *Registry) Get(bundleType types.BundleType) (Bundler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.bundlers[bundleType]
	return b, ok
}

// GetAll returns all registered bundlers.
func (r *Registry) GetAll() map[types.BundleType]Bundler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bundlers := make(map[types.BundleType]Bundler, len(r.bundlers))
	for k, v := range r.bundlers {
		bundlers[k] = v
	}
	return bundlers
}

// List returns all registered bundler types.
func (r *Registry) List() []types.BundleType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]types.BundleType, 0, len(r.bundlers))
	for k := range r.bundlers {
		types = append(types, k)
	}
	return types
}

// Unregister removes a bundler from this registry.
func (r *Registry) Unregister(bundleType types.BundleType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.bundlers[bundleType]; !ok {
		return fmt.Errorf("bundler type %s not registered", bundleType)
	}

	delete(r.bundlers, bundleType)
	return nil
}

// Count returns the number of registered bundlers.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.bundlers)
}

// IsEmpty returns true if no bundlers are registered.
// This is useful for checking if a registry has been populated.
func (r *Registry) IsEmpty() bool {
	return r.Count() == 0
}
