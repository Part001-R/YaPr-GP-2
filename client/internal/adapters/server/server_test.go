package server

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockClient struct {
	AuthenticationFunc func(ctx context.Context, req *proto.AuthenticationRequest, opts ...grpc.CallOption) (*proto.AuthenticationResponse, error)
}

func (t *MockClient) Authentication(ctx context.Context, in *proto.AuthenticationRequest, opts ...grpc.CallOption) (*proto.AuthenticationResponse, error) {
	return t.AuthenticationFunc(ctx, in, opts...)
}

func (t *MockClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *MockClient) LocalBackupFile(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.LocalBackupFileRequest, proto.LocalBackupFileResponse], error) {
	return nil, nil
}

func (m *MockClient) LocalRestoreFile(ctx context.Context, in *proto.LocalRestoreFileRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.LocalRestoreFileResponse], error) {
	return nil, nil
}

func (m *MockClient) LocalFilesInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.LocalFilesInfoResponse, error) {
	return nil, nil
}

func (m *MockClient) Registration(ctx context.Context, in *proto.RegistrationRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *MockClient) SendLoginPassword(ctx context.Context, in *proto.SendLoginPasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *MockClient) SendText(ctx context.Context, in *proto.SendTextRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *MockClient) SendBankCard(ctx context.Context, in *proto.SendBankCardRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *MockClient) SendFile(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[proto.SendFileRequest, proto.SendFileResponse], error) {
	return nil, nil
}

func (m *MockClient) RequestLoginPasswordName(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestLoginPasswordNameResponse, error) {
	return nil, nil
}

func (m *MockClient) RequestLoginPasswordByName(ctx context.Context, in *proto.RequestLoginPasswordByNameRequest, opts ...grpc.CallOption) (*proto.RequestLoginPasswordByNameResponse, error) {
	return nil, nil
}

func (m *MockClient) RequestTextName(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestTextNameResponse, error) {
	return nil, nil
}

func (m *MockClient) RequestTextByName(ctx context.Context, in *proto.RequestTextByNameRequest, opts ...grpc.CallOption) (*proto.RequestTextByNameResponse, error) {
	return nil, nil
}

func (m *MockClient) RequestBankCardName(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestBankCardNameResponse, error) {
	return nil, nil
}

func (m *MockClient) RequestBankCardByName(ctx context.Context, in *proto.RequestBankCardByNameRequest, opts ...grpc.CallOption) (*proto.RequestBankCardByNameResponse, error) {
	return nil, nil
}

func (m *MockClient) RequestFileInfo(ctx context.Context, in *proto.RequestFileInfoRequest, opts ...grpc.CallOption) (*proto.RequestFileInfoResponse, error) {
	return nil, nil
}

func (m *MockClient) RequestFileName(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.RequestFileNameResponse, error) {
	return nil, nil
}

func (m *MockClient) RequestFileByName(ctx context.Context, in *proto.RequestFileByNameRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.RequestFileByNameResponse], error) {
	return nil, nil
}

func (m *MockClient) DeleteLoginPassword(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *MockClient) DeleteText(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *MockClient) DeleteBankCard(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (m *MockClient) DeleteFile(ctx context.Context, in *proto.RequestDeleteName, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

//
// --- AuthenticationContext ---
//

func TestAuthenticationContext_SUCCESS(t *testing.T) {

	tokenAuth := "Foo"

	// Инициализация мок-клиента с имитацией grpc.SendHeader
	mockClient := &MockClient{
		AuthenticationFunc: func(ctx context.Context, req *proto.AuthenticationRequest, opts ...grpc.CallOption) (*proto.AuthenticationResponse, error) {

			// Входящие метаданные
			rxMD, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				return nil, status.Error(codes.PermissionDenied, "в запросе нет метаданных")
			}
			nameToken := "token"
			tokens := rxMD[nameToken]
			if len(tokens) == 0 || tokens[0] == "" {
				return nil, status.Error(codes.PermissionDenied, "в запросе нет токена <token>")
			}

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
			if !ok {
				return nil, status.Error(codes.PermissionDenied, "в запросе нет метаданных")
			}
			nameToken := "token"
			tokens := rxMD[nameToken]
			if len(tokens) == 0 || tokens[0] == "" {
				return nil, status.Error(codes.PermissionDenied, "в запросе нет токена <token>")
			}

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

// --- New ---
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
