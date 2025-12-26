package os

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/NVIDIA/cloud-native-stack/pkg/measurement"
)

// collectRelease gathers OS release information from /etc/os-release.
// It returns a measurement subtype containing key-value pairs of release data.
// Example keys include NAME, VERSION, ID, PRETTY_NAME, etc.
// Per freedesktop.org spec, falls back to /usr/lib/os-release if primary file doesn't exist.
//
//	NAME="Ubuntu"
//	ID=ubuntu
//	VERSION_ID="22.04"
//	PRETTY_NAME="Ubuntu 22.04.4 LTS"
func (c *Collector) collectRelease(ctx context.Context) (*measurement.Subtype, error) {
	// Check if context is canceled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Try primary location first, fall back to alternative per freedesktop.org spec
	root := "/etc/os-release"
	if _, err := os.Stat(root); os.IsNotExist(err) {
		root = "/usr/lib/os-release"
	}

	content, err := os.ReadFile(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read os release from %s: %w", root, err)
	}

	// Pre-allocate with typical capacity (most files have 10-15 fields)
	readings := make(map[string]measurement.Reading, 15)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments per freedesktop.org spec
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Release entries are in KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := strings.Trim(parts[1], `"'`) // Remove surrounding quotes if any
		readings[key] = measurement.Str(value)
	}

	res := &measurement.Subtype{
		Name: "release",
		Data: readings,
	}

	return res, nil
}
