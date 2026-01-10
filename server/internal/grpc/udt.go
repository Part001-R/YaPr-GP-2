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
	RxID        string // id клиента
	RxFor       string // принадлежность логин/пароль
	RxLogin     string // логин
	RxPassword  string // пароль
	RxCreatedAt string // дата создания
}

// Принятые данные текста.
type RxText struct {
	RxID        string // id клиента
	RxFor       string // принадлежность текста
	RxText      string // текст
	RxCreatedAt string // дата создания
}

// Принятые данные банковской карты.
type RxBankCard struct {
	RxID        string // id клиента
	RxFor       string // принадлежность текста
	RxOwner     string // владелец
	RxNumb      string // номер
	RxValidData string // валидность
	RxCode      string // код
	RxCreatedAt string // дата создания
}
