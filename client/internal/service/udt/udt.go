// Взаимодействие с экземпляром сервиса.
package udt

import (
	"database/sql"
	"sync"

	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/container"
	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/server"
	"github.com/Part001-R/YaPr-GP-2/client/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/logfile"
)

// Обеспечение единоразового выполняения инициализации конструктора.
var onceConf sync.Once

// Конфигурация сервиса.
type Configuration struct {
	LgrFile   *logfile.LogFile  // Логгер файл.
	ActionsDB domain.Actions    // Интерфес БД.
	Container container.Actions // Интерфейс контейнера.
	Flag      *flags.Config     // Флаги.
	dataBase  *sql.DB           // Указатель на БД.
	Server    server.ServerI    // Интерфейс сервера.
}

// Указатель на конфигурацию сервиса.
var inst *Configuration

// Создание экземпляра конфигурации сервиса. Возвращается конфигурация.
//
// Параметры:
//
//	l - указатель логгера.
//	a - интерфейс домена.
//	f - указатель на флаги.
func New(l *logfile.LogFile, a domain.Actions, f *flags.Config) (*Configuration, error) {

	// Проверка.
	if l == nil {
		return nil, NilPtrArgumentL
	}
	if f == nil {
		return nil, NilPtrArgumentF
	}

	// Логика.
	onceConf.Do(func() {
		inst = &Configuration{
			LgrFile:   l,
			ActionsDB: a,
			Flag:      f,
		}
	})
	return inst, nil
}

// Обновление подключения к серверу.
//
// Параметры:
//
//	s - интерфейс сервиса.
func NewServer(s server.ServerI) {
	inst.Server = s
}

// Проверка содержимого конфигурации. Возвращается ошибка.
func (c Configuration) CheckConf() error {

	if c.LgrFile == nil {
		return NilPtrLoggerFile
	}

	return nil
}
