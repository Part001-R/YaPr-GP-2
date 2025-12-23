package udt

import "errors"

var (
	// В аргументе Conf нет указателя.
	NilPtrArgumentConf = errors.New("в аргументе conf нет указателя")

	// Нет указателя на логгер терминала.
	NilPtrLogger = errors.New("нет указателя на логгер терминала")

	// Нет указателя на логгер файла.
	NilPtrLoggerFile = errors.New("нет указателя на логгер файла")
)
