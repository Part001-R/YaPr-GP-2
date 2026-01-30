// Статические ошибки пакета.
package flags

import "errors"

var (
	// нет указателя в f
	ErrNilPtrArgumentF = errors.New("нет указателя в f")
)
