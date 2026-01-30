// Основной функционал пакета.
package service

import (
	"fmt"
	"log"
)

// Запуск сервиса. Возвращается ошибка.
func Run() error {

	// Подготовка.
	conf, err := prepare()
	if err != nil {
		log.Fatalf("функция prepare, вернула ошибку: <%v>", err)
		return fmt.Errorf("функция prepare, вернула ошибку: <%w>", err)
	}

	// Действия.
	if err := actions(conf); err != nil {
		conf.LgrFile.Write(fmt.Sprintf("функция actions, вернула ошибку: <%v>", err))
		return fmt.Errorf("функция actions, вернула ошибку: <%w>", err)
	}

	return nil
}
