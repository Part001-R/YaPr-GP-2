package service

import (
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/client/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/client/internal/logfile"
	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	service "github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
)

// Подготовительные действия, перед запуском сервиса.
func prepare() (*udt.Configuration, error) {

	// Логгер файла.
	ptrLgrFile, err := logfile.New("log.txt")
	if err != nil {
		return nil, fmt.Errorf("функция logfile.New, вернула ошибку: <%w>", err)
	}

	// БД.
	actionsDB, err := domain.NewStorage("file:manager.db?cache=shared&foreign_keys=on&mode=rwc")
	if err != nil {
		return nil, fmt.Errorf("функция domain.NewStorage, вернула ошибку: <%w>", err)
	}

	// Создание конфигурации.
	conf := service.New(ptrLgrFile, actionsDB)

	// Завершение.
	ptrLgrFile.Write("Debug: Этап подготовки пройден")
	return conf, nil
}
