package types

import "testing"

func TestDeployerType_String(t *testing.T) {
	tests := []struct {
		name string
		dt   DeployerType
		want string
	}{
		{"script", DeployerTypeScript, "script"},
		{"argocd", DeployerTypeArgoCD, "argocd"},
		{"flux", DeployerTypeFlux, "flux"},
		{"custom", DeployerType("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dt.String(); got != tt.want {
				t.Errorf("DeployerType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeployerType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		dt   DeployerType
		want bool
	}{
		{"script valid", DeployerTypeScript, true},
		{"argocd valid", DeployerTypeArgoCD, true},
		{"flux valid", DeployerTypeFlux, true},
		{"empty invalid", DeployerType(""), false},
		{"unknown invalid", DeployerType("unknown"), false},
		{"custom invalid", DeployerType("custom"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dt.IsValid(); got != tt.want {
				t.Errorf("DeployerType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllDeployerTypes(t *testing.T) {
	got := AllDeployerTypes()

	if len(got) != 3 {
		t.Errorf("AllDeployerTypes() returned %d types, want 3", len(got))
	}

	expected := map[DeployerType]bool{
		DeployerTypeScript: false,
		DeployerTypeArgoCD: false,
		DeployerTypeFlux:   false,
	}

	for _, dt := range got {
		if _, exists := expected[dt]; !exists {
			t.Errorf("AllDeployerTypes() contains unexpected type: %s", dt)
		}
		expected[dt] = true
	}

	for dt, found := range expected {
		if !found {
			t.Errorf("AllDeployerTypes() missing expected type: %s", dt)
		}
	}
}

func TestAllDeployerTypes_AllValid(t *testing.T) {
	for _, dt := range AllDeployerTypes() {
		if !dt.IsValid() {
			t.Errorf("AllDeployerTypes() contains invalid type: %s", dt)
		}
	}
}
