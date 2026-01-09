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

// Обеспечение единоразового выполняения для конфигурации сервиса.
var onceConf sync.Once

// Конфигурация сервиса.
type Configuration struct {
	PtrLoggerFile *logfile.LogFile  // Логгер файл.
	DataBase      domain.StorageI   // Интерфес БД.
	Container     container.Actions // Интерфейс контейнера.
	Flag          *flags.Config     // Флаги.
	ptrDB         *sql.DB           // Указатель на БД.
	Server        server.ServerI    // Интерфейс сервера.
}

// Указатель на конфигурацию сервиса.
var inst *Configuration

// Создание экземпляра конфигурации сервиса.
func New(l *logfile.LogFile, a domain.StorageI, f *flags.Config) *Configuration {
	onceConf.Do(func() {
		inst = &Configuration{
			PtrLoggerFile: l,
			DataBase:      a,
			Flag:          f,
		}
	})
	return inst
}

// Обновление подключения к серверу.
func NewServer(s server.ServerI) {
	inst.Server = s
}

// Проверка содержимого конфигурации.
func (c Configuration) CheckConf() error {

	if c.PtrLoggerFile == nil {
		return NilPtrLoggerFile
	}

	return nil
}
