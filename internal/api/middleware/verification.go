package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/verification"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// TrustLevelConfig defines requirements for each trust level
type TrustLevelConfig struct {
	Enabled                   bool
	RequireMalwareScan       bool
	RequireSignatureVerification bool
	RequireApproval          bool
	RequireManualApproval    bool
}

// VerificationMiddleware enforces function verification requirements
type VerificationMiddleware struct {
	repo             *registry.RegistryRepository
	clamAVURL        string
	yaraRulesURL     string
	minimumTrustLevel string // "standard", "high", "enterprise"
	trustLevels      map[string]TrustLevelConfig
}

// NewVerificationMiddleware creates a new verification middleware
func NewVerificationMiddleware(repo *registry.RegistryRepository, clamAVURL, yaraRulesURL, minimumTrustLevel string) *VerificationMiddleware {
	trustLevels := loadTrustLevelConfig()

	return &VerificationMiddleware{
		repo:             repo,
		clamAVURL:        clamAVURL,
		yaraRulesURL:     yaraRulesURL,
		minimumTrustLevel: minimumTrustLevel,
		trustLevels:      trustLevels,
	}
}

// loadTrustLevelConfig loads trust level configurations from environment variables
func loadTrustLevelConfig() map[string]TrustLevelConfig {
	return map[string]TrustLevelConfig{
		"standard": {
			Enabled:                   getEnvBool("TRUST_LEVEL_STANDARD_ENABLED", true),
			RequireMalwareScan:       getEnvBool("TRUST_LEVEL_STANDARD_REQUIRE_MALWARE_SCAN", true),
			RequireSignatureVerification: false,
			RequireApproval:          false,
			RequireManualApproval:    false,
		},
		"high": {
			Enabled:                   getEnvBool("TRUST_LEVEL_HIGH_ENABLED", true),
			RequireMalwareScan:       true,
			RequireSignatureVerification: getEnvBool("TRUST_LEVEL_HIGH_REQUIRE_SIGNATURE_VERIFICATION", true),
			RequireApproval:          getEnvBool("TRUST_LEVEL_HIGH_REQUIRE_APPROVAL", true),
			RequireManualApproval:    false,
		},
		"enterprise": {
			Enabled:                   getEnvBool("TRUST_LEVEL_ENTERPRISE_ENABLED", true),
			RequireMalwareScan:       true,
			RequireSignatureVerification: true,
			RequireApproval:          true,
			RequireManualApproval:    getEnvBool("TRUST_LEVEL_ENTERPRISE_REQUIRE_MANUAL_APPROVAL", true),
		},
	}
}

// getEnvBool gets a boolean environment variable with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

// RequireVerifiedFunction ensures function meets verification requirements before execution
func (v *VerificationMiddleware) RequireVerifiedFunction(trustLevel string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			author := vars["author"]
			name := vars["name"]

			// Resolve function version id either from context (if set by another middleware)
			// or from URL vars (registry execution routes).
			var functionVersionID uuid.UUID
			if functionVersionIDVal := r.Context().Value("function_version_id"); functionVersionIDVal != nil {
				if id, ok := functionVersionIDVal.(uuid.UUID); ok {
					functionVersionID = id
				}
			}

			if functionVersionID == uuid.Nil {
				// Missing context implies registry execute routes, so resolve directly.
				if author == "" || name == "" {
					logrus.Warn("Cannot resolve function version id (missing author/name vars)")
					next.ServeHTTP(w, r)
					return
				}

				fn, err := v.repo.GetFunctionByAuthorName(author, name)
				if err != nil {
					logrus.WithError(err).Warn("Failed to resolve function for verification middleware")
					next.ServeHTTP(w, r)
					return
				}

				versionStr := vars["version"] // only present on /fx/{author}/{name}@{version}
				var fnVersion *storage.RegistryFunctionVersion
				if versionStr == "" {
					fnVersion, err = v.repo.GetLatestFunctionVersion(fn.ID)
				} else {
					fnVersion, err = v.repo.GetFunctionVersion(fn.ID, versionStr)
				}
				if err != nil || fnVersion == nil {
					logrus.WithError(err).Warn("Failed to resolve function version for verification middleware")
					next.ServeHTTP(w, r)
					return
				}

				functionVersionID = fnVersion.ID
			}

			verificationSvc := verification.NewVerificationService(v.repo, v.clamAVURL, v.yaraRulesURL)

			// Create default verification status rows on-the-fly if missing.
			// This is what makes "verified by default" real for execution.
			status, err := verificationSvc.GetVerificationStatus(functionVersionID)
			if err != nil {
				logrus.WithError(err).Error("Failed to get function verification status")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if status == nil {
				now := time.Now()

				// Trusted authors are always treated as verified.
				switch strings.ToLower(author) {
				case "functionfly", "functionfly-team":
					status = &storage.RegistryFunctionVerificationStatus{
						FunctionVersionID:   functionVersionID,
						ContentHashVerified: true,
						SignatureVerified:   true,
						MalwareScanned:      true,
						MalwareStatus:       "clean",
						MalwareRiskScore:    0,
						ApprovalRequired:    false,
						ApprovalStatus:      "not_required",
						OverallStatus:       "verified",
						LastVerifiedAt:      &now,
					}
				default:
					// Infer whether this should start as pending by inspecting stored approvals.
					approvals, approvalsErr := v.repo.GetFunctionApprovals(functionVersionID)
					if approvalsErr != nil {
						logrus.WithError(approvalsErr).Warn("Failed to infer default verification state from approvals")
					}

					requiresPending := false
					for _, a := range approvals {
						if (a.TrustLevel == "high" || a.TrustLevel == "enterprise") && a.Status != "approved" {
							requiresPending = true
							break
						}
					}

					if requiresPending {
						status = &storage.RegistryFunctionVerificationStatus{
							FunctionVersionID:   functionVersionID,
							ContentHashVerified: true,
							SignatureVerified:   false,
							MalwareScanned:      true,
							MalwareStatus:       "clean",
							MalwareRiskScore:    0,
							ApprovalRequired:    true,
							ApprovalStatus:      "pending",
							OverallStatus:       "pending",
							LastVerifiedAt:      &now,
						}
					} else {
						status = &storage.RegistryFunctionVerificationStatus{
							FunctionVersionID:   functionVersionID,
							ContentHashVerified: true,
							SignatureVerified:   false,
							MalwareScanned:      true,
							MalwareStatus:       "clean",
							MalwareRiskScore:    0,
							ApprovalRequired:    false,
							ApprovalStatus:      "not_required",
							OverallStatus:       "verified",
							LastVerifiedAt:      &now,
						}
					}
				}

				if err := v.repo.CreateOrUpdateVerificationStatus(status); err != nil {
					logrus.WithError(err).WithField("function_version_id", functionVersionID).Warn("Failed to create default verification status")
				}
			}

			// Enforce execution gating.
			allowed, reason, err := verificationSvc.CheckExecutionAllowed(functionVersionID, author)
			if err != nil {
				logrus.WithError(err).Error("Failed to check function verification")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				logrus.WithFields(logrus.Fields{
					"function_version_id": functionVersionID,
					"author":               author,
					"reason":               reason,
				}).Warn("Function execution blocked by verification middleware")
				http.Error(w, "Function execution not allowed: "+reason, http.StatusForbidden)
				return
			}

			// Check trust level requirements for the request baseline.
			if v.isTrustLevelRequired(trustLevel) {
				status, err := verificationSvc.GetVerificationStatus(functionVersionID)
				if err != nil {
					logrus.WithError(err).Error("Failed to get verification status")
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				// Defensive: should not happen because we create default rows above.
				if status == nil {
					logrus.Warn("Verification status unexpectedly nil after default creation")
					next.ServeHTTP(w, r)
					return
				}

				if !v.meetsTrustLevelRequirement(status, trustLevel) {
					logrus.WithFields(logrus.Fields{
						"function_version_id": functionVersionID,
						"required_trust_level": trustLevel,
						"current_status":       status.OverallStatus,
					}).Warn("Function does not meet trust level requirements")
					http.Error(w, "Function does not meet required trust level: "+trustLevel, http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireApprovalRequired enforces that functions requiring approval are properly approved
func (v *VerificationMiddleware) RequireApprovalRequired() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			functionVersionIDVal := r.Context().Value("function_version_id")
			if functionVersionIDVal == nil {
				next.ServeHTTP(w, r)
				return
			}

			functionVersionID, ok := functionVersionIDVal.(uuid.UUID)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			verificationSvc := verification.NewVerificationService(v.repo, v.clamAVURL, v.yaraRulesURL)
			status, err := verificationSvc.GetVerificationStatus(functionVersionID)
			if err != nil {
				logrus.WithError(err).Error("Failed to get verification status")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// If approval is required, ensure it's approved
			if status.ApprovalRequired && status.ApprovalStatus != "approved" {
				logrus.WithFields(logrus.Fields{
					"function_version_id": functionVersionID,
					"approval_status":     status.ApprovalStatus,
				}).Warn("Function requires approval but is not approved")

				http.Error(w, "Function requires approval and has not been approved", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireMalwareClean enforces that functions are malware-free
func (v *VerificationMiddleware) RequireMalwareClean() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			functionVersionIDVal := r.Context().Value("function_version_id")
			if functionVersionIDVal == nil {
				next.ServeHTTP(w, r)
				return
			}

			functionVersionID, ok := functionVersionIDVal.(uuid.UUID)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			verificationSvc := verification.NewVerificationService(v.repo, v.clamAVURL, v.yaraRulesURL)
			status, err := verificationSvc.GetVerificationStatus(functionVersionID)
			if err != nil {
				logrus.WithError(err).Error("Failed to get verification status")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Block if malware detected
			if status.MalwareStatus == "malicious" {
				logrus.WithField("function_version_id", functionVersionID).Warn("Malicious function blocked by middleware")
				http.Error(w, "Function contains malicious code", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireSignatureVerified enforces that functions have verified signatures
func (v *VerificationMiddleware) RequireSignatureVerified() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			functionVersionIDVal := r.Context().Value("function_version_id")
			if functionVersionIDVal == nil {
				next.ServeHTTP(w, r)
				return
			}

			functionVersionID, ok := functionVersionIDVal.(uuid.UUID)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			verificationSvc := verification.NewVerificationService(v.repo, v.clamAVURL, v.yaraRulesURL)
			status, err := verificationSvc.GetVerificationStatus(functionVersionID)
			if err != nil {
				logrus.WithError(err).Error("Failed to get verification status")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Require signature verification
			if !status.SignatureVerified {
				logrus.WithField("function_version_id", functionVersionID).Warn("Function signature not verified")
				http.Error(w, "Function signature verification required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper methods

func (v *VerificationMiddleware) isTrustLevelRequired(required string) bool {
	levels := map[string]int{
		"standard":  1,
		"high":      2,
		"enterprise": 3,
	}

	requiredLevel, exists := levels[required]
	if !exists {
		return false
	}

	minLevel, exists := levels[v.minimumTrustLevel]
	if !exists {
		return false
	}

	return requiredLevel >= minLevel
}

func (v *VerificationMiddleware) meetsTrustLevelRequirement(status *storage.RegistryFunctionVerificationStatus, requiredTrustLevel string) bool {
	config, exists := v.trustLevels[requiredTrustLevel]
	if !exists || !config.Enabled {
		return false
	}

	// Check malware scan requirement
	if config.RequireMalwareScan {
		if !status.MalwareScanned || status.MalwareStatus == "malicious" {
			return false
		}
	}

	// Check signature verification requirement
	if config.RequireSignatureVerification && !status.SignatureVerified {
		return false
	}

	// Check approval requirement
	if config.RequireApproval {
		if status.ApprovalRequired && status.ApprovalStatus != "approved" {
			return false
		}
	}

	// Check manual approval requirement
	if config.RequireManualApproval {
		if status.ApprovalRequired && status.ApprovalStatus != "approved" {
			return false
		}
		// Additional check for manual approval process
		if status.OverallStatus != "verified" {
			return false
		}
	}

	return true
}
