package server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
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
		return "", fmt.Errorf("Ошибка подключения к файлу: <%w>", err)
	}
	defer inputFile.Close()

	// Создание имени для шифрованной версии файла.
	extension := filepath.Ext(filePath)
	baseName := filepath.Base(filePath[:len(filePath)-len(extension)])
	fileName := baseName + "-enc" + extension

	encFilePath = filepath.Join(filepath.Dir(filePath), fileName)

	// Проверка существования файла.
	if isFileExists(encFilePath) {
		if err := deleteFile(encFilePath); err != nil {
			return "", fmt.Errorf("Ошибка удаления существующего файла: <%w>", err)
		}
	}

	// Создание шифрованного файла.
	encFile, err := os.Create(encFilePath)
	if err != nil {
		return "", fmt.Errorf("Ошибка создания зашифрованного файла: <%w>", err)
	}
	defer encFile.Close()

	// Создание шифра AES.
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("Ошибка создания AES шифра: <%w>", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("Ошибка генерации IV: <%w>", err)
	}

	cipherStream := cipher.NewCBCEncrypter(block, iv)

	// Запись IV в начало файла.
	if _, err := encFile.Write(iv); err != nil {
		return "", fmt.Errorf("Ошибка записи IV: <%w>", err)
	}

	buffer := make([]byte, aes.BlockSize)
	for {
		n, err := inputFile.Read(buffer)
		if n == 0 {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("Ошибка чтения: <%w>", err)
		}

		// Добавление паддинга PKCS#7.
		pad := aes.BlockSize - n%aes.BlockSize
		paddedData := append(buffer[:n], bytes.Repeat([]byte{byte(pad)}, pad)...)

		ciphertext := make([]byte, len(paddedData))
		cipherStream.CryptBlocks(ciphertext, paddedData)

		if _, err := encFile.Write(ciphertext); err != nil {
			return "", fmt.Errorf("Ошибка записи зашифрованных данных: <%w>", err)
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
	if resp.FileName != fileName {
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
// --- ReceiveFile ---
//

// Расшифровка файла.
func layerReceiveFileDecrypt(encFilePath string, key [32]byte) error {

	encFile, err := os.Open(encFilePath)
	if err != nil {
		return fmt.Errorf("Ошибка открытия файла: <%w>", err)
	}
	defer func() {
		if cerr := encFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("Ошибка закрытия encFile: <%w>", cerr)
		}
	}()

	// Чтение IV (16 байт)
	iv := make([]byte, aes.BlockSize)
	n, err := encFile.Read(iv)
	if err != nil && err != io.EOF {
		return fmt.Errorf("Ошибка чтения IV: <%w>", err)
	}
	if n != aes.BlockSize {
		return io.ErrUnexpectedEOF
	}

	// Создание AES-шифра
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return fmt.Errorf("Ошибка создания шифра: <%w>", err)
	}
	cipherStream := cipher.NewCBCDecrypter(block, iv)

	// Буферы
	const (
		base64Chunk = 4 * 1024 // 4 КБ
		plainChunk  = 3 * 1024 // 3 КБ
	)

	encodedBuf := make([]byte, base64Chunk)
	decodedBuf := make([]byte, plainChunk)
	cipherBuf := make([]byte, plainChunk)

	var leftover []byte

	// Создание расшифрованного файла
	ext := filepath.Ext(encFilePath)
	decFileName := strings.TrimSuffix(filepath.Base(encFilePath), "-enc"+ext)
	decFilePath := filepath.Join(filepath.Dir(encFilePath), decFileName+ext)

	decFile, err := os.Create(decFilePath)
	if err != nil {
		return fmt.Errorf("Ошибка создания файла: <%w>", err)
	}
	defer func() {
		if cerr := decFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("Ошибка закрытия decFile: <%w>", cerr)
		}
	}()

	for {
		// Чтение base64-данных
		n, err := encFile.Read(encodedBuf)
		if n == 0 {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("Ошибка чтения: <%w>", err)
		}

		// Объединяем с остатком предыдущего чтения
		chunk := append(leftover, encodedBuf[:n]...)

		// Убираем остаток
		remainder := len(chunk) % 4
		if remainder != 0 {
			leftover = chunk[len(chunk)-remainder:]
			chunk = chunk[:len(chunk)-remainder]
		} else {
			leftover = nil
		}

		// Декодирование base64
		decodedLen, err := base64.StdEncoding.Decode(decodedBuf, chunk)
		if err != nil {
			return fmt.Errorf("Ошибка base64-декодирования: <%w>", err)
		}

		// Расшифровка блока
		cipherStream.CryptBlocks(cipherBuf[:decodedLen], decodedBuf[:decodedLen])

		// Запись в файл
		if _, err := decFile.Write(cipherBuf[:decodedLen]); err != nil {
			return fmt.Errorf("Ошибка записи: <%w>", err)
		}
	}

	// Проверка остатка
	if len(leftover) > 0 {
		return fmt.Errorf("Неполный base64-блок в конце файла")
	}

	// Удаление паддинга
	if true {
		fileInfo, err := decFile.Stat()
		if err != nil {
			return fmt.Errorf("Ошибка Stat: <%w>", err)
		}
		fileSize := fileInfo.Size()

		if fileSize > 0 {
			paddingByte := make([]byte, 1)
			n, err := decFile.ReadAt(paddingByte, fileSize-1)
			if err != nil || n != 1 {
				return fmt.Errorf("Ошибка чтения последнего байта: <%w>", err)
			}

			padding := paddingByte[0]
			if int(padding) > aes.BlockSize || padding == 0 || fileSize < int64(padding) {
				return ErrInvalidPadding
			}

			if err := decFile.Truncate(fileSize - int64(padding)); err != nil {
				return fmt.Errorf("Ошибка Truncate: <%w>", err)
			}
		}
	}

	// Удаление зашифрованного файла
	return os.Remove(encFilePath)
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

// RequestLoginPasswordByName(ctx context.Context, tokenAuth, idClient, nameEntry string, key [32]byte)

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
