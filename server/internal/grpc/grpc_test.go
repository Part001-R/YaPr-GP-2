// Тесты пакета.
package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Мок.
type MockServer struct {
	mock.Mock
	proto.UnimplementedPasswordManagerServer
}

// Аутентификация.
func (m *MockServer) Authentication(ctx context.Context, req *proto.AuthenticationRequest) (*proto.AuthenticationResponse, error) {

	if req.UserName == "" || req.UserPwd == "" {
		return &proto.AuthenticationResponse{}, status.Error(codes.PermissionDenied, "ошибка в данных")
	}
	return &pb.AuthenticationResponse{Token: "tokenAuth"}, nil
}

// Проверка связи.
func (m *MockServer) Ping(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Резервное копирование (режим - локальный).
func (m *MockServer) LocalBackupFile(stream proto.PasswordManager_LocalBackupFileServer) error {
	args := m.Called(stream)
	return args.Error(0)
}

// Восстановление из резервной копии (режим локальный).
func (m *MockServer) LocalRestoreFile(req *proto.LocalRestoreFileRequest, stream grpc.ServerStreamingServer[proto.LocalRestoreFileResponse]) error {

	response := &proto.LocalRestoreFileResponse{
		FileName: req.FileName,
		Content:  []byte("..."),
	}

	if err := stream.SendMsg(response); err != nil {
		return err
	}

	return nil
}

// Информация по файлам.
func (m *MockServer) LocalFilesInfo(ctx context.Context, req *emptypb.Empty) (*proto.LocalFilesInfoResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*proto.LocalFilesInfoResponse), args.Error(1)
}

// Регистрация.
func (m *MockServer) Registration(ctx context.Context, in *proto.RegistrationRequest) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Передача данных логин/пароль.
func (m *MockServer) SendLoginPassword(ctx context.Context, in *proto.SendLoginPasswordRequest) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Передача текста.
func (m *MockServer) SendText(ctx context.Context, in *proto.SendTextRequest) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Передача банковской карты.
func (m *MockServer) SendBankCard(ctx context.Context, in *proto.SendBankCardRequest) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Передача файла.
func (m *MockServer) SendFile(stream grpc.ClientStreamingServer[proto.SendFileRequest, proto.SendFileResponse]) error {

	return nil
}

// Запрос имен записей логин/пароль.
func (m *MockServer) RequestLoginPasswordName(ctx context.Context, in *emptypb.Empty) (*proto.RequestLoginPasswordNameResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RequestLoginPasswordNameResponse), args.Error(1)
}

// Запрос данных логин/пароль по имени записи.
func (m *MockServer) RequestLoginPasswordByName(ctx context.Context, in *proto.RequestLoginPasswordByNameRequest) (*proto.RequestLoginPasswordByNameResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RequestLoginPasswordByNameResponse), args.Error(1)
}

// / Запрос имен записей данных текста.
func (m *MockServer) RequestTextName(ctx context.Context, in *emptypb.Empty) (*proto.RequestTextNameResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RequestTextNameResponse), args.Error(1)
}

// Запрос данных текста по имени записи.
func (m *MockServer) RequestTextByName(ctx context.Context, in *proto.RequestTextByNameRequest) (*proto.RequestTextByNameResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RequestTextByNameResponse), args.Error(1)
}

// Запрос имен записей банковских карт.
func (m *MockServer) RequestBankCardName(ctx context.Context, in *emptypb.Empty) (*proto.RequestBankCardNameResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RequestBankCardNameResponse), args.Error(1)
}

// Запрос данных банковской карты по имени записи.
func (m *MockServer) RequestBankCardByName(ctx context.Context, in *proto.RequestBankCardByNameRequest) (*proto.RequestBankCardByNameResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RequestBankCardByNameResponse), args.Error(1)
}

// Запрос информации по файлу.
func (m *MockServer) RequestFileInfo(ctx context.Context, in *proto.RequestFileInfoRequest) (*proto.RequestFileInfoResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RequestFileInfoResponse), args.Error(1)
}

// Запрос имен файлов.
func (m *MockServer) RequestFileName(ctx context.Context, in *emptypb.Empty) (*proto.RequestFileNameResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RequestFileNameResponse), args.Error(1)
}

// Запрос файла по имени.
func (m *MockServer) RequestFileByName(in *proto.RequestFileByNameRequest, stream grpc.ServerStreamingServer[proto.RequestFileByNameResponse]) error {

	args := m.Called(in, stream)

	if err := args.Error(0); err != nil {
		return err
	}

	response := &proto.RequestFileByNameResponse{
		// ...
	}

	if err := stream.SendMsg(response); err != nil {
		return err
	}

	return nil
}

// Удаление записи данных логин/пароль.
func (m *MockServer) DeleteLoginPassword(ctx context.Context, in *proto.RequestDeleteName) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Удаление записи текста.
func (m *MockServer) DeleteText(ctx context.Context, in *proto.RequestDeleteName) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Удаление записи банковской карты.
func (m *MockServer) DeleteBankCard(ctx context.Context, in *proto.RequestDeleteName) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// Удаление файла.
func (m *MockServer) DeleteFile(ctx context.Context, in *proto.RequestDeleteName) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

//
// --- Authentication ---
//

func TestAuthentication(t *testing.T) {

	mockServer := new(MockServer)

	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	proto.RegisterPasswordManagerServer(server, mockServer)

	// Запуск сервера.
	go func() {
		if err := server.Serve(lis); err != nil {
			t.Errorf("Server failed to serve: %v", err)
		}
	}()

	defer server.Stop()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	clientConn, err := grpc.DialContext(context.Background(), "", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	assert.NoError(t, err)
	defer clientConn.Close()

	// Клиент.
	client := proto.NewPasswordManagerClient(clientConn)

	// Создание запроса.
	req := &proto.AuthenticationRequest{
		UserName: "Foo",
		UserPwd:  "Bar",
	}

	// Запрос.
	resp, err := client.Authentication(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equalf(t, "tokenAuth", resp.Token, "Нет соответствия токена аутентификации")
}
