// Подготовительные действия покета.
package service

import (
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/client/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	service "github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/logfile"
)

var (
	// Имя файла логгера
	nameLogFile = "log.txt"
)

// Подготовительные действия, перед запуском сервиса. Возвращается конфигурация пакета и ошибка.
func prepare() (*udt.Configuration, error) {

	// Логгер файла.
	lgrFile, err := logfile.New(nameLogFile)
	if err != nil {
		return nil, fmt.Errorf("функция logfile.New, вернула ошибку: <%w>", err)
	}

	// Флаги.
	flg := flags.New()

	// БД.
	var storage domain.Actions

	if flg.Mode == flags.ModeLocal {
		storage, err = domain.NewStorage(flg.DSN)
		if err != nil {
			return nil, fmt.Errorf("функция NewStorage, вернула ошибку: <%w>", err)
		}
	}

	// Создание конфигурации.
	conf, err := service.New(lgrFile, storage, flg)
	if err != nil {
		return nil, fmt.Errorf("функция service.New, вернула ошибку: <%w>", err)
	}

	// Завершение.
	lgrFile.Write(fmt.Sprintf("Debug: Этап подготовки пройден. Режим работы клиента: <%s>", flg.Mode))
	return conf, nil
}
