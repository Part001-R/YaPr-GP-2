// Статические ошибки пакета.
package udt

import "errors"

var (
	// В аргументе Conf нет указателя.
	NilPtrArgumentConf = errors.New("в аргументе conf нет указателя")

	// Нет указателя на логгер.
	NilPtrLogger = errors.New("нет указателя на логгер")

	// Нет указателя на сервер.
	NilPtrServer = errors.New("нет указателя на сервер")

	// Нет указателя на хранилище.
	NilPtrStorage = errors.New("нет указателя на хранилище")

	// Нет указателя на флаги.
	NilPtrFlag = errors.New("нет указателя на флаги")

	// Нет пути к файлу приватного ключа.
	MissingPathPrivate = errors.New("нет пути к файлу приватного ключа")

	// Нет пути к файлу сертификата.
	MissingPathSert = errors.New("нет пути к файлу сертификата")

	// Нет указателя на gRPC.
	NilPtrGRPC = errors.New("нет указателя на gRPC")

	// нет указания порта".
	EmptyDataPort = errors.New("нет указания порта")
)
