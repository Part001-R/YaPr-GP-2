package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	stageNotActive = 0
	stageActive    = 1
)

// Статусы процессов.
type statusSrv struct {
	isBackUp  int32 // Признак Активности процесса BackUp (клиент передаёт данные).
	isRestore int32 // Признак Активности процесса Restore (клиент принимает данные).
}

// grpc
type Manager struct {
	pb.UnimplementedPasswordManagerServer
	logger    *zap.Logger     // Логгер.
	status    statusSrv       // Статусы.
	storage   domain.StorageI // База данных.
	token     string          // Выданный токен
	secretKey string          // Секретный ключ
}

var once sync.Once // единоразовая инициализация экземпляра
var inst *Manager  // экземпляр

// Конструктор.
func New(l *zap.Logger, s domain.StorageI) *Manager {
	once.Do(func() {
		inst = &Manager{
			UnimplementedPasswordManagerServer: pb.UnimplementedPasswordManagerServer{},
			logger:                             l,
			status: statusSrv{
				isBackUp:  0,
				isRestore: 0,
			},
			storage:   s,
			token:     "",
			secretKey: "",
		}
	})
	return inst
}

//
// Обработчики.
//

// Обработчик проверки связи.
func (s *Manager) Ping(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {

	s.logger.Debug("Принят Ping запрос")

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logger.Warn("В принятом Ping запросе, отсутствуют метаданные")
		return nil, status.Error(codes.NotFound, "отсутствуют метаданные")
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 || tokens[0] == "" {
		s.logger.Warn("В метаданных, принятого Ping запросе, нет токена ")
		return nil, status.Error(codes.InvalidArgument, "в метаданных нет токена")
	}

	// Создание метаданных для ответа
	txMD := metadata.Pairs(nameToken, tokens[0]) // возврат принятого токена

	if err := grpc.SendHeader(ctx, txMD); err != nil {
		s.logger.Warn("В принятом Ping запросе, ошибка при отправке заголовков ")
		return nil, status.Error(codes.Internal, "ошибка при отправке заголовков")
	}

	return &emptypb.Empty{}, nil
}

// Обработчик приёма файла.
func (s *Manager) BackupFile(stream pb.PasswordManager_BackupFileServer) (errReturn error) {

	// Проверка, что процесс уже активный.
	if s.GetStatusBackUp() == stageActive || s.GetStatusRestore() == stageActive {
		return status.Error(codes.PermissionDenied, "Есть активный процесс")
	}

	// Установка статуса.
	if err := s.UpdateStatusBackUp(stageActive); err != nil {
		return status.Error(codes.Internal, "Ошибка обновления статуса")
	}

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
				s.logger.Error("ошибка при удалении файла",
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
			s.logger.Error("ошибка stream.Recv", zap.String("ошибка", err.Error()))
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
				s.logger.Error("ошибка при открытии файла", zap.String("файл", rxFileName), zap.String("ошибка", err.Error()))
				errReturn = status.Error(codes.Internal, fmt.Sprintf("ошибка:<%v>, открытия файла:<%s>", err, rxFileName))
				return errReturn
			}
			defer func() {
				if err := file.Close(); err != nil {
					s.logger.Error("ошибка закрытия подключения к файлу", zap.String("файл", rxFileName), zap.String("ошибка", err.Error()))
				}
			}()
		}

		// Сохранение принятых данных потока, в файл.
		_, err = file.Write(req.GetContent())
		if err != nil {
			s.logger.Error("ошибка при добавлении данных файла", zap.String("файл", rxFileName), zap.String("ошибка", err.Error()))
			doDeferRemoveFile = true
			errReturn = status.Error(codes.Internal, fmt.Sprintf("ошибка:<%v>, добавления данных в файл:<%s>", err, rxFileName))
			return errReturn
		}
	}

	// Вычисление хэша файла.
	fileHash, err := hashFile(rxFileName)
	if err != nil {
		s.logger.Error("ошибка при вычислении эеша для файла", zap.String("файл", rxFileName))
		doDeferRemoveFile = true
		errReturn = status.Error(codes.Internal, fmt.Sprintf("ошибка:<%v>, при вычислении хэша для файла:<%s>", err, rxFileName))
		return errReturn
	}

	s.logger.Info("От клиента принят файл", zap.String("имя", rxFileName), zap.String("хэш", fileHash))

	// Установка трейлера с хэшем файла
	mdTrailer := metadata.Pairs("hash", fileHash, "token", token[0])
	if err := grpc.SetTrailer(stream.Context(), mdTrailer); err != nil {
		s.logger.Error("ошибка установки трейлера", zap.String("файл", rxFileName), zap.String("ошибка", err.Error()))
		return status.Error(codes.Internal, "ошибка установки трейлера")
	}

	// Сброс статуса.
	if err := s.UpdateStatusBackUp(stageNotActive); err != nil {
		return status.Error(codes.Internal, "Ошибка обновления статуса")
	}

	// Финальное сообщение сервера.
	return stream.SendAndClose(&pb.UploadResponse{
		FileName: rxFileName,
	})
}

// Обработчик передачи файлов.
func (s *Manager) RestoreFile(req *pb.DownloadRequest, stream pb.PasswordManager_RestoreFileServer) error {

	// Получение статуса isBackUp
	if s.GetStatusBackUp() == stageActive {
		return status.Error(codes.PermissionDenied, "Идёт процесс BackUp")
	}

	if err := s.UpdateStatusRestore(stageActive); err != nil {
		return status.Error(codes.PermissionDenied, "ошибка обновления статуса")
	}

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
		s.logger.Info("ошибка установки трейлера, для файла", zap.String("имя", reqFileName))
		return status.Error(codes.Aborted, fmt.Sprintf("ошибка установки трейлера, для файла:<%s>", reqFileName))
	}

	s.logger.Info("Клиенту отправлен файл", zap.String("имя", reqFileName), zap.String("хэш", fileHash))

	// Сброс статуса.
	if err := s.UpdateStatusRestore(stageNotActive); err != nil {
		return status.Error(codes.Internal, "Ошибка обновления статуса")
	}

	return nil
}

// Предоставление информации о файлах.
func (s *Manager) FilesInfo(ctx context.Context, req *emptypb.Empty) (*proto.FilesInfoResponse, error) {

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

	s.logger.Debug("Принят FilesInfo запрос")

	// Подготовка.
	fileInfos := &pb.FilesInfoResponse{}
	files := []string{"manager.db", "container.data"}

	// Сбор информации по файлам.
	for _, f := range files {
		var el pb.FileInfo

		size, err := sizeFile(f)
		if err != nil {
			if errors.Is(err, ErrNotFound) { // Если файл не найден - обработка следующего файла.
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
		s.logger.Error("ошибка установки трейлера", zap.String("ошибка", err.Error()))
		return nil, status.Error(codes.Internal, "ошибка установки трейлера")
	}

	// Результат.
	return fileInfos, nil
}

// Регистрация.
func (s *Manager) Registration(ctx context.Context, req *pb.RegistrationRequest) (*emptypb.Empty, error) {

	s.logger.Info("Принят запрос регистрации пользователя")

	// Получение токена клиента.
	rxToken, err := layerRegistrationGetToken(ctx)
	if err != nil {
		s.logger.Error("Ошибка в слое layerRegistrationGetToken", zap.String("ошибка", err.Error()))
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// Обработка приёма.
	rxData, err := layerRegistrationRx(req)
	if err != nil {
		s.logger.Error("Ошибка в слое layerRegistrationRx", zap.String("ошибка", err.Error()))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Логика обработчика.
	if err := layerRegistrationLogic(rxData, s.storage); err != nil {
		s.logger.Error("Ошибка в слое layerRegistrationLogic", zap.String("ошибка", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Ответ.
	if err := layerRegistrationTx(ctx, rxToken); err != nil {
		s.logger.Error("Ошибка в слое layerRegistrationTx", zap.String("ошибка", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.logger.Info("Регистрация пользователя пройдена", zap.String("имя", rxData.userName))

	return nil, nil
}

// Аутентификация.
func (s *Manager) Authentication(ctx context.Context, req *pb.AuthenticationRequest) (*pb.AuthenticationResponse, error) {

	// Данные запроса.
	rxData, err := layerAuthenticationRx(req)
	if err != nil {
		s.logger.Error("Функция layerAuthenticationRx, вернуля ошибку", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "ошибка обработки принятых данных")
	}

	// Получение токена, переданного клиентом.
	rxToken, err := layerAuthenticationGetToken(ctx)
	if err != nil {
		s.logger.Error("Функция layerAuthenticationGetToken, вернуля ошибку", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "ошибка получения токена")
	}

	// Логика аутентификации.
	if err := layerAuthenticationLogic(s, rxData.userName, rxData.userPwd); err != nil {
		s.logger.Error("Функция layerAuthenticationLogic, вернуля ошибку", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "ошибка аутентификации")
	}

	// Создание токена.
	s.secretKey, s.token, err = createServerToken()
	if err != nil {
		s.logger.Error("Функция createServerToken, вернуля ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка создания токена")
	}

	// Формирование ответа. Возвращается принятый токен и передаётся токен аутентификации.
	res, err := layerAuthenticationTx(ctx, rxToken, s.token)
	if err != nil {
		s.logger.Error("Функция layerAuthenticationTx, вернуля ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка подготовки ответа")
	}

	return res, nil
}

//
// Интерцепторы.
//

// Unar интерцептор.
func (s *Manager) AuthInterceptorUnar(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	// metadata из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "в запросе, отсутствуют метаданные")
	}

	// Проверка присутствия токена
	tokens, exists := md["token"]
	if !exists || len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "нет данных токена")
	}

	// Токен есть, передача управления.
	return handler(ctx, req)
}

// Stream интерцептор.
func (s *Manager) AuthInterceptorStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	// metadata из контекста.
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "в запросе, отсутствуют метаданные")
	}

	// Проверка наличия токена.
	tokens, exists := md["token"]
	if !exists || len(tokens) == 0 {
		return status.Error(codes.Unauthenticated, "нет данных токена")
	}

	// Токен есть, передача управления.
	return handler(srv, ss)
}

//
// Статусы.
//

// Получение значения статуса isBackUp. Возвращается текущее значение статуса.
func (s *Manager) GetStatusBackUp() int32 {
	return atomic.LoadInt32(&s.status.isBackUp)
}

// Обновление значения статуса isBackUp. Возвращается ошибка.
func (s *Manager) UpdateStatusBackUp(stage int32) error {

	if s.status.isBackUp != stageActive && s.status.isBackUp != stageNotActive {
		return fmt.Errorf("Принятое значение статуса: <%d>, не поддерживается", stage)
	}

	atomic.StoreInt32(&s.status.isBackUp, stage)

	return nil
}

// Получение значения статуса isRestore. Возвращается текущее значение статуса.
func (s *Manager) GetStatusRestore() int32 {
	return atomic.LoadInt32(&s.status.isRestore)
}

// Обновление значения статуса isRestore. Возвращается ошибка.
func (s *Manager) UpdateStatusRestore(stage int32) error {

	if s.status.isRestore != stageActive && s.status.isRestore != stageNotActive {
		return fmt.Errorf("Принятое значение статуса: <%d>, не поддерживается", stage)
	}

	atomic.StoreInt32(&s.status.isRestore, stage)

	return nil
}
