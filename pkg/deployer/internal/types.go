package internal

// ComponentInfo contains component information for templates.
// Used across all deployer implementations.
type ComponentInfo struct {
	Name    string
	Version string
}

// ReadmeData contains data for README template rendering.
// Used by script and argocd deployers.
type ReadmeData struct {
	Timestamp     string
	RecipeVersion string
	Components    []ComponentInfo
}
