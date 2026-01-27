// Вспомогательные функции пакета.
package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Part001-R/YaPr-GP-2/server/internal/grpc"
	"github.com/Part001-R/YaPr-GP-2/server/internal/service/udt"
	"go.uber.org/zap"
)

// Причины остановки приложения
type checkReasonStop struct {
	chErrServ chan error     // остановка остановка сервера gRPCS
	sigSys    chan os.Signal // сигнал Ctrl+C
}

// createTLSConfig создание конфигурацию TLS. Возвращается конфигурация и ошибка.
//
// Параметры:
//
// c - указатель на конфигурацию сервиса.
func createTLSConfig(c *udt.Configuration) (*tls.Config, error) {

	cert, err := tls.LoadX509KeyPair(c.TLS.Public, c.TLS.Private)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки сертификатов: %w", err)
	}

	// Создание новой конфигурации TLS
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

// Создание дочерней директории. Возвращается ошибка.
//
// Параметры:
//
//	subdirName - имя дирекории.
func createSubdirectory(subdirName string) error {

	if _, err := os.Stat(subdirName); os.IsNotExist(err) {

		err := os.Mkdir(subdirName, 0755)
		if err != nil {
			return fmt.Errorf("ошибка при создании директории: %v", err)
		}
		fmt.Printf("Поддиректория <%s> успешно создана.\n", subdirName)
	}
	return nil
}

// Функция определяет причину остановки выполнения. Возвращается ошибка.
//
// Параметры:
//
// data - набор данных для обеспечения работы функции.
// params - параметры.
func signalsStopRun(data checkReasonStop, params *udt.Configuration) error {

	// Проверка аргументов
	if data.chErrServ == nil {
		params.Lgr.Error("в аргументе chErrServ нет указателя")
		return errors.New("в аргументе chErrServ нет указателя")
	}
	if data.sigSys == nil {
		params.Lgr.Error("в аргументе data.sigSys нет указателя")
		return errors.New("в аргументе data.sigSys нет указателя")
	}
	if params == nil {
		return errors.New("в аргументе params нет указателя")
	}

	// Закрытие подключения к БД
	defer func() {
		params.Storage.Close()
	}()

	// Проверка на nil для полей структуры
	if data.sigSys == nil {
		return errors.New("канал sigSys не инициализирован")
	}
	if data.chErrServ == nil {
		return errors.New("канал chErrServ не инициализирован")
	}

	// Логика
	select {
	case <-data.sigSys: // Обработка сигнала Ctrl+C

		tTick := time.NewTicker(1 * time.Second)

		done := false
		for !done {
			<-tTick.C
			busy := isBusyServer(params)
			if !busy {
				done = true
			}
		}
		params.Lgr.Info("сервер остановлен штатно")
		return nil

	case err := <-data.chErrServ: // Обработка сигнала ошибки
		params.Lgr.Error("ошибка сервера", zap.String("ошибка", err.Error()))
		return err
	}
}

// Проверка активности процессов. Возвращается true - сервер занят.
//
// Параметры:
//
//	params - параметры.
func isBusyServer(params *udt.Configuration) bool {

	if params.Srv.GetStatusBackUp() != grpc.StageNotActive ||
		params.Srv.GetStatusRestore() != grpc.StageNotActive ||
		params.Srv.GetStatusTx() != grpc.StageNotActive ||
		params.Srv.GetStatusRx() != grpc.StageNotActive {
		return true
	}

	return false
}
