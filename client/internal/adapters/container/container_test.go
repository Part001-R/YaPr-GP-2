package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест создания экземпляра контейнера.
func TestNew(t *testing.T) {

	nameContainer := "test_container"

	var testKey [32]byte
	for i := range testKey {
		testKey[i] = byte(i)
	}
	_ = os.Remove(nameContainer)

	t.Run("Ошибка в аргументах", func(t *testing.T) {

		_, err := New("", testKey)
		require.Equalf(t, EmptyDataArgumentNAme, err, "Нет соответствия ошибки")

	})

	t.Run("Создание файла и инициализация контейнера", func(t *testing.T) {

		instContainer, err := New(nameContainer, testKey)
		require.NoError(t, err, "неожиданная ошибка")

		assert.NotNil(t, instContainer)

		_, err = os.Stat(nameContainer)
		assert.NoError(t, err, "Файл должен быть создан")

		fileData, err := os.ReadFile(nameContainer)
		assert.NoError(t, err, "Не удалось прочитать файл")

		var container Container
		err = json.Unmarshal(fileData, &container)
		assert.NoError(t, err, "Ошибка десериализации JSON")

		assert.Equal(t, nameContainer, container.Name, "Имя контейнера должно совпадать")
		assert.Equal(t, testKey, container.Key, "Ключ контейнера должен совпадать")
	})

	t.Run("Повторный вызов New (файл уже существует)", func(t *testing.T) {
		originalInst := inst
		result, err := New(nameContainer, testKey)
		require.NoError(t, err, "неожиданная ошибка")
		assert.Equal(t, originalInst, result, "Повторный вызов не должен создавать новый экземпляр")
	})

	t.Cleanup(func() {
		_ = os.Remove(nameContainer)
	})
}

// Тест добавления файла в контейнер и получение файла из контейнера.
func TestContainer_PushPopDelFileContainer(t *testing.T) {

	nameContainer := "test_container"

	var key [32]byte
	for i := range key {
		key[i] = byte(i)
	}
	_ = os.Remove(nameContainer)

	// Создание контейнера.
	instContainer, err := New(nameContainer, key)
	require.NoError(t, err, "неожиданная ошибка")

	// Создание временного файла.
	fileName := "Foo.txt"
	fullFileName, err := createFile(fileName)
	require.NoErrorf(t, err, "неожиданная ошиббка создания файла: <%v>", err)

	t.Run("Добавление файла в контейнер", func(t *testing.T) {
		txChProcess := make(chan float64)
		txChErr := make(chan error)
		txChDone := make(chan struct{})

		// Добавление файла в контейнер.
		go instContainer.AddFileToContainer(fullFileName, key, txChProcess, txChErr, txChDone)

		done := false
		for !done {
			select {
			case <-txChProcess:

			case err = <-txChErr:
				assert.NoError(t, err, "Неожиданная ошибка")
				done = true

			case <-txChDone:
				done = true
			}
		}

		_ = os.Remove(fullFileName)
	})

	t.Run("Проверка существования файла в контейнере", func(t *testing.T) {

		filesName, err := instContainer.ListFilesInContainer(key)
		assert.NoErrorf(t, err, "Неожиданная ошибка при получении имён файлов:<%w>", err)

		ok := false
		for _, v := range filesName {
			if v == fileName {
				ok = true
			}
		}
		assert.True(t, ok, "имя файла не найдено")
	})

	t.Run("Извлечение файла из контейнера", func(t *testing.T) {

		txChProcess := make(chan float64)
		txChErr := make(chan error)
		txChDone := make(chan struct{})
		txChData := make(chan []byte)
		rxChBreak := make(chan struct{})

		defer func() {
			close(rxChBreak)
		}()

		outFile, err := os.Create(fullFileName)
		require.NoErrorf(t, err, "неожиданная ошибка при создании файла:<%v>", err)

		done := false

		// Чтение файла из контейнера.
		go instContainer.GetFileFromContainer(fileName, key, txChProcess, txChErr, txChDone, txChData, rxChBreak)

		// Обработка каналов.
		for !done {
			select {
			case <-txChProcess:

			case err := <-txChErr:
				assert.NoErrorf(t, err, "Неожиданная ошибка:<%v>", err)
				done = true

			case <-txChDone:
				done = true

			// Сборка файла.
			case data := <-txChData:
				if isFileExists(fullFileName) {
					if _, err := outFile.Write(data); err != nil {
						assert.NoErrorf(t, err, "неожиданная ошибка при записи в файл:<%v>", err)
						rxChBreak <- struct{}{}
					}
				}
			}
		}
	})

	t.Run("Проверка существования извлечённого файла", func(t *testing.T) {
		assert.Truef(t, isFileExists(fullFileName), "отсутствует файл:<%s>", fullFileName)
	})

	t.Run("Удаление файла из контейнера", func(t *testing.T) {
		err := instContainer.RemoveFileFromContainer(fileName, key)
		require.NoErrorf(t, err, "неожиданная ошибка при удалении файла из контейнера:<%v>", err)
	})

	t.Run("Проверка существования файла в контейнере, после удаления", func(t *testing.T) {

		filesName, err := instContainer.ListFilesInContainer(key)
		assert.NoErrorf(t, err, "Неожиданная ошибка при получении имён файлов:<%w>", err)

		ok := false
		for _, v := range filesName {
			if v == fileName {
				ok = true
			}
		}
		assert.False(t, ok, "Файл не удалён")
	})

	// Освобождение ресурсов.
	t.Cleanup(func() {
		_ = os.Remove(nameContainer)
		_ = os.Remove(fullFileName)
	})
}

//
// -- вспомогательные функции ---
//

// Создание файла. Возвращается полный путь к файлу и ошибка.
//
// Параметры:
//
//	fileName - имя файла.
func createFile(fileName string) (fileFullName string, err error) {

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("не удалось получить текущую директорию: %v", err)
	}
	filePath := filepath.Join(dir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("не удалось создать файл %s: %v", filePath, err)
	}
	defer file.Close()

	return filePath, nil
}

// Проверка существования файла. Возвращается true - файл существует.
//
// Параметры:
//
//	fullFileName - полное имя файла.
func isFileExists(fullFileName string) bool {
	_, err := os.Stat(fullFileName)
	if os.IsNotExist(err) {
		return false // Файл не существует
	}
	return err == nil // Файл существует
}
