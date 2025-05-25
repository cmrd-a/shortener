package service

type OriginalExistError struct {
	Short string
}

func (oe *OriginalExistError) Error() string {
	return oe.Short
}

func NewOriginalExistError(short string) error {
	return &OriginalExistError{
		Short: short,
	}
}
