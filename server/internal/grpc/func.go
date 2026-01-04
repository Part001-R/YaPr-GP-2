package grpc

import (
	"fmt"
	"os"
)

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

// Определение размера файла в байтах. Возвращается размер в байтах и ошибка.
func sizeFile(fileName string) (int64, error) {

	// Проверка существования файла.
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return 0, NotFound
	}

	// Информация по файлу.
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return 0, fmt.Errorf("Ошибка получения данных по файлу: <%s>: %v", fileName, err)
	}

	return fileInfo.Size(), nil
}
