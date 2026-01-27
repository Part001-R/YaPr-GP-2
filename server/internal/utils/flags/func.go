package flags

import (
	"fmt"
	"strings"
)

// Получение имени БД из строки dsn. Возвращается имя БД и ошибка.
//
// Параметры:
//
//	dsn - строка подключения.
func GetNameDBFromDSN(dsn string) (string, error) {

	parts := strings.Split(dsn, "?")
	dbPart := parts[0]

	// SQLite
	if strings.HasPrefix(dbPart, "file:") {
		pathParts := strings.Split(dbPart, ":")
		if len(pathParts) != 2 {
			return "", fmt.Errorf("недопустимый формат DSN: %s", dsn)
		}
		return pathParts[1], nil
	}

	// ...

	return "", fmt.Errorf("недопустимый формат DSN: %s", dsn)
}
