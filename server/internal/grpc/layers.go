package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//
// --- Registration ---
//

// Получение токена из запроса.
func layerRegistrationGetToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 || tokens[0] == "" {
		return tokenData{}, ErrMissingToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Слой приёма данных. Возвращаются принятые данные и ошибка.
func layerRegistrationRx(req *pb.RegistrationRequest) (rxData registrationRX, err error) {

	rxData.userName = req.UserName
	rxData.userPwd = req.UserPwd
	rxData.userPwdRepeat = req.UserPwdRepeat

	// Проверка данных.
	if rxData.userName == "" || rxData.userPwd == "" || rxData.userPwdRepeat == "" {
		return registrationRX{}, ErrNotCorrectData
	}
	if rxData.userPwd != rxData.userPwdRepeat {
		return registrationRX{}, ErrNotCorrectData
	}

	return rxData, nil
}

// Логика обработчика.
func layerRegistrationLogic(rxData registrationRX, db domain.StorageI) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.AddUserContext(ctx, rxData.userName, rxData.userPwd); err != nil {
		return fmt.Errorf("Функция db.AddUserContext, вернула ошибку: <%w>", err)
	}

	return nil
}

// Ответ. Возврат приянтого токена.
func layerRegistrationTx(ctx context.Context, token tokenData) error {

	txMD := metadata.Pairs(token.name, token.token)

	if err := grpc.SendHeader(ctx, txMD); err != nil {
		return fmt.Errorf("функция grpc.SendHeader, вернула ошибку:<%w>", err)
	}

	return nil
}

//
// --- Authentication ---
//

// Слой приёма данных. Возвращаются принятые данные и ошибка.
func layerAuthenticationRx(req *pb.AuthenticationRequest) (rxData authenticationRX, err error) {

	rxData.userName = req.UserName
	rxData.userPwd = req.UserPwd

	// Проверка данных.
	if rxData.userName == "" || rxData.userPwd == "" {
		return authenticationRX{}, ErrNotCorrectData
	}

	return rxData, nil
}

// Получение токена из запроса.
func layerAuthenticationGetToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 || tokens[0] == "" {
		return tokenData{}, ErrMissingToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Логика аутентификации.
func layerAuthenticationLogic(s *Manager, userName, userPwd string) error {
	// Контекст для запроса.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Выполнение запроса.
	ok, err := s.storage.AuthenticateUserContext(ctx, userName, userPwd)
	if err != nil {
		return fmt.Errorf("Функция AuthenticateUserContext, вернуля ошибку:<%w>", err)
	}
	if !ok {
		return fmt.Errorf("Пользователь:<%s>, не прошел аутентификацию", userName)
	}
	s.logger.Info("Пользователь прошел аутентификацию", zap.String("имя", userName))

	return nil
}

// Ответ. Возврат приянтого токена.
func layerAuthenticationTx(ctx context.Context, rxToken tokenData, srvToken string) (res *pb.AuthenticationResponse, err error) {

	// Возврат токена, принятого от клиента.
	txMD := metadata.Pairs(rxToken.name, rxToken.token)

	if err := grpc.SendHeader(ctx, txMD); err != nil {
		return nil, fmt.Errorf("функция grpc.SendHeader, вернула ошибку:<%w>", err)
	}

	// Передача токена сервера.
	res = &proto.AuthenticationResponse{
		Token: srvToken,
	}

	return res, nil
}
