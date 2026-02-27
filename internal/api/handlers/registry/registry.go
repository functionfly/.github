// Package registry provides HTTP handlers for the function registry API.
//
// This package is organized into several focused files:
// - handlers.go: Main Handler struct and constructor
// - publish.go: Function publishing logic
// - execution.go: Function execution and replay logic
// - query.go: Function querying, listing, and search operations
// - stats.go: Statistics and ratings functionality
// - sdk.go: SDK code generation
// - utils.go: Utility functions and helpers
//
// This modular structure improves:
// - Code maintainability and readability
// - Testability through isolated concerns
// - Team collaboration with focused responsibilities
// - Easier debugging and feature development
package registry
