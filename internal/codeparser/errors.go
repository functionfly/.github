package codeparser

import "errors"

var (
	ErrEmptyCode      = errors.New("code cannot be empty")
	ErrCodeTooLarge   = errors.New("code exceeds maximum size of 100KB")
	ErrInvalidLanguage = errors.New("unsupported language")
	ErrParseFailed    = errors.New("failed to parse code")
)