package ui

// Информация по файлу.
type infoByFiles struct {
	name   string // имя файла.
	volume int64  // размер файла.
}

// Представление приянтых данных логин/пароль
type rxLoginPassword struct {
	name      string // имя записи
	login     string // логин
	password  string // пароль
	createdAt string // дата создания
}
