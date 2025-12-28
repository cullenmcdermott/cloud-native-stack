package bundler

import (
	"fmt"
	"sync"
)

// BundlerRegistry manages registered bundlers with thread-safe operations.
// It allows dynamic registration and retrieval of bundlers by type.
type BundlerRegistry struct {
	mu       sync.RWMutex
	bundlers map[BundleType]Bundler
}

var (
	defaultRegistry = NewRegistry()
)

// NewRegistry creates a new BundlerRegistry instance.
func NewRegistry() *BundlerRegistry {
	return &BundlerRegistry{
		bundlers: make(map[BundleType]Bundler),
	}
}

// Register registers a bundler with the default registry.
// This is typically called in package init() functions.
func Register(bundleType BundleType, b Bundler) {
	defaultRegistry.Register(bundleType, b)
}

// Register registers a bundler in this registry.
// If a bundler with the same type already exists, it will be replaced.
func (r *BundlerRegistry) Register(bundleType BundleType, b Bundler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bundlers[bundleType] = b
}

// Get retrieves a bundler by type from this registry.
// Returns the bundler and true if found, nil and false otherwise.
func (r *BundlerRegistry) Get(bundleType BundleType) (Bundler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.bundlers[bundleType]
	return b, ok
}

// GetAll returns all registered bundlers.
// Returns a copy of the bundlers map to prevent external modification.
func (r *BundlerRegistry) GetAll() map[BundleType]Bundler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bundlers := make(map[BundleType]Bundler, len(r.bundlers))
	for k, v := range r.bundlers {
		bundlers[k] = v
	}
	return bundlers
}

// List returns all registered bundler types.
func (r *BundlerRegistry) List() []BundleType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]BundleType, 0, len(r.bundlers))
	for k := range r.bundlers {
		types = append(types, k)
	}
	return types
}

// Unregister removes a bundler from this registry.
// Returns an error if the bundler type doesn't exist.
func (r *BundlerRegistry) Unregister(bundleType BundleType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.bundlers[bundleType]; !ok {
		return fmt.Errorf("bundler type %s not registered", bundleType)
	}

	delete(r.bundlers, bundleType)
	return nil
}

// GetFromDefault retrieves a bundler from the default registry.
func GetFromDefault(bundleType BundleType) (Bundler, bool) {
	return defaultRegistry.Get(bundleType)
}

// ListFromDefault returns all bundler types from the default registry.
func ListFromDefault() []BundleType {
	return defaultRegistry.List()
}
