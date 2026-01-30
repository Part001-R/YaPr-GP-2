// Тесты пакета.
package grpc

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path"
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

// Тест всех обработчиков gRPC.
func TestAllHandlersGRPC(t *testing.T) {

	//
	// Подготовка сервера для тестов.
	//

	// Подготовка конфигурации сервиса.
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

	// -------------------------------------------------------------------------------------------------
	//
	//                                               БД
	//
	// -------------------------------------------------------------------------------------------------
	// Данные для тестов.
	userName := "Foo"
	userPwd := "Bar"
	userPwdRepeat := "Bar"
	tokenAuth := ""
	namesLoginPassword := []string{}
	namesText := []string{}
	namesBankCard := []string{}

	// -------------------------------------------------------------------------------------------------
	//
	//                                        Регистрация
	//
	// -------------------------------------------------------------------------------------------------

	//
	// Тест регистрации пользователя (нет метаданных).
	//

	t.Run("Регистрация пользователя, без метаданных", func(t *testing.T) {

		req := &proto.RegistrationRequest{
			UserName:      "",
			UserPwd:       userPwd,
			UserPwdRepeat: userPwdRepeat,
		}

		var header metadata.MD

		_, err = client.Registration(context.Background(), req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Unauthenticated desc = нет данных токена", err.Error(), "Нет соответствия ошибки")
	})

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

	// -------------------------------------------------------------------------------------------------
	//
	//                                        Аутентификация
	//
	// -------------------------------------------------------------------------------------------------

	//
	// Тест аутентификации пользователя (нет метаданных).
	//

	t.Run("Регистрация пользователя, без метаданных", func(t *testing.T) {

		req := &proto.AuthenticationRequest{
			UserName: "Foo",
			UserPwd:  "Bar",
		}

		var header metadata.MD

		_, err = client.Authentication(context.Background(), req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Unauthenticated desc = нет данных токена", err.Error(), "Нет соответствия ошибки")
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

	// -------------------------------------------------------------------------------------------------
	//
	//                                         Логин/пароль
	//
	// -------------------------------------------------------------------------------------------------

	//
	// Тест добавления записи логин/пароль.
	//

	// Первая запись.
	txLoginPassword1 := RxLoginPassword{
		ID:        "AAA",
		For:       "For1",
		Login:     "Foo",
		Password:  "Bar",
		CreatedAt: time.Now().UTC().String(),
	}

	t.Run("Добавление записи логин/пароль", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  txLoginPassword1.ID,
			For:       txLoginPassword1.For,
			Login:     txLoginPassword1.Login,
			Password:  txLoginPassword1.Password,
			CreatedAt: txLoginPassword1.CreatedAt,
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

	// Вторая запись.
	txLoginPassword2 := RxLoginPassword{
		ID:        "AAA",
		For:       "For2",
		Login:     "Foo",
		Password:  "Bar",
		CreatedAt: time.Now().UTC().String(),
	}

	t.Run("Добавление записи логин/пароль, вторая запись", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendLoginPasswordRequest{
			IdClient:  txLoginPassword2.ID,
			For:       txLoginPassword2.For,
			Login:     txLoginPassword2.Login,
			Password:  txLoginPassword2.Password,
			CreatedAt: txLoginPassword2.CreatedAt,
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

	//
	// Тест получения данных логин/пароль, по имени записи. (подставной токен)
	//

	t.Run("Получение данных логин/пароль, подставной токен", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth+"1")

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestLoginPasswordByNameRequest{
			IdClient: "AAA",
			Name:     txLoginPassword1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.RequestLoginPasswordByName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест получения данных логин/пароль, по имени записи. (нет id клиента)
	//

	t.Run("Получение данных логин/пароль, без idClient", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestLoginPasswordByNameRequest{
			IdClient: "",
			Name:     txLoginPassword1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.RequestLoginPasswordByName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка обработки данных запроса", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест получения данных логин/пароль, по имени записи. (нет For)
	//

	t.Run("Получение данных логин/пароль, без idClient", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestLoginPasswordByNameRequest{
			IdClient: "AAA",
			Name:     "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.RequestLoginPasswordByName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка обработки данных запроса", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест получения данных логин/пароль, по имени записи. (первая запись)
	//

	rxLoginPassword1 := RxLoginPassword{}

	t.Run("Получение данных логин/пароль, по имени записи. 1", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestLoginPasswordByNameRequest{
			IdClient: "AAA",
			Name:     txLoginPassword1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		resp, err := client.RequestLoginPasswordByName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка запроса данных логин/пароль, по имени записи")

		rxLoginPassword1.For = resp.Name
		rxLoginPassword1.Login = resp.Login
		rxLoginPassword1.Password = resp.Password
	})

	//
	// Тест получения данных логин/пароль, по имени записи. (вторая запись)
	//

	rxLoginPassword2 := RxLoginPassword{}

	t.Run("Получение данных логин/пароль, по имени записи. 1", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestLoginPasswordByNameRequest{
			IdClient: "AAA",
			Name:     txLoginPassword2.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		resp, err := client.RequestLoginPasswordByName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка запроса данных логин/пароль, по имени записи")

		rxLoginPassword2.For = resp.Name
		rxLoginPassword2.Login = resp.Login
		rxLoginPassword2.Password = resp.Password
	})

	//
	// Проверка соответствия данных логин/пароль.
	//

	t.Run("Проверка соответствия данных логин/пароль", func(t *testing.T) {

		assert.Equalf(t, txLoginPassword1.For, rxLoginPassword1.For, "Запись-1. Нет соответствия For")
		assert.Equalf(t, txLoginPassword1.Login, rxLoginPassword1.Login, "Запись-1. Нет соответствия Login")
		assert.Equalf(t, txLoginPassword1.Password, rxLoginPassword1.Password, "Запись-1. Нет соответствия Password")

		assert.Equalf(t, txLoginPassword2.For, rxLoginPassword2.For, "Запись-2. Нет соответствия For")
		assert.Equalf(t, txLoginPassword2.Login, rxLoginPassword2.Login, "Запись-2. Нет соответствия Login")
		assert.Equalf(t, txLoginPassword2.Password, rxLoginPassword2.Password, "Запись-2. Нет соответствия Password")
	})

	//
	// Удаление данных логин/пароль. (подставной токен)
	//

	t.Run("Удаление данных логин/пароль. (подставной токен)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth+"1")

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     txLoginPassword1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "нет соответствия токена")
	})

	//
	// Удаление данных логин/пароль. (нет id клиента)
	//

	t.Run("Удаление данных логин/пароль. (нет id клиента)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "",
			Name:     txLoginPassword1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Unavailable desc = ошибка в данных запроса", err.Error(), "нет соответствия токена")
	})

	//
	// Удаление данных логин/пароль. (нет имени)
	//

	t.Run("Удаление данных логин/пароль. (нет имени)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Unavailable desc = ошибка в данных запроса", err.Error(), "нет соответствия токена")
	})

	//
	// Удаление данных логин/пароль. (первая запись)
	//

	t.Run("Удаление данных логин/пароль. 1", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     txLoginPassword1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteLoginPassword(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка удаления данных логин/пароль, по имени записи")
	})

	//
	// Удаление данных логин/пароль. (вторая запись)
	//

	t.Run("Удаление данных логин/пароль. 2", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     txLoginPassword2.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteLoginPassword(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка удаления данных логин/пароль, по имени записи")
	})

	//
	// Удаление отсутствующих данных логин/пароль.
	//

	t.Run("Удаление отсутствующих данных логин/пароль", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     "NotExistsName",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteLoginPassword(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка удаления записи логин/пароль", err.Error(), "Нет соответствия ошибки")
	})

	// -------------------------------------------------------------------------------------------------
	//
	//                                             Текст
	//
	// -------------------------------------------------------------------------------------------------

	//
	// Тест добавления записи текста (ошибочный токен аутентификации).
	//

	t.Run("Добавление записи текста, с ошибочным токеном аутентификации", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth+"1")

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendTextRequest{
			IdClient:  "AAA",
			For:       "For1",
			Text:      "Foo",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи текста (нет id клиента).
	//

	t.Run("Добавление записи текста, без id клиента", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendTextRequest{
			IdClient:  "",
			For:       "For1",
			Text:      "Foo",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = NotFound desc = нет данных ID клиента", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи текста (нет For).
	//

	t.Run("Добавление записи текста, без For", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendTextRequest{
			IdClient:  "AAA",
			For:       "",
			Text:      "Foo",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = отсутствуют данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи текста (нет Text).
	//

	t.Run("Добавление записи текста, без Text", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendTextRequest{
			IdClient:  "AAA",
			For:       "For1",
			Text:      "",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = отсутствуют данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи текста (нет CreateAt).
	//

	t.Run("Добавление записи текста, без CreateAt", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendTextRequest{
			IdClient:  "AAA",
			For:       "For1",
			Text:      "Text",
			CreatedAt: "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = отсутствуют данные", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи текста (корректные данные).
	//

	// Первая запись.
	txText1 := RxText{
		ID:        "AAA",
		For:       "For1",
		Text:      "Foo",
		CreatedAt: time.Now().UTC().String(),
	}

	t.Run("Добавление записи текста", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendTextRequest{
			IdClient:  txText1.ID,
			For:       txText1.For,
			Text:      txText1.Text,
			CreatedAt: txText1.CreatedAt,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendText(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка добавления данных текста")
	})

	//
	// Тест добавления записи текста (проверка коллизии).
	//

	t.Run("Добавление записи текста, коллизия", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendTextRequest{
			IdClient:  "AAA",
			For:       "For1",
			Text:      "Foo",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка добавления записи в БД", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи текста (корректные данные, второй записи).
	//

	// Первая запись.
	txText2 := RxText{
		ID:        "AAA",
		For:       "For2",
		Text:      "Foo",
		CreatedAt: time.Now().UTC().String(),
	}

	t.Run("Добавление записи текста, вторая запись", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendTextRequest{
			IdClient:  txText2.ID,
			For:       txText2.For,
			Text:      txText2.Text,
			CreatedAt: txText2.CreatedAt,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendText(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка добавления данных текста")
	})

	//
	// Тест получения имён записей текста (подставной токен аутентификации).
	//

	t.Run("Получение имён записей текста, подставной токен аутентификации", func(t *testing.T) {

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
		_, err := client.RequestTextName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест получения имён записей текста.
	//

	t.Run("Получение имён записей текста", func(t *testing.T) {

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
		resp, err := client.RequestTextName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка запроса имён данных текста")

		// Сохранение принятых данных
		namesText = resp.EntriesName
		_ = namesText
	})

	//
	// Тест получения записи текста, по имени (корректные данные, первой записи).
	//

	// Первая запись.
	rxText1 := RxText{}

	t.Run("Получение первой записи текста по имени", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestTextByNameRequest{
			IdClient: "AAA",
			Name:     txText1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		resp, err := client.RequestTextByName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка запроса данных текста, по имени записи")

		rxText1.For = resp.Name
		rxText1.Text = resp.Text
		rxText1.CreatedAt = resp.CreatedAt
	})

	//
	// Тест получения записи текста, по имени (корректные данные, второй записи).
	//

	// Первая запись.
	rxText2 := RxText{}

	t.Run("Получение первой записи текста по имени", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestTextByNameRequest{
			IdClient: "AAA",
			Name:     txText2.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		resp, err := client.RequestTextByName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка запроса данных текста, по имени записи")

		rxText2.For = resp.Name
		rxText2.Text = resp.Text
		rxText2.CreatedAt = resp.CreatedAt
	})

	//
	// Проверка соответствия данных Tx и Rx.
	//

	t.Run("Проверка соответствия данных Tx и Rx у данных текста", func(t *testing.T) {

		assert.Equalf(t, txText1.For, rxText1.For, "Нет соответствия For, у первой записи текста")
		assert.Equalf(t, txText1.Text, rxText1.Text, "Нет соответствия Text, у первой записи текста")

		assert.Equalf(t, txText2.For, rxText2.For, "Нет соответствия For, у второй записи текста")
		assert.Equalf(t, txText2.Text, rxText2.Text, "Нет соответствия Text, у второй записи текста")
	})

	//
	// Удаление данных текста. (подставной токен)
	//

	t.Run("Удаление данных текста. (подставной токен)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth+"1")

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     txText1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "нет соответствия токена")
	})

	//
	// Удаление данных текста. (нет id клиента)
	//

	t.Run("Удаление данных текста. (нет id клиента)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "",
			Name:     txText1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Unavailable desc = ошибка в данных запроса", err.Error(), "нет соответствия токена")
	})

	//
	// Удаление данных текста. (нет имени записи)
	//

	t.Run("Удаление данных текста. (нет имени записи)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Unavailable desc = ошибка в данных запроса", err.Error(), "нет соответствия токена")
	})

	//
	// Удаление данных текста. (удаление первой записи)
	//

	t.Run("Удаление данных текста. (удаление первой записи)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     txText1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteText(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка удаления записи")
	})

	//
	// Удаление данных текста. (удаление второй записи)
	//

	t.Run("Удаление данных текста. (удаление второй записи)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     txText2.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteText(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка удаления записи")
	})

	//
	// Удаление данных текста. (удаление отсутствующей записи)
	//

	t.Run("Удаление данных текста. (удаление отсутствующей записи)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     "UnavailableName",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteText(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка удаления записи текста", err.Error(), "Нет соответствия ошибки")
	})

	// -------------------------------------------------------------------------------------------------
	//
	//                                        Банковская карта
	//
	// -------------------------------------------------------------------------------------------------

	// Тест добавления записи банковской карты (корректные данные).
	//

	// Первая запись
	txBankCard1 := RxBankCard{
		ID:        "AAA",
		For:       "For1",
		Owner:     "Foo",
		Numb:      "4532015112830366",
		ValidData: "Foo",
		Code:      "Bar",
		CreatedAt: time.Now().UTC().String(),
	}

	t.Run("Добавление записи банковской карты", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  txBankCard1.ID,
			For:       txBankCard1.For,
			Owner:     txBankCard1.Owner,
			Numb:      txBankCard1.Numb,
			ValidData: txBankCard1.ValidData,
			Code:      txBankCard1.Code,
			CreatedAt: txBankCard1.CreatedAt,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка добавления данных банковской карты")
	})

	//
	// Тест добавления записи банковской карты (подставной токен).
	//

	t.Run("Добавление записи банковской карты, подставной токен", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth+"1")

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  "AAA",
			For:       "For2",
			Owner:     "Foo",
			Numb:      "4532015112830366",
			ValidData: "Foo",
			Code:      "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи банковской карты (нет id клиента).
	//

	t.Run("Добавление записи банковской карты, без id клиента", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  "",
			For:       "For2",
			Owner:     "Foo",
			Numb:      "4532015112830366",
			ValidData: "Foo",
			Code:      "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = NotFound desc = нет данных ID клиента", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи банковской карты (нет For).
	//

	t.Run("Добавление записи банковской карты, без For", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  "AAA",
			For:       "",
			Owner:     "Foo",
			Numb:      "4532015112830366",
			ValidData: "Foo",
			Code:      "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка получения отправленных данных", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи банковской карты (нет Owner).
	//

	t.Run("Добавление записи банковской карты, без Owner", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  "AAA",
			For:       "For2",
			Owner:     "",
			Numb:      "4532015112830366",
			ValidData: "Foo",
			Code:      "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка получения отправленных данных", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи банковской карты (нет Numb).
	//

	t.Run("Добавление записи банковской карты, без Numb", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  "AAA",
			For:       "For2",
			Owner:     "Foo",
			Numb:      "",
			ValidData: "Foo",
			Code:      "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка получения отправленных данных", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи банковской карты (нет ValidData).
	//

	t.Run("Добавление записи банковской карты, без ValidData", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  "AAA",
			For:       "For2",
			Owner:     "Foo",
			Numb:      "4532015112830366",
			ValidData: "",
			Code:      "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка получения отправленных данных", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи банковской карты (нет Code).
	//

	t.Run("Добавление записи банковской карты, без Code", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  "AAA",
			For:       "For2",
			Owner:     "Foo",
			Numb:      "4532015112830366",
			ValidData: "Foo",
			Code:      "",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка получения отправленных данных", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи банковской карты (нет CreatedAt).
	//

	t.Run("Добавление записи банковской карты, без CreatedAt", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  "AAA",
			For:       "For2",
			Owner:     "Foo",
			Numb:      "4532015112830366",
			ValidData: "Foo",
			Code:      "Bar",
			CreatedAt: "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка получения отправленных данных", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи банковской карты (ошибка в номере карты).
	//

	t.Run("Добавление записи банковской карты, ошибка в номере карты", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  "AAA",
			For:       "For3",
			Owner:     "Foo",
			Numb:      "4532015112830367",
			ValidData: "Foo",
			Code:      "Bar",
			CreatedAt: time.Now().UTC().String(),
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка получения отправленных данных", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест добавления записи банковской карты, второй (корректные данные).
	//

	// Первая запись
	txBankCard2 := RxBankCard{
		ID:        "AAA",
		For:       "For2",
		Owner:     "Foo",
		Numb:      "4532015112830366",
		ValidData: "Foo",
		Code:      "Bar",
		CreatedAt: time.Now().UTC().String(),
	}

	t.Run("Добавление записи банковской карты, второй", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.SendBankCardRequest{
			IdClient:  txBankCard2.ID,
			For:       txBankCard2.For,
			Owner:     txBankCard2.Owner,
			Numb:      txBankCard2.Numb,
			ValidData: txBankCard2.ValidData,
			Code:      txBankCard2.Code,
			CreatedAt: txBankCard2.CreatedAt,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.SendBankCard(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка добавления данных банковской карты")
	})

	//
	// Тест получения имён записей банковских карт (подставной токен).
	//

	t.Run("Получение имён записей банковских карт, с подставным токеном", func(t *testing.T) {

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
		_, err := client.RequestBankCardName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")

	})

	//
	// Тест получения имён записей банковских карт.
	//

	t.Run("Получение имён записей банковских карт", func(t *testing.T) {

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
		resp, err := client.RequestBankCardName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка запроса имён данных банковских карт")

		// Сохранение принятых данных
		namesBankCard = resp.EntriesName
		_ = namesBankCard
	})

	//
	// Тест получения записи банковской карты, по имени (корректные данные, первой записи).
	//

	// Первая запись.
	rxBankCard1 := RxBankCard{}

	t.Run("Получение первой записи банковской карты по имени", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestBankCardByNameRequest{
			IdClient: "AAA",
			Name:     txBankCard1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		resp, err := client.RequestBankCardByName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка запроса данных текста, по имени записи")

		rxBankCard1.For = resp.Name
		rxBankCard1.Owner = resp.Owner
		rxBankCard1.Numb = resp.Numb
		rxBankCard1.ValidData = resp.Valid
		rxBankCard1.Code = resp.Code
		rxBankCard1.CreatedAt = resp.CreatedAt
	})

	//
	// Тест получения записи банковской карты, по имени (корректные данные, второй записи).
	//

	// Первая запись.
	rxBankCard2 := RxBankCard{}

	t.Run("Получение второй записи банковской карты по имени", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestBankCardByNameRequest{
			IdClient: "AAA",
			Name:     txBankCard2.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		resp, err := client.RequestBankCardByName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка запроса данных текста, по имени записи")

		rxBankCard2.For = resp.Name
		rxBankCard2.Owner = resp.Owner
		rxBankCard2.Numb = resp.Numb
		rxBankCard2.ValidData = resp.Valid
		rxBankCard2.Code = resp.Code
		rxBankCard2.CreatedAt = resp.CreatedAt
	})

	//
	// Тест получения записи банковской карты, по имени (отсутствующая запись).
	//

	t.Run("Получение записи банковской карты по отсутствующему имени", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestBankCardByNameRequest{
			IdClient: "AAA",
			Name:     "UnavailableName",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.RequestBankCardByName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка запроса к БД", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест получения записи банковской карты, по имени (нет имени).
	//

	t.Run("Получение записи банковской карты без имени", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestBankCardByNameRequest{
			IdClient: "AAA",
			Name:     "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.RequestBankCardByName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка обработки данных запроса", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест получения записи банковской карты, по имени (без id клиента).
	//

	t.Run("Получение записи банковской карты без id клиента", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestBankCardByNameRequest{
			IdClient: "",
			Name:     "Foo",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err := client.RequestBankCardByName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Internal desc = ошибка обработки данных запроса", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Проверка соответствия данных карт.
	//

	t.Run("Проверка соответствия данных банковских карт", func(t *testing.T) {

		assert.Equalf(t, txBankCard1.For, rxBankCard1.For, "Нет соответствия For, у первой записи")
		assert.Equalf(t, txBankCard1.Owner, rxBankCard1.Owner, "Нет соответствия Owner, у первой записи")
		assert.Equalf(t, txBankCard1.Numb, rxBankCard1.Numb, "Нет соответствия Numb, у первой записи")
		assert.Equalf(t, txBankCard1.ValidData, rxBankCard1.ValidData, "Нет соответствия ValidData, у первой записи")
		assert.Equalf(t, txBankCard1.Code, rxBankCard1.Code, "Нет соответствия Code, у первой записи")

		assert.Equalf(t, txBankCard2.For, rxBankCard2.For, "Нет соответствия For, у второй записи")
		assert.Equalf(t, txBankCard2.Owner, rxBankCard2.Owner, "Нет соответствия Owner, у второй записи")
		assert.Equalf(t, txBankCard2.Numb, rxBankCard2.Numb, "Нет соответствия Numb, у второй записи")
		assert.Equalf(t, txBankCard2.ValidData, rxBankCard2.ValidData, "Нет соответствия ValidData, у второй записи")
		assert.Equalf(t, txBankCard2.Code, rxBankCard2.Code, "Нет соответствия Code, у второй записи")
	})

	//
	// Тест удаления данных банковской карты. (первой записи)
	//

	t.Run("Удаление данных банковской карты. (первой записи)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     txBankCard1.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteBankCard(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка удаления записи")
	})

	//
	// Тест удаления данных банковской карты. (второй записи)
	//

	t.Run("Удаление данных банковской карты. (второй записи)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     txBankCard2.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteBankCard(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка удаления записи")
	})

	//
	// Тест удаления данных банковской карты. (подставной токен)
	//

	t.Run("Удаление данных банковской карты. (подставной токен)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth+"1")

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     txBankCard2.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест удаления данных банковской карты. (нет id клиента)
	//

	t.Run("Удаление данных банковской карты. (нет id клиента)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "",
			Name:     txBankCard2.For,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Unavailable desc = ошибка в данных запроса", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест удаления данных банковской карты. (нет имени)
	//

	t.Run("Удаление данных банковской карты. (нет имени)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &proto.RequestDeleteName{
			IdClient: "AAA",
			Name:     "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteBankCard(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = Unavailable desc = ошибка в данных запроса", err.Error(), "Нет соответствия ошибки")
	})

	// -------------------------------------------------------------------------------------------------
	//
	//                                             Файлы
	//
	// -------------------------------------------------------------------------------------------------

	//
	// --- Подготовка ---
	//

	// Создание директории для файлов.
	dirFiles := flags.NameSubDirFiles
	err = createSubdirectory(dirFiles)
	require.NoErrorf(t, err, "Ошибка создания директории для файлов")
	defer func() {
		err := removeDirIfExists(dirFiles)
		assert.NoErrorf(t, err, "Ошибка удаления директории для файлов")
	}()

	// Создание файлов для передачи.
	nameFileA := "fileA.txt"
	nameFileB := "fileB.txt"
	dataText := "Foo"

	err = fileCreateAndWrite(dirFiles, nameFileA, dataText)
	require.NoErrorf(t, err, "Ошибка в файле <%s>", nameFileA)
	err = fileCreateAndWrite(dirFiles, nameFileB, dataText)
	require.NoErrorf(t, err, "Ошибка в файле <%s>", nameFileB)

	//
	// Получение имён файлов (подставной токен).
	//

	t.Run("Получение имён файлов (подставной токен)", func(t *testing.T) {

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
		_, err := client.RequestFileName(ctx, req, grpc.Header(&header))
		require.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Получение имён файлов.
	//

	rxNameFiles := []string{}

	t.Run("Получение имён файлов", func(t *testing.T) {

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
		resp, err := client.RequestFileName(ctx, req, grpc.Header(&header))
		require.NoErrorf(t, err, "Ошибка получения имён файлов")

		rxNameFiles = resp.EntriesName
		isBusy := resp.IsBusy
		assert.Falsef(t, isBusy, "Сервер не должен быть занятым")
	})

	//
	// Проверка соответствия имён файлов.
	//

	t.Run("Проверка соответствия имён файлов", func(t *testing.T) {

		require.Equalf(t, 2, len(rxNameFiles), "Нет соответствия длинны масива с названиеями файлов")

		isOk := false

		if rxNameFiles[0] == nameFileA && rxNameFiles[1] == nameFileB {
			isOk = true
		}
		if rxNameFiles[0] == nameFileB && rxNameFiles[1] == nameFileA {
			isOk = true
		}
		assert.Truef(t, isOk, "Нет соответствия имён файлов")
	})

	//
	// Тест удаления файла (А).
	//

	t.Run("Удаление файла (А)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &pb.RequestDeleteName{
			IdClient: "AAA",
			Name:     nameFileA,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteFile(ctx, req, grpc.Header(&header))
		assert.NoErrorf(t, err, "Ошибка удаления файла <%s>", nameFileA)
	})

	//
	// Тест удаления файла (B).
	//

	t.Run("Удаление файла (B)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &pb.RequestDeleteName{
			IdClient: "AAA",
			Name:     nameFileB,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteFile(ctx, req, grpc.Header(&header))
		assert.NoErrorf(t, err, "Ошибка удаления файла <%s>", nameFileB)
	})

	//
	// Тест удаления файла (подставной токен).
	//

	t.Run("Удаление файла (подставной токен)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth+"1")

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &pb.RequestDeleteName{
			IdClient: "AAA",
			Name:     nameFileB,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteFile(ctx, req, grpc.Header(&header))
		assert.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест удаления файла (нет id клиента).
	//

	t.Run("Удаление файла (нет id клиента)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &pb.RequestDeleteName{
			IdClient: "",
			Name:     nameFileB,
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteFile(ctx, req, grpc.Header(&header))
		assert.Equalf(t, "rpc error: code = Unavailable desc = ошибка в данных запроса", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест удаления файла (нет имени файла).
	//

	t.Run("Удаление файла (нет имени файла)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &pb.RequestDeleteName{
			IdClient: "ААА",
			Name:     "",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteFile(ctx, req, grpc.Header(&header))
		assert.Equalf(t, "rpc error: code = Unavailable desc = ошибка в данных запроса", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Тест удаления файла (отсутствующий файл).
	//

	t.Run("Удаление файла (отсутствующий файл)", func(t *testing.T) {

		// Создание метаданных с токеном аутентификации.
		nameToken := "token"
		txMD := metadata.Pairs(nameToken, tokenAuth)

		// Контекст для запроса.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &pb.RequestDeleteName{
			IdClient: "ААА",
			Name:     "MissingFile",
		}

		var header metadata.MD

		// Выполнение запроса.
		_, err = client.DeleteFile(ctx, req, grpc.Header(&header))
		assert.Equalf(t, "rpc error: code = Internal desc = ошибка удаления файла", err.Error(), "Нет соответствия ошибки")
	})

	//
	// Передача файла (подставной токен).
	//

	t.Run("Передача файла (подставной токен)", func(t *testing.T) {

		invalidToken := tokenAuth + "1"
		txMD := metadata.Pairs("token", invalidToken)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		ctx = metadata.NewOutgoingContext(ctx, txMD)

		var header metadata.MD

		// Тест
		stream, err := client.SendFile(ctx, grpc.Header(&header))
		if err != nil {
			t.Fatalf("Ошибка при создании потока: %v", err)
		}

		req := &pb.SendFileRequest{
			IdClient: "ААА",
			FileName: "Foo.txt",
			Content:  []byte("content"),
		}
		err = stream.Send(req)
		require.NoErrorf(t, err, "Ошибка при отправке чанка")

		_, err = stream.CloseAndRecv()
		assert.Equalf(t, "rpc error: code = PermissionDenied desc = токен не прошел проверку", err.Error(), "Нет соответствия ошибки")
	})
}

//
// Тесты ошибок
//

// Конструктор.
func TestNew(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			l        *zap.Logger
			s        domain.DomainI
			f        *flags.Config
			wantErr  error
		}{
			{
				nameTest: "Нет указателя на логгер",
				l:        nil,
				s:        serv.Storage,
				f:        serv.Flag,
				wantErr:  NilPtrArgumentL,
			},
			{
				nameTest: "Нет указателя на хранилище",
				l:        serv.Lgr,
				s:        nil,
				f:        serv.Flag,
				wantErr:  NilPtrArgumentS,
			},
			{
				nameTest: "Нет указателя на флаги",
				l:        serv.Lgr,
				s:        serv.Storage,
				f:        nil,
				wantErr:  NilPtrArgumentF,
			},
		}

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err := New(tt.l, tt.s, tt.f)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})
}

// Ping.
func TestPing(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			empty    *emptypb.Empty
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				empty:    &emptypb.Empty{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет пустого указателя",
				ctx:      context.Background(),
				empty:    nil,
				wantErr:  NilPtrArgumentEmpty,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.Ping(tt.ctx, tt.empty)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Нет указателя на логгер", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		inst.logger = nil

		_, err = inst.Ping(context.Background(), &emptypb.Empty{})
		assert.Equalf(t, NilPtrLogger, err, "Нет соответствия ошибки")

	})

	t.Run("Метаданные отсутствуют", func(t *testing.T) {
		err := resetForTest()
		require.NoError(t, err, "Ошибка сброса конструктора")

		serv, err := prepare()
		require.NoError(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoError(t, err, "Ошибка конструктора")

		_, err = inst.Ping(context.Background(), &emptypb.Empty{})
		assert.Equal(t, "rpc error: code = NotFound desc = отсутствуют метаданные", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Отсутствует токен в метаданных", func(t *testing.T) {
		err := resetForTest()
		require.NoError(t, err, "Ошибка сброса конструктора")

		serv, err := prepare()
		require.NoError(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoError(t, err, "Ошибка конструктора")

		nameToken := "token"
		txMD := metadata.Pairs(nameToken, "")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		_, err = inst.Ping(ctx, &emptypb.Empty{})
		assert.Equal(t, "rpc error: code = NotFound desc = отсутствуют метаданные", err.Error(), "Нет соответствия ошибки")
	})

}

// LocalBackupFile.
func TestLocalBackupFile(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			stream   pb.PasswordManager_LocalBackupFileServer
			wantErr  error
		}{
			{
				nameTest: "Нет потока",
				stream:   nil,
				wantErr:  NilPtrArgumentStream,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				err = inst.LocalBackupFile(tt.stream)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

}

// LocalRestoreFile.
func TestLocalRestoreFile(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			req      *pb.LocalRestoreFileRequest
			stream   pb.PasswordManager_LocalRestoreFileServer
			wantErr  error
		}{
			{
				nameTest: "Нет потока",
				req:      &pb.LocalRestoreFileRequest{},
				stream:   nil,
				wantErr:  NilPtrArgumentEmpty,
			},
			{
				nameTest: "Нет запроса",
				req:      nil,
				stream:   nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				err = inst.LocalRestoreFile(tt.req, tt.stream)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})
}

// LocalFilesInfo.
func TestLocalFilesInfo(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *emptypb.Empty
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &emptypb.Empty{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.LocalFilesInfo(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.LocalFilesInfo(ctx, &emptypb.Empty{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.LocalFilesInfo(context.Background(), &emptypb.Empty{})
		assert.Equalf(t, "rpc error: code = InvalidArgument desc = ошибка извлечения заголовков", err.Error(), "Нет соответствия ошибки")

	})

	t.Run("Отсутствует токен в метаданных", func(t *testing.T) {
		err := resetForTest()
		require.NoError(t, err, "Ошибка сброса конструктора")

		serv, err := prepare()
		require.NoError(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoError(t, err, "Ошибка конструктора")

		nameToken := "token"
		txMD := metadata.Pairs(nameToken, "")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		_, err = inst.LocalFilesInfo(context.Background(), &emptypb.Empty{})
		assert.Equalf(t, "rpc error: code = InvalidArgument desc = ошибка извлечения заголовков", err.Error(), "Нет соответствия ошибки")
	})
}

// Registration.
func TestRegistration(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.RegistrationRequest
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.RegistrationRequest{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.Registration(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}
	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.Registration(ctx, &pb.RegistrationRequest{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.RegistrationRequest{
			UserName:      "Foo",
			UserPwd:       "Bar",
			UserPwdRepeat: "Bar",
		}
		_, err = inst.Registration(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Unauthenticated desc = Отсутствуют метаданные", err.Error(), "Нет соответствия ошибки")

	})

	t.Run("Отсутствует токен в метаданных", func(t *testing.T) {
		err := resetForTest()
		require.NoError(t, err, "Ошибка сброса конструктора")

		serv, err := prepare()
		require.NoError(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoError(t, err, "Ошибка конструктора")

		nameToken := "token"
		txMD := metadata.Pairs(nameToken, "")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		_, err = inst.Registration(context.Background(), &pb.RegistrationRequest{})
		assert.Equalf(t, "rpc error: code = Unauthenticated desc = Отсутствуют метаданные", err.Error(), "Нет соответствия ошибки")
	})
}

// Authentication.
func TestAuthentication(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.AuthenticationRequest
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.AuthenticationRequest{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.Authentication(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.Authentication(ctx, &pb.AuthenticationRequest{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.AuthenticationRequest{
			UserName: "Foo",
			UserPwd:  "Bar",
		}
		_, err = inst.Authentication(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Unavailable desc = ошибка получения токена", err.Error(), "Нет соответствия ошибки")

	})

	t.Run("Отсутствует токен в метаданных", func(t *testing.T) {
		err := resetForTest()
		require.NoError(t, err, "Ошибка сброса конструктора")

		serv, err := prepare()
		require.NoError(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoError(t, err, "Ошибка конструктора")

		nameToken := "token"
		txMD := metadata.Pairs(nameToken, "")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		ctx = metadata.NewOutgoingContext(ctx, txMD)

		req := &pb.AuthenticationRequest{
			UserName: "Foo",
			UserPwd:  "Bar",
		}
		_, err = inst.Authentication(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Unavailable desc = ошибка получения токена", err.Error(), "Нет соответствия ошибки")
	})
}

// SendLoginPassword.
func TestSendLoginPassword(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.SendLoginPasswordRequest
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.SendLoginPasswordRequest{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.SendLoginPassword(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.SendLoginPassword(ctx, &pb.SendLoginPasswordRequest{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.SendLoginPasswordRequest{
			IdClient:  "A",
			For:       "B",
			Login:     "C",
			Password:  "D",
			CreatedAt: "E",
		}
		_, err = inst.SendLoginPassword(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Internal desc = ошибка получения токена запроса", err.Error(), "Нет соответствия ошибки")

	})
}

// SendText.
func TestSendText(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.SendTextRequest
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.SendTextRequest{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.SendText(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.SendText(ctx, &pb.SendTextRequest{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})
}

// SendBankCard.
func TestSendBankCard(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.SendBankCardRequest
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.SendBankCardRequest{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.SendBankCard(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.SendBankCard(ctx, &pb.SendBankCardRequest{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.SendBankCardRequest{
			IdClient:  "A",
			For:       "B",
			Owner:     "C",
			Numb:      "49927398716",
			ValidData: "E",
			Code:      "F",
			CreatedAt: "G",
		}
		_, err = inst.SendBankCard(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Internal desc = ошибка получения токена запроса", err.Error(), "Нет соответствия ошибки")

	})
}

// SendFile.
func TestSendFile_FAULT(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			stream   pb.PasswordManager_SendFileServer
			wantErr  error
		}{
			{
				nameTest: "Нет потока",
				stream:   nil,
				wantErr:  NilPtrArgumentStream,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				err = inst.SendFile(tt.stream)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Передача файла", func(t *testing.T) {

	})

}

// RequestLoginPasswordName.
func TestRequestLoginPasswordName(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			empty    *emptypb.Empty
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				empty:    &emptypb.Empty{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				empty:    nil,
				wantErr:  NilPtrArgumentEmpty,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.RequestLoginPasswordName(tt.ctx, tt.empty)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.RequestLoginPasswordName(ctx, &emptypb.Empty{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &emptypb.Empty{}
		_, err = inst.RequestLoginPasswordName(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Internal desc = Ошибка получения токена аутентификации", err.Error(), "Нет соответствия ошибки")

	})
}

// RequestLoginPasswordByName.
func TestRequestLoginPasswordByName(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.RequestLoginPasswordByNameRequest
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.RequestLoginPasswordByNameRequest{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.RequestLoginPasswordByName(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.RequestLoginPasswordByName(ctx, &pb.RequestLoginPasswordByNameRequest{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.RequestLoginPasswordByNameRequest{
			IdClient: "A",
			Name:     "B",
		}

		_, err = inst.RequestLoginPasswordByName(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Internal desc = Ошибка получения токена аутентификации", err.Error(), "Нет соответствия ошибки")

	})
}

// RequestTextName.
func TestRequestTextName(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			empty    *emptypb.Empty
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				empty:    &emptypb.Empty{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				empty:    nil,
				wantErr:  NilPtrArgumentEmpty,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.RequestTextName(tt.ctx, tt.empty)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.RequestTextName(ctx, &emptypb.Empty{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &emptypb.Empty{}

		_, err = inst.RequestTextName(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Internal desc = Ошибка получения токена аутентификации", err.Error(), "Нет соответствия ошибки")

	})
}

// RequestTextByName.
func TestRequestTextByName(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.RequestTextByNameRequest
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.RequestTextByNameRequest{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.RequestTextByName(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.RequestTextByName(ctx, &pb.RequestTextByNameRequest{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.RequestTextByNameRequest{
			IdClient: "A",
			Name:     "B",
		}

		_, err = inst.RequestTextByName(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Internal desc = Ошибка получения токена аутентификации", err.Error(), "Нет соответствия ошибки")

	})
}

// RequestBankCardName.
func TestRequestBankCardName(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *emptypb.Empty
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &emptypb.Empty{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentEmpty,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.RequestBankCardName(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.RequestBankCardName(ctx, &emptypb.Empty{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &emptypb.Empty{}

		_, err = inst.RequestBankCardName(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Internal desc = Ошибка получения токена аутентификации", err.Error(), "Нет соответствия ошибки")

	})
}

// RequestBankCardByName.
func TestRequestBankCardByName(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.RequestBankCardByNameRequest
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.RequestBankCardByNameRequest{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.RequestBankCardByName(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.RequestBankCardByName(ctx, &pb.RequestBankCardByNameRequest{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.RequestBankCardByNameRequest{
			IdClient: "A",
			Name:     "B",
		}

		_, err = inst.RequestBankCardByName(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Internal desc = Ошибка получения токена аутентификации", err.Error(), "Нет соответствия ошибки")

	})
}

// RequestFileName.
func TestRequestFileName(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *emptypb.Empty
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &emptypb.Empty{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentEmpty,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.RequestFileName(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.RequestFileName(ctx, &emptypb.Empty{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &emptypb.Empty{}

		_, err = inst.RequestFileName(context.Background(), req)
		assert.Equalf(t, "rpc error: code = Internal desc = Ошибка получения токена аутентификации", err.Error(), "Нет соответствия ошибки")

	})
}

// RequestFileByName.
func TestRequestFileByName(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			req      *pb.RequestFileByNameRequest
			stream   pb.PasswordManager_RequestFileByNameServer
			wantErr  error
		}{
			{
				nameTest: "Нет запроса",
				req:      nil,
				stream:   nil,
				wantErr:  NilPtrArgumentReq,
			},
			{
				nameTest: "Нет потока",
				req:      &pb.RequestFileByNameRequest{},
				stream:   nil,
				wantErr:  NilPtrArgumentStream,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				err = inst.RequestFileByName(tt.req, tt.stream)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}
	})
}

// RequestFileInfo.
func TestRequestFileInfo(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.RequestFileInfoRequest
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.RequestFileInfoRequest{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.RequestFileInfo(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.RequestFileInfo(ctx, &pb.RequestFileInfoRequest{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.RequestFileInfoRequest{
			IdClient: "A",
			Name:     "B",
		}

		_, err = inst.RequestFileInfo(context.Background(), req)
		assert.Equalf(t, "rpc error: code = PermissionDenied desc = ошибка получения токена", err.Error(), "Нет соответствия ошибки")

	})
}

// DeleteLoginPassword.
func TestDeleteLoginPassword(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.RequestDeleteName
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.RequestDeleteName{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.DeleteLoginPassword(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.DeleteLoginPassword(ctx, &pb.RequestDeleteName{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.RequestDeleteName{
			IdClient: "A",
			Name:     "B",
		}

		_, err = inst.DeleteLoginPassword(context.Background(), req)
		assert.Equalf(t, "rpc error: code = PermissionDenied desc = ошибка получения токена", err.Error(), "Нет соответствия ошибки")

	})
}

// DeleteText.
func TestDeleteText(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.RequestDeleteName
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.RequestDeleteName{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.DeleteText(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.DeleteText(ctx, &pb.RequestDeleteName{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.RequestDeleteName{
			IdClient: "A",
			Name:     "B",
		}

		_, err = inst.DeleteText(context.Background(), req)
		assert.Equalf(t, "rpc error: code = PermissionDenied desc = ошибка получения токена", err.Error(), "Нет соответствия ошибки")

	})
}

// DeleteBankCard.
func TestDeleteBankCard(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.RequestDeleteName
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.RequestDeleteName{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.DeleteBankCard(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.DeleteBankCard(ctx, &pb.RequestDeleteName{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.RequestDeleteName{
			IdClient: "A",
			Name:     "B",
		}

		_, err = inst.DeleteBankCard(context.Background(), req)
		assert.Equalf(t, "rpc error: code = PermissionDenied desc = ошибка получения токена", err.Error(), "Нет соответствия ошибки")

	})
}

// DeleteFile.
func TestDeleteFile(t *testing.T) {

	t.Run("Ошибки аргументов", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		// Данные тестов.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			req      *pb.RequestDeleteName
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				req:      &pb.RequestDeleteName{},
				wantErr:  NilPtrArgumentCtx,
			},
			{
				nameTest: "Нет запроса",
				ctx:      context.Background(),
				req:      nil,
				wantErr:  NilPtrArgumentReq,
			},
		}

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				_, err = inst.DeleteFile(tt.ctx, tt.req)
				assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
			})
		}

	})

	t.Run("Долгая обработка", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()

		time.Sleep(4 * time.Millisecond)

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		_, err = inst.DeleteFile(ctx, &pb.RequestDeleteName{})
		assert.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Нет метаданных", func(t *testing.T) {

		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		req := &pb.RequestDeleteName{
			IdClient: "A",
			Name:     "B",
		}

		_, err = inst.DeleteFile(context.Background(), req)
		assert.Equalf(t, "rpc error: code = PermissionDenied desc = ошибка получения токена", err.Error(), "Нет соответствия ошибки")

	})
}

//
// Тесты статусов.
//

// UpdateStatusBackUp и GetStatusBackUp
func TestStatusBackUp(t *testing.T) {

	t.Run("Корректный код статуса", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusBackUp(StageActive)
		assert.NoErrorf(t, err, "Неожиданная ошибка")
	})

	t.Run("Ошибочный код статуса", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusBackUp(StageActive + 100)
		assert.Equalf(t, IncorrectStage, err, "Нет соответствия ошибки")
	})

	t.Run("Проверка обновления", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusBackUp(StageActive)
		require.NoErrorf(t, err, "Неожиданная ошибка")

		data := inst.GetStatusBackUp()
		assert.Equalf(t, StageActive, int(data), "Нет соответствия кодов")
	})

}

// UpdateStatusRestore и GetStatusRestore
func TestStatusRestore(t *testing.T) {

	t.Run("Корректный код статуса", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusRestore(StageActive)
		assert.NoErrorf(t, err, "Неожиданная ошибка")
	})

	t.Run("Ошибочный код статуса", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusRestore(StageActive + 100)
		assert.Equalf(t, IncorrectStage, err, "Нет соответствия ошибки")
	})

	t.Run("Проверка обновления", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusRestore(StageActive)
		require.NoErrorf(t, err, "Неожиданная ошибка")

		data := inst.GetStatusRestore()
		assert.Equalf(t, StageActive, int(data), "Нет соответствия кодов")
	})

}

// UpdateStatusRx и GetStatusRx
func TestStatusRx(t *testing.T) {

	t.Run("Корректный код статуса", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusRx(StageActive)
		assert.NoErrorf(t, err, "Неожиданная ошибка")
	})

	t.Run("Ошибочный код статуса", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusRx(StageActive + 100)
		assert.Equalf(t, IncorrectStage, err, "Нет соответствия ошибки")
	})

	t.Run("Проверка обновления", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusRx(StageActive)
		require.NoErrorf(t, err, "Неожиданная ошибка")

		data := inst.GetStatusRx()
		assert.Equalf(t, StageActive, int(data), "Нет соответствия кодов")
	})

}

// UpdateStatusTx и GetStatusTx
func TestStatusTx(t *testing.T) {

	t.Run("Корректный код статуса", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusTx(StageActive)
		assert.NoErrorf(t, err, "Неожиданная ошибка")
	})

	t.Run("Ошибочный код статуса", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusTx(StageActive + 100)
		assert.Equalf(t, IncorrectStage, err, "Нет соответствия ошибки")
	})

	t.Run("Проверка обновления", func(t *testing.T) {
		err := resetForTest()
		require.NoErrorf(t, err, "Ошибка сброса конструктора")

		// Подготовка конфигурации сервиса.
		serv, err := prepare()
		require.NoErrorf(t, err, "Ошибка создания экземпляра сервиса")

		inst, err := New(serv.Lgr, serv.Storage, serv.Flag)
		require.NoErrorf(t, err, "Ошибка конструктора")

		err = inst.UpdateStatusTx(StageActive)
		require.NoErrorf(t, err, "Неожиданная ошибка")

		data := inst.GetStatusTx()
		assert.Equalf(t, StageActive, int(data), "Нет соответствия кодов")
	})

}

// =================================================================================================
//
//                               Вспомогательные функции для тестов
//
// =================================================================================================

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
	srvGRPC, err := New(lgr, storage, flag)
	if err != nil {
		return nil, fmt.Errorf("Конструктор сервиса вернул ошибку: <%w>", err)
	}

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

// Создание файла и запись в него данных. Возвращается ошибка.
//
// Параметры:
//
//	filePath - путь к файлу.
//	fileName - имя файла.
//	dataText - данные для записи в файл.
func fileCreateAndWrite(filePath, fileName, dataText string) (err error) {

	fullFileName := path.Join(filePath, fileName)

	file, err := os.Create(fullFileName)
	if err != nil {
		return fmt.Errorf("Ошибка создания файла <%w>", err)
	}
	defer file.Close()

	// Буфер.
	writer := bufio.NewWriter(file)

	// Запись в файл.
	_, err = writer.WriteString(dataText)
	if err != nil {
		return fmt.Errorf("Ошибка записи в файл <%w>", err)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("Ошибка сброса буфера <%w>", err)

	}

	return nil
}
