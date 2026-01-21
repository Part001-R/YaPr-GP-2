// Типы данных пакета.
package sqlitestor

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

// Данные для записи логин/пароль.
type DataLoginPassword struct {
	Field1    string // имя записи
	Field2    string // логин
	Field3    string // пароль
	CreatedAt string // дата создания
}

// Данные для записи текста.
type DataText struct {
	Field1    string // имя записи
	Field2    string // текст
	CreatedAt string // дата создания
}

// Данные для записи банковской карты.
type DataBankCard struct {
	Field1    string // имя записи
	Field2    string // владелец
	Field3    string // номер
	Field4    string // дата валидности
	Field5    string // код
	CreatedAt string // дата создания
}
