package gpuoperator

import (
	"time"

	"github.com/NVIDIA/cloud-native-stack/pkg/measurement"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

// ManifestData represents data for generating Kubernetes manifests.
type ManifestData struct {
	Timestamp           string
	Namespace           string
	EnableDriver        bool
	DriverVersion       string
	UseOpenKernelModule bool
	MIGStrategy         string
	EnableGDS           bool
	EnableVGPU          bool
	VGPULicenseServer   string
	EnableCDI           bool
	CustomLabels        map[string]string
	CustomAnnotations   map[string]string
}

// GenerateManifestData creates manifest data from a recipe and config.
func GenerateManifestData(recipe *recipe.Recipe, config map[string]string) *ManifestData {
	data := &ManifestData{
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		Namespace:         getConfigValue(config, "namespace", "gpu-operator"),
		EnableDriver:      true,
		MIGStrategy:       "single",
		EnableGDS:         false,
		EnableVGPU:        false,
		EnableCDI:         false,
		CustomLabels:      make(map[string]string),
		CustomAnnotations: make(map[string]string),
	}

	// Extract values from recipe (similar to HelmValues)
	helmValues := GenerateHelmValues(recipe, config)

	// Convert helm values to manifest data
	data.DriverVersion = helmValues.DriverVersion
	data.UseOpenKernelModule = helmValues.UseOpenKernelModule
	data.MIGStrategy = helmValues.MIGStrategy
	data.EnableGDS = helmValues.EnableGDS
	data.CustomLabels = helmValues.CustomLabels

	// Extract CDI setting from K8s config subtype
	for _, m := range recipe.Measurements {
		if m.Type == measurement.TypeK8s {
			for _, st := range m.Subtypes {
				if st.Name == "config" {
					if val, ok := st.Data["cdi"]; ok {
						if b, ok := val.Any().(bool); ok {
							data.EnableCDI = b
						}
					}
				}
			}
		}
	}

	// Apply config-specific manifest settings (overrides)
	if val, ok := config["enable_vgpu"]; ok {
		data.EnableVGPU = val == "true"
	}
	if val, ok := config["vgpu_license_server"]; ok && val != "" {
		data.VGPULicenseServer = val
	}
	if val, ok := config["enable_cdi"]; ok {
		data.EnableCDI = val == "true"
	}

	// Custom annotations
	for k, val := range config {
		if len(k) > 11 && k[:11] == "annotation_" {
			data.CustomAnnotations[k[11:]] = val
		}
	}

	return data
}

// ToMap converts ManifestData to a map for template rendering.
func (m *ManifestData) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Timestamp":           m.Timestamp,
		"Namespace":           m.Namespace,
		"EnableDriver":        m.EnableDriver,
		"DriverVersion":       m.DriverVersion,
		"UseOpenKernelModule": m.UseOpenKernelModule,
		"MIGStrategy":         m.MIGStrategy,
		"EnableGDS":           m.EnableGDS,
		"EnableVGPU":          m.EnableVGPU,
		"VGPULicenseServer":   m.VGPULicenseServer,
		"EnableCDI":           m.EnableCDI,
		"CustomLabels":        m.CustomLabels,
		"CustomAnnotations":   m.CustomAnnotations,
	}
}
