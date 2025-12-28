package gpuoperator

import (
	"fmt"
	"time"

	"github.com/NVIDIA/cloud-native-stack/pkg/measurement"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

const (
	strTrue = "true"
)

// HelmValues represents the data structure for GPU Operator Helm values.
type HelmValues struct {
	Timestamp                     string
	DriverRegistry                string
	GPUOperatorVersion            string
	EnableDriver                  bool
	DriverVersion                 string
	UseOpenKernelModule           bool
	NvidiaContainerToolkitVersion string
	DevicePluginVersion           string
	DCGMVersion                   string
	DCGMExporterVersion           string
	MIGStrategy                   string
	EnableGDS                     bool
	VGPULicenseServer             string
	EnableSecureBoot              bool
	CustomLabels                  map[string]string
	Namespace                     string
}

// GenerateHelmValues generates Helm values from a recipe.
func GenerateHelmValues(recipe *recipe.Recipe, config map[string]string) *HelmValues {
	values := &HelmValues{
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
		DriverRegistry:   getConfigValue(config, "driver_registry", "nvcr.io/nvidia"),
		EnableDriver:     true,
		MIGStrategy:      "single",
		EnableGDS:        false,
		EnableSecureBoot: false,
		CustomLabels:     make(map[string]string),
		Namespace:        getConfigValue(config, "namespace", "gpu-operator"),
	}

	// Extract GPU Operator configuration from recipe measurements
	for _, m := range recipe.Measurements {
		switch m.Type {
		case measurement.TypeK8s:
			values.extractK8sSettings(m)
		case measurement.TypeGPU:
			values.extractGPUSettings(m)
		case measurement.TypeSystemD, measurement.TypeOS:
			// Not used for Helm values generation
		}
	}

	// Apply config overrides
	values.applyConfigOverrides(config)

	return values
}

// extractK8sSettings extracts Kubernetes-related settings from measurements.
func (v *HelmValues) extractK8sSettings(m *measurement.Measurement) {
	for _, st := range m.Subtypes {
		// Extract version information from 'image' subtype
		if st.Name == "image" {
			if val, ok := st.Data["gpu-operator"]; ok {
				if s, ok := val.Any().(string); ok {
					v.GPUOperatorVersion = s
				}
			}
			if val, ok := st.Data["driver"]; ok {
				if s, ok := val.Any().(string); ok {
					v.DriverVersion = s
				}
			}
			if val, ok := st.Data["container-toolkit"]; ok {
				if s, ok := val.Any().(string); ok {
					v.NvidiaContainerToolkitVersion = s
				}
			}
			if val, ok := st.Data["k8s-device-plugin"]; ok {
				if s, ok := val.Any().(string); ok {
					v.DevicePluginVersion = s
				}
			}
			if val, ok := st.Data["dcgm"]; ok {
				if s, ok := val.Any().(string); ok {
					v.DCGMVersion = s
				}
			}
			if val, ok := st.Data["dcgm-exporter"]; ok {
				if s, ok := val.Any().(string); ok {
					v.DCGMExporterVersion = s
				}
			}
		}

		// Extract configuration flags from 'config' subtype
		if st.Name == "config" {
			// MIG configuration (boolean in recipe)
			if val, ok := st.Data["mig"]; ok {
				if b, ok := val.Any().(bool); ok && b {
					v.MIGStrategy = "mixed"
				}
			}
			// UseOpenKernelModule (camelCase in recipe)
			if val, ok := st.Data["useOpenKernelModule"]; ok {
				if b, ok := val.Any().(bool); ok {
					v.UseOpenKernelModule = b
				}
			}
			// RDMA support (affects GDS)
			if val, ok := st.Data["rdma"]; ok {
				if b, ok := val.Any().(bool); ok {
					v.EnableGDS = b
				}
			}
		}
	}
}

// extractGPUSettings extracts GPU-related settings from measurements.
func (v *HelmValues) extractGPUSettings(m *measurement.Measurement) {
	for _, st := range m.Subtypes {
		// Recipe uses 'smi' subtype for nvidia-smi output
		if st.Name == "smi" {
			if val, ok := st.Data["driver-version"]; ok {
				if s, ok := val.Any().(string); ok && v.DriverVersion == "" {
					v.DriverVersion = s
				}
			}
		}
	}
}

// applyConfigOverrides applies configuration overrides to values.
func (v *HelmValues) applyConfigOverrides(config map[string]string) {
	if val, ok := config["driver_version"]; ok && val != "" {
		v.DriverVersion = val
	}
	if val, ok := config["gpu_operator_version"]; ok && val != "" {
		v.GPUOperatorVersion = val
	}
	if val, ok := config["mig_strategy"]; ok && val != "" {
		v.MIGStrategy = val
	}
	if val, ok := config["enable_gds"]; ok {
		v.EnableGDS = val == strTrue
	}
	if val, ok := config["vgpu_license_server"]; ok && val != "" {
		v.VGPULicenseServer = val
	}
	if val, ok := config["namespace"]; ok && val != "" {
		v.Namespace = val
	}

	// Custom labels
	for k, val := range config {
		if len(k) > 6 && k[:6] == "label_" {
			v.CustomLabels[k[6:]] = val
		}
	}
}

// getConfigValue gets a value from config with a default fallback.
func getConfigValue(config map[string]string, key, defaultValue string) string {
	if val, ok := config[key]; ok && val != "" {
		return val
	}
	return defaultValue
}

// ToMap converts HelmValues to a map for template rendering.
func (v *HelmValues) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Timestamp":                     v.Timestamp,
		"DriverRegistry":                v.DriverRegistry,
		"GPUOperatorVersion":            v.GPUOperatorVersion,
		"EnableDriver":                  v.EnableDriver,
		"DriverVersion":                 v.DriverVersion,
		"UseOpenKernelModule":           v.UseOpenKernelModule,
		"NvidiaContainerToolkitVersion": v.NvidiaContainerToolkitVersion,
		"DevicePluginVersion":           v.DevicePluginVersion,
		"DCGMVersion":                   v.DCGMVersion,
		"DCGMExporterVersion":           v.DCGMExporterVersion,
		"MIGStrategy":                   v.MIGStrategy,
		"EnableGDS":                     v.EnableGDS,
		"VGPULicenseServer":             v.VGPULicenseServer,
		"EnableSecureBoot":              v.EnableSecureBoot,
		"CustomLabels":                  v.CustomLabels,
		"Namespace":                     v.Namespace,
	}
}

// Validate validates the Helm values.
func (v *HelmValues) Validate() error {
	if v.Namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	if v.DriverRegistry == "" {
		return fmt.Errorf("driver registry cannot be empty")
	}
	if v.MIGStrategy != "single" && v.MIGStrategy != "mixed" {
		return fmt.Errorf("invalid MIG strategy: %s (must be single or mixed)", v.MIGStrategy)
	}
	return nil
}
