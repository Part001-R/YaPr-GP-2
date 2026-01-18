// flags пакет для взаимодействия с флагами командной строки.
package flags

import (
	"flag"
	"log"
	"sync"
)

// Флаги сервиса.
type Config struct {
	Mode               string // Режим работы клиента.
	DSN                string // Строка подключения к БД.
	LocalNameDB        string // Имя локальной БД.
	LocalNameContainer string // Имя локального контейнера.
}

// Обеспечение однократного выполнения.
var once sync.Once

// Данные флагов.
var flags = Config{}

// Реализация парсинга флагов. Возвращаются флаги.
func New() *Config {

	once.Do(func() {

		flag.StringVar(&flags.Mode, "m", ModeLocal, "режим работы клиента")
		flag.StringVar(&flags.DSN, "d", DSN, "dsn БД")
		flag.StringVar(&flags.LocalNameContainer, "container", LocalNameContainer, "имя локального контейнера")

		flag.Parse()

		// Установка значений из флагов и переменных окружения.
		if err := setFlagsFromEnv(&flags); err != nil {
			log.Fatalf("ошибка подготовки флагов:<%v>", err)
		}

		// Получение имени БД из DSN.
		nameDB, err := GetNameDBFromDSN(flags.DSN)
		if err != nil {
			log.Fatalf("ошибка получения имени БД из dsn:<%v>", err)
		}
		flags.LocalNameDB = nameDB
	})

	// Проверка флагов.
	if err := checkFlags(flags); err != nil {
		log.Fatalf("Ошибка проверки параметров звпуска сервиса: <%v>", err)
	}

	return &flags
}
