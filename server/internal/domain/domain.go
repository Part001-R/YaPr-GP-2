// Основные функции пакета.
package domain

import (
	"context"
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/server/internal/adapters/sqlitestor"
)

// Интерфейс БД.
type StorageI interface {
	Close() error                                                                                               // Закрытие подключения.
	AddUserContext(ctx context.Context, userName, userPwd string) error                                         // Регистрация пользователя.
	AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error)                        // Аутентификация пользователя.
	UserExistContext(ctx context.Context) (bool, error)                                                         // Проверка, что в системе уже есть пользователь.
	AddDataLoginPasswordContext(ctx context.Context, field1, field2, field3, createdAt string) error            // Добавление данных - логин/пароль.
	ReadTableLoginPasswordContext(ctx context.Context) (list []LoginPassword, err error)                        // Чтение данных - логин/пароль.
	DelDataLoginPasswordContext(ctx context.Context, field1 string) error                                       // Удаление данных - логин/пароль.
	AddDataTextContext(ctx context.Context, field1, field2, createdAt string) error                             // Добавлеение данных - текст.
	ReadTableTextContext(ctx context.Context) (list []TextData, err error)                                      // Чтение данных - текст.
	DelTextContext(ctx context.Context, field1 string) error                                                    // Удаление данных - текст.
	AddDataBankCardContext(ctx context.Context, field1, field2, field3, field4, field5, createdAt string) error // Добавление данных - банковские карты.
	ReadTableBankCardContext(ctx context.Context) (list []BankCard, err error)                                  // Чтение данных - банковские карты.
	DelBankCardContext(ctx context.Context, field1 string) error                                                // Удаление данных - банковские карты.
}

// БД.
type storage struct {
	actions interface{} // Приём интерфейсов различных БД.
}

// Конструктор БД. Возвращается интерфейс и ошибка.
//
// Параметры:
//
//	dsn - строка подключения.
func NewStorage(dsn string) (StorageI, error) {

	// Выделение префикса из dsn.
	prefix := extractPrefixDSN(dsn)

	// Логика
	//
	switch prefix {
	// Если SQLite
	case "file":
		actions, err := sqlitestor.New(dsn)
		if err != nil {
			return nil, err
		}
		return &storage{actions: actions}, nil

	// ...
	default:
		return nil, fmt.Errorf("Тип БД неопределён, префикс: <%s>", prefix)
	}
}
