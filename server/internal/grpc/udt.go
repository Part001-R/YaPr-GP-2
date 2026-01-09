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
