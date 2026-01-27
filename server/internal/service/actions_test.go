package service

import (
	"testing"

	"github.com/Part001-R/YaPr-GP-2/server/internal/utils/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActions(t *testing.T) {

	t.Run("Нет указателя на конфигурацию", func(t *testing.T) {

		err := actions(nil)
		assert.Equalf(t, NilPtrArgumentConf, err, "Нет соответствия ошибки")
	})

	t.Run("Нет указателя на логгер", func(t *testing.T) {
		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания конфигурации")

		copy := conf.Lgr
		defer func() {
			conf.Lgr = copy
		}()
		conf.Lgr = nil // Сброс указателя на логгер.

		err = actions(conf)
		assert.Equalf(t, "функция CheckConf, вернула ошибку: <нет указателя на логгер>", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Нет указателя на флаги", func(t *testing.T) {
		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания конфигурации")

		copy := conf.Flag
		defer func() {
			conf.Flag = copy
		}()
		conf.Flag = nil // Сброс указателя на флаги.

		err = actions(conf)
		assert.Equalf(t, "функция CheckConf, вернула ошибку: <нет указателя на флаги>", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Нет указателя на сервер", func(t *testing.T) {
		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания конфигурации")

		copy := conf.Srv
		defer func() {
			conf.Srv = copy
		}()
		conf.Srv = nil // Сброс указателя на сервер.

		err = actions(conf)
		assert.Equalf(t, "функция CheckConf, вернула ошибку: <нет указателя на сервер>", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Нет указателя на хранилище", func(t *testing.T) {
		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания конфигурации")

		copy := conf.Storage
		defer func() {
			conf.Storage = copy
		}()
		conf.Storage = nil // Сброс указателя на хранилище.

		err = actions(conf)
		assert.Equalf(t, "функция CheckConf, вернула ошибку: <нет указателя на хранилище>", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Нет TLS.Private", func(t *testing.T) {
		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания конфигурации")

		copy := conf.TLS.Private
		defer func() {
			conf.TLS.Private = copy
		}()
		conf.TLS.Private = "" // Сброс TLS.Private.

		err = actions(conf)
		assert.Equalf(t, "функция CheckConf, вернула ошибку: <нет пути к файлу приватного ключа>", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Нет TLS.Public", func(t *testing.T) {
		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания конфигурации")

		copy := conf.TLS.Public
		defer func() {
			conf.TLS.Public = copy
		}()
		conf.TLS.Public = "" // Сброс TLS.Public.

		err = actions(conf)
		assert.Equalf(t, "функция CheckConf, вернула ошибку: <нет пути к файлу сертификата>", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Нет данных порта", func(t *testing.T) {
		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания конфигурации")

		copy := conf.Flag.Port
		defer func() {
			conf.Flag.Port = copy
		}()
		conf.Flag.Port = "" // Сброс порта.

		err = actions(conf)
		assert.Equalf(t, "функция CheckConf, вернула ошибку: <нет указания порта>", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Ошибка в номере порта", func(t *testing.T) {

		fl := flags.New()
		copy := fl.Port
		defer func() {
			fl.Port = copy
		}()
		fl.Port = "Foo" // Сброс порта.

		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания конфигурации")

		err = actions(conf)
		assert.Equalf(t, "ошибка подключения к порту: <listen tcp: lookup tcp/Foo: unknown port>", err.Error(), "Нет соответствия ошибки")
	})

}
