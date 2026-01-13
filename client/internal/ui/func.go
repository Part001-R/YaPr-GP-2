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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Part001-R/YaPr-GP-2/client/internal/adapters/server"
	"github.com/Part001-R/YaPr-GP-2/client/internal/utils/flags"
	"github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jroimartin/gocui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
		"indicatorNameClient":    {},
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
	txMD, nameToken, secretKey, err := layerDataPingContextPrepare(c)
	if err != nil {
		return false, fmt.Errorf("Функция layerPrepareDataPingContext, вернула ошибку: <%w>", err)
	}

	// Запрос.
	if err := layerPingContextRequest(ctx, txMD, client, nameToken, secretKey); err != nil {
		return false, fmt.Errorf("функция layerRequestPingContext, вернула ошибку: <%w>", err)
	}

	// Проверка пройдена.
	return true, nil
}

// Создание резервной копии файла БД, на сервере.
func backUp(c *handlerUI, txFileName string, client proto.PasswordManagerClient) error {

	// Создание токена.
	secretKey, txToken, err := layerBackUpCreateToken()
	if err != nil {
		return fmt.Errorf("функция layerBackUpCreateToken, вернула ошибку:<%w>", err)
	}

	// Передача файла на сервер.
	resp, rxFileHash, rxToken, err := layerBackUpTxFile(client, txFileName, txToken, c)
	if err != nil {
		return fmt.Errorf("Функция layerTxBackUpDB, вернула ошибку: <%w>", err)
	}

	// Вычисление хэша переданного файла.
	txFileHash, err := hashFile(txFileName)
	if err != nil {
		return fmt.Errorf("функция hashFile, вернула ошибку: <%w>", err)
	}

	// Проверка результата.
	if err := layerBackUpCheckResult(resp, txFileName, txFileHash, rxFileHash, rxToken, secretKey); err != nil {
		return fmt.Errorf("Функция layerCheckResultBackUpDB, вернула ошибку: <%w>", err)
	}

	return nil
}

// Восстановление из резервной копии файла БД.
func restore(c *handlerUI, fileName string, client proto.PasswordManagerClient) (err error) {

	needRestoreState := false // Признак необходимости воостановления состояния, при ошибке.
	var tempFileName string   // Имя временного файла.

	// Обработка перед выходом.
	defer func(needRestoreState bool, fileName, tempFileName string, errProcess error) {
		if errRestore := deferProcessRestoreByError(needRestoreState, fileName, tempFileName, errProcess); errRestore != nil {
			err = fmt.Errorf("функция deferProcessRestoreByError, вернула ошибку:<%w>, при ошибку процесса:<%w>", errRestore, errProcess)
		}
	}(needRestoreState, fileName, tempFileName, err)

	// Создание токена.
	secretKey, txToken, err := layerRestoreCreateToken()
	if err != nil {
		return fmt.Errorf("Функция layerRestoreCreateToken, вернула ошибку: <%w>", err)
	}

	// Запрос файла у сервера.
	content, srcFileHash, rxToken, err := layerRestoreRxFile(client, fileName, txToken, c)
	if err != nil {
		return fmt.Errorf("Функция layerRx, вернула ошибку: <%w>", err)
	}

	// Изменение имени существующего файла, чтобы в случае ошибки, не потерять данные.
	if isFileExists(fileName) {
		tempFileName, err = changeFileName(fileName, "-temp")
		if err != nil {
			return fmt.Errorf("функция changeFileName, вернула ошибку:<%w>", err)
		}
		needRestoreState = true
	}

	// Сохранение принятых данных в файл.
	if err := layerRestoreSaveFile(content, fileName); err != nil {
		return fmt.Errorf("Функция saveFile, вернула ошибку: <%w>", err)
	}

	// Вычисление хэша принятого файла.
	rxFileHash, err := hashFile(fileName)
	if err != nil {
		return fmt.Errorf("функция hashFile, вернула ошибку: <%w>", err)
	}

	// Проверка результата.
	if err := layerRestoreCheckResult(fileName, rxFileHash, srcFileHash, rxToken, secretKey); err != nil {

		// Удаление файла, если проверка не пройдена.
		if errRemove := os.Remove(fileName); errRemove != nil {
			return fmt.Errorf("ошибка:<%w> удаления файла:<%s>, после приёма. Базовая ошибка:<%w>", errRemove, fileName, err)
		}
		return fmt.Errorf("функция layerRestoreCheckResult, вернула ошибку:<%w>, для файла:<%s>", err, fileName)
	}

	// Файл успешно принят.
	// Удаление резервного файла, если он существует.
	if isFileExists(tempFileName) {
		if err := os.Remove(tempFileName); err != nil {
			return fmt.Errorf("ошибка:<%w> удаления резервного файла:<%s>, при успешном приёме", err, fileName)
		}
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

// Получение имени записи логин/пароль по индексу. Возвращается запись.
//
// Параметры:
//
//	с - конфигурация.
func nameLoginPasswordByIndex(c *handlerUI) string {

	return c.data.namesLoginPassword[c.index.loginPassword]
}

// Получение имени записи текста по индексу. Возвращается запись.
//
// Параметры:
//
//	с - конфигурация.
func nameTextByIndex(c *handlerUI) string {

	return c.data.namesText[c.index.text]
}

// Получение имени записи банковской карты по индексу. Возвращается запись.
//
// Параметры:
//
//	с - конфигурация.
func nameBankCardByIndex(c *handlerUI) string {

	return c.data.namesBankCard[c.index.bankCard]
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

// Получение данных файла по индексу. Возвращается запись.
//
// Параметры:
//
//	с - конфигурация.
func fileByIndex(c *handlerUI) string {

	return c.data.files[c.index.file]
}

// Получение данных файла по индексу. Возвращается запись.
//
// Параметры:
//
//	с - конфигурация.
func fileNameByIndex(c *handlerUI) string {

	return c.data.namesFile[c.index.file]
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

// Увеличение значения индекса для имён логин/пароль массива.
//
// Параметры:
//
//	с - конфигурация.
func incrIndexNamesloginPassword(c *handlerUI) {

	if c.index.loginPassword < len(c.data.namesLoginPassword)-1 {
		c.index.loginPassword++
	}
}

// Увеличение значения индекса для имён текст массива.
//
// Параметры:
//
//	с - конфигурация.
func incrIndexNamesText(c *handlerUI) {

	if c.index.text < len(c.data.namesText)-1 {
		c.index.text++
	}
}

// Увеличение значения индекса для имён массива банковских карт.
//
// Параметры:
//
//	с - конфигурация.
func incrIndexNamesBankCard(c *handlerUI) {

	if c.index.bankCard < len(c.data.namesBankCard)-1 {
		c.index.bankCard++
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

// Увеличение значения индекса для массива файлов.
//
// Параметры:
//
//	с - конфигурация.
func incrIndexNamesFile(c *handlerUI) {

	if c.index.file < len(c.data.namesFile)-1 {
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

// Уменьшение значения индекса для массива имён текста.
//
// Параметры:
//
//	с - конфигурация.
func decrIndexNamesText(c *handlerUI) {

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

// Уменьшение значения индекса для массива имён банковских карт.
//
// Параметры:
//
//	с - конфигурация.
func decrIndexBankCardName(c *handlerUI) {

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

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

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
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		name := "indicatorReadStatus"

		// Обработка индикатора получения данных.
		indicatorRead, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if indicatorRead == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}

		if c.status.readNameLoginPaaswordPassed { // обработка при чтении
			if c.status.readNameLoginPaaswordSUCCESS {
				indicatorRead.Clear()
				indicatorRead.Write([]byte(fmt.Sprintf("Всего записей: %d", len(c.data.namesLoginPassword))))
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

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

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
		} else {
			indicator.Clear()
			indicator.Write([]byte(""))
		}
		return nil
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		// Обработка индикатора получения данных.
		name := "indicatorReadStatus"

		indicatorRead, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if indicatorRead == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}

		if c.status.readNameTextPassed { // обработка при чтении
			if c.status.readNameTextSUCCESS {
				indicatorRead.Clear()
				indicatorRead.Write([]byte(fmt.Sprintf("Всего записей: %d", len(c.data.namesText))))
				indicatorRead.FgColor = gocui.ColorGreen
				indicatorRead.BgColor = gocui.ColorDefault
			} else {
				indicatorRead.Clear()
				indicatorRead.Write([]byte("Ошибка"))
				indicatorRead.FgColor = gocui.ColorRed
				indicatorRead.BgColor = gocui.ColorDefault
			}
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
		} else {
			indicator.Clear()
			indicator.Write([]byte(""))
		}
		return nil
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

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

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

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		// Обработка индикатора получения данных.
		name := "indicatorReadStatus"

		indicatorRead, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if indicatorRead == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}
		if c.status.readNameBankCardPassed { // обработка при чтении
			if c.status.readNameBankCardSUCCESS {
				indicatorRead.Clear()
				indicatorRead.Write([]byte(fmt.Sprintf("Всего записей: %d", len(c.data.namesBankCard))))
				indicatorRead.FgColor = gocui.ColorGreen
				indicatorRead.BgColor = gocui.ColorDefault
			} else {
				indicatorRead.Clear()
				indicatorRead.Write([]byte("Ошибка"))
				indicatorRead.FgColor = gocui.ColorRed
				indicatorRead.BgColor = gocui.ColorDefault
			}
			c.status.readNameBankCardPassed = false
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

	return nil
}

// Обновление цвета у индикторов окна viewBinaryData . Возвращается ошибка.
//
// Параметры:
//
//	g - указатель на интерфейс.
//	с - указатель на конфигурацию.
func indicatorViewBinaryData(g *gocui.Gui, c *handlerUI) error {

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		// Добавление файла в контейнер.
		statusPushContainer := c.getStatusPushContainer()

		if statusPushContainer != stageNotActive {
			name := "Save"
			btnSave, err := g.View(name)
			if err != nil {
				return fmt.Errorf("Функция View, вернула ошибку: <%w>", err)
			}
			if btnSave == nil {
				return fmt.Errorf("Нет указателя на элемент: <%s>", name)
			}

			switch statusPushContainer {
			case stageNotActive:
				btnSave.FgColor = gocui.ColorWhite
			case stageActive:
				btnSave.FgColor = gocui.ColorYellow
			case stageOk:
				btnSave.FgColor = gocui.ColorGreen
			case stageFault:
				btnSave.FgColor = gocui.ColorRed
			}
		}

		// Извлечение файла из контейнера.
		statusPopContainer := c.getStatusPopContainer()

		if statusPopContainer != stageNotActive {
			name := "Extraction"
			btnExtration, err := g.View(name)
			if err != nil {
				return fmt.Errorf("Функция View, вернула ошибку: <%w>", err)
			}
			if btnExtration == nil {
				return fmt.Errorf("Нет указателя на элемент: <%s>", name)
			}

			switch statusPopContainer {
			case stageNotActive:
				btnExtration.FgColor = gocui.ColorWhite
			case stageActive:
				btnExtration.FgColor = gocui.ColorYellow
			case stageOk:
				btnExtration.FgColor = gocui.ColorGreen
			case stageFault:
				btnExtration.FgColor = gocui.ColorRed
			}
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

		//
		// --- Индикатор процентов ---
		//

		if statusPushContainer != stageNotActive || statusPopContainer != stageNotActive {

			name := "indicatorPercent"
			indicator, err := g.View(name)
			if err != nil {
				return fmt.Errorf("Функция View, вернула ошибку: <%w>", err)
			}
			if indicator == nil {
				return fmt.Errorf("Нет указателя на элемент: <%s>", name)
			}

			indicator.Clear()
			indicator.Write([]byte(fmt.Sprintf("Выполнено: %.2f%%", c.getPercentTxRx())))
		}

		return nil
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		// Передача файла на сервер.
		statusTx := c.getStatusFileTx()

		if statusTx != stageNotActive {
			name := "Save"
			btnSave, err := g.View(name)
			if err != nil {
				return fmt.Errorf("Функция View, вернула ошибку: <%w>", err)
			}
			if btnSave == nil {
				return fmt.Errorf("Нет указателя на элемент: <%s>", name)
			}

			switch statusTx {
			case stageNotActive:
				btnSave.FgColor = gocui.ColorWhite
			case stageActive:
				btnSave.FgColor = gocui.ColorYellow
			case stageOk:
				btnSave.FgColor = gocui.ColorGreen
			case stageFault:
				btnSave.FgColor = gocui.ColorRed
			}
		}

		// Приём файла от сервера.
		statusRx := c.getStatusFileRx()

		if statusRx != stageNotActive {
			name := "Extraction"
			btnExtration, err := g.View(name)
			if err != nil {
				return fmt.Errorf("Функция View, вернула ошибку: <%w>", err)
			}
			if btnExtration == nil {
				return fmt.Errorf("Нет указателя на элемент: <%s>", name)
			}

			switch statusRx {
			case stageNotActive:
				btnExtration.FgColor = gocui.ColorWhite
			case stageActive:
				btnExtration.FgColor = gocui.ColorYellow
			case stageOk:
				btnExtration.FgColor = gocui.ColorGreen
			case stageFault:
				btnExtration.FgColor = gocui.ColorRed
			}
		}

		// Есть установлен признак удаления файла на сервере.
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

		// Если чтение имён файлов выполнено.
		if c.status.readNameFilePassed {
			name := "indicatorReadStatus"

			indicator, err := g.View(name)
			if err != nil {
				return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
			}
			if indicator == nil {
				return fmt.Errorf("Нет указателя на элемент: <%s>", name)
			}

			if c.status.readNameFileSUCCESS {

				indicator.Clear()
				indicator.Write([]byte(fmt.Sprintf("Всего файлов: %d", len(c.data.namesFile))))
				indicator.FgColor = gocui.ColorGreen
				indicator.BgColor = gocui.ColorDefault
			} else {
				indicator.Clear()
				indicator.Write([]byte("Ошибка чтения."))
				indicator.FgColor = gocui.ColorRed
				indicator.BgColor = gocui.ColorDefault
			}

			c.status.readNameFilePassed = false  // Для разовой отработки при открытии экрана.
			c.status.readNameFileSUCCESS = false // Для разовой отработки при открытии экрана.
		}

		//
		// --- Индикатор процентов ---
		//

		if statusTx != stageNotActive || statusRx != stageNotActive {

			name := "indicatorPercent"
			indicator, err := g.View(name)
			if err != nil {
				return fmt.Errorf("Функция View, вернула ошибку: <%w>", err)
			}
			if indicator == nil {
				return fmt.Errorf("Нет указателя на элемент: <%s>", name)
			}

			indicator.Clear()
			indicator.Write([]byte(fmt.Sprintf("Выполнено: %.2f%%", c.getPercentTxRx())))
		}
		return nil
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

	statusBackUp := c.getStatusBackUp()
	statusRestore := c.getStatusRestore()

	// Обработка в режиме - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

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

			name := "indicatorPercent"
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

	}

	// Обработка в режиме - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		//
		// --- Индикатор имени клиента ---
		//

		name := "indicatorNameClient"
		indicator, err := g.View(name)
		if err != nil {
			return fmt.Errorf("Фнукция View, вернула ошибку: <%w>", err)
		}
		if indicator == nil {
			return fmt.Errorf("Нет указателя на элемент: <%s>", name)
		}

		indicator.Clear()
		indicator.Write([]byte(fmt.Sprintf("ID клиента: %s", c.clientName)))
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

	// Создание токена.
	secretKey, txToken, err := layerRestoreCreateToken()
	if err != nil {
		return fmt.Errorf("Функция layerRestoreCreateToken, вернула ошибку: <%w>", err)
	}

	// Запрос информации по файлам у сервера.
	filesInfo, rxToken, err := layerFilesInfoRequest(client, txToken)
	if err != nil {
		return fmt.Errorf("Функция layerFilesInfoRequest, вернула ошибку: <%w>", err)
	}

	// Проверка результата запроса.
	if len(filesInfo) == 0 {
		return EmptyData
	}
	if err := layerReqFilesCheckResult(rxToken, secretKey); err != nil {
		return fmt.Errorf("Функция layerReqFilesCheckResult, вернула ошибку: <%w>", err)
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

// Процесс restrore.
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

// Добавление превикса к имени имени файла. Возвращается новое имя и ошибка.
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

// Воссстановление имени файла. Возвращается ошибка.
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

// Воостановление состояния файлов процесса restore, при ошибке в последовательности. Возвращается ошибка.
func deferProcessRestoreByError(dooRestore bool, fileName, tempFileName string, errProcess error) error {

	if errProcess != nil {
		// Удаление принятого файла.
		if isFileExists(fileName) && dooRestore {
			if err := os.Remove(fileName); err != nil {
				return fmt.Errorf("Error: ошибка:<%v>, при удалении файла:<%s> по ошибке процесса:<%v>", err, fileName, err)
			}
		}
		// Восстановление имени у исходного файла.
		if isFileExists(tempFileName) && dooRestore {
			if err := restoreFileName(tempFileName, "-temp"); err != nil {
				return fmt.Errorf("Error: ошибка:<%v>, при восстановлении файла:<%s> по ошибке процесса:<%v>", err, tempFileName, err)
			}
		}
	}
	return nil
}

// Функция принимает данные по каналам и транслирует их в экземпляр.
func bufferProcessPushContainer(c *handlerUI, rxChProcess <-chan float64, rxChErr <-chan error, rxChOk <-chan struct{}) {

	defer func() {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Завершён процесс добавления в контейнер файла:<%s>", c.typed.dataPathSrc))
	}()

	for {
		select {
		case percent, ok := <-rxChProcess:
			if !ok {
				c.updateStatusPushContainer(stageNotActive)
				return
			}
			c.setPercentTxRx(float32(percent))

		case err, ok := <-rxChErr:
			if !ok {
				c.updateStatusPushContainer(stageNotActive)
				return
			}
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка процесса добавления файла в контейнер:<%v>", err))
			c.updateStatusPushContainer(stageFault)
			return

		case <-rxChOk:
			c.updateStatusPushContainer(stageOk)
			return
		}
	}
}

// Приём данных и сборка файла.
func bufferProcessPopContainer(nameFile string, c *handlerUI, rxChProcess <-chan float64, rxChErr <-chan error, rxChDone <-chan struct{}, rxChData <-chan []byte, txChBreak chan<- struct{}) {

	var outFile *os.File
	var err error
	fileExists := false
	fullNameFile := c.typed.dataPathTrg + nameFile

	defer func(fileName, fullNameFile string, file *os.File, fileExist bool) {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Завершён процесс извлечения из контейнера файла:<%s>", fileName))
		close(txChBreak)

		// Закрытие подключения к файлу
		if fileExist {
			if err := file.Close(); err != nil {
				c.updateStatusPopContainer(stageFault)
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка:<%v>, при закрытии подключения к файлу:<%s>", err, fullNameFile))
				return
			}
		}

	}(nameFile, fullNameFile, outFile, fileExists)

	// Проверка существования файла.
	if isFileExists(fullNameFile) {
		c.updateStatusPopContainer(stageFault)
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Warn: Файл:<%s>, уже существует", fullNameFile))
		fileExists = true

		txChBreak <- struct{}{} // Передача сигнала - прекратить процесс.
	}

	// Создание файла.
	if !fileExists {
		outFile, err = os.Create(fullNameFile)
		if err != nil {
			c.updateStatusPopContainer(stageFault)
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Не удалось создать файл:<%s>, ошибка:<%v>", fullNameFile, err))
			return
		}
	}

	// Обработка каналов.
	for {
		select {
		case percent, ok := <-rxChProcess:
			if !ok {
				c.updateStatusPopContainer(stageNotActive)
				c.conf.PtrLoggerFile.Write("Error: Неожиданное закрытие канала rxChProcess")
				return
			}
			c.setPercentTxRx(float32(percent))

		case err, ok := <-rxChErr:
			if !ok {
				c.updateStatusPopContainer(stageNotActive)
				c.conf.PtrLoggerFile.Write("Error: Неожиданное закрытие канала rxChErr")
				return
			}
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка процесса извлечения файла из контейнера:<%v>", err))
			c.updateStatusPopContainer(stageFault)
			return

		case _, ok := <-rxChDone:
			if !ok {
				c.updateStatusPopContainer(stageNotActive)
				c.conf.PtrLoggerFile.Write("Error: Неожиданное закрытие канала rxChDone")
				return
			}
			c.updateStatusPopContainer(stageOk)
			return

		// Сборка файла.
		case data, ok := <-rxChData:
			if !ok {
				c.updateStatusPopContainer(stageNotActive)
				c.conf.PtrLoggerFile.Write("Error: Неожиданное закрытие канала rxChData")
				return
			}
			if !fileExists {
				if _, err := outFile.Write(data); err != nil {
					c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Не удалось записать данные в файл:<%s>, ошибка:<%v>", fullNameFile, err))
					c.updateStatusPopContainer(stageFault)
					return
				}
			}
		}
	}
}

// Приём данных процесса получения файла от сервера.
func bufferProcessRxFileByName(c *handlerUI, rxChProcess <-chan float32, rxChErr <-chan error, rxChDone <-chan struct{}) {

	// Обработка каналов.
	for {
		select {
		// Проценты процесса.
		case percent, ok := <-rxChProcess:
			if !ok {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("--- Проент выполнения: <%.2f>", percent)) // ---------------------------------------------------

				c.updateStatusFileRx(stageNotActive)
				c.conf.PtrLoggerFile.Write("Error: Неожиданное закрытие канала rxChProcess")
				return
			}
			c.setPercentTxRx(float32(percent))

			// Ошибка.
		case err, ok := <-rxChErr:
			if !ok {
				c.updateStatusFileRx(stageNotActive)
				c.conf.PtrLoggerFile.Write("Error: Неожиданное закрытие канала rxChErr")
				return
			}
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка процесса приёма файла от сервера:<%v>", err))
			c.updateStatusFileRx(stageFault)
			return

			// Приём выполнен.
		case _, ok := <-rxChDone:
			if !ok {
				c.updateStatusFileRx(stageNotActive)
				c.conf.PtrLoggerFile.Write("Error: Неожиданное закрытие канала rxChDone")
				return
			}
			c.updateStatusFileRx(stageOk)
			c.conf.PtrLoggerFile.Write("Info: файл успешно принят")
			return
		}
	}
}

// Логика процесса сохранения в окне BinaryData. Возвращается ошибка.
func doStoreViewBinaryData(c *handlerUI) error {

	// Если режим - локальный
	if c.conf.Flag.Mode == flags.ModeLocal {
		if c.getStatusPopContainer() == stageNotActive && c.getStatusPushContainer() == stageNotActive {

			c.updateStatusPushContainer(stageActive)
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Запущен процесс добавления в контейнер файла:<%s>", c.typed.dataPathSrc))

			// Сброс
			c.txrx.passedKB = 0
			c.txrx.percentTxRx = 0

			// Определение размера файла.
			var err error
			files := []string{c.typed.dataPathSrc}

			c.txrx.totalSizeKB, err = totalFileSize(files)
			if err != nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция totalFileSize, вернула ошибку: <%v>", err))
				c.updateStatusBackUp(stageFault)
				return nil
			}

			chPercent := make(chan float64)
			chErr := make(chan error)
			chDone := make(chan struct{})

			// Запуск процесса передачи файла в контейнер.
			go c.conf.Container.AddFileToContainer(c.typed.dataPathSrc, c.secret.secretKey, chPercent, chErr, chDone)

			// Буфер между каналами и экземпляром.
			go bufferProcessPushContainer(c, chPercent, chErr, chDone)
		}
	}

	// Если режим - удалённый
	if c.conf.Flag.Mode == flags.ModeRemote {

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Запущен процесс передачи на сервер файла:<%s>", c.typed.dataPathSrc))

		if err := c.conf.Server.SendFile(c.typed.dataPathSrc, c.conf.Server.GetTokenAuthentication(), c.clientName, c.secret.secretKey); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция Server.SendFile, вернула ошибку: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Завершен процесс передачи на сервер файла:<%s>", c.typed.dataPathSrc))
	}

	return nil
}

// Логика процесса сохранения в окне ViewBankCardData. Возвращается ошибка.
func doStoreViewBankCardData(c *handlerUI) error {

	c.status.addBankCardPassed = true
	c.status.addBankCardSUCCESS = false

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		// Шифрование данных
		encrFor, err := encrypt(c.typed.dataFor, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataFor: <%v>", err))
			return nil
		}
		encrOwner, err := encrypt(c.typed.dataOwner, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataOwner: <%v>", err))
			return nil
		}
		encrNumb, err := encrypt(c.typed.dataNumb, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataNumb: <%v>", err))
			return nil
		}
		encrValid, err := encrypt(c.typed.dataValidDate, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataValidDate: <%v>", err))
			return nil
		}
		encrCode, err := encrypt(c.typed.dataCode, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataCode: <%v>", err))
			return nil
		}
		tn := time.Now().UTC()
		strT := tn.Format(time.RFC3339)
		encrCreatedAt, err := encrypt(strT, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого времени создания: <%v>", err))
			return nil
		}

		// Проверка номера банковской карты на валидность.
		if !checkCardNumber(c.typed.dataNumb) {
			c.conf.PtrLoggerFile.Write("Error: номер карты, не прошел проверку")
			return nil
		}

		// Добавление зашифрованных данных в БД.
		if err := c.conf.DataBase.AddDataBankCardContext(ctx, encrFor, encrOwner, encrNumb, encrValid, encrCode, encrCreatedAt); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка добавления карты в БД: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Debug: карта добавлена в БД")
		c.status.addBankCardSUCCESS = true
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		// Дата создания записи.
		tn := time.Now().UTC()
		strT := tn.Format(time.RFC3339)

		// Подготовка.
		txData := server.TxBankCard{
			ID:        c.clientName,
			For:       c.typed.dataFor,
			Owner:     c.typed.dataOwner,
			Numb:      c.typed.dataNumb,
			ValidData: c.typed.dataValidDate,
			Code:      c.typed.dataCode,
			CreatedAt: strT,
		}

		// Логика.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token := c.conf.Server.GetTokenAuthentication()

		if err := c.conf.Server.SendBankCard(ctx, txData, token, c.secret.secretKey); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция SendBankCard, вернула ошибку: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Info: банковская карта, успешно передана на сервер")
		c.status.addBankCardSUCCESS = true
	}

	return nil
}

// Логика процесса сохранения в окне ViewTextData. Возвращается ошибка.
func doStoreViewTextData(c *handlerUI) error {

	c.status.addTextPassed = true
	c.status.addTextSUCCESS = false

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Шифрование данных
		encrFor, err := encrypt(c.typed.dataFor, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataFor: <%v>", err))
			return nil
		}
		encrText, err := encrypt(c.typed.dataText, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataText: <%v>", err))
			return nil
		}
		tn := time.Now().UTC()
		strT := tn.Format(time.RFC3339)
		encrCreatedAt, err := encrypt(strT, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого времени создания: <%v>", err))
			return nil
		}
		// Добавление зашифрованных данных в БД.
		if err := c.conf.DataBase.AddDataTextContext(ctx, encrFor, encrText, encrCreatedAt); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка добавления текста в БД: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Debug: текст добавлен в БД")
		c.status.addTextSUCCESS = true
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		// Дата создания записи.
		tn := time.Now().UTC()
		strT := tn.Format(time.RFC3339)

		// Подготовка.
		txData := server.TxText{
			ID:        c.clientName,
			For:       c.typed.dataFor,
			Text:      c.typed.dataText,
			CreatedAt: strT,
		}

		// Логика.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token := c.conf.Server.GetTokenAuthentication()

		if err := c.conf.Server.SendText(ctx, txData, token, c.secret.secretKey); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция SendText, вернула ошибку: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Info: текст, успешно передан на сервер")
		c.status.addTextSUCCESS = true
	}

	return nil
}

// Логика процесса сохранения в окне ViewLoginPasswordData. Возвращается ошибка.
func doStoreViewLoginPasswordData(c *handlerUI) error {

	c.status.addLoginPaaswordPassed = true
	c.status.addLoginPaaswordSUCCESS = false

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Шифрование данных
		encrFor, err := encrypt(c.typed.dataFor, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataFor: <%v>", err))
			return nil
		}
		encrLogin, err := encrypt(c.typed.dataLogin, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataLogin: <%v>", err))
			return nil
		}
		encrPassword, err := encrypt(c.typed.dataPassword, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого dataPassword: <%v>", err))
			return nil
		}
		tn := time.Now().UTC()
		strT := tn.Format(time.RFC3339)
		encrCreatedAt, err := encrypt(strT, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка шифрования содержимого времени создания: <%v>", err))
			return nil
		}

		// Добавление зашифрованных данных в БД.
		if err := c.conf.DataBase.AddDataLoginPasswordContext(ctx, encrFor, encrLogin, encrPassword, encrCreatedAt); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка добавления пары логин/пароль в БД: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Info: пара логин/пароль добавлена в БД")
		c.status.addLoginPaaswordSUCCESS = true
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		// Дата создания записи.
		tn := time.Now().UTC()
		strT := tn.Format(time.RFC3339)

		// Подготовка.
		txData := server.TxLoginPassword{
			ID:        c.clientName,
			For:       c.typed.dataFor,
			Login:     c.typed.dataLogin,
			Password:  c.typed.dataPassword,
			CreatedAt: strT,
		}

		// Логика.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token := c.conf.Server.GetTokenAuthentication()

		if err := c.conf.Server.SendLoginPassword(ctx, txData, token, c.secret.secretKey); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция SendLoginPassword, вернула ошибку: <%v>", err))
			return nil
		}

		c.conf.PtrLoggerFile.Write("Info: пара логин/пароль, успешно передана на сервер")
		c.status.addLoginPaaswordSUCCESS = true
	}

	return nil
}

// Логика перевода фокуса в окне viewLoginPasswordData. Возвращается ошибка.
func doShowNextElementViewLoginPasswordData(c *handlerUI, gui *gocui.Gui) error {

	// Если режим - локальный
	if c.conf.Flag.Mode == flags.ModeLocal {

		if len(c.data.loginPassword) == 0 {
			return nil
		}

		el := loginPasswordByIndex(c) // получение записи по индексу
		incrIndexloginPassword(c)     // увеличение значения индекса

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el.name != "" {
			fieldName.Clear()
			fieldName.Write([]byte(el.name))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Login.
		fieldLogin, err := gui.View("fieldShowLogin")
		if err != nil || fieldLogin == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowLogin: <%v>", err))
			return nil
		}
		if el.login != "" {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(el.login))

		} else {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(""))
		}

		// отображение содержимого поля Password.
		fieldPassword, err := gui.View("fieldShowPassword")
		if err != nil || fieldPassword == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowPassword: <%v>", err))
			return nil
		}
		if el.password != "" {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(el.password))

		} else {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(""))
		}
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		if len(c.data.namesLoginPassword) == 0 {
			return nil
		}

		name := nameLoginPasswordByIndex(c) // получение имени записи по индексу
		incrIndexNamesloginPassword(c)      // увеличение значения индекса

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Запрос данных логин/пароль по имени: <%s>", name))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Запрос логин/пароль у сервера, по имени записи
		rxData, err := c.conf.Server.RequestLoginPasswordByName(ctx, c.conf.Server.GetTokenAuthentication(), c.clientName, name, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция RequestLoginPasswordByName, вернула ошибку: <%v>", err))
			return nil
		}

		// ---------------

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if rxData.For != "" {
			fieldName.Clear()
			fieldName.Write([]byte(rxData.For))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Login.
		fieldLogin, err := gui.View("fieldShowLogin")
		if err != nil || fieldLogin == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowLogin: <%v>", err))
			return nil
		}
		if rxData.Login != "" {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(rxData.Login))

		} else {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(""))
		}

		// отображение содержимого поля Password.
		fieldPassword, err := gui.View("fieldShowPassword")
		if err != nil || fieldPassword == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowPassword: <%v>", err))
			return nil
		}
		if rxData.Password != "" {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(rxData.Password))

		} else {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(""))
		}
	}

	return nil
}

// Логика перевода фокуса в окне viewTextData. Возвращается ошибка.
func doShowNextElementViewTextData(c *handlerUI, gui *gocui.Gui) error {

	// Если режим - локальный
	if c.conf.Flag.Mode == flags.ModeLocal {

		if len(c.data.textData) == 0 {
			return nil
		}

		el := textByIndex(c) // получение записи по индексу
		incrIndexText(c)     // увеличение значения индекса

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el.name != "" {
			fieldName.Clear()
			fieldName.Write([]byte(el.name))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Login.
		fieldText, err := gui.View("fieldShowText")
		if err != nil || fieldText == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowText: <%v>", err))
			return nil
		}
		if el.text != "" {
			fieldText.Clear()
			fieldText.Write([]byte(el.text))

		} else {
			fieldText.Clear()
			fieldText.Write([]byte(""))
		}

		return nil
	}

	// Если режим - локальный
	if c.conf.Flag.Mode == flags.ModeRemote {

		if len(c.data.namesText) == 0 {
			return nil
		}

		name := nameTextByIndex(c) // получение записи по индексу
		incrIndexNamesText(c)      // увеличение значения индекса

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Запрос данных текста по имени: <%s>", name))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Запрос текста у сервера, по имени записи
		rxData, err := c.conf.Server.RequestTextByName(ctx, c.conf.Server.GetTokenAuthentication(), c.clientName, name, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция RequestTextByName, вернула ошибку: <%v>", err))
			return nil
		}

		// ----------------

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if rxData.For != "" {
			fieldName.Clear()
			fieldName.Write([]byte(rxData.For))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Login.
		fieldText, err := gui.View("fieldShowText")
		if err != nil || fieldText == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowText: <%v>", err))
			return nil
		}
		if rxData.Text != "" {
			fieldText.Clear()
			fieldText.Write([]byte(rxData.Text))

		} else {
			fieldText.Clear()
			fieldText.Write([]byte(""))
		}

		return nil
	}

	return nil
}

// Логика перевода фокуса в окне viewBankCardData. Возвращается ошибка.
func doShowNextElementViewBankCardData(c *handlerUI, gui *gocui.Gui) error {

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		if len(c.data.bankCard) == 0 {
			return nil
		}

		el := bankCardByIndex(c) // получение записи по индексу
		incrIndexBankCard(c)     // увеличение значения индекса

		// отображение содержимого поля Для.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el.name != "" {
			fieldName.Clear()
			fieldName.Write([]byte(el.name))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Владелец.
		fieldOwner, err := gui.View("fieldShowOwner")
		if err != nil || fieldOwner == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowOwner: <%v>", err))
			return nil
		}
		if el.owner != "" {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(el.owner))

		} else {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(""))
		}

		// отображение содержимого поля Номер.
		fieldNumb, err := gui.View("fieldShowNumber")
		if err != nil || fieldNumb == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowNumber: <%v>", err))
			return nil
		}
		if el.numb != "" {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(el.numb))

		} else {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(""))
		}

		// отображение содержимого поля Валидность.
		fieldValid, err := gui.View("fieldShowValid")
		if err != nil || fieldValid == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowValid: <%v>", err))
			return nil
		}
		if el.valid != "" {
			fieldValid.Clear()
			fieldValid.Write([]byte(el.valid))

		} else {
			fieldValid.Clear()
			fieldValid.Write([]byte(""))
		}

		// отображение содержимого поля Код.
		fieldCode, err := gui.View("fieldShowCode")
		if err != nil || fieldCode == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowCode: <%v>", err))
			return nil
		}
		if el.code != "" {
			fieldCode.Clear()
			fieldCode.Write([]byte(el.code))

		} else {
			fieldCode.Clear()
			fieldCode.Write([]byte(""))
		}
		return nil
	}
	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		if len(c.data.namesBankCard) == 0 {
			return nil
		}

		name := nameBankCardByIndex(c) // получение записи по индексу
		incrIndexNamesBankCard(c)      // увеличение значения индекса

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Запрос данных банковской карты по имени: <%s>", name))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Запрос банковской карты у сервера, по имени записи
		rxData, err := c.conf.Server.RequestBankCardByName(ctx, c.conf.Server.GetTokenAuthentication(), c.clientName, name, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция RequestBankCardByName, вернула ошибку: <%v>", err))
			return nil
		}

		// -------------------

		// отображение содержимого поля Для.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if rxData.For != "" {
			fieldName.Clear()
			fieldName.Write([]byte(rxData.For))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Владелец.
		fieldOwner, err := gui.View("fieldShowOwner")
		if err != nil || fieldOwner == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowOwner: <%v>", err))
			return nil
		}
		if rxData.Owner != "" {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(rxData.Owner))

		} else {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(""))
		}

		// отображение содержимого поля Номер.
		fieldNumb, err := gui.View("fieldShowNumber")
		if err != nil || fieldNumb == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowNumber: <%v>", err))
			return nil
		}
		if rxData.Numb != "" {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(rxData.Numb))

		} else {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(""))
		}

		// отображение содержимого поля Валидность.
		fieldValid, err := gui.View("fieldShowValid")
		if err != nil || fieldValid == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowValid: <%v>", err))
			return nil
		}
		if rxData.Valid != "" {
			fieldValid.Clear()
			fieldValid.Write([]byte(rxData.Valid))

		} else {
			fieldValid.Clear()
			fieldValid.Write([]byte(""))
		}

		// отображение содержимого поля Код.
		fieldCode, err := gui.View("fieldShowCode")
		if err != nil || fieldCode == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowCode: <%v>", err))
			return nil
		}
		if rxData.Code != "" {
			fieldCode.Clear()
			fieldCode.Write([]byte(rxData.Code))

		} else {
			fieldCode.Clear()
			fieldCode.Write([]byte(""))
		}
		return nil
	}
	return nil
}

// Логика перевода фокуса в окне viewBinaryData. Возвращается ошибка.
func doShowNextElementViewBinaryData(c *handlerUI, gui *gocui.Gui) error {

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		if c.getStatusPopContainer() != stageActive && c.getStatusPushContainer() != stageActive {

			c.updateStatusPopContainer(stageNotActive)
			c.updateStatusPushContainer(stageNotActive)

			if len(c.data.files) == 0 {
				return nil
			}

			el := fileByIndex(c) // получение записи по индексу
			incrIndexFile(c)     // увеличение значения индекса

			// отображение содержимого поля Код.
			fieldCode, err := gui.View("fieldShowFor")
			if err != nil || fieldCode == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
				return nil
			}
			if el != "" {
				fieldCode.Clear()
				fieldCode.Write([]byte(el))

			} else {
				fieldCode.Clear()
				fieldCode.Write([]byte(""))
			}
		}
		return nil
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		if c.getStatusFileTx() != stageActive && c.getStatusFileRx() != stageActive {

			c.updateStatusFileTx(stageNotActive)
			c.updateStatusFileRx(stageNotActive)

			if len(c.data.namesFile) == 0 {
				return nil
			}

			el := fileNameByIndex(c) // получение записи по индексу
			incrIndexNamesFile(c)    // увеличение значения индекса

			// отображение содержимого поля Код.
			fieldCode, err := gui.View("fieldShowFor")
			if err != nil || fieldCode == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
				return nil
			}
			if el != "" {
				fieldCode.Clear()
				fieldCode.Write([]byte(el))

			} else {
				fieldCode.Clear()
				fieldCode.Write([]byte(""))
			}
		}
		return nil
	}

	return nil
}

// Логика перевода фокуса в окне viewLoginPasswordData. Возвращается ошибка.
func doShowPrevElementViewLoginPasswordData(c *handlerUI, gui *gocui.Gui) error {

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		decrIndexloginPassword(c)     // уменьшение значения индекса
		el := loginPasswordByIndex(c) // получение записи по индексу

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el.name != "" {
			fieldName.Clear()
			fieldName.Write([]byte(el.name))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Login.
		fieldLogin, err := gui.View("fieldShowLogin")
		if err != nil || fieldLogin == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowLogin: <%v>", err))
			return nil
		}
		if el.login != "" {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(el.login))

		} else {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(""))
		}

		// отображение содержимого поля Password.
		fieldPassword, err := gui.View("fieldShowPassword")
		if err != nil || fieldPassword == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowPassword: <%v>", err))
			return nil
		}
		if el.password != "" {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(el.password))

		} else {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(""))
		}

		return nil
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		name := nameLoginPasswordByIndex(c) // получение имени записи по индексу
		decrIndexloginPassword(c)           // увеличение значения индекса

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Запрос данных логин/пароль по имени: <%s>", name))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Запрос логин/пароль у сервера, по имени записи
		rxData, err := c.conf.Server.RequestLoginPasswordByName(ctx, c.conf.Server.GetTokenAuthentication(), c.clientName, name, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция RequestLoginPasswordByName, вернула ошибку: <%v>", err))
			return nil
		}

		// ---------------------

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if rxData.For != "" {
			fieldName.Clear()
			fieldName.Write([]byte(rxData.For))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Login.
		fieldLogin, err := gui.View("fieldShowLogin")
		if err != nil || fieldLogin == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowLogin: <%v>", err))
			return nil
		}
		if rxData.Login != "" {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(rxData.Login))

		} else {
			fieldLogin.Clear()
			fieldLogin.Write([]byte(""))
		}

		// отображение содержимого поля Password.
		fieldPassword, err := gui.View("fieldShowPassword")
		if err != nil || fieldPassword == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowPassword: <%v>", err))
			return nil
		}
		if rxData.Password != "" {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(rxData.Password))

		} else {
			fieldPassword.Clear()
			fieldPassword.Write([]byte(""))
		}

		return nil
	}

	return nil
}

// Логика перевода фокуса в окне viewTextData. Возвращается ошибка.
func doShowPrevElementViewTextData(c *handlerUI, gui *gocui.Gui) error {

	// Если режим - локальный
	if c.conf.Flag.Mode == flags.ModeLocal {

		decrIndexText(c)     // уменьшение значения индекса
		el := textByIndex(c) // получение записи по индексу

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el.name != "" {
			fieldName.Clear()
			fieldName.Write([]byte(el.name))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Login.
		fieldText, err := gui.View("fieldShowText")
		if err != nil || fieldText == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowText: <%v>", err))
			return nil
		}
		if el.text != "" {
			fieldText.Clear()
			fieldText.Write([]byte(el.text))

		} else {
			fieldText.Clear()
			fieldText.Write([]byte(""))
		}
		return nil
	}

	// Если режим - удалённый
	if c.conf.Flag.Mode == flags.ModeRemote {

		decrIndexNamesText(c)      // уменьшение значения индекса
		name := nameTextByIndex(c) // получение записи по индексу

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Запрос данных текста по имени: <%s>", name))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Запрос текста у сервера, по имени записи
		rxData, err := c.conf.Server.RequestTextByName(ctx, c.conf.Server.GetTokenAuthentication(), c.clientName, name, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция RequestTextByName, вернула ошибку: <%v>", err))
			return nil
		}

		// ---------------

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if name != "" {
			fieldName.Clear()
			fieldName.Write([]byte(rxData.For))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля текст.
		fieldText, err := gui.View("fieldShowText")
		if err != nil || fieldText == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowText: <%v>", err))
			return nil
		}
		if rxData.Text != "" {
			fieldText.Clear()
			fieldText.Write([]byte(rxData.Text))

		} else {
			fieldText.Clear()
			fieldText.Write([]byte(""))
		}
		return nil
	}
	return nil
}

// Логика перевода фокуса в окне viewBankCardData. Возвращается ошибка.
func doShowPrevElementViewBankCardData(c *handlerUI, gui *gocui.Gui) error {

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {

		decrIndexBankCard(c)     // уменьшение значения индекса
		el := bankCardByIndex(c) // получение записи по индексу

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if el.name != "" {
			fieldName.Clear()
			fieldName.Write([]byte(el.name))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Владелец.
		fieldOwner, err := gui.View("fieldShowOwner")
		if err != nil || fieldOwner == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowOwner: <%v>", err))
			return nil
		}
		if el.owner != "" {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(el.owner))

		} else {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(""))
		}

		// отображение содержимого поля Номер.
		fieldNumb, err := gui.View("fieldShowNumber")
		if err != nil || fieldNumb == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowNumber: <%v>", err))
			return nil
		}
		if el.numb != "" {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(el.numb))

		} else {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(""))
		}

		// отображение содержимого поля Валидность.
		fieldValid, err := gui.View("fieldShowValid")
		if err != nil || fieldValid == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowValid: <%v>", err))
			return nil
		}
		if el.valid != "" {
			fieldValid.Clear()
			fieldValid.Write([]byte(el.valid))

		} else {
			fieldValid.Clear()
			fieldValid.Write([]byte(""))
		}

		// отображение содержимого поля Код.
		fieldCode, err := gui.View("fieldShowCode")
		if err != nil || fieldCode == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowCode: <%v>", err))
			return nil
		}
		if el.code != "" {
			fieldCode.Clear()
			fieldCode.Write([]byte(el.code))

		} else {
			fieldCode.Clear()
			fieldCode.Write([]byte(""))
		}

		return nil
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		if len(c.data.namesBankCard) == 0 {
			return nil
		}

		decrIndexBankCardName(c)       // уменьшение значения индекса
		name := nameBankCardByIndex(c) // Получение значения по индексу

		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: Запрос данных банковской карты по имени: <%s>", name))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Запрос банковской карты у сервера, по имени записи
		rxData, err := c.conf.Server.RequestBankCardByName(ctx, c.conf.Server.GetTokenAuthentication(), c.clientName, name, c.secret.secretKey)
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Функция RequestBankCardByName, вернула ошибку: <%v>", err))
			return nil
		}

		// -------------------

		// отображение содержимого поля For.
		fieldName, err := gui.View("fieldShowFor")
		if err != nil || fieldName == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
			return nil
		}
		if rxData.For != "" {
			fieldName.Clear()
			fieldName.Write([]byte(rxData.For))

		} else {
			fieldName.Clear()
			fieldName.Write([]byte(""))
		}

		// отображение содержимого поля Владелец.
		fieldOwner, err := gui.View("fieldShowOwner")
		if err != nil || fieldOwner == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowOwner: <%v>", err))
			return nil
		}
		if rxData.Owner != "" {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(rxData.Owner))

		} else {
			fieldOwner.Clear()
			fieldOwner.Write([]byte(""))
		}

		// отображение содержимого поля Номер.
		fieldNumb, err := gui.View("fieldShowNumber")
		if err != nil || fieldNumb == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowNumber: <%v>", err))
			return nil
		}
		if rxData.Numb != "" {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(rxData.Numb))

		} else {
			fieldNumb.Clear()
			fieldNumb.Write([]byte(""))
		}

		// отображение содержимого поля Валидность.
		fieldValid, err := gui.View("fieldShowValid")
		if err != nil || fieldValid == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowValid: <%v>", err))
			return nil
		}
		if rxData.Valid != "" {
			fieldValid.Clear()
			fieldValid.Write([]byte(rxData.Valid))

		} else {
			fieldValid.Clear()
			fieldValid.Write([]byte(""))
		}

		// отображение содержимого поля Код.
		fieldCode, err := gui.View("fieldShowCode")
		if err != nil || fieldCode == nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowCode: <%v>", err))
			return nil
		}
		if rxData.Code != "" {
			fieldCode.Clear()
			fieldCode.Write([]byte(rxData.Code))

		} else {
			fieldCode.Clear()
			fieldCode.Write([]byte(""))
		}

		return nil
	}

	return nil
}

// Логика перевода фокуса в окне viewBinaryData. Возвращается ошибка.
func doShowPrevElementViewBinaryData(c *handlerUI, gui *gocui.Gui) error {

	// Если режим - локальный.
	if c.conf.Flag.Mode == flags.ModeLocal {
		if c.getStatusPopContainer() != stageActive && c.getStatusPushContainer() != stageActive {

			c.updateStatusPopContainer(stageNotActive)
			c.updateStatusPushContainer(stageNotActive)

			decrIndexFile(c)     // уменьшение значения индекса
			el := fileByIndex(c) // получение записи по индексу

			// отображение содержимого.
			fieldCode, err := gui.View("fieldShowFor")
			if err != nil || fieldCode == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
				return nil
			}
			if el != "" {
				fieldCode.Clear()
				fieldCode.Write([]byte(el))

			} else {
				fieldCode.Clear()
				fieldCode.Write([]byte(""))
			}
		}

		return nil
	}

	// Если режим - удалённый.
	if c.conf.Flag.Mode == flags.ModeRemote {

		if c.getStatusFileTx() != stageActive && c.getStatusFileRx() != stageActive {

			c.updateStatusFileTx(stageNotActive)
			c.updateStatusFileRx(stageNotActive)

			if len(c.data.namesFile) == 0 {
				return nil
			}

			decrIndexFile(c)         // уменьшение значения индекса
			el := fileNameByIndex(c) // получение записи по индексу

			// отображение содержимого.
			fieldCode, err := gui.View("fieldShowFor")
			if err != nil || fieldCode == nil {
				c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка в функции View, при взаимодействии с fieldShowFor: <%v>", err))
				return nil
			}
			if el != "" {
				fieldCode.Clear()
				fieldCode.Write([]byte(el))

			} else {
				fieldCode.Clear()
				fieldCode.Write([]byte(""))
			}
		}

		return nil
	}
	return nil
}

// Логика удаления в окне viewLoginPasswordData. Возвращается ошибка.
func doDeleteElementViewLoginPasswordData(c *handlerUI, gui *gocui.Gui) error {

	c.status.delLoginPaaswordPassed = true
	c.status.delLoginPaaswordSUCCESS = false

	v, err := gui.View("fieldShowFor")
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка получения вида fieldShowFor, при удалении записи логин/пароль: <%v>", err))
		return nil
	}
	textEl := v.Buffer() // Получаем содержимое поля ввода
	textEl = strings.ReplaceAll(textEl, "\n", "")

	textEl, err = encrypt(textEl, c.secret.secretKey)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция encrypt, вернула ошибку: <%v>", err))
		return nil
	}

	// Удаление записи в БД.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := c.conf.DataBase.DelDataLoginPasswordContext(ctx, textEl); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция DelDataLoginPasswordContext, вернула ошибку: <%v>", err))
		return nil
	}

	c.status.delLoginPaaswordSUCCESS = true
	c.conf.PtrLoggerFile.Write(("Debug: данные логин/пароль, успешно удалены"))

	return nil
}

// Логика удаления в окне viewTextData. Возвращается ошибка.
func doDeleteElementViewTextData(c *handlerUI, gui *gocui.Gui) error {

	c.status.delTextPassed = true
	c.status.delTextSUCCESS = false

	v, err := gui.View("fieldShowFor")
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка получения вида fieldShowFor, при удалении записи логин/пароль: <%v>", err))
		return nil
	}
	textEl := v.Buffer() // Получаем содержимое поля ввода
	textEl = strings.ReplaceAll(textEl, "\n", "")

	textEl, err = encrypt(textEl, c.secret.secretKey)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция encrypt, вернула ошибку: <%v>", err))
		return nil
	}

	// Удаление записи в БД.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := c.conf.DataBase.DelTextContext(ctx, textEl); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция DelTextContext, вернула ошибку: <%v>", err))
		return nil
	}

	c.status.delTextSUCCESS = true
	c.conf.PtrLoggerFile.Write(("Debug: данные текста, успешно удалены"))

	return nil
}

// Логика удаления в окне viewBankCardData. Возвращается ошибка.
func doDeleteElementViewBankCardData(c *handlerUI, gui *gocui.Gui) error {

	c.status.delBankCardPassed = true
	c.status.delBankCardSUCCESS = false

	v, err := gui.View("fieldShowFor")
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка получения вида fieldShowFor, при удалении записи логин/пароль: <%v>", err))
		return nil
	}
	textEl := v.Buffer() // Получаем содержимое поля ввода
	textEl = strings.ReplaceAll(textEl, "\n", "")

	textEl, err = encrypt(textEl, c.secret.secretKey)
	if err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция encrypt, вернула ошибку: <%v>", err))
		return nil
	}

	// Удаление записи в БД.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := c.conf.DataBase.DelBankCardContext(ctx, textEl); err != nil {
		c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция DelBankCardContext, вернула ошибку: <%v>", err))
		return nil
	}

	c.status.delBankCardSUCCESS = true
	c.conf.PtrLoggerFile.Write(("Debug: данные карты, успешно удалены"))

	return nil
}

// Логика удаления в окне viewBinaryData. Возвращается ошибка.
func doDeleteElementViewBinaryData(c *handlerUI, gui *gocui.Gui) error {

	if c.getStatusPopContainer() != stageActive && c.getStatusPushContainer() != stageActive {

		c.updateStatusPopContainer(stageNotActive)
		c.updateStatusPushContainer(stageNotActive)

		c.status.delFilePassed = true
		c.status.delFileSUCCESS = false

		// Чтение буфера.
		v, err := gui.View("fieldShowFor")
		if err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: ошибка доступа к элементу fieldShowFor: <%v>", err))
			return nil
		}
		name := v.ViewBuffer()
		name = strings.ReplaceAll(name, "\n", "") // удаление символа

		// Удаление файла.
		if err := c.conf.Container.RemoveFileFromContainer(name, c.secret.secretKey); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: функция RemoveFileFromContainer, вернула ошибку: <%v>", err))
			return nil
		}
		c.status.delFileSUCCESS = true
	}

	return nil
}

// Регистрации пользователя в режиме - Локальный.
func doRegistrationUserLocal(c *handlerUI) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+W")

	c.status.addUserSUCCESS = false // сброс признака успешности регистрации пользователя.
	c.status.addUserPassed = false  // сброс признака, что процедура регистрации быд запущена.
	c.status.addUserRegBusy = false // сброс признака, что в системе уже есть зарегистрированный пользователь.

	userName := c.typed.login
	userPwd1 := c.typed.password1
	userPwd2 := c.typed.password2

	// Проверка корректности введённых пользователем данных
	if err := checkDataRegistration(userName, userPwd1, userPwd2); err != nil {
		return fmt.Errorf("функция checkDataRegistration, вернула ошибку: <%w>", err)
	}

	// Контекст для запроса.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Проверка, что в БД уже есть регистрация пользователя.
	busy, err := c.conf.DataBase.UserExistContext(ctx)
	if err != nil {
		return fmt.Errorf("Error: функция UserExistContext, вернуля ошибку: <%v>", err)
	}

	if busy {
		c.status.addUserRegBusy = true // установка признака, что в системе уже есть зарегистрированный пользоатель.
		return nil
	}

	// Добавление пользователя в БД.
	if err := c.conf.DataBase.AddUserContext(ctx, userName, userPwd1); err != nil {
		return fmt.Errorf("Error: функция AddUserContext, вернуля ошибку: <%v>", err)
	}

	c.status.addUserPassed = true  // установка признака, что процедура регистрации была запущена.
	c.status.addUserSUCCESS = true // установка признака, что пользователь зарегистрировался в системе.

	c.conf.PtrLoggerFile.Write(fmt.Sprintf("Info: выполнена регистрация пользователя с именем: <%s>", userName))

	return nil
}

// Регистрации пользователя в режиме - Удалённый.
func doRegistrationUserRemote(c *handlerUI) error {

	c.conf.PtrLoggerFile.Write("Info: Нажата комбинация Ctrl+W")

	c.status.addUserSUCCESS = false // сброс признака успешности регистрации пользователя.
	c.status.addUserPassed = false  // сброс признака, что процедура регистрации была запущена.
	c.status.addUserRegBusy = false // сброс признака, что в системе уже есть зарегистрированный пользователь.

	c.conf.PtrLoggerFile.Write("Запущен процесс регистрации пользователя. Режим -удалённый.")

	// Проверка аргументов
	if c == nil {
		return NilPtrArgumentC
	}

	// Подключение к серверу.
	client, conn, err := connectSrv(c)
	if err != nil {
		return fmt.Errorf("функция layerConnectSrv, вернула ошибку: <%w>", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			c.conf.PtrLoggerFile.Write(fmt.Sprintf("Error: Ошибка закрытия подключения: <%v>", err))
		}
	}()

	// Создание метаданных с токеном.
	txMD, secretKey, nameToken, err := createTokenForRegistration()
	if err != nil {
		return fmt.Errorf("функция createTokenForRegistration, вернула ошибку: <%w>", err)
	}

	// Подготовка запроса.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = metadata.NewOutgoingContext(ctx, txMD)

	req := &proto.RegistrationRequest{
		UserName:      c.typed.login,
		UserPwd:       c.typed.password1,
		UserPwdRepeat: c.typed.password2,
	}

	var header metadata.MD

	// Запрос регистрации пользователя.
	_, err = client.Registration(ctx, req, grpc.Header(&header))
	if err != nil {
		c.status.addUserRegBusy = true
		return fmt.Errorf("функция client.Registration, вернула ошибку: <%w>", err)
	}

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

	c.status.addUserPassed = true  // установка признака, что процедура регистрации была запущена.
	c.status.addUserSUCCESS = true // установка признака, что пользователь зарегистрировался в системе.

	c.conf.PtrLoggerFile.Write("Регистрация пользователя на удалённом сервере, выполнена. Режим - удалённый.")

	return nil
}

// Создание токена для регистрации пользователя в режиме  - удалённый.
func createTokenForRegistration() (txMD metadata.MD, secretKey, nameToken string, err error) {

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

// Аутентификация в режиме - локальный.
func doAuthenticationUserModeLocal(c *handlerUI) error {

	userName := c.typed.login
	userPwd1 := c.typed.password1

	// Контекст для запроса.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Выполнение запроса.
	ok, err := c.conf.DataBase.AuthenticateUserContext(ctx, userName, userPwd1)
	if err != nil {
		return fmt.Errorf("Error: функция AuthenticateUserContext, вернуля ошибку: <%v>", err)
	}

	// Обработка результата
	if !ok {
		return fmt.Errorf("Info: пользователь <%s>, не прошел аутентификацию.", userName)
	}

	return nil
}

// Аутентификация в режиме - удалённый.
func doAuthenticationUserModeRemote(c *handlerUI) error {

	userName := c.typed.login
	userPwd := c.typed.password1

	// Контекст для запроса.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Выполнение запроса.
	tokenAuth, err := c.conf.Server.AuthenticationContext(ctx, userName, userPwd)
	if err != nil {
		return fmt.Errorf("Error: функция AuthenticateUserContext, вернуля ошибку: <%v>", err)
	}

	c.tokenAuth = tokenAuth

	return nil
}

// Расшифровка принятых данных логин/пароль на запрос по имени
func DecryptRxLoginPasswordByName(rxData server.RxLoginPassword, key [32]byte) (data rxLoginPassword, err error) {

	// Расшифровка For
	data.name, err = decrypt(rxData.For, key)
	if err != nil {
		return rxLoginPassword{}, fmt.Errorf("Ошибка расшифровки For:<%w>", err)
	}

	// Расшифровка Login
	data.login, err = decrypt(rxData.Login, key)
	if err != nil {
		return rxLoginPassword{}, fmt.Errorf("Ошибка расшифровки Login:<%w>", err)
	}

	// Расшифровка Password
	data.password, err = decrypt(rxData.Password, key)
	if err != nil {
		return rxLoginPassword{}, fmt.Errorf("Ошибка расшифровки Password:<%w>", err)
	}

	// Расшифровка CreatedAt
	data.createdAt, err = decrypt(rxData.CreatedAt, key)
	if err != nil {
		return rxLoginPassword{}, fmt.Errorf("Ошибка расшифровки CreatedAt:<%w>", err)
	}

	return data, nil
}
