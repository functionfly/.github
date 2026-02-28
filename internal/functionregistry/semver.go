package functionregistry

import (
	"fmt"
	"strconv"
	"strings"
)

// SemVer represents a parsed semantic version
type SemVer struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Raw        string
}

// ParseSemVer parses a semantic version string (supports v-prefix and pre-release)
func ParseSemVer(version string) (*SemVer, error) {
	if version == "" {
		return nil, fmt.Errorf("version string is empty")
	}
	v := version
	if len(v) > 0 && v[0] == 'v' {
		v = v[1:]
	}

	var preRelease string
	if idx := strings.IndexByte(v, '+'); idx >= 0 {
		v = v[:idx] // strip build metadata
	}
	if idx := strings.IndexByte(v, '-'); idx >= 0 {
		preRelease = v[idx+1:]
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid semver format '%s': expected MAJOR.MINOR.PATCH", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return nil, fmt.Errorf("invalid major version '%s' in '%s'", parts[0], version)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil || minor < 0 {
		return nil, fmt.Errorf("invalid minor version '%s' in '%s'", parts[1], version)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil || patch < 0 {
		return nil, fmt.Errorf("invalid patch version '%s' in '%s'", parts[2], version)
	}

	return &SemVer{Major: major, Minor: minor, Patch: patch, PreRelease: preRelease, Raw: version}, nil
}

// ValidateSemVer validates that a version string is valid semver
func ValidateSemVer(version string) error {
	_, err := ParseSemVer(version)
	if err != nil {
		return fmt.Errorf("invalid version '%s': %w. Use semantic versioning (e.g., 1.0.0, 2.1.3-beta.1)", version, err)
	}
	return nil
}

// Compare compares two SemVer values. Returns -1, 0, or 1.
func (a *SemVer) Compare(b *SemVer) int {
	if a.Major != b.Major {
		return cmpInt(a.Major, b.Major)
	}
	if a.Minor != b.Minor {
		return cmpInt(a.Minor, b.Minor)
	}
	if a.Patch != b.Patch {
		return cmpInt(a.Patch, b.Patch)
	}
	if a.PreRelease == "" && b.PreRelease != "" {
		return 1
	}
	if a.PreRelease != "" && b.PreRelease == "" {
		return -1
	}
	return strings.Compare(a.PreRelease, b.PreRelease)
}

// IsGreaterThan returns true if a > b
func (a *SemVer) IsGreaterThan(b *SemVer) bool { return a.Compare(b) > 0 }

// FindLatestVersion finds the highest semver from a list of version strings
func FindLatestVersion(versions []string) string {
	var latest *SemVer
	var latestStr string
	for _, v := range versions {
		parsed, err := ParseSemVer(v)
		if err != nil {
			continue
		}
		if latest == nil || parsed.IsGreaterThan(latest) {
			latest = parsed
			latestStr = v
		}
	}
	return latestStr
}

func cmpInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
