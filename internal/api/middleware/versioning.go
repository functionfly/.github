package middleware

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultAPIVersion is the default API version when no version is specified
	DefaultAPIVersion = "v1"
)

// APIVersion represents an API version with its metadata
type APIVersion struct {
	Version          string     `json:"version"`
	IsActive         bool       `json:"is_active"`
	IsDefault        bool       `json:"is_default"`
	DeprecationDate  *time.Time `json:"deprecation_date,omitempty"`
	SunsetDate       *time.Time `json:"sunset_date,omitempty"`
	SuccessorVersion string     `json:"successor_version,omitempty"`
}

// VersionManager manages API versions
type VersionManager struct {
	versions       map[string]*APIVersion
	defaultVersion string
}

// NewVersionManager creates a new version manager
func NewVersionManager() *VersionManager {
	vm := &VersionManager{
		versions: make(map[string]*APIVersion),
	}

	// Register v1 - initial version
	vm.versions["v1"] = &APIVersion{
		Version:   "v1",
		IsActive:  true,
		IsDefault: true,
	}

	// Register v2 - current development version
	vm.versions["v2"] = &APIVersion{
		Version:          "v2",
		IsActive:         true,
		IsDefault:        false,
		SuccessorVersion: "",
	}

	vm.defaultVersion = "v1"

	return vm
}

// GetVersion returns the API version from request
func (vm *VersionManager) GetVersion(r *http.Request) string {
	// First, try to get version from URL path
	if vars := mux.Vars(r); vars != nil {
		if version, ok := vars["version"]; ok {
			if _, exists := vm.versions[version]; exists {
				return version
			}
		}
	}

	// Try to get version from URL path directly by parsing the path
	path := r.URL.Path
	if match := versionPathRegex.FindStringSubmatch(path); match != nil {
		version := match[1]
		if _, exists := vm.versions[version]; exists {
			return version
		}
	}

	// Fallback to Accept-Version header
	if acceptVersion := r.Header.Get("Accept-Version"); acceptVersion != "" {
		// Handle version ranges (e.g., "v1-v2")
		if strings.Contains(acceptVersion, "-") {
			// Simple range support - return the highest available
			if strings.HasPrefix(acceptVersion, "v1") {
				return "v1"
			}
		}
		if _, exists := vm.versions[acceptVersion]; exists {
			return acceptVersion
		}
	}

	// Return default version
	return vm.defaultVersion
}

// GetVersionInfo returns version metadata
func (vm *VersionManager) GetVersionInfo(version string) *APIVersion {
	if v, ok := vm.versions[version]; ok {
		return v
	}
	return vm.versions[vm.defaultVersion]
}

// IsDeprecated checks if a version is deprecated
func (vm *VersionManager) IsDeprecated(version string) bool {
	v := vm.GetVersionInfo(version)
	if v == nil {
		return false
	}
	if v.DeprecationDate != nil {
		return time.Now().After(*v.DeprecationDate)
	}
	return false
}

// GetSunsetDate returns the sunset date for a version
func (vm *VersionManager) GetSunsetDate(version string) *time.Time {
	v := vm.GetVersionInfo(version)
	if v == nil {
		return nil
	}
	return v.SunsetDate
}

// GetSupportedVersions returns all supported versions
func (vm *VersionManager) GetSupportedVersions() []string {
	var versions []string
	for v := range vm.versions {
		versions = append(versions, v)
	}
	return versions
}

var versionPathRegex = regexp.MustCompile(`/v(\d+)/`)

// API version context key type
type apiVersionContextKey string

const apiVersionKey apiVersionContextKey = "api_version"

// WithAPIVersion adds API version to context
func WithAPIVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, apiVersionKey, version)
}

// GetAPIVersionFromContext retrieves API version from context
func GetAPIVersionFromContext(ctx context.Context) string {
	if v := ctx.Value(apiVersionKey); v != nil {
		if version, ok := v.(string); ok {
			return version
		}
	}
	return ""
}

// GetAPIVersion retrieves API version from request
func GetAPIVersion(r *http.Request) string {
	// Try to get from path first
	path := r.URL.Path
	if match := versionPathRegex.FindStringSubmatch(path); match != nil {
		return match[0][1:] // Remove leading /
	}

	// Check header
	if acceptVersion := r.Header.Get("Accept-Version"); acceptVersion != "" {
		return acceptVersion
	}

	// Default to v1
	return DefaultAPIVersion
}

// VersionMiddleware creates middleware for API versioning
func VersionMiddleware(vm *VersionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			version := vm.GetVersion(r)

			// Check if version is deprecated
			if vm.IsDeprecated(version) {
				// Add deprecation headers
				if sunsetDate := vm.GetSunsetDate(version); sunsetDate != nil {
					w.Header().Set("Sunset", sunsetDate.Format(http.TimeFormat))
				}
				w.Header().Set("Deprecation", "true")
				w.Header().Set("X-API-Warning", "This API version is deprecated. Please migrate to a supported version.")

				// Add Link header to successor version
				if v := vm.GetVersionInfo(version); v != nil && v.SuccessorVersion != "" {
					w.Header().Set("Link", "</v"+v.SuccessorVersion+"/>; rel=\"successor-version\"")
				}
			}

			// Store version in context for handlers to use
			ctx := r.Context()
			ctx = WithAPIVersion(ctx, version)
			r = r.WithContext(ctx)

			// Add version header to response
			w.Header().Set("X-API-Version", version)

			next.ServeHTTP(w, r)
		})
	}
}

// AddVersionHeaders adds appropriate version headers to response
func AddVersionHeaders(w http.ResponseWriter, version string, successorVersion string, sunsetDate *time.Time) {
	w.Header().Set("X-API-Version", version)

	if sunsetDate != nil {
		w.Header().Set("Sunset", sunsetDate.Format(http.TimeFormat))
		w.Header().Set("Deprecation", "true")
	}

	if successorVersion != "" {
		w.Header().Set("Link", "</v"+successorVersion+"/>; rel=\"successor-version\"")
	}
}

// RequireVersion creates middleware that requires a specific version
func RequireVersion(vm *VersionManager, requiredVersion string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			version := vm.GetVersion(r)

			if version != requiredVersion {
				http.Error(w, "Invalid API version", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// VersionNegotiator handles version negotiation for responses
type VersionNegotiator struct {
	vm *VersionManager
}

// NewVersionNegotiator creates a new version negotiator
func NewVersionNegotiator(vm *VersionManager) *VersionNegotiator {
	return &VersionNegotiator{vm: vm}
}

// NegotiateResponseVersion determines the best version for response
func (vn *VersionNegotiator) NegotiateResponseVersion(r *http.Request, requestedVersion string) string {
	// If client explicitly requested a version, use it if supported
	if requestedVersion != "" {
		if _, ok := vn.vm.versions[requestedVersion]; ok {
			return requestedVersion
		}
		logrus.Warnf("Requested version %s not supported, falling back to %s", requestedVersion, vn.vm.defaultVersion)
	}

	// Fall back to default
	return vn.vm.defaultVersion
}

// GetVersionCompatibilityInfo returns compatibility information for a version
func (vm *VersionManager) GetVersionCompatibilityInfo(version string) map[string]interface{} {
	info := map[string]interface{}{
		"version":    version,
		"is_default": version == vm.defaultVersion,
		"is_active":  true,
	}

	if v, ok := vm.versions[version]; ok {
		info["is_deprecated"] = v.DeprecationDate != nil && time.Now().After(*v.DeprecationDate)
		info["sunset_date"] = v.SunsetDate
		info["successor_version"] = v.SuccessorVersion

		if v.SunsetDate != nil {
			daysUntilSunset := v.SunsetDate.Sub(time.Now()).Hours() / 24
			info["days_until_sunset"] = daysUntilSunset
		}
	}

	return info
}
