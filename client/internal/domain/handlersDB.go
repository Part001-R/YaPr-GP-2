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
func (s *storage) AddDataLoginPasswordContext(ctx context.Context, field1, field2, field3, createdAt string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AddDataLoginPasswordContext(ctx, field1, field2, field3, createdAt)
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
			return nil, err
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
func (s *storage) AddDataTextContext(ctx context.Context, field1, field2, createdAt string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AddDataTextContext(ctx, field1, field2, createdAt)
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
		rxArr, err := actions.ReadTableTextContext(ctx)
		if err != nil {
			return nil, err
		}
		for _, v := range rxArr {
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

//
// --- банковские карты ---
//

// Добавление банковской карты. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
//	field2 - поле имени владельца.
//	field3 - поле номера.
//	field4 - поле даты вылидности.
//	field5 - поле кода.
//	createdAt - поле даты создания записи.
func (s *storage) AddDataBankCardContext(ctx context.Context, field1, field2, field3, field4, field5, createdAt string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AddDataBankCardContext(ctx, field1, field2, field3, field4, field5, createdAt)
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
		rxArr, err := actions.ReadTableBankCardContext(ctx)
		if err != nil {
			return nil, err
		}
		for _, v := range rxArr {
			var el BankCard

			el.Name = v.Name
			el.Owner = v.Owner
			el.Numb = v.Numb
			el.Valid = v.Valid
			el.Code = v.Code
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

// ReadLoginPassworByNameContext(ctx context.Context, name string) (data LoginPassword, err error)

// получение имён записей логин/пароль. Возвращается ошибка.
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
