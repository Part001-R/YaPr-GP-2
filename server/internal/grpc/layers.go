// Слои обработчиков.
package grpc

import (
	"context"
	"fmt"
	"os"
	"path"
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

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
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
//
// Параметры:
//
//	req - запрос.
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

// Логика обработчика. Возвращается ошибка.
//
// Параметры:
//
//	rxData - принятые данные.
//	domain - указатель домена.
func layerRegistrationLogic(rxData registrationRX, domain domain.DomainI) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := domain.AddUserContext(ctx, rxData.userName, rxData.userPwd); err != nil {
		return fmt.Errorf("Функция db.AddUserContext, вернула ошибку: <%w>", err)
	}

	return nil
}

// Ответ. Возврат приянтого токена. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	token - токен.
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
//
// Параметры:
//
//	req - запрос.
func layerAuthenticationRx(req *pb.AuthenticationRequest) (rxData authenticationRX, err error) {

	rxData.userName = req.UserName
	rxData.userPwd = req.UserPwd

	// Проверка данных.
	if rxData.userName == "" || rxData.userPwd == "" {
		return authenticationRX{}, ErrNotCorrectData
	}

	return rxData, nil
}

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
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

// Логика. Возвращается ошибка.
//
// Параметры:
//
//	s - указатель на экземпляр сервиса.
//	userName - имя пользователя.
//	userPwd - пароль пользователя.
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

// Формирование ответа. Возвращается ответ и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	rxToken - принятый токен.
//	srvToken - токен сервера.
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

//
// --- SendLoginPassword ---
//

// Получение данных запроса. Возвращается принятые данные и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerSendLoginPasswordRx(req *pb.SendLoginPasswordRequest) (rxData RxLoginPassword, err error) {

	rxData.ID = req.IdClient
	rxData.For = req.For
	rxData.Login = req.Login
	rxData.Password = req.Password
	rxData.CreatedAt = req.CreatedAt

	return rxData, nil
}

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerSendLoginPasswordGetToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Логика. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	rxData - принятые данные.
//	m - указатель на экземпляр сервиса.
func layerSendLoginPasswordLogicContext(ctx context.Context, rxData RxLoginPassword, m *Manager) error {

	if err := m.storage.AddDataLoginPasswordContext(ctx, rxData.For, rxData.Login, rxData.Password, rxData.CreatedAt); err != nil {
		return fmt.Errorf("Функция AddDataLoginPasswordContext, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- SendText ---
//

// Получение данных запроса. Возвращаются данные и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerSendTextRx(req *pb.SendTextRequest) (rxData RxText, err error) {

	rxData.ID = req.IdClient
	rxData.For = req.For
	rxData.Text = req.Text
	rxData.CreatedAt = req.CreatedAt

	return rxData, nil
}

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerSendTextGetToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Логика. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	rxData - принятые данные.
//	m - указатель на экземпляр сервиса.
func layerSendTextContext(ctx context.Context, rxData RxText, m *Manager) error {

	if err := m.storage.AddDataTextContext(ctx, rxData.For, rxData.Text, rxData.CreatedAt); err != nil {
		return fmt.Errorf("Функция AddDataTextContext, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- SendBankCard ---
//

// Получение данных запроса. Возвращаются данные и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerSendBankCardRx(req *pb.SendBankCardRequest) (rxData RxBankCard, err error) {

	rxData.ID = req.IdClient
	rxData.For = req.For
	rxData.Owner = req.Owner
	rxData.Numb = req.Numb
	rxData.ValidData = req.ValidData
	rxData.Code = req.Code
	rxData.CreatedAt = req.CreatedAt

	return rxData, nil
}

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerSendBankCardGetToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Логика. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	rxData - принятые данные.
//	m - указатель на экземпляр сервиса.
func layerSendBankCardContext(ctx context.Context, rxData RxBankCard, m *Manager) error {

	if err := m.storage.AddDataBankCardContext(ctx, rxData.For, rxData.Owner, rxData.Numb, rxData.ValidData, rxData.Code, rxData.CreatedAt); err != nil {
		return fmt.Errorf("Функция AddDataBankCardContext, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// -- RequestLoginPasswordName ---
//

// Логика. Возвращаются данные и ошибка.
//
// Параметры:
//
//	req - запрос.
func LayerRequestLoginPasswordName(ctx context.Context, s *Manager) (rxData []string, err error) {

	// Получение из БД имён записей логин/пароль.
	rxData, err = s.storage.GetNamesLoginPasswordContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Функция GetNamesLoginPasswordContext, вернула ошибку: <%w>", err)
	}

	return rxData, nil
}

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func LayerRequestLoginPasswordNameToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Формирование ответа. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	rxData - принятые данные.
//	m - указатель на экземпляр сервиса.
func LayerRequestLoginPasswordNameTx(data []string) (res *pb.RequestLoginPasswordNameResponse, err error) {

	res = &pb.RequestLoginPasswordNameResponse{
		EntriesName: data,
	}

	return res, nil
}

//
// --- RequestLoginPasswordByName ---
//

// Получение токена из запроса. Возвращаются данные и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerRequestLoginPasswordByNameToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Получение данных запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerRequestLoginPasswordByName(req *pb.RequestLoginPasswordByNameRequest) (name RxReqLoginPasswordByName, err error) {

	name.ClientID = req.IdClient
	name.Name = req.Name

	return name, nil
}

// Формирование ответа. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	rxData - принятые данные.
//	m - указатель на экземпляр сервиса.
func layerRequestLoginPasswordByNameTx(txData TxLoginPassword) (*pb.RequestLoginPasswordByNameResponse, error) {

	resp := &pb.RequestLoginPasswordByNameResponse{
		Name:      txData.For,
		Login:     txData.Login,
		Password:  txData.Password,
		CreatedAt: txData.CreatedAt,
	}
	return resp, nil
}

//
// -- RequestTextName ---
//

// Логика. Возвращаются данные и ошибка.
//
// Параметры:
//
//	req - запрос.
//	s - указатель на экземпляр сервиса.
func LayerRequestTextName(ctx context.Context, s *Manager) (rxData []string, err error) {

	// Получение из БД имён записей текста.
	rxData, err = s.storage.GetNamesTextContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Функция GetNamesTextContext, вернула ошибку: <%w>", err)
	}

	return rxData, nil
}

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func LayerRequestTextNameToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Формирование ответа. Возвращается ответ и ошибка.
//
// Параметры:
//
//	data - данные.
func LayerRequestTextNameTx(data []string) (res *pb.RequestTextNameResponse, err error) {

	res = &pb.RequestTextNameResponse{
		EntriesName: data,
	}

	return res, nil
}

//
// --- RequestTextByName ---
//

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerRequestTextByNameToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Получение данных запроса. Возвращается имя и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerRequestTextByName(req *pb.RequestTextByNameRequest) (name RxReqTextByName, err error) {

	name.ClientID = req.IdClient
	name.Name = req.Name

	return name, nil
}

// Формирование ответа. Возвращается ответ и ошибка.
//
// Параметры:
//
//	txData - данные.
func layerRequestTextByNameTx(txData TxText) (*pb.RequestTextByNameResponse, error) {

	resp := &pb.RequestTextByNameResponse{
		Name:      txData.For,
		Text:      txData.Text,
		CreatedAt: txData.CreatedAt,
	}
	return resp, nil
}

//
// -- RequestBankCardName ---
//

// Логика. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	s - указатель на экземпляр сервиса.
func LayerRequestBankCardName(ctx context.Context, s *Manager) (rxData []string, err error) {

	// Получение из БД имён записей текста.
	rxData, err = s.storage.GetNamesBankCardContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Функция GetNamesTextContext, вернула ошибку: <%w>", err)
	}

	return rxData, nil
}

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func LayerRequestBankCardNameToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Формирование ответа. Возвращается ответ и ошибка.
//
// Параметры:
//
//	data - данные.
func LayerRequestBankCardNameTx(data []string) (res *pb.RequestBankCardNameResponse, err error) {

	res = &pb.RequestBankCardNameResponse{
		EntriesName: data,
	}

	return res, nil
}

//
// --- RequestBankCardByName ---
//

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerRequestBankCardByNameToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Получение данных запроса. Возвращается имя и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerRequestBankCardByName(req *pb.RequestBankCardByNameRequest) (name RxReqBankCardByName, err error) {

	name.ClientID = req.IdClient
	name.Name = req.Name

	return name, nil
}

// Формирование ответа. Возвращается ответ и ошибка.
//
// Параметры:
//
//	txData - данные.
func layerRequestBankCardByNameTx(txData TxBankCard) (*pb.RequestBankCardByNameResponse, error) {

	resp := &pb.RequestBankCardByNameResponse{
		Name:      txData.For,
		Owner:     txData.Owner,
		Numb:      txData.Numb,
		Valid:     txData.Valid,
		Code:      txData.Code,
		CreatedAt: txData.CreatedAt,
	}

	return resp, nil
}

//
// --- RequestFileName ---
//

// Получение имён файлов. Возвращаются имена файловб признак занятости сервера и ошибка.
//
// Параметры:
//
//	dir - директория расположенния файлов.
//	s - указатель на экземпляр сервиса.
func layerRequestFileNameScanDir(dir string, s *Manager) (fileNames []string, isBusyServer bool, err error) {

	// Проверка активности по работе с файлами
	if s.GetStatusRx() == stageActive {
		return nil, true, nil
	}

	// Чтение содержимого директории
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, false, err
	}

	// Сбор имён файлов
	for _, file := range files {
		if !file.IsDir() { // отсев дирекотрий
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, false, nil
}

// Подготовка ответа. Возвращается ответ и ошибка.
//
// Параметры:
//
//	fileNames - имена файлов.
//	isBusyServer - признак занятости сервера.
func layerRequestFileNameTx(fileNames []string, isBusyServer bool) (*pb.RequestFileNameResponse, error) {

	// Ответ.
	resp := &pb.RequestFileNameResponse{
		EntriesName: make([]string, 0, len(fileNames)),
		IsBusy:      isBusyServer,
	}

	// Заполнение.
	for _, v := range fileNames {
		resp.EntriesName = append(resp.EntriesName, v)
	}

	// Результат.
	return resp, nil
}

//
// --- RequestFileInfo ---
//

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerRequestFileInfoToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Получение данных запроса. Возвращается id клиента, имя файла токен и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerRequestFileInfoRx(req *pb.RequestFileInfoRequest) (idClient, fileName string, err error) {

	idClient = req.IdClient
	fileName = req.Name

	if idClient == "" {
		return "", "", EmptyDataIDClient
	}

	if fileName == "" {
		return "", "", EmptyDataFileNAme
	}

	return idClient, fileName, nil
}

// Логика обработчика. Возвращается информация по файлу и ошибка.
//
// Параметры:
//
//	filePath - путь к файлу.
func layerRequestFileInfo(filePath string) (data fileInfo, err error) {

	// Получение размера файла.
	data.size, err = sizeFile(filePath)
	if err != nil {
		return fileInfo{}, fmt.Errorf("Функция sizeFile, вернула ошибку: <%w>", err)
	}

	// Получение хэш файла.
	data.hash, err = hashFile(filePath)
	if err != nil {
		return fileInfo{}, fmt.Errorf("Функция hashFile, вернула ошибку: <%w>", err)
	}

	// Получение имени файла.
	data.name = path.Base(filePath)

	return data, nil
}

// Подготовка ответа. Возвращается ответ и ошибка.
//
// Параметры:
//
//	data  - данные ответа.
func layerRequestFileInfoTx(data fileInfo) (res *pb.RequestFileInfoResponse, err error) {

	r := &pb.RequestFileInfoResponse{
		Name: data.name,
		Hash: data.hash,
		Size: data.size,
	}

	return r, nil
}

//
// --- DeleteLoginPassword ---
//

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx  - контекст.
func layerDeleteLoginPasswordToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Получение данных запроса. Возвращается id клиента, имя и ошибка.
//
// Параметры:
//
//	req  - запрос.
func layerDeleteLoginPasswordRx(req *pb.RequestDeleteName) (idClient, name string, err error) {

	idClient = req.IdClient
	name = req.Name

	if idClient == "" {
		return "", "", EmptyDataIDClient
	}

	if name == "" {
		return "", "", EmptyDataFileNAme
	}

	return idClient, name, nil
}

// Логика. Возвращается ошибка.
//
// Параметры:
//
//	name  - имя.
//	s  - указатель на экземпляр сервиса.
//	ctx - контекст.
func layerDeleteLoginPassword(name string, s *Manager, ctx context.Context) error {

	// Удаление.
	if err := s.storage.DelDataLoginPasswordContext(ctx, name); err != nil {
		return fmt.Errorf("Функция DelDataLoginPasswordContext, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- DeleteText ---
//

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerDeleteTextToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Получение данных запроса. Возвращается id клиента, имя и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerDeleteTextRx(req *pb.RequestDeleteName) (idClient, name string, err error) {

	idClient = req.IdClient
	name = req.Name

	if idClient == "" {
		return "", "", EmptyDataIDClient
	}

	if name == "" {
		return "", "", EmptyDataFileNAme
	}

	return idClient, name, nil
}

// Логика. Возвращается ошибка.
//
// Параметры:
//
//	name - имя.
//	s - указатель на экземпляр сервиса.
//	ctx - контекст.
func layerDeleteText(name string, s *Manager, ctx context.Context) error {

	// Удаление.
	if err := s.storage.DelTextContext(ctx, name); err != nil {
		return fmt.Errorf("Функция DelTextContext, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- DeleteBankCard ---
//

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerDeleteBankCardToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Получение данных запроса.  Возвращается id клиента, имя и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerDeleteBankCardRx(req *pb.RequestDeleteName) (idClient, name string, err error) {

	idClient = req.IdClient
	name = req.Name

	if idClient == "" {
		return "", "", EmptyDataIDClient
	}

	if name == "" {
		return "", "", EmptyDataFileNAme
	}

	return idClient, name, nil
}

// Логика. Возвращается ошибка.
//
// Параметры:
//
//	name - имя.
//	s - указатель на экземпляр сервиса.
//	ctx - контекст.
func layerDeleteBankCard(name string, s *Manager, ctx context.Context) error {

	// Удаление.
	if err := s.storage.DelBankCardContext(ctx, name); err != nil {
		return fmt.Errorf("Функция DelBankCardContext, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- DeleteFile ---
//

// Получение токена из запроса. Возвращается токен и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func layerDeleteFileToken(ctx context.Context) (token tokenData, err error) {

	// Считывание заголовков
	rxMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return tokenData{}, ErrMissingMetadata
	}

	nameToken := "token"

	// Извлечение метаданных.
	tokens := rxMD[nameToken]
	if len(tokens) == 0 {
		return tokenData{}, ErrMissingToken
	}
	if tokens[0] == "" {
		return tokenData{}, ErrIsEmptyToken
	}

	token.name = nameToken
	token.token = tokens[0]

	return token, nil
}

// Получение данных запроса. Возвращается id клиента, имя токен и ошибка.
//
// Параметры:
//
//	req - запрос.
func layerDeleteFileRx(req *pb.RequestDeleteName) (idClient, name string, err error) {

	idClient = req.IdClient
	name = req.Name

	if idClient == "" {
		return "", "", EmptyDataIDClient
	}

	if name == "" {
		return "", "", EmptyDataFileNAme
	}

	return idClient, name, nil
}

// Логика. Возвращается  ошибка.
//
// Параметры:
//
//	name - имя.
//	s - указатель на экземпляр сервиса.
func layerDeleteFile(name string, s *Manager) error {

	filePath := path.Join(s.flag.NameSubDirFiles, name)

	if fileExists(filePath) {
		if err := os.Remove(filePath); err != nil {
			s.logger.Error("Ошибка удаления файла", zap.String("ошибка", err.Error()), zap.String("файл", name))
			return fmt.Errorf("Ошибка: <%w> удаления файла:<%s>", err, name)
		}
		return nil
	}

	s.logger.Error("Отсутствует файл", zap.String("файл", filePath))
	return fmt.Errorf("Отсутствует файл: <%s>", name)
}
