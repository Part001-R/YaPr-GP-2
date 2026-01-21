// Обработчики БД.
package domain

import (
	"context"
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/sqlitestor"
)

// Закрытие подключения. Возвращается ошибка.
func (s *storage) Close() error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.Close()
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

//
// --- пользователь ---
//

// Регистрация нового пользователя. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	userName - имя пользователя.
//	userPwd - пароль пользователя.
func (s *storage) AddUserContext(ctx context.Context, userName, userPwd string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AddUserContext(ctx, userName, userPwd)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Аутентификация пользователя. Возвращается true - если пользователь аутентифицирован и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	userName - имя пользователя.
//	userPwd - пароль пользователя.
func (s *storage) AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AuthenticateUserContext(ctx, userName, userPwd)
	// ...
	default:
		return false, fmt.Errorf("Неизвестный тип БД")
	}
}

// Проверка, что уже есть зарегистрированный пользователь. Возвращается true - если уже есть запись пользователя и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) UserExistContext(ctx context.Context) (bool, error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.UserExistContext(ctx)
	// ...
	default:
		return false, fmt.Errorf("Неизвестный тип БД")
	}
}

//
// --- логин/пароль ---
//

// Добавление пары логин/пароль. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
//	field2 - поле имени пользователя.
//	field3 - поле пароля пользователя.
//	createdAt - поле даты создания записи.
func (s *storage) AddDataLoginPasswordContext(ctx context.Context, data DataLoginPassword) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		var txData sqlitestor.DataLoginPassword
		txData.Field1 = data.Field1
		txData.Field2 = data.Field2
		txData.Field3 = data.Field3
		txData.CreatedAt = data.CreatedAt
		return actions.AddDataLoginPasswordContext(ctx, txData)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Удаление пары логин/пароль. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
func (s *storage) DelDataLoginPasswordContext(ctx context.Context, field1 string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.DelDataLoginPasswordContext(ctx, field1)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

//
// --- текст ---
//

// Добавление текста. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
//	field2 - поле текста.
//	createdAt - поле даты создания записи.
func (s *storage) AddDataTextContext(ctx context.Context, data DataText) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		var txData sqlitestor.DataText
		txData.Field1 = data.Field1
		txData.Field2 = data.Field2
		txData.CreatedAt = data.CreatedAt
		return actions.AddDataTextContext(ctx, txData)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Удаление текста. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
func (s *storage) DelTextContext(ctx context.Context, field1 string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.DelTextContext(ctx, field1)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

//
// --- банковские карты ---
//

// Добавление банковской карты. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	data - данные.
func (s *storage) AddDataBankCardContext(ctx context.Context, data DataBankCard) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		var txData sqlitestor.DataBankCard
		txData.Field1 = data.Field1
		txData.Field2 = data.Field2
		txData.Field3 = data.Field3
		txData.Field4 = data.Field4
		txData.Field5 = data.Field5
		txData.CreatedAt = data.CreatedAt
		return actions.AddDataBankCardContext(ctx, txData)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Удаление банковской карты. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
func (s *storage) DelBankCardContext(ctx context.Context, field1 string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.DelBankCardContext(ctx, field1)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// получение имён записей логин/пароль. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) ReadNamesTableLoginPasswordContext(ctx context.Context) ([]string, error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.ReadNamesTableLoginPasswordContext(ctx)
	// ...
	default:
		return nil, fmt.Errorf("Неизвестный тип БД")
	}
}

// получение записи логин/пароль, по имени. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) ReadLoginPassworByNameContext(ctx context.Context, name string) (data LoginPassword, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		var rxData sqlitestor.LoginPassword
		rxData, err = actions.ReadLoginPassworByNameContext(ctx, name)
		if err != nil {
			return LoginPassword{}, err
		}
		data.Name = rxData.Name
		data.Login = rxData.Login
		data.Password = rxData.Password
		data.CreatedAt = rxData.CreatedAt
		return data, nil
	// ...
	default:
		return LoginPassword{}, fmt.Errorf("Неизвестный тип БД")
	}
}

// получение имён записей текста. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) ReadNamesTableTextContext(ctx context.Context) (names []string, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.ReadNamesTableTextContext(ctx)
	// ...
	default:
		return nil, fmt.Errorf("Неизвестный тип БД")
	}
}

// получение имени записи текста по имени. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) ReadTextByNameContext(ctx context.Context, name string) (data TextData, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		var rxData sqlitestor.TextData
		rxData, err = actions.ReadTextByNameContext(ctx, name)
		if err != nil {
			return TextData{}, err
		}
		data.Name = rxData.Name
		data.Text = rxData.Text
		data.CreatedAt = rxData.CreatedAt
		return data, nil
	// ...
	default:
		return TextData{}, fmt.Errorf("Неизвестный тип БД")
	}
}

// получение имён записей текста. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) ReadNamesTableBankCardContext(ctx context.Context) (names []string, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.ReadNamesTableBankCardContext(ctx)
	// ...
	default:
		return nil, fmt.Errorf("Неизвестный тип БД")
	}
}

// получение имени записи текста по имени. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) ReadBankCardByNameContext(ctx context.Context, name string) (data BankCard, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		var rxData sqlitestor.BankCard
		rxData, err = actions.ReadBankCardByNameContext(ctx, name)
		if err != nil {
			return BankCard{}, err
		}
		data.Name = rxData.Name
		data.Owner = rxData.Owner
		data.Numb = rxData.Numb
		data.Valid = rxData.Valid
		data.Code = rxData.Code
		data.CreatedAt = rxData.CreatedAt
		return data, nil
	// ...
	default:
		return BankCard{}, fmt.Errorf("Неизвестный тип БД")
	}
}
