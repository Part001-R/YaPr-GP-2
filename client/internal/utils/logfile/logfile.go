// Обработчики пакета.
package logfile

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// Файл логов.
type LogFile struct {
	PtrLogFile *os.File   // Указатель на файл.
	mu         sync.Mutex // Мьютекс доступа.
}

// Экземпляр.
var inst *LogFile

// Конструктор. Возвращается экземпляр и ошибка.
//
// Параметры:
//
//	nameLogFile  - имя файла.
func New(nameLogFile string) (*LogFile, error) {

	// Проверка.
	if nameLogFile == "" {
		return nil, EmptyDataArgumentNameLogFile

	}

	// Логика.
	var logFile *os.File

	logFile, err := os.OpenFile(nameLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("Функция OpenFile, вернула ошибку:<%w>", err)

	}
	inst = &LogFile{
		PtrLogFile: logFile,
	}

	return inst, nil
}

// Функция реализации записи логов в файл. Возвращается ошибка.
//
// Параметры:
//
//	msg - сообщение для записи.
func (f *LogFile) Write(msg string) error {

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.PtrLogFile == nil {
		return NilPtrLogger
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
