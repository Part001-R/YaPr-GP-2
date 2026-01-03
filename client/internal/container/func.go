package container

import (
	"crypto/rand"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/nacl/secretbox"
)

// Шифрование данных
func encrypt(data []byte, key [32]byte) ([]byte, error) {

	var nonce [24]byte

	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	encrypted := secretbox.Seal(nonce[:], data, &nonce, &key)
	return encrypted, nil
}

// Дешифрование данных
func decrypt(encrypted []byte, key [32]byte) ([]byte, error) {

	var nonce [24]byte

	copy(nonce[:], encrypted[:nonceSize])
	decrypted, ok := secretbox.Open(nil, encrypted[nonceSize:], &nonce, &key)
	if !ok {
		return nil, fmt.Errorf("ошибка дешифрования")
	}
	return decrypted, nil
}

// Выделение имени файла и его тип из полного пути. Возвращается имя файла и его тип.
//
// Параметры:
//
//	fullPath - полный путь к файлу.
func getFileNameAndExtension(fullPath string) string {

	fileName := filepath.Base(fullPath)

	fileName = strings.ReplaceAll(fileName, "`", "")

	fileType := filepath.Ext(fileName)

	if strings.Contains(fileName, fileType) {
		return fileName
	}

	return fileName + fileType
}
