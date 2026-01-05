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
func (s *PasswordManager) BackupFile(stream pb.PasswordManager_BackupFileServer) (errReturn error) {

	// Извлечение метаданных из контекста
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.InvalidArgument, "ошибка извлечения заголовков")
	}

	// Получение значения токена
	token := md.Get("token")
	if len(token) == 0 {
		return status.Error(codes.Unauthenticated, "токен не передан")
	}

	//
	// Логика
	//

	var rxFileName string
	var file *os.File
	var doDeferRemoveFile bool // Флаг необходимости удаления файла
	var isRemovedFile bool     // Флаг, что файл уже был удалён

	// Проверка необходимости удаления файла, перед выходом.
	defer func(doRemove bool, fileName string) {
		if fileExists(fileName) && doRemove {
			if errRemove := os.Remove(fileName); errRemove != nil {
				s.ptrLogger.Error("ошибка при удалении файла",
					zap.String("файл", fileName),
					zap.String("ошибка", errRemove.Error()),
					zap.String("причина удаления", errReturn.Error()))

				errReturn = fmt.Errorf("ошибка при удалении файла:<%v>, файл:<%s>, ошибка:<%v>, причина удаления:<%v>", errRemove, fileName, errRemove, errReturn)
			}
		}
	}(doDeferRemoveFile, rxFileName)

	// Приём данных файла.
	for {
		req, err := stream.Recv()
		if err == io.EOF { // конец потока
			break
		}
		if err != nil {
			s.ptrLogger.Error("ошибка stream.Recv", zap.String("ошибка", err.Error()))
			break
		}

		// Создание файла.
		if rxFileName == "" {
			rxFileName = req.GetFileName()

			// Предварительное удаление уже существующего файла.
			if !isRemovedFile {
				isRemovedFile = true

				if fileExists(rxFileName) {
					if err := os.Remove(rxFileName); err != nil {
						errReturn = status.Error(codes.Internal, fmt.Sprintf("ошибка:<%v>, предварительного удаления существующего файла:<%s>", err, rxFileName))
						return errReturn
					}
				}
			}

			// Новая версия файла.
			file, err = os.OpenFile(rxFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				s.ptrLogger.Error("ошибка при открытии файла", zap.String("файл", rxFileName), zap.String("ошибка", err.Error()))
				errReturn = status.Error(codes.Internal, fmt.Sprintf("ошибка:<%v>, открытия файла:<%s>", err, rxFileName))
				return errReturn
			}
			defer func() {
				if err := file.Close(); err != nil {
					s.ptrLogger.Error("ошибка закрытия подключения к файлу", zap.String("файл", rxFileName), zap.String("ошибка", err.Error()))
				}
			}()
		}

		// Сохранение принятых данных потока, в файл.
		_, err = file.Write(req.GetContent())
		if err != nil {
			s.ptrLogger.Error("ошибка при добавлении данных файла", zap.String("файл", rxFileName), zap.String("ошибка", err.Error()))
			doDeferRemoveFile = true
			errReturn = status.Error(codes.Internal, fmt.Sprintf("ошибка:<%v>, добавления данных в файл:<%s>", err, rxFileName))
			return errReturn
		}
	}

	// Вычисление хэша файла.
	fileHash, err := hashFile(rxFileName)
	if err != nil {
		s.ptrLogger.Error("ошибка при вычислении эеша для файла", zap.String("файл", rxFileName))
		doDeferRemoveFile = true
		errReturn = status.Error(codes.Internal, fmt.Sprintf("ошибка:<%v>, при вычислении хэша для файла:<%s>", err, rxFileName))
		return errReturn
	}

	s.ptrLogger.Info("От клиента принят файл", zap.String("имя", rxFileName), zap.String("хэш", fileHash))

	// Установка трейлера с хэшем файла
	mdTrailer := metadata.Pairs("hash", fileHash, "token", token[0])
	if err := grpc.SetTrailer(stream.Context(), mdTrailer); err != nil {
		s.ptrLogger.Error("ошибка установки трейлера", zap.String("файл", rxFileName), zap.String("ошибка", err.Error()))
		return status.Error(codes.Internal, "ошибка установки трейлера")
	}

	// Финальное сообщение сервера.
	return stream.SendAndClose(&pb.UploadResponse{
		FileName: rxFileName,
	})
}

// Обработчик передачи файлов.
func (s *PasswordManager) RestoreFile(req *pb.DownloadRequest, stream pb.PasswordManager_RestoreFileServer) error {

	// Извлечение метаданных из контекста
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.InvalidArgument, "ошибка извлечения заголовков")
	}

	// Получение значения токена
	token := md.Get("token")
	if len(token) == 0 {
		return status.Error(codes.Unauthenticated, "токен не передан")
	}

	//
	// Логика
	//

	reqFileName := req.FileName
	fileContent, err := os.ReadFile(reqFileName)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Запрошенный файл <%s>, отсутствует", req.FileName))
	}

	fileHash, err := hashFile(reqFileName)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Ошибка вычисления хэша:<%v>, для файла:<%s> отсутствует", err, req.FileName))
	}

	// Передача файла через stream.
	size := 1024

	for i := 0; i < len(fileContent); i += size {
		end := i + size
		if end > len(fileContent) {
			end = len(fileContent)
		}

		if err := stream.Send(&pb.DownloadResponse{
			FileName: reqFileName,
			Content:  fileContent[i:end],
		}); err != nil {
			return err
		}
	}

	// Добавление метаданных Trailer
	mdTrailer := metadata.Pairs("hash", fileHash, "token", token[0])
	if err := grpc.SetTrailer(stream.Context(), mdTrailer); err != nil {
		s.ptrLogger.Info("ошибка установки трейлера, для файла", zap.String("имя", reqFileName))
		return status.Error(codes.Aborted, fmt.Sprintf("ошибка установки трейлера, для файла:<%s>", reqFileName))
	}

	s.ptrLogger.Info("Клиенту отправлен файл", zap.String("имя", reqFileName), zap.String("хэш", fileHash))

	return nil
}

// Предоставление информации о файлах.
func (s *PasswordManager) FilesInfo(ctx context.Context, req *emptypb.Empty) (*proto.FilesInfoResponse, error) {

	// Извлечение метаданных из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "ошибка извлечения заголовков")
	}

	// Получение значения токена
	token := md.Get("token")
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, "токен не передан")
	}

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

	// Установка трейлера
	mdTrailer := metadata.Pairs("token", token[0])
	if err := grpc.SetTrailer(ctx, mdTrailer); err != nil {
		s.ptrLogger.Error("ошибка установки трейлера", zap.String("ошибка", err.Error()))
		return nil, status.Error(codes.Internal, "ошибка установки трейлера")
	}

	// Результат.
	return fileInfos, nil
}
