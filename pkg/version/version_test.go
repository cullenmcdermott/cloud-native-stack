package version

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      Version
		expectedError bool
	}{
		{
			name:  "major only",
			input: "1",
			expected: Version{
				Major:     1,
				Minor:     0,
				Patch:     0,
				Precision: 1,
			},
			expectedError: false,
		},
		{
			name:  "major only with v prefix",
			input: "v2",
			expected: Version{
				Major:     2,
				Minor:     0,
				Patch:     0,
				Precision: 1,
			},
			expectedError: false,
		},
		{
			name:  "major.minor",
			input: "1.2",
			expected: Version{
				Major:     1,
				Minor:     2,
				Patch:     0,
				Precision: 2,
			},
			expectedError: false,
		},
		{
			name:  "major.minor with v prefix",
			input: "v0.1",
			expected: Version{
				Major:     0,
				Minor:     1,
				Patch:     0,
				Precision: 2,
			},
			expectedError: false,
		},
		{
			name:  "full version",
			input: "1.2.3",
			expected: Version{
				Major:     1,
				Minor:     2,
				Patch:     3,
				Precision: 3,
			},
			expectedError: false,
		},
		{
			name:  "full version with v prefix",
			input: "v1.2.3",
			expected: Version{
				Major:     1,
				Minor:     2,
				Patch:     3,
				Precision: 3,
			},
			expectedError: false,
		},
		{
			name:  "version with zeros",
			input: "v0.0.0",
			expected: Version{
				Major:     0,
				Minor:     0,
				Patch:     0,
				Precision: 3,
			},
			expectedError: false,
		},
		{
			name:          "invalid - too many components",
			input:         "1.2.3.4",
			expected:      Version{},
			expectedError: true,
		},
		{
			name:          "invalid - non-numeric",
			input:         "v1.2.a",
			expected:      Version{},
			expectedError: true,
		},
		{
			name:          "invalid - empty string",
			input:         "",
			expected:      Version{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVersion(tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result.Major != tt.expected.Major {
				t.Errorf("Major: got %d, want %d", result.Major, tt.expected.Major)
			}
			if result.Minor != tt.expected.Minor {
				t.Errorf("Minor: got %d, want %d", result.Minor, tt.expected.Minor)
			}
			if result.Patch != tt.expected.Patch {
				t.Errorf("Patch: got %d, want %d", result.Patch, tt.expected.Patch)
			}
			if result.Precision != tt.expected.Precision {
				t.Errorf("Precision: got %d, want %d", result.Precision, tt.expected.Precision)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		expected string
	}{
		{
			name: "major only",
			version: Version{
				Major:     1,
				Minor:     0,
				Patch:     0,
				Precision: 1,
			},
			expected: "1",
		},
		{
			name: "major.minor",
			version: Version{
				Major:     1,
				Minor:     2,
				Patch:     0,
				Precision: 2,
			},
			expected: "1.2",
		},
		{
			name: "full version",
			version: Version{
				Major:     1,
				Minor:     2,
				Patch:     3,
				Precision: 3,
			},
			expected: "1.2.3",
		},
		{
			name: "zero version with precision 2",
			version: Version{
				Major:     0,
				Minor:     1,
				Patch:     5,
				Precision: 2,
			},
			expected: "0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestEqualsOrNewer(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		other    Version
		expected bool
	}{
		{
			name: "major only - equal",
			version: Version{
				Major:     1,
				Minor:     0,
				Patch:     0,
				Precision: 1,
			},
			other: Version{
				Major:     1,
				Minor:     5,
				Patch:     10,
				Precision: 3,
			},
			expected: true,
		},
		{
			name: "major only - newer",
			version: Version{
				Major:     2,
				Minor:     0,
				Patch:     0,
				Precision: 1,
			},
			other: Version{
				Major:     1,
				Minor:     9,
				Patch:     9,
				Precision: 3,
			},
			expected: true,
		},
		{
			name: "major only - older",
			version: Version{
				Major:     1,
				Minor:     0,
				Patch:     0,
				Precision: 1,
			},
			other: Version{
				Major:     2,
				Minor:     0,
				Patch:     0,
				Precision: 3,
			},
			expected: false,
		},
		{
			name: "major.minor - equal (example from user: v0.1 matches 0.1.1)",
			version: Version{
				Major:     0,
				Minor:     1,
				Patch:     0,
				Precision: 2,
			},
			other: Version{
				Major:     0,
				Minor:     1,
				Patch:     1,
				Precision: 3,
			},
			expected: true,
		},
		{
			name: "major.minor - newer minor",
			version: Version{
				Major:     1,
				Minor:     3,
				Patch:     0,
				Precision: 2,
			},
			other: Version{
				Major:     1,
				Minor:     2,
				Patch:     99,
				Precision: 3,
			},
			expected: true,
		},
		{
			name: "major.minor - older minor",
			version: Version{
				Major:     1,
				Minor:     1,
				Patch:     0,
				Precision: 2,
			},
			other: Version{
				Major:     1,
				Minor:     2,
				Patch:     0,
				Precision: 3,
			},
			expected: false,
		},
		{
			name: "full version - equal",
			version: Version{
				Major:     1,
				Minor:     2,
				Patch:     3,
				Precision: 3,
			},
			other: Version{
				Major:     1,
				Minor:     2,
				Patch:     3,
				Precision: 3,
			},
			expected: true,
		},
		{
			name: "full version - newer patch",
			version: Version{
				Major:     1,
				Minor:     2,
				Patch:     4,
				Precision: 3,
			},
			other: Version{
				Major:     1,
				Minor:     2,
				Patch:     3,
				Precision: 3,
			},
			expected: true,
		},
		{
			name: "full version - older patch",
			version: Version{
				Major:     1,
				Minor:     2,
				Patch:     2,
				Precision: 3,
			},
			other: Version{
				Major:     1,
				Minor:     2,
				Patch:     3,
				Precision: 3,
			},
			expected: false,
		},
		{
			name: "full version - newer major",
			version: Version{
				Major:     2,
				Minor:     0,
				Patch:     0,
				Precision: 3,
			},
			other: Version{
				Major:     1,
				Minor:     9,
				Patch:     9,
				Precision: 3,
			},
			expected: true,
		},
		{
			name: "full version - newer minor",
			version: Version{
				Major:     1,
				Minor:     3,
				Patch:     0,
				Precision: 3,
			},
			other: Version{
				Major:     1,
				Minor:     2,
				Patch:     99,
				Precision: 3,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.EqualsOrNewer(tt.other)
			if result != tt.expected {
				t.Errorf("got %v, want %v (comparing %s vs %s)", result, tt.expected, tt.version.String(), tt.other.String())
			}
		})
	}
}

func TestParseVersionRoundTrip(t *testing.T) {
	tests := []string{
		"1",
		"v2",
		"1.2",
		"v0.1",
		"1.2.3",
		"v1.2.3",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			v, err := ParseVersion(input)
			if err != nil {
				t.Fatalf("ParseVersion failed: %v", err)
			}
			// Parse again from the string representation
			v2, err := ParseVersion(v.String())
			if err != nil {
				t.Fatalf("ParseVersion round-trip failed: %v", err)
			}
			if v.Major != v2.Major || v.Minor != v2.Minor || v.Patch != v2.Patch || v.Precision != v2.Precision {
				t.Errorf("round-trip mismatch: %+v != %+v", v, v2)
			}
		})
	}
}
