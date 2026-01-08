// Вызовы обработчиков БД, в зависимости от типа БД.
package domain

import (
	"context"
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/server/internal/adapters/sqlitestor"
)

// Закрытие подключения.
func (s *storage) Close() error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.Close()
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Регистрация пользователя.
func (s *storage) AddUserContext(ctx context.Context, userName, userPwd string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AddUserContext(ctx, userName, userPwd)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Аутентификация пользователя.
func (s *storage) AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AuthenticateUserContext(ctx, userName, userPwd)
	// ...
	default:
		return false, fmt.Errorf("Неизвестный тип БД")
	}
}

// Проверка, что в системе уже есть зарегистрированный пользователь.
func (s *storage) UserExistContext(ctx context.Context) (bool, error) {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.UserExistContext(ctx)
	// ...
	default:
		return false, fmt.Errorf("Неизвестный тип БД")
	}
}

// Добавление данных - логин/пароль.
func (s *storage) AddDataLoginPasswordContext(ctx context.Context, field1, field2, field3, createdAt string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AddDataLoginPasswordContext(ctx, field1, field2, field3, createdAt)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Чтение данных - логин/пароль.
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

// Удаление данных - логин/пароль.
func (s *storage) DelDataLoginPasswordContext(ctx context.Context, field1 string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.DelDataLoginPasswordContext(ctx, field1)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Добавление данных - текст.
func (s *storage) AddDataTextContext(ctx context.Context, field1, field2, createdAt string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AddDataTextContext(ctx, field1, field2, createdAt)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Чтение данных - текст.
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

// Удаление данных - текст.
func (s *storage) DelTextContext(ctx context.Context, field1 string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.DelTextContext(ctx, field1)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Добавление данных - бакновские карты.
func (s *storage) AddDataBankCardContext(ctx context.Context, field1, field2, field3, field4, field5, createdAt string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.AddDataBankCardContext(ctx, field1, field2, field3, field4, field5, createdAt)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}

// Чтение данных - банковские карты.
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

// Удаление данных - банковские карты.
func (s *storage) DelBankCardContext(ctx context.Context, field1 string) error {

	switch actions := s.actions.(type) {
	case sqlitestor.Actions:
		return actions.DelBankCardContext(ctx, field1)
	// ...
	default:
		return fmt.Errorf("Неизвестный тип БД")
	}
}
