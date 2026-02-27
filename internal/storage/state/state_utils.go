package state

import "strings"

// ParseStatePath parses a state path into tenant and name components
// Path format: tenant/name or just name (uses context tenant)
func ParseStatePath(path string) (tenant string, name string, err error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return "", parts[0], nil
	}
	return parts[0], parts[1], nil
}