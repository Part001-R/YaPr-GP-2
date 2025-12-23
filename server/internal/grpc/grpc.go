package grpc

import (
	"context"
	"sync"

	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// grpc
type PasswordManager struct {
	pb.UnimplementedPasswordManagerServer
	ptrLogger *zap.Logger
}

var once sync.Once        // единоразовая инициализация экземпляра
var inst *PasswordManager // экземпляр

// Конструктор.
func New(l *zap.Logger) *PasswordManager {
	once.Do(func() {
		inst = &PasswordManager{
			UnimplementedPasswordManagerServer: pb.UnimplementedPasswordManagerServer{},
			ptrLogger:                          l,
		}
	})
	return inst
}

// Функция проверки связи.
func (s *PasswordManager) Ping(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {

	s.ptrLogger.Debug("Принят Ping запрос")

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.ptrLogger.Warn("В принятом Ping запросе, отсутствуют метаданные")
		return nil, status.Error(codes.NotFound, "отсутствуют метаданные")
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 || tokens[0] == "" {
		s.ptrLogger.Warn("В метаданных, принятого Ping запросе, нет токена ")
		return nil, status.Error(codes.InvalidArgument, "в метаданных нет токена")
	}

	// Создание метаданных для ответа
	txMD := metadata.Pairs(nameToken, tokens[0]) // возврат принятого токена

	if err := grpc.SendHeader(ctx, txMD); err != nil {
		s.ptrLogger.Warn("В принятом Ping запросе, ошибка при отправке заголовков ")
		return nil, status.Error(codes.Internal, "ошибка при отправке заголовков")
	}

	return &emptypb.Empty{}, nil
}
