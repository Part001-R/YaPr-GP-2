package server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Part001-R/YaPr-GP-2/proto"
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

//
// --- SendLoginPassword ---
//

// Шифрование передаваемых данных.
func layerSendLoginPasswordEncode(data TxLoginPassword, key [32]byte) (eData TxLoginPassword, err error) {

	eData.ID = data.ID

	eData.For, err = encrypt(data.For, key)
	if err != nil {
		return TxLoginPassword{}, fmt.Errorf("Error: ошибка шифрования содержимого txFor: <%v>", err)
	}

	eData.Login, err = encrypt(data.Login, key)
	if err != nil {
		return TxLoginPassword{}, fmt.Errorf("Error: ошибка шифрования содержимого txLogin: <%v>", err)
	}

	eData.Password, err = encrypt(data.Password, key)
	if err != nil {
		return TxLoginPassword{}, fmt.Errorf("Error: ошибка шифрования содержимого txPassword: <%v>", err)
	}

	eData.CreatedAt, err = encrypt(data.CreatedAt, key)
	if err != nil {
		return TxLoginPassword{}, fmt.Errorf("Error: ошибка шифрования содержимого TxCreatedAt: <%v>", err)
	}

	return eData, nil
}

//
// --- SendText ---
//

// Шифрование передаваемых данных.
func layerSendTextEncode(data TxText, secretKey [32]byte) (eData TxText, err error) {

	eData.ID = data.ID

	eData.For, err = encrypt(data.For, secretKey)
	if err != nil {
		return TxText{}, fmt.Errorf("Error: ошибка шифрования содержимого txFor: <%v>", err)
	}

	eData.Text, err = encrypt(data.Text, secretKey)
	if err != nil {
		return TxText{}, fmt.Errorf("Error: ошибка шифрования содержимого txLogin: <%v>", err)
	}

	eData.CreatedAt, err = encrypt(data.CreatedAt, secretKey)
	if err != nil {
		return TxText{}, fmt.Errorf("Error: ошибка шифрования содержимого TxCreatedAt: <%v>", err)
	}

	return eData, nil
}

//
// --- SendBankCard ---
//

// Шифрование передаваемых данных.
func layerSendBankCardEncode(data TxBankCard, secretKey [32]byte) (eData TxBankCard, err error) {

	eData.ID = data.ID

	eData.For, err = encrypt(data.For, secretKey)
	if err != nil {
		return TxBankCard{}, fmt.Errorf("Error: ошибка шифрования содержимого txFor: <%w>", err)
	}

	eData.Owner, err = encrypt(data.Owner, secretKey)
	if err != nil {
		return TxBankCard{}, fmt.Errorf("Error: ошибка шифрования содержимого TxOwner: <%w>", err)
	}

	eData.Numb, err = encrypt(data.Numb, secretKey)
	if err != nil {
		return TxBankCard{}, fmt.Errorf("Error: ошибка шифрования содержимого TxNumb: <%w>", err)
	}

	eData.ValidData, err = encrypt(data.ValidData, secretKey)
	if err != nil {
		return TxBankCard{}, fmt.Errorf("Error: ошибка шифрования содержимого TxValidData: <%w>", err)
	}

	eData.Code, err = encrypt(data.Code, secretKey)
	if err != nil {
		return TxBankCard{}, fmt.Errorf("Error: ошибка шифрования содержимого TxCode: <%w>", err)
	}

	eData.CreatedAt, err = encrypt(data.CreatedAt, secretKey)
	if err != nil {
		return TxBankCard{}, fmt.Errorf("Error: ошибка шифрования содержимого TxCreatedAt: <%w>", err)
	}

	return eData, nil
}

//
// --- SendFile ---
//

// Шифрование файла.
func layerSendFileEncrypt(filePath string, key [32]byte) (encFilePath string, err error) {

	// Подключение к исходному файлу.
	inputFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if errCl := inputFile.Close(); errCl != nil {
			err = fmt.Errorf("Ошибка:<%w>, закрытия подключения к файлу:<%s>. Базовая ошибка:<%w>", errCl, filePath, err)
		}
	}()

	// Создание зашифрованного файла
	fileExt := path.Ext(filePath)
	fileName := path.Base(filePath)
	fileName = strings.TrimSuffix(fileName, fileExt)

	encFilePath = fileName + "-enc" + fileExt
	outputFile, err := os.Create(encFilePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if errCl := outputFile.Close(); errCl != nil {
			err = fmt.Errorf("Ошибка:<%w>, закрытия подключения к файлу:<%s>. Базовая ошибка:<%w>", errCl, encFilePath, err)
		}
	}()

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	// Генерация случайного IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	// IV в начало файла
	if _, err := outputFile.Write(iv); err != nil {
		return "", err
	}

	stream := cipher.NewCBCEncrypter(block, iv)

	buffer := make([]byte, 4096)
	for {
		n, err := inputFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// PKCS#7 паддинг (только для последнего блока)
		isLastBlock := (err == io.EOF || n < len(buffer))
		if isLastBlock {
			pad := aes.BlockSize - n%aes.BlockSize
			if pad == 0 {
				pad = aes.BlockSize
			}
			buffer = append(buffer[:n], bytes.Repeat([]byte{byte(pad)}, pad)...)
			n += pad
		}

		// Шифрация
		ciphertext := make([]byte, n)
		stream.CryptBlocks(ciphertext, buffer[:n])

		// Пишем в файл
		if _, err := outputFile.Write(ciphertext); err != nil {
			return "", err
		}
	}

	return encFilePath, nil
}

// Передача файла на сервер.
func layerSendFileTx(client proto.PasswordManagerClient, filePath, tokenAuth, idClient string) (rxHash string, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", tokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Инициализация стрима для передачи файла
	stream, err := client.SendFile(ctx)
	if err != nil {
		return "", fmt.Errorf("ошибка создания stream, для передачи данных: <%w>", err)
	}

	// Открытие файла для отправки
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла передачи: <%w>", err)
	}
	defer func() {
		if errClose := file.Close(); errClose != nil {
			err = fmt.Errorf("Ошибка при закрытии подключения к файлу: <%w>. Исходная ошибка: <%w>", errClose, err)
		}
	}()

	// Чтение файла по частям
	reader := bufio.NewReader(file)
	buf := make([]byte, 1024)

	// Получение имени и тапа файла из полного пути
	fileName := filepath.Base(filePath)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				return "", fmt.Errorf("ошибка при чтении файла: <%w>", err)
			}
			break
		}

		req := &pb.SendFileRequest{
			FileName: fileName,
			Content:  buf[:n],
			IdClient: idClient,
		}

		if err := stream.Send(req); err != nil {
			return "", fmt.Errorf("ошибка отправки данных файла: <%w>", err)
		}
	}

	// Закрытие потока передачи и ожидание ответа от сервера
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return "", fmt.Errorf("ошибка получения ответа от сервера: <%w>", err)
	}

	// Проверка имени файла из ответа.
	if path.Base(resp.FileName) != path.Base(fileName) {
		return "", fmt.Errorf("нет соответствия имени файла в ответе. Ожидалось:<%s>, а принято:<%s>", filePath, resp.FileName)
	}

	// Получение трейлера хэша
	rxTrailer := stream.Trailer()
	if hash, ok := rxTrailer["hash"]; ok {
		rxHash = hash[0]
	} else {
		return "", fmt.Errorf("Сервер не предоставил трейлер с данными хэша для файла: <%s>", filePath)
	}

	// Результат
	return rxHash, nil
}

// Проверка хэша.
func layerSendFileCheckHash(nameFile, rxHash string) error {

	// Вычисление хэша у переданного файла.
	hash, err := hashFile(nameFile)
	if err != nil {
		return fmt.Errorf("функция hashFile, вернула ошибку: <%w>", err)
	}

	// Проверка соответствия
	if hash != rxHash {
		return fmt.Errorf("нет соответствия хэш. Ожидалось:<%s>, а принято:<%s>", hash, rxHash)
	}

	return nil
}

// Удаление файла.
func layerSendFileRemove(filePath string) error {

	if !isFileExists(filePath) {
		return fmt.Errorf("отсутствует удаляемый файл:<%s>", filePath)
	}

	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("ошибка удаления файла:<%w>", err)
	}

	return nil
}

//
// --- RequestLoginPasswordNames ---
//

// Передача запроса
func layerRequestLoginPasswordNamesTx(client proto.PasswordManagerClient, tokenAuth, idClient string) (rxData []string, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", tokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Запрос у сервера информации.
	emptyRequest := &emptypb.Empty{}
	resp, err := client.RequestLoginPasswordName(
		ctx,
		emptyRequest,
	)
	if err != nil {
		return nil, fmt.Errorf("Функция client.RequestLoginPasswordName, вернула ошибку: <%v>", err)
	}

	// Получение данных ответа.
	for _, v := range resp.EntriesName {
		rxData = append(rxData, v)
	}

	// Результат.
	return rxData, nil
}

// Расшифровка принятых данных
func layerRequestLoginPasswordNamesDecrypt(enRxData []string, key [32]byte) (rxData []string, err error) {

	// Проверка
	if len(enRxData) == 0 {
		return []string{}, nil
	}

	// Расшифровка
	for _, v := range enRxData {
		d, err := decrypt(v, key)
		if err != nil {
			return []string{}, fmt.Errorf("Функция decrypt, вернула ошибку: <%w>", err)
		}
		rxData = append(rxData, d)
	}

	// Результат
	return rxData, nil
}

//
// --- RequestLoginPasswordByName ---
//

// Шифрование передаваемых данных
func LayerRequestLoginPasswordByNameEncrypt(nameEntry string, key [32]byte) (enNameEntry string, err error) {

	// Шифрование имени записи
	enNameEntry, err = encrypt(nameEntry, key)
	if err != nil {
		return "", fmt.Errorf("функция encrypt, вернула ошибку: <%w>", err)
	}

	return enNameEntry, nil
}

// Запрос к серверу
func LayerRequestLoginPasswordByNameTx(client proto.PasswordManagerClient, tokenAuth, idClient, enNameEntry string) (resp *pb.RequestLoginPasswordByNameResponse, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных
	md := metadata.Pairs("token", tokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Данные запроса
	req := &proto.RequestLoginPasswordByNameRequest{
		IdClient: idClient,
		Name:     enNameEntry,
	}

	// Запрос
	resp, err = client.RequestLoginPasswordByName(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Функция RequestLoginPasswordByName, вернула ошибку: <%w>", err)
	}

	// Результат
	return resp, nil
}

// Обработка ответа
func LayerRequestLoginPasswordByDecrypt(resp *pb.RequestLoginPasswordByNameResponse, key [32]byte) (rxData RxLoginPassword, err error) {

	rxData.For, err = decrypt(resp.Name, key)
	if err != nil {
		return RxLoginPassword{}, fmt.Errorf("ошибка расшифровки Name:<%w>", err)
	}

	rxData.Login, err = decrypt(resp.Login, key)
	if err != nil {
		return RxLoginPassword{}, fmt.Errorf("ошибка расшифровки Login:<%w>", err)
	}

	rxData.Password, err = decrypt(resp.Password, key)
	if err != nil {
		return RxLoginPassword{}, fmt.Errorf("ошибка расшифровки Password:<%w>", err)
	}

	rxData.CreatedAt, err = decrypt(resp.CreatedAt, key)
	if err != nil {
		return RxLoginPassword{}, fmt.Errorf("ошибка расшифровки CreatedAt:<%w>", err)
	}

	return rxData, nil
}

//
// --- RequestTextNames ---
//

// Передача запроса
func layerRequestTextNamesTx(client proto.PasswordManagerClient, tokenAuth, idClient string) (rxData []string, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", tokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Запрос у сервера информации.
	emptyRequest := &emptypb.Empty{}
	resp, err := client.RequestTextName(
		ctx,
		emptyRequest,
	)
	if err != nil {
		return nil, fmt.Errorf("Функция client.RequestTextName, вернула ошибку: <%v>", err)
	}

	// Получение данных ответа.
	for _, v := range resp.EntriesName {
		rxData = append(rxData, v)
	}

	// Результат.
	return rxData, nil
}

// Расшифровка принятых данных
func layerRequestTextNamesDecrypt(enRxData []string, key [32]byte) (rxData []string, err error) {

	// Проверка
	if len(enRxData) == 0 {
		return []string{}, nil
	}

	// Расшифровка
	for _, v := range enRxData {
		d, err := decrypt(v, key)
		if err != nil {
			return []string{}, fmt.Errorf("Функция decrypt, вернула ошибку: <%w>", err)
		}
		rxData = append(rxData, d)
	}

	// Результат
	return rxData, nil
}

//
// --- RequestTextByName ---
//

// Шифрование передаваемых данных
func LayerRequestTextByNameEncrypt(nameEntry string, key [32]byte) (enNameEntry string, err error) {

	// Шифрование имени записи
	enNameEntry, err = encrypt(nameEntry, key)
	if err != nil {
		return "", fmt.Errorf("функция encrypt, вернула ошибку: <%w>", err)
	}

	return enNameEntry, nil
}

// Запрос к серверу
func LayerRequestTextByNameTx(client proto.PasswordManagerClient, tokenAuth, idClient, enNameEntry string) (resp *pb.RequestTextByNameResponse, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных
	md := metadata.Pairs("token", tokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Данные запроса
	req := &proto.RequestTextByNameRequest{
		IdClient: idClient,
		Name:     enNameEntry,
	}

	// Запрос
	resp, err = client.RequestTextByName(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Функция RequestTextByName, вернула ошибку: <%w>", err)
	}

	// Результат
	return resp, nil
}

// Обработка ответа
func LayerRequestTextByNameDecrypt(resp *pb.RequestTextByNameResponse, key [32]byte) (rxData RxText, err error) {

	rxData.For, err = decrypt(resp.Name, key)
	if err != nil {
		return RxText{}, fmt.Errorf("ошибка расшифровки Name:<%w>", err)
	}

	rxData.Text, err = decrypt(resp.Text, key)
	if err != nil {
		return RxText{}, fmt.Errorf("ошибка расшифровки Text:<%w>", err)
	}

	rxData.CreatedAt, err = decrypt(resp.CreatedAt, key)
	if err != nil {
		return RxText{}, fmt.Errorf("ошибка расшифровки CreatedAt:<%w>", err)
	}

	return rxData, nil
}

//
// --- RequestBankCardNames ---
//

// Передача запроса
func layerRequestBankCardNamesTx(client proto.PasswordManagerClient, tokenAuth, idClient string) (rxData []string, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", tokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Запрос у сервера информации.
	emptyRequest := &emptypb.Empty{}
	resp, err := client.RequestBankCardName(
		ctx,
		emptyRequest,
	)
	if err != nil {
		return nil, fmt.Errorf("Функция client.RequestBankCardName, вернула ошибку: <%v>", err)
	}

	// Получение данных ответа.
	for _, v := range resp.EntriesName {
		rxData = append(rxData, v)
	}

	// Результат.
	return rxData, nil
}

// Расшифровка принятых данных
func layerRequestBankCardNamesDecrypt(enRxData []string, key [32]byte) (rxData []string, err error) {

	// Проверка
	if len(enRxData) == 0 {
		return []string{}, nil
	}

	// Расшифровка
	for _, v := range enRxData {
		d, err := decrypt(v, key)
		if err != nil {
			return []string{}, fmt.Errorf("Функция decrypt, вернула ошибку: <%w>", err)
		}
		rxData = append(rxData, d)
	}

	// Результат
	return rxData, nil
}

//
// --- RequestBankCardByName ---
//

// Шифрование передаваемых данных
func LayerRequestBankCardByNameEncrypt(nameEntry string, key [32]byte) (enNameEntry string, err error) {

	// Шифрование имени записи
	enNameEntry, err = encrypt(nameEntry, key)
	if err != nil {
		return "", fmt.Errorf("функция encrypt, вернула ошибку: <%w>", err)
	}

	return enNameEntry, nil
}

// Запрос к серверу
func LayerRequestBankCardByNameTx(client proto.PasswordManagerClient, tokenAuth, idClient, enNameEntry string) (resp *pb.RequestBankCardByNameResponse, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных
	md := metadata.Pairs("token", tokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Данные запроса
	req := &proto.RequestBankCardByNameRequest{
		IdClient: idClient,
		Name:     enNameEntry,
	}

	// Запрос
	resp, err = client.RequestBankCardByName(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Функция RequestBankCardByName, вернула ошибку: <%w>", err)
	}

	// Результат
	return resp, nil
}

// Обработка ответа
func LayerRequestBankCardByNameDecrypt(resp *pb.RequestBankCardByNameResponse, key [32]byte) (rxData RxBankCard, err error) {

	rxData.For, err = decrypt(resp.Name, key)
	if err != nil {
		return RxBankCard{}, fmt.Errorf("ошибка расшифровки Name:<%w>", err)
	}

	rxData.Owner, err = decrypt(resp.Owner, key)
	if err != nil {
		return RxBankCard{}, fmt.Errorf("ошибка расшифровки Owner:<%w>", err)
	}

	rxData.Numb, err = decrypt(resp.Numb, key)
	if err != nil {
		return RxBankCard{}, fmt.Errorf("ошибка расшифровки Numb:<%w>", err)
	}

	rxData.Valid, err = decrypt(resp.Valid, key)
	if err != nil {
		return RxBankCard{}, fmt.Errorf("ошибка расшифровки Valid:<%w>", err)
	}

	rxData.Code, err = decrypt(resp.Code, key)
	if err != nil {
		return RxBankCard{}, fmt.Errorf("ошибка расшифровки Code:<%w>", err)
	}

	rxData.CreatedAt, err = decrypt(resp.CreatedAt, key)
	if err != nil {
		return RxBankCard{}, fmt.Errorf("ошибка расшифровки CreatedAt:<%w>", err)
	}

	return rxData, nil
}

//
// --- RequestFileNames ---
//

// Передача запроса
func layerRequestFileNamesTx(client proto.PasswordManagerClient, tokenAuth, idClient string) (rxData []string, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", tokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Запрос у сервера информации.
	emptyRequest := &emptypb.Empty{}
	resp, err := client.RequestFileName(
		ctx,
		emptyRequest,
	)
	if err != nil {
		return nil, fmt.Errorf("Функция client.RequestTextName, вернула ошибку: <%v>", err)
	}

	// Получение данных ответа.
	for _, v := range resp.EntriesName {
		rxData = append(rxData, v)
	}

	// Результат.
	return rxData, nil
}

//
// --- RequestFileInfo ---
//

// Передача запроса
func layerRequestFileInfoTx(client proto.PasswordManagerClient, tokenAuth, idClient, nameFile string) (fileName, fileHash string, fileSize int64, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", tokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Запрос.
	req := &pb.RequestFileInfoRequest{
		IdClient: idClient,
		Name:     nameFile,
	}
	resp, err := client.RequestFileInfo(
		ctx,
		req,
	)
	if err != nil {
		return "", "", 0, fmt.Errorf("Функция client.RequestFileInfo, вернула ошибку: <%v>", err)
	}

	// Получение данных ответа.
	fileName = resp.Name
	fileHash = resp.Hash
	fileSize = resp.Size

	// Результат.
	return fileName, fileHash, fileSize, nil
}

// Проверка имён.
func layerRequestFileInfoCheck(fileName, rxFileName string) error {

	if fileName != rxFileName {
		return fmt.Errorf("Нет соответствия имён файлов. Ожидалось:<%s>, а принято:<%s>", fileName, rxFileName)
	}

	return nil
}

//
// --- RequestFileByName ---
//

// Приём файла.
func layerRequestFileByNameRx(s *server, data *dataRequestFile, chProcess chan<- float32) (err error) {

	needRestoreState := false // Признак необходимости воостановления состояния, при ошибке.
	var tempFileName string   // Имя временного файла.

	// Обработка перед выходом.
	defer func(needRestoreState bool, fileName, tempFileName string, errProcess error) {
		if errRestore := deferProcessRestoreByError(needRestoreState, fileName, tempFileName, errProcess); errRestore != nil {
			err = fmt.Errorf("функция deferProcessRestoreByError, вернула ошибку:<%w>, при ошибку процесса:<%w>", errRestore, errProcess)
		}
	}(needRestoreState, data.fileName, tempFileName, err)

	// Запрос файла у сервера.
	srcFileHash, err := requestFile(s, data, chProcess)
	if err != nil {
		return fmt.Errorf("Функция requestFile, вернула ошибку: <%w>", err)
	}

	// Вычисление хэша принятого файла.
	rxFileHash, err := hashFile(data.fileName)
	if err != nil {
		return fmt.Errorf("функция hashFile, вернула ошибку: <%w>", err)
	}

	// Проверка результата.
	if err := checkResultRequestFile(data.fileName, rxFileHash, srcFileHash); err != nil {

		// Удаление файла, если проверка не пройдена.
		if errRemove := os.Remove(data.fileName); errRemove != nil {
			return fmt.Errorf("ошибка:<%w> удаления файла:<%s>, после приёма. Базовая ошибка:<%w>", errRemove, data.fileName, err)
		}
		return fmt.Errorf("функция layerRestoreCheckResult, вернула ошибку:<%w>, для файла:<%s>", err, data.fileName)
	}

	return nil
}

// Изменение имени существующего файла.
func layerRequestFileByNameCreateTemp(fileName string) (tempFileName string, err error) {

	if isFileExists(fileName) {
		tempFileName, err = changeFileName(fileName, "-temp")
		if err != nil {
			return "", fmt.Errorf("функция changeFileName, вернула ошибку:<%w>", err)
		}
	}

	return tempFileName, nil
}

// Удаление временного файла.
func layerRequestFileByNameRemoveTemp(tempFileName string) error {

	if isFileExists(tempFileName) {
		if err := os.Remove(tempFileName); err != nil {
			return fmt.Errorf("ошибка:<%w> удаления резервного файла:<%s>, при успешном приёме", err, tempFileName)
		}
	}

	return nil
}

// Расшифровка файла.
func layerRequestFileByNameDecrypt(encFilePath string, key [32]byte) (err error) {

	defer func(encFilePath string) {
		if errDel := deleteFile(encFilePath); errDel != nil {
			errDel = fmt.Errorf("Ошибка: <%w> при удалении файла: <%s>. Базовая ошибка:<%v>", errDel, encFilePath, err)
		}
	}(encFilePath)

	encFile, err := os.Open(encFilePath)
	if err != nil {
		return err
	}
	defer encFile.Close()

	// Создание расшифрованного файла
	decFileExt := path.Ext(encFilePath)
	decFileName := path.Base(encFilePath)
	decFileName = strings.TrimSuffix(decFileName, decFileExt)
	decFileName = strings.TrimSuffix(decFileName, "-enc")

	decFileName = decFileName + decFileExt

	// Предварительное удаление файла, если существует.
	if isFileExists(decFileName) {
		if err := deleteFile(decFileName); err != nil {
			return fmt.Errorf("Ошибка при удалении файла: <%s>", decFileName)
		}
	}

	outputFile, err := os.Create(decFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	// Читаем IV (первые 16 байт)
	iv := make([]byte, aes.BlockSize)
	n, err := encFile.Read(iv)
	if err != nil || n != aes.BlockSize {
		return fmt.Errorf("не удалось прочитать IV")
	}

	stream := cipher.NewCBCDecrypter(block, iv)

	buffer := make([]byte, 4096)
	var leftover []byte

	for {
		n, err := encFile.Read(buffer)
		if n == 0 {
			if err == io.EOF {
				break
			}
			return err
		}

		chunk := append(leftover, buffer[:n]...)

		// Если данные не кратны 16 Б, оставляем остаток
		remainder := len(chunk) % aes.BlockSize
		if remainder != 0 {
			leftover = chunk[len(chunk)-remainder:]
			chunk = chunk[:len(chunk)-remainder]
		} else {
			leftover = nil
		}

		if len(chunk) == 0 {
			continue
		}

		// Расшифровываем
		plaintext := make([]byte, len(chunk))
		stream.CryptBlocks(plaintext, chunk)

		// Пишем расшифрованные данные
		if _, err := outputFile.Write(plaintext); err != nil {
			return err
		}
	}

	// Обрабатываем остаток (последний блок с паддингом)
	if len(leftover) > 0 {
		plaintext := make([]byte, len(leftover))
		stream.CryptBlocks(plaintext, leftover)

		// Удаляем PKCS#7 паддинг
		padding := plaintext[len(plaintext)-1]
		if padding == 0 || int(padding) > aes.BlockSize || len(plaintext) < int(padding) {
			return ErrInvalidPadding
		}

		for i := len(plaintext) - int(padding); i < len(plaintext); i++ {
			if plaintext[i] != padding {
				return ErrInvalidPadding
			}
		}

		plaintext = plaintext[:len(plaintext)-int(padding)]
		if _, err := outputFile.Write(plaintext); err != nil {
			return err
		}
	}

	return nil
}
