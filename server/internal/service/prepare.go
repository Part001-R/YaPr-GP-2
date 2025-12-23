package service

// Подготовительные действия, перед запуском сервиса.
import (
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/internal/utils/logger"
	"github.com/Part001-R/YaPr-GP-2/server/internal/grpc"
	"github.com/Part001-R/YaPr-GP-2/server/internal/service/udt"
)

// Подготовительные действия перед запуском сервиса. Возвращается конфигурация и ошибка.
func prepare() (*udt.Configuration, error) {

	// Создание логгера.
	ptrLgr, err := logger.New("debug")
	if err != nil {
		return nil, fmt.Errorf("функция logger.NewLogger, вернула ошибку: <%w>", err)
	}
	// Создание grpc.
	ptrGRPC := grpc.New(ptrLgr)

	// Пути к TLS файлам.
	pathTLSsert := "tls/server.crt"
	pathTLSPriv := "tls/server.key"

	// Создание сводной конфигурации сервиса.
	conf := udt.New(ptrLgr, ptrGRPC, pathTLSsert, pathTLSPriv)

	// Завершение.
	ptrLgr.Debug("подготовка пройдена")
	return conf, nil
}
