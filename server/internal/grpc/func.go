// Вспомогательные функции пакета.
package grpc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Определение размера файла в байтах. Возвращается размер в байтах и ошибка.
//
// Параметры:
//
//	fileName - имя файла.
func sizeFile(fileName string) (int64, error) {

	// Проверка существования файла.
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return 0, ErrNotFound
	}

	// Информация по файлу.
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return 0, fmt.Errorf("Ошибка получения данных по файлу: <%s>: %v", fileName, err)
	}

	return fileInfo.Size(), nil
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

// Проверка существования файла. Возвращается true - файл существует.
//
// Параметры:
//
//	fullNameFile - имя файла.
func fileExists(fullNameFile string) bool {
	_, err := os.Stat(fullNameFile)
	if os.IsNotExist(err) {
		return false // Файл не существует
	}
	return err == nil // Файл существует
}

// Создание токена. Возвращается ключ, токен и ошибка.
func createServerToken() (secretKey, token string, err error) {

	// Создание ключа.
	secretKey, err = generateRandomString(50)
	if err != nil {
		return "", "", fmt.Errorf("функция generateRandomString, вернула ошибку: <%w>", err)
	}

	// Создание токена.
	timeValidToken := time.Duration(24 * time.Hour)
	token, err = createToken("serverManager", secretKey, timeValidToken)
	if err != nil {
		return "", "", fmt.Errorf("функция createToken, вернула ошибку: <%w>", err)
	}

	// Результат.
	return secretKey, token, nil
}

// Генерация строки заданной длины, из случайных символов. Возвращается сгенерированная строка и ошибка.
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

// Получение название файлов в директории. Возвращается массив имён и ошибка.
//
// Параметры:
//
//	dirPath - директория с файлами.
func ReadFilesInDirectory(dirPath string) (files []string, err error) {

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("директория не найдена: %s", dirPath)
	}

	// Чтение файлов в директории
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Проверка, что это - файл.
		if !info.IsDir() {
			files = append(files, info.Name())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// Проверка номера алгоритмом Луна. Возвращается true - проверка успешна.
//
// Параметры:
//
//	number - номер для проверки.
func checkNumbByLuhn(number string) bool {

	number = strings.ReplaceAll(number, " ", "")
	if len(number) == 0 {
		return false
	}

	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
