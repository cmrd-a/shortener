package service

// OriginalExistError represents an error when trying to shorten a URL that already exists.
type OriginalExistError struct {
	Short string
}

// Error returns the short URL as the error message.
func (oe *OriginalExistError) Error() string {
	return oe.Short
}

// NewOriginalExistError creates a new OriginalExistError with the provided short URL.
func NewOriginalExistError(short string) error {
	return &OriginalExistError{
		Short: short,
	}
}
