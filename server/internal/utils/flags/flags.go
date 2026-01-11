// flags пакет для взаимодействия с флагами командной строки.
package flags

import (
	"flag"
	"log"
	"os"
	"sync"
)

// Флаги сервиса.
type Config struct {
	NameSubDirFiles string // Имя поддиректории принятых файлов.
	DSN             string // dsn БД
}

// Обеспечение однократного выполнения.
var once sync.Once

// Данные флагов.
var flags = Config{}

// Реализация парсинга флагов. Возвращаются флаги.
func New() *Config {

	once.Do(func() {

		flag.StringVar(&flags.NameSubDirFiles, "sd", NameSubDirFiles, "дочерняя директория для хранения файлов")
		flag.StringVar(&flags.DSN, "d", DSN, "dsn БД")

		flag.Parse()

		// Установка значений из флагов и переменных окружения.
		if err := setFlagsFromEnv(&flags); err != nil {
			log.Fatalf("ошибка подготовки флагов:<%v>", err)
		}
	})

	return &flags
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
	if envValue := os.Getenv("SUBDIR_FILES"); envValue != "" {
		f.NameSubDirFiles = envValue
	}
	if envValue := os.Getenv("DB_DSN"); envValue != "" {
		f.DSN = envValue
	}

	return nil
}
