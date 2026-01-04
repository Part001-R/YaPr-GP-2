package container

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	onceInst sync.Once // для разовой инициализации экземпляра.

	nonceSize = 24
)

// Структура записи в контейнере
type FileEntry struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

// Защифрованный контейнер.
type Container struct {
	name  string      // имя контейнера.
	key   [32]byte    // ключ шифрования.
	Files []FileEntry `json:"files"` // файлы в контейнере.
}

// Интерфейс.
type Actions interface {
	AddFileToContainer(fileName string, key [32]byte) error
	GetFileFromContainer(fileName string, key [32]byte) ([]byte, error)
	ListFilesInContainer(key [32]byte) (files []string, err error)
	RemoveFileFromContainer(fileName string, key [32]byte) error
}

// Экземпляр.
var inst *Container

// Конструктор.
func New(name string, key [32]byte) Actions {
	onceInst.Do(func() {
		// Проверка присутствия файла контейнера.
		if _, err := os.Stat(name); os.IsNotExist(err) {

			// Если файла нет - создание
			emptyContainer := Container{
				name: name,
				key:  key,
			}
			data, err := json.Marshal(emptyContainer)
			if err != nil {
				log.Fatalf("Ошибка сериализации контейнера: <%v>", err)
				return
			}

			if err := os.WriteFile(name, data, 0644); err != nil {
				log.Fatalf("Ошибка создания файла контейнера: <%v>", err)
				return
			}
		}
		// Инициализация экземпляра.
		inst = &Container{
			name: name,
			key:  key,
		}
	})
	// Рузультат.
	return inst
}

// Чтение контейнера из файла.
func (c Container) readContainer(key [32]byte) (*Container, error) {
	data, err := os.ReadFile(c.name)
	if err != nil {
		if os.IsNotExist(err) {
			return &Container{}, nil // пустой контейнер
		}
		return nil, fmt.Errorf("Функция os.ReadFile, вернула ошибку:<%w>, при чтении файла:<%s>", err, c.name)
	}

	var container Container
	err = json.Unmarshal(data, &container)
	if err != nil {
		return nil, fmt.Errorf("Функция json.Unmarshal, вернула ошибку:<%w>, при обработке данных файла:<%s>", err, c.name)
	}

	// Дешифруем содержимое каждого файла
	for i, file := range container.Files {
		decrypted, err := decrypt(file.Content, key)
		if err != nil {
			return nil, fmt.Errorf("не удалось дешифровать файл %s: %v", file.Name, err)
		}
		container.Files[i].Content = decrypted
	}

	return &container, nil
}

// Запись контейнера в файл.
func (c *Container) writeContainer(container *Container, key [32]byte) error {

	// Шифрование каждого файла перед сохранением.
	encryptedFiles := make([]FileEntry, len(container.Files))
	for i, file := range container.Files {
		encrypted, err := encrypt(file.Content, key)
		if err != nil {
			return err
		}
		encryptedFiles[i] = FileEntry{
			Name:    file.Name,
			Content: encrypted,
		}
	}

	containerToSave := Container{Files: encryptedFiles}
	data, err := json.Marshal(containerToSave)
	if err != nil {
		return err
	}

	return os.WriteFile(c.name, data, 0644)
}

// Добавление файла в контейнер.
func (c *Container) AddFileToContainer(fileName string, key [32]byte) error {

	// Чтение содиржимого файла.
	content, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("ошибка: <%v>, при чтении файла: <%s>", err, fileName)
	}

	container, err := c.readContainer(key)
	if err != nil {
		return err
	}

	// Выделение имени файла и его тип из полного пути.
	fileN := getFileNameAndExtension(fileName)

	// Проверка на присутствие такого файла в контейнере.
	for _, file := range container.Files {
		if file.Name == fileN {
			return fmt.Errorf("файл <%s> уже существует в контейнере", fileN)
		}
	}

	// Добавление файла в контейнер.
	container.Files = append(container.Files, FileEntry{
		Name:    fileN,
		Content: content,
	})

	if err := c.writeContainer(container, key); err != nil {
		return fmt.Errorf("функция writeContainer, вернула ошибку: <%w>", err)
	}

	return nil
}

// Получение файла из контейнера по имени
func (c Container) GetFileFromContainer(fileName string, key [32]byte) ([]byte, error) {

	// Чтение контейнера
	container, err := c.readContainer(key)
	if err != nil {
		return nil, err
	}

	// Поиск файла
	for _, file := range container.Files {
		if file.Name == fileName {
			return file.Content, nil
		}
	}

	return nil, fmt.Errorf("файл <%s>, не найден в контейнере", fileName)
}

// Получение списка файлов в контейнере.
func (c Container) ListFilesInContainer(key [32]byte) (files []string, err error) {

	// чтение контейнера.
	container, err := c.readContainer(key)
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

// Удаление файла из контейнера.
func (c *Container) RemoveFileFromContainer(fileName string, key [32]byte) error {
	// Чтение контейнера
	container, err := c.readContainer(key)
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

	// Записываем обновленный контейнер обратно в файл
	if err := c.writeContainer(container, key); err != nil {
		return fmt.Errorf("ошибка при записи контейнера: <%w>", err)
	}

	return nil
}
