package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProvider(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		want       string
	}{
		{
			name:       "AWS EKS",
			providerID: "aws:///us-west-2a/i-0123456789abcdef0",
			want:       "eks",
		},
		{
			name:       "GCP GKE",
			providerID: "gce://my-project/us-central1-a/gke-cluster-default-pool-node-abc123",
			want:       "gke",
		},
		{
			name:       "Azure AKS",
			providerID: "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachines/aks-nodepool1-12345678-vmss000000",
			want:       "aks",
		},
		{
			name:       "OCI OKE",
			providerID: "oci://ocid1.instance.oc1.phx.abcdef123456",
			want:       "oke",
		},
		{
			name:       "empty provider",
			providerID: "",
			want:       "",
		},
		{
			name:       "unknown format",
			providerID: "custom-provider://some-id",
			want:       "custom-provider",
		},
		{
			name:       "no scheme separator",
			providerID: "just-a-string",
			want:       "just-a-string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseProvider(tt.providerID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetNodeName(t *testing.T) {
	tests := []struct {
		name     string
		setEnv   map[string]string
		expected string
	}{
		{
			name:     "NODE_NAME set",
			setEnv:   map[string]string{"NODE_NAME": "test-node-1"},
			expected: "test-node-1",
		},
		{
			name:     "KUBERNETES_NODE_NAME fallback",
			setEnv:   map[string]string{"KUBERNETES_NODE_NAME": "k8s-node-2"},
			expected: "k8s-node-2",
		},
		{
			name:     "HOSTNAME fallback",
			setEnv:   map[string]string{"HOSTNAME": "host-3"},
			expected: "host-3",
		},
		{
			name:     "NODE_NAME takes precedence",
			setEnv:   map[string]string{"NODE_NAME": "node-1", "KUBERNETES_NODE_NAME": "node-2", "HOSTNAME": "node-3"},
			expected: "node-1",
		},
		{
			name:     "no env vars",
			setEnv:   map[string]string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			t.Setenv("NODE_NAME", "")
			t.Setenv("KUBERNETES_NODE_NAME", "")
			t.Setenv("HOSTNAME", "")

			// Set test environment
			for k, v := range tt.setEnv {
				t.Setenv(k, v)
			}

			got := GetNodeName()
			assert.Equal(t, tt.expected, got)
		})
	}
}
