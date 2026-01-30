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
type BankCardData struct {
	Name      string // наименование записи.
	Owner     string // владелец.
	Numb      string // номер.
	Valid     string // валидность.
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

// Формат записи банковской карты.
type BankCard struct {
	Field1    string // наименование записи.
	Field2    string // вдажелец.
	Field3    string // номер карты.
	Field4    string // дата валидности.
	Field5    string // код.
	CreatedAt string // время создания/обновления.
}

// Представление данных текста.
type DataUser struct {
	Field1    string // имя пользователь
	Field2    string // хэш пароля
	CreatedAt string // дата создания
}

// Представление данных текста.
type DataText struct {
	Field1    string // принадлежность
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
