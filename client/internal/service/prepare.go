package service

import (
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/client/internal/logfile"
	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	service "github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/client/internal/storage/sqlitestor"
	"github.com/Part001-R/YaPr-GP-2/internal/utils/logger"
)

// Подготовительные действия, перед запуском сервиса.
func prepare() (*udt.Configuration, error) {

	// Логгер файла.
	ptrLgrFile, err := logfile.New("log.txt")
	if err != nil {
		return nil, fmt.Errorf("функция logfile.New, вернула ошибку: <%w>", err)
	}

	// Логгер терминала.
	ptrLgr, err := logger.New("debug")
	if err != nil {
		return nil, fmt.Errorf("функция logger.New, вернула ошибку: <%w>", err)
	}

	// БД.
	actionsDB, err := sqlitestor.New("file:manager.db?cache=shared&foreign_keys=on&mode=rwc")
	if err != nil {
		return nil, fmt.Errorf("функция sqlitestor.New, вернула ошибку: <%w>", err)
	}

	// Создание конфигурации.
	conf := service.New(ptrLgr, ptrLgrFile, actionsDB)

	// Завершение.
	ptrLgrFile.Write("Debug: Этап подготовки пройден")
	return conf, nil
}
