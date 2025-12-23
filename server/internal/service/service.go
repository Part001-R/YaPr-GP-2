package service

import "fmt"

func Run() error {

	// Подготовка.
	conf, err := prepare()
	if err != nil {
		return fmt.Errorf("функция prepare, вернула ошибку: <%w>", err)
	}

	// Действия.
	if err := actions(conf); err != nil {
		return fmt.Errorf("функция actions, вернула ошибку: <%w>", err)
	}

	return nil
}
