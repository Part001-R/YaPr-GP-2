// Вспомогательные функции пакета.
package flags

import (
	"fmt"
	"os"
	"strings"
)

// Проверка значений влагов. Возвращается ошибка.
func checkFlags(f Config) error {

	// Проверка режима работы.
	if f.Mode != ModeLocal && f.Mode != ModeRemote {
		return fmt.Errorf("Проверьте режим запуска. Указан: <%s>", f.Mode)
	}

	// Проверка DSN
	if f.Mode == ModeLocal {
		if f.DSN == "" {
			return ErrEmptyDSN
		}
	}

	return nil
}

// setFlagsFromEnv, функция выполняет установку значений исходя из содержимого флагов и env. Возвращает ошибку.
//
// Парамметры:
//
//	f - указатель на структуру.
func setFlagsFromEnv(f *Config) error {

	// Проверка
	if f == nil {
		return ErrNilPtrArgumentF
	}

	// Логика
	if envValue := os.Getenv("MODE_CLIENT"); envValue != "" {
		f.Mode = envValue
	}
	if envValue := os.Getenv("DSN_STORAGE"); envValue != "" {
		f.DSN = envValue
	}
	if envValue := os.Getenv("LOCAL_NAME_CONTAINER"); envValue != "" {
		f.LocalNameContainer = envValue
	}

	return nil
}

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

	// MySQL
	if strings.HasPrefix(dbPart, "mysql://") {
		nameParts := strings.Split(dbPart, "/")
		return nameParts[len(nameParts)-1], nil
	}

	// PostgreSQL
	if strings.HasPrefix(dbPart, "postgres://") {
		nameParts := strings.Split(dbPart, "/")
		return nameParts[len(nameParts)-1], nil
	}

	// SQL Server
	if strings.HasPrefix(dbPart, "sqlserver://") {
		subParts := strings.Split(dbPart, ";")
		for _, part := range subParts {
			if strings.HasPrefix(part, "database=") {
				return strings.TrimPrefix(part, "database="), nil
			}
		}
	}

	// ...

	return "", fmt.Errorf("недопустимый формат DSN: %s", dsn)
}
