package categorization

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// FunctionSpec represents the function specification for categorization
type FunctionSpec struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	InputSchema  map[string]any `json:"input_schema"`
	OutputSchema map[string]any `json:"output_schema"`
	Code         string         `json:"code"`
	Runtime      string         `json:"runtime"`
}

// CategorizationResult represents the result of auto-categorization
type CategorizationResult struct {
	FunctionID        uuid.UUID `json:"function_id"`
	PrimaryCategory   string    `json:"primary_category"`
	SecondaryCategory string    `json:"secondary_category,omitempty"`
	Tags              []string  `json:"tags"`
	Confidence        float64   `json:"confidence"`
	Reasoning         string    `json:"reasoning"`
	CodePatterns      []string  `json:"code_patterns,omitempty"`
	InputTypes        []string  `json:"input_types,omitempty"`
	OutputTypes       []string  `json:"output_types,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// FunctionCategory stores the categorization result for a function
type FunctionCategory struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID        uuid.UUID      `json:"function_id" gorm:"type:uuid;uniqueIndex;not null"`
	PrimaryCategory   string         `json:"primary_category" gorm:"not null;index"`
	SecondaryCategory string         `json:"secondary_category"`
	Tags              pq.StringArray `json:"tags" gorm:"type:text[]"`
	Confidence        float64        `json:"confidence" gorm:"type:decimal(5,4);not null"`
	Reasoning         string         `json:"reasoning"`
	CodePatterns      pq.StringArray `json:"code_patterns" gorm:"type:text[]"`
	InputTypes        pq.StringArray `json:"input_types" gorm:"type:text[]"`
	OutputTypes       pq.StringArray `json:"output_types" gorm:"type:text[]"`
	ManuallyEdited    bool           `json:"manually_edited" gorm:"default:false"`
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

func (FunctionCategory) TableName() string {
	return "function_categories"
}

// Service provides auto-categorization functionality
type Service struct {
	db *gorm.DB
}

// NewService creates a new categorization service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// AutoMigrate runs database migrations for categorization models
func (s *Service) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&FunctionCategory{})
}

// CategorizeFunction analyzes a function and returns categorization results
func (s *Service) CategorizeFunction(ctx context.Context, spec *FunctionSpec) (*CategorizationResult, error) {
	result := &CategorizationResult{
		FunctionID: uuid.New(),
		Tags:       []string{},
		CreatedAt:  time.Now(),
	}

	// Analyze code patterns
	codePatterns := s.analyzeCodePatterns(spec.Code)
	result.CodePatterns = codePatterns

	// Analyze input/output types
	inputTypes := s.analyzeSchemaTypes(spec.InputSchema)
	outputTypes := s.analyzeSchemaTypes(spec.OutputSchema)
	result.InputTypes = inputTypes
	result.OutputTypes = outputTypes

	// Determine primary category
	primaryCategory, confidence, reasoning := s.determineCategory(spec, codePatterns, inputTypes, outputTypes)
	result.PrimaryCategory = primaryCategory
	result.Confidence = confidence
	result.Reasoning = reasoning

	// Determine secondary category
	secondaryCategory, _ := s.determineSecondaryCategory(spec, codePatterns, primaryCategory)
	result.SecondaryCategory = secondaryCategory

	// Extract tags
	result.Tags = s.extractTags(spec, codePatterns, inputTypes, outputTypes)

	return result, nil
}

// CategorizeAndStore categorizes a function and stores the result
func (s *Service) CategorizeAndStore(ctx context.Context, functionID uuid.UUID, spec *FunctionSpec) (*FunctionCategory, error) {
	result, err := s.CategorizeFunction(ctx, spec)
	if err != nil {
		return nil, err
	}

	// Check if categorization already exists
	var existing FunctionCategory
	err = s.db.WithContext(ctx).Where("function_id = ?", functionID).First(&existing).Error
	if err == nil {
		// Update existing if not manually edited
		if !existing.ManuallyEdited {
			existing.PrimaryCategory = result.PrimaryCategory
			existing.SecondaryCategory = result.SecondaryCategory
			existing.Tags = result.Tags
			existing.Confidence = result.Confidence
			existing.Reasoning = result.Reasoning
			existing.CodePatterns = result.CodePatterns
			existing.InputTypes = result.InputTypes
			existing.OutputTypes = result.OutputTypes
			if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
				return nil, err
			}
		}
		return &existing, nil
	}

	// Create new categorization
	fc := &FunctionCategory{
		ID:                uuid.New(),
		FunctionID:        functionID,
		PrimaryCategory:   result.PrimaryCategory,
		SecondaryCategory: result.SecondaryCategory,
		Tags:              result.Tags,
		Confidence:        result.Confidence,
		Reasoning:         result.Reasoning,
		CodePatterns:      result.CodePatterns,
		InputTypes:        result.InputTypes,
		OutputTypes:       result.OutputTypes,
	}

	if err := s.db.WithContext(ctx).Create(fc).Error; err != nil {
		return nil, err
	}

	return fc, nil
}

// GetCategorization retrieves the categorization for a function
func (s *Service) GetCategorization(ctx context.Context, functionID uuid.UUID) (*FunctionCategory, error) {
	var fc FunctionCategory
	if err := s.db.WithContext(ctx).Where("function_id = ?", functionID).First(&fc).Error; err != nil {
		return nil, err
	}
	return &fc, nil
}

// UpdateCategorization manually updates a function's categorization
func (s *Service) UpdateCategorization(ctx context.Context, functionID uuid.UUID, primaryCategory, secondaryCategory string, tags []string) (*FunctionCategory, error) {
	var fc FunctionCategory
	if err := s.db.WithContext(ctx).Where("function_id = ?", functionID).First(&fc).Error; err != nil {
		return nil, err
	}

	fc.PrimaryCategory = primaryCategory
	fc.SecondaryCategory = secondaryCategory
	fc.Tags = tags
	fc.ManuallyEdited = true
	fc.Confidence = 1.0 // Manual edits have full confidence
	fc.Reasoning = "Manually set by user"

	if err := s.db.WithContext(ctx).Save(&fc).Error; err != nil {
		return nil, err
	}

	return &fc, nil
}

// GetFunctionsByCategory retrieves all functions in a category
func (s *Service) GetFunctionsByCategory(ctx context.Context, category string, limit, offset int) ([]FunctionCategory, int64, error) {
	var total int64
	var fcs []FunctionCategory

	query := s.db.WithContext(ctx).Model(&FunctionCategory{}).
		Where("primary_category = ? OR secondary_category = ?", category, category)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("confidence DESC, created_at DESC").
		Limit(limit).Offset(offset).Find(&fcs).Error; err != nil {
		return nil, 0, err
	}

	return fcs, total, nil
}

// GetFunctionsByTag retrieves all functions with a specific tag
func (s *Service) GetFunctionsByTag(ctx context.Context, tag string, limit, offset int) ([]FunctionCategory, int64, error) {
	var total int64
	var fcs []FunctionCategory

	query := s.db.WithContext(ctx).Model(&FunctionCategory{}).
		Where("? = ANY(tags)", tag)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("confidence DESC, created_at DESC").
		Limit(limit).Offset(offset).Find(&fcs).Error; err != nil {
		return nil, 0, err
	}

	return fcs, total, nil
}

// AnalyzeCodePatterns extracts patterns from the function code (exported for API use)
func (s *Service) AnalyzeCodePatterns(code string) []string {
	return s.analyzeCodePatterns(code)
}

// analyzeCodePatterns extracts patterns from the function code
func (s *Service) analyzeCodePatterns(code string) []string {
	var patterns []string

	// Define code pattern matchers
	patternMatchers := []struct {
		name    string
		pattern *regexp.Regexp
	}{
		// String operations
		{"string-concat", regexp.MustCompile(`(?i)(\+\s*["']|["']\s*\+|\.join\(|concat|append\s*\()`)},
		{"string-format", regexp.MustCompile(`(?i)(\.format\(|f["']|%\s*\(|format_string|template)`)},
		{"string-split", regexp.MustCompile(`(?i)(\.split\(|partition|rsplit)`)},
		{"string-replace", regexp.MustCompile(`(?i)(\.replace\(|sub\s*\(|substitute)`)},
		{"string-trim", regexp.MustCompile(`(?i)(\.strip\(|\.trim\(|lstrip|rstrip)`)},
		{"string-case", regexp.MustCompile(`(?i)(\.lower\(|\.upper\(|\.capitalize\(|\.title\(|\.swapcase\()`)},
		{"substring", regexp.MustCompile(`(?i)(\[.*:.*\]|slice|substr)`)},

		// Encoding/Decoding
		{"base64-encode", regexp.MustCompile(`(?i)(base64.*encode|b64encode|encode.*base64)`)},
		{"base64-decode", regexp.MustCompile(`(?i)(base64.*decode|b64decode|decode.*base64)`)},
		{"url-encode", regexp.MustCompile(`(?i)(url.*encode|quote\(|urllib|encodeURI)`)},
		{"url-decode", regexp.MustCompile(`(?i)(url.*decode|unquote\(|urllib|decodeURI)`)},
		{"html-encode", regexp.MustCompile(`(?i)(html.*escape|escape\(|htmlentities)`)},
		{"html-decode", regexp.MustCompile(`(?i)(html.*unescape|unescape\(|html_entity_decode)`)},

		// JSON operations
		{"json-parse", regexp.MustCompile(`(?i)(json\.loads|JSON\.parse|from_json|parse.*json)`)},
		{"json-stringify", regexp.MustCompile(`(?i)(json\.dumps|JSON\.stringify|to_json|serialize)`)},
		{"json-minify", regexp.MustCompile(`(?i)(json\.dumps.*separators|minify.*json|compact.*json)`)},
		{"json-prettify", regexp.MustCompile(`(?i)(json\.dumps.*indent|prettify|pretty.*print|beautify)`)},

		// Hashing
		{"hash-md5", regexp.MustCompile(`(?i)(md5|MD5)`)},
		{"hash-sha1", regexp.MustCompile(`(?i)(sha1|SHA1)`)},
		{"hash-sha256", regexp.MustCompile(`(?i)(sha256|SHA256|sha-256)`)},
		{"hash-sha512", regexp.MustCompile(`(?i)(sha512|SHA512|sha-512)`)},

		// UUID/ID generation
		{"uuid-gen", regexp.MustCompile(`(?i)(uuid|UUID|guid|GUID|uuid4|uuid1)`)},
		{"random-id", regexp.MustCompile(`(?i)(random.*id|generate.*id|unique.*id|nanoid)`)},

		// Validation
		{"email-validate", regexp.MustCompile(`(?i)(email.*valid|valid.*email|is_email|check.*email)`)},
		{"regex-match", regexp.MustCompile(`(?i)(re\.match|re\.search|regex|RegExp|pattern.*match)`)},
		{"type-check", regexp.MustCompile(`(?i)(isinstance|typeof|type\(|is_\w+|check.*type)`)},

		// Math operations
		{"math-basic", regexp.MustCompile(`(?i)(\+\s*\w|\-\s*\w|\*\s*\w|\/\s*\w|add|subtract|multiply|divide)`)},
		{"math-round", regexp.MustCompile(`(?i)(round|floor|ceil|trunc|truncate)`)},
		{"math-statistical", regexp.MustCompile(`(?i)(mean|median|mode|average|std|variance|sum\(|count\()`)},
		{"number-format", regexp.MustCompile(`(?i)(format.*number|number.*format|currency|decimal|precision)`)},

		// Date/Time
		{"date-parse", regexp.MustCompile(`(?i)(datetime\.parse|strptime|Date\.parse|parse.*date)`)},
		{"date-format", regexp.MustCompile(`(?i)(strftime|datetime\.strftime|format.*date|toISOString)`)},
		{"date-calc", regexp.MustCompile(`(?i)(timedelta|date.*add|date.*sub|days.*between|diff.*date)`)},
		{"timestamp", regexp.MustCompile(`(?i)(timestamp|unix.*time|epoch|time\(\))`)},

		// Array/List operations
		{"array-map", regexp.MustCompile(`(?i)(\.map\(|map\s*\(|for.*in)`)},
		{"array-filter", regexp.MustCompile(`(?i)(\.filter\(|filter\s*\(|where|select)`)},
		{"array-reduce", regexp.MustCompile(`(?i)(\.reduce\(|reduce\s*\(|fold|aggregate)`)},
		{"array-sort", regexp.MustCompile(`(?i)(\.sort\(|sorted\(|order.*by)`)},
		{"array-reverse", regexp.MustCompile(`(?i)(\.reverse\(|reversed\(|\[::-1\])`)},

		// CSV operations
		{"csv-parse", regexp.MustCompile(`(?i)(csv\.reader|csv\.DictReader|parse.*csv|from_csv)`)},
		{"csv-generate", regexp.MustCompile(`(?i)(csv\.writer|csv\.DictWriter|to_csv|generate.*csv)`)},

		// Network/URL
		{"http-request", regexp.MustCompile(`(?i)(requests\.|fetch\(|http\.|axios|urllib\.request)`)},
		{"url-parse", regexp.MustCompile(`(?i)(urlparse|URL\(|parse.*url|url.*parse)`)},
		{"query-string", regexp.MustCompile(`(?i)(parse_qs|query.*string|URLSearchParams)`)},

		// File operations
		{"file-read", regexp.MustCompile(`(?i)(open\(.*['\"]r|read.*file|fs\.read)`)},
		{"file-write", regexp.MustCompile(`(?i)(open\(.*['\"]w|write.*file|fs\.write)`)},

		// Crypto
		{"encrypt", regexp.MustCompile(`(?i)(encrypt|cipher|AES|RSA|Fernet)`)},
		{"decrypt", regexp.MustCompile(`(?i)(decrypt|decipher)`)},
		{"hmac", regexp.MustCompile(`(?i)(hmac|HMAC)`)},

		// Error handling
		{"try-catch", regexp.MustCompile(`(?i)(try\s*:|try\s*\{|except|catch)`)},
		{"raise-error", regexp.MustCompile(`(?i)(raise\s+|throw\s+|Error\()`)},

		// Determinism indicators
		{"pure-function", regexp.MustCompile(`(?i)(return\s+\w+[^=]*$|def\s+\w+\([^)]*\)\s*:\s*return)`)},
		{"no-state", regexp.MustCompile(`(?i)(def\s+\w+\([^)]*\)\s*:)`)},
	}

	for _, matcher := range patternMatchers {
		if matcher.pattern.MatchString(code) {
			patterns = append(patterns, matcher.name)
		}
	}

	return patterns
}

// analyzeSchemaTypes extracts types from a JSON schema
func (s *Service) analyzeSchemaTypes(schema map[string]any) []string {
	var types []string
	if schema == nil {
		return types
	}

	// Check for type field
	if t, ok := schema["type"]; ok {
		switch v := t.(type) {
		case string:
			types = append(types, v)
		case []any:
			for _, t := range v {
				if ts, ok := t.(string); ok {
					types = append(types, ts)
				}
			}
		}
	}

	// Check properties for object types
	if props, ok := schema["properties"].(map[string]any); ok {
		for _, prop := range props {
			if propMap, ok := prop.(map[string]any); ok {
				if t, ok := propMap["type"]; ok {
					if ts, ok := t.(string); ok {
						types = append(types, ts)
					}
				}
			}
		}
	}

	// Check items for array types
	if items, ok := schema["items"].(map[string]any); ok {
		if t, ok := items["type"]; ok {
			if ts, ok := t.(string); ok {
				types = append(types, "array<"+ts+">")
			}
		}
	}

	return types
}

// determineCategory determines the primary category based on analysis
func (s *Service) determineCategory(spec *FunctionSpec, codePatterns []string, inputTypes, outputTypes []string) (string, float64, string) {
	scores := make(map[string]float64)
	reasons := make(map[string][]string)

	// Score based on function name
	nameLower := strings.ToLower(spec.Name)
	for _, cat := range CategoryTaxonomy {
		for _, keyword := range cat.Keywords {
			if strings.Contains(nameLower, keyword) {
				scores[cat.ID] += 0.3
				reasons[cat.ID] = append(reasons[cat.ID], "name contains '"+keyword+"'")
			}
		}
	}

	// Score based on description
	descLower := strings.ToLower(spec.Description)
	for _, cat := range CategoryTaxonomy {
		for _, keyword := range cat.Keywords {
			if strings.Contains(descLower, keyword) {
				scores[cat.ID] += 0.2
				reasons[cat.ID] = append(reasons[cat.ID], "description contains '"+keyword+"'")
			}
		}
	}

	// Score based on code patterns
	patternCategoryMap := map[string]string{
		"string-concat":    "text-manipulation",
		"string-format":    "text-manipulation.formatting",
		"string-split":     "text-manipulation.parsing",
		"string-replace":   "text-manipulation",
		"string-trim":      "text-manipulation",
		"string-case":      "text-manipulation",
		"substring":        "text-manipulation",
		"base64-encode":    "text-manipulation.encoding",
		"base64-decode":    "text-manipulation.encoding",
		"url-encode":       "text-manipulation.encoding",
		"url-decode":       "text-manipulation.encoding",
		"html-encode":      "text-manipulation.encoding",
		"html-decode":      "text-manipulation.encoding",
		"json-parse":       "conversion.format",
		"json-stringify":   "conversion.format",
		"json-minify":      "conversion.format",
		"json-prettify":    "conversion.format",
		"hash-md5":         "cryptography.hashing",
		"hash-sha1":        "cryptography.hashing",
		"hash-sha256":      "cryptography.hashing",
		"hash-sha512":      "cryptography.hashing",
		"uuid-gen":         "generation.identifier",
		"random-id":        "generation.identifier",
		"email-validate":   "validation.email",
		"regex-match":      "validation",
		"type-check":       "validation.type",
		"math-basic":       "mathematical.arithmetic",
		"math-round":       "mathematical",
		"math-statistical": "mathematical.statistical",
		"number-format":    "mathematical.formatting",
		"date-parse":       "datetime.conversion",
		"date-format":      "datetime.formatting",
		"date-calc":        "datetime.calculation",
		"timestamp":        "datetime",
		"array-map":        "data-processing",
		"array-filter":     "data-processing.filtering",
		"array-reduce":     "data-processing.aggregation",
		"array-sort":       "data-processing.sorting",
		"array-reverse":    "data-processing",
		"csv-parse":        "conversion.format",
		"csv-generate":     "conversion.format",
		"http-request":     "network.http",
		"url-parse":        "network.url",
		"query-string":     "network.url",
		"encrypt":          "cryptography.encryption",
		"decrypt":          "cryptography.encryption",
		"hmac":             "cryptography",
	}

	for _, pattern := range codePatterns {
		if cat, ok := patternCategoryMap[pattern]; ok {
			scores[cat] += 0.4
			reasons[cat] = append(reasons[cat], "code pattern: "+pattern)
		}
	}

	// Score based on input/output types
	for _, t := range inputTypes {
		switch t {
		case "string":
			scores["text-manipulation"] += 0.15
		case "number", "integer":
			scores["mathematical"] += 0.15
		case "array":
			scores["data-processing"] += 0.15
		case "object":
			scores["conversion.format"] += 0.1
		}
	}

	// Find the best category
	var bestCategory string
	var bestScore float64
	for cat, score := range scores {
		if score > bestScore {
			bestScore = score
			bestCategory = cat
		}
	}

	// Default to utility if no clear winner
	if bestCategory == "" {
		bestCategory = "utility"
		bestScore = 0.5
	}

	// Calculate confidence (normalize to 0-1)
	confidence := bestScore
	if confidence > 1.0 {
		confidence = 1.0
	}

	// Build reasoning
	reasoning := "Auto-categorized based on: "
	if r, ok := reasons[bestCategory]; ok && len(r) > 0 {
		reasoning += strings.Join(r[:min(3, len(r))], ", ")
	} else {
		reasoning += "default classification"
	}

	return bestCategory, confidence, reasoning
}

// determineSecondaryCategory determines a secondary category if applicable
func (s *Service) determineSecondaryCategory(spec *FunctionSpec, codePatterns []string, primaryCategory string) (string, float64) {
	scores := make(map[string]float64)

	// Pattern-based secondary scoring
	patternCategoryMap := map[string]string{
		"json-parse":     "validation",
		"json-stringify": "validation",
		"try-catch":      "validation",
		"regex-match":    "validation",
		"hash-sha256":    "security",
		"encrypt":        "security",
	}

	for _, pattern := range codePatterns {
		if cat, ok := patternCategoryMap[pattern]; ok && cat != primaryCategory {
			scores[cat] += 0.3
		}
	}

	var bestCategory string
	var bestScore float64
	for cat, score := range scores {
		if score > bestScore && cat != primaryCategory {
			bestScore = score
			bestCategory = cat
		}
	}

	return bestCategory, bestScore
}

// extractTags extracts intelligent tags from the function analysis
func (s *Service) extractTags(spec *FunctionSpec, codePatterns []string, inputTypes, outputTypes []string) []string {
	tags := make(map[string]bool)

	// Add runtime tag
	switch strings.ToLower(spec.Runtime) {
	case "python", "python3.11", "python3.10", "python3.9":
		tags["python"] = true
	case "node", "nodejs", "nodejs20", "javascript", "js":
		tags["javascript"] = true
	case "wasm", "webassembly":
		tags["wasm"] = true
	}

	// Add I/O type tags
	for _, t := range inputTypes {
		switch t {
		case "string":
			tags["string-io"] = true
		case "number", "integer":
			tags["number-io"] = true
		case "object":
			tags["json-io"] = true
		case "array":
			tags["array-io"] = true
		}
	}

	for _, t := range outputTypes {
		switch t {
		case "string":
			tags["string-io"] = true
		case "number", "integer":
			tags["number-io"] = true
		case "object":
			tags["json-io"] = true
		case "array":
			tags["array-io"] = true
		}
	}

	// Add pattern-based tags
	for _, pattern := range codePatterns {
		switch pattern {
		case "pure-function", "no-state":
			tags["deterministic"] = true
			tags["ephemeral"] = true
		case "try-catch":
			tags["safe"] = true
		case "base64-encode", "base64-decode", "url-encode", "url-decode", "html-encode", "html-decode":
			tags["sanitized"] = true
		case "hash-sha256", "hash-sha512":
			tags["safe"] = true
		}
	}

	// Check for external calls
	hasExternalCalls := false
	externalPatterns := []string{"http-request", "fetch", "axios", "requests."}
	for _, pattern := range codePatterns {
		for _, ext := range externalPatterns {
			if pattern == ext {
				hasExternalCalls = true
				break
			}
		}
	}
	if !hasExternalCalls {
		tags["no-external-calls"] = true
	}

	// Convert to slice
	result := make([]string, 0, len(tags))
	for tag := range tags {
		if t := GetTagByID(tag); t != nil {
			result = append(result, t.Name)
		} else {
			result = append(result, tag)
		}
	}

	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
