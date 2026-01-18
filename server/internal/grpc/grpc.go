// Обработчики пакета.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"sync/atomic"

	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/server/internal/utils/flags"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Представленние сервиса.
type Manager struct {
	pb.UnimplementedPasswordManagerServer
	logger    *zap.Logger    // Логгер.
	status    statusSrv      // Статусы.
	storage   domain.DomainI // База данных.
	token     string         // Выданный токен
	secretKey string         // Секретный ключ
	flag      *flags.Config  // Флаги
}

var once sync.Once // единоразовая инициализация экземпляра
var inst *Manager  // экземпляр

// Конструктор. Возвращается указатель на экземпляр.
//
// Параметры:
//
// l - логгер.
// s - интерфейс домена.
// f - флаги.
func New(l *zap.Logger, s domain.DomainI, f *flags.Config) *Manager {
	once.Do(func() {
		inst = &Manager{
			UnimplementedPasswordManagerServer: pb.UnimplementedPasswordManagerServer{},
			logger:                             l,
			status: statusSrv{
				backUp:  0,
				restore: 0,
			},
			storage:   s,
			token:     "",
			secretKey: "",
			flag:      f,
		}
	})
	return inst
}

//
// Обработчики.
//

// Обработчик проверки связи. Возвращается пустой указатель и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	empty - указатель на пустые данные.
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

// Обработчик приёма файла. Возвращается ошибка.
//
// Параметры:
//
//	stream - поток.
func (s *Manager) LocalBackupFile(stream pb.PasswordManager_LocalBackupFileServer) (errReturn error) {

	s.logger.Info("Принят запрос BackUp")

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

				rxFileName = path.Join(s.flag.NameSubDirBackUp, rxFileName) // Добавление директории расположения

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

	s.logger.Info("BackUp, выполнен")

	// Финальное сообщение сервера.
	rxFileName = path.Base(rxFileName) // Выделение имени файла.

	return stream.SendAndClose(&pb.LocalBackupFileResponse{
		FileName: rxFileName,
	})
}

// Обработчик передачи файлов. Возвращается ошибка.
//
// Параметры:
//
//	req - данные запроса.
//	stream - поток.
func (s *Manager) LocalRestoreFile(req *pb.LocalRestoreFileRequest, stream pb.PasswordManager_LocalRestoreFileServer) error {

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
	reqFileName = path.Join(s.flag.NameSubDirBackUp, reqFileName) // Добавление названия директории

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

		if err := stream.Send(&pb.LocalRestoreFileResponse{
			FileName: path.Base(reqFileName),
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

// Предоставление информации о файлах. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
func (s *Manager) LocalFilesInfo(ctx context.Context, req *emptypb.Empty) (*proto.LocalFilesInfoResponse, error) {

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

	fileInfos := &pb.LocalFilesInfoResponse{}

	// Получение имён файлов в директории.
	files, err := ReadFilesInDirectory(s.flag.NameSubDirBackUp)
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка получения имён файлов backUp")
	}

	// Сбор информации по файлам.
	for _, f := range files {
		var el pb.FileInfo

		f = path.Join(s.flag.NameSubDirBackUp, f) // Добавление имени директории.

		size, err := sizeFile(f)
		if err != nil {
			if errors.Is(err, ErrNotFound) { // Если файл не найден - обработка следующего файла.
				continue
			}
			return nil, status.Error(codes.Internal, fmt.Sprintf("Ошибка:<%v>, при обработке файла:<%s>", err, f))
		}

		f = path.Base(f) // Выделение имени файла.

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

// Регистрация. Возвращается пустой указатель и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
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

// Аутентификация. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
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

// Добавление данных - логин/пароль. Возвращается пустой указатель и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
func (s *Manager) SendLoginPassword(ctx context.Context, req *pb.SendLoginPasswordRequest) (*emptypb.Empty, error) {

	if req.IdClient == "" {
		return nil, status.Error(codes.NotFound, "нет данных ID клиента")
	}
	s.logger.Info("Принят запрос добавления логин/пароль", zap.String("ID клиента", req.IdClient))

	// Получение данных запроса.
	rxData, err := layerSendLoginPasswordRx(req)
	if err != nil {
		s.logger.Error("ошибка получения отправленных данных", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка получения отправленных данных")
	}

	// Получение токена запроса.
	rxToken, err := layerSendLoginPasswordGetToken(ctx)
	if err != nil {
		s.logger.Error("ошибка получения токена запроса", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка получения токена запроса")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Логика.
	if err := layerSendLoginPasswordLogicContext(ctx, rxData, s); err != nil {
		s.logger.Error("ошибка добавления записи в БД", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка добавления записи в БД")
	}

	s.logger.Info("Данные запроса логин/пароль, успешно добавлены", zap.String("ID клиента", req.IdClient))

	return nil, nil
}

// Добавление данных - текст. Возвращается пустой указатель и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
func (s *Manager) SendText(ctx context.Context, req *pb.SendTextRequest) (*emptypb.Empty, error) {

	if req.IdClient == "" {
		return nil, status.Error(codes.NotFound, "нет данных ID клиента")
	}
	s.logger.Info("Принят запрос добавления текста", zap.String("ID клиента", req.IdClient))

	// Получение данных запроса.
	rxData, err := layerSendTextRx(req)
	if err != nil {
		s.logger.Error("ошибка получения отправленных данных", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка получения отправленных данных")
	}

	// Получение токена запроса.
	rxToken, err := layerSendTextGetToken(ctx)
	if err != nil {
		s.logger.Error("ошибка получения токена запроса", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка получения токена запроса")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Логика.
	if err := layerSendTextContext(ctx, rxData, s); err != nil {
		s.logger.Error("ошибка добавления записи в БД", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка добавления записи в БД")
	}

	s.logger.Info("Данные запроса текста, успешно добавлены", zap.String("ID клиента", req.IdClient))

	return nil, nil
}

// Добавление данных - банковская карта. Возвращается пустой указатель и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
func (s *Manager) SendBankCard(ctx context.Context, req *pb.SendBankCardRequest) (*emptypb.Empty, error) {

	if req.IdClient == "" {
		return nil, status.Error(codes.NotFound, "нет данных ID клиента")
	}
	s.logger.Info("Принят запрос добавления банковской карты", zap.String("ID клиента", req.IdClient))

	// Получение данных запроса.
	rxData, err := layerSendBankCardRx(req)
	if err != nil {
		s.logger.Error("ошибка получения отправленных данных", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка получения отправленных данных")
	}

	// Получение токена запроса.
	rxToken, err := layerSendBankCardGetToken(ctx)
	if err != nil {
		s.logger.Error("ошибка получения токена запроса", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка получения токена запроса")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Логика.
	if err := layerSendBankCardContext(ctx, rxData, s); err != nil {
		s.logger.Error("ошибка добавления записи в БД", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка добавления записи в БД")
	}

	s.logger.Info("Данные запроса банковской карты, успешно добавлены", zap.String("ID клиента", req.IdClient))

	return nil, nil
}

// Добавление данных - файл. Возвращается ошибка.
//
// Параметры:
//
//	stream - поток.
func (s *Manager) SendFile(stream pb.PasswordManager_SendFileServer) (errReturn error) {

	s.logger.Info("Принят запрос добавления файла")

	// Проверка уже запущенного процесса приёма файла.
	if s.GetStatusRx() != stageNotActive {
		return status.Error(codes.Unavailable, "Приём отклонён. Уже идёт передача файла.")
	}

	// Установка признака, что начат процесс приёма файла.
	if err := s.UpdateStatusRx(stageActive); err != nil {
		return status.Error(codes.Internal, "ошибка установки признака активности")
	}

	// Извлечение метаданных из контекста.
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.InvalidArgument, "ошибка извлечения заголовков")
	}

	// Получение значения токена.
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
	// Сброс признака активности.
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
		// Сброс статуса активности.
		if err := s.UpdateStatusRx(stageNotActive); err != nil {
			s.logger.Error("ошибка сброса признака активности", zap.String("ошибка", err.Error()))
			errReturn = fmt.Errorf("ошибка:<%w>, сброса признака активности. Базовая ошибка:<%w>", err, errReturn)
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
			rxFileName = path.Join(s.flag.NameSubDirFiles, rxFileName)

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

	s.logger.Info("Файл успешно принят")

	// Финальное сообщение сервера.
	return stream.SendAndClose(&pb.SendFileResponse{
		FileName: rxFileName,
	})
}

// Получение имён для логин/пароль. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	empty - пустой указатель.
func (s *Manager) RequestLoginPasswordName(ctx context.Context, empty *emptypb.Empty) (resp *pb.RequestLoginPasswordNameResponse, err error) {

	s.logger.Info("Принят запрос на получение имён записей логин/пароль")

	// Получение токена.
	rxToken, err := LayerRequestLoginPasswordNameToken(ctx)
	if err != nil {
		s.logger.Error("Функция LayerRequestLoginPasswordNameToken, вернула ошибку", zap.Error(err))
		return &pb.RequestLoginPasswordNameResponse{}, status.Error(codes.Internal, "Ошибка получения токена аутентификации")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Выполнение запроса к БД.
	rxData, err := LayerRequestLoginPasswordName(ctx, s)
	if err != nil {
		s.logger.Error("Функция LayerRequestLoginPasswordName, вернула ошибку", zap.Error(err))
		return &pb.RequestLoginPasswordNameResponse{}, status.Error(codes.Internal, "Ошибка при получении имён записей логин/пароль")
	}

	// Формирование ответа.
	resp, err = LayerRequestLoginPasswordNameTx(rxData)
	if err != nil {
		s.logger.Error("Функция LayerRequestLoginPasswordNameTx, вернула ошибку", zap.Error(err))
		return &pb.RequestLoginPasswordNameResponse{}, status.Error(codes.Internal, "Ошибка при формировании ответа")
	}

	s.logger.Info("Обработка запроса имён записей логин/пароль, выполнена")
	return resp, nil
}

// Получение данных логин/пароль по имени записи. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
func (s *Manager) RequestLoginPasswordByName(ctx context.Context, req *pb.RequestLoginPasswordByNameRequest) (resp *pb.RequestLoginPasswordByNameResponse, err error) {

	s.logger.Info("Принят запрос на получение данных записи логин/пароль, по имени записи")

	// Полуение токена аутентификации
	rxToken, err := layerRequestLoginPasswordByNameToken(ctx)
	if err != nil {
		s.logger.Error("Функция layerRequestLoginPasswordByNameToken, вернула ошибку", zap.Error(err))
		return &pb.RequestLoginPasswordByNameResponse{}, status.Error(codes.Internal, "Ошибка получения токена аутентификации")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Получение данных запроса.
	rxData, err := layerRequestLoginPasswordByName(req)
	if err != nil {
		s.logger.Error("Функция layerRequestLoginPasswordByName, вернула ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка обработки данных запроса")
	}

	// Запрос к БД.
	dataDB, err := s.storage.GetLoginPasswordByNameContext(ctx, rxData.Name)
	if err != nil {
		s.logger.Error("Функция GetLoginPasswordByNameContext, вернула ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка запроса к БД")
	}

	// Формирование ответа.
	var data TxLoginPassword
	data.For = dataDB.Name
	data.Login = dataDB.Login
	data.Password = dataDB.Password
	data.CreatedAt = dataDB.CreatedAt

	resp, err = layerRequestLoginPasswordByNameTx(data)
	if err != nil {
		s.logger.Error("Функция layerRequestLoginPasswordByNameTx, вернула ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка подготовки ответа")
	}

	// Результат.
	return resp, nil
}

// Получение имён для текста. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	empty - пустой указатель.
func (s *Manager) RequestTextName(ctx context.Context, empty *emptypb.Empty) (resp *pb.RequestTextNameResponse, err error) {

	s.logger.Info("Принят запрос на получение имён записей текста")

	// Получение токена.
	rxToken, err := LayerRequestTextNameToken(ctx)
	if err != nil {
		s.logger.Error("Функция LayerRequestTextNameToken, вернула ошибку", zap.Error(err))
		return &pb.RequestTextNameResponse{}, status.Error(codes.Internal, "Ошибка получения токена аутентификации")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Выполнение запроса к БД.
	rxData, err := LayerRequestTextName(ctx, s)
	if err != nil {
		s.logger.Error("Функция LayerRequestTextName, вернула ошибку", zap.Error(err))
		return &pb.RequestTextNameResponse{}, status.Error(codes.Internal, "Ошибка при получении имён записей текста")
	}

	// Формирование ответа.
	resp, err = LayerRequestTextNameTx(rxData)
	if err != nil {
		s.logger.Error("Функция LayerRequestTextNameTx, вернула ошибку", zap.Error(err))
		return &pb.RequestTextNameResponse{}, status.Error(codes.Internal, "Ошибка при формировании ответа")
	}

	s.logger.Info("Обработка запроса имён записей текста, выполнена")
	return resp, nil
}

// Получение данных текста по имени записи. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
func (s *Manager) RequestTextByName(ctx context.Context, req *pb.RequestTextByNameRequest) (resp *pb.RequestTextByNameResponse, err error) {

	s.logger.Info("Принят запрос на получение данных записи текста, по имени записи")

	// Полуение токена аутентификации
	rxToken, err := layerRequestTextByNameToken(ctx)
	if err != nil {
		s.logger.Error("Функция layerRequestTextByNameToken, вернула ошибку", zap.Error(err))
		return &pb.RequestTextByNameResponse{}, status.Error(codes.Internal, "Ошибка получения токена аутентификации")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Получение данных запроса.
	rxData, err := layerRequestTextByName(req)
	if err != nil {
		s.logger.Error("Функция layerRequestTextByName, вернула ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка обработки данных запроса")
	}

	// Запрос к БД.
	dataDB, err := s.storage.GetTextByNameContext(ctx, rxData.Name)
	if err != nil {
		s.logger.Error("Функция GetTextByNameContext, вернула ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка запроса к БД")
	}

	// Формирование ответа.
	var data TxText
	data.For = dataDB.Name
	data.Text = dataDB.Text
	data.CreatedAt = dataDB.CreatedAt

	resp, err = layerRequestTextByNameTx(data)
	if err != nil {
		s.logger.Error("Функция layerRequestTextByNameTx, вернула ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка подготовки ответа")
	}

	// Результат.
	return resp, nil
}

// Получение имён для банковских карт. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	empty - пустой указатель.
func (s *Manager) RequestBankCardName(ctx context.Context, empty *emptypb.Empty) (resp *pb.RequestBankCardNameResponse, err error) {

	s.logger.Info("Принят запрос на получение имён записей банковских карт")

	// Получение токена.
	rxToken, err := LayerRequestBankCardNameToken(ctx)
	if err != nil {
		s.logger.Error("Функция LayerRequestBankCardNameToken, вернула ошибку", zap.Error(err))
		return &pb.RequestBankCardNameResponse{}, status.Error(codes.Internal, "Ошибка получения токена аутентификации")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Выполнение запроса к БД.
	rxData, err := LayerRequestBankCardName(ctx, s)
	if err != nil {
		s.logger.Error("Функция LayerRequestBankCardName, вернула ошибку", zap.Error(err))
		return &pb.RequestBankCardNameResponse{}, status.Error(codes.Internal, "Ошибка при получении имён записей текста")
	}

	// Формирование ответа.
	resp, err = LayerRequestBankCardNameTx(rxData)
	if err != nil {
		s.logger.Error("Функция LayerRequestBankCardNameTx, вернула ошибку", zap.Error(err))
		return &pb.RequestBankCardNameResponse{}, status.Error(codes.Internal, "Ошибка при формировании ответа")
	}

	s.logger.Info("Обработка запроса имён записей банковских карт, выполнена")
	return resp, nil
}

// Получение данных банковской карты по имени записи. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
func (s *Manager) RequestBankCardByName(ctx context.Context, req *pb.RequestBankCardByNameRequest) (resp *pb.RequestBankCardByNameResponse, err error) {

	s.logger.Info("Принят запрос на получение данных записи банковской карты, по имени записи")

	// Полуение токена аутентификации
	rxToken, err := layerRequestBankCardByNameToken(ctx)
	if err != nil {
		s.logger.Error("Функция layerRequestBankCardByNameToken, вернула ошибку", zap.Error(err))
		return &pb.RequestBankCardByNameResponse{}, status.Error(codes.Internal, "Ошибка получения токена аутентификации")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Получение данных запроса.
	rxData, err := layerRequestBankCardByName(req)
	if err != nil {
		s.logger.Error("Функция layerRequestBankCardByName, вернула ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка обработки данных запроса")
	}

	// Запрос к БД.
	dataDB, err := s.storage.GetBankCardByNameContext(ctx, rxData.Name)
	if err != nil {
		s.logger.Error("Функция GetBankCardByNameContext, вернула ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка запроса к БД")
	}

	// Формирование ответа.
	var data TxBankCard
	data.For = dataDB.Name
	data.Owner = dataDB.Owner
	data.Numb = dataDB.Numb
	data.Valid = dataDB.Valid
	data.Code = dataDB.Code
	data.CreatedAt = dataDB.CreatedAt

	resp, err = layerRequestBankCardByNameTx(data)
	if err != nil {
		s.logger.Error("Функция layerRequestBankCardByNameTx, вернула ошибку", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка подготовки ответа")
	}

	// Результат.
	s.logger.Info("Обработка запроса банковской карты, выполнена")
	return resp, nil
}

// Получение имён файлов. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	empty - пустой указатель.
func (s *Manager) RequestFileName(ctx context.Context, empty *emptypb.Empty) (resp *pb.RequestFileNameResponse, err error) {

	s.logger.Info("Принят запрос имён файлов")

	//Получение имён файлов.
	rxData, isBusyServer, err := layerRequestFileNameScanDir(flags.NameSubDirFiles, s)
	if err != nil {
		s.logger.Error("Ошибка чтения имён файлов", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка чтения имён файлов")
	}

	// Формирование ответа
	resp, err = layerRequestFileNameTx(rxData, isBusyServer)
	if err != nil {
		s.logger.Error("Ошибка подготовки ответа", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка чтения имён файлов")
	}

	// Результат
	s.logger.Info("Запрос имён файлов, обработан")
	return resp, nil
}

// Обработчик передачи файлов. Возвращается ошибка.
//
// Параметры:
//
//	req - данные запроса.
//	stream - поток.
func (s *Manager) RequestFileByName(req *pb.RequestFileByNameRequest, stream pb.PasswordManager_RequestFileByNameServer) error {

	s.logger.Info("Принят запрос на передачу файла.")

	// Проверка активности процесса приёма файла
	if s.GetStatusRx() == stageActive {
		return status.Error(codes.PermissionDenied, "Идёт процесс приёма файла")
	}

	// Установка признака активности процесса передачи файла.
	if err := s.UpdateStatusTx(stageActive); err != nil {
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
	reqFilePath := path.Join(s.flag.NameSubDirFiles, reqFileName)

	fileContent, err := os.ReadFile(reqFilePath)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Запрошенный файл <%s>, отсутствует", req.FileName))
	}

	fileHash, err := hashFile(reqFilePath)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Ошибка вычисления хэша:<%v>, для файла:<%s> отсутствует", err, reqFilePath))
	}

	// Передача файла через stream.
	size := 1024

	for i := 0; i < len(fileContent); i += size {
		end := i + size
		if end > len(fileContent) {
			end = len(fileContent)
		}

		if err := stream.Send(&pb.RequestFileByNameResponse{
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
	if err := s.UpdateStatusTx(stageNotActive); err != nil {
		return status.Error(codes.Internal, "Ошибка обновления статуса")
	}

	s.logger.Info("Файл успешно отправлен")

	return nil
}

// Обработчик информации по файлу. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - данные запроса.
func (s *Manager) RequestFileInfo(ctx context.Context, req *pb.RequestFileInfoRequest) (resp *pb.RequestFileInfoResponse, err error) {

	s.logger.Info("Принят запрос на информацию по файлу.")

	// Получение токена аутентификации.
	rxToken, err := layerRequestFileInfoToken(ctx)
	if err != nil {
		s.logger.Error("ошибка получения токена аутентификации", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "ошибка получения токена")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Получение данных запроса.
	_, fileName, err := layerRequestFileInfoRx(req)
	if err != nil {
		s.logger.Error("ошибка в данных запроса", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "ошибка в данных запроса")
	}

	// Логика обработчика.
	filePath := path.Join(s.flag.NameSubDirFiles, fileName)
	dataFile, err := layerRequestFileInfo(filePath)
	if err != nil {
		s.logger.Error("ошибка сбора информации", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка при сборе информации")
	}

	// Формирование ответа.
	resp, err = layerRequestFileInfoTx(dataFile)
	if err != nil {
		s.logger.Error("ошибка формирования ответа", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка формирования ответа")
	}

	s.logger.Info("Запрос данных файла, обработан успешно.")

	return resp, nil
}

// Обработчик удаления записи логин/пароль. Возвращается пустой ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - запрос.
func (s *Manager) DeleteLoginPassword(ctx context.Context, req *pb.RequestDeleteName) (*emptypb.Empty, error) {

	s.logger.Info("Принят запрос на удаление записи логин/пароль.")

	// Получение токена аутентификации.
	rxToken, err := layerDeleteLoginPasswordToken(ctx)
	if err != nil {
		s.logger.Error("ошибка получения токена аутентификации", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "ошибка получения токена")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Получение данных запроса.
	_, name, err := layerDeleteLoginPasswordRx(req)
	if err != nil {
		s.logger.Error("ошибка получения данных запроса", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "ошибка в данных запроса")
	}

	// Удаление.
	if err := layerDeleteLoginPassword(name, s, ctx); err != nil {
		s.logger.Error("ошибка удаления записи логин/пароль", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка удаления записи логин/пароль")
	}

	s.logger.Info("Запись логин/пароль, удалена")
	return nil, nil
}

// Обработчик удаления записи текста. Возвращается пустой ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - запрос.
func (s *Manager) DeleteText(ctx context.Context, req *pb.RequestDeleteName) (*emptypb.Empty, error) {

	s.logger.Info("Принят запрос на удаление записи текста.")

	// Получение токена аутентификации.
	rxToken, err := layerDeleteTextToken(ctx)
	if err != nil {
		s.logger.Error("ошибка получения токена аутентификации", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "ошибка получения токена")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Получение данных запроса.
	_, name, err := layerDeleteTextRx(req)
	if err != nil {
		s.logger.Error("ошибка получения данных запроса", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "ошибка в данных запроса")
	}

	// Удаление.
	if err := layerDeleteText(name, s, ctx); err != nil {
		s.logger.Error("ошибка удаления записи текста", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка удаления записи текста")
	}

	s.logger.Info("Запись текста, удалена")
	return nil, nil
}

// Обработчик удаления записи банковской карты. Возвращается пустой ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - запрос.
func (s *Manager) DeleteBankCard(ctx context.Context, req *pb.RequestDeleteName) (*emptypb.Empty, error) {

	s.logger.Info("Принят запрос на удаление записи банковской карты.")

	// Получение токена аутентификации.
	rxToken, err := layerDeleteBankCardToken(ctx)
	if err != nil {
		s.logger.Error("ошибка получения токена аутентификации", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "ошибка получения токена")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Получение данных запроса.
	_, name, err := layerDeleteBankCardRx(req)
	if err != nil {
		s.logger.Error("ошибка получения данных запроса", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "ошибка в данных запроса")
	}

	// Удаление.
	if err := layerDeleteBankCard(name, s, ctx); err != nil {
		s.logger.Error("ошибка удаления записи банковской карты", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка удаления записи банковской карты")
	}

	s.logger.Info("Запись банковской карты, удалена")
	return nil, nil
}

// Обработчик удаления файла. Возвращается пустой ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - запрос.
func (s *Manager) DeleteFile(ctx context.Context, req *pb.RequestDeleteName) (*emptypb.Empty, error) {

	s.logger.Info("Принят запрос на удаление файла.")

	// Получение токена аутентификации.
	rxToken, err := layerDeleteFileToken(ctx)
	if err != nil {
		s.logger.Error("ошибка получения токена аутентификации", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "ошибка получения токена")
	}

	// Проверка токена.
	if err := checkToken(rxToken.token, s.secretKey); err != nil {
		s.logger.Error("ошибка проверки токена", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "токен не прошел проверку")
	}

	// Получение данных запроса.
	_, name, err := layerDeleteFileRx(req)
	if err != nil {
		s.logger.Error("ошибка получения данных запроса", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "ошибка в данных запроса")
	}

	// Удаление.
	if err := layerDeleteFile(name, s); err != nil {
		s.logger.Error("ошибка удаления файла", zap.Error(err))
		return nil, status.Error(codes.Internal, "ошибка удаления файла")
	}

	s.logger.Info("Файл, удалён")
	return nil, nil
}

//
// Интерцепторы.
//

// Unar интерцептор. Возвращается интерфейс и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - запрос.
//	info - информация.
//	handler - обработчик.
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

// Stream интерцептор. Возвращается ошибка.
//
// Параметры:
//
//	srv - интерфейс сервера.
//	ss - поток сервера.
//	info - информация.
//	handler - обработчик.
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
	return atomic.LoadInt32(&s.status.backUp)
}

// Обновление значения статуса isBackUp. Возвращается ошибка.
//
// Параметры6:
//
//	stage - новое значение.
func (s *Manager) UpdateStatusBackUp(stage int32) error {

	if stage != stageActive && stage != stageNotActive {
		return fmt.Errorf("Принятое значение статуса: <%d>, не поддерживается", stage)
	}

	atomic.StoreInt32(&s.status.backUp, stage)

	return nil
}

// Получение значения статуса isRestore. Возвращается текущее значение статуса.
func (s *Manager) GetStatusRestore() int32 {
	return atomic.LoadInt32(&s.status.restore)
}

// Обновление значения статуса isRestore. Возвращается ошибка.
//
// Параметры6:
//
//	stage - новое значение.
func (s *Manager) UpdateStatusRestore(stage int32) error {

	if stage != stageActive && stage != stageNotActive {
		return fmt.Errorf("Принятое значение статуса: <%d>, не поддерживается", stage)
	}

	atomic.StoreInt32(&s.status.restore, stage)

	return nil
}

// Получение значения статуса приёма файла от клиента. Возвращается текущее значение статуса.
func (s *Manager) GetStatusRx() int32 {
	return atomic.LoadInt32(&s.status.rxFile)
}

// Обновление значения статуса приёма файла от клиента. Возвращается ошибка.
//
// Параметры6:
//
//	stage - новое значение.
func (s *Manager) UpdateStatusRx(stage int32) error {

	if stage != stageActive && stage != stageNotActive {
		return fmt.Errorf("Принятое значение статуса: <%d>, не поддерживается", stage)
	}

	atomic.StoreInt32(&s.status.rxFile, stage)

	return nil
}

// Получение значения статуса передачи файла клиенту. Возвращается текущее значение статуса.
func (s *Manager) GetStatusTx() int32 {
	return atomic.LoadInt32(&s.status.txFile)
}

// Обновление значения статуса передачи файла клиенту. Возвращается ошибка.
//
// Параметры6:
//
//	stage - новое значение.
func (s *Manager) UpdateStatusTx(stage int32) error {

	if stage != stageActive && stage != stageNotActive {
		return fmt.Errorf("Принятое значение статуса: <%d>, не поддерживается", stage)
	}

	atomic.StoreInt32(&s.status.txFile, stage)

	return nil
}
