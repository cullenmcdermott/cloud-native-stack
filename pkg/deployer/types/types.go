package types

// DeployerType identifies different deployment method types.
type DeployerType string

const (
	// DeployerTypeScript generates shell scripts for manual deployment.
	DeployerTypeScript DeployerType = "script"

	// DeployerTypeArgoCD generates ArgoCD Application manifests.
	DeployerTypeArgoCD DeployerType = "argocd"

	// DeployerTypeFlux generates Flux Kustomization resources.
	DeployerTypeFlux DeployerType = "flux"
)

// String returns the string representation of the deployer type.
func (d DeployerType) String() string {
	return string(d)
}

// IsValid checks if the deployer type is one of the supported types.
func (d DeployerType) IsValid() bool {
	switch d {
	case DeployerTypeScript, DeployerTypeArgoCD, DeployerTypeFlux:
		return true
	default:
		return false
	}
}

// AllDeployerTypes returns all valid deployer types.
func AllDeployerTypes() []DeployerType {
	return []DeployerType{
		DeployerTypeScript,
		DeployerTypeArgoCD,
		DeployerTypeFlux,
	}
}
