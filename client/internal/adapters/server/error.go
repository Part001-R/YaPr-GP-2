package server

import "errors"

var (
	// В аргументе ip, нет содержимого
	NilPtrArgumentIP = errors.New("В аргументе ip, нет содержимого")

	// В аргументе port, нет содержимого
	NilPtrArgumentPort = errors.New("В аргументе port, нет содержимого")

	// В аргументе tokenStr, нет данных.
	MissingDataArgumentTokenStr = errors.New("в аргументе tokenStr, нет данных")

	// Длинна ключа не соответутсвует 32-м.
	LenSecretKey = errors.New("длина <secretKey> должна быть не менее 32 символов")

	// Неизвестный метод подписи.
	SigningMethodUnknown = errors.New("неизвестный метод подписи")

	// Время валидности токена истекло
	ValidTokenExpired = errors.New("время валидности токена истекло")

	// В аргументе subjectName, нет данных.
	EmptyDataArgumentSubjectName = errors.New("в аргументе <subjectName>, нет данных")

	// В аргументе userName, нет данных.
	EmptyDataArgumentUserName = errors.New("в аргументе <userName>, нет данных")

	// В аргументе userPwd, нет данных.
	EmptyDataArgumentUserPwd = errors.New("в аргументе <userPwd>, нет данных")

	// Нет указателя connect
	NilPtrConnect = errors.New("Нет указателя connect")

	// Отсутствуют данные токена.
	MissingTokenData = errors.New("Отсутствуют данные токена")

	// Отсутствуют данные токена от сервера.
	MissingTokenSrvData = errors.New("Отсутствуют данные токена от сервера")

	// В аргументе txID, нет данных.
	EmptyDataArgumentTxID = errors.New("в аргументе <txID>, нет данных")

	// В аргументе txFor, нет данных.
	EmptyDataArgumentTxFor = errors.New("в аргументе <txFor>, нет данных")

	// В аргументе txLogin, нет данных.
	EmptyDataArgumentTxLogin = errors.New("в аргументе <txLogin>, нет данных")

	// В аргументе txText, нет данных.
	EmptyDataArgumentTxText = errors.New("в аргументе <txText>, нет данных")

	// Некорректная длина зашифрованных данных
	NotCorrectLenData = errors.New("некорректная длина зашифрованных данных")

	// Некорректное заполнение
	NotCorrectDataFill = errors.New("некорректное заполнение")
)
