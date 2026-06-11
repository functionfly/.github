package middleware

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ValidatedRequest interface {
	Validate() error
}

type Validatable interface {
	Validate() error
}

type ValidationResult struct {
	Valid        bool
	ErrorCode    string
	ErrorMessage string
	Field        string
}

func ValidateRequestMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}

		if req, ok := r.Context().Value("validated_request").(ValidatedRequest); ok {
			if err := req.Validate(); err != nil {
				err := apierror.NewValidation(err.Error())
				apierror.WriteError(w, err)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}

func RequireJSONValidation[T ValidatedRequest](target T, handler func(http.ResponseWriter, *http.Request, T)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req T
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			err := apierror.NewBadRequest("Invalid JSON: " + err.Error())
			apierror.WriteError(w, err)
			return
		}

		if err := req.Validate(); err != nil {
			validationErr := apierror.NewValidation(err.Error())
			apierror.WriteError(w, validationErr)
			return
		}

		handler(w, r, req)
	}
}

type FieldValidator struct {
	Field string
	Value interface{}
	Rules []ValidationRule
}

type ValidationRule struct {
	Name    string
	Check   func(interface{}) bool
	Message string
}

func ValidateField(field string, value interface{}, required bool, validator func(interface{}) bool, errorMsg string) *apierror.APIError {
	if required && value == nil {
		return apierror.ValidationFieldError(field, "This field is required")
	}
	if value != nil && !validator(value) {
		return apierror.ValidationFieldError(field, errorMsg)
	}
	return nil
}

var (
	NonEmpty = func(v interface{}) bool {
		s, ok := v.(string)
		return ok && s != ""
	}

	MinLength = func(min int) func(interface{}) bool {
		return func(v interface{}) bool {
			s, ok := v.(string)
			return ok && len(s) >= min
		}
	}

	MaxLength = func(max int) func(interface{}) bool {
		return func(v interface{}) bool {
			s, ok := v.(string)
			return ok && len(s) <= max
		}
	}

	Email = func(v interface{}) bool {
		s, ok := v.(string)
		if !ok || s == "" {
			return false
		}
		for i := 0; i < len(s); i++ {
			if s[i] == '@' {
				return i > 0 && i < len(s)-1
			}
		}
		return false
	}

	UUID = func(v interface{}) bool {
		s, ok := v.(string)
		if !ok || s == "" {
			return false
		}
		if len(s) != 36 {
			return false
		}
		return s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-'
	}

	PositiveInt = func(v interface{}) bool {
		switch n := v.(type) {
		case int:
			return n > 0
		case int64:
			return n > 0
		case float64:
			return n > 0
		default:
			return false
		}
	}

	NonNegativeInt = func(v interface{}) bool {
		switch n := v.(type) {
		case int:
			return n >= 0
		case int64:
			return n >= 0
		case float64:
			return n >= 0
		default:
			return false
		}
	}
)

type QueryParamValidator struct {
	Param   string
	Rules   []QueryParamRule
	Message string
}

type QueryParamRule struct {
	Check   func(string) bool
	Message string
}

func QueryParam(param string, rules ...QueryParamRule) QueryParamValidator {
	return QueryParamValidator{Param: param, Rules: rules}
}

func (v QueryParamValidator) Validate(values map[string]string) *apierror.APIError {
	value, exists := values[v.Param]
	if !exists {
		for _, rule := range v.Rules {
			if _, ok := rule.Check(""); ok {
				return apierror.ValidationFieldError(v.Param, rule.Message)
			}
		}
		return nil
	}
	for _, rule := range v.Rules {
		if !rule.Check(value) {
			return apierror.ValidationFieldError(v.Param, rule.Message)
		}
	}
	return nil
}

func Required() QueryParamRule {
	return QueryParamRule{
		Check:   func(v string) bool { return v != "" },
		Message: "This parameter is required",
	}
}

func MinLen(min int) QueryParamRule {
	return QueryParamRule{
		Check:   func(v string) bool { return len(v) >= min },
		Message: "Value must be at least " + strconv.Itoa(min) + " characters",
	}
}

func MaxLen(max int) QueryParamRule {
	return QueryParamRule{
		Check:   func(v string) bool { return len(v) <= max },
		Message: "Value must be at most " + strconv.Itoa(max) + " characters",
	}
}

func MatchesPattern(pattern string) QueryParamRule {
	re := regexp.MustCompile(pattern)
	return QueryParamRule{
		Check:   func(v string) bool { return re.MatchString(v) },
		Message: "Value does not match required format",
	}
}

func OneOf(allowed ...string) QueryParamRule {
	allowedSet := make(map[string]bool)
	for _, a := range allowed {
		allowedSet[a] = true
	}
	return QueryParamRule{
		Check:   func(v string) bool { return allowedSet[v] },
		Message: "Value must be one of: " + joinStrings(allowed),
	}
}

func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

type QueryValidationMiddleware struct {
	validators []QueryParamValidator
}

func NewQueryValidationMiddleware(validators ...QueryParamValidator) *QueryValidationMiddleware {
	return &QueryValidationMiddleware{validators: validators}
}

func (m *QueryValidationMiddleware) ValidateQueryParams(values map[string]string) *apierror.APIError {
	for _, v := range m.validators {
		if err := v.Validate(values); err != nil {
			return err
		}
	}
	return nil
}

func (m *QueryValidationMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values := make(map[string]string)
		for key, vals := range r.URL.Query() {
			if len(vals) > 0 {
				values[key] = vals[0]
			}
		}
		if err := m.ValidateQueryParams(values); err != nil {
			apierror.WriteError(w, err)
			return
		}
		next.ServeHTTP(w, r)
	}
}

type PathParamValidator struct {
	Param   string
	Rules   []PathParamRule
	Message string
}

type PathParamRule struct {
	Check   func(string) bool
	Message string
}

func PathParam(param string, rules ...PathParamRule) PathParamValidator {
	return PathParamValidator{Param: param, Rules: rules}
}

func (v PathParamValidator) Validate(value string) *apierror.APIError {
	for _, rule := range v.Rules {
		if !rule.Check(value) {
			return apierror.ValidationFieldError(v.Param, rule.Message)
		}
	}
	return nil
}

func IsUUID() PathParamRule {
	return PathParamRule{
		Check: func(v string) bool {
			if v == "" {
				return false
			}
			_, err := uuid.Parse(v)
			return err == nil
		},
		Message: "Must be a valid UUID",
	}
}

func IsInt() PathParamRule {
	return PathParamRule{
		Check: func(v string) bool {
			if v == "" {
				return false
			}
			_, err := strconv.Atoi(v)
			return err == nil
		},
		Message: "Must be a valid integer",
	}
}

func IsPositiveInt() PathParamRule {
	return PathParamRule{
		Check: func(v string) bool {
			if v == "" {
				return false
			}
			n, err := strconv.Atoi(v)
			return err == nil && n > 0
		},
		Message: "Must be a positive integer",
	}
}

func MatchesPathPattern(pattern string) PathParamRule {
	re := regexp.MustCompile(pattern)
	return PathParamRule{
		Check:   func(v string) bool { return re.MatchString(v) },
		Message: "Value does not match required format",
	}
}

type PathValidationMiddleware struct {
	validators []PathParamValidator
}

func NewPathValidationMiddleware(validators ...PathParamValidator) *PathValidationMiddleware {
	return &PathValidationMiddleware{validators: validators}
}

func (m *PathValidationMiddleware) ValidatePathParams(values map[string]string) *apierror.APIError {
	for _, v := range m.validators {
		value, exists := values[v.Param]
		if !exists {
			value = ""
		}
		if err := v.Validate(value); err != nil {
			return err
		}
	}
	return nil
}

func (m *PathValidationMiddleware) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values := mux.Vars(r)
		if err := m.ValidatePathParams(values); err != nil {
			apierror.WriteError(w, err)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func ValidateUUIDParam(vars map[string]string, param string) *apierror.APIError {
	value, exists := vars[param]
	if !exists || value == "" {
		return apierror.NewBadRequest("Missing required path parameter: " + param)
	}
	if _, err := uuid.Parse(value); err != nil {
		return apierror.ValidationFieldError(param, "Must be a valid UUID")
	}
	return nil
}

func ValidateUUIDParamWithRequestID(vars map[string]string, param string, requestID string) *apierror.APIError {
	err := ValidateUUIDParam(vars, param)
	if err != nil && requestID != "" {
		err = err.WithRequestID(requestID)
	}
	return err
}

func ValidateIntParam(vars map[string]string, param string) (int, *apierror.APIError) {
	value, exists := vars[param]
	if !exists || value == "" {
		return 0, apierror.NewBadRequest("Missing required path parameter: " + param)
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, apierror.ValidationFieldError(param, "Must be a valid integer")
	}
	return n, nil
}

func ValidateRequiredQuery(r *http.Request, params ...string) *apierror.APIError {
	for _, param := range params {
		if val := r.URL.Query().Get(param); val == "" {
			return apierror.ValidationFieldError(param, "This parameter is required")
		}
	}
	return nil
}

func ParsePositiveIntParam(query string, defaultVal int) (int, *apierror.APIError) {
	if query == "" {
		return defaultVal, nil
	}
	n, err := strconv.Atoi(query)
	if err != nil {
		return 0, apierror.NewInvalidLimit("limit must be a valid integer")
	}
	if n < 0 {
		return 0, apierror.NewInvalidOffset("value cannot be negative")
	}
	if n > 100 {
		n = 100
	}
	return n, nil
}

func ParseOffsetParam(query string) (int, *apierror.APIError) {
	if query == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(query)
	if err != nil {
		return 0, apierror.NewInvalidOffset("offset must be a valid integer")
	}
	if n < 0 {
		return 0, apierror.NewInvalidOffset("offset cannot be negative")
	}
	return n, nil
}
