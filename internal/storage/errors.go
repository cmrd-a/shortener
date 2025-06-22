package storage

import (
	"errors"
	"fmt"
)

type ErrOriginalExist struct {
	Short string
}

func (oe *ErrOriginalExist) Error() string {
	return fmt.Sprintf("original is exist with short:%s", oe.Short)
}

func NewOriginalExistError(short string) error {
	return &ErrOriginalExist{
		Short: short,
	}
}

var ErrURLIsDeleted = errors.New("url is deleted")
