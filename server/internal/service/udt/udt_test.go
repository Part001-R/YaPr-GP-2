package udt

import (
	"os"
	"testing"

	"github.com/Part001-R/YaPr-GP-2/internal/utils/logger"
	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/server/internal/grpc"
	"github.com/Part001-R/YaPr-GP-2/server/internal/utils/flags"
	"github.com/stretchr/testify/require"
)

// CheckConf

func TestCheckConf(t *testing.T) {

	// Логгер.
	lgr, err := logger.New("debug")
	require.NoErrorf(t, err, "Ошибка создания логгера")

	// Флаги.
	flag := flags.New()

	// БД.
	storage, err := domain.NewStorage(flag.DSN)
	require.NoErrorf(t, err, "Ошибка создания хранилища")

	defer os.Remove("remote.db")

	// Создание grpc.
	srvGRPC, err := grpc.New(lgr, storage, flag)
	require.NoErrorf(t, err, "Ошибка создания сервера grpc")

	// Пути к TLS файлам.
	pathTLSsert := "tls/server.crt"
	pathTLSPriv := "tls/server.key"

	// Создание сводной конфигурации сервиса.
	conf := New(lgr, srvGRPC, pathTLSsert, pathTLSPriv, storage, flag)

	t.Run("Корректные данные", func(t *testing.T) {

		err = conf.CheckConf()
		require.NoErrorf(t, err, "Неожиданная ошибка проверки")
	})

	t.Run("Нет указателя на логгер", func(t *testing.T) {

		copy := conf.Lgr
		defer func() {
			conf.Lgr = copy
		}()
		conf.Lgr = nil

		err = conf.CheckConf()
		require.Equalf(t, NilPtrLogger, err, "Неожиданная ошибка проверки")
	})

	t.Run("Нет указателя на сервер", func(t *testing.T) {

		copy := conf.Srv
		defer func() {
			conf.Srv = copy
		}()
		conf.Srv = nil

		err = conf.CheckConf()
		require.Equalf(t, NilPtrServer, err, "Неожиданная ошибка проверки")
	})

	t.Run("Нет данных TLS.Private", func(t *testing.T) {

		copy := conf.TLS.Private
		defer func() {
			conf.TLS.Private = copy
		}()
		conf.TLS.Private = ""

		err = conf.CheckConf()
		require.Equalf(t, MissingPathPrivate, err, "Неожиданная ошибка проверки")
	})

	t.Run("Нет данных TLS.Public", func(t *testing.T) {

		copy := conf.TLS.Public
		defer func() {
			conf.TLS.Public = copy
		}()
		conf.TLS.Public = ""

		err = conf.CheckConf()
		require.Equalf(t, MissingPathSert, err, "Неожиданная ошибка проверки")
	})

	t.Run("Нет укзателя на хранилище", func(t *testing.T) {

		copy := conf.Storage
		defer func() {
			conf.Storage = copy
		}()
		conf.Storage = nil

		err = conf.CheckConf()
		require.Equalf(t, NilPtrStorage, err, "Неожиданная ошибка проверки")
	})

	t.Run("Нет укзателя на флаги", func(t *testing.T) {

		copy := conf.Flag
		defer func() {
			conf.Flag = copy
		}()
		conf.Flag = nil

		err = conf.CheckConf()
		require.Equalf(t, NilPtrFlag, err, "Неожиданная ошибка проверки")
	})

	t.Run("Нет данных порта", func(t *testing.T) {

		copy := conf.Flag.Port
		defer func() {
			conf.Flag.Port = copy
		}()
		conf.Flag.Port = ""

		err = conf.CheckConf()
		require.Equalf(t, EmptyDataPort, err, "Неожиданная ошибка проверки")
	})
}
