package grpc

// Принятые данные регистрации.
type registrationRX struct {
	userName      string // Имя пользователя.
	userPwd       string // Пароль пользователя.
	userPwdRepeat string // Подтверждение пароля.
}

// Принятые данные аутентификации.
type authenticationRX struct {
	userName string // Имя пользователя.
	userPwd  string // Пароль пользователя.
}

// Представление токена.
type tokenData struct {
	name  string // имя
	token string // токен
}

// Принятые данные логин/пароль.
type RxLoginPassword struct {
	ID        string // id клиента
	For       string // принадлежность логин/пароль
	Login     string // логин
	Password  string // пароль
	CreatedAt string // дата создания
}

// Передаваемые данные логин/пароль.
type TxLoginPassword struct {
	For       string // принадлежность логин/пароль
	Login     string // логин
	Password  string // пароль
	CreatedAt string // дата создания
}

// Принятые данные текста.
type RxText struct {
	ID        string // id клиента
	For       string // принадлежность текста
	Text      string // текст
	CreatedAt string // дата создания
}

// Принятые данные банковской карты.
type RxBankCard struct {
	ID        string // id клиента
	For       string // принадлежность текста
	Owner     string // владелец
	Numb      string // номер
	ValidData string // валидность
	Code      string // код
	CreatedAt string // дата создания
}

// Приянтые данные для запроса логин/пароль по имени записи.
type RxReqLoginPasswordByName struct {
	ClientID string // id клиента
	Name     string // имя записи логин/пароль
}
