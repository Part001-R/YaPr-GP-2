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

	// Ошибка отправки заголовков
	ErrSendHeaders = errors.New("Ошибка отправки заголовков")
)
