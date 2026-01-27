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

// Тест выделения имени БД из dsn.
func TestGetNameDBFromDSN(t *testing.T) {

	t.Run("dsn для SQlite", func(t *testing.T) {
		dsn := "file:foo.db?cache=shared&foreign_keys=on&mode=rwc"
		wantName := "foo.db"

		name, err := GetNameDBFromDSN(dsn)
		require.NoErrorf(t, err, "ошибка выделения имени БД")
		assert.Equalf(t, wantName, name, "нет соответствия имени")
	})

}

// Тест установки флагов из переменных окружения.
func TestSetFlagsFromEnv(t *testing.T) {
	t.Run("Нет конфигурации", func(t *testing.T) {
		var cfg *Config = nil
		err := setFlagsFromEnv(cfg)
		require.Error(t, err)
		assert.Equal(t, ErrNilPtrArgumentF, err)
	})

	t.Run("все переменные окружения заданы", func(t *testing.T) {
		t.Setenv("SERVER_SUBDIR_FILES", "uploads")
		t.Setenv("SERVER_SUBDIR_BACKUP", "backups")
		t.Setenv("SERVER_DB_DSN", "postgres://user:pass@localhost/db")
		t.Setenv("SERVER_PORT", "8080")

		cfg := &Config{}
		err := setFlagsFromEnv(cfg)
		require.NoError(t, err)

		assert.Equal(t, "uploads", cfg.NameSubDirFiles)
		assert.Equal(t, "backups", cfg.NameSubDirBackUp)
		assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DSN)
		assert.Equal(t, "8080", cfg.Port)
	})

	t.Run("Заданы DSN и PORT", func(t *testing.T) {
		t.Setenv("SERVER_SUBDIR_FILES", "")
		t.Setenv("SERVER_SUBDIR_BACKUP", "")
		t.Setenv("SERVER_DB_DSN", "sqlite:///data.db")
		t.Setenv("SERVER_PORT", "9000")

		cfg := &Config{
			NameSubDirFiles:  "default_files",
			NameSubDirBackUp: "default_backup",
			DSN:              "old_dsn",
			Port:             "80",
		}
		err := setFlagsFromEnv(cfg)
		require.NoError(t, err)

		assert.Equal(t, "default_files", cfg.NameSubDirFiles)
		assert.Equal(t, "default_backup", cfg.NameSubDirBackUp)
		assert.Equal(t, "sqlite:///data.db", cfg.DSN)
		assert.Equal(t, "9000", cfg.Port)
	})

	t.Run("все переменные пустые", func(t *testing.T) {
		t.Setenv("SERVER_SUBDIR_FILES", "")
		t.Setenv("SERVER_SUBDIR_BACKUP", "")
		t.Setenv("SERVER_DB_DSN", "")
		t.Setenv("SERVER_PORT", "")

		cfg := &Config{
			NameSubDirFiles:  "prev_files",
			NameSubDirBackUp: "prev_backup",
			DSN:              "prev_dsn",
			Port:             "prev_port",
		}
		err := setFlagsFromEnv(cfg)
		require.NoError(t, err)

		assert.Equal(t, "prev_files", cfg.NameSubDirFiles)
		assert.Equal(t, "prev_backup", cfg.NameSubDirBackUp)
		assert.Equal(t, "prev_dsn", cfg.DSN)
		assert.Equal(t, "prev_port", cfg.Port)
	})
}

// Тест изменения порта.
func TestChangePortForTest(t *testing.T) {

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

	wantPort := "50102"

	// Вызов конструктора.
	config := New()

	ChangePortForTest(wantPort)

	//Проверки.
	assert.Equalf(t, wantPort, config.Port, "Нет соответствия порта")

}
