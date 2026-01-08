package grpc

// Приянтые данные регистрации.
type registrationRX struct {
	userName      string // Имя пользователя.
	userPwd       string // Пароль пользователя.
	userPwdRepeat string // Подтверждение пароля.
}

// Предстваление токена.
type tokenData struct {
	name  string // имя
	token string // токен
}
