package networkoperator

import (
	"context"
	"testing"

	"github.com/NVIDIA/cloud-native-stack/pkg/bundler/config"
	"github.com/NVIDIA/cloud-native-stack/pkg/bundler/internal"
)

func TestNewBundler(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "with nil config",
			cfg:  nil,
		},
		{
			name: "with valid config",
			cfg: config.NewConfig(
				config.WithNamespace("test-namespace"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBundler(tt.cfg)
			if b == nil {
				t.Fatal("NewBundler() returned nil")
			}
			if b.Config == nil {
				t.Error("Bundler config should not be nil")
			}
		})
	}
}

func TestBundler_Make(t *testing.T) {
	tests := []struct {
		name         string
		recipeFunc   func() *internal.RecipeBuilder
		wantErr      bool
		validateFunc func(*testing.T, string)
	}{
		{
			name:       "valid recipe",
			recipeFunc: createTestRecipe,
			wantErr:    false,
		},
		{
			name:       "invalid recipe",
			recipeFunc: internal.NewRecipeBuilder, // Empty recipe
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			b := NewBundler(nil)
			ctx := context.Background()

			rec := tt.recipeFunc().Build()
			result, err := b.Make(ctx, rec, tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("Make() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("Make() returned nil result")
					return
				}
				if len(result.Files) == 0 {
					t.Error("Make() returned no files")
				}

				if tt.validateFunc != nil {
					tt.validateFunc(t, tmpDir)
				}
			}
		})
	}
}

func TestGetTemplate(t *testing.T) {
	internal.TestTemplateGetter(t, GetTemplate, []string{
		"values.yaml",
		"nicclusterpolicy",
		"install.sh",
		"uninstall.sh",
		"README.md",
	})
}

func TestBundler_validateRecipe(t *testing.T) {
	b := NewBundler(nil)
	internal.TestValidateRecipe(t, b.validateRecipe)
}

// Helper function to create a test recipe
func createTestRecipe() *internal.RecipeBuilder {
	return internal.NewRecipeBuilder().
		WithK8sMeasurement(
			internal.ConfigSubtype(map[string]interface{}{
				"rdma-enabled":             true,
				"sr-iov-enabled":           true,
				"ofed-version":             "24.07",
				"network-operator-version": "25.4.0",
			}),
		)
}
