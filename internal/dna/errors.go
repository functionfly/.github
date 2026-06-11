package dna

import "errors"

var (
	ErrAccessDenied           = errors.New("access denied")
	ErrMutationNotFound       = errors.New("mutation not found")
	ErrMutationNotProposed    = errors.New("mutation is not in proposed status")
	ErrInsufficientCredits    = errors.New("insufficient credits")
	ErrMutationNotRollbackable = errors.New("mutation is not in a rollback-eligible status")
)