package bundler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/NVIDIA/cloud-native-stack/pkg/measurement"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

// MockBundler for testing.
type mockBundler struct{}

func (m *mockBundler) Make(ctx context.Context, r *recipe.Recipe, outputDir string) (*BundleResult, error) {
	result := NewBundleResult("mock")
	result.AddFile("test.txt", 100)
	result.MarkSuccess()
	return result, nil
}

func init() {
	Register("mock", &mockBundler{})
}

func TestMake(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a test recipe
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

	output, err := Make(ctx, rec, tmpDir, WithBundlers("mock"))
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	if output == nil {
		t.Fatal("Make() returned nil output")
	}

	if len(output.Results) == 0 {
		t.Error("Make() produced no results")
	}

	if output.OutputDir != tmpDir {
		t.Errorf("OutputDir = %s, want %s", output.OutputDir, tmpDir)
	}
}

func TestMakeWithNilRecipe(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, err := Make(ctx, nil, tmpDir)
	if err == nil {
		t.Error("Make() with nil recipe should return error")
	}
}

func TestMakeWithEmptyMeasurements(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	rec := &recipe.Recipe{
		Measurements: []*measurement.Measurement{},
	}

	_, err := Make(ctx, rec, tmpDir)
	if err == nil {
		t.Error("Make() with empty measurements should return error")
	}
}

func TestMakeWithOptions(t *testing.T) {
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

	config := DefaultBundlerConfig()
	config.Namespace = "test-namespace"

	output, err := Make(ctx, rec, tmpDir,
		WithBundlers("mock"),
		WithConfig(config),
		WithParallel(),
	)
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	if output == nil {
		t.Fatal("Make() returned nil output")
	}
}

func TestMakeCreatesDirectory(t *testing.T) {
	ctx := context.Background()
	tmpDir := filepath.Join(t.TempDir(), "nested", "dir")

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

	_, err := Make(ctx, rec, tmpDir, WithBundlers("mock"))
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Make() did not create output directory")
	}
}

func TestMakeWithDryRun(t *testing.T) {
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

	output, err := Make(ctx, rec, tmpDir, WithBundlers("mock"), WithDryRun())
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	if output == nil {
		t.Fatal("Make() returned nil output")
	}

	// In dry run mode, no files should be created
	if output.TotalFiles > 0 {
		t.Errorf("DryRun should not create files, got %d files", output.TotalFiles)
	}
}

func TestBundleOutput_Summary(t *testing.T) {
	output := &BundleOutput{
		TotalFiles: 5,
		TotalSize:  1024,
		Results: []*BundleResult{
			{Success: true},
			{Success: true},
			{Success: false},
		},
	}

	summary := output.Summary()
	if summary == "" {
		t.Error("Summary() returned empty string")
	}
}

func TestBundleOutput_HasErrors(t *testing.T) {
	tests := []struct {
		name   string
		output *BundleOutput
		want   bool
	}{
		{
			name: "no errors",
			output: &BundleOutput{
				Errors: []BundleError{},
			},
			want: false,
		},
		{
			name: "with errors",
			output: &BundleOutput{
				Errors: []BundleError{
					{BundlerType: "test", Error: "test error"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.output.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBundleResult_AddFile(t *testing.T) {
	result := NewBundleResult("test")
	result.AddFile("/path/to/file", 100)

	if len(result.Files) != 1 {
		t.Errorf("AddFile() did not add file, got %d files", len(result.Files))
	}

	if result.Size != 100 {
		t.Errorf("AddFile() size = %d, want 100", result.Size)
	}
}

func TestValidateRecipeStructure(t *testing.T) {
	tests := []struct {
		name    string
		recipe  *recipe.Recipe
		wantErr bool
	}{
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
		{
			name: "valid recipe",
			recipe: &recipe.Recipe{
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
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRecipeStructure(tt.recipe)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRecipeStructure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
