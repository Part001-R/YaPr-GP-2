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
	EmptyDataFileNAme = errors.New("нет данных <fileName>")
)
