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
	RequestLoginPasswordNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) ([]string, error)
	RequestLoginPasswordByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte) (rxData RxLoginPassword, err error)
	RequestTextNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) ([]string, error)
	RequestTextByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte) (rxData RxText, err error)
	RequestBankCardNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) ([]string, error)
	RequestBankCardByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte) (rxData RxBankCard, err error)
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
	if data.For == "" {
		return EmptyDataArgumentTxID
	}
	if data.For == "" {
		return EmptyDataArgumentTxFor
	}
	if data.Login == "" {
		return EmptyDataArgumentTxLogin
	}
	if s.client == nil {
		return NilPtrConnect
	}

	// Шифрование передаваемых данных.
	txData := TxLoginPassword{
		ID:        data.ID,
		For:       data.For,
		Login:     data.Login,
		Password:  data.Password,
		CreatedAt: data.CreatedAt,
	}

	enTxData, err := layerSendLoginPasswordEncode(txData, key)
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
		IdClient:  enTxData.ID,
		For:       enTxData.For,
		Login:     enTxData.Login,
		Password:  enTxData.Password,
		CreatedAt: enTxData.CreatedAt,
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
	if data.For == "" {
		return EmptyDataArgumentTxID
	}
	if data.For == "" {
		return EmptyDataArgumentTxFor
	}
	if data.Text == "" {
		return EmptyDataArgumentTxText
	}
	if s.client == nil {
		return NilPtrConnect
	}

	// Шифрование передаваемых данных.
	txData := TxText{
		ID:        data.ID,
		For:       data.For,
		Text:      data.Text,
		CreatedAt: data.CreatedAt,
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
		IdClient:  eData.ID,
		For:       eData.For,
		Text:      eData.Text,
		CreatedAt: eData.CreatedAt,
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
	if data.For == "" {
		return EmptyDataArgumentTxID
	}
	if data.For == "" {
		return EmptyDataArgumentTxFor
	}
	if data.Owner == "" {
		return EmptyDataArgumentTxOwner
	}
	if data.Numb == "" {
		return EmptyDataArgumentTxNumb
	}
	if data.ValidData == "" {
		return EmptyDataArgumentTxValidData
	}
	if data.Code == "" {
		return EmptyDataArgumentTxCode
	}
	if s.client == nil {
		return NilPtrConnect
	}

	// Шифрование передаваемых данных.
	txData := TxBankCard{
		ID:        data.ID,
		For:       data.For,
		Owner:     data.Owner,
		Numb:      data.Numb,
		ValidData: data.ValidData,
		Code:      data.Code,
		CreatedAt: data.CreatedAt,
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
		IdClient:  eData.ID,
		For:       eData.For,
		Owner:     eData.Owner,
		Numb:      eData.Numb,
		ValidData: eData.ValidData,
		Code:      eData.Code,
		CreatedAt: eData.CreatedAt,
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
	enNameFile, err := layerSendFileEncrypt(fileName, key)
	if err != nil {
		return fmt.Errorf("Функция layerSendFileEncrypt, вернула ошибку: <%w>", err)
	}

	// Передача файла.
	rxHash, err := layerSendFileTx(s.client, enNameFile, tokenSrv, idClient)
	if err != nil {
		return fmt.Errorf("Функция layerSendFileTx, вернула ошибку: <%w>", err)
	}

	// Вычисление хеша переданного файла
	if err := layerSendFileCheckHash(enNameFile, rxHash); err != nil {
		return fmt.Errorf("Функция layerSendFileCheckHash, вернула ошибку: <%w>", err)
	}

	// Удаление созданного зашифрованного файла
	if err := layerSendFileRemove(enNameFile); err != nil {
		return fmt.Errorf("Функция layerSendFileRemove, вернула ошибку: <%w>", err)
	}

	return nil
}

// Запрос у сервера имён записей для логин/пароль
func (s *server) RequestLoginPasswordNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) ([]string, error) {

	// Запрос у сервера имён записей для логин/пароль.
	enRxData, err := layerRequestLoginPasswordNamesTx(s.client, tokenAuth, idClient)
	if err != nil {
		return nil, fmt.Errorf("Функция layerRequestLoginPasswordNamesTx, вернула ошибку: <%w>", err)
	}

	// Расшифровка принятых данных.
	rxData, err := layerRequestLoginPasswordNamesDecrypt(enRxData, key)
	if err != nil {
		return nil, fmt.Errorf("Функция layerRequestLoginPasswordNamesDecrypt, вернула ошибку: <%w>", err)
	}

	// Результат.
	return rxData, nil
}

// Запрос у сервера записи логин/пароль по его имени
func (s *server) RequestLoginPasswordByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte) (rxData RxLoginPassword, err error) {

	// Шифрование имени записи логин/пароль
	enNameEntry, err := LayerRequestLoginPasswordByNameEncrypt(nameEntry, key)
	if err != nil {
		return RxLoginPassword{}, fmt.Errorf("Функция LayerRequestLoginPasswordByNameEncrypt, вернула ошибку: <%w>", err)
	}

	// Запрос
	resp, err := LayerRequestLoginPasswordByNameTx(s.client, tokenAuth, idClient, enNameEntry)
	if err != nil {
		return RxLoginPassword{}, fmt.Errorf("Функция LayerRequestLoginPasswordByNameTx, вернула ошибку: <%w>", err)
	}

	// Обработка ответа
	rxData, err = LayerRequestLoginPasswordByDecrypt(resp, key)
	if err != nil {
		return RxLoginPassword{}, fmt.Errorf("Функция LayerRequestLoginPasswordByDecrypt, вернула ошибку: <%w>", err)
	}

	// Результат
	return rxData, nil
}

// Запрос у сервера имён записей для текста
func (s *server) RequestTextNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) ([]string, error) {

	// Запрос у сервера имён записей для логин/пароль.
	enRxData, err := layerRequestRequestTextNamesTx(s.client, tokenAuth, idClient)
	if err != nil {
		return nil, fmt.Errorf("Функция layerRequestRequestTextNamesTx, вернула ошибку: <%w>", err)
	}

	// Расшифровка принятых данных.
	rxData, err := layerRequestTextNamesDecrypt(enRxData, key)
	if err != nil {
		return nil, fmt.Errorf("Функция layerRequestTextNamesDecrypt, вернула ошибку: <%w>", err)
	}

	// Результат.
	return rxData, nil
}

// Запрос у сервера записи текста по его имени
func (s *server) RequestTextByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte) (rxData RxText, err error) {

	// Шифрование имени записи логин/пароль
	enNameEntry, err := LayerRequestTextByNameEncrypt(nameEntry, key)
	if err != nil {
		return RxText{}, fmt.Errorf("Функция LayerRequestTextByNameEncrypt, вернула ошибку: <%w>", err)
	}

	// Запрос
	resp, err := LayerRequestTextByNameTx(s.client, tokenAuth, idClient, enNameEntry)
	if err != nil {
		return RxText{}, fmt.Errorf("Функция LayerRequestTextByNameTx, вернула ошибку: <%w>", err)
	}

	// Обработка ответа
	rxData, err = LayerRequestTextByNameDecrypt(resp, key)
	if err != nil {
		return RxText{}, fmt.Errorf("Функция LayerRequestTextByDecrypt, вернула ошибку: <%w>", err)
	}

	// Результат
	return rxData, nil
}

// Запрос у сервера имён записей для банковских карт.
func (s *server) RequestBankCardNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) ([]string, error) {

	// Запрос у сервера имён записей для банковских карт.
	enRxData, err := layerRequestBankCardNamesTx(s.client, tokenAuth, idClient)
	if err != nil {
		return nil, fmt.Errorf("Функция layerRequestRequestTextNamesTx, вернула ошибку: <%w>", err)
	}

	// Расшифровка принятых данных.
	rxData, err := layerRequestBankCardNamesDecrypt(enRxData, key)
	if err != nil {
		return nil, fmt.Errorf("Функция layerRequestTextNamesDecrypt, вернула ошибку: <%w>", err)
	}

	// Результат.
	return rxData, nil
}

// Запрос у сервера записи банковской карты по его имени
func (s *server) RequestBankCardByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte) (rxData RxBankCard, err error) {

	// Шифрование имени записи банковской карты
	enNameEntry, err := LayerRequestBankCardByNameEncrypt(nameEntry, key)
	if err != nil {
		return RxBankCard{}, fmt.Errorf("Функция LayerRequestBankCardByNameEncrypt, вернула ошибку: <%w>", err)
	}

	// Запрос
	resp, err := LayerRequestBankCardByNameTx(s.client, tokenAuth, idClient, enNameEntry)
	if err != nil {
		return RxBankCard{}, fmt.Errorf("Функция LayerRequestBankCardByNameTx, вернула ошибку: <%w>", err)
	}

	// Обработка ответа
	rxData, err = LayerRequestBankCardByNameDecrypt(resp, key)
	if err != nil {
		return RxBankCard{}, fmt.Errorf("Функция LayerRequestBankCardByNameDecrypt, вернула ошибку: <%w>", err)
	}

	// Результат
	return rxData, nil
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
