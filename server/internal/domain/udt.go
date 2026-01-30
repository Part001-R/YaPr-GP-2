// Типы данных пакета.
package domain

// Формат записи логин/пароль
type LoginPassword struct {
	Name      string // наименование записи.
	Login     string // логин.
	Password  string // пароль.
	CreatedAt string // время создания/обновления.
}

// Формат записи текстовых данных.
type TextData struct {
	Name      string // наименование записи.
	Text      string // текст.
	CreatedAt string // время создания/обновления.
}

// Формат записи банковской карты.
type BankCard struct {
	Name      string // наименование записи.
	Owner     string // вдажелец.
	Numb      string // номер карты.
	Valid     string // дата валидности.
	Code      string // код.
	CreatedAt string // время создания/обновления.
}

// Представление данных текста.
type DataLoginPassword struct {
	Field1    string // принадлежность
	Field2    string // логин
	Field3    string // пароль
	CreatedAt string // дата создания
}

// Представление данных текста.
type DataText struct {
	IDClient  string // id клиента
	Field1    string // принадлежность
	Field2    string // текст
	CreatedAt string // дата создания
}

// Данные для записи банковской карты.
type DataBankCard struct {
	IDClient  string // id клиента
	Field1    string // имя записи
	Field2    string // владелец
	Field3    string // номер
	Field4    string // дата валидности
	Field5    string // код
	CreatedAt string // дата создания
}

// Представление данных текста.
type DataUser struct {
	Field1    string // имя пользователь
	Field2    string // хэш пароля
	CreatedAt string // дата создания
}
