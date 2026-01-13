package script

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

	// Create test directory
	tmpDir, err := os.MkdirTemp("", "script-deployer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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

	// Verify README was created
	readmePath := filepath.Join(tmpDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("Generate() did not create README.md")
	}

	// Verify README content
	if artifacts.ReadmeContent == "" {
		t.Error("Generate() did not set ReadmeContent")
	}

	// Verify README contains expected content
	expectedContent := []string{
		"gpu-operator",
		"network-operator",
		"v25.3.3",
		"v25.4.0",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(artifacts.ReadmeContent, expected) {
			t.Errorf("README does not contain expected content: %s", expected)
		}
	}

	// Verify Files includes the README
	found := false
	for _, f := range artifacts.Files {
		if strings.HasSuffix(f, "README.md") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Generate() did not include README.md in Files")
	}

	// Verify Duration is set
	if artifacts.Duration == 0 {
		t.Error("Generate() did not set Duration")
	}
}

func TestDeployer_Generate_EmptyComponents(t *testing.T) {
	deployer := &Deployer{}
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "script-deployer-test-*")
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

func TestInternalReadmeData_Fields(t *testing.T) {
	data := internal.ReadmeData{
		Timestamp:     time.Now().Format(time.RFC3339),
		RecipeVersion: "v1.0.0",
		Components: []internal.ComponentInfo{
			{Name: "test-component", Version: "v1.2.3"},
		},
	}

	if data.Timestamp == "" {
		t.Error("ReadmeData.Timestamp is empty")
	}

	if data.RecipeVersion != "v1.0.0" {
		t.Errorf("ReadmeData.RecipeVersion = %s, want v1.0.0", data.RecipeVersion)
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
			name:     "invalid template",
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
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("RenderTemplate() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestInit_Registers(t *testing.T) {
	// Verify that the init() function registered the script deployer
	// This test relies on the global registry from the registry package
	r := registry.NewFromGlobal()

	deployer, found := r.Get(types.DeployerTypeScript)
	if !found {
		t.Error("Script deployer not found in global registry after init()")
	}

	if deployer == nil {
		t.Error("Script deployer is nil in global registry")
	}
}

func TestDeployer_Generate_DeploymentOrder(t *testing.T) {
	deployer := &Deployer{}
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "script-order-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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

	// Verify README lists components in deployment order
	readme := artifacts.ReadmeContent

	// Find positions of each component in the README
	certManagerPos := strings.Index(readme, "cert-manager")
	gpuOperatorPos := strings.Index(readme, "gpu-operator")
	skyhookPos := strings.Index(readme, "skyhook")

	if certManagerPos == -1 {
		t.Fatal("README does not contain cert-manager")
	}
	if gpuOperatorPos == -1 {
		t.Fatal("README does not contain gpu-operator")
	}
	if skyhookPos == -1 {
		t.Fatal("README does not contain skyhook")
	}

	// Verify order: cert-manager should appear before gpu-operator before skyhook
	if certManagerPos > gpuOperatorPos {
		t.Errorf("cert-manager (pos %d) should appear before gpu-operator (pos %d) in README",
			certManagerPos, gpuOperatorPos)
	}
	if gpuOperatorPos > skyhookPos {
		t.Errorf("gpu-operator (pos %d) should appear before skyhook (pos %d) in README",
			gpuOperatorPos, skyhookPos)
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
