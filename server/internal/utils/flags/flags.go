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
	NameSubDirFiles  string // Имя поддиректории принятых файлов.
	NameSubDirBackUp string // Имя поддиректории backUp.
	DSN              string // dsn БД
	Port             string // Порт прослушивания.
}

// Обеспечение однократного выполнения.
var once sync.Once

// Данные флагов.
var flags = Config{}

// Реализация парсинга флагов. Возвращается указатель на экземпляр.
func New() *Config {

	once.Do(func() {

		flag.StringVar(&flags.NameSubDirFiles, "sdf", NameSubDirFiles, "дочерняя директория для хранения файлов")
		flag.StringVar(&flags.NameSubDirBackUp, "sdb", NameSubDirBackUp, "дочерняя директория для хранения backUp")
		flag.StringVar(&flags.DSN, "d", DSN, "dsn БД")
		flag.StringVar(&flags.Port, "p", Port, "порт прослушивания")

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
	if envValue := os.Getenv("SERVER_SUBDIR_FILES"); envValue != "" {
		f.NameSubDirFiles = envValue
	}
	if envValue := os.Getenv("SERVER_SUBDIR_BACKUP"); envValue != "" {
		f.NameSubDirBackUp = envValue
	}
	if envValue := os.Getenv("SERVER_DB_DSN"); envValue != "" {
		f.DSN = envValue
	}
	if envValue := os.Getenv("SERVER_PORT"); envValue != "" {
		f.Port = envValue
	}

	return nil
}
