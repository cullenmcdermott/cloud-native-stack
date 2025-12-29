package networkoperator

import (
	"time"

	"github.com/NVIDIA/cloud-native-stack/pkg/measurement"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

// ManifestData represents data for generating Kubernetes manifests.
type ManifestData struct {
	Timestamp              string
	Namespace              string
	EnableRDMA             bool
	EnableSRIOV            bool
	EnableHostDevice       bool
	EnableIPAM             bool
	DeployOFED             bool
	OFEDVersion            string
	NicType                string
	ContainerRuntimeSocket string
	CustomLabels           map[string]string
	CustomAnnotations      map[string]string
}

// GenerateManifestData creates manifest data from a recipe and config.
func GenerateManifestData(recipe *recipe.Recipe, config map[string]string) *ManifestData {
	data := &ManifestData{
		Timestamp:              time.Now().UTC().Format(time.RFC3339),
		Namespace:              getConfigValue(config, "namespace", "nvidia-network-operator"),
		EnableRDMA:             false,
		EnableSRIOV:            false,
		EnableHostDevice:       true,
		EnableIPAM:             true,
		DeployOFED:             false,
		NicType:                "ConnectX",
		ContainerRuntimeSocket: "/var/run/containerd/containerd.sock",
		CustomLabels:           make(map[string]string),
		CustomAnnotations:      make(map[string]string),
	}

	// Extract values from recipe (similar to HelmValues)
	helmValues := GenerateHelmValues(recipe, config)

	// Convert helm values to manifest data
	data.EnableRDMA = helmValues.EnableRDMA
	data.EnableSRIOV = helmValues.EnableSRIOV
	data.EnableHostDevice = helmValues.EnableHostDevice
	data.EnableIPAM = helmValues.EnableIPAM
	data.DeployOFED = helmValues.DeployOFED
	data.OFEDVersion = helmValues.OFEDVersion
	data.NicType = helmValues.NicType
	data.ContainerRuntimeSocket = helmValues.ContainerRuntimeSocket
	data.CustomLabels = helmValues.CustomLabels

	// Extract additional settings from K8s config subtype
	for _, m := range recipe.Measurements {
		if m.Type == measurement.TypeK8s {
			for _, st := range m.Subtypes {
				if st.Name == configSubtype {
					// Additional manifest-specific settings can be extracted here
					// Currently all settings are extracted via helm values
					_ = st // Avoid unused variable
				}
			}
		}
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
		"Timestamp":              m.Timestamp,
		"Namespace":              m.Namespace,
		"EnableRDMA":             m.EnableRDMA,
		"EnableSRIOV":            m.EnableSRIOV,
		"EnableHostDevice":       m.EnableHostDevice,
		"EnableIPAM":             m.EnableIPAM,
		"DeployOFED":             m.DeployOFED,
		"OFEDVersion":            m.OFEDVersion,
		"NicType":                m.NicType,
		"ContainerRuntimeSocket": m.ContainerRuntimeSocket,
		"CustomLabels":           m.CustomLabels,
		"CustomAnnotations":      m.CustomAnnotations,
	}
}
