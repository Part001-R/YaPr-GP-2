package flags

import "errors"

var (
	// нет указателя в f
	ErrNilPtrArgumentF = errors.New("нет указателя в f")

	// нет содержимого dsn
	ErrEmptyDSN = errors.New("нет содержимого dsn")
)
