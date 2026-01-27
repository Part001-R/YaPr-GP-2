// Статические ошибки пакета.
package grpc

import "errors"

var (
	// Не найдено
	ErrNotFound = errors.New("не найдено")

	// Некорректные данные
	ErrNotCorrectData = errors.New("некорректные данные")

	// Отсутствуют метаданные
	ErrMissingMetadata = errors.New("Отсутствуют метаданные")

	// Отсутствует токен
	ErrMissingToken = errors.New("Отсутствует токен")

	// В токене нет содержимого"
	ErrIsEmptyToken = errors.New("В токене нет содержимого")

	// Ошибка отправки заголовков
	ErrSendHeaders = errors.New("Ошибка отправки заголовков")

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

	// нет данных idClient.
	EmptyDataIDClient = errors.New("нет данных <idClient>")

	// нет данных fileName.
	EmptyDataFileName = errors.New("нет данных <fileName>")

	// нет данных name.
	EmptyDataName = errors.New("нет данных <name>")

	// В аргументе Conf нет указателя.
	NilPtrArgumentConf = errors.New("в аргументе conf нет указателя")

	// Отсутствуют данные.
	MissingData = errors.New("отсутствуют данные")

	// Некоректный номер.
	IncorrecrNumb = errors.New("Некоректный номер")

	// Нет указателя в аргументе l.
	NilPtrArgumentL = errors.New("Нет указателя в аргументе l")

	// Нет указателя в аргументе s.
	NilPtrArgumentS = errors.New("Нет указателя в аргументе s")

	// Нет указателя в аргументе f.
	NilPtrArgumentF = errors.New("Нет указателя в аргументе f")

	// Нет указателя в аргументе ctx.
	NilPtrArgumentCtx = errors.New("Нет указателя в аргументе ctx")

	// Нет указателя в аргументе Empty.
	NilPtrArgumentEmpty = errors.New("Нет указателя в аргументе empty")

	// Нет указателя в аргументе stream.
	NilPtrArgumentStream = errors.New("Нет указателя в аргументе stream")

	// Нет указателя в аргументе req.
	NilPtrArgumentReq = errors.New("Нет указателя в аргументе req")

	// Некорректный код статуса.
	IncorrectStage = errors.New("Не корректный код статуса")

	// Нет указателя на логгер.
	NilPtrLogger = errors.New("Нет указателя на логгер")
)
