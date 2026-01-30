package service

import (
	"os"
	"testing"

	"github.com/Part001-R/YaPr-GP-2/server/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/server/internal/utils/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест подготовительных действий и проверки конфигурации.
func TestPrepare(t *testing.T) {

	t.Run("Успешное конфигурирование", func(t *testing.T) {

		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка подготовки")

		nameDB, err := flags.GetNameDBFromDSN(conf.Flag.DSN)
		require.NoErrorf(t, err, "Ошибка выделения имени БД из dsn")

		err = os.Remove(nameDB)
		require.NoErrorf(t, err, "Ошибка удаления файла БД")

		// Тест проверки конфигурации.
		err = conf.CheckConf()
		require.NoErrorf(t, err, "Ошибка в конфигурации")

	})

	t.Run("Нет указателя на флаги", func(t *testing.T) {

		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка подготовки")

		ptr := conf.Flag
		conf.Flag = nil

		// Тест проверки конфигурации.
		err = conf.CheckConf()
		assert.Equalf(t, udt.NilPtrFlag, err, "Нет соответствия ошибки")

		conf.Flag = ptr
	})

	t.Run("Нет указателя на логгер", func(t *testing.T) {

		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка подготовки")

		ptr := conf.Lgr
		conf.Lgr = nil

		// Тест проверки конфигурации.
		err = conf.CheckConf()
		assert.Equalf(t, udt.NilPtrLogger, err, "Нет соответствия ошибки")

		conf.Lgr = ptr
	})

	t.Run("Нет указателя на сервер", func(t *testing.T) {

		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка подготовки")

		ptr := conf.Srv
		conf.Srv = nil

		// Тест проверки конфигурации.
		err = conf.CheckConf()
		assert.Equalf(t, udt.NilPtrServer, err, "Нет соответствия ошибки")

		conf.Srv = ptr
	})

	t.Run("Нет указателя на хранилище", func(t *testing.T) {

		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка подготовки")

		ptr := conf.Storage
		conf.Storage = nil

		// Тест проверки конфигурации.
		err = conf.CheckConf()
		assert.Equalf(t, udt.NilPtrStorage, err, "Нет соответствия ошибки")

		conf.Storage = ptr
	})

	t.Run("Нет информации по TLS Private", func(t *testing.T) {

		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка подготовки")

		data := conf.TLS.Private
		conf.TLS.Private = ""

		// Тест проверки конфигурации.
		err = conf.CheckConf()
		assert.Equalf(t, udt.MissingPathPrivate, err, "Нет соответствия ошибки")

		conf.TLS.Private = data
	})

	t.Run("Нет информации по TLS Public", func(t *testing.T) {

		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка подготовки")

		data := conf.TLS.Public
		conf.TLS.Public = ""

		// Тест проверки конфигурации.
		err = conf.CheckConf()
		assert.Equalf(t, udt.MissingPathSert, err, "Нет соответствия ошибки")

		conf.TLS.Public = data
	})

}
