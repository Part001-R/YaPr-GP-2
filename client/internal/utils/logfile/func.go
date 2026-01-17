// Вспомогательные функции пакета.
package logfile

import (
	"fmt"
	"os"
	"time"
)

// Проверка файла логов. Если размер больше 100МБ, создаётся новый файл. Возвращается указатель на файл и ошибка.
//
//	Параметры:
//
//	file - указатель на файл.
func checkFile(file *os.File) (*os.File, error) {

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о файле: <%w>", err)
	}

	if fileInfo.Size() > 100*1024*1024 { // 100 МБ в байтах
		// Закрываем старый файл
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("ошибка закрытия файла логов: <%w>", err)
		}

		// Создаем новый файл с тем же именем
		fileName := fmt.Sprintf("%s %s", time.Now().Format("2006-01-02 15:04:05"), "log")

		newLogFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, fmt.Errorf("ошибка создания нового файла логов: <%w>", err)
		}

		// Обновляем указатель на файл логов.
		return newLogFile, nil
	}

	// Размер в норме. Возврат исходного указателя.
	return file, nil
}
