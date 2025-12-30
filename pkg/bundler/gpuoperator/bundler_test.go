package gpuoperator

import (
	"testing"

	"github.com/NVIDIA/cloud-native-stack/pkg/bundler/config"
	"github.com/NVIDIA/cloud-native-stack/pkg/bundler/internal"
	"github.com/NVIDIA/cloud-native-stack/pkg/measurement"
)

const testNamespace = "test-ns"

func TestBundler_Make(t *testing.T) {
	harness := internal.NewTestHarness(t, "gpu-operator").
		WithExpectedFiles([]string{
			"values.yaml",
			"manifests/clusterpolicy.yaml",
			"scripts/install.sh",
			"scripts/uninstall.sh",
			"README.md",
			"checksums.txt",
		}).
		WithRecipeBuilder(createTestRecipe)

	b := NewBundler(config.NewConfig())
	harness.TestMake(b)
}

func TestBundler_Validate(t *testing.T) {
	tests := []struct {
		name    string
		recipe  *internal.RecipeBuilder
		wantErr bool
	}{
		{
			name: "valid recipe",
			recipe: internal.NewRecipeBuilder().
				WithK8sMeasurement(internal.ImageSubtype(map[string]string{
					"gpu-operator": "v25.3.3",
				})),
			wantErr: false,
		},
		{
			name: "missing K8s measurements",
			recipe: internal.NewRecipeBuilder().
				WithGPUMeasurement(internal.SMISubtype(map[string]string{
					"driver-version": "580.82.07",
				})),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := tt.recipe.Build()
			err := rec.ValidateMeasurementExists(measurement.TypeK8s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMeasurementExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBundler_Configure(t *testing.T) {
	cfg := config.NewConfig()
	// Note: Config is now immutable. We cannot assign fields directly.
	// Instead, bundlers should receive pre-configured Config instances.

	// Create a test config for verification purposes
	b := NewBundler(cfg)

	// Verify default values
	if b.Config.Namespace() != "default" {
		t.Errorf("Configure() namespace = %s, want default", b.Config.Namespace())
	}
}

func TestGenerateHelmValues(t *testing.T) {
	rec := createTestRecipe().Build()
	config := make(map[string]string)
	config["namespace"] = testNamespace

	values := GenerateHelmValues(rec, config)

	if values == nil {
		t.Fatal("GenerateHelmValues() returned nil")
		return
	}

	if values.Namespace != testNamespace {
		t.Errorf("Namespace = %s, want %s", values.Namespace, testNamespace)
	}

	// Verify extracted values from recipe match expected structure
	internal.AssertValueWithContext(t, values.GPUOperatorVersion, "v25.3.3", "GPUOperatorVersion")
	internal.AssertValueWithContext(t, values.DriverVersion, "580.82.07", "DriverVersion")
	internal.AssertValueWithContext(t, values.NvidiaContainerToolkitVersion, "v1.17.8", "NvidiaContainerToolkitVersion")
	internal.AssertValueWithContext(t, values.DevicePluginVersion, "v0.17.4", "DevicePluginVersion")
	internal.AssertValueWithContext(t, values.DCGMVersion, "4.3.1-1", "DCGMVersion")
	internal.AssertValueWithContext(t, values.DCGMExporterVersion, "4.3.1", "DCGMExporterVersion")
	internal.AssertValueWithContext(t, values.UseOpenKernelModule, true, "UseOpenKernelModule")
	internal.AssertValueWithContext(t, values.EnableGDS, true, "EnableGDS")
	internal.AssertValueWithContext(t, values.MIGStrategy, "single", "MIGStrategy")

	if err := values.Validate(); err != nil {
		t.Errorf("HelmValues.Validate() error = %v", err)
	}
}

func TestGenerateManifestData(t *testing.T) {
	rec := createTestRecipe().Build()
	config := make(map[string]string)
	config["namespace"] = testNamespace

	data := GenerateManifestData(rec, config)

	if data == nil {
		t.Fatal("GenerateManifestData() returned nil")
		return
	}

	if data.Namespace != testNamespace {
		t.Errorf("Namespace = %s, want %s", data.Namespace, testNamespace)
	}

	// Verify CDI is extracted from recipe config subtype
	if !data.EnableCDI {
		t.Error("EnableCDI = false, want true (from recipe config)")
	}

	// Verify other values propagated from HelmValues
	if data.DriverVersion != "580.82.07" {
		t.Errorf("DriverVersion = %s, want 580.82.07", data.DriverVersion)
	}

	if !data.UseOpenKernelModule {
		t.Error("UseOpenKernelModule = false, want true")
	}

	if !data.EnableGDS {
		t.Error("EnableGDS = false, want true")
	}
}

func TestGenerateScriptData(t *testing.T) {
	rec := createTestRecipe().Build()
	config := make(map[string]string)
	config["helm_repository"] = "https://test.repo"

	data := GenerateScriptData(rec, config)

	if data == nil {
		t.Fatal("GenerateScriptData() returned nil")
		return
	}

	internal.AssertConfigValue(t, config, "helm_repository", "https://test.repo")
}

func TestGetTemplate(t *testing.T) {
	internal.TestTemplateGetter(t, GetTemplate, []string{
		"values.yaml",
		"clusterpolicy",
		"install.sh",
		"uninstall.sh",
		"README.md",
	})
}

func TestHelmValues_Validate(t *testing.T) {
	tests := []struct {
		name    string
		values  *HelmValues
		wantErr bool
	}{
		{
			name: "valid values",
			values: &HelmValues{
				Namespace:      "test",
				DriverRegistry: internal.ValueWithContext{Value: "nvcr.io/nvidia"},
				MIGStrategy:    internal.ValueWithContext{Value: "single"},
			},
			wantErr: false,
		},
		{
			name: "empty namespace",
			values: &HelmValues{
				Namespace:      "",
				DriverRegistry: internal.ValueWithContext{Value: "nvcr.io/nvidia"},
				MIGStrategy:    internal.ValueWithContext{Value: "single"},
			},
			wantErr: true,
		},
		{
			name: "invalid MIG strategy",
			values: &HelmValues{
				Namespace:      "test",
				DriverRegistry: internal.ValueWithContext{Value: "nvcr.io/nvidia"},
				MIGStrategy:    internal.ValueWithContext{Value: "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.values.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// createTestRecipe creates a recipe for testing that matches the actual recipe.yaml structure.
func createTestRecipe() *internal.RecipeBuilder {
	return internal.NewRecipeBuilder().
		WithK8sMeasurement(
			internal.ImageSubtype(map[string]string{
				"gpu-operator":      "v25.3.3",
				"driver":            "580.82.07",
				"container-toolkit": "v1.17.8",
				"k8s-device-plugin": "v0.17.4",
				"dcgm":              "4.3.1-1",
				"dcgm-exporter":     "4.3.1",
			}),
			internal.ConfigSubtype(map[string]interface{}{
				"cdi":                 true,
				"mig":                 false,
				"rdma":                true,
				"useOpenKernelModule": true,
			}),
		).
		WithGPUMeasurement(
			internal.SMISubtype(map[string]string{
				"driver-version": "580.82.07",
				"cuda-version":   "13.1",
			}),
		)
}
