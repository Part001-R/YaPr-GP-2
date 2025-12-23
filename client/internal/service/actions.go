package service

import (
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/client/internal/ui"
)

// Действия сервиса.
func actions(conf *udt.Configuration) error {

	// Проверка аргументов.
	if conf == nil {
		return NilPtrArgumentConf
	}
	if err := conf.CheckConf(); err != nil {
		return err
	}

	conf.PtrLoggerFile.Write("Info: Запуск CLI UI")

	// Запуск CLI UI
	if err := ui.Run(conf); err != nil {
		return fmt.Errorf("ошибка в ui.Run: <%w>", err)
	}

	return nil
}
