// Статические ошибки пакета.
package udt

import "errors"

var (
	// В аргументе Conf нет указателя.
	NilPtrArgumentConf = errors.New("в аргументе conf нет указателя")

	// В аргументе L нет указателя.
	NilPtrArgumentL = errors.New(" В аргументе L нет указателя")

	// В аргументе A нет указателя.
	NilPtrArgumentA = errors.New(" В аргументе A нет указателя")

	// В аргументе F нет указателя.
	NilPtrArgumentF = errors.New(" В аргументе F нет указателя")

	// Нет указателя на логгер терминала.
	NilPtrLogger = errors.New("нет указателя на логгер терминала")

	// Нет указателя на логгер файла.
	NilPtrLoggerFile = errors.New("нет указателя на логгер файла")
)
