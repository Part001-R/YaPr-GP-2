package service

import (
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/client/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	service "github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/logfile"
)

// Подготовительные действия, перед запуском сервиса.
func prepare() (*udt.Configuration, error) {

	// Логгер файла.
	lgrFile, err := logfile.New("log.txt")
	if err != nil {
		return nil, fmt.Errorf("функция logfile.New, вернула ошибку: <%w>", err)
	}

	// Флаги.
	flg := flags.New()

	// БД.
	var storage domain.StorageI

	if flg.Mode == flags.ModeLocal { // Если клиент в режиме Local
		storage, err = domain.NewStorage(flg.DSN)
		if err != nil {
			return nil, fmt.Errorf("функция domain.NewStorage, вернула ошибку: <%w>", err)
		}
	}

	// Создание конфигурации.
	conf := service.New(lgrFile, storage, flg)

	// Завершение.
	lgrFile.Write(fmt.Sprintf("Debug: Этап подготовки пройден. Режим работы клиента: <%s>", flg.Mode))
	return conf, nil
}
