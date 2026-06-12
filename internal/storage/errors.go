package storage

import (
	"errors"

	"github.com/lib/pq"
)

var (
	ErrDuplicateKey       = errors.New("duplicate key violation")
	ErrUniqueViolation    = errors.New("unique constraint violation")
	ErrRecordNotFound     = errors.New("record not found")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

func ClassifyError(err error) error {
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return ErrDuplicateKey
		case "23503":
			return ErrForeignKeyViolation
		case "23502":
			return ErrRecordNotFound
		}
	}

	if errors.Is(err, ErrRecordNotFound) ||
	   errors.Is(err, ErrDuplicateKey) ||
	   errors.Is(err, ErrForeignKeyViolation) {
		return err
	}

	return err
}

func IsDuplicateKeyError(err error) bool {
	return errors.Is(ClassifyError(err), ErrDuplicateKey)
}

func IsNotFoundError(err error) bool {
	return errors.Is(ClassifyError(err), ErrRecordNotFound)
}

func IsForeignKeyError(err error) bool {
	return errors.Is(ClassifyError(err), ErrForeignKeyViolation)
}