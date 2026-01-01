package grpc

import "os"

// Добавление порции принятых данных к файлу. Возвращается ошибка.
func appendToFile(filename string, content []byte) error {

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content)
	return err
}
