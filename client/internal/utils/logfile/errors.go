// Статические ошибки пакета.
package logfile

import "errors"

var (
	// Нет содержимого в названии файла логов
	EmptyDataArgumentNameLogFile = errors.New("нет содержимого в названии файла логов")

	// Нет содержимого в аргументе appName
	EmptyDataArgumentAppName = errors.New("Нет содержимого в аргументе appName")

	// Нет содержимого в аргументе projectName
	EmptyDataArgumentProjectName = errors.New("Нет содержимого в аргументе projectName")

	// Нет содержимого в аргументе fullPath
	EmptyDataArgumentFullPath = errors.New("Нет содержимого в аргументе fullPath")

	// Нет указателя на логгер"
	NilPtrLogger = errors.New("Нет указателя на логгер")

	// Нет ожидаемой размерности
	ErrVolume = errors.New("Нет ожидаемой размерности")

	// Нет содержимого
	ErrEmpty = errors.New("Нет содержимого")
)
