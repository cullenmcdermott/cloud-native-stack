package examples

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/NVIDIA/cloud-native-stack/pkg/bundler"
	"github.com/NVIDIA/cloud-native-stack/pkg/measurement"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

func TestSimpleBundler_Make(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	rec := &recipe.Recipe{
		Measurements: []*measurement.Measurement{
			{
				Type: measurement.TypeK8s,
				Subtypes: []measurement.Subtype{
					{
						Name: "cluster",
						Data: map[string]measurement.Reading{
							"version": measurement.Str("1.28.0"),
						},
					},
				},
			},
		},
	}

	b := NewSimpleBundler()
	result, err := b.Make(ctx, rec, tmpDir)
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	if result == nil {
		t.Fatal("Make() returned nil result")
	}

	if !result.Success {
		t.Error("Make() should succeed")
	}

	if len(result.Files) == 0 {
		t.Error("Make() produced no files")
	}

	// Verify bundle directory
	bundleDir := filepath.Join(tmpDir, "simple")
	if _, err := os.Stat(bundleDir); os.IsNotExist(err) {
		t.Error("Make() did not create simple directory")
	}

	// Verify expected files
	expectedFiles := []string{"summary.txt", "metadata.txt"}
	for _, file := range expectedFiles {
		path := filepath.Join(bundleDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s not found", file)
		}
	}
}

func TestSimpleBundler_Validate(t *testing.T) {
	ctx := context.Background()
	b := NewSimpleBundler().(*SimpleBundler)

	tests := []struct {
		name    string
		recipe  *recipe.Recipe
		wantErr bool
	}{
		{
			name: "valid recipe",
			recipe: &recipe.Recipe{
				Measurements: []*measurement.Measurement{
					{Type: measurement.TypeK8s},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil recipe",
			recipe:  nil,
			wantErr: true,
		},
		{
			name: "empty measurements",
			recipe: &recipe.Recipe{
				Measurements: []*measurement.Measurement{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := b.Validate(ctx, tt.recipe)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSimpleBundler_Configure(t *testing.T) {
	b := NewSimpleBundler().(*SimpleBundler)

	config := bundler.DefaultBundlerConfig()
	config.Namespace = "test-namespace"
	config.CustomLabels = map[string]string{
		"app": "test",
	}

	err := b.Configure(config)
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
}

func TestSimpleBundlerRegistered(t *testing.T) {
	// Verify the simple bundler is registered in the global registry
	types := []string{"simple", "gpu-operator"}
	found := false
	for _, typ := range types {
		if bundler.BundleType(typ) == SimpleBundlerType {
			found = true
			break
		}
	}
	if !found {
		t.Error("Simple bundler type not defined correctly")
	}
}
