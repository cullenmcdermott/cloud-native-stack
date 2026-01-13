package flux

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/internal"
	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/registry"
	"github.com/NVIDIA/cloud-native-stack/pkg/deployer/types"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

const testVersion = "v1.0.0"

// Verify Deployer implements registry.Deployer interface at compile time
var _ registry.Deployer = (*Deployer)(nil)

func TestNewDeployer(t *testing.T) {
	deployer := NewDeployer()

	if deployer == nil {
		t.Fatal("NewDeployer() returned nil")
	}
}

func TestDeployer_Generate(t *testing.T) {
	deployer := &Deployer{}
	ctx := context.Background()

	// Create test directory with component subdirectories
	tmpDir, err := os.MkdirTemp("", "flux-deployer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create component directories that the Flux deployer expects
	for _, component := range []string{"gpu-operator", "network-operator"} {
		componentDir := filepath.Join(tmpDir, component)
		if mkErr := os.MkdirAll(componentDir, 0755); mkErr != nil {
			t.Fatalf("Failed to create component dir %s: %v", component, mkErr)
		}
	}

	// Create test recipe result
	recipeResult := &recipe.RecipeResult{}
	recipeResult.Metadata.Version = testVersion
	recipeResult.ComponentRefs = []recipe.ComponentRef{
		{
			Name:    "gpu-operator",
			Version: "v25.3.3",
			Source:  "https://helm.ngc.nvidia.com/nvidia",
		},
		{
			Name:    "network-operator",
			Version: "v25.4.0",
			Source:  "https://helm.ngc.nvidia.com/nvidia",
		},
	}

	// Call Generate
	artifacts, err := deployer.Generate(ctx, recipeResult, tmpDir)

	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if artifacts == nil {
		t.Fatal("Generate() returned nil artifacts")
	}

	// Verify success
	if !artifacts.Success {
		t.Errorf("Generate() artifacts.Success = false, want true (error: %s)", artifacts.Error)
	}

	// Verify flux directory was created
	fluxDir := filepath.Join(tmpDir, "flux")
	if _, err := os.Stat(fluxDir); os.IsNotExist(err) {
		t.Error("Generate() did not create flux directory")
	}

	// Verify expected files were created
	expectedFiles := []string{
		"flux/kustomization.yaml",
		"gpu-operator/helmrelease.yaml",
		"network-operator/helmrelease.yaml",
		"README.md",
	}

	for _, expectedFile := range expectedFiles {
		fullPath := filepath.Join(tmpDir, expectedFile)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Generate() did not create %s", expectedFile)
		}
	}

	// Verify Files includes all expected files
	if len(artifacts.Files) < len(expectedFiles) {
		t.Errorf("Generate() returned %d files, want at least %d", len(artifacts.Files), len(expectedFiles))
	}

	// Verify README content
	if artifacts.ReadmeContent == "" {
		t.Error("Generate() did not set ReadmeContent")
	}

	// Verify Duration is set
	if artifacts.Duration == 0 {
		t.Error("Generate() did not set Duration")
	}
}

func TestDeployer_Generate_EmptyComponents(t *testing.T) {
	deployer := &Deployer{}
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "flux-deployer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create recipe with no components
	recipeResult := &recipe.RecipeResult{}
	recipeResult.Metadata.Version = testVersion
	recipeResult.ComponentRefs = []recipe.ComponentRef{}

	artifacts, err := deployer.Generate(ctx, recipeResult, tmpDir)

	if err != nil {
		t.Fatalf("Generate() with empty components failed: %v", err)
	}

	if !artifacts.Success {
		t.Errorf("Generate() with empty components failed: %s", artifacts.Error)
	}

	// Should still create flux directory and parent kustomization
	fluxDir := filepath.Join(tmpDir, "flux")
	if _, err := os.Stat(fluxDir); os.IsNotExist(err) {
		t.Error("Generate() did not create flux directory for empty components")
	}
}

func TestDeployer_Generate_InvalidDirectory(t *testing.T) {
	deployer := &Deployer{}
	ctx := context.Background()

	// Use a non-existent directory
	nonExistentDir := "/nonexistent/path/that/does/not/exist"

	recipeResult := &recipe.RecipeResult{}
	recipeResult.Metadata.Version = testVersion
	recipeResult.ComponentRefs = []recipe.ComponentRef{
		{Name: "test", Version: testVersion},
	}

	artifacts, err := deployer.Generate(ctx, recipeResult, nonExistentDir)

	if err == nil {
		t.Error("Generate() with invalid directory did not return error")
	}

	if artifacts.Success {
		t.Error("Generate() with invalid directory has Success=true")
	}

	if artifacts.Error == "" {
		t.Error("Generate() with invalid directory did not set Error message")
	}
}

func TestInternalGetNamespaceForComponent(t *testing.T) {
	tests := []struct {
		componentName string
		want          string
	}{
		{"gpu-operator", "gpu-operator"},
		{"network-operator", "network-operator"},
		{"cert-manager", "cert-manager"},
		{"unknown-component", "default"},
		{"", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.componentName, func(t *testing.T) {
			got := internal.GetNamespaceForComponent(tt.componentName)
			if got != tt.want {
				t.Errorf("GetNamespaceForComponent(%s) = %s, want %s", tt.componentName, got, tt.want)
			}
		})
	}
}

func TestHelmReleaseData_Fields(t *testing.T) {
	data := HelmReleaseData{
		Namespace: "test-ns",
		Name:      "test-component",
		Source:    "https://example.com",
		Version:   "v1.2.3",
	}

	if data.Namespace != "test-ns" {
		t.Errorf("HelmReleaseData.Namespace = %s, want test-ns", data.Namespace)
	}

	if data.Name != "test-component" {
		t.Errorf("HelmReleaseData.Name = %s, want test-component", data.Name)
	}

	if data.Source != "https://example.com" {
		t.Errorf("HelmReleaseData.Source = %s, want https://example.com", data.Source)
	}

	if data.Version != "v1.2.3" {
		t.Errorf("HelmReleaseData.Version = %s, want v1.2.3", data.Version)
	}
}

func TestInternalReadmeData_Fields(t *testing.T) {
	data := internal.ReadmeData{
		Timestamp: time.Now().Format(time.RFC3339),
		Components: []internal.ComponentInfo{
			{Name: "test-component", Version: "v1.2.3"},
		},
	}

	if data.Timestamp == "" {
		t.Error("ReadmeData.Timestamp is empty")
	}

	if len(data.Components) != 1 {
		t.Errorf("ReadmeData.Components length = %d, want 1", len(data.Components))
	}
}

func TestInternalComponentInfo_Fields(t *testing.T) {
	info := internal.ComponentInfo{
		Name:    "test-component",
		Version: "v1.2.3",
	}

	if info.Name != "test-component" {
		t.Errorf("ComponentInfo.Name = %s, want test-component", info.Name)
	}

	if info.Version != "v1.2.3" {
		t.Errorf("ComponentInfo.Version = %s, want v1.2.3", info.Version)
	}
}

func TestInternalRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{ .Name }}",
			data:     struct{ Name string }{Name: "World"},
			want:     "Hello World",
			wantErr:  false,
		},
		{
			name:     "invalid template syntax",
			template: "Hello {{ .Name",
			data:     struct{ Name string }{Name: "World"},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "nil data",
			template: "Static content",
			data:     nil,
			want:     "Static content",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := internal.RenderTemplate(tt.template, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("renderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("renderTemplate() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDeployer_Generate_VerifiesHelmReleaseContent(t *testing.T) {
	deployer := &Deployer{}
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "flux-deployer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create component directory
	componentDir := filepath.Join(tmpDir, "gpu-operator")
	if mkErr := os.MkdirAll(componentDir, 0755); mkErr != nil {
		t.Fatalf("Failed to create component dir: %v", mkErr)
	}

	recipeResult := &recipe.RecipeResult{}
	recipeResult.Metadata.Version = testVersion
	recipeResult.ComponentRefs = []recipe.ComponentRef{
		{
			Name:    "gpu-operator",
			Version: "v25.3.3",
			Source:  "https://helm.ngc.nvidia.com/nvidia",
		},
	}

	_, err = deployer.Generate(ctx, recipeResult, tmpDir)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Read and verify the component HelmRelease file
	helmReleasePath := filepath.Join(tmpDir, "gpu-operator", "helmrelease.yaml")
	content, err := os.ReadFile(helmReleasePath)
	if err != nil {
		t.Fatalf("Failed to read HelmRelease file: %v", err)
	}

	// Verify HelmRelease YAML contains expected fields
	expectedStrings := []string{
		"apiVersion: helm.toolkit.fluxcd.io/v2",
		"kind: HelmRelease",
		"gpu-operator",
		"HelmRepository",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(string(content), expected) {
			t.Errorf("HelmRelease file does not contain expected string: %s", expected)
		}
	}
}

func TestInit_Registers(t *testing.T) {
	// Verify that the init() function registered the flux deployer
	r := registry.NewFromGlobal()

	deployer, found := r.Get(types.DeployerTypeFlux)
	if !found {
		t.Error("Flux deployer not found in global registry after init()")
	}

	if deployer == nil {
		t.Error("Flux deployer is nil in global registry")
	}
}

func TestDeployer_Generate_DeploymentOrder(t *testing.T) {
	deployer := &Deployer{}
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "flux-order-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create component directories
	for _, component := range []string{"cert-manager", "gpu-operator", "skyhook"} {
		componentDir := filepath.Join(tmpDir, component)
		if mkErr := os.MkdirAll(componentDir, 0755); mkErr != nil {
			t.Fatalf("Failed to create component dir %s: %v", component, mkErr)
		}
	}

	// Create recipe with components in reverse order but deployment order specified
	recipeResult := &recipe.RecipeResult{}
	recipeResult.Metadata.Version = testVersion
	recipeResult.ComponentRefs = []recipe.ComponentRef{
		{Name: "skyhook", Version: "v1.0.0", Source: "https://example.com"},
		{Name: "gpu-operator", Version: "v25.3.3", Source: "https://example.com"},
		{Name: "cert-manager", Version: "v1.14.0", Source: "https://example.com"},
	}
	// Deployment order: cert-manager first, then gpu-operator, then skyhook
	recipeResult.DeploymentOrder = []string{"cert-manager", "gpu-operator", "skyhook"}

	artifacts, err := deployer.Generate(ctx, recipeResult, tmpDir)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if !artifacts.Success {
		t.Fatalf("Generate() failed: %s", artifacts.Error)
	}

	// Verify dependsOn chain
	testCases := []struct {
		component        string
		expectsDependsOn bool
		dependsOnName    string
	}{
		{"cert-manager", false, ""},            // First component has no dependency
		{"gpu-operator", true, "cert-manager"}, // Depends on cert-manager
		{"skyhook", true, "gpu-operator"},      // Depends on gpu-operator
	}

	for _, tc := range testCases {
		helmReleasePath := filepath.Join(tmpDir, tc.component, "helmrelease.yaml")
		content, err := os.ReadFile(helmReleasePath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", helmReleasePath, err)
		}

		contentStr := string(content)

		if tc.expectsDependsOn {
			// Verify dependsOn is present with correct name
			if !strings.Contains(contentStr, "dependsOn:") {
				t.Errorf("HelmRelease %s should have dependsOn field", tc.component)
			}
			if !strings.Contains(contentStr, "name: "+tc.dependsOnName) {
				t.Errorf("HelmRelease %s should depend on %s\nContent:\n%s",
					tc.component, tc.dependsOnName, contentStr)
			}
		} else if strings.Contains(contentStr, "dependsOn:") {
			// First component should not have dependsOn
			t.Errorf("HelmRelease %s should NOT have dependsOn field (it's first)\nContent:\n%s",
				tc.component, contentStr)
		}
	}
}

func TestOrderComponentsByDeployment(t *testing.T) {
	tests := []struct {
		name       string
		components []recipe.ComponentRef
		order      []string
		wantOrder  []string
	}{
		{
			name: "orders by deployment order",
			components: []recipe.ComponentRef{
				{Name: "c"},
				{Name: "b"},
				{Name: "a"},
			},
			order:     []string{"a", "b", "c"},
			wantOrder: []string{"a", "b", "c"},
		},
		{
			name: "empty order returns original",
			components: []recipe.ComponentRef{
				{Name: "c"},
				{Name: "b"},
				{Name: "a"},
			},
			order:     []string{},
			wantOrder: []string{"c", "b", "a"},
		},
		{
			name: "components not in order go last",
			components: []recipe.ComponentRef{
				{Name: "unknown"},
				{Name: "b"},
				{Name: "a"},
			},
			order:     []string{"a", "b"},
			wantOrder: []string{"a", "b", "unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := orderComponentsByDeployment(tt.components, tt.order)

			if len(result) != len(tt.wantOrder) {
				t.Fatalf("got %d components, want %d", len(result), len(tt.wantOrder))
			}

			for i, comp := range result {
				if comp.Name != tt.wantOrder[i] {
					t.Errorf("position %d: got %s, want %s", i, comp.Name, tt.wantOrder[i])
				}
			}
		})
	}
}
