package config

import "testing"

func TestConfigImmutability(t *testing.T) {
	cfg := NewConfig()

	// Verify we can only read values, not modify them
	namespace := cfg.Namespace()
	if namespace != "default" {
		t.Errorf("Namespace() = %s, want default", namespace)
	}

	outputFormat := cfg.OutputFormat()
	if outputFormat != "yaml" {
		t.Errorf("OutputFormat() = %s, want yaml", outputFormat)
	}

	// Verify getters return expected default values
	if !cfg.IncludeScripts() {
		t.Error("IncludeScripts() = false, want true")
	}

	if !cfg.IncludeReadme() {
		t.Error("IncludeReadme() = false, want true")
	}

	if !cfg.IncludeChecksums() {
		t.Error("IncludeChecksums() = false, want true")
	}

	if cfg.Compression() {
		t.Error("Compression() = true, want false")
	}

	if cfg.Verbose() {
		t.Error("Verbose() = true, want false")
	}
}

func TestConfigMapGetters(t *testing.T) {
	cfg := NewConfig()

	// Verify map getters return defensive copies
	labels := cfg.CustomLabels()
	if labels == nil {
		t.Fatal("CustomLabels() returned nil")
	}

	// Modify the returned map
	labels["test"] = "value"

	// Verify the original config was not affected
	newLabels := cfg.CustomLabels()
	if _, exists := newLabels["test"]; exists {
		t.Error("modifying returned map affected original config - not properly immutable")
	}

	// Same test for annotations
	annotations := cfg.CustomAnnotations()
	if annotations == nil {
		t.Fatal("CustomAnnotations() returned nil")
	}

	annotations["test"] = "value"

	newAnnotations := cfg.CustomAnnotations()
	if _, exists := newAnnotations["test"]; exists {
		t.Error("modifying returned map affected original config - not properly immutable")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  NewConfig(),
			wantErr: false,
		},
		{
			name:    "valid yaml format",
			config:  NewConfig(WithOutputFormat("yaml")),
			wantErr: false,
		},
		{
			name:    "valid json format",
			config:  NewConfig(WithOutputFormat("json")),
			wantErr: false,
		},
		{
			name:    "valid helm format",
			config:  NewConfig(WithOutputFormat("helm")),
			wantErr: false,
		},
		{
			name: "invalid output format",
			config: &Config{
				outputFormat: "invalid",
				namespace:    "default",
			},
			wantErr: true,
		},
		{
			name: "empty namespace",
			config: &Config{
				outputFormat: "yaml",
				namespace:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConfigWithOptions(t *testing.T) {
	cfg := NewConfig(
		WithOutputFormat("json"),
		WithCompression(true),
		WithIncludeScripts(false),
		WithIncludeReadme(false),
		WithIncludeChecksums(false),
		WithHelmChartVersion("v1.2.3"),
		WithHelmRepository("https://example.com/helm"),
		WithNamespace("custom-ns"),
		WithVerbose(true),
		WithCustomLabels(map[string]string{"env": "prod", "team": "platform"}),
		WithCustomAnnotations(map[string]string{"owner": "john", "project": "alpha"}),
	)

	// Verify all options were applied
	if cfg.OutputFormat() != "json" {
		t.Errorf("OutputFormat() = %s, want json", cfg.OutputFormat())
	}
	if !cfg.Compression() {
		t.Error("Compression() = false, want true")
	}
	if cfg.IncludeScripts() {
		t.Error("IncludeScripts() = true, want false")
	}
	if cfg.IncludeReadme() {
		t.Error("IncludeReadme() = true, want false")
	}
	if cfg.IncludeChecksums() {
		t.Error("IncludeChecksums() = true, want false")
	}
	if cfg.HelmChartVersion() != "v1.2.3" {
		t.Errorf("HelmChartVersion() = %s, want v1.2.3", cfg.HelmChartVersion())
	}
	if cfg.HelmRepository() != "https://example.com/helm" {
		t.Errorf("HelmRepository() = %s, want https://example.com/helm", cfg.HelmRepository())
	}
	if cfg.Namespace() != "custom-ns" {
		t.Errorf("Namespace() = %s, want custom-ns", cfg.Namespace())
	}
	if !cfg.Verbose() {
		t.Error("Verbose() = false, want true")
	}

	// Verify custom labels
	labels := cfg.CustomLabels()
	if labels["env"] != "prod" {
		t.Errorf("CustomLabels()[env] = %s, want prod", labels["env"])
	}
	if labels["team"] != "platform" {
		t.Errorf("CustomLabels()[team] = %s, want platform", labels["team"])
	}

	// Verify custom annotations
	annotations := cfg.CustomAnnotations()
	if annotations["owner"] != "john" {
		t.Errorf("CustomAnnotations()[owner] = %s, want john", annotations["owner"])
	}
	if annotations["project"] != "alpha" {
		t.Errorf("CustomAnnotations()[project] = %s, want alpha", annotations["project"])
	}
}

func TestAllGetters(t *testing.T) {
	cfg := NewConfig(
		WithOutputFormat("helm"),
		WithCompression(true),
		WithIncludeScripts(false),
		WithIncludeReadme(true),
		WithIncludeChecksums(false),
		WithHelmChartVersion("v2.0.0"),
		WithHelmRepository("https://charts.example.com"),
		WithNamespace("production"),
		WithVerbose(true),
	)

	tests := []struct {
		name     string
		got      interface{}
		want     interface{}
		getterFn string
	}{
		{"OutputFormat", cfg.OutputFormat(), "helm", "OutputFormat()"},
		{"Compression", cfg.Compression(), true, "Compression()"},
		{"IncludeScripts", cfg.IncludeScripts(), false, "IncludeScripts()"},
		{"IncludeReadme", cfg.IncludeReadme(), true, "IncludeReadme()"},
		{"IncludeChecksums", cfg.IncludeChecksums(), false, "IncludeChecksums()"},
		{"HelmChartVersion", cfg.HelmChartVersion(), "v2.0.0", "HelmChartVersion()"},
		{"HelmRepository", cfg.HelmRepository(), "https://charts.example.com", "HelmRepository()"},
		{"Namespace", cfg.Namespace(), "production", "Namespace()"},
		{"Verbose", cfg.Verbose(), true, "Verbose()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.getterFn, tt.got, tt.want)
			}
		})
	}
}

func TestCustomLabelsImmutability(t *testing.T) {
	// Test that custom labels added via option are immutable
	cfg := NewConfig(
		WithCustomLabels(map[string]string{"key1": "value1"}),
	)

	// Get labels and try to modify
	labels := cfg.CustomLabels()
	labels["key2"] = "value2"
	delete(labels, "key1")

	// Verify original config unchanged
	newLabels := cfg.CustomLabels()
	if _, exists := newLabels["key2"]; exists {
		t.Error("adding key to returned map affected config - not immutable")
	}
	if _, exists := newLabels["key1"]; !exists {
		t.Error("deleting key from returned map affected config - not immutable")
	}
	if newLabels["key1"] != "value1" {
		t.Errorf("CustomLabels()[key1] = %s, want value1", newLabels["key1"])
	}
}

func TestCustomAnnotationsImmutability(t *testing.T) {
	// Test that custom annotations added via option are immutable
	cfg := NewConfig(
		WithCustomAnnotations(map[string]string{"ann1": "val1"}),
	)

	// Get annotations and try to modify
	annotations := cfg.CustomAnnotations()
	annotations["ann2"] = "val2"
	delete(annotations, "ann1")

	// Verify original config unchanged
	newAnnotations := cfg.CustomAnnotations()
	if _, exists := newAnnotations["ann2"]; exists {
		t.Error("adding key to returned map affected config - not immutable")
	}
	if _, exists := newAnnotations["ann1"]; !exists {
		t.Error("deleting key from returned map affected config - not immutable")
	}
	if newAnnotations["ann1"] != "val1" {
		t.Errorf("CustomAnnotations()[ann1] = %s, want val1", newAnnotations["ann1"])
	}
}

func TestMultipleCustomLabelsOptions(t *testing.T) {
	// Test that multiple WithCustomLabels calls merge correctly
	cfg := NewConfig(
		WithCustomLabels(map[string]string{"key1": "value1"}),
		WithCustomLabels(map[string]string{"key2": "value2"}),
		WithCustomLabels(map[string]string{"key1": "updated"}), // Should override
	)

	labels := cfg.CustomLabels()
	if labels["key1"] != "updated" {
		t.Errorf("CustomLabels()[key1] = %s, want updated (should be overridden)", labels["key1"])
	}
	if labels["key2"] != "value2" {
		t.Errorf("CustomLabels()[key2] = %s, want value2", labels["key2"])
	}
}

func TestMultipleCustomAnnotationsOptions(t *testing.T) {
	// Test that multiple WithCustomAnnotations calls merge correctly
	cfg := NewConfig(
		WithCustomAnnotations(map[string]string{"ann1": "val1"}),
		WithCustomAnnotations(map[string]string{"ann2": "val2"}),
		WithCustomAnnotations(map[string]string{"ann1": "updated"}), // Should override
	)

	annotations := cfg.CustomAnnotations()
	if annotations["ann1"] != "updated" {
		t.Errorf("CustomAnnotations()[ann1] = %s, want updated (should be overridden)", annotations["ann1"])
	}
	if annotations["ann2"] != "val2" {
		t.Errorf("CustomAnnotations()[ann2] = %s, want val2", annotations["ann2"])
	}
}

func TestEmptyStringGetters(t *testing.T) {
	// Test that string getters return empty strings for unset values
	cfg := NewConfig()

	if cfg.HelmChartVersion() != "" {
		t.Errorf("HelmChartVersion() = %s, want empty string", cfg.HelmChartVersion())
	}
	if cfg.HelmRepository() != "" {
		t.Errorf("HelmRepository() = %s, want empty string", cfg.HelmRepository())
	}
}

func TestValidateErrorMessages(t *testing.T) {
	tests := []struct {
		name            string
		config          *Config
		wantErrContains string
	}{
		{
			name: "invalid format error message",
			config: &Config{
				outputFormat: "xml",
				namespace:    "default",
			},
			wantErrContains: "invalid output format: xml",
		},
		{
			name: "empty namespace error message",
			config: &Config{
				outputFormat: "yaml",
				namespace:    "",
			},
			wantErrContains: "namespace cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
			if tt.wantErrContains != "" {
				errMsg := err.Error()
				if len(errMsg) < len(tt.wantErrContains) || errMsg[:len(tt.wantErrContains)] != tt.wantErrContains {
					t.Errorf("Validate() error = %q, want error containing %q", errMsg, tt.wantErrContains)
				}
			}
		})
	}
}
