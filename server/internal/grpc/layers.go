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

// Логика.
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

//
// --- SendLoginPassword ---
//

// Получение данных запроса.
func layerSendLoginPasswordRx(req *pb.SendLoginPasswordRequest) (rxData RxLoginPassword, err error) {

	rxData.ID = req.IdClient
	rxData.For = req.For
	rxData.Login = req.Login
	rxData.Password = req.Password
	rxData.CreatedAt = req.CreatedAt

	return rxData, nil
}

// Получение токена из запроса.
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

// Логика.
func layerSendLoginPasswordLogicContext(ctx context.Context, rxData RxLoginPassword, m *Manager) error {

	if err := m.storage.AddDataLoginPasswordContext(ctx, rxData.For, rxData.Login, rxData.Password, rxData.CreatedAt); err != nil {
		return fmt.Errorf("Функция AddDataLoginPasswordContext, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- SendText ---
//

// Получение данных запроса.
func layerSendTextRx(req *pb.SendTextRequest) (rxData RxText, err error) {

	rxData.ID = req.IdClient
	rxData.For = req.For
	rxData.Text = req.Text
	rxData.CreatedAt = req.CreatedAt

	return rxData, nil
}

// Получение токена из запроса.
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

// Логика.
func layerSendTextContext(ctx context.Context, rxData RxText, m *Manager) error {

	if err := m.storage.AddDataTextContext(ctx, rxData.For, rxData.Text, rxData.CreatedAt); err != nil {
		return fmt.Errorf("Функция AddDataTextContext, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- SendBankCard ---
//

// Получение данных запроса.
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

// Получение токена из запроса.
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

// Логика.
func layerSendBankCardContext(ctx context.Context, rxData RxBankCard, m *Manager) error {

	if err := m.storage.AddDataBankCardContext(ctx, rxData.For, rxData.Owner, rxData.Numb, rxData.ValidData, rxData.Code, rxData.CreatedAt); err != nil {
		return fmt.Errorf("Функция AddDataBankCardContext, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// -- RequestLoginPasswordName ---
//

// Логика.
func LayerRequestLoginPasswordName(ctx context.Context, s *Manager) (rxData []string, err error) {

	// Получение из БД имён записей логин/пароль.
	rxData, err = s.storage.GetNamesLoginPasswordContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Функция GetNamesLoginPasswordContext, вернула ошибку: <%w>", err)
	}

	return rxData, nil
}

// Получение токена из запроса.
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

// Формирование ответа.
func LayerRequestLoginPasswordNameTx(data []string) (res *pb.RequestLoginPasswordNameResponse, err error) {

	res = &pb.RequestLoginPasswordNameResponse{
		EntriesName: data,
	}

	return res, nil
}

//
// --- RequestLoginPasswordByName ---
//

// Получение токена из запроса.
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

// Получение данных запроса.
func layerRequestLoginPasswordByName(req *pb.RequestLoginPasswordByNameRequest) (name RxReqLoginPasswordByName, err error) {

	name.ClientID = req.IdClient
	name.Name = req.Name

	return name, nil
}

// Формирование ответа.
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

// Логика.
func LayerRequestTextName(ctx context.Context, s *Manager) (rxData []string, err error) {

	// Получение из БД имён записей текста.
	rxData, err = s.storage.GetNamesTextContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Функция GetNamesTextContext, вернула ошибку: <%w>", err)
	}

	return rxData, nil
}

// Получение токена из запроса.
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

// Формирование ответа.
func LayerRequestTextNameTx(data []string) (res *pb.RequestTextNameResponse, err error) {

	res = &pb.RequestTextNameResponse{
		EntriesName: data,
	}

	return res, nil
}

//
// --- RequestTextByName ---
//

// Получение токена из запроса.
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

// Получение данных запроса.
func layerRequestTextByName(req *pb.RequestTextByNameRequest) (name RxReqTextByName, err error) {

	name.ClientID = req.IdClient
	name.Name = req.Name

	return name, nil
}

// Формирование ответа.
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

// Логика.
func LayerRequestBankCardName(ctx context.Context, s *Manager) (rxData []string, err error) {

	// Получение из БД имён записей текста.
	rxData, err = s.storage.GetNamesBankCardContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Функция GetNamesTextContext, вернула ошибку: <%w>", err)
	}

	return rxData, nil
}

// Получение токена из запроса.
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

// Формирование ответа.
func LayerRequestBankCardNameTx(data []string) (res *pb.RequestBankCardNameResponse, err error) {

	res = &pb.RequestBankCardNameResponse{
		EntriesName: data,
	}

	return res, nil
}

//
// --- RequestBankCardByName ---
//

// Получение токена из запроса.
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

// Получение данных запроса.
func layerRequestBankCardByName(req *pb.RequestBankCardByNameRequest) (name RxReqBankCardByName, err error) {

	name.ClientID = req.IdClient
	name.Name = req.Name

	return name, nil
}

// Формирование ответа.
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

// Получение имён файлов.
func layerRequestFileNameScanDir(dir string) (fileNames []string, err error) {

	// Чтение содержимого директории
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Сбор имён файлов
	for _, file := range files {
		if !file.IsDir() { // отсев дирекотрий
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}

// Подготовка ответа.
func layerRequestFileNameTx(fileNames []string) (*pb.RequestFileNameResponse, error) {

	resp := &pb.RequestFileNameResponse{
		EntriesName: make([]string, 0, len(fileNames)),
	}

	// Заполнение
	for _, v := range fileNames {
		resp.EntriesName = append(resp.EntriesName, v)
	}

	return resp, nil
}

//
// --- RequestFileInfo ---
//

// Получение токена из запроса.
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

// Получение данных запроса.
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

// Логика обработчика.
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

// Подготовка ответа.
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

// Получение токена из запроса.
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

// Получение данных запроса.
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

// Логика.
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

// Получение токена из запроса.
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

// Получение данных запроса.
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

// Логика.
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

// Получение токена из запроса.
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

// Получение данных запроса.
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

// Логика.
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

// Получение токена из запроса.
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

// Получение данных запроса.
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

// Логика.
func layerDeleteFile(name string, s *Manager) error {

	filePath := s.flag.NameSubDirFiles + "/" + name

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
