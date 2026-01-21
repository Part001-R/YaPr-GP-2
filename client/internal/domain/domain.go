// Основные функции пакета.
package domain

import (
	"context"
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/sqlitestor"
)

// Интерфейс БД.
type actionsDB interface {
	Close() error                                                                                   // Закрытие подключения.
	AddUserContext(ctx context.Context, userName, userPwd string) error                             // Регистрация пользователя.
	AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error)            // Аутентификация пользователя.
	UserExistContext(ctx context.Context) (bool, error)                                             // Проверка, что в системе уже есть пользователь.
	AddDataLoginPasswordContext(ctx context.Context, data DataLoginPassword) error                  // Добавление данных - логин/пароль.
	DelDataLoginPasswordContext(ctx context.Context, field1 string) error                           // Удаление данных - логин/пароль.
	AddDataTextContext(ctx context.Context, data DataText) error                                    // Добавлеение данных - текст.
	DelTextContext(ctx context.Context, field1 string) error                                        // Удаление данных - текст.
	AddDataBankCardContext(ctx context.Context, data DataBankCard) error                            // Добавление данных - банковские карты.
	DelBankCardContext(ctx context.Context, field1 string) error                                    // Удаление данных - банковские карты.
	ReadNamesTableLoginPasswordContext(ctx context.Context) (names []string, err error)             // Чтение имён записей логин/пароль.
	ReadLoginPassworByNameContext(ctx context.Context, name string) (data LoginPassword, err error) // Чтение записи логин/пароль по имени.
	ReadNamesTableTextContext(ctx context.Context) (names []string, err error)                      // Чтение имён записей текст.
	ReadTextByNameContext(ctx context.Context, name string) (data TextData, err error)              // Чтение записи текста по имени.
	ReadNamesTableBankCardContext(ctx context.Context) (names []string, err error)                  // Чтение имён записей банковских карт.
	ReadBankCardByNameContext(ctx context.Context, name string) (data BankCard, err error)          // Чтение записи банковской карты по имени.
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
