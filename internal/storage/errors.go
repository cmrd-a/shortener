package storage

import "fmt"

type OriginalExistError struct {
	Short string
}

func (oe *OriginalExistError) Error() string {
	return fmt.Sprintf("original is exist with short:%s", oe.Short)
}

func NewOriginalExistError(short string) error {
	return &OriginalExistError{
		Short: short,
	}
}
