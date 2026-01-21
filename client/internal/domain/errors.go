// Статические ошибки пакета.
package domain

import "errors"

var (

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
)
