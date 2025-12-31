package service

import (
	"fmt"
)

func Run() error {

	// Подготовка.
	conf, err := prepare()
	if err != nil {
		conf.PtrLoggerFile.Write(fmt.Sprintf("функция prepare, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция prepare, вернула ошибку: <%w>", err)
	}

	// Действия.
	if err := actions(conf); err != nil {
		conf.PtrLoggerFile.Write(fmt.Sprintf("функция actions, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция actions, вернула ошибку: <%w>", err)
	}

	return nil
}
