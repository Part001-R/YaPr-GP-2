// Тесты бработчиков.
package ui

import (
	"context"

	"github.com/Part001-R/YaPr-GP-2/proto"
	"google.golang.org/grpc"
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
// --- new ---
//
