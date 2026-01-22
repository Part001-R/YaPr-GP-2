// Основные функции пакета.
package domain

import (
	"context"
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/server/internal/adapters/sqlitestor"
)

type StorageI interface {
	// Закрытие подключения.
	Close() error
	// Регистрация пользователя.
	AddUserContext(ctx context.Context, userName, userPwd string) error
	// Аутентификация пользователя.
	AuthenticateUserContext(ctx context.Context, data DataUser) (bool, error)
	// Проверка, что в системе уже есть пользователь.
	UserExistContext(ctx context.Context) (bool, error)
	// Добавление данных - логин/пароль.
	AddDataLoginPasswordContext(ctx context.Context, data DataLoginPassword) error
	// Чтение данных - логин/пароль.
	ReadTableLoginPasswordContext(ctx context.Context) (list []LoginPassword, err error)
	// Удаление данных - логин/пароль.
	DelDataLoginPasswordContext(ctx context.Context, field1 string) error
	// Добавлеение данных - текст.
	AddDataTextContext(ctx context.Context, data DataText) error
	// Чтение данных - текст.
	ReadTableTextContext(ctx context.Context) (list []TextData, err error)
	// Удаление данных - текст.
	DelTextContext(ctx context.Context, field1 string) error
	// Добавление данных - банковские карты.
	AddDataBankCardContext(ctx context.Context, data DataBankCard) error
	// Чтение данных - банковские карты.
	ReadTableBankCardContext(ctx context.Context) (list []BankCard, err error)
	// Удаление данных - банковские карты.
	DelBankCardContext(ctx context.Context, field1 string) error
	// Полчение имён записей логин/пароль.
	GetNamesLoginPasswordContext(ctx context.Context) ([]string, error)
	// Получение строки по имени
	GetLoginPasswordByNameContext(ctx context.Context, name string) (data LoginPassword, err error)
	// Полчение имён записей текста.
	GetNamesTextContext(ctx context.Context) ([]string, error)
	// Получение строки по имени
	GetTextByNameContext(ctx context.Context, name string) (data TextData, err error)
	// Полчение имён записей банковских карт
	GetNamesBankCardContext(ctx context.Context) ([]string, error)
	// Получение банковской карты по имени
	GetBankCardByNameContext(ctx context.Context, name string) (data BankCard, err error)
}

// Интерфейс домена.
type DomainI interface {
	StorageI // Интерфейс БД
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
func NewStorage(dsn string) (DomainI, error) {

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
