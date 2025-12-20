package collectors

import (
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/NVIDIA/cloud-native-stack/pkg/client"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelmCollector collects information about Helm releases in the cluster.
type HelmCollector struct {
}

// HelmType is the type identifier for Helm measurements.
const HelmType string = "Helm"

// Collect retrieves all Helm releases across all namespaces.
// This provides a reliable snapshot of installed charts for cluster comparison.
// Helm 3 stores release information as Secrets in the cluster with label owner=helm.
func (h *HelmCollector) Collect(ctx context.Context) ([]Measurement, error) {
	// Check if context is canceled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	releases, err := h.collectHelmReleases(ctx)
	if err != nil {
		return nil, err
	}

	res := []Measurement{
		{
			Type: HelmType,
			Data: releases,
		},
	}

	return res, nil
}

// helmRelease represents the structure of Helm release data stored in Kubernetes.
type helmRelease struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Info      struct {
		Status string `json:"status"`
	} `json:"info"`
	Chart struct {
		Metadata struct {
			Name       string `json:"name"`
			Version    string `json:"version"`
			AppVersion string `json:"appVersion"`
		} `json:"metadata"`
	} `json:"chart"`
}

// collectHelmReleases retrieves all Helm releases by querying Kubernetes secrets.
// Helm 3 stores release information as secrets with label "owner=helm".
func (h *HelmCollector) collectHelmReleases(ctx context.Context) (map[string]any, error) {
	k8sClient, _, err := client.GetKubeClient("")
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes client: %w", err)
	}

	// Query all secrets across all namespaces with label owner=helm
	secrets, err := k8sClient.CoreV1().Secrets("").List(ctx, v1.ListOptions{
		LabelSelector: "owner=helm",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list helm secrets: %w", err)
	}

	result := make(map[string]any)
	for _, secret := range secrets.Items {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Only process deployed releases (skip pending-*, superseded, etc.)
		status := secret.Labels["status"]
		if status != "deployed" {
			continue
		}

		releaseName := secret.Labels["name"]
		if releaseName == "" {
			continue
		}

		// Parse the release data from the secret
		release, err := h.parseHelmSecret(&secret)
		if err != nil {
			slog.Warn("failed to parse helm secret",
				slog.String("namespace", secret.Namespace),
				slog.String("name", releaseName),
				slog.String("error", err.Error()))
			continue
		}

		key := fmt.Sprintf("%s/%s", secret.Namespace, releaseName)
		result[key] = map[string]string{
			"chart":      fmt.Sprintf("%s-%s", release.Chart.Metadata.Name, release.Chart.Metadata.Version),
			"appVersion": release.Chart.Metadata.AppVersion,
			"status":     release.Info.Status,
		}
	}

	slog.Debug("collected helm releases", slog.Int("count", len(result)))
	return result, nil
}

// parseHelmSecret decodes and decompresses Helm release data from a Kubernetes secret.
func (h *HelmCollector) parseHelmSecret(secret *corev1.Secret) (*helmRelease, error) {
	// Helm stores release data in the "release" key
	releaseData, ok := secret.Data["release"]
	if !ok {
		return nil, fmt.Errorf("release data not found in secret")
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(string(releaseData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Decompress gzip
	gzReader, err := gzip.NewReader(strings.NewReader(string(decoded)))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}

	// Parse JSON
	var release helmRelease
	if err := json.Unmarshal(decompressed, &release); err != nil {
		return nil, fmt.Errorf("failed to parse helm release JSON: %w", err)
	}

	return &release, nil
}
