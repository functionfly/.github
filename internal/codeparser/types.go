package codeparser

import (
	"time"
)

type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	HasDefault  bool   `json:"has_default"`
	DefaultValue string `json:"default_value,omitempty"`
}

type ParsedFunction struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Language    string     `json:"language"`
	Signature   string     `json:"signature"`
	Parameters  []Parameter `json:"parameters"`
	ReturnType  string     `json:"return_type,omitempty"`
	Docstring   string     `json:"docstring,omitempty"`
	Code        string     `json:"code"`
	StartLine   int        `json:"start_line"`
	EndLine     int        `json:"end_line"`
}

type ParseResult struct {
	Language        string          `json:"language"`
	Confidence      float64         `json:"confidence"`
	Functions       []ParsedFunction `json:"functions"`
	RawCode         string          `json:"raw_code"`
	RawCodeLength   int             `json:"raw_code_length"`
	DetectedAt      time.Time       `json:"detected_at"`
}

type ParseRequest struct {
	Code         string `json:"code" validate:"required,max=102400"`
	ForceLanguage string `json:"force_language,omitempty"`
}

type CreateFunctionsRequest struct {
	Functions []CreateFunctionInput `json:"functions" validate:"required,min=1,max=50"`
	Visibility string `json:"visibility" validate:"required,oneof=private public"`
	Providers []string `json:"providers,omitempty"`
	Region    string `json:"region,omitempty"`
	Author    string `json:"author,omitempty"`
	Changelog string `json:"changelog,omitempty"`
}

type CreateFunctionInput struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Code     string `json:"code" validate:"required"`
	Language string `json:"language" validate:"required"`
}

type CreateFunctionsResponse struct {
	Created []CreatedFunction `json:"created"`
	Failed  []FailedFunction   `json:"failed,omitempty"`
}

type CreatedFunction struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type FailedFunction struct {
	Name   string `json:"name"`
	Error  string `json:"error"`
}

type ImportError struct {
	FunctionName string `json:"function_name"`
	Error        string `json:"error"`
}