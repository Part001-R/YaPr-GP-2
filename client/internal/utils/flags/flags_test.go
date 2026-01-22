// Тесты пакета.
package flags

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест конструктора.
func TestNew(t *testing.T) {

	// Установим флаги окружения для теста
	os.Setenv("MODE_CLIENT", "local")
	os.Setenv("DSN_STORAGE", "file:test.db?cache=shared&foreign_keys=on&mode=rwc")
	os.Setenv("LOCAL_NAME_CONTAINER", "testContainer")

	defer func() {
		os.Unsetenv("MODE_CLIENT")
		os.Unsetenv("DSN_STORAGE")
		os.Unsetenv("LOCAL_NAME_CONTAINER")
	}()

	wantMode := "local"
	wantDSN := "file:test.db?cache=shared&foreign_keys=on&mode=rwc"
	wantContainer := "testContainer"
	wantDBName := "test.db"

	// Вызов конструктора.
	config := New()

	//Проверки.
	assert.Equalf(t, wantMode, config.Mode, "Нет соответствия режима работы")
	assert.Equalf(t, wantDSN, config.DSN, "Нет соответствия dsn")
	assert.Equalf(t, wantContainer, config.LocalNameContainer, "Нет соответствия имени контейнера")

	dbName, err := GetNameDBFromDSN(config.DSN)
	require.NoErrorf(t, err, "Ошибка получения имени БД из DSN")
	assert.Equalf(t, wantDBName, dbName, "Нет соответствия имени БД")

	err = checkFlags(*config)
	require.NoErrorf(t, err, "Ошибка проверки флагов")
}
