package logfile

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

var once sync.Once

// Файл логов.
type LogFile struct {
	PtrLogFile *os.File
	mu         sync.Mutex
}

var inst *LogFile

// Создание файла логов.
func New(nameLogFile string) (*LogFile, error) {

	var err error

	once.Do(func() {

		if nameLogFile == "" {
			err = errors.New("нет содержимого в названии файла логов")
			return
		}

		var logFile *os.File

		logFile, err = os.OpenFile(nameLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			err = fmt.Errorf("ошибка подключения к файлу логов %s: <%w>", nameLogFile, err)
			return
		}
		inst = &LogFile{
			PtrLogFile: logFile,
		}
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания экземпляра файла логирования: <%w>", err)
	}

	return inst, nil
}

// Функция реализации записи логов в файл.
func (f *LogFile) Write(msg string) error {

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.PtrLogFile == nil {
		return errors.New("файл логов не инициализирован")
	}

	// Проверка.
	// Если размер файл больше 100МБ, создаётся новый файл.
	var err error
	f.PtrLogFile, err = checkFile(f.PtrLogFile)

	// Получение информации о месте вызова.
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown_file"
		line = 0
	}

	// Подготовка сообщения для записи в файл.
	logMessage := fmt.Sprintf("%s [%s:%d] %s\n", time.Now().Format("2006-01-02 15:04:05.000"), file, line, msg)

	// Запись.
	_, err = f.PtrLogFile.WriteString(logMessage)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл логов: <%w>", err)
	}

	return nil
}
