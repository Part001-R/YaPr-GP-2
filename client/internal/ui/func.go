package ui

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jroimartin/gocui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Part001-R/YaPr-GP-2/proto"
)

const (
	inputWidth  = 80
	inputHeight = 2
)

// Флаг: создан ли интерфейс
var layoutInitialized bool

// Удаление видов.
func deleteViews(g *gocui.Gui) error {

	for name := range map[string]struct{}{
		viewRegistration:         {},
		viewAutentification:      {},
		viewSettings:             {},
		viewRequestSecretKey:     {},
		viewUserData:             {},
		"Login":                  {},
		"Password-1":             {},
		"Password-2":             {},
		"indicator-match":        {},
		"indicator-registration": {},
		"TAB":                    {},
		"Enter":                  {},
		"MainMenu":               {},
		"Exit":                   {},
		"IP":                     {},
		"Port":                   {},
		"TestConnect":            {},
		"DoAuthentication":       {},
		"DoRegistration":         {},
		"scrtKey":                {},
	} {
		if err := g.DeleteView(name); err != nil && err != gocui.ErrUnknownView {
			return err
		}
	}
	return nil
}

// Реализация проверки связи с сервером.
func pingContext(ctx context.Context, c *handlerUI) (bool, error) {
	// Проверка аргументов
	if c == nil {
		return false, errors.New("в аргументе <c> нет указателя")
	}

	// Логика
	c.flag.checkConnectPassed = true
	srvAddr := c.typed.ip + ":" + c.typed.port

	// Настройка TLS.
	creds, err := credentials.NewClientTLSFromFile("tls/server.crt", "")
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("ошибка получения TLS-креденциалов: <%v>", err))
		return false, fmt.Errorf("ошибка получения TLS-креденциалов: <%w>", err)
	}

	// Подключение к серверу.
	conn, err := grpc.NewClient(srvAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("ошибка подключения к серверу: <%v>", err))
		return false, fmt.Errorf("ошибка подключения к серверу: <%w>", err)
	}
	defer conn.Close()

	client := pb.NewPasswordManagerClient(conn)

	// Подготовка данных к запросу.
	secretKey, err := generateRandomString(50)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция generateRandomString, вернула ошибку: <%v>", err))
		return false, fmt.Errorf("функция generateRandomString, вернула ошибку: <%w>", err)
	}

	timeValidToken := time.Duration(5 * time.Second)
	txToken, err := createToken("clientManager", secretKey, timeValidToken)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция createToken, вернула ошибку: <%v>", err))
		return false, fmt.Errorf("функция createToken, вернула ошибку: <%w>", err)
	}

	nameToken := "token"
	txMD := metadata.Pairs(nameToken, txToken)

	// Создание нового контекста, основанного на переданном контексте, с метаданными.
	ctx = metadata.NewOutgoingContext(ctx, txMD)

	// Запрос.
	emptyRequest := &emptypb.Empty{}
	var header metadata.MD

	_, err = client.Ping(ctx, emptyRequest, grpc.Header(&header))
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("ошибка выполнения Ping: <%v>", err))
		return false, fmt.Errorf("ошибка выполнения Ping: <%w>", err)
	}

	// Получение токена из метаданных ответа.
	token := header[nameToken]
	if len(token) == 0 || token[0] == "" {
		c.conf.PtrLoggerFile.Write("в ответе на запрос ping, отсутствуют данные токена")
		return false, errors.New("в ответе на запрос ping, отсутствуют данные токена")
	}
	rxToken := token[0]

	// Проверка метаданных ответа.
	if err := checkToken(rxToken, secretKey); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("функция checkToken, вернула ошибку: <%v>", err))
		return false, err
	}

	// Проверка пройдена успешно.
	return true, nil
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
		return "", errors.New("в аргументе <subjectName>, нет данных")
	}
	if len(secretKey) < 32 {
		return "", errors.New("длина <secretKey> должна быть не менее 32 символов")
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
		return errors.New("в аргументе tokenStr, нет данных")
	}
	if len(secretKey) < 32 {
		return errors.New("длинная secretKey, меньше 32")
	}

	// Логика.
	//
	// Проверка токена.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неизвестный метод подписи")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return fmt.Errorf("ошибка парсинга токена: <%w>", err)
	}

	// Проверка на истечение срока действия токена.
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && claims.ExpiresAt != nil {
		if time.Now().After(claims.ExpiresAt.Time) {
			return errors.New("время валидности токена истекло")
		}
	}

	// Проверка пройдена.
	return nil
}

// Проверка данных регистрации.
func checkDataRegistration(userName, userPwd1, userPwd2 string) error {

	// Проверка аргументов.
	if userName == "" {
		return errors.New("в аргументе <userName>, нет данных")
	}
	if userPwd1 == "" {
		return errors.New("в аргументе <userPwd1>, нет данных")
	}
	if userPwd2 == "" {
		return errors.New("в аргументе <userPwd2>, нет данных")
	}

	// Проверка пароля.
	if userPwd1 != userPwd2 {
		return errors.New("данные пароля не эквивалентны")
	}

	return nil
}
