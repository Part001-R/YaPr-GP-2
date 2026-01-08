package grpc

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// Определение размера файла в байтах. Возвращается размер в байтах и ошибка.
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
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false // Файл не существует
	}
	return err == nil // Файл существует
}
