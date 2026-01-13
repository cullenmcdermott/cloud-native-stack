// Package internal provides shared utilities for deployer implementations.
package internal

import (
	"bytes"
	"fmt"
	"text/template"
)

// RenderTemplate renders a template with the given data.
func RenderTemplate(tmplContent string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
