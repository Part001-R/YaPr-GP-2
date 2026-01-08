package sqlitestor

import "errors"

var (

	// Нет содержимого в аргументе dsn
	EmptyDataArgumentDSN = errors.New("нет содержимого в аргументе dsn")

	// В аргументе db, нет указателя
	NilPtrArgumentDB = errors.New("в аргументе db, нет указателя")

	// Удаление не выполнено
	FaultDelete = errors.New("удаление не выполнено")
)
