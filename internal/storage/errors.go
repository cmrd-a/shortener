package storage

import (
	"errors"
	"fmt"
)

// ErrOriginalExist represents an error when trying to add a URL that already exists in storage.
type ErrOriginalExist struct {
	Short string
}

// Error returns a formatted error message indicating the original URL already exists.
func (oe *ErrOriginalExist) Error() string {
	return fmt.Sprintf("original is exist with short:%s", oe.Short)
}

// NewOriginalExistError creates a new ErrOriginalExist error with the provided short URL.
func NewOriginalExistError(short string) error {
	return &ErrOriginalExist{
		Short: short,
	}
}

// ErrURLIsDeleted is returned when attempting to access a URL that has been marked as deleted.
var ErrURLIsDeleted = errors.New("url is deleted")
