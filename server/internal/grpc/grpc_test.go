// Тесты пакета.
package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Part001-R/YaPr-GP-2/internal/utils/logger"
	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/server/internal/utils/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//
// --- Authentication ---
//

func TestAuthentication(t *testing.T) {

	//
	// Подготовка сервера для тестов.
	//

	// подготовка конфигурации сервиса.
	serv, err := prepare()
	require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

	defer func() {
		err := removeDirIfExists(serv.Flag.NameSubDirFiles)
		assert.NoErrorf(t, err, "Ошибка удаления тестовой директории с файлами")

		err = removeDirIfExists(serv.Flag.NameSubDirBackUp)
		assert.NoErrorf(t, err, "Ошибка удаления тестовой директории с бэкап")

		nameDB, err := flags.GetNameDBFromDSN(serv.Flag.DSN)
		require.NoErrorf(t, err, "Ошибка быделения имени БД из DSN")

		err = os.Remove(nameDB)
		assert.NoErrorf(t, err, "Ошибка удаления тестовой БД")
	}()

	fmt.Printf("Подготовка сервиса пройдена\n")

	// Запуск gRPCS сервера.
	go func() {
		err := actions(serv)
		require.NoErrorf(t, err, "Ошибка запуска:<%v>", err)
	}()

	fmt.Printf("Ожидание запуска горутины сервера\n")
	time.Sleep(1 * time.Second)

	// Подключение к gRPCS серверу.
	client, conn, err := connectSrv(serv.Flag.Port)
	require.NoErrorf(t, err, "Ошибка подключения к серверу")

	defer func() {
		err := conn.Close()
		assert.NoErrorf(t, err, "ошибка закрытия подключения к серверу")
	}()

	userName := "Foo"
	userPwd := "Bar"
	userPwdRepeat := "Bar"
	tokenAuth := ""
	namesLoginPassword := []string{}

	//
	// Тест регистрации пользователя (нет userName).
	//

	t.Run("Регистрация пользователя, без userName", func(t *testing.T) {

		// Создание метаданных с токеном.
		txMD, _, _, err := createTokenForRegistration()
		require.NoErrorf(t, err, "ошибка создания токена для регистрации: <%v>", err)

		// Подготовка запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RegistrationRequest{
			UserName:      "",
			UserPwd:       userPwd,
			UserPwdRepeat: userPwdRepeat,
		}

		var header metadata.MD

		// Запрос регистрации пользователя.
		_, err = client.Registration(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = InvalidArgument desc = некорректные данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест регистрации пользователя (нет userPassword).
	//

	t.Run("Регистрация пользователя, без userPassword", func(t *testing.T) {

		// Создание метаданных с токеном.
		txMD, _, _, err := createTokenForRegistration()
		require.NoErrorf(t, err, "ошибка создания токена для регистрации: <%v>", err)

		// Подготовка запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RegistrationRequest{
			UserName:      userName,
			UserPwd:       "",
			UserPwdRepeat: userPwdRepeat,
		}

		var header metadata.MD

		// Запрос регистрации пользователя.
		_, err = client.Registration(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = InvalidArgument desc = некорректные данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест регистрации пользователя (нет UserPwdRepeat).
	//

	t.Run("Регистрация пользователя, без UserPwdRepeat", func(t *testing.T) {

		// Создание метаданных с токеном.
		txMD, _, _, err := createTokenForRegistration()
		require.NoErrorf(t, err, "ошибка создания токена для регистрации: <%v>", err)

		// Подготовка запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RegistrationRequest{
			UserName:      userName,
			UserPwd:       userPwd,
			UserPwdRepeat: "",
		}

		var header metadata.MD

		// Запрос регистрации пользователя.
		_, err = client.Registration(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = InvalidArgument desc = некорректные данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест регистрации пользователя (разные userPwd и UserPwdRepeat).
	//

	t.Run("Регистрация пользователя, без соответствия подтверждения пароля", func(t *testing.T) {

		// Создание метаданных с токеном.
		txMD, _, _, err := createTokenForRegistration()
		require.NoErrorf(t, err, "ошибка создания токена для регистрации: <%v>", err)

		// Подготовка запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RegistrationRequest{
			UserName:      userName,
			UserPwd:       userPwd,
			UserPwdRepeat: userPwdRepeat + "1",
		}

		var header metadata.MD

		// Запрос регистрации пользователя.
		_, err = client.Registration(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = InvalidArgument desc = некорректные данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест регистрации пользователя (успешная регистрация).
	//

	t.Run("Регистрация пользователя", func(t *testing.T) {

		// Создание метаданных с токеном.
		txMD, secretKey, nameToken, err := createTokenForRegistration()
		require.NoErrorf(t, err, "ошибка создания токена для регистрации: <%v>", err)

		// Подготовка запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RegistrationRequest{
			UserName:      userName,
			UserPwd:       userPwd,
			UserPwdRepeat: userPwdRepeat,
		}

		var header metadata.MD

		// Запрос регистрации пользователя.
		_, err = client.Registration(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "ошибка регистрации пользователя")

		// Получение токена из метаданных ответа.
		token := header[nameToken]
		require.Falsef(t, len(token) == 0 || token[0] == "", "Нет токена метаданных")

		rxToken := token[0]

		// Проверка токена.
		err = checkToken(rxToken, secretKey)
		assert.NoErrorf(t, err, "ошибка проверки токена ответа")
	})

	//
	// Тест аутентификации пользователя (без userName).
	//

	t.Run("Аутентификация пользователя, без userName", func(t *testing.T) {

		// Создание метаданных с токеном.
		txMD, _, _, err := createTokenForRegistration()
		require.NoErrorf(t, err, "ошибка создания токена для регистрации: <%v>", err)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.AuthenticationRequest{
			UserName: "",
			UserPwd:  userPwd,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.Authentication(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = InvalidArgument desc = некорректные данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест аутентификации пользователя (без userPwd).
	//

	t.Run("Аутентификация пользователя, без userPwd", func(t *testing.T) {

		// Создание метаданных с токеном.
		txMD, _, _, err := createTokenForRegistration()
		require.NoErrorf(t, err, "ошибка создания токена для регистрации: <%v>", err)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.AuthenticationRequest{
			UserName: userName,
			UserPwd:  "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.Authentication(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = InvalidArgument desc = некорректные данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест аутентификации пользователя.
	//

	t.Run("Аутентификация пользователя", func(t *testing.T) {

		// Создание метаданных с токеном.
		txMD, secretKey, nameToken, err := createTokenForRegistration()
		require.NoErrorf(t, err, "ошибка создания токена для регистрации: <%v>", err)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.AuthenticationRequest{
			UserName: userName,
			UserPwd:  userPwd,
		}

		var header metadata.MD

		// Выполнение запроса.
		resp, err := client.Authentication(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка аутентификации")

		// Получение токена из метаданных ответа.
		token := header[nameToken]
		require.Falsef(t, len(token) == 0 || token[0] == "", "Нет токена метаданных")

		rxToken := token[0]

		// Проверка токена.
		err = checkToken(rxToken, secretKey)
		assert.NoErrorf(t, err, "ошибка проверки токена ответа")

		// Сохранение токена аутентификации.
		tokenAuth = resp.Token
		_ = tokenAuth
	})

	//
	// Тест добавления записи логин/пароль (корректные данные).
	//

	t.Run("Добавление записи логин/пароль", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  "AAA",
			For:       "For1",
			Login:     "Foo",
			Password:  "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendLoginPassword(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка добавления данных логин/пароль")
	})

	//
	// Тест добавления записи логин/пароль (ошибочный токен аутентификации).
	//

	t.Run("Добавление записи логин/пароль с ошибочным токеном аутентификации", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, "incorrectTokenAuth")

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  "AAA",
			For:       "For2",
			Login:     "Foo",
			Password:  "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи логин/пароль (добавление уже существующей записи).
	//

	t.Run("Добавление записи логин/пароль с коллизией в данных", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  "AAA",
			For:       "For1",
			Login:     "Foo",
			Password:  "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка добавления записи в БД", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи логин/пароль, вторая запись (корректные данные).
	//

	t.Run("Добавление записи логин/пароль, вторая запись", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  "AAA",
			For:       "For2",
			Login:     "Foo",
			Password:  "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendLoginPassword(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка добавления данных логин/пароль")
	})

	//
	// Тест добавления записи логин/пароль (нет id клиента).
	//

	t.Run("Добавление записи логин/пароль, без id клиента", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  "",
			For:       "For3",
			Login:     "Foo",
			Password:  "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = NotFound desc = нет данных ID клиента", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи логин/пароль (нет принадлежности).
	//

	t.Run("Добавление записи логин/пароль, без For", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  "AAA",
			For:       "",
			Login:     "Foo",
			Password:  "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = отсутствуют данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи логин/пароль (нет логина).
	//

	t.Run("Добавление записи логин/пароль, без Login", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  "AAA",
			For:       "For3",
			Login:     "",
			Password:  "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = отсутствуют данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи логин/пароль (нет пароля).
	//

	t.Run("Добавление записи логин/пароль, без Password", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  "AAA",
			For:       "For3",
			Login:     "Foo",
			Password:  "",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = отсутствуют данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи логин/пароль (нет даты создания).
	//

	t.Run("Добавление записи логин/пароль, без CreatedAt", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  "AAA",
			For:       "For3",
			Login:     "Foo",
			Password:  "Bar",
			CreatedAt: "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = отсутствуют данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест получения имён записей логин/пароль (другой токен аутентификции).
	//

	t.Run("Получение имён записей логин/пароль, другой токен аутентификации", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth+"1")

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &emptypb.Empty{}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.RequestLoginPasswordName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест получения имён записей логин/пароль.
	//

	t.Run("Получение имён записей логин/пароль", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &emptypb.Empty{}

		var header metadata.MD

		// Выполнение запроса.
		resp, err := client.RequestLoginPasswordName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка запроса имён данных логин/пароль")

		// Сохранение принятых данных
		namesLoginPassword = resp.EntriesName
		_ = namesLoginPassword
	})

}

// =======================================================================================================
//
//                               Вспомогательные функции для тестов
//
// =======================================================================================================

// инициализация сервиса для теста. Возвращается указатель и ошибка.
func prepare() (*Configuration, error) {

	// Создание логгера.
	lgr, err := logger.New("debug")
	if err != nil {
		return nil, fmt.Errorf("функция logger.NewLogger, вернула ошибку: <%w>", err)
	}

	// Флаги.
	flag := flags.New()

	// Создание директории для файлов.
	if err := createSubdirectory(flag.NameSubDirFiles); err != nil {
		return nil, fmt.Errorf("Создание директории для файлов. функция createSubdirectory, вернула ошибку: <%w>", err)
	}
	// Создание директории для backUp.
	if err := createSubdirectory(flag.NameSubDirBackUp); err != nil {
		return nil, fmt.Errorf("Создание директории для backUp. функция createSubdirectory, вернула ошибку: <%w>", err)
	}

	// БД.
	storage, err := domain.NewStorage(flag.DSN)
	if err != nil {
		return nil, fmt.Errorf("функция domain.NewStorage, вернула ошибку: <%w>", err)
	}

	// Экземляр сервиса.
	srvGRPC := New(lgr, storage, flag)

	// Пути к TLS файлам.
	pathTLSsert := "tls/server.crt"
	pathTLSPriv := "tls/server.key"

	// Создание сводной конфигурации сервиса.
	serviceInst := NewServ(lgr, srvGRPC, pathTLSsert, pathTLSPriv, storage, flag)

	return serviceInst, nil
}

// Удаление директории. Возвращается ошибка.
//
// Параметры:
//
//	dirPath - путь к директории.
func removeDirIfExists(dirPath string) error {

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil
	}

	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("ошибка <%w>, при удалении директории <%s>", err, dirPath)
	}

	return nil
}

// Подключение к серверу. Возвращается клиент, подключение и ошибка.
//
// Параметры:
//
//	c - экземпляр интерфейса.
func connectSrv(port string) (pb.PasswordManagerClient, *grpc.ClientConn, error) {

	// Логика
	//
	srvAddr := ":" + port

	// Настройка TLS.
	creds, err := credentials.NewClientTLSFromFile("tls/server.crt", "")
	if err != nil {
		return nil, nil, fmt.Errorf("функция credentials.NewClientTLSFromFile, вернула ошибку: <%w>", err)
	}

	// Подключение к серверу.
	conn, err := grpc.NewClient(srvAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, nil, fmt.Errorf("функция grpc.NewClient, вернула ошибку: <%w>", err)
	}

	// создание клиента.
	client := pb.NewPasswordManagerClient(conn)

	return client, conn, nil
}

// Создание токена для регистрации пользователя в режиме  - удалённый. Возвращаеются метаданные, ключ, имя токена и ошибка.
func createTokenForRegistration() (txMD metadata.MD, secretKey, nameToken string, err error) {

	// Создание ключа.
	secretKey, err = generateRandomString(50)
	if err != nil {
		return nil, "", "", fmt.Errorf("функция generateRandomString, вернула ошибку: <%w>", err)
	}

	// Создание токена.
	timeValidToken := time.Duration(5 * time.Second)
	txToken, err := createToken("clientManager", secretKey, timeValidToken)
	if err != nil {
		return nil, "", "", fmt.Errorf("функция createToken, вернула ошибку: <%w>", err)
	}

	// Заполнение метаданных.
	nameToken = "token"
	txMD = metadata.Pairs(nameToken, txToken)

	return txMD, secretKey, nameToken, nil
}

// Функция содержит действия сервиса. Возвращается ошибка.
//
// Параметры:
//
// с - конфигурация сервиса.
func actions(c *Configuration) error {

	port := ":" + c.Flag.Port

	// Проверка аргументов
	if c == nil {
		return NilPtrArgumentConf
	}

	// Подключение к порту.
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("ошибка подключения к порту: <%w>", err)
	}

	// Создание конфигурации для TLS
	tlsConfig, err := createTLSConfig(c)
	if err != nil {
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
	fmt.Printf("Запуск сервера на порту <%s>\n", port)
	if err := s.Serve(lis); err != nil {
		fmt.Printf("Ошибка сервера <%v>\n", err)
		return fmt.Errorf("ошибка в работе gRPC сервера: <%w>", err)
	}

	return nil
}

// createTLSConfig создание конфигурацию TLS. Возвращается конфигурация и ошибка.
//
// Параметры:
//
// c - указатель на конфигурацию сервиса.
func createTLSConfig(c *Configuration) (*tls.Config, error) {

	cert, err := tls.LoadX509KeyPair(c.TLS.Public, c.TLS.Privae)
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
	}
	return nil
}

// Конструктор. Возвращается указатель на экземпляр сервиса.
//
// Параметры:
//
//	l - логгер.
//	g - указатель на grpc.
//	keyPublic - публичный ключ.
//	keyPrivate - приватный ключ.
//	storage - указатель на домен.
//	flag - указатель на флаги.
func NewServ(l *zap.Logger, g *Manager, keyPublic, keyPrivate string, storage domain.DomainI, flag *flags.Config) *Configuration {

	return &Configuration{
		Lgr: l,
		Srv: g,
		TLS: TLSdata{
			Public: keyPublic,
			Privae: keyPrivate,
		},
		Storage: storage,
		Flag:    flag,
	}

}

// Unar интерцептор. Возвращается интерфейс и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	req - запрос.
//	info - информация.
//	handler - обработчик.
func (s *TestManager) AuthInterceptorUnar(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

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
func (s *TestManager) AuthInterceptorStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

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
