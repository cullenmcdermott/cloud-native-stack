package internal

// GetNamespaceForComponent returns the appropriate namespace for a component.
// This centralizes namespace mapping logic across all deployers.
func GetNamespaceForComponent(componentName string) string {
	switch componentName {
	case "gpu-operator":
		return "gpu-operator"
	case "network-operator":
		return "network-operator"
	case "cert-manager":
		return "cert-manager"
	case "nvsentinel":
		return "nvsentinel"
	case "skyhook":
		return "skyhook"
	default:
		return "default"
	}
}
