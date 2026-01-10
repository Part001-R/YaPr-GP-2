package server

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Подключение к серверу.
func connect(ip, port string) (conn *grpc.ClientConn, client pb.PasswordManagerClient, err error) {

	// Проверка аргументов
	if ip == "" {
		return nil, nil, NilPtrArgumentIP
	}
	if port == "" {
		return nil, nil, NilPtrArgumentPort
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

// Создание токена для регистрации пользователя в режиме  - удалённый.
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
		return "", err
	}

	// Создание AES-256
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
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

// Проверка существования файла
func isFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

// Функция для удаления файла
func deleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("не удалось удалить файл <%s>: <%w>", filePath, err)
	}
	return nil
}

// Вычисление хэша у файла.
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
