// Представление функционала контейнера.
package container

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	nonceSize = 24
)

// Структура записи в контейнере
type FileEntry struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

type Container struct {
	Name  string      `json:"name"`
	Key   [32]byte    `json:"key"`
	Files []FileEntry `json:"files"` // файлы в контейнере.
}

// Интерфейс.
type Actions interface {
	AddFileToContainer(fileName string, key [32]byte, chProcess chan<- float64, chError chan<- error, chDone chan<- struct{})
	GetFileFromContainer(fileName string, key [32]byte, txChPercent chan<- float64, txChErr chan<- error, txChDone chan<- struct{}, txChData chan<- []byte, rxChBreak <-chan struct{})
	ListFilesInContainer(key [32]byte) (files []string, err error)
	RemoveFileFromContainer(fileName string, key [32]byte) error
}

// Экземпляр.
var inst *Container

// Конструктор. Возвращается интерфейс.
//
// Параметры:
//
//	name - имя контейнера.
//	key - ключ шифрования.
func New(name string, key [32]byte) (Actions, error) {

	// Проверка.
	if name == "" {
		return nil, EmptyDataArgumentNAme
	}

	// Логика.
	if _, err := os.Stat(name); os.IsNotExist(err) {
		container := Container{
			Name: name,
			Key:  key,
		}
		data, err := json.Marshal(container)
		if err != nil {
			return nil, fmt.Errorf("ошибка сериализации контейнера: %w", err)

		}

		// Проверка записью.
		if err := os.WriteFile(name, data, 0644); err != nil {
			return nil, fmt.Errorf("ошибка создания файла контейнера: %w", err)

		}
	}
	inst = &Container{
		Name: name,
		Key:  key,
	}

	return inst, nil
}

// Добавление файла в контейнер. Для запуска в горутине.
//
// Параметры:
//
//	fileName - имя файла.
//	key - ключ шифрования.
//	chProcess - канал передачи процентов процесса.
//	chError - канал передачи ошибки.
//	chDone - канал передачи признака завершения процесса.
func (c *Container) AddFileToContainer(fileName string, key [32]byte, chProcess chan<- float64, chError chan<- error, chDone chan<- struct{}) {

	defer func() {
		close(chProcess)
		close(chError)
		close(chDone)
	}()

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		chError <- fmt.Errorf("ошибка: <%v>, при чтении файла: <%s>", err, fileName)
		return
	}
	fileSize := fileInfo.Size()

	file, err := os.Open(fileName)
	if err != nil {
		chError <- fmt.Errorf("ошибка: <%v>, при открытии файла: <%s>", err, fileName)
		return
	}
	defer file.Close()

	var totalRead int64
	buffer := make([]byte, 1024) // Буфер 1КБ

	// Чтение контейнера.
	container, err := readContainer(key, c)
	if err != nil {
		chError <- fmt.Errorf("функция c.readContainer, вернула ошибку: <%v>", err)
		return
	}

	// Выделение имени файла из полного пути.
	fileN := filepath.Base(fileName)

	// Проверка наличия такого имени файла в контейнере.
	for _, file := range container.Files {
		if file.Name == fileN {
			chError <- fmt.Errorf("файл <%s> уже существует в контейнере", fileN)
			return
		}
	}

	// Создание нового файла для контейнера
	newFileEntry := FileEntry{Name: fileN, Content: make([]byte, 0, fileSize)}

	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			chError <- fmt.Errorf("ошибка при чтении файла: <%v>", err)
			return
		}
		totalRead += int64(n)

		// Добавление данных
		newFileEntry.Content = append(newFileEntry.Content, buffer[:n]...)

		// Процент процесса
		progress := (float64(totalRead) / float64(fileSize)) * 100.0
		chProcess <- progress
	}

	// Добавление файла в контейнер
	container.Files = append(container.Files, newFileEntry)

	if err := writeContainer(container, key, c); err != nil {
		chError <- fmt.Errorf("функция writeContainer, вернула ошибку: <%w>", err)
		return
	}

	// Установка признака успешного завершения процесса
	chDone <- struct{}{}
}

// Получение файла из контейнера, по имени. Для запуска в горутине.
//
// Параметры:
//
//	fileName - имя файла.
//	key - ключ шифрования.
//	txChPercent - канал передачи процентов процесса.
//	txChErr - канал передачи ошибки.
//	txChDone - канал передачи признака успешного завершения процесса.
//	txChData - канал передачи данных.
//	rxChBreak - канал приёма сигнала остановки процесса.
func (c *Container) GetFileFromContainer(fileName string, key [32]byte, txChPercent chan<- float64, txChErr chan<- error, txChDone chan<- struct{}, txChData chan<- []byte, rxChBreak <-chan struct{}) {
	defer func() {
		close(txChPercent)
		close(txChErr)
		close(txChDone)
		close(txChData)
	}()

	// Чтение контейнера
	container, err := readContainer(key, c)
	if err != nil {
		txChErr <- err
		return
	}

	// Поиск файла.
	for _, file := range container.Files {

		// Порционная передача файла.
		if file.Name == fileName {

			fileSize := len(file.Content)

			for start := 0; start < fileSize; start += 1024 {

				select {
				case <-rxChBreak: // Ожидание сигнала - прекратить обработку.
					txChErr <- ExternalErr
					return

				default:
					end := start + 1024
					if end > fileSize {
						end = fileSize
					}
					txChData <- file.Content[start:end]

					progressPercent := float64(start+1024) / float64(fileSize) * 100
					if progressPercent > 100 {
						progressPercent = 100
					}
					txChPercent <- progressPercent
				}
			}
			txChPercent <- 100.0
			txChDone <- struct{}{} // Передача сигнала - обработка завершена.
			return
		}
	}

	// Если файл не найден
	txChErr <- fmt.Errorf("файл:<%s>, в контейнере, не найден", fileName)
}

// Получение списка файлов в контейнере. Возвращается массив имен файлов и ошибка.
//
// Параметры:
//
//	key - ключ шифрования.
func (c *Container) ListFilesInContainer(key [32]byte) (files []string, err error) {

	// чтение контейнера.
	container, err := readContainer(key, c)
	if err != nil {
		return nil, fmt.Errorf("функция c.readContainer, вернула ошибку:<%w>", err)
	}
	// Формирование списка из имён файлов.
	for _, file := range container.Files {
		files = append(files, file.Name)
	}

	// Результат.
	return files, nil
}

// Удаление файла из контейнера. Возвращается ошибка.
//
// Параметры:
//
//	fileName - имя файла.
//	key - ключ шифрования.
func (c *Container) RemoveFileFromContainer(fileName string, key [32]byte) error {

	// Чтение контейнера
	container, err := readContainer(key, c)
	if err != nil {
		return err
	}

	// Определение индекса файла для удаления
	var indexToRemove = -1
	for i, file := range container.Files {
		if file.Name == fileName {
			indexToRemove = i
			break
		}
	}
	if indexToRemove == -1 {
		return fmt.Errorf("файл <%s> не найден в контейнере", fileName)
	}

	// Удаление файла из контейнера.
	container.Files = append(container.Files[:indexToRemove], container.Files[indexToRemove+1:]...)

	// Запись обновленного контейнер обратно в файл
	if err := writeContainer(container, key, c); err != nil {
		return fmt.Errorf("ошибка при записи контейнера: <%w>", err)
	}

	return nil
}
