// Статические ошибки пакета.
package logfile

import "errors"

var (
	// Нет содержимого в названии файла логов
	EmptyDataArgumentNameLogFile = errors.New("нет содержимого в названии файла логов")

	// Нет указателя на логгер"
	NilPtrLogger = errors.New("Нет указателя на логгер")

	// Нет ожидаемой размерности
	ErrVolume = errors.New("Нет ожидаемой размерности")

	// Нет содержимого
	ErrEmpty = errors.New("Нет содержимого")
)
