package udt

import (
	"database/sql"
	"sync"

	"github.com/Part001-R/YaPr-GP-2/client/internal/container"
	"github.com/Part001-R/YaPr-GP-2/client/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/client/internal/logfile"
)

// Обеспечение единоразового выполняения для конфигурации сервиса.
var onceConf sync.Once

// Конфигурация сервиса.
type Configuration struct {
	PtrLoggerFile *logfile.LogFile  // Логгер файл.
	DataBase      domain.StorageI   // Интерфес БД.
	Container     container.Actions // Интерфейс контейнера.
	ptrDB         *sql.DB           // Указатель на БД.
}

// Указатель на конфигурацию сервиса.
var confInst *Configuration

// Создание экземпляра конфигурации сервиса.
func New(f *logfile.LogFile, a domain.StorageI) *Configuration {
	onceConf.Do(func() {
		confInst = &Configuration{
			PtrLoggerFile: f,
			DataBase:      a,
		}
	})
	return confInst
}

// Проверка содержимого конфигурации.
func (c Configuration) CheckConf() error {

	if c.PtrLoggerFile == nil {
		return NilPtrLoggerFile
	}

	return nil
}
