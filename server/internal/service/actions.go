// Действия пакета.
package service

import (
	"fmt"
	"net"

	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/Part001-R/YaPr-GP-2/server/internal/service/udt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Функция содержит действия сервиса. Возвращается ошибка.
//
// Параметры:
//
// с - конфигурация сервиса.
func actions(c *udt.Configuration) error {

	port := c.Flag.Port

	// Проверка аргументов
	if c == nil {
		return NilPtrArgumentConf
	}
	if err := c.CheckConf(); err != nil {
		return fmt.Errorf("функция CheckConf, вернула ошибку: <%w>", err)
	}

	// Подключение к порту.
	lis, err := net.Listen("tcp", port)
	if err != nil {
		c.Lgr.Error("Ошибка подключения к порту",
			zap.String("ошибка", err.Error()),
			zap.String("порт", port))
		return fmt.Errorf("ошибка подключения к порту: <%w>", err)
	}

	// Создание конфигурации для TLS
	tlsConfig, err := createTLSConfig(c)
	if err != nil {
		c.Lgr.Error("Ошибка создания TLS конфигурации",
			zap.String("ошибка", err.Error()))
		return fmt.Errorf("ошибка создания TLS конфигурации: <%w>", err)
	}

	// Создание gRPC сервера с TLS
	s := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
		grpc.UnaryInterceptor(c.Srv.AuthInterceptorUnar),
		grpc.StreamInterceptor(c.Srv.AuthInterceptorStream),
	)

	pb.RegisterPasswordManagerServer(s, c.Srv)

	// Запуск gRPCS сервера.
	c.Lgr.Info("Запуск gRPCS сервера",
		zap.String("порт", port))

	if err := s.Serve(lis); err != nil {
		c.Lgr.Error("Ошибка в работе gRPC сервера",
			zap.String("ошибка", err.Error()))
		return fmt.Errorf("ошибка в работе gRPC сервера: <%w>", err)
	}

	return nil
}
