package service

import (
	"fmt"

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

	// Создание конфигурации.
	// Указатель на БД, тут не передаётся. Указатель формируется после ввода данных аутентификации, т.к. при неактивности пользователя, подключение закрывается.
	conf := service.New(lgrFile, nil, flg)

	// Завершение.
	lgrFile.Write(fmt.Sprintf("Debug: Этап подготовки пройден. Режим работы клиента: <%s>", flg.Mode))
	return conf, nil
}
