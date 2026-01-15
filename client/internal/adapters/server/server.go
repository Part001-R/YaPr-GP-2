package server

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/logfile"
	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Мьютексы.
type mutexes struct {
	processTxFile        sync.Mutex // Процесс передачи файла.
	processRxFile        sync.Mutex // Процесс приёма файла.
	processBackUpRestore sync.Mutex // Процесс BackUp/Restore.

}

// Представление сервера.
type server struct {
	ip          string                   // ip сервера.
	port        string                   // port сервера.
	client      pb.PasswordManagerClient // клиент.
	connect     *grpc.ClientConn         // коннект.
	tokenSrv    string                   // токен сервера.
	mtx         mutexes                  // мьютексы.
	dataRxFile  dataRequestFile          // данные для приёма файла.
	dataTxFile  dataSendFile             // данные для передачи файла.
	dataBackUp  dataBackUp               // данные для процесса BackUp.
	dataRestore dataRestore              // данные для процесса BackUp.
}

// Интерфейс действий.
type ActionsI interface {
	InitDataBackUp(listFiles []string, sizeSendFile int64, secretKey [32]byte) error
	BackUp(chProcess chan<- float32, chErr chan<- error, chDone chan<- struct{}, lgr *logfile.LogFile)
	InitDataRestore(listFiles []InfoByFiles) error
	RestoreRequestFilesInfo() (data []InfoByFiles, err error)
	Restore(chProcess chan<- float32, chErr chan<- error, chDone chan<- struct{}, lgr *logfile.LogFile)
	ConnectClose() error
	AuthenticationContext(ctx context.Context, userName, userPwd string) (tokenAuth string, err error)
	SendLoginPassword(ctx context.Context, data TxLoginPassword, tokenSrv string, key [32]byte) error
	SendText(ctx context.Context, data TxText, tokenSrv string, key [32]byte) error
	SendBankCard(ctx context.Context, data TxBankCard, tokenSrv string, key [32]byte) error
	SendFile(chProcess chan<- float32, chErr chan<- error, chDone chan<- struct{})
	RequestLoginPasswordNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) ([]string, error)
	RequestLoginPasswordByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte) (rxData RxLoginPassword, err error)
	RequestTextNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) ([]string, error)
	RequestTextByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte) (rxData RxText, err error)
	RequestBankCardNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) ([]string, error)
	RequestBankCardByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte) (rxData RxBankCard, err error)
	RequestFileNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) (rxData []string, err error)
	RequestFileInfo(tokenAuth, idClient, nameFile string) (fileName, fileHash string, fileSize int64, err error)
	RequestFileByName(chProcess chan<- float32, chErr chan<- error, chDone chan<- struct{})
	InitDataRequestFileByName(fileName, tokenAuth, clientID string, sizeReqFile, sizePassed int64, secretKey [32]byte) error
	InitDataSendFile(filePath, tokenAuth, clientID string, sizeSendFile, sizePassed int64, secretKey [32]byte) error
	DeleteLoginPassword(idClient, name string) error
	DeleteText(idClient, name string) error
	DeleteBankCard(idClient, name string) error
	DeleteFile(idClient, name string) error
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
		mtx: mutexes{
			processTxFile:        sync.Mutex{},
			processRxFile:        sync.Mutex{},
			processBackUpRestore: sync.Mutex{},
		},
		dataRxFile:  dataRequestFile{},
		dataTxFile:  dataSendFile{},
		dataBackUp:  dataBackUp{},
		dataRestore: dataRestore{},
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
func (s *server) AuthenticationContext(ctx context.Context, userName, userPwd string) (tokenAuth string, err error) {

	// Проверка аргументов.
	if userName == "" {
		return "", EmptyDataArgumentUserName
	}
	if userPwd == "" {
		return "", EmptyDataArgumentUserPwd
	}
	if s.client == nil {
		return "", NilPtrConnect
	}

	//
	// логика
	//

	// Создание метаданных с токеном.
	txMD, secretKey, nameToken, err := createTokenForAuthentication()
	if err != nil {
		return "", fmt.Errorf("функция createTokenForAuthentication, вернула ошибку: <%w>", err)
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
		return "", fmt.Errorf("функция client.Authentication, вернула ошибку: <%w>", err)
	}

	// --- Получение отправленного токена.

	// Получение токена из метаданных ответа.
	token := header[nameToken]
	if len(token) == 0 || token[0] == "" {
		return "", MissingTokenData
	}
	rxToken := token[0]

	// Проверка токена.
	if err := checkToken(rxToken, secretKey); err != nil {
		return "", fmt.Errorf("функция checkToken, вернула ошибку: <%w>", err)
	}

	// --- Получение токена регистрации.

	s.tokenSrv = res.Token

	if s.tokenSrv == "" {
		return "", MissingTokenSrvData
	}

	return s.tokenSrv, nil
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
func (s *server) SendFile(chProcess chan<- float32, chErr chan<- error, chDone chan<- struct{}) {

	defer func() {
		close(chProcess)
		close(chErr)
		close(chDone)
	}()

	// Создание зашифрованной версии файла.
	enNameFile, err := layerSendFileEncrypt(s.dataTxFile.filePath, s.dataTxFile.secretKey)
	if err != nil {
		chErr <- fmt.Errorf("Функция layerSendFileEncrypt, вернула ошибку: <%w>", err)
	}

	// Передача файла.
	s.dataTxFile.filePath = enNameFile

	rxHash, err := layerSendFileTx(s, chProcess)
	if err != nil {
		chErr <- fmt.Errorf("Функция layerSendFileTx, вернула ошибку: <%w>", err)
	}

	// Проверка хэш.
	if err := layerSendFileCheckHash(enNameFile, rxHash); err != nil {
		chErr <- fmt.Errorf("Функция layerSendFileCheckHash, вернула ошибку: <%w>", err)
	}

	// Удаление созданного зашифрованного файла
	if err := layerSendFileRemove(enNameFile); err != nil {
		chErr <- fmt.Errorf("Функция layerSendFileRemove, вернула ошибку: <%w>", err)
	}

	chDone <- struct{}{}
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
	enRxData, err := layerRequestTextNamesTx(s.client, tokenAuth, idClient)
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

// Запрос у сервера имён файлов
func (s *server) RequestFileNames(ctx context.Context, tokenAuth, idClient string, key [32]byte) (rxData []string, err error) {

	// Запрос.
	rxData, err = layerRequestFileNamesTx(s.client, tokenAuth, idClient)
	if err != nil {
		return nil, fmt.Errorf("Функция layerRequestFileNamesTx, вернула ошибку: <%w>", err)
	}

	// Результат.
	return rxData, nil
}

// Запрос у сервера хэш файла
func (s *server) RequestFileInfo(tokenAuth, idClient, nameFile string) (fileName, fileHash string, fileSize int64, err error) {

	// Запрос.
	fileName, fileHash, fileSize, err = layerRequestFileInfoTx(s.client, tokenAuth, idClient, nameFile)
	if err != nil {
		return "", "", 0, fmt.Errorf("Функция layerRequestFileInfoTx, вернула ошибку: <%w>", err)
	}

	// Проверка имён
	if err := layerRequestFileInfoCheck(nameFile, fileName); err != nil {
		return "", "", 0, fmt.Errorf("Функция layerRequestFileInfoCheck, вернула ошибку: <%w>", err)
	}

	// Результат.
	return fileName, fileHash, fileSize, nil
}

// Запрос файла у сервера. Для запуска как горутина.
func (s *server) RequestFileByName(chProcess chan<- float32, chErr chan<- error, chDone chan<- struct{}) {

	defer func() {
		close(chProcess)
		close(chErr)
		close(chDone)
	}()

	// Создание резервной копии файла. Если он существует.
	tempFileName, err := layerRequestFileByNameCreateTemp(s.dataRxFile.fileName)
	if err != nil {
		chErr <- fmt.Errorf("Функция layerRequestFileByNameCreateTemp, вернула ошибку: <%w>", err)
		return
	}
	// Удаление резервной копии.
	defer func(tempFileName string) {
		if err := layerRequestFileByNameRemoveTemp(tempFileName); err != nil {
			chErr <- fmt.Errorf("Функция layerRequestFileByNameRemoveTemp, вернула ошибку: <%w>", err)
			return
		}
	}(tempFileName)

	// Получение файла.
	if err := layerRequestFileByNameRx(s, &s.dataRxFile, chProcess); err != nil {
		chErr <- fmt.Errorf("Функция layerRequestFileByNameRx, вернула ошибку: <%w>", err)
		return
	}

	// Расшифровка принятого файла.
	if err := layerRequestFileByNameDecrypt(s.dataRxFile.fileName, s.dataRxFile.secretKey); err != nil {
		chErr <- fmt.Errorf("Функция layerRequestFileByNameDecrypt, вернула ошибку: <%w>", err)
		return
	}

	// Признак успешного завершения процесса.
	chDone <- struct{}{}
}

// Инициализация данных, для процесса приёма файла.
func (s *server) InitDataRequestFileByName(fileName, tokenAuth, clientID string, sizeReqFile, sizePassed int64, secretKey [32]byte) error {

	// Проверка аргументов
	if fileName == "" {
		return EmptyDataArgumentFileName
	}
	if tokenAuth == "" {
		return EmptyDataArgumentTokenAuth
	}
	if clientID == "" {
		return EmptyDataArgumentClientID
	}
	if len(secretKey) != 32 {
		return NotCorrectLenData
	}

	// Данные
	s.dataRxFile = dataRequestFile{
		fileName:    fileName,
		tokenAuth:   tokenAuth,
		clientID:    clientID,
		sizeReqFile: sizeReqFile,
		sizePassed:  sizePassed,
		secretKey:   secretKey,
	}

	return nil
}

// Инициализация данных, для процесса приёма файла.
func (s *server) InitDataSendFile(filePath, tokenAuth, clientID string, sizeSendFile, sizePassed int64, secretKey [32]byte) error {

	// Проверка аргументов
	if filePath == "" {
		return EmptyDataArgumentFilePath
	}
	if tokenAuth == "" {
		return EmptyDataArgumentTokenAuth
	}
	if clientID == "" {
		return EmptyDataArgumentClientID
	}
	if len(secretKey) != 32 {
		return NotCorrectLenData
	}

	// Данные
	s.dataTxFile = dataSendFile{
		filePath:     filePath,
		tokenAuth:    tokenAuth,
		clientID:     clientID,
		sizeSendFile: sizeSendFile,
		sizePassed:   0,
		secretKey:    secretKey,
	}

	return nil
}

// Режим - локальный. Резервное копирование файлов на сервер.
func (s *server) BackUp(chProcess chan<- float32, chErr chan<- error, chDone chan<- struct{}, lgr *logfile.LogFile) {

	defer func() {
		close(chProcess)
		close(chErr)
		close(chDone)
	}()

	// Создание токена.
	secretKey, txToken, err := layerBackUpCreateToken()
	if err != nil {
		chErr <- fmt.Errorf("функция layerBackUpCreateToken, вернула ошибку:<%w>", err)
	}

	// Передача файлов.
	for _, txFileName := range s.dataBackUp.listFiles {

		// Передача файла на сервер.
		resp, rxFileHash, rxToken, err := layerBackUpTxFile(s, txFileName, txToken, chProcess, lgr)
		if err != nil {
			chErr <- fmt.Errorf("Функция layerBackUpTxFile, вернула ошибку: <%w>", err)
		}

		// Вычисление хэша переданного файла.
		txFileHash, err := hashFile(txFileName)
		if err != nil {
			chErr <- fmt.Errorf("функция hashFile, вернула ошибку: <%w>", err)
		}

		// Проверка результата.
		if err := layerBackUpCheckResult(resp, txFileName, txFileHash, rxFileHash, rxToken, secretKey); err != nil {
			chErr <- fmt.Errorf("Функция layerBackUpCheckResult, вернула ошибку: <%w>", err)
		}
	}

	chDone <- struct{}{}
}

// Инициализация данных, для процесса BackUp.
func (s *server) InitDataBackUp(listFiles []string, sizeSendFile int64, secretKey [32]byte) error {

	// Проверка аргументов
	if len(listFiles) != 2 {
		return NotCorrectLenLestFiles
	}
	if sizeSendFile <= 0 {
		return EmptyDataArgumentSizeSendFile
	}
	if len(secretKey) != 32 {
		return NotCorrectLenData
	}

	// Инициализация.
	s.dataBackUp.listFiles = listFiles
	s.dataBackUp.secretKey = secretKey
	s.dataBackUp.sizeSendFile = sizeSendFile

	return nil
}

// Режим - локальный. Восстановление файлов из сервера.
func (s *server) Restore(chProcess chan<- float32, chErr chan<- error, chDone chan<- struct{}, lgr *logfile.LogFile) {

	defer func() {
		close(chProcess)
		close(chErr)
		close(chDone)
	}()

	// Создание токена.
	secretKey, txToken, err := layerRestoreCreateToken()
	if err != nil {
		chErr <- fmt.Errorf("функция layerRestoreCreateToken, вернула ошибку:<%w>", err)
	}

	// Получение файлов.
	for _, fileInfo := range s.dataRestore.listFiles {

		needRestoreState := false // Признак необходимости воостановления состояния, при ошибке.
		var tempFileName string   // Имя временного файла.

		// Обработка перед выходом.
		defer func(needRestoreState bool, fileName, tempFileName string, errProcess error) {
			if errRestore := deferProcessRestoreByError(needRestoreState, fileName, tempFileName, errProcess); errRestore != nil {
				err = fmt.Errorf("функция deferProcessRestoreByError, вернула ошибку:<%w>, при ошибку процесса:<%w>", errRestore, errProcess)
				return
			}
		}(needRestoreState, fileInfo.Name, tempFileName, err)

		// Запрос файла у сервера.
		content, srcFileHash, rxToken, err := layerRestoreRxFile(s, fileInfo.Name, txToken, chProcess)
		if err != nil {
			chErr <- fmt.Errorf("Функция layerRestoreRxFile, вернула ошибку: <%w>", err)
			return
		}

		// Изменение имени существующего файла, чтобы в случае ошибки, не потерять данные.
		if isFileExists(fileInfo.Name) {
			tempFileName, err = changeFileName(fileInfo.Name, "-temp")
			if err != nil {
				chErr <- fmt.Errorf("функция changeFileName, вернула ошибку:<%w>", err)
				return
			}
			needRestoreState = true
		}

		// Сохранение принятых данных в файл.
		if err := layerRestoreSaveFile(content, fileInfo.Name); err != nil {
			chErr <- fmt.Errorf("Функция layerRestoreSaveFile, вернула ошибку: <%w>", err)
			return
		}

		// Вычисление хэша принятого файла.
		rxFileHash, err := hashFile(fileInfo.Name)
		if err != nil {
			chErr <- fmt.Errorf("функция hashFile, вернула ошибку: <%w>", err)
			return
		}

		// Проверка результата.
		if err := layerRestoreCheckResult(fileInfo.Name, rxFileHash, srcFileHash, rxToken, secretKey); err != nil {

			// Удаление файла, если проверка не пройдена.
			if errRemove := os.Remove(fileInfo.Name); errRemove != nil {
				chErr <- fmt.Errorf("ошибка:<%w> удаления файла:<%s>, после приёма. Базовая ошибка:<%w>", errRemove, fileInfo.Name, err)
				return
			}
			chErr <- fmt.Errorf("функция layerRestoreCheckResult, вернула ошибку:<%w>, для файла:<%s>", err, fileInfo.Name)
			return
		}

		// Файл успешно принят.
		// Удаление резервного файла, если он существует.
		if isFileExists(tempFileName) {
			if err := os.Remove(tempFileName); err != nil {
				chErr <- fmt.Errorf("ошибка:<%w> удаления резервного файла:<%s>, при успешном приёме", err, fileInfo.Name)
				return
			}
		}
	}

	chDone <- struct{}{}
}

// Инициализация данных, для процесса Restore.
func (s *server) InitDataRestore(listFiles []InfoByFiles) error {

	s.dataRestore = dataRestore{}

	s.dataRestore.listFiles = listFiles

	// Формирование общего размера в КБайт
	for _, f := range listFiles {
		s.dataRestore.totalSizeFiles += f.Volume
	}
	s.dataRestore.totalSizeFiles /= 1024

	return nil
}

// Запрос у сервера информации по файлам, которые будут приняты при Restore.
func (s *server) RestoreRequestFilesInfo() (data []InfoByFiles, err error) {

	// Создание токена.
	secretKey, token, err := layerRestoreRequestFilesInfoCreateToken()
	if err != nil {
		return nil, fmt.Errorf("Функция layerRestoreRequestFilesInfoCreateToken, вернула ошибку: <%w>", err)
	}

	// Запрос.
	rxData, rxToken, err := layerRestoreRequestFilesInfoRequest(s.client, token)
	if err != nil {
		return nil, fmt.Errorf("Функция layerRestoreRequestFilesInfoRequest, вернула ошибку: <%w>", err)
	}

	// Проверка токена.
	if err := layerRestoreRequestFilesInfoCheckToken(rxToken, secretKey); err != nil {
		return nil, fmt.Errorf("Функция layerRestoreRequestFilesInfoCheckToken, вернула ошибку: <%w>", err)
	}

	// Результат.
	return rxData, nil
}

// Удаление записи логин/пароль.
func (s *server) DeleteLoginPassword(idClient, name string) error {

	// Удаление записи.
	if err := layerDeleteLoginPassword(name, idClient, s); err != nil {
		return fmt.Errorf("Функция layerDeleteLoginPassword, вернула ошибку:<%w>", err)
	}

	return nil
}

// Удаление записи текста.
func (s *server) DeleteText(idClient, name string) error {

	// Удаление записи.
	if err := layerDeleteText(name, idClient, s); err != nil {
		return fmt.Errorf("Функция layerDeleteText, вернула ошибку:<%w>", err)
	}

	return nil
}

// Удаление записи банковской карты.
func (s *server) DeleteBankCard(idClient, name string) error {

	// Удаление записи.
	if err := layerDeleteBankCard(name, idClient, s); err != nil {
		return fmt.Errorf("Функция layerDeleteBankCard, вернула ошибку:<%w>", err)
	}

	return nil
}

// Удаление файла.
func (s *server) DeleteFile(idClient, name string) error {

	// Удаление файла.
	if err := layerDeleteFile(name, idClient, s); err != nil {
		return fmt.Errorf("Функция layerDeleteFile, вернула ошибку:<%w>", err)
	}

	return nil
}

//
// --- токен ---
//

// Получение токена аутентификации.
func (s *server) GetTokenAuthentication() string {
	return s.tokenSrv
}

// Обновление токена аутентификации.
func (s *server) UpdateTokenAuthentication(token string) {
	s.tokenSrv = token
}
