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

	// В аргументе <resp>, нет указателя
	NilPtrArgumentResp = errors.New("в аргументе <resp>, нет указателя")

	// В аргументе txFileName, нет данных.
	EmptyDataArgumentTxFileName = errors.New("в аргументе <txFileName>, нет данных")

	// В аргументе txFileHash, нет данных.
	EmptyDataArgumentTxFileHash = errors.New("в аргументе <txFileHash>, нет данных")

	// В аргументе rxFileHash, нет данных.
	EmptyDataArgumentRxFileHash = errors.New("в аргументе <rxFileHash>, нет данных")

	// В аргументе srcFileHash, нет данных.
	EmptyDataArgumentSrcFileHash = errors.New("в аргументе <srcFileHash>, нет данных")

	// В аргументе rxToken, нет данных.
	EmptyDataArgumentRxToken = errors.New("в аргументе <rxToken>, нет данных")

	// В аргументе secretKey, нет данных.
	EmptyDataArgumentSecretKey = errors.New("в аргументе <secretKey>, нет данных")

	// В аргументе fileName, нет данных.
	EmptyDataArgumentFileName = errors.New("в аргументе <fileName>, нет данных")

	// В аргументе txToken, нет данных.
	EmptyDataArgumentTxToken = errors.New("в аргументе <txToken>, нет данных")

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

	// Нет содержимого
	EmptyData = errors.New("нет содержимого")

	// Некорректные данные.
	IncorrectData = errors.New("некорретные данные")

	// Не удалось извлечь метаданные из контекста ответа
	ErrMetadata = errors.New("не удалось извлечь метаданные из контекста ответа")

	// Токены не эквивалентны
	NotEqualTokens = errors.New("nокены не эквивалентны")
)
