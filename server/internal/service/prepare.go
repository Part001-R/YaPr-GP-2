// Подготовительные действия.
package service

// Подготовительные действия, перед запуском сервиса.
import (
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/internal/utils/logger"
	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/server/internal/grpc"
	"github.com/Part001-R/YaPr-GP-2/server/internal/service/udt"
	"github.com/Part001-R/YaPr-GP-2/server/internal/utils/flags"
)

// Подготовительные действия перед запуском сервиса. Возвращается конфигурация и ошибка.
func prepare() (*udt.Configuration, error) {

	// Создание логгера.
	lgr, err := logger.New("debug")
	if err != nil {
		return nil, fmt.Errorf("функция logger.NewLogger, вернула ошибку: <%w>", err)
	}

	// Флаги.
	flag := flags.New()

	// Создание дирекотрии для файлов.
	if err := createSubdirectory(flag.NameSubDirFiles); err != nil {
		return nil, fmt.Errorf("Создание директории для файлов. функция createSubdirectory, вернула ошибку: <%w>", err)
	}
	// Создание дирекотрии для backUp.
	if err := createSubdirectory(flag.NameSubDirBackUp); err != nil {
		return nil, fmt.Errorf("Создание директории для backUp. функция createSubdirectory, вернула ошибку: <%w>", err)
	}

	// БД.
	storage, err := domain.NewStorage(flag.DSN)
	if err != nil {
		return nil, fmt.Errorf("функция domain.NewStorage, вернула ошибку: <%w>", err)
	}

	// Создание grpc.
	srvGRPC := grpc.New(lgr, storage, flag)

	// Пути к TLS файлам.
	pathTLSsert := "tls/server.crt"
	pathTLSPriv := "tls/server.key"

	// Создание сводной конфигурации сервиса.
	conf := udt.New(lgr, srvGRPC, pathTLSsert, pathTLSPriv, storage, flag)

	// Завершение.
	lgr.Debug("подготовка пройдена")
	return conf, nil
}
