package server

import (
	"context"
	"fmt"
	"time"

	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Представление сервера.
type server struct {
	ip       string                   // ip сервера
	port     string                   // port сервера
	client   pb.PasswordManagerClient // клиент
	connect  *grpc.ClientConn         // коннект
	tokenSrv string                   // токен сервера
}

// Интерфейс действий.
type ActionsI interface {
	ConnectClose() error
	AuthenticationContext(ctx context.Context, userName, userPwd string) error
	SendLoginPassword(ctx context.Context, data TxLoginPassword, tokenSrv string, key [32]byte) error
	SendText(ctx context.Context, data TxText, tokenSrv string, key [32]byte) error
	SendBankCard(ctx context.Context, data TxBankCard, tokenSrv string, key [32]byte) error
	SendFile(fileName, tokenSrv, idClient string, key [32]byte) error
	GetTokenAuthentication() string
	UpdateTokenAuthentication(token string)
}

// Интерфейс.
type ServerI interface {
	ActionsI
}

var inst *server

// Конструктор.
func New(ip, port string) (act ServerI, err error) {

	// Закрытие подключения, если было установлено ранее.
	if inst != nil {
		if inst.connect != nil {
			if err := inst.connect.Close(); err != nil {
				return nil, fmt.Errorf("Ошибка закрытия подключения к серверу, при обновлении: <%v>", err)
			}
		}
	}

	// Подключение.
	conn, client, err := connect(ip, port)
	if err != nil {
		return nil, fmt.Errorf("Функция connect, вернула ошибку: <%w>", err)
	}

	inst = &server{
		ip:       ip,
		port:     port,
		client:   client,
		connect:  conn,
		tokenSrv: "",
	}

	return inst, nil
}

// Закрытие коннекта.
func (s *server) ConnectClose() error {

	if s.connect != nil {
		if err := s.connect.Close(); err != nil {
			return fmt.Errorf("Error: Ошибка закрытия подключения к серверу: <%v>", err)
		}
	}
	return nil
}

// Аутентификация.
func (s *server) AuthenticationContext(ctx context.Context, userName, userPwd string) error {

	// Проверка аргументов.
	if userName == "" {
		return EmptyDataArgumentUserName
	}
	if userPwd == "" {
		return EmptyDataArgumentUserPwd
	}
	if s.client == nil {
		return NilPtrConnect
	}

	//
	// логика
	//

	// Создание метаданных с токеном.
	txMD, secretKey, nameToken, err := createTokenForAuthentication()
	if err != nil {
		return fmt.Errorf("функция createTokenForAuthentication, вернула ошибку: <%w>", err)
	}

	// --- Запрос

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, txMD)

	req := &proto.AuthenticationRequest{
		UserName: userName,
		UserPwd:  userPwd,
	}

	var header metadata.MD

	// Запрос регистрации пользователя.
	res, err := s.client.Authentication(ctx, req, grpc.Header(&header))
	if err != nil {
		return fmt.Errorf("функция client.Authentication, вернула ошибку: <%w>", err)
	}

	// --- Получение отправленного токена.

	// Получение токена из метаданных ответа.
	token := header[nameToken]
	if len(token) == 0 || token[0] == "" {
		return MissingTokenData
	}
	rxToken := token[0]

	// Проверка токена.
	if err := checkToken(rxToken, secretKey); err != nil {
		return fmt.Errorf("функция checkToken, вернула ошибку: <%w>", err)
	}

	// --- Получение токена регистрации.

	s.tokenSrv = res.Token

	if s.tokenSrv == "" {
		return MissingTokenSrvData
	}

	return nil
}

// Передача логин/пароль.
func (s *server) SendLoginPassword(ctx context.Context, data TxLoginPassword, tokenSrv string, key [32]byte) error {

	// Проверка аргументов.
	if data.TxFor == "" {
		return EmptyDataArgumentTxID
	}
	if data.TxFor == "" {
		return EmptyDataArgumentTxFor
	}
	if data.TxLogin == "" {
		return EmptyDataArgumentTxLogin
	}
	if s.client == nil {
		return NilPtrConnect
	}

	// Шифрование передаваемых данных.
	txData := TxLoginPassword{
		TxID:        data.TxID,
		TxFor:       data.TxFor,
		TxLogin:     data.TxLogin,
		TxPassword:  data.TxPassword,
		TxCreatedAt: data.TxCreatedAt,
	}
	eData, err := layerSendLoginPasswordEncode(txData, key)
	if err != nil {
		return fmt.Errorf("функция layerSendLoginPasswordEncode, вернула ошибку: <%w>", err)
	}

	// Подготовка метаданных
	nameToken := "token"
	txMD := metadata.Pairs(nameToken, tokenSrv)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, txMD)

	// Подготовка данных
	req := &proto.SendLoginPasswordRequest{
		IdClient:  eData.TxID,
		For:       eData.TxFor,
		Login:     eData.TxLogin,
		Password:  eData.TxPassword,
		CreatedAt: eData.TxCreatedAt,
	}

	var header metadata.MD

	// Запрос.
	_, err = s.client.SendLoginPassword(ctx, req, grpc.Header(&header))
	if err != nil {
		return fmt.Errorf("функция client.SendLoginPassword, вернула ошибку: <%w>", err)
	}

	return nil
}

// Передача текста.
func (s *server) SendText(ctx context.Context, data TxText, tokenSrv string, key [32]byte) error {

	// Проверка аргументов.
	if data.TxFor == "" {
		return EmptyDataArgumentTxID
	}
	if data.TxFor == "" {
		return EmptyDataArgumentTxFor
	}
	if data.TxText == "" {
		return EmptyDataArgumentTxText
	}
	if s.client == nil {
		return NilPtrConnect
	}

	// Шифрование передаваемых данных.
	txData := TxText{
		TxID:        data.TxID,
		TxFor:       data.TxFor,
		TxText:      data.TxText,
		TxCreatedAt: data.TxCreatedAt,
	}
	eData, err := layerSendTextEncode(txData, key)
	if err != nil {
		return fmt.Errorf("функция layerSendTextEncode, вернула ошибку: <%w>", err)
	}

	// Подготовка метаданных
	nameToken := "token"
	txMD := metadata.Pairs(nameToken, tokenSrv)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, txMD)

	// Подготовка данных
	req := &proto.SendTextRequest{
		IdClient:  eData.TxID,
		For:       eData.TxFor,
		Text:      eData.TxText,
		CreatedAt: eData.TxCreatedAt,
	}

	var header metadata.MD

	// Запрос.
	_, err = s.client.SendText(ctx, req, grpc.Header(&header))
	if err != nil {
		return fmt.Errorf("функция client.SendText, вернула ошибку: <%w>", err)
	}

	return nil
}

// Передача банковской карты.
func (s *server) SendBankCard(ctx context.Context, data TxBankCard, tokenSrv string, key [32]byte) error {

	// Проверка аргументов.
	if data.TxFor == "" {
		return EmptyDataArgumentTxID
	}
	if data.TxFor == "" {
		return EmptyDataArgumentTxFor
	}
	if data.TxOwner == "" {
		return EmptyDataArgumentTxOwner
	}
	if data.TxNumb == "" {
		return EmptyDataArgumentTxNumb
	}
	if data.TxValidData == "" {
		return EmptyDataArgumentTxValidData
	}
	if data.TxCode == "" {
		return EmptyDataArgumentTxCode
	}
	if s.client == nil {
		return NilPtrConnect
	}

	// Шифрование передаваемых данных.
	txData := TxBankCard{
		TxID:        data.TxID,
		TxFor:       data.TxFor,
		TxOwner:     data.TxOwner,
		TxNumb:      data.TxNumb,
		TxValidData: data.TxValidData,
		TxCreatedAt: data.TxCreatedAt,
	}
	eData, err := layerSendBankCardEncode(txData, key)
	if err != nil {
		return fmt.Errorf("функция layerSendBankCardEncode, вернула ошибку: <%w>", err)
	}

	// Подготовка метаданных
	nameToken := "token"
	txMD := metadata.Pairs(nameToken, tokenSrv)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, txMD)

	// Подготовка данных
	req := &proto.SendBankCardRequest{
		IdClient:  eData.TxID,
		For:       eData.TxFor,
		Owner:     eData.TxOwner,
		Numb:      eData.TxNumb,
		Code:      eData.TxCode,
		CreatedAt: eData.TxCreatedAt,
	}

	var header metadata.MD

	// Запрос.
	_, err = s.client.SendBankCard(ctx, req, grpc.Header(&header))
	if err != nil {
		return fmt.Errorf("функция client.SendBankCard, вернула ошибку: <%w>", err)
	}

	return nil
}

// Передача файла.
func (s *server) SendFile(fileName, tokenSrv, idClient string, key [32]byte) error {

	// Проверка аргументов.
	if fileName == "" {
		return EmptyDataArgumentFileName
	}
	if s.client == nil {
		return NilPtrConnect
	}

	// Создание зашифрованной версии файла.
	nameFileEncr, err := layerSendFileEncrypt(fileName, key)
	if err != nil {
		return fmt.Errorf("Функция layerSendFileEncrypt, вернула ошибку: <%w>", err)
	}

	// Передача файла.
	rxHash, err := layerSendFileTx(s.client, nameFileEncr, tokenSrv, idClient)
	if err != nil {
		return fmt.Errorf("Функция layerSendFileTx, вернула ошибку: <%w>", err)
	}

	// Вычисление хеша переданного файла
	if err := layerSendFileCheckHash(nameFileEncr, rxHash); err != nil {
		return fmt.Errorf("Функция layerSendFileCheckHash, вернула ошибку: <%w>", err)
	}

	// Удаление созданного зашифрованного файла
	if err := layerSendFileRemove(nameFileEncr); err != nil {
		return fmt.Errorf("Функция layerSendFileRemove, вернула ошибку: <%w>", err)
	}

	return nil
}

//
// --- токен ---
//

// Получение токена аутентификации.
func (s server) GetTokenAuthentication() string {
	return s.tokenSrv
}

// Обновление токена аутентификации.
func (s *server) UpdateTokenAuthentication(token string) {
	s.tokenSrv = token
}
