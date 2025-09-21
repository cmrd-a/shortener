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

// IsOriginalExistError checks if the error is an OriginalExistError and returns it.
func IsOriginalExistError(err error, target **OriginalExistError) bool {
	if err == nil {
		return false
	}
	if origErr, ok := err.(*OriginalExistError); ok {
		if target != nil {
			*target = origErr
		}
		return true
	}
	return false
}
