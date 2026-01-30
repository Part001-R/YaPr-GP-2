// Тесты пакета.
package udt

import (
	"os"
	"testing"

	"github.com/Part001-R/YaPr-GP-2/client/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/logfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Конструктор.
func TestNew(t *testing.T) {

	// Логгер файла.
	lgrFile, err := logfile.New("log.txt")
	require.NoErrorf(t, err, "Ошибка создания логгера")
	defer func() {
		err := os.Remove("log.txt")
		assert.NoErrorf(t, err, "Ошибка удаления файла логгера")
	}()

	// Флаги.
	flg := flags.New()

	// БД.
	storage, err := domain.NewStorage(flg.DSN)
	require.NoErrorf(t, err, "Ошибка создания хранилища")
	defer func() {
		err := storage.Close()
		assert.NoErrorf(t, err, "Ошибка отключения от хранилища")

		err = os.Remove("localStorage.db")
		assert.NoErrorf(t, err, "Ошибка удаления файла БД")
	}()

	t.Run("Успешное создание", func(t *testing.T) {

		_, err := New(lgrFile, storage, flg)
		require.NoErrorf(t, err, "Ошибка конструктора")
	})

	t.Run("Нет логгера", func(t *testing.T) {

		_, err := New(nil, storage, flg)
		require.Equalf(t, NilPtrArgumentL, err, "нет соответствия ошибки")
	})

	t.Run("Нет флагов", func(t *testing.T) {

		_, err := New(lgrFile, storage, nil)
		require.Equalf(t, NilPtrArgumentF, err, "нет соответствия ошибки")
	})
}
