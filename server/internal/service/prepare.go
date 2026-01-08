package service

// Подготовительные действия, перед запуском сервиса.
import (
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/internal/utils/logger"
	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/server/internal/grpc"
	"github.com/Part001-R/YaPr-GP-2/server/internal/service/udt"
)

// Подготовительные действия перед запуском сервиса. Возвращается конфигурация и ошибка.
func prepare() (*udt.Configuration, error) {

	// Создание логгера.
	lgr, err := logger.New("debug")
	if err != nil {
		return nil, fmt.Errorf("функция logger.NewLogger, вернула ошибку: <%w>", err)
	}

	// БД.
	storage, err := domain.NewStorage("file:remote.db?cache=shared&foreign_keys=on&mode=rwc")
	if err != nil {
		return nil, fmt.Errorf("функция domain.NewStorage, вернула ошибку: <%w>", err)
	}

	// Создание grpc.
	srvGRPC := grpc.New(lgr, storage)

	// Пути к TLS файлам.
	pathTLSsert := "tls/server.crt"
	pathTLSPriv := "tls/server.key"

	// Создание сводной конфигурации сервиса.
	conf := udt.New(lgr, srvGRPC, pathTLSsert, pathTLSPriv, storage)

	// Завершение.
	lgr.Debug("подготовка пройдена")
	return conf, nil
}
