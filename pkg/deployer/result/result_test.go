package result

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		deployerType string
	}{
		{"script type", "script"},
		{"argocd type", "argocd"},
		{"flux type", "flux"},
		{"empty type", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.deployerType)

			if got == nil {
				t.Fatal("New() returned nil")
			}

			if got.Type != tt.deployerType {
				t.Errorf("Type = %v, want %v", got.Type, tt.deployerType)
			}

			if got.Success != true {
				t.Errorf("Success = %v, want true", got.Success)
			}

			if got.Files == nil {
				t.Error("Files is nil, want empty slice")
			}

			if len(got.Files) != 0 {
				t.Errorf("Files has %d elements, want 0", len(got.Files))
			}

			if got.ReadmeContent != "" {
				t.Errorf("ReadmeContent = %q, want empty string", got.ReadmeContent)
			}

			if got.Duration != 0 {
				t.Errorf("Duration = %v, want 0", got.Duration)
			}

			if got.Error != "" {
				t.Errorf("Error = %q, want empty string", got.Error)
			}
		})
	}
}

func TestArtifacts_Fields(t *testing.T) {
	a := &Artifacts{
		Type:          "test",
		Files:         []string{"file1.yaml", "file2.yaml"},
		ReadmeContent: "Test README",
		Duration:      5 * time.Second,
		Success:       false,
		Error:         "test error",
	}

	if a.Type != "test" {
		t.Errorf("Type = %v, want test", a.Type)
	}

	if len(a.Files) != 2 {
		t.Errorf("Files length = %d, want 2", len(a.Files))
	}

	if a.ReadmeContent != "Test README" {
		t.Errorf("ReadmeContent = %v, want 'Test README'", a.ReadmeContent)
	}

	if a.Duration != 5*time.Second {
		t.Errorf("Duration = %v, want 5s", a.Duration)
	}

	if a.Success != false {
		t.Errorf("Success = %v, want false", a.Success)
	}

	if a.Error != "test error" {
		t.Errorf("Error = %v, want 'test error'", a.Error)
	}
}

func TestArtifacts_AppendFiles(t *testing.T) {
	a := New("test")

	a.Files = append(a.Files, "file1.yaml")
	if len(a.Files) != 1 {
		t.Errorf("Files length = %d after append, want 1", len(a.Files))
	}

	a.Files = append(a.Files, "file2.yaml", "file3.yaml")
	if len(a.Files) != 3 {
		t.Errorf("Files length = %d after second append, want 3", len(a.Files))
	}
}
