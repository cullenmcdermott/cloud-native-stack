package recipe

import (
	"context"
	"testing"
)

// TestOverlayAddsNewComponent verifies that overlay recipes can add components
// that don't exist in the base recipe.
func TestOverlayAddsNewComponent(t *testing.T) {
	ctx := context.Background()

	// Build recipe for H100 inference workload
	// h100-inference.yaml adds network-operator which is NOT in base.yaml
	builder := NewBuilder()
	criteria := &Criteria{
		Accelerator: CriteriaAcceleratorH100,
		Intent:      CriteriaIntentInference,
	}

	result, err := builder.BuildFromCriteria(ctx, criteria)
	if err != nil {
		t.Fatalf("BuildFromCriteria failed: %v", err)
	}

	if result == nil {
		t.Fatal("Recipe result is nil")
	}

	// Verify base components exist
	baseComponents := []string{"cert-manager", "gpu-operator", "nvsentinel", "skyhook"}
	for _, name := range baseComponents {
		if comp := result.GetComponentRef(name); comp == nil {
			t.Errorf("Base component %q not found in result", name)
		}
	}

	// Verify overlay-added component exists
	networkOp := result.GetComponentRef("network-operator")
	if networkOp == nil {
		t.Fatalf("network-operator not found (should be added by h100-inference overlay)")
	}

	// Verify network-operator properties
	if networkOp.Version == "" {
		t.Error("network-operator has empty version")
	}
	if networkOp.Type != "Helm" {
		t.Errorf("network-operator type = %q, want Helm", networkOp.Type)
	}
	if len(networkOp.DependencyRefs) == 0 {
		t.Error("network-operator has no dependencies (should depend on cert-manager)")
	}

	t.Logf("✅ Successfully verified overlay can add new components")
	t.Logf("   Base components: %d", len(baseComponents))
	t.Logf("   Total components: %d", len(result.ComponentRefs))
	t.Logf("   network-operator version: %s", networkOp.Version)
}

// TestOverlayMergeDoesNotLoseBaseComponents verifies that when overlays add
// components, base components are preserved.
func TestOverlayMergeDoesNotLoseBaseComponents(t *testing.T) {
	ctx := context.Background()
	builder := NewBuilder()

	// Build H100 inference recipe (matches overlay that adds network-operator)
	criteria := &Criteria{
		Accelerator: CriteriaAcceleratorH100,
		Intent:      CriteriaIntentInference,
	}

	result, err := builder.BuildFromCriteria(ctx, criteria)
	if err != nil {
		t.Fatalf("BuildFromCriteria failed: %v", err)
	}

	// Verify all 4 base components exist
	expectedBaseComponents := []string{"cert-manager", "gpu-operator", "nvsentinel", "skyhook"}
	for _, name := range expectedBaseComponents {
		if comp := result.GetComponentRef(name); comp == nil {
			t.Errorf("Base component %q missing from overlay result", name)
		}
	}

	// Verify network-operator was added
	networkOp := result.GetComponentRef("network-operator")
	if networkOp == nil {
		t.Error("network-operator not found (should be added by overlay)")
	}

	// Result should have at least 5 components (4 base + 1 added)
	if len(result.ComponentRefs) < 5 {
		t.Errorf("Expected at least 5 components, got %d", len(result.ComponentRefs))
	}

	t.Logf("✅ Base components preserved when overlay adds new components")
	t.Logf("   Total components: %d (4 base + additions)", len(result.ComponentRefs))
	if networkOp != nil {
		t.Logf("   network-operator added: version %s", networkOp.Version)
	}
}
