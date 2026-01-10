package recipe

import (
	"testing"
)

func TestMergeValues(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]interface{}
		overlay  map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "simple override",
			base: map[string]interface{}{
				"enabled": true,
				"version": "1.0.0",
			},
			overlay: map[string]interface{}{
				"version": "2.0.0",
			},
			expected: map[string]interface{}{
				"enabled": true,
				"version": "2.0.0",
			},
		},
		{
			name: "nested map merge",
			base: map[string]interface{}{
				"driver": map[string]interface{}{
					"enabled":    true,
					"repository": "nvcr.io/nvidia",
					"version":    "1.0.0",
				},
			},
			overlay: map[string]interface{}{
				"driver": map[string]interface{}{
					"version": "2.0.0",
				},
			},
			expected: map[string]interface{}{
				"driver": map[string]interface{}{
					"enabled":    true,
					"repository": "nvcr.io/nvidia",
					"version":    "2.0.0",
				},
			},
		},
		{
			name: "add new key",
			base: map[string]interface{}{
				"enabled": true,
			},
			overlay: map[string]interface{}{
				"newFeature": true,
			},
			expected: map[string]interface{}{
				"enabled":    true,
				"newFeature": true,
			},
		},
		{
			name: "deep nested merge",
			base: map[string]interface{}{
				"driver": map[string]interface{}{
					"config": map[string]interface{}{
						"timeout": 30,
						"retry":   3,
					},
				},
			},
			overlay: map[string]interface{}{
				"driver": map[string]interface{}{
					"config": map[string]interface{}{
						"timeout": 60,
					},
				},
			},
			expected: map[string]interface{}{
				"driver": map[string]interface{}{
					"config": map[string]interface{}{
						"timeout": 60,
						"retry":   3,
					},
				},
			},
		},
		{
			name: "type mismatch - overlay wins",
			base: map[string]interface{}{
				"value": map[string]interface{}{
					"nested": "data",
				},
			},
			overlay: map[string]interface{}{
				"value": "string",
			},
			expected: map[string]interface{}{
				"value": "string",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of base to avoid modifying the test data
			dst := make(map[string]interface{})
			for k, v := range tt.base {
				dst[k] = v
			}

			// Merge overlay into dst
			mergeValues(dst, tt.overlay)

			// Compare results
			if !mapsEqual(dst, tt.expected) {
				t.Errorf("mergeValues() result mismatch\ngot:  %+v\nwant: %+v", dst, tt.expected)
			}
		})
	}
}

// mapsEqual compares two maps recursively.
func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, aVal := range a {
		bVal, exists := b[key]
		if !exists {
			return false
		}

		// If both are maps, compare recursively
		if aMap, aOK := aVal.(map[string]interface{}); aOK {
			if bMap, bOK := bVal.(map[string]interface{}); bOK {
				if !mapsEqual(aMap, bMap) {
					return false
				}
				continue
			}
		}

		// For non-map types, use direct comparison
		if aVal != bVal {
			return false
		}
	}

	return true
}
