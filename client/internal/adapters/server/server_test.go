package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Перечень мок функций.
type MockClient struct {
	AuthenticationFunc             func(ctx context.Context, req *proto.AuthenticationRequest, opts ...grpc.CallOption) (*proto.AuthenticationResponse, error)
	PingFunc                       func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	LocalBackupFileFunc            func(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.LocalBackupFileRequest, proto.LocalBackupFileResponse], error)
	LocalRestoreFileFunc           func(ctx context.Context, in *proto.LocalRestoreFileRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.LocalRestoreFileResponse], error)
	LocalFilesInfoFunc             func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.LocalFilesInfoResponse, error)
	RegistrationFunc               func(ctx context.Context, in *proto.RegistrationRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendLoginPasswordFunc          func(ctx context.Context, in *proto.SendLoginPasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendTextFunc                   func(ctx context.Context, in *proto.SendTextRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendBankCardFunc               func(ctx context.Context, in *proto.SendBankCardRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendFileFunc                   func(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.SendFileRequest, proto.SendFileResponse], error)
	RequestLoginPasswordNameFunc   func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestLoginPasswordNameResponse, error)
	RequestLoginPasswordByNameFunc func(ctx context.Context, in *proto.RequestLoginPasswordByNameRequest, opts ...grpc.CallOption) (*proto.RequestLoginPasswordByNameResponse, error)
	RequestTextNameFunc            func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestTextNameResponse, error)
	RequestTextByNameFunc          func(ctx context.Context, in *proto.RequestTextByNameRequest, opts ...grpc.CallOption) (*proto.RequestTextByNameResponse, error)
	RequestBankCardNameFunc        func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestBankCardNameResponse, error)
	RequestBankCardByNameFunc      func(ctx context.Context, in *proto.RequestBankCardByNameRequest, opts ...grpc.CallOption) (*proto.RequestBankCardByNameResponse, error)
	RequestFileInfoFunc            func(ctx context.Context, in *proto.RequestFileInfoRequest, opts ...grpc.CallOption) (*proto.RequestFileInfoResponse, error)
	RequestFileNameFunc            func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestFileNameResponse, error)
	RequestFileByNameFunc          func(ctx context.Context, in *proto.RequestFileByNameRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.RequestFileByNameResponse], error)
	DeleteLoginPasswordFunc        func(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteTextFunc                 func(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteBankCardFunc             func(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteFileFunc                 func(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

// Аутентификация.
func (t *MockClient) Authentication(ctx context.Context, in *proto.AuthenticationRequest, opts ...grpc.CallOption) (*proto.AuthenticationResponse, error) {
	return t.AuthenticationFunc(ctx, in, opts...)
}

// Проверка связи.
func (t *MockClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.PingFunc(ctx, in, opts...)
}

// Резервное копирования (режим - локальный).
func (t *MockClient) LocalBackupFile(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.LocalBackupFileRequest, proto.LocalBackupFileResponse], error) {
	return t.LocalBackupFileFunc(ctx, opts...)
}

// Восстановление из резервной копии (режим-локальный).
func (t *MockClient) LocalRestoreFile(ctx context.Context, in *proto.LocalRestoreFileRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.LocalRestoreFileResponse], error) {
	return t.LocalRestoreFileFunc(ctx, in, opts...)
}

// Информация по файлам.
func (t *MockClient) LocalFilesInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.LocalFilesInfoResponse, error) {
	return t.LocalFilesInfoFunc(ctx, in, opts...)
}

// Регистрация.
func (t *MockClient) Registration(ctx context.Context, in *proto.RegistrationRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.RegistrationFunc(ctx, in, opts...)
}

// Передача данных логин/пароль.
func (t *MockClient) SendLoginPassword(ctx context.Context, in *proto.SendLoginPasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.SendLoginPasswordFunc(ctx, in, opts...)
}

// Передача данных текста.
func (t *MockClient) SendText(ctx context.Context, in *proto.SendTextRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.SendTextFunc(ctx, in, opts...)
}

// Передача данных банковской карты.
func (t *MockClient) SendBankCard(ctx context.Context, in *proto.SendBankCardRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.SendBankCardFunc(ctx, in, opts...)
}

// Передача файла.
func (t *MockClient) SendFile(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.SendFileRequest, proto.SendFileResponse], error) {
	return t.SendFileFunc(ctx, opts...)
}

// Запрос имён записей логин/пароль.
func (t *MockClient) RequestLoginPasswordName(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestLoginPasswordNameResponse, error) {
	return t.RequestLoginPasswordNameFunc(ctx, in, opts...)
}

// Запрос данных логин/пароль по имени записи.
func (t *MockClient) RequestLoginPasswordByName(ctx context.Context, in *proto.RequestLoginPasswordByNameRequest, opts ...grpc.CallOption) (*proto.RequestLoginPasswordByNameResponse, error) {
	return t.RequestLoginPasswordByNameFunc(ctx, in, opts...)
}

// Запрос имён записей данных текста.
func (t *MockClient) RequestTextName(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestTextNameResponse, error) {
	return t.RequestTextNameFunc(ctx, in, opts...)
}

// Запрос данных текста по имени записи.
func (t *MockClient) RequestTextByName(ctx context.Context, in *proto.RequestTextByNameRequest, opts ...grpc.CallOption) (*proto.RequestTextByNameResponse, error) {
	return t.RequestTextByNameFunc(ctx, in, opts...)
}

// Запрос имён записей банковских карт.
func (t *MockClient) RequestBankCardName(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestBankCardNameResponse, error) {
	return t.RequestBankCardNameFunc(ctx, in, opts...)
}

// Запрос данных банковской карты по имени записи.
func (t *MockClient) RequestBankCardByName(ctx context.Context, in *proto.RequestBankCardByNameRequest, opts ...grpc.CallOption) (*proto.RequestBankCardByNameResponse, error) {
	return t.RequestBankCardByNameFunc(ctx, in, opts...)
}

// Запрос информации по файлу.
func (t *MockClient) RequestFileInfo(ctx context.Context, in *proto.RequestFileInfoRequest, opts ...grpc.CallOption) (*proto.RequestFileInfoResponse, error) {
	return t.RequestFileInfoFunc(ctx, in, opts...)
}

// Запрос имён файлов.
func (t *MockClient) RequestFileName(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestFileNameResponse, error) {
	return t.RequestFileNameFunc(ctx, in, opts...)
}

// Запрос файла по имени.
func (t *MockClient) RequestFileByName(ctx context.Context, in *proto.RequestFileByNameRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.RequestFileByNameResponse], error) {
	return t.RequestFileByNameFunc(ctx, in, opts...)
}

// Удаление записи данных логин/пароль.
func (t *MockClient) DeleteLoginPassword(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.DeleteLoginPasswordFunc(ctx, in, opts...)
}

// УУдаление записи текста.
func (t *MockClient) DeleteText(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.DeleteTextFunc(ctx, in, opts...)
}

// Удаление записи банковской карты.
func (t *MockClient) DeleteBankCard(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.DeleteBankCardFunc(ctx, in, opts...)
}

// Удаление файла.
func (t *MockClient) DeleteFile(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return t.DeleteFile(ctx, in, opts...)
}

//
// --- AuthenticationContext ---
//

func TestAuthenticationContext_SUCCESS(t *testing.T) {

	tokenAuth := "Foo"

	// Мок.
	mockClient := &MockClient{
		AuthenticationFunc: func(ctx context.Context, req *proto.AuthenticationRequest, opts ...grpc.CallOption) (*proto.AuthenticationResponse, error) {

			// Входящие метаданные
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Подготовка header
			header := metadata.Pairs(nameToken, tokens[0])
			for _, opt := range opts {
				if hdr, ok := opt.(grpc.HeaderCallOption); ok {
					*hdr.HeaderAddr = header
				}
			}

			// Токен аутентификации
			res := &proto.AuthenticationResponse{
				Token: tokenAuth,
			}
			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxTokenAuth, err := s.AuthenticationContext(context.Background(), "testUser", "testPwd")
	require.NoErrorf(t, err, "ошибка запроса")

	assert.Equalf(t, tokenAuth, rxTokenAuth, "Нет соответствия токенов аутентификации")
}

func TestAuthenticationContext_FAULT(t *testing.T) {

	// Мок
	mockClient := &MockClient{
		AuthenticationFunc: func(ctx context.Context, req *proto.AuthenticationRequest, opts ...grpc.CallOption) (*proto.AuthenticationResponse, error) {

			// Входящие метаданные
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Подготовка header
			header := metadata.Pairs(nameToken, tokens[0])
			for _, opt := range opts {
				if hdr, ok := opt.(grpc.HeaderCallOption); ok {
					*hdr.HeaderAddr = header
				}
			}

			// Токен аутентификации
			res := &proto.AuthenticationResponse{
				Token: "Foo",
			}
			return res, nil
		},
	}

	s := &server{client: mockClient}

	// Данные для тестов.
	dataTest := []struct {
		nameTest string
		userName string
		userPwd  string
		wantErr  error
	}{
		{
			nameTest: "Без имени",
			userName: "",
			userPwd:  "Foo",
			wantErr:  errors.New("в аргументе <userName>, нет данных"),
		},
		{
			nameTest: "Без пароля",
			userName: "Foo",
			userPwd:  "",
			wantErr:  errors.New("в аргументе <userPwd>, нет данных"),
		},
	}

	// Тесты.
	for _, data := range dataTest {
		t.Run(data.nameTest, func(t *testing.T) {
			_, err := s.AuthenticationContext(context.Background(), data.userName, data.userPwd)
			assert.Equalf(t, data.wantErr, err, "нет соответствия ошибки при отсутствии userName")
		})
	}

}

//
// --- New ---
//

func TestNew(t *testing.T) {
	t.Run("Успешное создание", func(t *testing.T) {

		certPath := "tls/server.crt"
		require.FileExists(t, certPath)

		listener, err := net.Listen("tcp", "localhost:50061")
		require.NoError(t, err)

		server := grpc.NewServer()
		go func() { server.Serve(listener) }()
		defer func() {
			server.GracefulStop()
			listener.Close()
		}()
		time.Sleep(100 * time.Millisecond)

		// Вызов New.
		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		// Закрытие подключения.
		act.ConnectClose()
		require.NoError(t, err)
	})
}

//
// --- InitDataBackUp ---
//

func TestInitDataBackUp(t *testing.T) {

	t.Run("Успешная инициализация", func(t *testing.T) {

		resetInstServer()

		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		listFiles := []string{"Foo.txt", "Bar.txt"}
		sizeSendFile := 100
		key := createKey()

		err = act.InitDataBackUp(listFiles, int64(sizeSendFile), key)
		require.NoError(t, err)
	})

	t.Run("Нет имён файлов", func(t *testing.T) {

		resetInstServer()

		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		listFiles := []string{}
		sizeSendFile := 100
		key := createKey()

		err = act.InitDataBackUp(listFiles, int64(sizeSendFile), key)
		require.Equalf(t, NotCorrectLenLestFiles, err, "Нет соответствия ошибки")
	})

	t.Run("Одно имя файла", func(t *testing.T) {

		resetInstServer()

		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		listFiles := []string{"Foo.txt"}
		sizeSendFile := 100
		key := createKey()

		err = act.InitDataBackUp(listFiles, int64(sizeSendFile), key)
		require.Equalf(t, NotCorrectLenLestFiles, err, "Нет соответствия ошибки")
	})

	t.Run("Размер файла нулевой", func(t *testing.T) {

		resetInstServer()

		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		listFiles := []string{"Foo.txt", "Bar.txt"}
		sizeSendFile := 0
		key := createKey()

		err = act.InitDataBackUp(listFiles, int64(sizeSendFile), key)
		require.Equalf(t, EmptyDataArgumentSizeSendFile, err, "Нет соответствия ошибки")
	})

	t.Run("Размер файла отрицательный", func(t *testing.T) {

		resetInstServer()

		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		listFiles := []string{"Foo.txt", "Bar.txt"}
		sizeSendFile := -10
		key := createKey()

		err = act.InitDataBackUp(listFiles, int64(sizeSendFile), key)
		require.Equalf(t, EmptyDataArgumentSizeSendFile, err, "Нет соответствия ошибки")
	})
}

//
// --- BackUp ---
//

/*
func TestBackUp(t *testing.T) {

	mockClient := &MockClient{
		LocalBackupFileFunc: func(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.LocalBackupFileRequest, proto.LocalBackupFileResponse], error) {

		},
	}

	// Создание экземпляра.
	act := &server{client: mockClient}

	// Создание временных файлов.
	err := createFileAndFill("Foo.txt", "Foo")
	require.NoErrorf(t, err, "Ошибка создания Foo.txt")
	err = createFileAndFill("Bar.txt", "Bar")
	require.NoErrorf(t, err, "Ошибка создания Bar.txt")

	// Определение общего размера файлов.
	files := []string{"Foo.txt", "Bar.txt"}
	totalSize, err := totalFileSize(files)
	require.NoErrorf(t, err, "Ошибка определения общего размера файлов")

	// Создание ключа шифрования.
	key := createKey()

	// Инициализация данных процесса BackUp.
	err = act.InitDataBackUp(files, totalSize, key)
	require.NoErrorf(t, err, "Ошибка инициализации данных")

	// Каналы для процесса BackUp.
	chProcess := make(chan float32)
	chErr := make(chan error)
	chDone := make(chan struct{})

	// Запуск обработчика.
	go act.BackUp(chProcess, chErr, chDone)

	// Приём признаков процесса передачи.
	done := false
	for !done {
		select {
		case <-chProcess:
			// Логика для обработки прогресса может быть добавлена здесь.

		case err = <-chErr:
			require.NoErrorf(t, err, "Ошибка процесса BackUp:<%v>", err)

		case <-chDone:
			done = true
		}
	}
}
*/

//
// --- InitDataRestore ---
//

func TestInitDataRestore(t *testing.T) {

	t.Run("Успешная инициализация", func(t *testing.T) {
		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		listFiles := []InfoByFiles{
			{
				Name:   "Foo.txt",
				Volume: 10,
			},
			{
				Name:   "Bar.txt",
				Volume: 10,
			},
		}

		err = act.InitDataRestore(listFiles)
		require.NoError(t, err)
	},
	)

	t.Run("Нет данных", func(t *testing.T) {
		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		listFiles := []InfoByFiles{}

		err = act.InitDataRestore(listFiles)
		require.Equalf(t, NotCorrectLenLestFiles, err, "Нет соответствия ошибки")
	},
	)

	t.Run("Один файл", func(t *testing.T) {
		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		listFiles := []InfoByFiles{
			{
				Name:   "Foo.txt",
				Volume: 10,
			},
		}

		err = act.InitDataRestore(listFiles)
		require.Equalf(t, NotCorrectLenLestFiles, err, "Нет соответствия ошибки")
	},
	)

	t.Run("Отрицательный размер", func(t *testing.T) {
		act, err := New("localhost", "50061")
		require.NoError(t, err)
		require.NotNil(t, act)

		listFiles := []InfoByFiles{
			{
				Name:   "Foo.txt",
				Volume: 10,
			},
			{
				Name:   "Bar.txt",
				Volume: -10,
			},
		}

		err = act.InitDataRestore(listFiles)
		require.Equalf(t, NotCorrectDataFill, err, "Нет соответствия ошибки")
	},
	)
}

//
// --- RestoreRequestFilesInfo ---
//

func TestRestoreRequestFilesInfo(t *testing.T) {

	// Мок.
	mockClient := &MockClient{
		LocalFilesInfoFunc: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.LocalFilesInfoResponse, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Формирование данных файлов.
			//
			filesInfo := make([]*pb.FileInfo, 0)
			for i := 0; i < 2; i++ {
				el := &pb.FileInfo{
					FileName: fmt.Sprintf("file-%d.txt", i),
					Size:     10,
				}
				filesInfo = append(filesInfo, el)
			}

			// Подготовка трейлера.
			trailer := metadata.Pairs(nameToken, tokens[0])

			// Ответ.
			res := &proto.LocalFilesInfoResponse{
				FileInfo: filesInfo,
			}

			// Передача трейлера.
			for _, opt := range opts {
				if tlr, ok := opt.(grpc.TrailerCallOption); ok {
					*tlr.TrailerAddr = trailer
				}
			}

			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxData, err := s.RestoreRequestFilesInfo()
	require.NoErrorf(t, err, "Ошибка:<%v>", err)
	assert.Equalf(t, 2, len(rxData), "Нет соответствия размера массива")
}

//
// --- Restore ---
//

// {...}

//
// --- ConnectClose ---
//

func TestConnectClose(t *testing.T) {

	act, err := New("localhost", "50010")
	require.NoErrorf(t, err, "Ошибка создания экземпляра")

	err = act.ConnectClose()
	assert.NoErrorf(t, err, "Ошибка:<%v>", err)
}

//
// --- SendLoginPassword ---
//

func TestSendLoginPassword(t *testing.T) {

	tokenAuth := "Foo"

	// Мок.
	mockClient := &MockClient{
		SendLoginPasswordFunc: func(ctx context.Context, in *proto.SendLoginPasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			return &emptypb.Empty{}, nil
		},
	}

	s := &server{client: mockClient}
	key := createKey()

	// Данные для тестов
	dataTest := []struct {
		nameTest      string
		loginPassword TxLoginPassword
		wantErr       error
	}{
		{
			nameTest: "Корректные данные",
			loginPassword: TxLoginPassword{
				ID:        "1",
				For:       "A",
				Login:     "B",
				Password:  "C",
				CreatedAt: "D",
			},
			wantErr: nil,
		},
		{
			nameTest: "Нет ID",
			loginPassword: TxLoginPassword{
				ID:        "",
				For:       "A",
				Login:     "B",
				Password:  "C",
				CreatedAt: "D",
			},
			wantErr: EmptyDataArgumentTxID,
		},
		{
			nameTest: "Нет For",
			loginPassword: TxLoginPassword{
				ID:        "1",
				For:       "",
				Login:     "B",
				Password:  "C",
				CreatedAt: "D",
			},
			wantErr: EmptyDataArgumentTxFor,
		},
		{
			nameTest: "Нет Login",
			loginPassword: TxLoginPassword{
				ID:        "1",
				For:       "A",
				Login:     "",
				Password:  "C",
				CreatedAt: "D",
			},
			wantErr: EmptyDataArgumentTxLogin,
		},
		{
			nameTest: "Нет Password",
			loginPassword: TxLoginPassword{
				ID:        "1",
				For:       "A",
				Login:     "B",
				Password:  "",
				CreatedAt: "D",
			},
			wantErr: EmptyDataArgumentTxPassword,
		},
		{
			nameTest: "Нет CreatedAt",
			loginPassword: TxLoginPassword{
				ID:        "1",
				For:       "A",
				Login:     "B",
				Password:  "C",
				CreatedAt: "",
			},
			wantErr: EmptyDataArgumentTxCreatedAt,
		},
	}

	// Тесты
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {
			err := s.SendLoginPassword(context.Background(), tt.loginPassword, tokenAuth, key)
			assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
		})
	}
}

//
// --- SendText ---
//

func TestSendText(t *testing.T) {

	tokenAuth := "Foo"

	// Мок.
	mockClient := &MockClient{
		SendTextFunc: func(ctx context.Context, in *proto.SendTextRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {

			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			return &emptypb.Empty{}, nil
		},
	}

	s := &server{client: mockClient}
	key := createKey()

	// Данные для тестов
	dataTest := []struct {
		nameTest      string
		loginPassword TxText
		wantErr       error
	}{
		{
			nameTest: "Корректные данные",
			loginPassword: TxText{
				ID:        "1",
				For:       "A",
				Text:      "B",
				CreatedAt: "C",
			},
			wantErr: nil,
		},
		{
			nameTest: "Нет ID",
			loginPassword: TxText{
				ID:        "",
				For:       "A",
				Text:      "B",
				CreatedAt: "C",
			},
			wantErr: EmptyDataArgumentTxID,
		},
		{
			nameTest: "Нет For",
			loginPassword: TxText{
				ID:        "1",
				For:       "",
				Text:      "B",
				CreatedAt: "C",
			},
			wantErr: EmptyDataArgumentTxFor,
		},
		{
			nameTest: "Нет Text",
			loginPassword: TxText{
				ID:        "1",
				For:       "A",
				Text:      "",
				CreatedAt: "C",
			},
			wantErr: EmptyDataArgumentTxText,
		},
		{
			nameTest: "Нет CreatedAt",
			loginPassword: TxText{
				ID:        "1",
				For:       "A",
				Text:      "B",
				CreatedAt: "",
			},
			wantErr: EmptyDataArgumentTxCreatedAt,
		},
	}

	// Тесты
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {
			err := s.SendText(context.Background(), tt.loginPassword, tokenAuth, key)
			assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
		})
	}
}

//
// --- SendBankCard ---
//

func TestSendBankCard(t *testing.T) {

	tokenAuth := "Foo"

	// Мок.
	mockClient := &MockClient{
		SendBankCardFunc: func(ctx context.Context, in *proto.SendBankCardRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			return &emptypb.Empty{}, nil
		},
	}

	s := &server{client: mockClient}
	key := createKey()

	// Данные для тестов
	dataTest := []struct {
		nameTest      string
		loginPassword TxBankCard
		wantErr       error
	}{
		{
			nameTest: "Корректные данные",
			loginPassword: TxBankCard{
				ID:        "1",
				For:       "A",
				Owner:     "B",
				Numb:      "4532015112830366",
				ValidData: "C",
				Code:      "D",
				CreatedAt: "E",
			},
			wantErr: nil,
		},
		{
			nameTest: "Нет ID",
			loginPassword: TxBankCard{
				ID:        "",
				For:       "A",
				Owner:     "B",
				Numb:      "4532015112830366",
				ValidData: "C",
				Code:      "D",
				CreatedAt: "E",
			},
			wantErr: EmptyDataArgumentTxID,
		},
		{
			nameTest: "Нет For",
			loginPassword: TxBankCard{
				ID:        "1",
				For:       "",
				Owner:     "B",
				Numb:      "4532015112830366",
				ValidData: "C",
				Code:      "D",
				CreatedAt: "E",
			},
			wantErr: EmptyDataArgumentTxFor,
		},
		{
			nameTest: "Нет Owner",
			loginPassword: TxBankCard{
				ID:        "1",
				For:       "A",
				Owner:     "",
				Numb:      "4532015112830366",
				ValidData: "C",
				Code:      "D",
				CreatedAt: "E",
			},
			wantErr: EmptyDataArgumentTxOwner,
		},
		{
			nameTest: "Нет Numb",
			loginPassword: TxBankCard{
				ID:        "1",
				For:       "A",
				Owner:     "B",
				Numb:      "",
				ValidData: "C",
				Code:      "D",
				CreatedAt: "E",
			},
			wantErr: EmptyDataArgumentTxNumb,
		},
		{
			nameTest: "Нет ValidData",
			loginPassword: TxBankCard{
				ID:        "1",
				For:       "A",
				Owner:     "B",
				Numb:      "4532015112830366",
				ValidData: "",
				Code:      "D",
				CreatedAt: "E",
			},
			wantErr: EmptyDataArgumentTxValidData,
		},
		{
			nameTest: "Нет Code",
			loginPassword: TxBankCard{
				ID:        "1",
				For:       "A",
				Owner:     "B",
				Numb:      "4532015112830366",
				ValidData: "C",
				Code:      "",
				CreatedAt: "E",
			},
			wantErr: EmptyDataArgumentTxCode,
		},
		{
			nameTest: "Нет CreatedAt",
			loginPassword: TxBankCard{
				ID:        "1",
				For:       "A",
				Owner:     "B",
				Numb:      "4532015112830366",
				ValidData: "C",
				Code:      "D",
				CreatedAt: "",
			},
			wantErr: EmptyDataArgumentTxCreatedAt,
		},
		{
			nameTest: "Ошибка в Numb",
			loginPassword: TxBankCard{
				ID:        "1",
				For:       "A",
				Owner:     "B",
				Numb:      "4532015112830367",
				ValidData: "C",
				Code:      "D",
				CreatedAt: "E",
			},
			wantErr: NotCorrectDataNumb,
		},
	}

	// Тесты
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {
			err := s.SendBankCard(context.Background(), tt.loginPassword, tokenAuth, key)
			assert.Equalf(t, tt.wantErr, err, "Нет соответствия ошибки")
		})
	}
}

//
// --- SendFile ---
//

// {...}

//
// --- RequestLoginPasswordNames ---
//

func TestRequestLoginPasswordNames(t *testing.T) {

	tokenAuth := "ValidToken"
	names := []string{"Foo", "Bar"}
	key := createKey()

	// Мок.
	mockClient := &MockClient{
		RequestLoginPasswordNameFunc: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestLoginPasswordNameResponse, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Шифрование, чтобы клиент мог расшифровать.
			encrNames := make([]string, 0)
			for _, v := range names {
				encStr, err := encrypt(v, key)
				require.NoErrorf(t, err, "ошибка шифрования")

				encrNames = append(encrNames, encStr)
			}

			// Ответ.
			res := &proto.RequestLoginPasswordNameResponse{
				EntriesName: encrNames,
			}

			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxData, err := s.RequestLoginPasswordNames(context.Background(), tokenAuth, key)
	require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)
	assert.Equalf(t, names, rxData, "Нет соответствия данных")
}

//
// --- RequestLoginPasswordByName ---
//

func TestRequestLoginPasswordByName(t *testing.T) {

	nameEntry := "nameEntry"
	tokenAuth := "ValidToken"
	name := "A"
	login := "B"
	password := "C"
	createdAt := "D"
	key := createKey()

	// Мок.
	mockClient := &MockClient{
		RequestLoginPasswordByNameFunc: func(ctx context.Context, in *proto.RequestLoginPasswordByNameRequest, opts ...grpc.CallOption) (*proto.RequestLoginPasswordByNameResponse, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Проверка принятых данных.
			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия токена")

			encNameEntry, err := encrypt(nameEntry, key)
			require.NoErrorf(t, err, "Ошибка шифрования nameEntry")
			require.Equalf(t, encNameEntry, in.Name, "Нет соответствия имён")

			// Шифрование, чтобы клиент мог расшифровать.
			encName, err := encrypt(name, key)
			require.NoErrorf(t, err, "Ошибка шифрования name")

			encLogin, err := encrypt(login, key)
			require.NoErrorf(t, err, "Ошибка шифрования login")

			encPassword, err := encrypt(password, key)
			require.NoErrorf(t, err, "Ошибка шифрования password")

			encCreatedAt, err := encrypt(createdAt, key)
			require.NoErrorf(t, err, "Ошибка шифрования createdAt")

			// Ответ.
			res := &proto.RequestLoginPasswordByNameResponse{
				Name:      encName,
				Login:     encLogin,
				Password:  encPassword,
				CreatedAt: encCreatedAt,
			}
			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxData, err := s.RequestLoginPasswordByName(context.Background(), tokenAuth, "idClient", nameEntry, key)
	require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

	assert.Equalf(t, name, rxData.For, "Нет соответствия Name")
	assert.Equalf(t, login, rxData.Login, "Нет соответствия Login")
	assert.Equalf(t, password, rxData.Password, "Нет соответствия Password")
	assert.Equalf(t, createdAt, rxData.CreatedAt, "Нет соответствия CreatedAt")
}

//
// --- RequestTextNames ---
//

func TestRequestTextNames(t *testing.T) {

	tokenAuth := "ValidToken"
	names := []string{"Foo", "Bar"}
	key := createKey()

	// Мок.
	mockClient := &MockClient{
		RequestTextNameFunc: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestTextNameResponse, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия токенов")

			// Шифрование, чтобы клиент мог расшифровать.
			encrNames := make([]string, 0)
			for _, v := range names {
				encStr, err := encrypt(v, key)
				require.NoErrorf(t, err, "ошибка шифрования")

				encrNames = append(encrNames, encStr)
			}

			// Ответ.
			res := &proto.RequestTextNameResponse{
				EntriesName: encrNames,
			}

			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxData, err := s.RequestTextNames(context.Background(), tokenAuth, key)
	require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)
	assert.Equalf(t, names, rxData, "Нет соответствия данных")
}

//
// --- RequestTextByName ---
//

func TestRequestTextByName(t *testing.T) {

	nameEntry := "nameEntry"
	tokenAuth := "ValidToken"
	name := "A"
	text := "B"
	createdAt := "C"
	key := createKey()

	// Мок.
	mockClient := &MockClient{
		RequestTextByNameFunc: func(ctx context.Context, in *proto.RequestTextByNameRequest, opts ...grpc.CallOption) (*proto.RequestTextByNameResponse, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Проверка принятых данных.
			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия токенов")

			encNameEntry, err := encrypt(nameEntry, key)
			require.NoErrorf(t, err, "ошибка шифрования nameEntry")
			require.Equalf(t, encNameEntry, in.Name, "нет соответствия имёни записи")

			// Шифрование, чтобы клиент мог расшифровать.
			encName, err := encrypt(name, key)
			require.NoErrorf(t, err, "ошибка шифрования name")

			encText, err := encrypt(text, key)
			require.NoErrorf(t, err, "ошибка шифрования text")

			encCreatedAt, err := encrypt(createdAt, key)
			require.NoErrorf(t, err, "ошибка шифрования createdAt")

			// Ответ.
			res := &proto.RequestTextByNameResponse{
				Name:      encName,
				Text:      encText,
				CreatedAt: encCreatedAt,
			}
			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxData, err := s.RequestTextByName(context.Background(), tokenAuth, "idClient", "nameEntry", key)
	require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

	assert.Equalf(t, name, rxData.For, "Нет соответствия Name")
	assert.Equalf(t, text, rxData.Text, "Нет соответствия Text")
	assert.Equalf(t, createdAt, rxData.CreatedAt, "Нет соответствия CreatedAt")
}

//
// --- RequestBankCardNames ---
//

func TestRequestBankCardNames(t *testing.T) {

	tokenAuth := "ValidToken"
	names := []string{"Foo", "Bar"}
	key := createKey()

	// Мок.
	mockClient := &MockClient{
		RequestBankCardNameFunc: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestBankCardNameResponse, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Обработка входных данных.
			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия токенов")

			// Шифрование, чтобы клиент мог расшифровать.
			encrNames := make([]string, 0)
			for _, v := range names {
				encStr, err := encrypt(v, key)
				require.NoErrorf(t, err, "ошибка шифрования")

				encrNames = append(encrNames, encStr)
			}

			// Ответ.
			res := &proto.RequestBankCardNameResponse{
				EntriesName: encrNames,
			}

			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxData, err := s.RequestBankCardNames(context.Background(), tokenAuth, key)
	require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)
	assert.Equalf(t, names, rxData, "Нет соответствия данных")
}

//
// --- RequestBankCardByName ---
//

func TestRequestBankCardByName(t *testing.T) {

	nameEntry := "nameEntry"
	tokenAuth := "ValidToken"
	name := "A"
	owner := "B"
	numb := "C"
	valid := "D"
	code := "E"
	createdAt := "F"
	key := createKey()

	// Мок.
	mockClient := &MockClient{
		RequestBankCardByNameFunc: func(ctx context.Context, in *proto.RequestBankCardByNameRequest, opts ...grpc.CallOption) (*proto.RequestBankCardByNameResponse, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "В запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Обработка принятых данных
			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия токенов")

			encNameEntry, err := encrypt(nameEntry, key)
			require.NoErrorf(t, err, "ошибка шифрования nameEntry")
			require.Equalf(t, encNameEntry, in.Name, "нет соответствия имён")

			// Шифрование, чтобы клиент мог расшифровать.
			encName, err := encrypt(name, key)
			require.Equalf(t, encNameEntry, in.Name, "ошибка шифрования name")

			encOwner, err := encrypt(owner, key)
			require.Equalf(t, encNameEntry, in.Name, "ошибка шифрования owner")

			encNumb, err := encrypt(numb, key)
			require.Equalf(t, encNameEntry, in.Name, "ошибка шифрования numb")

			encValid, err := encrypt(valid, key)
			require.Equalf(t, encNameEntry, in.Name, "ошибка шифрования valid")

			encCode, err := encrypt(code, key)
			require.Equalf(t, encNameEntry, in.Name, "ошибка шифрования code")

			encCreatedAt, err := encrypt(createdAt, key)
			require.Equalf(t, encNameEntry, in.Name, "ошибка шифрования createdAt")

			// Ответ.
			res := &proto.RequestBankCardByNameResponse{
				Name:      encName,
				Owner:     encOwner,
				Numb:      encNumb,
				Valid:     encValid,
				Code:      encCode,
				CreatedAt: encCreatedAt,
			}
			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxData, err := s.RequestBankCardByName(context.Background(), tokenAuth, "idClient", "nameEntry", key)
	require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

	assert.Equalf(t, name, rxData.For, "Нет соответствия Name")
	assert.Equalf(t, owner, rxData.Owner, "Нет соответствия Owner")
	assert.Equalf(t, numb, rxData.Numb, "Нет соответствия Numb")
	assert.Equalf(t, valid, rxData.Valid, "Нет соответствия Valid")
	assert.Equalf(t, code, rxData.Code, "Нет соответствия Code")
	assert.Equalf(t, createdAt, rxData.CreatedAt, "Нет соответствия CreatedAt")
}

//
// --- RequestFileNames ---
//

func TestRequestFileNames(t *testing.T) {

	tokenAuth := "ValidToken"
	names := []string{"Foo.txt", "Bar.txt"}
	key := createKey()

	// Мок.
	mockClient := &MockClient{
		RequestFileNameFunc: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestFileNameResponse, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "в запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Обработка входных данных.
			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия токенов")

			// Ответ.
			res := &proto.RequestFileNameResponse{
				EntriesName: names,
			}

			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxData, _, err := s.RequestFileNames(context.Background(), tokenAuth, key)
	require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)
	assert.Equalf(t, names, rxData, "Нет соответствия данных")
}

//
// --- RequestFileInfo ---
//

func TestRequestFileInfo(t *testing.T) {

	tokenAuth := "ValidToken"
	name := "A"
	hash := "B"
	size := int64(10)

	// Мок.
	mockClient := &MockClient{
		RequestFileInfoFunc: func(ctx context.Context, in *proto.RequestFileInfoRequest, opts ...grpc.CallOption) (*proto.RequestFileInfoResponse, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "В запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Обработка принятых данных
			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия токенов")

			// Ответ.
			res := &proto.RequestFileInfoResponse{
				Name: name,
				Hash: hash,
				Size: int64(size),
			}
			return res, nil
		},
	}

	s := &server{client: mockClient}

	rxName, rxHash, rxSize, err := s.RequestFileInfo(tokenAuth, "idClient", name)
	require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

	assert.Equalf(t, name, rxName, "Нет соответствия Name")
	assert.Equalf(t, hash, rxHash, "Нет соответствия Hash")
	assert.Equalf(t, size, rxSize, "Нет соответствия Size")
}

//
// --- RequestFileByName ---
//

// {...}

//
// --- InitDataRequestFileByName ---
//

func TestInitDataRequestFileByName(t *testing.T) {

	resetInstServer()
	key := createKey()

	act, err := New("", "")
	require.NoErrorf(t, err, "ошибка конструктора")

	defer resetInstServer()

	dataTest := []struct {
		nameTest string
		data     DataRequestFile
		wantErr  error
	}{
		{
			nameTest: "Корректные данные",
			data: DataRequestFile{
				FilePath:    "./A",
				FileName:    "A",
				TokenAuth:   "B",
				ClientID:    "C",
				SizeReqFile: 10,
				SizePassed:  0,
				SecretKey:   key,
			},
			wantErr: nil,
		},
		{
			nameTest: "Нет пути",
			data: DataRequestFile{
				FilePath:    "",
				FileName:    "A",
				TokenAuth:   "B",
				ClientID:    "C",
				SizeReqFile: 10,
				SizePassed:  0,
				SecretKey:   key,
			},
			wantErr: nil,
		},
		{
			nameTest: "Нет имени",
			data: DataRequestFile{
				FilePath:    "./A",
				FileName:    "",
				TokenAuth:   "B",
				ClientID:    "C",
				SizeReqFile: 10,
				SizePassed:  0,
				SecretKey:   key,
			},
			wantErr: EmptyDataArgumentName,
		},
		{
			nameTest: "Нет токена",
			data: DataRequestFile{
				FilePath:    "./A",
				FileName:    "A",
				TokenAuth:   "",
				ClientID:    "C",
				SizeReqFile: 10,
				SizePassed:  0,
				SecretKey:   key,
			},
			wantErr: EmptyDataArgumentTokenAuth,
		},
		{
			nameTest: "Нет ID клиента",
			data: DataRequestFile{
				FilePath:    "./A",
				FileName:    "A",
				TokenAuth:   "B",
				ClientID:    "",
				SizeReqFile: 10,
				SizePassed:  0,
				SecretKey:   key,
			},
			wantErr: EmptyDataArgumentClientID,
		},
		{
			nameTest: "Отрицательный размера файла",
			data: DataRequestFile{
				FilePath:    "./A",
				FileName:    "A",
				TokenAuth:   "B",
				ClientID:    "C",
				SizeReqFile: -1,
				SizePassed:  0,
				SecretKey:   key,
			},
			wantErr: IncorrectSizeFile,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.nameTest, func(t *testing.T) {

			err = act.InitDataRequestFileByName(tt.data)
			require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
		})
	}
}

//
// --- DeleteLoginPassword ---
//

func TestDeleteLoginPassword(t *testing.T) {

	tokenAuth := "Foo"
	name := ""
	idClient := ""

	// Мок.
	mockClient := &MockClient{
		DeleteLoginPasswordFunc: func(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "В запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Обработка принятых данных
			require.Equalf(t, name, in.Name, "нет соответствия name")
			require.Equalf(t, idClient, in.IdClient, "нет соответствия idClient")
			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия tokenAuth")

			return &emptypb.Empty{}, nil
		},
	}

	s := &server{
		client:   mockClient,
		tokenSrv: tokenAuth,
	}

	// Данные тестов
	dataTest := []struct {
		nameTest string
		idClient string
		name     string
		wantErr  error
	}{
		{
			nameTest: "Корректные данные",
			idClient: "A",
			name:     "B",
			wantErr:  nil,
		},
		{
			nameTest: "Нет ID клиента",
			idClient: "",
			name:     "B",
			wantErr:  EmptyDataArgumentClientID,
		},
		{
			nameTest: "Нет имени",
			idClient: "A",
			name:     "",
			wantErr:  EmptyDataArgumentName,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.name, func(t *testing.T) {

			name = tt.name
			idClient = tt.idClient

			err := s.DeleteLoginPassword(tt.idClient, tt.name)
			require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
		})
	}
}

//
// --- DeleteText ---
//

func TestDeleteText(t *testing.T) {

	tokenAuth := "Foo"
	name := ""
	idClient := ""

	// Мок.
	mockClient := &MockClient{
		DeleteTextFunc: func(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "В запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Обработка принятых данных
			require.Equalf(t, name, in.Name, "нет соответствия name")
			require.Equalf(t, idClient, in.IdClient, "нет соответствия idClient")
			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия tokenAuth")

			return &emptypb.Empty{}, nil
		},
	}

	s := &server{
		client:   mockClient,
		tokenSrv: tokenAuth,
	}

	// Данные тестов
	dataTest := []struct {
		nameTest string
		idClient string
		name     string
		wantErr  error
	}{
		{
			nameTest: "Корректные данные",
			idClient: "A",
			name:     "B",
			wantErr:  nil,
		},
		{
			nameTest: "Нет ID клиента",
			idClient: "",
			name:     "B",
			wantErr:  EmptyDataArgumentClientID,
		},
		{
			nameTest: "Нет имени",
			idClient: "A",
			name:     "",
			wantErr:  EmptyDataArgumentName,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.name, func(t *testing.T) {

			name = tt.name
			idClient = tt.idClient

			err := s.DeleteText(tt.idClient, tt.name)
			require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
		})
	}
}

//
// --- DeleteBankCard ---
//

func TestDeleteBankCard(t *testing.T) {

	tokenAuth := "Foo"
	name := ""
	idClient := ""

	// Мок.
	mockClient := &MockClient{
		DeleteBankCardFunc: func(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {

			// Приём данных запроса.
			//
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			require.Truef(t, ok, "В запросе нет метаданных")

			nameToken := "token"
			tokens := rxMD[nameToken]
			require.Falsef(t, len(tokens) == 0 || tokens[0] == "", "в запросе нет токена <token>")

			// Обработка принятых данных
			require.Equalf(t, name, in.Name, "нет соответствия name")
			require.Equalf(t, idClient, in.IdClient, "нет соответствия idClient")
			require.Equalf(t, tokenAuth, tokens[0], "нет соответствия tokenAuth")

			return &emptypb.Empty{}, nil
		},
	}

	s := &server{
		client:   mockClient,
		tokenSrv: tokenAuth,
	}

	// Данные тестов
	dataTest := []struct {
		nameTest string
		idClient string
		name     string
		wantErr  error
	}{
		{
			nameTest: "Корректные данные",
			idClient: "A",
			name:     "B",
			wantErr:  nil,
		},
		{
			nameTest: "Нет ID клиента",
			idClient: "",
			name:     "B",
			wantErr:  EmptyDataArgumentClientID,
		},
		{
			nameTest: "Нет имени",
			idClient: "A",
			name:     "",
			wantErr:  EmptyDataArgumentName,
		},
	}

	// Тесты.
	for _, tt := range dataTest {
		t.Run(tt.name, func(t *testing.T) {

			name = tt.name
			idClient = tt.idClient

			err := s.DeleteBankCard(tt.idClient, tt.name)
			require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
		})
	}
}

//
// --- GetTokenAuthentication ---
//

func TestGetTokenAuthentication(t *testing.T) {

	tokenAuth := "Foo"

	s := &server{
		tokenSrv: tokenAuth,
	}

	rxToken := s.GetTokenAuthentication()

	require.Equalf(t, tokenAuth, rxToken, "Нет соответствия")
}

//
// --- UpdateTokenAuthentication ---
//

func TestUpdateTokenAuthentication(t *testing.T) {

	tokenAuth := "Foo"
	newTokenAuth := "Bar"

	s := &server{
		tokenSrv: tokenAuth,
	}

	rxToken := s.GetTokenAuthentication()
	require.Equalf(t, tokenAuth, rxToken, "Нет соответствия")

	s.UpdateTokenAuthentication(newTokenAuth)

	rxToken = s.GetTokenAuthentication()
	require.Equalf(t, newTokenAuth, rxToken, "Нет соответствия, после обновления.")

}

//
// --- IsConnectSuccess ---
//

func TestIsConnectSuccess(t *testing.T) {

	t.Run("Нет подключения", func(t *testing.T) {
		act, err := New("", "")
		require.NoErrorf(t, err, "Ошибка конструктора")

		defer resetInstServer()

		connStatus := act.IsConnectSuccess()
		require.Falsef(t, connStatus, "Ошибка логики")
	})

	t.Run("Есть подключение", func(t *testing.T) {
		act, err := New("localhost", "50002")
		require.NoErrorf(t, err, "Ошибка конструктора")

		defer resetInstServer()

		connStatus := act.IsConnectSuccess()
		require.Truef(t, connStatus, "Ошибка логики")
	})
}

// -------------------------------------------------------------------
//
//                        Вспомогательные функции
//
// -------------------------------------------------------------------

// Создание ключа. Возвращается ключ.
func createKey() (secretKey [32]byte) {

	copy(secretKey[:], "FooBar")
	return secretKey
}
