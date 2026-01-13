package registry

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/result"
	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/types"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

// mockDeployer is a test implementation of the Deployer interface.
type mockDeployer struct {
	name string
}

func (m *mockDeployer) Generate(ctx context.Context, recipeResult *recipe.RecipeResult,
	bundleDir string) (*result.Artifacts, error) {

	return result.New(m.name), nil
}

func newMockDeployer(name string) Factory {
	return func() Deployer {
		return &mockDeployer{name: name}
	}
}

// saveAndResetGlobalRegistry saves the current global registry and creates a fresh one.
// Returns a cleanup function that restores the original registry.
func saveAndResetGlobalRegistry() func() {
	globalMu.Lock()
	origRegistry := globalRegistry
	globalRegistry = &Registry{
		deployers: make(map[types.DeployerType]Factory),
	}
	globalMu.Unlock()

	return func() {
		globalMu.Lock()
		globalRegistry = origRegistry
		globalMu.Unlock()
	}
}

func TestRegister(t *testing.T) {
	cleanup := saveAndResetGlobalRegistry()
	defer cleanup()

	tests := []struct {
		name         string
		deployerType types.DeployerType
		wantErr      bool
	}{
		{"register new type", types.DeployerType("test1"), false},
		{"register another type", types.DeployerType("test2"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Register(tt.deployerType, newMockDeployer(string(tt.deployerType)))
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegister_Duplicate(t *testing.T) {
	cleanup := saveAndResetGlobalRegistry()
	defer cleanup()

	deployerType := types.DeployerType("test")
	factory := newMockDeployer("test")

	// First registration should succeed
	if err := Register(deployerType, factory); err != nil {
		t.Fatalf("First Register() failed: %v", err)
	}

	// Second registration should fail
	err := Register(deployerType, factory)
	if err == nil {
		t.Error("Register() with duplicate type did not return error")
	}

	expectedMsg := "deployer type test already registered"
	if err.Error() != expectedMsg {
		t.Errorf("Register() error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestMustRegister_Panic(t *testing.T) {
	cleanup := saveAndResetGlobalRegistry()
	defer cleanup()

	deployerType := types.DeployerType("test")
	factory := newMockDeployer("test")

	// First registration should succeed
	MustRegister(deployerType, factory)

	// Second registration should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustRegister() with duplicate type did not panic")
		}
	}()

	MustRegister(deployerType, factory)
}

func TestRegistry_Get(t *testing.T) {
	r := &Registry{
		deployers: make(map[types.DeployerType]Factory),
	}

	testType := types.DeployerType("test")
	r.deployers[testType] = newMockDeployer("test")

	tests := []struct {
		name         string
		deployerType types.DeployerType
		wantFound    bool
	}{
		{"existing type", testType, true},
		{"non-existing type", types.DeployerType("nonexistent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployer, found := r.Get(tt.deployerType)

			if found != tt.wantFound {
				t.Errorf("Get() found = %v, want %v", found, tt.wantFound)
			}

			if tt.wantFound && deployer == nil {
				t.Error("Get() returned nil deployer for existing type")
			}

			if !tt.wantFound && deployer != nil {
				t.Error("Get() returned non-nil deployer for non-existing type")
			}
		})
	}
}

func TestRegistry_GetAll(t *testing.T) {
	r := &Registry{
		deployers: make(map[types.DeployerType]Factory),
	}

	// Register multiple deployers
	dtypes := []types.DeployerType{
		types.DeployerType("test1"),
		types.DeployerType("test2"),
		types.DeployerType("test3"),
	}

	for _, dt := range dtypes {
		r.deployers[dt] = newMockDeployer(string(dt))
	}

	// Get all deployers
	all := r.GetAll()

	if len(all) != len(dtypes) {
		t.Errorf("GetAll() returned %d deployers, want %d", len(all), len(dtypes))
	}

	// Verify all types are present
	for _, dt := range dtypes {
		if _, exists := all[dt]; !exists {
			t.Errorf("GetAll() missing deployer type: %s", dt)
		}
	}
}

func TestRegistry_Types(t *testing.T) {
	r := &Registry{
		deployers: make(map[types.DeployerType]Factory),
	}

	// Register multiple deployers
	expectedTypes := []types.DeployerType{
		types.DeployerType("test1"),
		types.DeployerType("test2"),
		types.DeployerType("test3"),
	}

	for _, dt := range expectedTypes {
		r.deployers[dt] = newMockDeployer(string(dt))
	}

	// Get all types
	gotTypes := r.Types()

	if len(gotTypes) != len(expectedTypes) {
		t.Errorf("Types() returned %d types, want %d", len(gotTypes), len(expectedTypes))
	}

	// Verify all expected types are present
	typeMap := make(map[types.DeployerType]bool)
	for _, dt := range gotTypes {
		typeMap[dt] = true
	}

	for _, dt := range expectedTypes {
		if !typeMap[dt] {
			t.Errorf("Types() missing expected type: %s", dt)
		}
	}
}

func TestNewFromGlobal(t *testing.T) {
	cleanup := saveAndResetGlobalRegistry()
	defer cleanup()

	testTypes := []types.DeployerType{
		types.DeployerType("test1"),
		types.DeployerType("test2"),
	}

	for _, dt := range testTypes {
		_ = Register(dt, newMockDeployer(string(dt)))
	}

	// Create registry from global
	r := NewFromGlobal()

	if r == nil {
		t.Fatal("NewFromGlobal() returned nil")
	}

	// Verify all types were copied
	for _, dt := range testTypes {
		if _, exists := r.deployers[dt]; !exists {
			t.Errorf("NewFromGlobal() missing deployer type: %s", dt)
		}
	}

	// Verify it's a copy, not the same instance
	if r == globalRegistry {
		t.Error("NewFromGlobal() returned same instance as global registry")
	}
}

func TestRegistry_Concurrency(t *testing.T) {
	cleanup := saveAndResetGlobalRegistry()
	defer cleanup()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // Register + Get operations

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				dt := types.DeployerType(fmt.Sprintf("test-%d-%d", id, j))
				_ = Register(dt, newMockDeployer(string(dt)))
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			r := NewFromGlobal()
			for j := 0; j < numOperations; j++ {
				_ = r.GetAll()
				_ = r.Types()
			}
		}()
	}

	wg.Wait()

	// Verify registry is still functional
	r := NewFromGlobal()
	all := r.GetAll()

	if len(all) == 0 {
		t.Error("Registry is empty after concurrent operations")
	}
}

func TestRegistry_GetCreatesNewInstance(t *testing.T) {
	r := &Registry{
		deployers: make(map[types.DeployerType]Factory),
	}

	testType := types.DeployerType("test")
	r.deployers[testType] = newMockDeployer("test")

	// Get two instances
	d1, ok1 := r.Get(testType)
	d2, ok2 := r.Get(testType)

	if !ok1 || !ok2 {
		t.Fatal("Get() failed to retrieve deployer")
	}

	// Verify they are different instances
	if d1 == d2 {
		t.Error("Get() returned same instance, expected new instances from factory")
	}
}
