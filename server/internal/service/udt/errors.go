// Статические ошибки пакета.
package udt

import "errors"

var (
	// В аргументе Conf нет указателя.
	NilPtrArgumentConf = errors.New("в аргументе conf нет указателя")

	// Нет указателя на логгер.
	NilPtrLogger = errors.New("нет указателя на логгер")

	// Нет указателя на gRPC.
	NilPtrGRPC = errors.New("нет указателя на gRPC")
)
