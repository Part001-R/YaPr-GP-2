package ui

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
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jroimartin/gocui"
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
		"Extraction":             {},
		"fieldPathSource":        {},
		"fieldPathTarget":        {},
		"Backup":                 {},
		"Restore":                {},
		"indicatorPercent":       {},
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
		return false, NilPtrArgumentC
	}

	// Подключение к серверу.
	client, conn, err := connectSrv(c)
	if err != nil {
		return false, fmt.Errorf("функция layerConnectSrv, вернула ошибку: <%w>", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка закрытия подключения: <%v>", err))
		}
	}()

	// Подготовка данных к запросу.
	txMD, nameToken, secretKey, err := layerPrepareDataPingContext(c)
	if err != nil {
		return false, fmt.Errorf("Функция layerPrepareDataPingContext, вернула ошибку: <%w>", err)
	}

	// Запрос.
	if err := layerRequestPingContext(ctx, txMD, client, nameToken, secretKey); err != nil {
		return false, fmt.Errorf("функция layerRequestPingContext, вернула ошибку: <%w>", err)
	}

	// Проверка пройдена.
	return true, nil
}

// Создание резервной копии файла БД, на сервере.
func backUp(c *handlerUI, fileName string, client proto.PasswordManagerClient) error {

	token := "123" //-------------------------------

	// Передача файла на сервер.
	resp, rxHash, rxToken, err := layerBackUpTx(client, fileName, token, c)
	if err != nil {
		return fmt.Errorf("Функция layerTxBackUpDB, вернула ошибку: <%w>", err)
	}

	// Вычисление хэша переданного файла.
	fileHash, err := hashFile(fileName)
	if err != nil {
		return fmt.Errorf("функция hashFile, вернула ошибку: <%w>", err)
	}

	// Анализ данных ответа от сервера.
	if err := layerBackUpCheckResult(resp, fileName, fileHash, rxHash, token, rxToken); err != nil {
		return fmt.Errorf("Функция layerCheckResultBackUpDB, вернула ошибку: <%w>", err)
	}

	return nil
}

// Восстановление из резервной копии файла БД.
func restore(c *handlerUI, fileName string, client proto.PasswordManagerClient) error {

	token := "123" //-------------------------------

	// Запрос файла у сервера.
	content, rxFileHash, rxToken, err := layerRx(client, fileName, token, c)
	if err != nil {
		return fmt.Errorf("Функция layerRx, вернула ошибку: <%w>", err)
	}

	// Предварительное удаление файла.
	if isFileExists(fileName) {
		if err := os.Remove(fileName); err != nil {
			return fmt.Errorf("ошибка:<%w> предварительного удаления файла:<%s>", err, fileName)
		}
	}

	// Сохранение файла.
	if err := saveFile(content, fileName); err != nil {
		return fmt.Errorf("Функция saveFile, вернула ошибку: <%w>", err)
	}

	// Вычисление хэша переданного файла.
	fileHash, err := hashFile(fileName)
	if err != nil {
		return fmt.Errorf("функция hashFile, вернула ошибку: <%w>", err)
	}

	// Проверка результата.
	if err := layerRestoreCheckResult(fileName, fileHash, rxFileHash, token, rxToken); err != nil {

		// Удаление файла, если проверка не пройдена.
		if errRemove := os.Remove(fileName); errRemove != nil {
			return fmt.Errorf("ошибка:<%w> удаления файла:<%s>, после приёма. Базовая ошибка:<%w>", errRemove, fileName, err)
		}
		return fmt.Errorf("функция layerRestoreCheckResult, вернула ошибку:<%w>, для файла:<%s>", err, fileName)
	}

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

// Проверка данных регистрации.
func checkDataRegistration(userName, userPwd1, userPwd2 string) error {

	// Проверка аргументов.
	if userName == "" {
		return MissingDataArgumentUserName
	}
	if userPwd1 == "" {
		return MissingDataArgumentUserPwd1
	}
	if userPwd2 == "" {
		return MissingDataArgumentUserPwd2
	}

	// Проверка пароля.
	if userPwd1 != userPwd2 {
		return NotEqualPassword
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

	encodeRxData, err := c.conf.DataBase.ReadTableLoginPasswordContext(ctx)
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

	encodeRxData, err := c.conf.DataBase.ReadTableBankCardContext(ctx)
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

	encodeRxData, err := c.conf.DataBase.ReadTableTextContext(ctx)
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

// Получение данных afqkf по индексу. Возвращается запись.
//
// Параметры:
//
//	с - конфигурация.
func fileByIndex(c *handlerUI) string {

	return c.data.files[c.index.file]
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

// Увеличение значения индекса для массива файлов.
//
// Параметры:
//
//	с - конфигурация.
func incrIndexFile(c *handlerUI) {

	if c.index.file < len(c.data.files)-1 {
		c.index.file++
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

// Уменьшение значения индекса для массива файлов.
//
// Параметры:
//
//	с - конфигурация.
func decrIndexFile(c *handlerUI) {

	if c.index.file > 0 {
		c.index.file--
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

// Очистка введённой строки от пробелов и \n. Возвращается очищенная трока.
//
// Параметры:
//
//	typed - введённая строка.
func cleaningTypedString(typed string) string {

	var str string

	str = strings.TrimSpace(typed)
	str = strings.TrimSuffix(str, "\n")

	return str
}

//
// --- Enter ---
//

// Обработка нажатия Enter в окне viewRegistration. Возвращается ошибка.
//
// Параметры:
//
//	v - указатель на вид.
//	с - указатель на конфигурацию.
func enterViewRegistration(v *gocui.View, c *handlerUI) error {

	// Обработка полей ввода.
	switch v.Name() {
	case "Login":
		c.typed.login = cleaningTypedString(v.Buffer())
	case "Password-1":
		c.typed.password1 = cleaningTypedString(v.Buffer())
	case "Password-2":
		c.typed.password2 = cleaningTypedString(v.Buffer())
	default:
	}

	return nil
}

// Обработка нажатия Enter в окне viewAutentification. Возвращается ошибка.
//
// Параметры:
//
//	v - указатель на вид.
//	с - указатель на конфигурацию.
func enterViewAutentification(v *gocui.View, c *handlerUI) error {

	switch v.Name() {
	case "Login":
		c.typed.login = cleaningTypedString(v.Buffer())
	case "Password-1":
		c.typed.password1 = cleaningTypedString(v.Buffer())
	default:
	}

	return nil
}

// Обработка нажатия Enter в окне viewSettings. Возвращается ошибка.
//
// Параметры:
//
//	v - указатель на вид.
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func enterViewSettings(v *gocui.View, g *gocui.Gui, c *handlerUI) error {

	c.status.checkConnectPassed = false // Сброс признака процесса проверки связи.
	c.status.checkConnectStatus = false // Сброс статуса результата проверки связи.

	testConnectView, err := g.View("TestConnect")
	if err != nil {
		return fmt.Errorf("Ошибка в функции View: <%w>", err)
	}
	testConnectView.FgColor = gocui.ColorWhite // Сбор статусного цвета надписи.

	switch v.Name() {
	case "IP":
		c.typed.ip = cleaningTypedString(v.Buffer())
	case "Port":
		c.typed.port = cleaningTypedString(v.Buffer())
	default:
	}
	return nil
}

// Обработка нажатия Enter в окне viewRequestSecretKey. Возвращается ошибка.
//
// Параметры:
//
//	v - указатель на вид.
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func enterViewRequestSecretKey(v *gocui.View, g *gocui.Gui, c *handlerUI) error {

	// Получение данных секретного ключа.
	switch v.Name() {
	case "scrtKey":
		str := c.typed.login + c.typed.password1 + strings.TrimSuffix(v.Buffer(), "\n")
		c.secret.secretKey = generateSecretKey(str) // создание ключа шифрования из введённых данных.
	default:
	}

	// Отображение окно, выбранного типа.
	if err := c.showSelectType(g, v); err != nil {
		return fmt.Errorf("функция c.showSelectType, вернула ошибку: <%w>", err)
	}

	return nil
}

// Обработка нажатия Enter в окне viewSelectType. Возвращается ошибка.
//
// Параметры:
//
//	v - указатель на вид.
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func enterViewSelectType(v *gocui.View, g *gocui.Gui, c *handlerUI) error {

	// определение, какое окно открыть.
	switch c.view.currentFocus {
	case "selectLoginPassword":
		if err := c.showLoginPassword(g, v); err != nil {
			return fmt.Errorf("функция c.showLoginPassword, вернула ошибку: <%w>", err)
		}
	case "selectText":
		if err := c.showText(g, v); err != nil {
			return fmt.Errorf("функция c.showText, вернула ошибку: <%w>", err)
		}
	case "SelectBinary":
		if err := c.showBinary(g, v); err != nil {
			return fmt.Errorf("функция c.showBinary, вернула ошибку: <%w>", err)
		}
	case "SelectBankCard":
		if err := c.showBankCard(g, v); err != nil {
			return fmt.Errorf("функция c.showBankCard, вернула ошибку: <%w>", err)
		}
	default:
	}

	return nil
}

// Обработка нажатия Enter в окне viewLoginPasswordData. Возвращается ошибка.
//
// Параметры:
//
//	v - указатель на вид.
//	с - указатель на конфигурацию.
func enterViewLoginPasswordData(v *gocui.View, c *handlerUI) error {

	c.status.addLoginPaaswordPassed = false
	c.status.addLoginPaaswordSUCCESS = false

	switch v.Name() {
	case "fieldAddFor":
		c.typed.dataFor = cleaningTypedString(v.Buffer())
	case "fieldAddLogin":
		c.typed.dataLogin = cleaningTypedString(v.Buffer())
	case "fieldAddPassword":
		c.typed.dataPassword = cleaningTypedString(v.Buffer())
	default:
	}

	return nil
}

// Обработка нажатия Enter в окне viewTextData. Возвращается ошибка.
//
// Параметры:
//
//	v - указатель на вид.
//	с - указатель на конфигурацию.
func enterViewTextData(v *gocui.View, c *handlerUI) error {

	c.status.addTextPassed = false  // Сброс признака.
	c.status.addTextSUCCESS = false // Сброс статуса.

	switch v.Name() {
	case "fieldAddFor":
		c.typed.dataFor = cleaningTypedString(v.Buffer())
	case "fieldAddText":
		c.typed.dataText = cleaningTypedString(v.Buffer())
	default:
	}

	return nil
}

// Обработка нажатия Enter в окне viewBankCardData. Возвращается ошибка.
//
// Параметры:
//
//	v - указатель на вид.
//	с - указатель на конфигурацию.
func enterViewBankCardData(v *gocui.View, c *handlerUI) error {

	c.status.addBankCardPassed = false  // Сброс признака.
	c.status.addBankCardSUCCESS = false // Сброс статуса.

	switch v.Name() {
	case "fieldAddFor":
		c.typed.dataFor = cleaningTypedString(v.Buffer())
	case "fieldAddOwner":
		c.typed.dataOwner = cleaningTypedString(v.Buffer())
	case "fieldAddNumber":
		c.typed.dataNumb = cleaningTypedString(v.Buffer())
	case "fieldAddValid":
		c.typed.dataValidDate = cleaningTypedString(v.Buffer())
	case "fieldAddCode":
		c.typed.dataCode = cleaningTypedString(v.Buffer())
	default:
	}

	return nil
}

// Обработка нажатия Enter в окне viewBinaryData. Возвращается ошибка.
//
// Параметры:
//
//	v - указатель на вид.
//	с - указатель на конфигурацию.
func enterViewBinaryData(v *gocui.View, c *handlerUI) error {

	c.status.addFilePassed = false  // Сброс признака.
	c.status.addFileSUCCESS = false // Сброс статуса.

	switch v.Name() {
	case "fieldPathSource":
		c.typed.dataPathSrc = cleaningTypedString(v.Buffer())
	case "fieldPathTarget":
		c.typed.dataPathTrg = cleaningTypedString(v.Buffer())
	default:
	}

	return nil
}

//
// --- Индикаторы ---
//

// Обновление цвета у индикторов окна viewRegistration. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func indicatorViewRegistration(g *gocui.Gui, c *handlerUI) error {

	// Обработка индикатора проверки введённых данных
	name := "indicator-match"

	indicator, err := g.View(name)
	if err != nil {
		return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
	}
	if indicator == nil {
		return fmt.Errorf("Нет указателя на элемент: <%s>", name)
	}
	if c.typed.password1 != "" && c.typed.password2 != "" && c.typed.password1 == c.typed.password2 {
		indicator.Clear()
		indicator.Write([]byte("Данные приняты!"))
		indicator.FgColor = gocui.ColorGreen
		indicator.BgColor = gocui.ColorDefault
	} else {
		indicator.Clear()
		indicator.Write([]byte("Укажите данные"))
		indicator.FgColor = gocui.ColorRed
		indicator.BgColor = gocui.ColorDefault
	}

	// Обработка индикатора процесса регистрации.
	if c.status.addUserPassed {
		name := "indicator-registration"

		indicator, err = g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if indicator == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}
		if c.status.addUserSUCCESS {
			indicator.Clear()
			indicator.Write([]byte("Пользователь зарегистрирован. Выполните вход."))
			indicator.FgColor = gocui.ColorGreen
			indicator.BgColor = gocui.ColorDefault
		} else {
			indicator.Clear()
			indicator.Write([]byte("Ошибка при регистрации нового пользователя."))
			indicator.FgColor = gocui.ColorRed
			indicator.BgColor = gocui.ColorDefault
		}

		c.status.addUserPassed = false
	}

	// Обработка случая, если при регистрации пользователя, в системе уже присутствует запись.
	if c.status.addUserRegBusy {
		name := "indicator-registration"

		indicator, err = g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if indicator == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}
		indicator.Clear()
		indicator.Write([]byte("Уже есть зарегистрированный пользователь."))
		indicator.FgColor = gocui.ColorRed
		indicator.BgColor = gocui.ColorDefault

		c.status.addUserRegBusy = false
	}

	return nil
}

// Обновление цвета у индикторов окна viewSettings. Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func indicatorViewSettings(g *gocui.Gui, c *handlerUI) error {

	// Есть установлен признак отработки проверки связи.
	if c.status.checkConnectPassed {

		name := "TestConnect"

		// Изменение цвета, в зависимости от результата.
		indicator, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if indicator == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}

		if c.status.checkConnectStatus {
			indicator.FgColor = gocui.ColorGreen
		} else {
			indicator.FgColor = gocui.ColorRed
		}

		c.status.checkConnectPassed = false

	}

	return nil
}

// Обновление цвета у индикторов окна viewLoginPasswordData . Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func indicatorViewLoginPasswordData(g *gocui.Gui, c *handlerUI) error {

	name := "indicatorReadStatus"

	// Обработка индикатора получения данных.
	indicatorRead, err := g.View(name)
	if err != nil {
		return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
	}
	if indicatorRead == nil {
		return fmt.Errorf("Нет указателя на элемент: <%s>", name)
	}

	if c.status.readLoginPaaswordPassed { // обработка при чтении
		if c.status.readLoginPaaswordSUCCESS {
			indicatorRead.Clear()
			indicatorRead.Write([]byte(fmt.Sprintf("Всего записей: %d", len(c.data.loginPassword))))
			indicatorRead.FgColor = gocui.ColorGreen
			indicatorRead.BgColor = gocui.ColorDefault
		} else {
			indicatorRead.Clear()
			indicatorRead.Write([]byte("Ошибка"))
			indicatorRead.FgColor = gocui.ColorRed
			indicatorRead.BgColor = gocui.ColorDefault
		}
	}
	if c.status.delLoginPaaswordPassed { // обработка при удалении
		if c.status.delLoginPaaswordSUCCESS {
			indicatorRead.Clear()
			indicatorRead.Write([]byte("Запись удалена"))
			indicatorRead.FgColor = gocui.ColorGreen
			indicatorRead.BgColor = gocui.ColorDefault
		} else {
			indicatorRead.Clear()
			indicatorRead.Write([]byte("Ошибка удаления"))
			indicatorRead.FgColor = gocui.ColorRed
			indicatorRead.BgColor = gocui.ColorDefault
		}
	}

	// Обработка индикатора добавления записи.
	name = "indicatorAddSuccess"

	indicator, err := g.View(name)
	if err != nil {
		return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
	}
	if indicator == nil {
		return fmt.Errorf("Нет указателя на элемент: <%s>", name)
	}

	if c.status.addLoginPaaswordPassed {
		if c.status.addLoginPaaswordSUCCESS {
			indicator.Clear()
			indicator.Write([]byte("Данные приняты!"))
			indicator.FgColor = gocui.ColorGreen
			indicator.BgColor = gocui.ColorDefault
		} else {
			indicator.Clear()
			indicator.Write([]byte("Ошибка добавления."))
			indicator.FgColor = gocui.ColorRed
			indicator.BgColor = gocui.ColorDefault
		}
	} else {
		indicator.Clear()
		indicator.Write([]byte(""))
	}

	return nil
}

// Обновление цвета у индикторов окна viewTextData . Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func indicatorViewTextData(g *gocui.Gui, c *handlerUI) error {

	// Обработка индикатора получения данных.
	name := "indicatorReadStatus"

	indicatorRead, err := g.View(name)
	if err != nil {
		return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
	}
	if indicatorRead == nil {
		return fmt.Errorf("Нет указателя на элемент: <%s>", name)
	}

	if c.status.readTextPassed { // обработка при чтении
		if c.status.readTextSUCCESS {
			indicatorRead.Clear()
			indicatorRead.Write([]byte(fmt.Sprintf("Всего записей: %d", len(c.data.textData))))
			indicatorRead.FgColor = gocui.ColorGreen
			indicatorRead.BgColor = gocui.ColorDefault
		} else {
			indicatorRead.Clear()
			indicatorRead.Write([]byte("Ошибка"))
			indicatorRead.FgColor = gocui.ColorRed
			indicatorRead.BgColor = gocui.ColorDefault
		}
		c.status.readTextPassed = false
	}
	if c.status.delTextPassed { // обработка при удалении
		if c.status.delTextSUCCESS {
			indicatorRead.Clear()
			indicatorRead.Write([]byte("Запись удалена"))
			indicatorRead.FgColor = gocui.ColorGreen
			indicatorRead.BgColor = gocui.ColorDefault
		} else {
			indicatorRead.Clear()
			indicatorRead.Write([]byte("Ошибка удаления"))
			indicatorRead.FgColor = gocui.ColorRed
			indicatorRead.BgColor = gocui.ColorDefault
		}
		c.status.delTextPassed = false
	}

	// Обработка индикатора добавления записи.
	name = "indicatorAddSuccess"

	indicator, err := g.View(name)
	if err != nil {
		return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
	}
	if indicator == nil {
		return fmt.Errorf("Нет указателя на элемент: <%s>", name)
	}

	if c.status.addTextPassed {
		if c.status.addTextSUCCESS {
			indicator.Clear()
			indicator.Write([]byte("Данные приняты!"))
			indicator.FgColor = gocui.ColorGreen
			indicator.BgColor = gocui.ColorDefault
		} else {
			indicator.Clear()
			indicator.Write([]byte("Ошибка добавления."))
			indicator.FgColor = gocui.ColorRed
			indicator.BgColor = gocui.ColorDefault
		}
		c.status.addTextPassed = false
	} else {
		indicator.Clear()
		indicator.Write([]byte(""))
	}

	return nil
}

// Обновление цвета у индикторов окна viewBankCardData . Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func indicatorViewBankCardData(g *gocui.Gui, c *handlerUI) error {

	// Обработка индикатора получения данных.
	name := "indicatorReadStatus"

	indicatorRead, err := g.View(name)
	if err != nil {
		return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
	}
	if indicatorRead == nil {
		return fmt.Errorf("Нет указателя на элемент: <%s>", name)
	}
	if c.status.readBankCardPassed { // обработка при чтении
		if c.status.readBankCardSUCCESS {
			indicatorRead.Clear()
			indicatorRead.Write([]byte(fmt.Sprintf("Всего записей: %d", len(c.data.bankCard))))
			indicatorRead.FgColor = gocui.ColorGreen
			indicatorRead.BgColor = gocui.ColorDefault
		} else {
			indicatorRead.Clear()
			indicatorRead.Write([]byte("Ошибка"))
			indicatorRead.FgColor = gocui.ColorRed
			indicatorRead.BgColor = gocui.ColorDefault
		}
		c.status.readBankCardPassed = false
	}

	if c.status.delBankCardPassed { // обработка при удалении
		if c.status.delBankCardSUCCESS {
			indicatorRead.Clear()
			indicatorRead.Write([]byte("Запись удалена"))
			indicatorRead.FgColor = gocui.ColorGreen
			indicatorRead.BgColor = gocui.ColorDefault
		} else {
			indicatorRead.Clear()
			indicatorRead.Write([]byte("Ошибка удаления"))
			indicatorRead.FgColor = gocui.ColorRed
			indicatorRead.BgColor = gocui.ColorDefault
		}
		c.status.delBankCardPassed = false
	}

	// Обработка индикатора добавления записи.
	name = "indicatorAddSuccess"

	indicator, err := g.View(name)
	if err != nil {
		return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
	}
	if indicator == nil {
		return fmt.Errorf("Нет указателя на элемент: <%s>", name)
	}

	if c.status.addBankCardPassed {
		if c.status.addBankCardSUCCESS {
			indicator.Clear()
			indicator.Write([]byte("Данные приняты!"))
			indicator.FgColor = gocui.ColorGreen
			indicator.BgColor = gocui.ColorDefault
		} else {
			indicator.Clear()
			indicator.Write([]byte("Ошибка добавления."))
			indicator.FgColor = gocui.ColorRed
			indicator.BgColor = gocui.ColorDefault
		}
	} else {
		indicator.Clear()
		indicator.Write([]byte(""))
	}

	return nil
}

// Обновление цвета у индикторов окна viewBinaryData . Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func indicatorViewBinaryData(g *gocui.Gui, c *handlerUI) error {

	// Есть установлен признак отработки добавления файла в контейнер.
	if c.status.addFilePassed {
		name := "Save"

		element, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if element == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}

		if c.status.addFileSUCCESS {
			element.FgColor = gocui.ColorGreen
		} else {
			element.FgColor = gocui.ColorRed
		}

		c.status.addFilePassed = false
	}

	// Есть установлен признак извлечения файла из контейнера.
	if c.status.extractFilePassed {
		name := "Extraction"

		element, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if element == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}

		if c.status.extractFileSUCCESS {
			element.FgColor = gocui.ColorGreen
		} else {
			element.FgColor = gocui.ColorRed
		}

		c.status.extractFilePassed = false
	}

	// Есть установлен признак удаления файла из контейнера.
	if c.status.delFilePassed {
		name := "DeleteElement"

		element, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if element == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}

		if c.status.delFileSUCCESS {
			element.FgColor = gocui.ColorGreen
		} else {
			element.FgColor = gocui.ColorRed
		}

		c.status.delFilePassed = false
		c.status.delFileSUCCESS = false
	}

	// Если чтение файлов контейнера выполнено.
	if c.status.readFilePassed {
		name := "indicatorReadStatus"

		indicator, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if indicator == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}

		if c.status.readFileSUCCESS {

			indicator.Clear()
			indicator.Write([]byte(fmt.Sprintf("Всего файлов: %d", len(c.data.files))))
			indicator.FgColor = gocui.ColorGreen
			indicator.BgColor = gocui.ColorDefault
		} else {
			indicator.Clear()
			indicator.Write([]byte("Ошибка чтения."))
			indicator.FgColor = gocui.ColorRed
			indicator.BgColor = gocui.ColorDefault
		}

		c.status.readFilePassed = false  // Для разовой отработки при открытии экрана.
		c.status.readFileSUCCESS = false // Для разовой отработки при открытии экрана.
	}

	return nil
}

// Обновление индикторов окна viewSelectType . Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func indicatorViewSelectType(g *gocui.Gui, c *handlerUI) error {

	//
	// Обновление вида элемента backUp (---> сервер).
	//

	name := "Backup"
	btnBackup, err := g.View(name)
	if err != nil {
		return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
	}
	if btnBackup == nil {
		return fmt.Errorf("Нет указателя на элемент: <%s>", name)
	}

	statusBackUp := c.getStatusBackUp()

	switch statusBackUp {
	case stageNotActive: // Нет активности процесса.
		btnBackup.FgColor = gocui.ColorWhite

	case stageActive: // Есть активность процесса.
		btnBackup.FgColor = gocui.ColorYellow

	case stageOk: // Процесс завершён успешно.
		btnBackup.FgColor = gocui.ColorGreen

	case stageFault: // Ошибка процесса.
		btnBackup.FgColor = gocui.ColorRed

	default:
	}

	//
	// Обновление вида элемента restore (<--- сервер).
	//

	name = "Restore"
	btnRestore, err := g.View(name)
	if err != nil {
		return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
	}
	if btnRestore == nil {
		return fmt.Errorf("Нет указателя на элемент: <%s>", name)
	}

	statusRestore := c.getStatusRestore()

	switch statusRestore {
	case stageNotActive: // Нет активности процесса.
		btnRestore.FgColor = gocui.ColorWhite

	case stageActive: // Есть активность процесса.
		btnRestore.FgColor = gocui.ColorYellow

	case stageOk: // Процесс завершён успешно.
		btnRestore.FgColor = gocui.ColorGreen

	case stageFault: // Ошибка процесса.
		btnRestore.FgColor = gocui.ColorRed

	default:
	}

	//
	// --- Индикатор процентов ---
	//

	if statusBackUp == stageActive || statusBackUp == stageOk ||
		statusRestore == stageActive || statusRestore == stageOk {

		name = "indicatorPercent"
		indicator, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if indicator == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}

		indicator.Clear()
		indicator.Write([]byte(fmt.Sprintf("Выполнено: %.2f%%", c.getPercentTxRx())))
	}

	return nil
}

// Определение общего размера файлов в КБайт. Возвращается общий размер и ошибка.
//
// Параметры:
//
//	files - названия файлов.
func totalFileSize(files []string) (int64, error) {

	var totalSize int64

	for _, filename := range files {
		fileInfo, err := os.Stat(filename)
		if err != nil {
			return 0, fmt.Errorf("Ошибка получения данных по файлу: <%s>", filename)
		}

		// Проверка переполнения.
		fileSize := fileInfo.Size()
		if totalSize > math.MaxInt64-fileSize {
			return 0, fmt.Errorf("Переполнение при суммировании размеров файлов")
		}

		totalSize += fileInfo.Size() / 1024
	}

	return totalSize, nil
}

// Вычисление процента выполнения.
//
// Параметры:
//
//	c - конфигурация.
//	b - количество переданных байт.
func updateDataBackUpRestoreProcess(c *handlerUI, b int) {

	c.mutex.processTxRx.Lock()
	defer c.mutex.processTxRx.Unlock()

	// Получение КБайт из Байт.
	volumeKB := b / 1024

	// Обновление данных накопителя.
	c.txrx.passedKB += int64(volumeKB)

	// Вычисление процентов.
	if c.txrx.totalSizeKB > 0 {
		c.txrx.percentTxRx = float32(float64(c.txrx.passedKB) / float64(c.txrx.totalSizeKB) * 100.0)
	} else {
		c.txrx.percentTxRx = 0
	}
}

// Запрос информации о файлах.
func requestFilesInfo(client proto.PasswordManagerClient, c *handlerUI) error {

	token := "123" //-----------------------------------------------------------

	// Запрос информации по файлам у сервера.
	filesInfo, rxToken, err := layerFilesInfoRequest(client, token)
	if err != nil {
		return fmt.Errorf("Функция layerFilesInfoRequest, вернула ошибку: <%w>", err)
	}

	// Проверка результата запроса.
	if len(filesInfo) == 0 {
		return EmptyData
	}
	if token != rxToken {
		return NotEqualTokens
	}

	// Заполнение данных по результатам запроса.
	if err := layerFilesInfoFillData(filesInfo, c); err != nil {
		return fmt.Errorf("Функция layerFilesInfoFillData, вернула ошибку: <%w>", err)
	}

	return nil
}

// Процесс backUp.
func doBackupProcess(filesList []string, c *handlerUI) {

	// Подключение к серверу.
	client, conn, err := connectSrv(c)
	if err != nil {
		c.updateStatusRestore(stageFault)
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Функция connectSrv, вернула ошибку: <%v>", err))
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция conn.Close, вернула ошибку: <%v>", err))
		}
	}()

	// Передача файлов.
	for _, f := range filesList {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Запуск процесса резервного копирования <%s>", f))

		if err := backUp(c, f, client); err != nil {
			c.updateStatusBackUp(stageFault)
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка backUp:<%v>, файла:<%s> ", err, f))
			return
		}
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Резервное копирование <%s>, выполнено", f))

	}

	// Установка признака, что процесс выполнен.
	c.updateStatusBackUp(stageOk)
}

func doRestoreProcess(filesList []string, c *handlerUI) {

	// Подключение к серверу.
	client, conn, err := connectSrv(c)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Функция connectSrv, вернула ошибку: <%v>", err))
		c.updateStatusRestore(stageFault)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция conn.Close, вернула ошибку: <%v>", err))
		}
	}()

	// Запрос у сервера информации по файлам (имя, размер), которые будут приняты.
	if err := requestFilesInfo(client, c); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция requestFilesInfo, вернула ошибку: <%v>", err))
		c.updateStatusRestore(stageFault)
		return
	}

	// Получение файлов.
	for _, f := range filesList {
		if err := restore(c, f, client); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка restore: <%v>, при приёме: <%s> ", err, f))
			c.updateStatusRestore(stageFault)
			return
		}
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Восстановление файла <%s>, выполнено", f))
	}

	// Установка признака, что восстановление выполнено.
	c.updateStatusRestore(stageOk)
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

// Проверка существования файла.
func isFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false // Файл не существует
	}
	return err == nil // Файл существует
}
