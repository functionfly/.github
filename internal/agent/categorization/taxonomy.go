package categorization

// Category represents a function category in the taxonomy
type Category struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ParentID    *string  `json:"parent_id,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

// Tag represents an intelligent tag derived from function analysis
type Tag struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category,omitempty"` // Tag category (performance, security, etc.)
	AutoApply   bool     `json:"auto_apply"`         // Whether this tag can be auto-applied
	Keywords    []string `json:"keywords,omitempty"`
	Patterns    []string `json:"patterns,omitempty"` // Regex patterns for code matching
}

// Category Taxonomy - Hierarchical function categories
var CategoryTaxonomy = []Category{
	// Data Processing
	{
		ID:          "data-processing",
		Name:        "Data Processing",
		Description: "Functions that process, transform, or manipulate data",
		Keywords:    []string{"process", "transform", "convert", "parse", "data"},
	},
	{
		ID:          "data-processing.aggregation",
		Name:        "Data Aggregation",
		Description: "Functions that aggregate or summarize data",
		ParentID:    strPtr("data-processing"),
		Keywords:    []string{"aggregate", "sum", "average", "count", "group"},
	},
	{
		ID:          "data-processing.filtering",
		Name:        "Data Filtering",
		Description: "Functions that filter or select data",
		ParentID:    strPtr("data-processing"),
		Keywords:    []string{"filter", "select", "where", "exclude", "include"},
	},
	{
		ID:          "data-processing.sorting",
		Name:        "Data Sorting",
		Description: "Functions that sort or order data",
		ParentID:    strPtr("data-processing"),
		Keywords:    []string{"sort", "order", "rank", "arrange"},
	},

	// Text Manipulation
	{
		ID:          "text-manipulation",
		Name:        "Text Manipulation",
		Description: "Functions that manipulate or transform text",
		Keywords:    []string{"text", "string", "character", "substring"},
	},
	{
		ID:          "text-manipulation.formatting",
		Name:        "Text Formatting",
		Description: "Functions that format text output",
		ParentID:    strPtr("text-manipulation"),
		Keywords:    []string{"format", "pretty", "indent", "align"},
	},
	{
		ID:          "text-manipulation.encoding",
		Name:        "Text Encoding",
		Description: "Functions that encode or decode text",
		ParentID:    strPtr("text-manipulation"),
		Keywords:    []string{"encode", "decode", "base64", "url", "html"},
	},
	{
		ID:          "text-manipulation.parsing",
		Name:        "Text Parsing",
		Description: "Functions that parse text into structured data",
		ParentID:    strPtr("text-manipulation"),
		Keywords:    []string{"parse", "extract", "tokenize", "split"},
	},
	{
		ID:          "text-manipulation.truncation",
		Name:        "Text Truncation",
		Description: "Functions that truncate or shorten text",
		ParentID:    strPtr("text-manipulation"),
		Keywords:    []string{"truncate", "shorten", "limit", "cut", "ellipsis"},
	},

	// Conversion
	{
		ID:          "conversion",
		Name:        "Conversion",
		Description: "Functions that convert between formats or types",
		Keywords:    []string{"convert", "transform", "cast", "change"},
	},
	{
		ID:          "conversion.format",
		Name:        "Format Conversion",
		Description: "Functions that convert between data formats",
		ParentID:    strPtr("conversion"),
		Keywords:    []string{"json", "xml", "csv", "yaml", "toml"},
	},
	{
		ID:          "conversion.unit",
		Name:        "Unit Conversion",
		Description: "Functions that convert between units",
		ParentID:    strPtr("conversion"),
		Keywords:    []string{"unit", "measure", "temperature", "length", "weight"},
	},
	{
		ID:          "conversion.encoding",
		Name:        "Encoding Conversion",
		Description: "Functions that convert between encodings",
		ParentID:    strPtr("conversion"),
		Keywords:    []string{"encoding", "charset", "utf", "ascii"},
	},

	// Validation
	{
		ID:          "validation",
		Name:        "Validation",
		Description: "Functions that validate data or inputs",
		Keywords:    []string{"validate", "check", "verify", "assert"},
	},
	{
		ID:          "validation.email",
		Name:        "Email Validation",
		Description: "Functions that validate email addresses",
		ParentID:    strPtr("validation"),
		Keywords:    []string{"email", "mail", "address"},
	},
	{
		ID:          "validation.schema",
		Name:        "Schema Validation",
		Description: "Functions that validate data against schemas",
		ParentID:    strPtr("validation"),
		Keywords:    []string{"schema", "jsonschema", "validate", "structure"},
	},
	{
		ID:          "validation.type",
		Name:        "Type Validation",
		Description: "Functions that validate data types",
		ParentID:    strPtr("validation"),
		Keywords:    []string{"type", "typeof", "instanceof", "is"},
	},

	// Cryptography & Security
	{
		ID:          "cryptography",
		Name:        "Cryptography",
		Description: "Functions for cryptographic operations",
		Keywords:    []string{"crypto", "encrypt", "decrypt", "hash", "cipher"},
	},
	{
		ID:          "cryptography.hashing",
		Name:        "Hashing",
		Description: "Functions that generate hashes",
		ParentID:    strPtr("cryptography"),
		Keywords:    []string{"hash", "md5", "sha", "digest", "checksum"},
	},
	{
		ID:          "cryptography.encryption",
		Name:        "Encryption",
		Description: "Functions for encryption and decryption",
		ParentID:    strPtr("cryptography"),
		Keywords:    []string{"encrypt", "decrypt", "cipher", "aes", "rsa"},
	},

	// Generation
	{
		ID:          "generation",
		Name:        "Generation",
		Description: "Functions that generate data or content",
		Keywords:    []string{"generate", "create", "produce", "random"},
	},
	{
		ID:          "generation.identifier",
		Name:        "Identifier Generation",
		Description: "Functions that generate unique identifiers",
		ParentID:    strPtr("generation"),
		Keywords:    []string{"uuid", "guid", "id", "unique", "identifier"},
	},
	{
		ID:          "generation.random",
		Name:        "Random Generation",
		Description: "Functions that generate random values",
		ParentID:    strPtr("generation"),
		Keywords:    []string{"random", "rand", "shuffle", "pick"},
	},

	// Mathematical
	{
		ID:          "mathematical",
		Name:        "Mathematical",
		Description: "Functions for mathematical operations",
		Keywords:    []string{"math", "calculate", "compute", "arithmetic"},
	},
	{
		ID:          "mathematical.arithmetic",
		Name:        "Arithmetic",
		Description: "Basic arithmetic operations",
		ParentID:    strPtr("mathematical"),
		Keywords:    []string{"add", "subtract", "multiply", "divide", "sum"},
	},
	{
		ID:          "mathematical.statistical",
		Name:        "Statistical",
		Description: "Statistical calculations",
		ParentID:    strPtr("mathematical"),
		Keywords:    []string{"mean", "median", "mode", "variance", "stddev"},
	},
	{
		ID:          "mathematical.formatting",
		Name:        "Number Formatting",
		Description: "Functions that format numbers",
		ParentID:    strPtr("mathematical"),
		Keywords:    []string{"format", "number", "currency", "percentage", "decimal"},
	},

	// Date & Time
	{
		ID:          "datetime",
		Name:        "Date & Time",
		Description: "Functions for date and time operations",
		Keywords:    []string{"date", "time", "datetime", "timestamp", "timezone"},
	},
	{
		ID:          "datetime.formatting",
		Name:        "Date/Time Formatting",
		Description: "Functions that format dates and times",
		ParentID:    strPtr("datetime"),
		Keywords:    []string{"format", "strftime", "iso", "utc"},
	},
	{
		ID:          "datetime.conversion",
		Name:        "Date/Time Conversion",
		Description: "Functions that convert between date/time formats",
		ParentID:    strPtr("datetime"),
		Keywords:    []string{"convert", "timezone", "epoch", "unix"},
	},
	{
		ID:          "datetime.calculation",
		Name:        "Date/Time Calculation",
		Description: "Functions for date/time calculations",
		ParentID:    strPtr("datetime"),
		Keywords:    []string{"diff", "add", "subtract", "duration", "interval"},
	},

	// Network & API
	{
		ID:          "network",
		Name:        "Network",
		Description: "Functions for network operations",
		Keywords:    []string{"http", "request", "api", "url", "network"},
	},
	{
		ID:          "network.http",
		Name:        "HTTP Operations",
		Description: "Functions for HTTP requests and responses",
		ParentID:    strPtr("network"),
		Keywords:    []string{"http", "get", "post", "request", "response"},
	},
	{
		ID:          "network.url",
		Name:        "URL Operations",
		Description: "Functions for URL manipulation",
		ParentID:    strPtr("network"),
		Keywords:    []string{"url", "uri", "query", "param", "path"},
	},

	// Utility
	{
		ID:          "utility",
		Name:        "Utility",
		Description: "General utility functions",
		Keywords:    []string{"utility", "helper", "tool", "misc"},
	},
	{
		ID:          "utility.comparison",
		Name:        "Comparison",
		Description: "Functions for comparing values",
		ParentID:    strPtr("utility"),
		Keywords:    []string{"compare", "equal", "diff", "match"},
	},
	{
		ID:          "utility.debugging",
		Name:        "Debugging",
		Description: "Functions for debugging and logging",
		ParentID:    strPtr("utility"),
		Keywords:    []string{"debug", "log", "trace", "inspect"},
	},
}

// Tag Taxonomy - Intelligent tags for function classification
var TagTaxonomy = []Tag{
	// Input/Output Type Tags
	{ID: "io.string", Name: "string-io", Description: "Accepts or returns string input/output", Category: "io", AutoApply: true, Keywords: []string{"str", "string", "text"}},
	{ID: "io.number", Name: "number-io", Description: "Accepts or returns numeric input/output", Category: "io", AutoApply: true, Keywords: []string{"int", "float", "number", "decimal"}},
	{ID: "io.json", Name: "json-io", Description: "Accepts or returns JSON data", Category: "io", AutoApply: true, Keywords: []string{"json", "object", "dict"}},
	{ID: "io.array", Name: "array-io", Description: "Accepts or returns array/list data", Category: "io", AutoApply: true, Keywords: []string{"array", "list", "[]", "slice"}},
	{ID: "io.binary", Name: "binary-io", Description: "Accepts or returns binary data", Category: "io", AutoApply: true, Keywords: []string{"bytes", "binary", "buffer"}},
	{ID: "io.file", Name: "file-io", Description: "Accepts or returns file data", Category: "io", AutoApply: true, Keywords: []string{"file", "path", "read", "write"}},

	// Performance Tags
	{ID: "perf.deterministic", Name: "deterministic", Description: "Produces same output for same input", Category: "performance", AutoApply: true, Keywords: []string{"deterministic", "pure"}},
	{ID: "perf.idempotent", Name: "idempotent", Description: "Multiple calls produce same result", Category: "performance", AutoApply: true, Keywords: []string{"idempotent"}},
	{ID: "perf.fast", Name: "fast-execution", Description: "Executes in sub-millisecond time", Category: "performance", AutoApply: false, Keywords: []string{"fast", "quick", "instant"}},
	{ID: "perf.memory-efficient", Name: "memory-efficient", Description: "Low memory footprint", Category: "performance", AutoApply: false, Keywords: []string{"memory", "efficient"}},

	// Security Tags
	{ID: "sec.safe", Name: "safe", Description: "No security concerns", Category: "security", AutoApply: true, Keywords: []string{"safe"}},
	{ID: "sec.sanitized", Name: "sanitized", Description: "Output is sanitized", Category: "security", AutoApply: true, Keywords: []string{"sanitize", "escape", "clean"}},
	{ID: "sec.no-external", Name: "no-external-calls", Description: "Does not make external network calls", Category: "security", AutoApply: true, Keywords: []string{}},
	{ID: "sec.pii-safe", Name: "pii-safe", Description: "Does not expose PII", Category: "security", AutoApply: false, Keywords: []string{}},

	// Complexity Tags
	{ID: "complexity.simple", Name: "simple", Description: "Simple logic, easy to understand", Category: "complexity", AutoApply: true, Keywords: []string{}},
	{ID: "complexity.moderate", Name: "moderate", Description: "Moderate complexity", Category: "complexity", AutoApply: true, Keywords: []string{}},
	{ID: "complexity.complex", Name: "complex", Description: "Complex logic requiring careful review", Category: "complexity", AutoApply: true, Keywords: []string{}},

	// Use Case Tags
	{ID: "use.ai-ready", Name: "ai-ready", Description: "Suitable for AI agent consumption", Category: "use-case", AutoApply: true, Keywords: []string{}},
	{ID: "use.api-friendly", Name: "api-friendly", Description: "Ideal for API integration", Category: "use-case", AutoApply: true, Keywords: []string{}},
	{ID: "use.ephemeral", Name: "ephemeral", Description: "No persistent state required", Category: "use-case", AutoApply: true, Keywords: []string{}},

	// Runtime Tags
	{ID: "runtime.python", Name: "python", Description: "Python runtime function", Category: "runtime", AutoApply: true, Keywords: []string{"python", "py"}},
	{ID: "runtime.javascript", Name: "javascript", Description: "JavaScript runtime function", Category: "runtime", AutoApply: true, Keywords: []string{"javascript", "js", "node"}},
	{ID: "runtime.wasm", Name: "wasm", Description: "WebAssembly function", Category: "runtime", AutoApply: true, Keywords: []string{"wasm", "webassembly"}},
}

// strPtr is a helper to get a string pointer
func strPtr(s string) *string {
	return &s
}

// GetCategoryByID retrieves a category by its ID
func GetCategoryByID(id string) *Category {
	for _, cat := range CategoryTaxonomy {
		if cat.ID == id {
			return &cat
		}
	}
	return nil
}

// GetTagByID retrieves a tag by its ID
func GetTagByID(id string) *Tag {
	for _, tag := range TagTaxonomy {
		if tag.ID == id {
			return &tag
		}
	}
	return nil
}

// GetRootCategories returns all top-level categories (no parent)
func GetRootCategories() []Category {
	var roots []Category
	for _, cat := range CategoryTaxonomy {
		if cat.ParentID == nil {
			roots = append(roots, cat)
		}
	}
	return roots
}

// GetSubCategories returns all sub-categories of a parent category
func GetSubCategories(parentID string) []Category {
	var subs []Category
	for _, cat := range CategoryTaxonomy {
		if cat.ParentID != nil && *cat.ParentID == parentID {
			subs = append(subs, cat)
		}
	}
	return subs
}

// GetAutoApplyTags returns all tags that can be auto-applied
func GetAutoApplyTags() []Tag {
	var tags []Tag
	for _, tag := range TagTaxonomy {
		if tag.AutoApply {
			tags = append(tags, tag)
		}
	}
	return tags
}

// GetTagsByCategory returns all tags in a specific category
func GetTagsByCategory(category string) []Tag {
	var tags []Tag
	for _, tag := range TagTaxonomy {
		if tag.Category == category {
			tags = append(tags, tag)
		}
	}
	return tags
}
