package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Part001-R/YaPr-GP-2/proto"
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

// Обработчик проверки связи.
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

// Обработчик приёма файла.
func (s *PasswordManager) BackupFile(stream pb.PasswordManager_BackupFileServer) error {

	var fileName string

	// Приём данных файла.
	for {
		req, err := stream.Recv()
		if err == io.EOF { // конец потока
			break
		}
		if err != nil {
			s.ptrLogger.Error("ошибка при получении данных файла", zap.String("файл", fileName), zap.String("ошибка", err.Error()))
			return status.Error(codes.Internal, fmt.Sprintf("ошибка при чтении потока: <%v>", err))
		}

		if fileName == "" {
			fileName = req.GetFilename()
		}

		// Сохранение файла
		err = appendToFile(fileName, req.GetContent())
		if err != nil {
			s.ptrLogger.Error("ошибка при добавлении данных файла", zap.String("файл", fileName), zap.String("ошибка", err.Error()))
			return status.Error(codes.Internal, fmt.Sprintf("ошибка при добавлении принятых данных: <%v>", err))
		}
	}

	s.ptrLogger.Info("От клиента принят файл", zap.String("имя", fileName))

	// Финальное сообщение сервера.
	return stream.SendAndClose(&pb.UploadResponse{Message: fileName})
}

// Обработчик передачи файлов.
func (s *PasswordManager) RestoreFile(req *pb.DownloadRequest, stream pb.PasswordManager_RestoreFileServer) error {

	// Загружаем файл; пример кода для чтения файла.
	fileContent, err := os.ReadFile(req.Filename)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Запрошенный файл <%s>, отсутствует", req.Filename))
	}

	// Передача файла через stream.
	size := 1024

	for i := 0; i < len(fileContent); i += size {
		end := i + size
		if end > len(fileContent) {
			end = len(fileContent)
		}

		if err := stream.Send(&pb.DownloadResponse{
			Filename: req.Filename,
			Content:  fileContent[i:end],
		}); err != nil {
			return err
		}
	}

	s.ptrLogger.Info("Клиенту отправлен файл", zap.String("имя", req.Filename))

	return nil
}

// Предоставление информации о файлах.
func (s *PasswordManager) FilesInfo(ctx context.Context, req *emptypb.Empty) (*proto.FilesInfoResponse, error) {

	s.ptrLogger.Debug("Принят FilesInfo запрос")

	// Подготовка.
	fileInfos := &pb.FilesInfoResponse{}
	files := []string{"manager.db", "container.data"}

	// Сбор информации по файлам.
	for _, f := range files {
		var el pb.FileInfo

		size, err := sizeFile(f)
		if err != nil {
			if errors.Is(err, NotFound) { // Если файл не найден - обработка следующего файла.
				continue
			}
			return nil, status.Error(codes.Internal, fmt.Sprintf("Ошибка:<%v>, при обработке файла:<%s>", err, f))
		}

		el.FileName = f
		el.Size = size

		fileInfos.FileInfo = append(fileInfos.FileInfo, &el)
	}

	// Результат.
	return fileInfos, nil
}
