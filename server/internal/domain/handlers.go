// Вызовы обработчиков БД, в зависимости от типа БД.
package domain

import (
	"context"
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/server/internal/adapters/sqlitestor"
)

// Закрытие подключения к БД. Возвращается ошибка.
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
func (s *storage) AuthenticateUserContext(ctx context.Context, data DataUser) (bool, error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		var txData sqlitestor.DataUser
		txData.Field1 = data.Field1
		txData.Field2 = data.Field2
		return actions.AuthenticateUserContext(ctx, txData)
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
//	data - данные.
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

// Получение всех записей логин/пароль из БД. Возвращается массив записей и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) ReadTableLoginPasswordContext(ctx context.Context) (list []LoginPassword, err error) {

	switch actions := s.actions.(type) {

	case sqlitestor.Actions:
		rxArr, err := actions.ReadTableLoginPasswordContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("SQlite. Функция ReadTableLoginPasswordContext, вернула ошибку: <%w>", err)
		}
		for _, v := range rxArr {
			var el LoginPassword

			el.Name = v.Name
			el.Login = v.Login
			el.Password = v.Password
			el.CreatedAt = v.CreatedAt

			list = append(list, el)
		}
		return list, nil
	// ...
	default:
		return nil, fmt.Errorf("Неизвестный тип БД")
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

// Чтение имён записей - логин/пароль. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) GetNamesLoginPasswordContext(ctx context.Context) ([]string, error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		rxData, err := actions.GetNamesLoginPasswordContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("SQlite. Функция GetNamesLoginPasswordContext, вернула ошибку: <%w>", err)
		}
		return rxData, nil
	// ...
	default:
		return []string{}, fmt.Errorf("Неизвестный тип БД")
	}
}

// Чтение данных логин/пароль по имени. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	name - имя записи.
func (s *storage) GetLoginPasswordByNameContext(ctx context.Context, name string) (data LoginPassword, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		rxData, err := actions.GetLoginPasswordByNameContext(ctx, name)
		if err != nil {
			return LoginPassword{}, fmt.Errorf("SQlite. Функция GetLoginPasswordByNameContext, вернула ошибку: <%w>", err)
		}
		data.Name = rxData.Field1
		data.Login = rxData.Field2
		data.Password = rxData.Field3
		data.CreatedAt = rxData.CreatedAt
		return data, nil
	// ...
	default:
		return LoginPassword{}, fmt.Errorf("Неизвестный тип БД")
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
//	data - данные.
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

// Получение всех записей текста из БД. Возвращается массив записей и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) ReadTableTextContext(ctx context.Context) (list []TextData, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		rxData, err := actions.ReadTableTextContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("SQlite. Функция ReadTableTextContext, вернула ошибку: <%w>", err)
		}
		for _, v := range rxData {
			var el TextData

			el.Name = v.Name
			el.Text = v.Text
			el.CreatedAt = v.CreatedAt

			list = append(list, el)
		}
		return list, nil
	// ...
	default:
		return nil, fmt.Errorf("Неизвестный тип БД")
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

// Получение имен записей текста. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) GetNamesTextContext(ctx context.Context) ([]string, error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		rxData, err := actions.GetNamesTextContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("SQlite. Функция GetNamesTextContext, вернула ошибку: <%w>", err)
		}
		return rxData, nil
	// ...
	default:
		return []string{}, fmt.Errorf("Неизвестный тип БД")
	}
}

// Получение записb текста, по имени. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	name - имя записи.
func (s *storage) GetTextByNameContext(ctx context.Context, name string) (data TextData, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		rxData, err := actions.GetTextByNameContext(ctx, name)
		if err != nil {
			return TextData{}, fmt.Errorf("SQlite. Функция GetTextByNameContext, вернула ошибку: <%w>", err)
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

// Получение всех записей банковских карт из БД. Возвращается массив записей и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) ReadTableBankCardContext(ctx context.Context) (list []BankCard, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		rxData, err := actions.ReadTableBankCardContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("SQlite. Функция ReadTableBankCardContext, вернула ошибку: <%w>", err)
		}
		for _, v := range rxData {
			var el BankCard

			el.Name = v.Field1
			el.Owner = v.Field2
			el.Numb = v.Field3
			el.Valid = v.Field4
			el.Code = v.Field5
			el.CreatedAt = v.CreatedAt

			list = append(list, el)
		}
		return list, nil
	// ...
	default:
		return nil, fmt.Errorf("Неизвестный тип БД")
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

// Получение имен записей банковских карт. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (s *storage) GetNamesBankCardContext(ctx context.Context) ([]string, error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		rxData, err := actions.GetNamesBankCardContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("SQlite. Функция GetNamesBankCardContext, вернула ошибку: <%w>", err)
		}
		return rxData, nil
	// ...
	default:
		return []string{}, fmt.Errorf("Неизвестный тип БД")
	}
}

// Получение записи банковской карты, по имени. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	name - имя записи.
func (s *storage) GetBankCardByNameContext(ctx context.Context, name string) (data BankCard, err error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		rxData, err := actions.GetBankCardByNameContext(ctx, name)
		if err != nil {
			return BankCard{}, fmt.Errorf("SQlite. Функция GetBankCardByNameContext, вернула ошибку: <%w>", err)
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
