// Вспомогательные функции пакета.
package container

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/nacl/secretbox"
)

// Шифрование данных. Возвращаются зашифрованные данные и ошибка.
//
// Параметры:
//
//	data - данные для шифрования.
//	key - ключ шифрования.
func encrypt(data []byte, key [32]byte) ([]byte, error) {

	var nonce [24]byte

	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	encrypted := secretbox.Seal(nonce[:], data, &nonce, &key)
	return encrypted, nil
}

// Дешифрование данных. Возвращаются дешифрованные данные и ошибка.
//
// Параметры:
//
//	encrypted - зашифрованные данные.
//	key - ключ шифрования.
func decrypt(encrypted []byte, key [32]byte) ([]byte, error) {

	var nonce [24]byte

	copy(nonce[:], encrypted[:nonceSize])
	decrypted, ok := secretbox.Open(nil, encrypted[nonceSize:], &nonce, &key)
	if !ok {
		return nil, fmt.Errorf("ошибка дешифрования")
	}
	return decrypted, nil
}

// Чтение контейнера из файла. Возвращается указатель на контейнер и ошибка.
//
// Параметры:
//
//	key - ключ шифрования.
func readContainer(key [32]byte, c *Container) (*Container, error) {

	data, err := os.ReadFile(c.Name)
	if err != nil {
		if os.IsNotExist(err) {
			return &Container{}, nil // пустой контейнер
		}
		return nil, fmt.Errorf("Функция os.ReadFile, вернула ошибку:<%w>, при чтении файла:<%s>", err, c.Name)
	}

	var container Container
	err = json.Unmarshal(data, &container)
	if err != nil {
		return nil, fmt.Errorf("Функция json.Unmarshal, вернула ошибку:<%w>, при обработке данных файла:<%s>", err, c.Name)
	}

	// Дешифрация файлов.
	for i, file := range container.Files {
		decrypted, err := decrypt(file.Content, key)
		if err != nil {
			return nil, fmt.Errorf("не удалось дешифровать файл %s: %v", file.Name, err)
		}
		container.Files[i].Content = decrypted
	}

	return &container, nil
}

// Запись контейнера в файл. Возвращается ошибка.
//
// Параметры:
//
//	container - указатель на контейнер.
//	key - ключ шифрования.
func writeContainer(container *Container, key [32]byte, c *Container) error {

	// Шифрование каждого файла перед сохранением.
	encryptedFiles := make([]FileEntry, len(container.Files))

	for i, file := range container.Files {
		encrypted, err := encrypt(file.Content, key)
		if err != nil {
			return err
		}
		encryptedFiles[i] = FileEntry{
			Name:    file.Name,
			Content: encrypted,
		}
	}

	containerToSave := Container{Files: encryptedFiles}
	data, err := json.Marshal(containerToSave)
	if err != nil {
		return err
	}

	return os.WriteFile(c.Name, data, 0644)
}
