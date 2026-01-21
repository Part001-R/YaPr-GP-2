// Основные функции пакета.
package domain

import (
	"context"
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/sqlitestor"
)

// Интерфейс БД.
type actionsDB interface {
	// Закрытие подключения.
	Close() error
	// Регистрация пользователя.
	AddUserContext(ctx context.Context, userName, userPwd string) error
	// Аутентификация пользователя.
	AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error)
	// Проверка, что в системе уже есть пользователь.
	UserExistContext(ctx context.Context) (bool, error)
	// Добавление данных - логин/пароль.
	AddDataLoginPasswordContext(ctx context.Context, data DataLoginPassword) error
	// Удаление данных - логин/пароль.
	DelDataLoginPasswordContext(ctx context.Context, field1 string) error
	// Добавлеение данных - текст.
	AddDataTextContext(ctx context.Context, data DataText) error
	// Удаление данных - текст.
	DelTextContext(ctx context.Context, field1 string) error
	// Добавление данных - банковские карты.
	AddDataBankCardContext(ctx context.Context, data DataBankCard) error
	// Удаление данных - банковские карты.
	DelBankCardContext(ctx context.Context, field1 string) error
	// Чтение имён записей логин/пароль.
	ReadNamesTableLoginPasswordContext(ctx context.Context) (names []string, err error)
	// Чтение записи логин/пароль по имени.
	ReadLoginPassworByNameContext(ctx context.Context, name string) (data LoginPassword, err error)
	// Чтение имён записей текст.
	ReadNamesTableTextContext(ctx context.Context) (names []string, err error)
	// Чтение записи текста по имени.
	ReadTextByNameContext(ctx context.Context, name string) (data TextData, err error)
	// Чтение имён записей банковских карт.
	ReadNamesTableBankCardContext(ctx context.Context) (names []string, err error)
	// Чтение записи банковской карты по имени.
	ReadBankCardByNameContext(ctx context.Context, name string) (data BankCard, err error)
}

// Интерфейс.
type Actions interface {
	actionsDB // Интерфейс БД
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
func NewStorage(dsn string) (Actions, error) {

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
