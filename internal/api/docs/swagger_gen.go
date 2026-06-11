package docs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Paths      map[string]PathItem    `json:"paths"`
	Components Components             `json:"components"`
	Servers    []Server               `json:"servers,omitempty"`
	Tags       []Tag                  `json:"tags,omitempty"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
	Contact     *Contact `json:"contact,omitempty"`
	License     *License `json:"license,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type Server struct {
	URL         string            `json:"url"`
	Description string            `json:"description,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type PathItem struct {
	Summary     string         `json:"summary,omitempty"`
	Description string         `json:"description,omitempty"`
	Get         *Operation     `json:"get,omitempty"`
	Post        *Operation     `json:"post,omitempty"`
	Put         *Operation     `json:"put,omitempty"`
	Delete      *Operation     `json:"delete,omitempty"`
	Patch       *Operation     `json:"patch,omitempty"`
	Parameters  []Parameter    `json:"parameters,omitempty"`
}

type Operation struct {
	Tags        []string               `json:"tags,omitempty"`
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	OperationID string                 `json:"operationId,omitempty"`
	Parameters  []Parameter            `json:"parameters,omitempty"`
	RequestBody *RequestBody           `json:"requestBody,omitempty"`
	Responses   map[string]Response    `json:"responses"`
	Deprecated  bool                   `json:"deprecated,omitempty"`
	Security    []map[string][]string  `json:"security,omitempty"`
}

type Parameter struct {
	Name        string   `json:"name"`
	In          string   `json:"in"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Schema      Schema   `json:"schema"`
	Example     interface{} `json:"example,omitempty"`
}

type RequestBody struct {
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Response struct {
	Description string                 `json:"description"`
	Content     map[string]MediaType   `json:"content,omitempty"`
	Headers     map[string]Header      `json:"headers,omitempty"`
}

type Header struct {
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema"`
}

type Components struct {
	Schemas     map[string]Schema `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type        string `json:"type"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Scheme      string `json:"scheme,omitempty"`
	Description string `json:"description,omitempty"`
}

type Schema struct {
	Type        string        `json:"type,omitempty"`
	Format      string        `json:"format,omitempty"`
	Description string        `json:"description,omitempty"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Items       *Schema       `json:"items,omitempty"`
	Required    []string      `json:"required,omitempty"`
	Enum        []interface{} `json:"enum,omitempty"`
	Example     interface{}    `json:"example,omitempty"`
	Ref         string        `json:"$ref,omitempty"`
}

type SwaggerGenerator struct {
	spec       *OpenAPISpec
	host       string
	basePath   string
}

func NewSwaggerGenerator(title, version, host, basePath string) *SwaggerGenerator {
	return &SwaggerGenerator{
		spec: &OpenAPISpec{
			OpenAPI: "3.0.0",
			Info: OpenAPIInfo{
				Title:       title,
				Description: "API documentation for FunctionFly",
				Version:     version,
			},
			Paths:      make(map[string]PathItem),
			Components: Components{
				Schemas:         make(map[string]Schema),
				SecuritySchemes: make(map[string]SecurityScheme),
			},
		},
		host:     host,
		basePath: basePath,
	}
}

func (g *SwaggerGenerator) AddSecurityScheme(name, schemeType, description string) {
	g.spec.Components.SecuritySchemes[name] = SecurityScheme{
		Type:        schemeType,
		Description: description,
		BearerFormat: "JWT",
	}
}

func (g *SwaggerGenerator) AddPath(path string, method string, op *Operation) {
	if _, exists := g.spec.Paths[path]; !exists {
		g.spec.Paths[path] = PathItem{}
	}

	switch strings.ToUpper(method) {
	case "GET":
		g.spec.Paths[path].Get = op
	case "POST":
		g.spec.Paths[path].Post = op
	case "PUT":
		g.spec.Paths[path].Put = op
	case "DELETE":
		g.spec.Paths[path].Delete = op
	case "PATCH":
		g.spec.Paths[path].Patch = op
	}
}

func (g *SwaggerGenerator) AddSchema(name string, schema Schema) {
	if g.spec.Components.Schemas == nil {
		g.spec.Components.Schemas = make(map[string]Schema)
	}
	g.spec.Components.Schemas[name] = schema
}

func (g *SwaggerGenerator) AddResponse(path, method string, statusCode string, response Response) {
	var op *Operation
	pathItem, exists := g.spec.Paths[path]
	if !exists {
		return
	}

	switch strings.ToUpper(method) {
	case "GET":
		op = pathItem.Get
	case "POST":
		op = pathItem.Post
	case "PUT":
		op = pathItem.Put
	case "DELETE":
		op = pathItem.Delete
	case "PATCH":
		op = pathItem.Patch
	}

	if op == nil {
		return
	}

	if op.Responses == nil {
		op.Responses = make(map[string]Response)
	}
	op.Responses[statusCode] = response
}

func (g *SwaggerGenerator) Generate() *OpenAPISpec {
	return g.spec
}

func (g *SwaggerGenerator) ToJSON() ([]byte, error) {
	return json.MarshalIndent(g.spec, "", "  ")
}

func SchemaFromType(v interface{}) Schema {
	t := reflect.TypeOf(v)
	return Schema{Type: mapGoTypeToOpenAPI(t.Kind().String())}
}

func mapGoTypeToOpenAPI(kind string) string {
	switch kind {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "array", "slice":
		return "array"
	case "object", "struct":
		return "object"
	default:
		return "string"
	}
}

func (g *SwaggerGenerator) AddTag(name, description string) {
	g.spec.Tags = append(g.spec.Tags, Tag{Name: name, Description: description})
}

func (g *SwaggerGenerator) SetServer(url, description string) {
	g.spec.Servers = append(g.spec.Servers, Server{
		URL:         url,
		Description: description,
	})
}

type EndpointDoc struct {
	Path        string
	Method      string
	Summary     string
	Description string
	OperationID string
	Tags        []string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   map[string]Response
	Deprecated  bool
	Security    []map[string][]string
}

func (g *SwaggerGenerator) RegisterEndpoint(doc EndpointDoc) {
	op := &Operation{
		Summary:     doc.Summary,
		Description: doc.Description,
		OperationID: doc.OperationID,
		Tags:        doc.Tags,
		Parameters:  doc.Parameters,
		RequestBody: doc.RequestBody,
		Responses:   doc.Responses,
		Deprecated:  doc.Deprecated,
		Security:    doc.Security,
	}

	g.AddPath(doc.Path, doc.Method, op)
}

func CreateSuccessResponse(description string, schema Schema) Response {
	return Response{
		Description: description,
		Content: map[string]MediaType{
			"application/json": {Schema: schema},
		},
	}
}

func CreateErrorResponse(statusCode, description string) Response {
	return Response{
		Description: description,
		Content: map[string]MediaType{
			"application/json": {
				Schema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"code":    {Type: "string", Example: "ERROR_CODE"},
						"message": {Type: "string", Example: "Error message"},
						"detail":  {Type: "string", Example: "Additional details"},
						"request_id": {Type: "string", Example: "uuid"},
						"field":   {Type: "string", Example: "field_name"},
					},
				},
			},
		},
	}
}

type ValidationErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func CreateValidationErrorResponse() Response {
	return Response{
		Description: "Validation error",
		Content: map[string]MediaType{
			"application/json": {
				Schema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"code":    {Type: "string", Example: "VALIDATION_ERROR"},
						"message": {Type: "string", Example: "Validation failed"},
						"field":   {Type: "string", Example: "field_name"},
					},
				},
			},
		},
	}
}

func PaginationParameters() []Parameter {
	return []Parameter{
		{
			Name:     "limit",
			In:       "query",
			Description: "Maximum number of items to return (default 20, max 100)",
			Required:    false,
			Schema:      Schema{Type: "integer", Format: "int32", Example: 20},
		},
		{
			Name:     "offset",
			In:       "query",
			Description: "Number of items to skip before returning results",
			Required:    false,
			Schema:      Schema{Type: "integer", Format: "int32", Example: 0},
		},
		{
			Name:     "cursor",
			In:       "query",
			Description: "Cursor for pagination (alternative to offset)",
			Required:    false,
			Schema:      Schema{Type: "string"},
		},
	}
}

func SortParameters(allowedFields []string) []Parameter {
	return []Parameter{
		{
			Name:        "sort_by",
			In:          "query",
			Description: fmt.Sprintf("Field to sort by (allowed: %s)", strings.Join(allowedFields, ", ")),
			Required:    false,
			Schema:      Schema{Type: "string"},
		},
		{
			Name:        "sort_dir",
			In:          "query",
			Description: "Sort direction (asc or desc)",
			Required:    false,
			Schema:      Schema{Type: "string", Enum: []interface{}{"asc", "desc"}},
		},
	}
}

type PaginationResponseSchema struct {
	Data       interface{} `json:"data"`
	Total      int64        `json:"total"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
	HasMore    bool         `json:"has_more"`
	NextCursor string       `json:"next_cursor,omitempty"`
}

func (p PaginationResponseSchema) ToSchema(itemSchema Schema) Schema {
	return Schema{
		Type: "object",
		Properties: map[string]Schema{
			"data":       Schema{Type: "array", Items: &itemSchema},
			"total":      Schema{Type: "integer", Format: "int64"},
			"limit":      Schema{Type: "integer", Format: "int32"},
			"offset":     Schema{Type: "integer", Format: "int32"},
			"has_more":   Schema{Type: "boolean"},
			"next_cursor": Schema{Type: "string"},
		},
	}
}

type APIErrorSchema struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Detail    string `json:"detail,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Field     string `json:"field,omitempty"`
}

func APIErrorSchemaRef() string {
	return "#/components/schemas/APIError"
}

func (g *SwaggerGenerator) AddAPIErrorSchema() {
	g.AddSchema("APIError", Schema{
		Type: "object",
		Properties: map[string]Schema{
			"code":       {Type: "string", Description: "Machine-readable error code"},
			"message":    {Type: "string", Description: "Human-readable error message"},
			"detail":     {Type: "string", Description: "Additional context (optional)"},
			"request_id": {Type: "string", Description: "Request ID for tracing"},
			"field":      {Type: "string", Description: "Field that failed validation"},
		},
		Required: []string{"code", "message"},
	})
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func CreateUnauthorizedResponse() Response {
	return Response{
		Description: "Unauthorized - Authentication required",
		Content: map[string]MediaType{
			"application/json": {
				Schema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"code":    {Type: "string", Example: "UNAUTHORIZED"},
						"message": {Type: "string", Example: "Authentication required"},
					},
				},
			},
		},
	}
}

func CreateForbiddenResponse() Response {
	return Response{
		Description: "Forbidden - Insufficient permissions",
		Content: map[string]MediaType{
			"application/json": {
				Schema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"code":    {Type: "string", Example: "FORBIDDEN"},
						"message": {Type: "string", Example: "Insufficient permissions"},
					},
				},
			},
		},
	}
}

func CreateNotFoundResponse(resource string) Response {
	return Response{
		Description: fmt.Sprintf("%s not found", resource),
		Content: map[string]MediaType{
			"application/json": {
				Schema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"code":    {Type: "string", Example: "NOT_FOUND"},
						"message": {Type: "string", Example: fmt.Sprintf("%s not found", resource)},
					},
				},
			},
		},
	}
}

func CreateRateLimitedResponse() Response {
	return Response{
		Description: "Rate limit exceeded",
		Headers: map[string]Header{
			"Retry-After": {
				Description: "Seconds to wait before retrying",
				Schema:      Schema{Type: "integer"},
			},
			"X-RateLimit-Limit": {
				Description: "Rate limit ceiling",
				Schema:      Schema{Type: "integer"},
			},
			"X-RateLimit-Remaining": {
				Description: "Requests remaining in window",
				Schema:      Schema{Type: "integer"},
			},
			"X-RateLimit-Reset": {
				Description: "Unix timestamp when rate limit resets",
				Schema:      Schema{Type: "integer"},
			},
		},
		Content: map[string]MediaType{
			"application/json": {
				Schema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"code":    {Type: "string", Example: "RATE_LIMITED"},
						"message": {Type: "string", Example: "Rate limit exceeded"},
					},
				},
			},
		},
	}
}

func CreatedAt() time.Time {
	return time.Now()
}