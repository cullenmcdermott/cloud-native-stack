package collectors

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelmCollector_Collect(t *testing.T) {
	ctx := context.Background()
	collector := &HelmCollector{}

	// This test requires a kubernetes cluster to be available
	// It will gracefully handle both no cluster and no helm releases
	measurements, err := collector.Collect(ctx)

	// If no cluster is available, we expect an error
	if err != nil {
		// Accept errors related to kubernetes connectivity
		errMsg := err.Error()
		validError := strings.Contains(errMsg, "failed to get kubernetes client") ||
			strings.Contains(errMsg, "failed to list helm secrets")
		assert.True(t, validError, "unexpected error: %v", err)
		t.Logf("Helm collector failed as expected (no cluster available): %v", err)
		return
	}

	// If successful, validate the measurement
	assert.NoError(t, err)
	assert.Len(t, measurements, 1)
	assert.Equal(t, HelmType, measurements[0].Type)
	assert.NotNil(t, measurements[0].Data)

	// Data should be a map
	data, ok := measurements[0].Data.(map[string]any)
	assert.True(t, ok, "Data should be a map[string]any")

	// Log the number of releases found (may be 0 if no helm releases in cluster)
	t.Logf("Found %d Helm releases", len(data))
}

func TestHelmCollector_CollectWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	collector := &HelmCollector{}
	_, err := collector.Collect(ctx)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
