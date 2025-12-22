package version

import (
	"fmt"
	"strings"
)

// Version represents a semantic version with Major, Minor, and Patch components.
// Precision indicates how many components are significant (1=Major, 2=Major.Minor, 3=Major.Minor.Patch).
type Version struct {
	Major     int
	Minor     int
	Patch     int
	Precision int // Number of version components specified (1, 2, or 3)
}

// String returns the version as a string respecting its precision
func (v Version) String() string {
	switch v.Precision {
	case 1:
		return fmt.Sprintf("%d", v.Major)
	case 2:
		return fmt.Sprintf("%d.%d", v.Major, v.Minor)
	default:
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
}

// ParseVersion parses a version string in the format "Major", "Major.Minor", "Major.Minor.Patch", or with "v" prefix
func ParseVersion(s string) (Version, error) {
	// Strip 'v' prefix if present
	s = strings.TrimPrefix(s, "v")
	var v Version

	// Count dots to determine precision
	dots := strings.Count(s, ".")

	switch dots {
	case 0:
		// Major only
		n, err := fmt.Sscanf(s, "%d", &v.Major)
		if n != 1 || err != nil {
			return Version{}, fmt.Errorf("invalid version format: %w", err)
		}
		v.Precision = 1
	case 1:
		// Major.Minor
		n, err := fmt.Sscanf(s, "%d.%d", &v.Major, &v.Minor)
		if n != 2 || err != nil {
			return Version{}, fmt.Errorf("invalid version format: %w", err)
		}
		v.Precision = 2
	case 2:
		// Major.Minor.Patch
		n, err := fmt.Sscanf(s, "%d.%d.%d", &v.Major, &v.Minor, &v.Patch)
		if n != 3 || err != nil {
			return Version{}, fmt.Errorf("invalid version format: %w", err)
		}
		v.Precision = 3
	default:
		return Version{}, fmt.Errorf("invalid version format: too many components")
	}

	return v, nil
}

// EqualsOrNewer returns true if v is equal to or newer than other.
// Only compares components up to the precision of v (e.g., v0.1 matches v0.1.x)
func (v Version) EqualsOrNewer(other Version) bool {
	// Always compare Major
	if v.Major > other.Major {
		return true
	}
	if v.Major < other.Major {
		return false
	}

	// If precision is 1 (Major only), we're equal
	if v.Precision == 1 {
		return true
	}

	// Major versions are equal, compare Minor
	if v.Minor > other.Minor {
		return true
	}
	if v.Minor < other.Minor {
		return false
	}

	// If precision is 2 (Major.Minor), we're equal
	if v.Precision == 2 {
		return true
	}

	// Minor versions are equal, compare Patch
	return v.Patch >= other.Patch
}
