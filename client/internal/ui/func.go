package ui

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
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
		viewSelectType:           {},
		viewLoginPasswordData:    {},
		viewTextData:             {},
		viewBinaryData:           {},
		viewBankCardData:         {},
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
		"selectLoginPassword":    {},
		"selectText":             {},
		"SelectBinary":           {},
		"SelectBankCard":         {},
		"Back":                   {},
		"fieldShowFor":           {},
		"fieldShowLogin":         {},
		"fieldShowPassword":      {},
		"fieldAddFor":            {},
		"fieldAddLogin":          {},
		"fieldAddPassword":       {},
		"indicatorAddSuccess":    {},
		"Save":                   {},
		"indicatorReadStatus":    {},
		"NextElement":            {},
		"PrevElement":            {},
		"DeleteElement":          {},
		"fieldShowText":          {},
		"fieldAddText":           {},
		"fieldShowOwner":         {},
		"fieldShowNumber":        {},
		"fieldShowValid":         {},
		"fieldShowCode":          {},
		"fieldAddOwner":          {},
		"fieldAddNumber":         {},
		"fieldAddValid":          {},
		"fieldAddCode":           {},
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

// Функция генерирует секретный ключ из входной строки. Возвращается массив байт.
//
// Параметры:
//
//	input - данные, на основе которых формируется ключ.
func generateSecretKey(input string) [32]byte {

	return sha256.Sum256([]byte(input))
}

// Шифрование данных. Возвращается результат шифрования и ошибка.
//
// Параметры:
//
//	data - данные для шифрования.
//	key - ключ шифрования.
func encrypt(data string, key [32]byte) (string, error) {
	// Создаём AES-256 шифр
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	// Преобразуем текст в байты
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

	// Кодируем в Base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Расшифровка данных. Возвращается результат расшифровки и ошибка.
//
// Параметры:
//
//	data - зашифрованные данные.
//	key - ключ шифрования.
func decrypt(data string, key [32]byte) (string, error) {
	// Декодируем из Base64
	ciphertextBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	// Создаём AES-256 шифр
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	if len(ciphertextBytes)%blockSize != 0 {
		return "", errors.New("некорректная длина зашифрованных данных")
	}

	// Вектор инициализации (такой же, как при шифровании)
	iv := make([]byte, blockSize)

	// Расшифровка
	plaintextPadded := make([]byte, len(ciphertextBytes))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintextPadded, ciphertextBytes)

	// Удаляем PKCS#7 заполнение
	padding := int(plaintextPadded[len(plaintextPadded)-1])
	if padding > blockSize || padding == 0 {
		return "", errors.New("некорректное заполнение")
	}
	plaintext := plaintextPadded[:len(plaintextPadded)-padding]

	return string(plaintext), nil
}

// Декодирование данных логин/пароль. Возвращаются декодированные данные и ошибка.
//
// Параметры:
//
//	encryptData - закодированные данные.
//	key - секретный ключ.
func decryptDataLoginPassword(encryptData []loginPassword, key [32]byte) (decryptData []loginPassword, err error) {

	for _, v := range encryptData {
		var el loginPassword

		// Обработка поля - name
		str, err := decrypt(v.name, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании name, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.name = str

		// Обработка поля - login
		str, err = decrypt(v.login, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании login, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.login = str

		// Обработка поля - password
		str, err = decrypt(v.password, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании password, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.password = str

		// Обработка поля - createdAt
		str, err = decrypt(v.createdAt, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании createdAt, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.createdAt = str

		decryptData = append(decryptData, el)
	}

	// результат
	return decryptData, nil
}

// Декодирование данных текста. Возвращаются декодированные данные и ошибка.
//
// Параметры:
//
//	encryptData - закодированные данные.
//	key - секретный ключ.
func decryptDataText(encryptData []textData, key [32]byte) (decryptData []textData, err error) {

	for _, v := range encryptData {
		var el textData

		// Обработка поля - name
		str, err := decrypt(v.name, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании name, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.name = str

		// Обработка поля - text
		str, err = decrypt(v.text, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании text, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.text = str

		// Обработка поля - createdAt
		str, err = decrypt(v.createdAt, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании createdAt, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.createdAt = str

		decryptData = append(decryptData, el)
	}

	// результат
	return decryptData, nil
}

// Декодирование данных банковских карт. Возвращаются декодированные данные и ошибка.
//
// Параметры:
//
//	encryptData - закодированные данные.
//	key - секретный ключ.
func decryptDataBankCard(encryptData []bankCard, key [32]byte) (decryptData []bankCard, err error) {

	for _, v := range encryptData {
		var el bankCard

		// Обработка поля - имя
		str, err := decrypt(v.name, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании name, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.name = str

		// Обработка поля - владелец
		str, err = decrypt(v.owner, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании owner, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.owner = str

		// Обработка поля - номер карты
		str, err = decrypt(v.numb, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании numb, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.numb = str

		// Обработка поля - валидность
		str, err = decrypt(v.valid, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании valid, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.valid = str

		// Обработка поля - код
		str, err = decrypt(v.code, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании code, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.code = str

		// Обработка поля - createdAt
		str, err = decrypt(v.createdAt, key)
		if err != nil {
			return nil, fmt.Errorf("при декодировании createdAt, функция decrypt, вернула ошибку: <%w>", err)
		}
		el.createdAt = str

		decryptData = append(decryptData, el)
	}

	// результат
	return decryptData, nil
}

// Функция реализует получение пар логин/пароль из БД и выполняет декодирование. Возвращается количество записей и ошибка.
//
// Параметры:
//
//	с - конфигурация.
func showLoginPasswordWorkDB(c *handlerUI) (int, error) {

	// Чтение из БД всех записей таблицы логин/пароль (data1).
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	encodeRxData, err := c.conf.DB.ReadTableLoginPasswordContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("функция ReadTableLoginPasswordContext, вернула ошибку: <%v>", err)
	}

	// Перенос принятых закодированных данных логин/пароль, в in-memory.
	c.data.encryptLoginPassword = []loginPassword{} // сброс содержимого слайса

	for _, v := range encodeRxData {
		var el loginPassword
		el.name = v.Name
		el.login = v.Login
		el.password = v.Password
		el.createdAt = v.CreatedAt

		c.data.encryptLoginPassword = append(c.data.encryptLoginPassword, el)
	}
	// Декодирование принятых данных.
	c.data.loginPassword, err = decryptDataLoginPassword(c.data.encryptLoginPassword, c.secret.secretKey)
	if err != nil {
		return 0, fmt.Errorf("ошибка декодирования данных логин/пароль: <%v>", err)
	}

	// Результат.
	return len(c.data.loginPassword), nil
}

// Функция реализует получение банковских карт из БД и выполняет декодирование. Возвращается количество записей и ошибка.
//
// Параметры:
//
//	с - конфигурация.
func showBankCardWorkDB(c *handlerUI) (int, error) {

	// Чтение из БД всех записей таблицы логин/пароль (data1).
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	encodeRxData, err := c.conf.DB.ReadTableBankCardContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("функция ReadTableLoginPasswordContext, вернула ошибку: <%v>", err)
	}

	// Перенос принятых закодированных данных логин/пароль, в in-memory.
	c.data.encryptBankCard = []bankCard{} // сброс содержимого слайса

	for _, v := range encodeRxData {
		var el bankCard
		el.name = v.Name
		el.owner = v.Owner
		el.numb = v.Numb
		el.valid = v.Valid
		el.code = v.Code
		el.createdAt = v.CreatedAt

		c.data.encryptBankCard = append(c.data.encryptBankCard, el)
	}
	// Декодирование принятых данных.
	c.data.bankCard, err = decryptDataBankCard(c.data.encryptBankCard, c.secret.secretKey)
	if err != nil {
		return 0, fmt.Errorf("ошибка декодирования данных логин/пароль: <%v>", err)
	}

	// Результат.
	return len(c.data.bankCard), nil
}

// Функция реализует получение данных текста из БД и выполняет декодирование. Возвращается количество записей и ошибка.
//
// Параметры:
//
//	с - конфигурация.
func showTextWorkDB(c *handlerUI) (int, error) {

	// Чтение из БД всех записей таблицы логин/пароль (data1).
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	encodeRxData, err := c.conf.DB.ReadTableTextContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("функция ReadTableTextContext, вернула ошибку: <%v>", err)
	}

	// Перенос принятых закодированных данных логин/пароль, в in-memory.
	c.data.encryptTextData = []textData{} // сброс содержимого слайса

	for _, v := range encodeRxData {
		var el textData
		el.name = v.Name
		el.text = v.Text
		el.createdAt = v.CreatedAt

		c.data.encryptTextData = append(c.data.encryptTextData, el)
	}
	// Декодирование принятых данных.
	c.data.textData, err = decryptDataText(c.data.encryptTextData, c.secret.secretKey)
	if err != nil {
		return 0, fmt.Errorf("ошибка декодирования данных логин/пароль: <%v>", err)
	}

	// Результат.
	return len(c.data.loginPassword), nil
}

// Получение данных логин/пароль по индексу. Возвращается запись.
//
// Параметры:
//
//	с - конфигурация.
func loginPasswordByIndex(c *handlerUI) (el loginPassword) {

	el.name = c.data.loginPassword[c.index.loginPassword].name
	el.login = c.data.loginPassword[c.index.loginPassword].login
	el.password = c.data.loginPassword[c.index.loginPassword].password
	el.createdAt = c.data.loginPassword[c.index.loginPassword].createdAt

	return el
}

// Получение данных текста по индексу. Возвращается запись.
//
// Параметры:
//
//	с - конфигурация.
func textByIndex(c *handlerUI) (el textData) {

	el.name = c.data.textData[c.index.text].name
	el.text = c.data.textData[c.index.text].text
	el.createdAt = c.data.textData[c.index.text].createdAt

	return el
}

// Получение данных банковской карты по индексу. Возвращается запись.
//
// Параметры:
//
//	с - конфигурация.
func bankCardByIndex(c *handlerUI) (el bankCard) {

	el.name = c.data.bankCard[c.index.bankCard].name
	el.owner = c.data.bankCard[c.index.bankCard].owner
	el.numb = c.data.bankCard[c.index.bankCard].numb
	el.valid = c.data.bankCard[c.index.bankCard].valid
	el.code = c.data.bankCard[c.index.bankCard].code
	el.createdAt = c.data.bankCard[c.index.bankCard].createdAt

	return el
}

// Увеличение значения индекса для логин/пароль массива.
//
// Параметры:
//
//	с - конфигурация.
func incrIndexloginPassword(c *handlerUI) {

	if c.index.loginPassword < len(c.data.loginPassword)-1 {
		c.index.loginPassword++
	}
}

// Увеличение значения индекса для текст массива.
//
// Параметры:
//
//	с - конфигурация.
func incrIndexText(c *handlerUI) {

	if c.index.text < len(c.data.textData)-1 {
		c.index.text++
	}
}

// Увеличение значения индекса для массива банковских карт.
//
// Параметры:
//
//	с - конфигурация.
func incrIndexBankCard(c *handlerUI) {

	if c.index.bankCard < len(c.data.bankCard)-1 {
		c.index.bankCard++
	}
}

// Уменьшение значения индекса для логин/пароль массива.
//
// Параметры:
//
//	с - конфигурация.
func decrIndexloginPassword(c *handlerUI) {

	if c.index.loginPassword > 0 {
		c.index.loginPassword--
	}
}

// Уменьшение значения индекса для текст массива.
//
// Параметры:
//
//	с - конфигурация.
func decrIndexText(c *handlerUI) {

	if c.index.text > 0 {
		c.index.text--
	}
}

// Уменьшение значения индекса для массива банковских карт.
//
// Параметры:
//
//	с - конфигурация.
func decrIndexBankCard(c *handlerUI) {

	if c.index.bankCard > 0 {
		c.index.bankCard--
	}
}

// Проверка номера банковской карты. Возвращает true - номер корректный.
//
// Параметры:
//
//	cardNumber - номер карты.
func checkCardNumber(cardNumber string) bool {

	sum := 0
	alternate := false

	// С конца строки
	for i := len(cardNumber) - 1; i >= 0; i-- {

		// Отсев лишних символов.
		if cardNumber[i] < '0' || cardNumber[i] > '9' {
			continue
		}

		digit, err := strconv.Atoi(string(cardNumber[i]))
		if err != nil {
			return false
		}

		// *2 для каждого второго символа
		if alternate {
			digit *= 2
			// Если число больше 9, то минус 9
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}

	// Проверка.
	return sum%10 == 0
}
