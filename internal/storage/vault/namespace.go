package vault

import "strings"

// hasPathPrefix reports whether s starts with prefix AND has at least
// one more segment after the prefix. The caller is responsible for
// ensuring prefix ends with "/".
func hasPathPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// JoinNamespacePath joins a parent path with a child segment, ensuring
// a single "/" separator.
func JoinNamespacePath(parent, child string) string {
	parent = strings.TrimRight(parent, "/")
	child = strings.TrimLeft(child, "/")
	if child == "" {
		return parent
	}
	if parent == "" {
		return child
	}
	return parent + "/" + child
}

// SplitNamespacePath returns the segments of a namespace path.
func SplitNamespacePath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

// IsValidNamespacePath reports whether p is a syntactically valid
// namespace path: lowercase letters, digits, "-", "_", and "/".
// Each segment must be non-empty. No leading or trailing slash.
func IsValidNamespacePath(p string) bool {
	if p == "" {
		return false
	}
	if p[0] == '/' || p[len(p)-1] == '/' {
		return false
	}
	for _, r := range p {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '/' || r == '-' || r == '_':
		default:
			return false
		}
	}
	// No empty segments (defensive — already rejected "/" as a single char).
	for _, seg := range SplitNamespacePath(p) {
		if seg == "" {
			return false
		}
	}
	return true
}
