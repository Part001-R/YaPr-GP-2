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
	Mode string // Режим работы клиента.
	DSN  string
}

// Обеспечение однократного выполнения.
var once sync.Once

// Данные флагов.
var flags = Config{}

// Реализация парсинга флагов. Возвращаются флаги.
func New() *Config {

	once.Do(func() {

		flag.StringVar(&flags.Mode, "m", ModeLocal, "режим работы клиента") // Определяется режим работы клиента.
		flag.StringVar(&flags.DSN, "d", DSN, "dsn БД")                      // Строка подключения к БД.

		flag.Parse()

		// Установка значений из флагов и переменных окружения.
		if err := setFlagsFromEnv(&flags); err != nil {
			log.Fatalf("ошибка подготовки флагов:<%v>", err)
		}
	})

	// Проверка флагов.
	if err := checkFlags(flags); err != nil {
		log.Fatalf("Ошибка проверки параметров звпуска сервиса: <%v>", err)
	}

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
	if envValue := os.Getenv("MODE_CLIENT"); envValue != "" {
		f.Mode = envValue
	}
	if envValue := os.Getenv("DSN_STORAGE"); envValue != "" {
		f.DSN = envValue
	}

	return nil
}
