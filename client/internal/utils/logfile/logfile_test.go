// Тесты пакета.
package logfile

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Конструтор.
func TestNew(t *testing.T) {

	t.Run("Ошибка в аргументах", func(t *testing.T) {

		_, err := New("")
		assert.Equalf(t, EmptyDataArgumentNameLogFile, err, "Нет соответствия ошибки")
	})

	t.Run("Успешное создание", func(t *testing.T) {

		name := "testLog.txt"

		_, err := New(name)
		require.NoErrorf(t, err, "Ошибка создания логгера")

		err = os.Remove(name)
		assert.NoErrorf(t, err, "Ошибка удаления")
	})

}

// Запись.
func TestWrite(t *testing.T) {

	t.Run("Успешная запись", func(t *testing.T) {
		name := "testLog.txt"

		lgr, err := New(name)
		require.NoErrorf(t, err, "Ошибка создания логгера")
		defer func() {
			err = os.Remove(name)
			assert.NoErrorf(t, err, "Ошибка удаления")
		}()

		// Запись.
		dataWr := "Foo"
		err = lgr.Write(dataWr)
		require.NoErrorf(t, err, "Ошибка записи в файл")

		// Чтение файла.
		file, err := os.Open(name)
		require.NoErrorf(t, err, "Ошибка открытия файла")
		defer func() {
			err := file.Close()
			require.NoErrorf(t, err, "Ошибка закрытия файла")
		}()

		rxDAta, err := io.ReadAll(file)
		require.NoErrorf(t, err, "Ошибка чтения файла")

		// Проверка.
		assert.Truef(t, strings.Contains(string(rxDAta), dataWr), "Нет соответствия в записи")
	})

	t.Run("Проверка переполнения размера файла", func(t *testing.T) {
		name := "testLog.txt"

		lgr, err := New(name)
		require.NoErrorf(t, err, "Ошибка создания логгера")
		defer func() {
			err = os.Remove(name)
			assert.NoErrorf(t, err, "Ошибка удаления")
		}()

		// Получение списка файлов до записи.
		namesStart, err := os.ReadDir("./")
		require.NoErrorf(t, err, "Ошибка получения имён до записи")

		// Запись.
		dataWr := "Foo"
		for i := 0; i < 800000; i++ {
			err = lgr.Write(dataWr)
			require.NoErrorf(t, err, "Ошибка записи в файл")
		}

		// Получение списка файлов после записи.
		namesStop, err := os.ReadDir("./")
		require.NoErrorf(t, err, "Ошибка получения имён после записи")

		// Проверка.
		assert.Truef(t, len(namesStart) != len(namesStop), "Нет добавленного файла")

		// Удаление дополнительного файла лога.
		existingFiles := make(map[string]struct{})
		for _, file := range namesStart {
			existingFiles[file.Name()] = struct{}{}
		}

		for _, newFile := range namesStop {
			if _, exists := existingFiles[newFile.Name()]; !exists {
				os.Remove(newFile.Name())
			}
		}

	})

}
