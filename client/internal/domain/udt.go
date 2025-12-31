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
