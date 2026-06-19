// Package versioning provides HTTP middleware for API version detection and management.
package versioning

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

const (
	// DefaultAPIVersion is the default API version when no version is specified
	DefaultAPIVersion = "v1"

	// Context key for version info
	versionInfoKey contextKey = "version_info"
)

// contextKey is a type for context keys
type contextKey string

// VersionInfo contains version information extracted from the request
type VersionInfo struct {
	Version         string
	IsActive        bool
	IsDeprecated    bool
	IsSunset        bool
	IsDefault       bool
	DeprecationInfo *DeprecationWarning
}

// DeprecationWarning contains warning information for deprecated versions
type DeprecationWarning struct {
	DeprecatedAt     time.Time
	SunsetAt         *time.Time
	SunsetMessage    string
	SuccessorVersion string
}

// VersionMiddleware creates middleware for API version detection
type VersionMiddleware struct {
	repo           *Repository
	cache          map[string]*APIVersion
	defaultVersion string
}

// NewVersionMiddleware creates a new version middleware with caching
func NewVersionMiddleware(repo *Repository) *VersionMiddleware {
	return &VersionMiddleware{
		repo:           repo,
		cache:          make(map[string]*APIVersion),
		defaultVersion: DefaultAPIVersion,
	}
}

// Handler returns the middleware handler function
func (vm *VersionMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract version from URL path
			version := vm.extractVersion(r)

			// Get version info (from cache or database)
			info := vm.getVersionInfo(r.Context(), version)

			// Check if version is sunset - return 410 Gone
			if info.IsSunset {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Deprecation", "true")
				if info.DeprecationInfo != nil && info.DeprecationInfo.SunsetAt != nil {
					w.Header().Set("Sunset", info.DeprecationInfo.SunsetAt.Format(http.TimeFormat))
					if info.DeprecationInfo.SunsetMessage != "" {
						w.Header().Set("X-API-Sunset-Message", info.DeprecationInfo.SunsetMessage)
					}
				}
				w.WriteHeader(http.StatusGone)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":            "VERSION_SUNSET",
					"message":         "This API version has been sunset",
					"sunsetAt":        info.DeprecationInfo.SunsetAt,
					"successorVersion": info.DeprecationInfo.SuccessorVersion,
				})
				return
			}

			// Add deprecation headers if version is deprecated
			if info.IsDeprecated && info.DeprecationInfo != nil {
				w.Header().Set("Deprecation", "true")
				w.Header().Set("X-API-Warning", "This API version is deprecated")

				if info.DeprecationInfo.SunsetAt != nil {
					w.Header().Set("Sunset", info.DeprecationInfo.SunsetAt.Format(http.TimeFormat))
				}

				if info.DeprecationInfo.SunsetMessage != "" {
					w.Header().Set("X-API-Sunset-Message", info.DeprecationInfo.SunsetMessage)
				}

				if info.DeprecationInfo.SuccessorVersion != "" {
					w.Header().Set("Link", "</"+info.DeprecationInfo.SuccessorVersion+"/>; rel=\"successor-version\"")
				}
			}

			// Add version header to response
			w.Header().Set("X-API-Version", info.Version)

			// Store version info in context
			ctx := context.WithValue(r.Context(), versionInfoKey, info)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// extractVersion extracts the version from the URL path
func (vm *VersionMiddleware) extractVersion(r *http.Request) string {
	// Try to extract from URL path like /v1/, /v2/
	path := r.URL.Path

	// Match /v{n}/ pattern at the beginning of the path
	if match := versionPathRegex.FindStringSubmatch(path); match != nil {
		return match[1] // Returns "v1", "v2", etc.
	}

	// Try to get from header
	if acceptVersion := r.Header.Get("Accept-Version"); acceptVersion != "" {
		normalized := strings.TrimSpace(acceptVersion)
		if normalized != "" && !strings.HasPrefix(normalized, "v") {
			// Allow clients to send "1" or "2" instead of "v1"/"v2".
			if regexp.MustCompile(`^\d+$`).MatchString(normalized) {
				normalized = "v" + normalized
			}
		}
		if strings.HasPrefix(normalized, "v") {
			return normalized
		}
	}

	// Return default version
	return vm.defaultVersion
}

// getVersionInfo retrieves version info, using cache when available
func (vm *VersionMiddleware) getVersionInfo(ctx context.Context, version string) *VersionInfo {
	// Check cache first
	if cached, ok := vm.cache[version]; ok {
		return vm.toVersionInfo(cached)
	}

	// Try to get from database
	var apiVersion *APIVersion
	if vm.repo != nil {
		var err error
		apiVersion, err = vm.repo.GetAPIVersionByVersion(ctx, version)
		if err != nil {
			logrus.WithError(err).WithField("version", version).Warn("Failed to get version from database, using default")
		}
	}

	// If not found in database, try default
	if apiVersion == nil && version != vm.defaultVersion {
		if vm.repo != nil {
			apiVersion, _ = vm.repo.GetAPIVersionByVersion(ctx, vm.defaultVersion)
		}
	}

	// If still not found, create a basic version info
	if apiVersion == nil {
		return &VersionInfo{
			Version:      vm.defaultVersion,
			IsActive:     true,
			IsDefault:    true,
			IsDeprecated: false,
		}
	}

	// Cache the result
	vm.cache[version] = apiVersion

	return vm.toVersionInfo(apiVersion)
}

// toVersionInfo converts an APIVersion to VersionInfo
func (vm *VersionMiddleware) toVersionInfo(apiVersion *APIVersion) *VersionInfo {
	isSunset := apiVersion.Status == APIVersionStatusSunset ||
		(apiVersion.SunsetAt != nil && apiVersion.SunsetAt.Before(time.Now()))

	info := &VersionInfo{
		Version:      apiVersion.Version,
		IsActive:     apiVersion.Status == APIVersionStatusActive && !isSunset,
		IsDefault:    apiVersion.Version == vm.defaultVersion,
		IsDeprecated: apiVersion.Status == APIVersionStatusDeprecated || isSunset,
		IsSunset:    isSunset,
	}

	if info.IsDeprecated && apiVersion.DeprecatedAt != nil {
		info.DeprecationInfo = &DeprecationWarning{
			DeprecatedAt:     *apiVersion.DeprecatedAt,
			SunsetAt:         apiVersion.SunsetAt,
			SunsetMessage:    apiVersion.SunsetMessage,
			SuccessorVersion: vm.getSuccessorVersion(apiVersion),
		}
	}

	// Sunset can be driven by status or sunset_at even without deprecated_at; still populate warning info.
	if info.DeprecationInfo == nil && info.IsSunset && apiVersion.SunsetAt != nil {
		deprecatedAt := time.Now()
		if apiVersion.DeprecatedAt != nil {
			deprecatedAt = *apiVersion.DeprecatedAt
		}
		info.DeprecationInfo = &DeprecationWarning{
			DeprecatedAt:     deprecatedAt,
			SunsetAt:         apiVersion.SunsetAt,
			SunsetMessage:    apiVersion.SunsetMessage,
			SuccessorVersion: vm.getSuccessorVersion(apiVersion),
		}
	}

	return info
}

// getSuccessorVersion determines the successor version
func (vm *VersionMiddleware) getSuccessorVersion(v *APIVersion) string {
	// Extract major version number
	var major int
	if _, err := fmt.Sscanf(v.Version, "v%d", &major); err != nil {
		return ""
	}

	// Check if there's a known successor
	successorMap := map[int]string{
		1: "v2",
		2: "v3",
	}

	if successor, ok := successorMap[major]; ok {
		return successor
	}

	return ""
}

// versionPathRegex matches version patterns in URL paths
var versionPathRegex = regexp.MustCompile(`/(v\d+)/`)

// GetVersionInfo retrieves version info from context
func GetVersionInfo(ctx context.Context) *VersionInfo {
	if info, ok := ctx.Value(versionInfoKey).(*VersionInfo); ok {
		return info
	}
	return nil
}

// RequireAPIVersion creates middleware that requires a specific API version
func RequireAPIVersion(requiredVersion string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := GetVersionInfo(r.Context())
			if info == nil || info.Version != requiredVersion {
				apierror.WriteError(w, apierror.NewBadRequest("Invalid API version"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ResponseWithVersionInfo adds version information to a JSON response
func ResponseWithVersionInfo(w http.ResponseWriter, version string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-API-Version", version)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithError(err).Error("Failed to encode response")
		apierror.WriteError(w, apierror.NewInternal("Internal server error"))
	}
}

// GetAPIVersionFromContext retrieves the API version from context
func GetAPIVersionFromContext(ctx context.Context) string {
	info := GetVersionInfo(ctx)
	if info != nil {
		return info.Version
	}
	return DefaultAPIVersion
}

// GetUserFromContext retrieves the authenticated user (auth claims) from context when only context is available.
// The value is set by the auth middleware; returns nil if not authenticated.
func GetUserFromContext(ctx context.Context) interface{} {
	return middleware.GetClaimsFromContext(ctx)
}
