package functions

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/codeparser"
	"github.com/functionfly/functionfly/internal/storage"
	registryrepo "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

type PasteHandler struct {
	repo        storage.Repository
	registryRepo *registryrepo.RegistryRepository
}

func NewPasteHandler(repo storage.Repository, registryRepo *registryrepo.RegistryRepository) *PasteHandler {
	return &PasteHandler{
		repo:        repo,
		registryRepo: registryRepo,
	}
}

type ParseCodeRequest struct {
	Code         string `json:"code" validate:"required,max=102400"`
	ForceLanguage string `json:"force_language,omitempty"`
}

type ParseCodeResponse struct {
	Language      string                   `json:"language"`
	Confidence    float64                  `json:"confidence"`
	Functions     []codeparser.ParsedFunction `json:"functions"`
	RawCodeLength int                      `json:"raw_code_length"`
}

type CreateFromCodeRequest struct {
	Functions  []CreateFunctionInput `json:"functions" validate:"required,min=1,max=50"`
	Visibility string               `json:"visibility" validate:"required,oneof=private public"`
	Providers  []string             `json:"providers,omitempty"`
	Region     string               `json:"region,omitempty"`
	Author     string               `json:"author,omitempty"`
	Changelog  string               `json:"changelog,omitempty"`
}

type CreateFunctionInput struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Code        string `json:"code" validate:"required"`
	Language    string `json:"language" validate:"required"`
	Description string `json:"description,omitempty"`
}

type CreateFromCodeResponse struct {
	Created []CreatedFunction `json:"created"`
	Failed  []FailedFunction  `json:"failed,omitempty"`
}

type CreatedFunction struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type FailedFunction struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

var (
	functionNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,99}$`)
	codeSanitizer     = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f]`)
)

func sanitizeFunctionName(name string) string {
	name = strings.TrimSpace(name)
	name = codeSanitizer.ReplaceAllString(name, "")
	name = strings.ReplaceAll(name, "\n", "")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\t", "")
	if len(name) > 100 {
		name = name[:100]
	}
	return name
}

func sanitizeCode(code string) string {
	return codeSanitizer.ReplaceAllString(code, "")
}

func validateFunctionName(name string) error {
	if name == "" {
		return &ValidationError{Field: "name", Message: "Function name is required"}
	}
	if len(name) > 100 {
		return &ValidationError{Field: "name", Message: "Function name must be 100 characters or less"}
	}
	if !functionNameRegex.MatchString(name) {
		return &ValidationError{Field: "name", Message: "Function name must start with a letter and contain only letters, numbers, underscores, and hyphens"}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

var validProviders = map[string]bool{"cloud": true, "edge": true, "local": true}
var validRegions = map[string]bool{
	"us-east-1": true, "us-west-1": true, "us-west-2": true,
	"eu-west-1": true, "eu-central-1": true,
	"ap-northeast-1": true, "ap-southeast-1": true,
}

func validateProviders(providers []string) error {
	if len(providers) == 0 {
		return &ValidationError{Field: "providers", Message: "At least one provider is required"}
	}
	for _, p := range providers {
		if !validProviders[p] {
			return &ValidationError{Field: "providers", Message: "Invalid provider: " + p}
		}
	}
	return nil
}

func validateRegion(region string) error {
	if region == "" {
		return nil
	}
	if !validRegions[region] {
		return &ValidationError{Field: "region", Message: "Invalid region: " + region}
	}
	return nil
}

func languageTags(language string) ([]byte, error) {
	return json.Marshal([]string{language})
}

func (h *PasteHandler) HandleParseCode(w http.ResponseWriter, r *http.Request) {
	var req ParseCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Code == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Code is required"))
		return
	}

	if len(req.Code) > 102400 {
		apierror.WriteError(w, apierror.NewBadRequest("Code exceeds maximum size of 100KB"))
		return
	}

	language := req.ForceLanguage
	if language == "" || language == "auto" {
		language = ""
	} else if !codeparser.IsValidLanguage(language) {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid language specified"))
		return
	}

	result, err := codeparser.Parse(req.Code, language)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse code")
		apierror.WriteError(w, apierror.NewInternal("Failed to parse code. Check server logs for details."))
		return
	}

	logrus.WithFields(logrus.Fields{
		"language":  result.Language,
		"functions": len(result.Functions),
		"code_size": result.RawCodeLength,
	}).Info("Code parsed successfully")

	response := ParseCodeResponse{
		Language:      result.Language,
		Confidence:    result.Confidence,
		Functions:     result.Functions,
		RawCodeLength: result.RawCodeLength,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PasteHandler) HandleCreateFromCode(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	tenantID := user.TenantID

	var req CreateFromCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if len(req.Functions) == 0 {
		apierror.WriteError(w, apierror.NewBadRequest("At least one function is required"))
		return
	}

	if len(req.Functions) > 50 {
		apierror.WriteError(w, apierror.NewBadRequest("Maximum 50 functions can be created at once"))
		return
	}

	if req.Visibility != "private" && req.Visibility != "public" {
		apierror.WriteError(w, apierror.NewBadRequest("Visibility must be 'private' or 'public'"))
		return
	}

	var created []CreatedFunction
	var failed []FailedFunction

	for _, fn := range req.Functions {
		sanitizedName := sanitizeFunctionName(fn.Name)

		if err := validateFunctionName(sanitizedName); err != nil {
			failed = append(failed, FailedFunction{
				Name:  fn.Name,
				Error: err.Error(),
			})
			logrus.WithFields(logrus.Fields{
				"name":      fn.Name,
				"tenant_id": tenantID,
				"reason":    "invalid_name",
			}).Warn("Function creation rejected - invalid name")
			continue
		}

		if fn.Code == "" {
			failed = append(failed, FailedFunction{
				Name:  sanitizedName,
				Error: "Function code is required",
			})
			continue
		}

		if len(fn.Code) > 102400 {
			failed = append(failed, FailedFunction{
				Name:  sanitizedName,
				Error: "Function code exceeds maximum size of 100KB",
			})
			continue
		}

		providers := req.Providers
		if len(providers) == 0 {
			providers = []string{"cloud"}
		}

		region := req.Region
		if region == "" {
			region = "us-east-1"
		}

		if req.Visibility == "private" {
			function := &storage.FunctionConfig{
				TenantID:  tenantID,
				Name:      sanitizedName,
				Providers: providers,
				Region:    region,
				Code:      fn.Code,
				Status:    "draft",
			}

			createdFn, err := h.repo.CreateFunction(r.Context(), function)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"name":      sanitizedName,
					"tenant_id": tenantID,
					"error":     err.Error(),
				}).Error("Failed to create function from code")
				failed = append(failed, FailedFunction{
					Name:  sanitizedName,
					Error: "Failed to create function",
				})
				continue
			}

			logrus.WithFields(logrus.Fields{
				"function_id": createdFn.ID,
				"name":        sanitizedName,
				"tenant_id":   tenantID,
				"language":    fn.Language,
			}).Info("Function created from code paste")

			created = append(created, CreatedFunction{
				ID:     createdFn.ID.String(),
				Name:   createdFn.Name,
				Status: createdFn.Status,
			})
		} else {
			function := &registryrepo.RegistryFunction{
				Author:       tenantID.String(),
				Name:         sanitizedName,
				Title:        nullString(fn.Name),
				Description:  nullString(fn.Description),
				Visibility:   "public",
				Providers:    providers,
				Region:       region,
				Code:         fn.Code,
				Status:       "draft",
				TenantID:     &tenantID,
				OwnerUserID:  &user.UserID,
				Capabilities: json.RawMessage(`["code_execution"]`),
			}

			if fn.Language != "" {
				function.Tags = json.RawMessage(`["` + fn.Language + `"]`)
			}

			err := h.registryRepo.CreateFunction(r.Context(), function)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"name":      sanitizedName,
					"tenant_id": tenantID,
					"error":     err.Error(),
				}).Error("Failed to create public function")
				failed = append(failed, FailedFunction{
					Name:  sanitizedName,
					Error: "Failed to create public function",
				})
				continue
			}

			logrus.WithFields(logrus.Fields{
				"function_id": function.ID,
				"name":       sanitizedName,
				"tenant_id":  tenantID,
				"language":   fn.Language,
			}).Info("Public function created from code paste")

			created = append(created, CreatedFunction{
				ID:     function.ID.String(),
				Name:   function.Name,
				Status: function.Status,
			})
		}
	}

	response := CreateFromCodeResponse{
		Created: created,
		Failed:  failed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getTenantIDFromRequest(r *http.Request) uuid.UUID {
	if claims := middleware.GetUserFromContext(r); claims != nil {
		return claims.TenantID
	}
	return uuid.Nil
}
