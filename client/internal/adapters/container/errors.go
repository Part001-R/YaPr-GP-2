// Статические ошибки пакета.
package container

import "errors"

var (

	// Возникла внешняя ошибка
	ExternalErr = errors.New("Возникла внешняя ошибка")
)
