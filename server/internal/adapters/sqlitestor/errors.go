// Статические ошибки пакета.
package sqlitestor

import "errors"

var (

	// Отсутствуют данные.
	MissingData = errors.New("отсутствуют данные")

	// Нет содержимого в аргументе dsn
	EmptyDataArgumentDSN = errors.New("нет содержимого в аргументе dsn")

	// В аргументе db, нет указателя
	NilPtrArgumentDB = errors.New("в аргументе db, нет указателя")

	// Удаление не выполнено
	FaultDelete = errors.New("удаление не выполнено")

	// Нет содержимого в аргументе field1
	EmptyDataArgumentField1 = errors.New("нет содержимого в аргументе field1")

	// Нет содержимого в аргументе field2
	EmptyDataArgumentField2 = errors.New("нет содержимого в аргументе field2")

	// Нет содержимого в аргументе field3
	EmptyDataArgumentField3 = errors.New("нет содержимого в аргументе field3")

	// Нет содержимого в аргументе field4
	EmptyDataArgumentField4 = errors.New("нет содержимого в аргументе field4")

	// Нет содержимого в аргументе field5
	EmptyDataArgumentField5 = errors.New("нет содержимого в аргументе field5")

	// Нет содержимого в аргументе createdAt
	EmptyDataArgumentCreatedAt = errors.New("нет содержимого в аргументе createdAt")

	// Нет содержимого в аргументе name
	EmptyDataArgumentName = errors.New("нет содержимого в аргументе name")

	// Нет содержимого в аргументе pwd
	EmptyDataArgumentPwd = errors.New("нет содержимого в аргументе pwd")

	// Нет содержимого в аргументе ctx
	EmptyDataArgumentCtx = errors.New("нет содержимого в аргументе ctx")

	// Нет указателя на БД.
	NilPtrDB = errors.New("нет указателя на БД")

	// Нет указателя на мьютексы.
	NilPtrMutex = errors.New("нет указателя на мьютексы")
)
