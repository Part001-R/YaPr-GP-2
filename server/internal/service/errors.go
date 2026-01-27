// Статические ошибки пакета.
package service

import "errors"

var (
	// В аргументе Conf нет указателя.
	NilPtrArgumentConf = errors.New("в аргументе conf нет указателя")

	// Нет указателя на логгер.
	NilPtrLogger = errors.New("нет указателя на логгер")
)
