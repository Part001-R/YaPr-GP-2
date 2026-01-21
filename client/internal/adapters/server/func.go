// Вспомогательные функции пакета.
package server

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Подключение к серверу. Возвращается подключение, клиент и ошибка.
//
// Параметры:
//
//	ip - ip адрес.
//	port - номер порта.
func connect(ip, port string) (conn *grpc.ClientConn, client pb.PasswordManagerClient, err error) {

	// Проверка аргументов
	if ip == "" || port == "" {
		return nil, nil, nil // Возвращается nil по ошибке, т.к. в локальном режиме нет постоянной необходимости в подключении.
	}

	//
	// Логика
	//

	// Настройка TLS.
	creds, err := credentials.NewClientTLSFromFile("tls/server.crt", "")
	if err != nil {
		return nil, nil, fmt.Errorf("функция credentials.NewClientTLSFromFile, вернула ошибку: <%w>", err)
	}

	// Подключение к серверу.
	srvAddr := ip + ":" + port

	conn, err = grpc.NewClient(srvAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, nil, fmt.Errorf("функция grpc.NewClient, вернула ошибку: <%w>", err)
	}

	// создание клиента.
	client = pb.NewPasswordManagerClient(conn)

	return conn, client, nil
}

// Создание токена для регистрации пользователя в режиме  - удалённый. Возвращаются передаваемые метаданные, секретный ключ, имя токена и ошибка.
func createTokenForAuthentication() (txMD metadata.MD, secretKey, nameToken string, err error) {

	// Создание ключа.
	secretKey, err = generateRandomString(50)
	if err != nil {
		return nil, "", "", fmt.Errorf("функция generateRandomString, вернула ошибку: <%w>", err)
	}

	// Создание токена.
	timeValidToken := time.Duration(5 * time.Second)
	txToken, err := createToken("clientManager", secretKey, timeValidToken)
	if err != nil {
		return nil, "", "", fmt.Errorf("функция createToken, вернула ошибку: <%w>", err)
	}

	// Заполнение метаданных.
	nameToken = "token"
	txMD = metadata.Pairs(nameToken, txToken)

	return txMD, secretKey, nameToken, nil
}

// Генерация строки из случайных символов, заданной длинны. Возарвщается строка и ошибка.
//
// Параметры:
//
//	length - уставка длинны строки.
func generateRandomString(length int) (string, error) {

	// Проверка аргументов.
	if length <= 0 {
		return "", fmt.Errorf("длина должна быть больше 0")
	}

	// Логика.
	//
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength := len(charset)

	// Массив для хранения случайных индексов символов.
	bytes := make([]byte, length)

	// Заполнение массива случайными индексами.
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Преобразование случайных байтов в индексы символов из charset.
	for i := 0; i < length; i++ {
		bytes[i] = charset[int(bytes[i])%charsetLength]
	}

	// Рузультат.
	return string(bytes), nil
}

// Проверка токена. Возвращается ошибка.
//
// Параметры:
//
//	tokenStr - проверяемый токен.
//	secretKey - секретный ключ.
func checkToken(tokenStr string, secretKey string) error {

	// Проверка аргументов.
	if tokenStr == "" {
		return MissingDataArgumentTokenStr
	}
	if len(secretKey) < 32 {
		return LenSecretKey
	}

	// Логика.
	//
	// Проверка токена.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, SigningMethodUnknown
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return fmt.Errorf("ошибка парсинга токена: <%w>", err)
	}

	// Проверка на истечение срока действия токена.
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && claims.ExpiresAt != nil {
		if time.Now().After(claims.ExpiresAt.Time) {
			return ValidTokenExpired
		}
	}

	// Проверка пройдена.
	return nil
}

// Создание JWT токена.
//
// Параметры:
//
//	subjectName - имя, для кого выдаётся токен.
//	secretKey - секретный ключ.
//	validTime - время валидности токена.
func createToken(subjectName, secretKey string, validTime time.Duration) (string, error) {

	// Проверка аргументов.
	if subjectName == "" {
		return "", EmptyDataArgumentSubjectName
	}
	if len(secretKey) < 32 {
		return "", LenSecretKey
	}

	// Логика.
	//
	// Создание токена.
	expirationTime := time.Now().Add(validTime)

	claims := &jwt.RegisteredClaims{
		Subject:   subjectName,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подпись токена.
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("ошибка при подписи токена: <%w>", err)
	}

	// Результат.
	return tokenString, nil
}

// Шифрование данных. Возвращается результат шифрования и ошибка.
//
// Параметры:
//
//	data - данные для шифрования.
//	key - ключ шифрования.
func encrypt(data string, key [32]byte) (string, error) {

	// Создание AES-256
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	plaintextBytes := []byte(data)

	// PKCS#7 заполнение
	blockSize := block.BlockSize()
	padding := blockSize - len(plaintextBytes)%blockSize
	padText := append(plaintextBytes, bytes.Repeat([]byte{byte(padding)}, padding)...)

	// Вектор инициализации (нулевой — для воспроизводимости)
	iv := make([]byte, blockSize)

	// Шифрование
	ciphertext := make([]byte, len(padText))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padText)

	// Кодирование в Base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Расшифровка данных. Возвращается результат расшифровки и ошибка.
//
// Параметры:
//
//	data - зашифрованные данные.
//	key - ключ шифрования.
func decrypt(data string, key [32]byte) (string, error) {

	// Декодирование из Base64
	ciphertextBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("Функция DecodeString, вернула ошибку:<%w>", err)
	}

	// Создание AES-256
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("Функция NewCipher, вернула ошибку:<%w>", err)
	}

	blockSize := block.BlockSize()
	if len(ciphertextBytes)%blockSize != 0 {
		return "", NotCorrectLenData
	}

	// Вектор инициализации (такой же, как при шифровании)
	iv := make([]byte, blockSize)

	// Расшифровка
	plaintextPadded := make([]byte, len(ciphertextBytes))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintextPadded, ciphertextBytes)

	// Удаление PKCS#7
	padding := int(plaintextPadded[len(plaintextPadded)-1])
	if padding > blockSize || padding == 0 {
		return "", NotCorrectDataFill
	}
	plaintext := plaintextPadded[:len(plaintextPadded)-padding]

	return string(plaintext), nil
}

// Проверка существования файла. Возвращается true - если файл существует.
//
// Параметры:
//
//	filePath - путь к файлу.
func isFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

// Функция для удаления файла. Возвращается ошибка.
//
// Параметры:
//
//	filePath - путь к файлу.
func deleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("не удалось удалить файл <%s>: <%w>", filePath, err)
	}
	return nil
}

// Вычисление хэша у файла. Возвращается хэш и ошибка.
//
// Параметры:
//
//	fileName - имя файла.
func hashFile(fileName string) (string, error) {

	file, err := os.Open(fileName)
	if err != nil {
		return "", fmt.Errorf("ошибка при открытии файла: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("ошибка при вычислении хэша: %w", err)
	}

	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}

// Воостановление состояния файлов процесса restore, при ошибке в последовательности. Возвращается ошибка.
//
// Параметры:
//
//	doRestore - выполнить восстановление.
//	fileName - имя файла.
//	tempFileName - имя временного файла.
//	errProcess - основная ошибка логики.
func deferProcessRestoreByError(doRestore bool, fileName, tempFileName string, errProcess error) error {

	if errProcess != nil {
		// Удаление принятого файла.
		if isFileExists(fileName) && doRestore {
			if err := os.Remove(fileName); err != nil {
				return fmt.Errorf("Error: ошибка:<%v>, при удалении файла:<%s> по ошибке процесса:<%v>", err, fileName, err)
			}
		}
		// Восстановление имени у исходного файла.
		if isFileExists(tempFileName) && doRestore {
			if err := restoreFileName(tempFileName, "-temp"); err != nil {
				return fmt.Errorf("Error: ошибка:<%v>, при восстановлении файла:<%s> по ошибке процесса:<%v>", err, tempFileName, err)
			}
		}
	}
	return nil
}

// Воссстановление имени файла. Возвращается ошибка.
//
// Параметры:
//
//	fileName - имя файла.
//	suffix - постфикс.
func restoreFileName(fileName string, suffix string) (err error) {

	// Проверка, существует ли файл.
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return fmt.Errorf("файл с именем:<%s>, не найден", fileName)
	}

	// Исходные данные.
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)

	// Восстановление
	name = strings.TrimSuffix(name, suffix)

	// Формирование нового имени.
	newFileName := fmt.Sprintf("%s%s", name, ext)

	// Переименование файла.
	err = os.Rename(fileName, newFileName)
	if err != nil {
		return fmt.Errorf("ошибка:<%w> переименования файла:<%s>", err, fileName)
	}

	return nil
}

// Приём файла. Возвращается хэш принятого файла и ошибка.
//
// Параметры:
//
//	s - указатель на сервер.
//	data - данные для процесса.
//	chProcess - канал для передачи процентов процесса.
func requestFile(s *server, data *DataRequestFile, chProcess chan<- float32) (rxFileHash string, err error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Установка метаданных с токеном
	md := metadata.Pairs("token", data.TokenAuth)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Запрос
	req := &pb.RequestFileByNameRequest{
		IdClient: data.ClientID,
		FileName: data.FileName,
	}
	stream, err := s.client.RequestFileByName(ctx, req)
	if err != nil {
		return "", fmt.Errorf("Функция client.RequestFileByName, вернула ошибку: <%w>", err)
	}

	// Предварительное удаление, если такой файл уже существует.
	if isFileExists(data.FileName) {
		if err := os.Remove(data.FileName); err != nil {
			return "", fmt.Errorf("ошибка удаления файла: <%w>", err)
		}
	}

	// Создание файла для записи
	file, err := os.Create(data.FileName)
	if err != nil {
		return "", fmt.Errorf("не удалось создать файл <%s>: <%w>", data.FileName, err)
	}
	defer file.Close()

	// Чтение потоком
	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("Функция stream.Recv, вернула ошибку: <%w>", err)
		}
		if res.FileName != data.FileName {
			return "", fmt.Errorf("Приняты данные для другого файла: <%s>", res.FileName)
		}

		// Обновление статистики процесса
		updateDataRxProcess(s, chProcess, len(res.Content), data)

		// Запись данных в файл
		if _, err := file.Write(res.Content); err != nil {
			return "", fmt.Errorf("ошибка при записи в файл <%s>: <%w>", data.FileName, err)
		}
	}

	// Получение трейлера после завершения потока
	rxTrailer := stream.Trailer()
	if hash, ok := rxTrailer["hash"]; ok {
		rxFileHash = hash[0]
	} else {
		return "", fmt.Errorf("Сервер не предоставил трейлер с данными хэша, для файла: <%s>", data.FileName)
	}

	// Результат
	return rxFileHash, nil
}

// Добавление превикса к имени имени файла. Возвращается новое имя и ошибка.
//
// Параметры:
//
//	fileName - имя файла.
//	suffix - суффикс.
func changeFileName(fileName string, suffix string) (newFileName string, err error) {

	// Проверка, существует ли файл.
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return "", fmt.Errorf("файл с именем:<%s>, не найден", fileName)
	}

	// Исходные данные.
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)

	// Новое имя.
	newFileName = fmt.Sprintf("%s%s%s", name, suffix, ext)

	// Проверка существование файла по новому имени.
	// Если есть - удаляется.
	_, err = os.Stat(newFileName)
	if err == nil {
		if err := os.Remove(newFileName); err != nil {
			return "", fmt.Errorf("ошибка:<%w> удаления резервного файла:<%s>", err, newFileName)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("ошибка:<%w> при проверке существования файла:<%s>", err, newFileName)
	}

	// Переименование файла.
	err = os.Rename(fileName, newFileName)
	if err != nil {
		return "", fmt.Errorf("ошибка:<%w> переименования файла:<%s>", err, fileName)
	}

	return newFileName, nil
}

// Проверка ответа от сервера. Возвращается ошибка.
//
// Параметры:
//
//	fileName - имя файла.
//	rxFileHash - хэш принятого файла.
//	srcFileHash - хэш исходного файла.
func checkResultRequestFile(fileName, rxFileHash, srcFileHash string) error {

	// Проверка аргументов.
	if fileName == "" {
		return EmptyDataArgumentName
	}
	if rxFileHash == "" {
		return EmptyDataArgumentRxFileHash
	}
	if srcFileHash == "" {
		return EmptyDataArgumentSrcFileHash
	}

	// Анализ данных ответа от сервера.
	rxFileHash, err := hashFile(fileName)
	if err != nil {
		return fmt.Errorf("Функция hashFile, вернула ошибку:<%w>", err)
	}
	if srcFileHash != rxFileHash {
		return fmt.Errorf("Для файла:<%s>, нет соответствия хэша. Ожидался:<%s>, а принято:<%s>", fileName, srcFileHash, rxFileHash)
	}

	return nil
}

// Вычисление процента выполнения.
//
// Параметры:
//
//	s - указатель на сервер.
//	chProcess - канал передачи процентов процесса.
//	b - количество байт.
//	data - данные процесса.
func updateDataRxProcess(s *server, chProcess chan<- float32, b int, data *DataRequestFile) {

	s.mtx.processRxFile.Lock()
	defer s.mtx.processRxFile.Unlock()

	// Получение КБайт из Байт.
	volumeKB := b / 1024

	// Обновление данных накопителя.
	data.SizePassed += int64(volumeKB)

	// Вычисление процентов.
	if data.SizeReqFile > 0 {
		chProcess <- float32(float64(data.SizePassed) / float64(data.SizeReqFile) * 100.0)
	} else {
		chProcess <- 0
	}
}

// Вычисление процента выполнения.
//
// Параметры:
//
//	s - указатель на сервер.
//	chProcess - канал передачи процентов процесса.
//	b - количество байт.
//	data - данные процесса.
func updateDataTxProcess(s *server, chProcess chan<- float32, b int, data *dataSendFile) {

	s.mtx.processTxFile.Lock()
	defer s.mtx.processTxFile.Unlock()

	// Получение КБайт из Байт.
	volumeKB := b / 1024

	// Обновление данных накопителя.
	data.sizePassed += int64(volumeKB)

	// Вычисление процентов.
	if data.sizeSendFile > 0 {
		chProcess <- float32(float64(data.sizePassed) / float64(data.sizeSendFile) * 100.0)
	} else {
		chProcess <- 0
	}
}

// Вычисление процента выполнения.
//
// Параметры:
//
//	s - указатель на сервер.
//	chProcess - канал передачи процентов процесса.
//	b - количество байт.
func updateDataBackUpRestoreProcess(s *server, chProcess chan<- float32, b int) {

	s.mtx.processBackUpRestore.Lock()
	defer s.mtx.processBackUpRestore.Unlock()

	// Получение КБайт из Байт.
	volumeKB := b / 1024

	// Обновление данных накопителя.
	s.dataRestore.sizePassed += int64(volumeKB)

	// Вычисление процентов.
	if s.dataRestore.totalSizeFiles > 0 {
		chProcess <- float32(float64(s.dataRestore.sizePassed) / float64(s.dataRestore.totalSizeFiles) * 100.0)
	} else {
		chProcess <- 0
	}
}

// Проверка номера алгоритмом Luhn. Возвращается true - проверка пройдена.
//
// Параметры:
//
//	cardNumber - номер.
func isCheckByLuhn(number string) bool {

	// Удаление лишних символов
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, number)

	if len(cleaned) == 0 {
		return false
	}

	var sum int
	length := len(cleaned)

	// Проход справа налево
	for i := 0; i < length; i++ {
		digit := int(cleaned[length-1-i] - '0')

		if i%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}

// Сброс экземпляра сервера, для тестов.
func resetInstServer() {

	// Определяем место вызова
	_, callerFile, _, ok := runtime.Caller(1)
	if !ok {
		panic("не удалось получить информацию о месте вызова")
	}

	// Является ли вызывающий файл тестовым
	if !strings.HasSuffix(callerFile, "_test.go") {
		return
	}

	inst = nil
}
