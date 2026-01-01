package ui

import "errors"

var (
	// В аргументе Conf нет указателя.
	NilPtrArgumentConf = errors.New("в аргументе conf нет указателя")

	// В аргументе С нет указателя.
	NilPtrArgumentC = errors.New("в аргументе <c> нет указателя")

	// Отсутствуют данные токена.
	MissingTokenData = errors.New("в ответе на запрос ping, отсутствуют данные токена")

	// В аргументе subjectName, нет данных.
	EmptyDataArgumentSubjectName = errors.New("в аргументе <subjectName>, нет данных")

	// В аргументе fileName, нет данных.
	EmptyDataArgumentFileName = errors.New("в аргументе <fileName>, нет данных")

	// Длинна ключа не соответутсвует 32-м.
	LenSecretKey = errors.New("длина <secretKey> должна быть не менее 32 символов")

	// В аргументе tokenStr, нет данных.
	MissingDataArgumentTokenStr = errors.New("в аргументе tokenStr, нет данных")

	// Неизвестный метод подписи.
	SigningMethodUnknown = errors.New("неизвестный метод подписи")

	// Время валидности токена истекло
	ValidTokenExpired = errors.New("время валидности токена истекло")

	// В аргументе userName, нет данных
	MissingDataArgumentUserName = errors.New("в аргументе <userName>, нет данных")

	// В аргументе userPwd1, нет данных
	MissingDataArgumentUserPwd1 = errors.New("в аргументе <userPwd1>, нет данных")

	// В аргументе userPwd2, нет данных
	MissingDataArgumentUserPwd2 = errors.New("в аргументе <userPwd2>, нет данных")

	// Данные пароля не эквивалентны
	NotEqualPassword = errors.New("данные пароля не эквивалентны")

	// Некорректная длина зашифрованных данных
	NotCorrectLenData = errors.New("некорректная длина зашифрованных данных")

	// Некорректное заполнение
	NotCorrectDataFill = errors.New("некорректное заполнение")

	// Нет подтверждения
	NotConfirm = errors.New("нет подтверждения")
)
