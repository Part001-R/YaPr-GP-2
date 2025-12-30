package udt

import (
	"database/sql"
	"sync"

	"github.com/Part001-R/YaPr-GP-2/client/internal/container"
	"github.com/Part001-R/YaPr-GP-2/client/internal/logfile"
	"github.com/Part001-R/YaPr-GP-2/client/internal/storage/sqlitestor"
	"go.uber.org/zap"
)

// Обеспечение единоразового выполняения для конфигурации сервиса.
var onceConf sync.Once

// Конфигурация сервиса.
type Configuration struct {
	PtrLogger     *zap.Logger        // Логгер терминала.
	PtrLoggerFile *logfile.LogFile   // Логгер файл.
	DataBase      sqlitestor.Actions // Интерфес БД.
	Container     container.Actions  // Интерфейс контейнера.
	ptrDB         *sql.DB            // Указатель на БД.
}

// Указатель на конфигурацию сервиса.
var confInst *Configuration

// Создание экземпляра конфигурации сервиса.
func New(l *zap.Logger, f *logfile.LogFile, a sqlitestor.Actions) *Configuration {
	onceConf.Do(func() {
		confInst = &Configuration{
			PtrLogger:     l,
			PtrLoggerFile: f,
			DataBase:      a,
		}
	})
	return confInst
}

// Проверка содержимого конфигурации.
func (c Configuration) CheckConf() error {

	if c.PtrLogger == nil {
		return NilPtrLogger
	}
	if c.PtrLoggerFile == nil {
		return NilPtrLoggerFile
	}

	return nil
}
