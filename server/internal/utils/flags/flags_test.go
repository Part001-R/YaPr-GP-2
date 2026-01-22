// Тесты пакета.
package flags

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тест конструктора.
func TestNew(t *testing.T) {

	// переменные окружения.
	os.Setenv("SERVER_SUBDIR_FILES", "files")
	os.Setenv("SERVER_SUBDIR_BACKUP", "dirBackUp")
	os.Setenv("SERVER_DB_DSN", "file:test.db?cache=shared&foreign_keys=on&mode=rwc")
	os.Setenv("SERVER_PORT", "50101")

	defer func() {
		os.Unsetenv("SERVER_SUBDIR_FILES")
		os.Unsetenv("SERVER_SUBDIR_BACKUP")
		os.Unsetenv("SERVER_DB_DSN")
		os.Unsetenv("SERVER_PORT")
	}()

	wantSubDirFiles := "files"
	wantSubDirBackUp := "dirBackUp"
	wantDSN := "file:test.db?cache=shared&foreign_keys=on&mode=rwc"
	wantPort := "50101"

	// Вызов конструктора.
	config := New()

	//Проверки.
	assert.Equalf(t, wantSubDirFiles, config.NameSubDirFiles, "Нет соответствия директории файлов")
	assert.Equalf(t, wantSubDirBackUp, config.NameSubDirBackUp, "Нет соответствия директории бэкапа")
	assert.Equalf(t, wantDSN, config.DSN, "Нет соответствия dsn")
	assert.Equalf(t, wantPort, config.Port, "Нет соответствия порта")

}
