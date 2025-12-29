package networkoperator

import (
	"fmt"
	"time"

	"github.com/NVIDIA/cloud-native-stack/pkg/measurement"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

const (
	strTrue = "true"
)

// HelmValues represents the data structure for Network Operator Helm values.
type HelmValues struct {
	Timestamp              string
	NetworkOperatorVersion string
	OFEDVersion            string
	EnableRDMA             bool
	EnableSRIOV            bool
	EnableHostDevice       bool
	EnableIPAM             bool
	EnableMultus           bool
	EnableWhereabouts      bool
	DeployOFED             bool
	NicType                string
	ContainerRuntimeSocket string
	CustomLabels           map[string]string
	Namespace              string
}

// GenerateHelmValues generates Helm values from a recipe.
func GenerateHelmValues(recipe *recipe.Recipe, config map[string]string) *HelmValues {
	values := &HelmValues{
		Timestamp:              time.Now().UTC().Format(time.RFC3339),
		EnableRDMA:             false,
		EnableSRIOV:            false,
		EnableHostDevice:       true,
		EnableIPAM:             true,
		EnableMultus:           true,
		EnableWhereabouts:      true,
		DeployOFED:             false,
		NicType:                "ConnectX",
		ContainerRuntimeSocket: "/var/run/containerd/containerd.sock",
		CustomLabels:           make(map[string]string),
		Namespace:              getConfigValue(config, "namespace", "nvidia-network-operator"),
	}

	// Extract Network Operator configuration from recipe measurements
	for _, m := range recipe.Measurements {
		switch m.Type {
		case measurement.TypeK8s:
			values.extractK8sSettings(m)
		case measurement.TypeSystemD, measurement.TypeOS, measurement.TypeGPU:
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
			if val, ok := st.Data["network-operator"]; ok {
				if s, ok := val.Any().(string); ok {
					v.NetworkOperatorVersion = s
				}
			}
			if val, ok := st.Data["ofed-driver"]; ok {
				if s, ok := val.Any().(string); ok {
					v.OFEDVersion = s
				}
			}
		}

		// Extract configuration flags from 'config' subtype
		if st.Name == "config" {
			// RDMA configuration
			if val, ok := st.Data["rdma"]; ok {
				if b, ok := val.Any().(bool); ok {
					v.EnableRDMA = b
				}
			}
			// SR-IOV configuration
			if val, ok := st.Data["sr-iov"]; ok {
				if b, ok := val.Any().(bool); ok {
					v.EnableSRIOV = b
				}
			}
			// OFED deployment
			if val, ok := st.Data["deploy-ofed"]; ok {
				if b, ok := val.Any().(bool); ok {
					v.DeployOFED = b
				}
			}
			// Host device plugin
			if val, ok := st.Data["host-device"]; ok {
				if b, ok := val.Any().(bool); ok {
					v.EnableHostDevice = b
				}
			}
			// IPAM plugin
			if val, ok := st.Data["ipam"]; ok {
				if b, ok := val.Any().(bool); ok {
					v.EnableIPAM = b
				}
			}
			// Multus CNI
			if val, ok := st.Data["multus"]; ok {
				if b, ok := val.Any().(bool); ok {
					v.EnableMultus = b
				}
			}
			// Whereabouts IPAM
			if val, ok := st.Data["whereabouts"]; ok {
				if b, ok := val.Any().(bool); ok {
					v.EnableWhereabouts = b
				}
			}
			// NIC type
			if val, ok := st.Data["nic-type"]; ok {
				if s, ok := val.Any().(string); ok {
					v.NicType = s
				}
			}
		}

		// Extract container runtime from 'server' subtype
		if st.Name == "server" {
			if val, ok := st.Data["container-runtime"]; ok {
				if s, ok := val.Any().(string); ok {
					switch s {
					case "containerd":
						v.ContainerRuntimeSocket = "/var/run/containerd/containerd.sock"
					case "docker":
						v.ContainerRuntimeSocket = "/var/run/docker.sock"
					case "cri-o":
						v.ContainerRuntimeSocket = "/var/run/crio/crio.sock"
					}
				}
			}
		}
	}
}

// applyConfigOverrides applies configuration overrides to values.
func (v *HelmValues) applyConfigOverrides(config map[string]string) {
	if val, ok := config["network_operator_version"]; ok && val != "" {
		v.NetworkOperatorVersion = val
	}
	if val, ok := config["ofed_version"]; ok && val != "" {
		v.OFEDVersion = val
	}
	if val, ok := config["enable_rdma"]; ok {
		v.EnableRDMA = val == strTrue
	}
	if val, ok := config["enable_sriov"]; ok {
		v.EnableSRIOV = val == strTrue
	}
	if val, ok := config["deploy_ofed"]; ok {
		v.DeployOFED = val == strTrue
	}
	if val, ok := config["enable_host_device"]; ok {
		v.EnableHostDevice = val == strTrue
	}
	if val, ok := config["enable_ipam"]; ok {
		v.EnableIPAM = val == strTrue
	}
	if val, ok := config["enable_multus"]; ok {
		v.EnableMultus = val == strTrue
	}
	if val, ok := config["enable_whereabouts"]; ok {
		v.EnableWhereabouts = val == strTrue
	}
	if val, ok := config["nic_type"]; ok && val != "" {
		v.NicType = val
	}
	if val, ok := config["container_runtime_socket"]; ok && val != "" {
		v.ContainerRuntimeSocket = val
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
		"Timestamp":              v.Timestamp,
		"NetworkOperatorVersion": v.NetworkOperatorVersion,
		"OFEDVersion":            v.OFEDVersion,
		"EnableRDMA":             v.EnableRDMA,
		"EnableSRIOV":            v.EnableSRIOV,
		"EnableHostDevice":       v.EnableHostDevice,
		"EnableIPAM":             v.EnableIPAM,
		"EnableMultus":           v.EnableMultus,
		"EnableWhereabouts":      v.EnableWhereabouts,
		"DeployOFED":             v.DeployOFED,
		"NicType":                v.NicType,
		"ContainerRuntimeSocket": v.ContainerRuntimeSocket,
		"CustomLabels":           v.CustomLabels,
		"Namespace":              v.Namespace,
	}
}

// Validate validates the Helm values.
func (v *HelmValues) Validate() error {
	if v.Namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	if v.NicType == "" {
		return fmt.Errorf("NIC type cannot be empty")
	}
	if v.ContainerRuntimeSocket == "" {
		return fmt.Errorf("container runtime socket cannot be empty")
	}
	return nil
}
