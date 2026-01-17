// Вспомогательные функции пакета.
package service

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/Part001-R/YaPr-GP-2/server/internal/service/udt"
)

// createTLSConfig создание конфигурацию TLS. Возвращается конфигурация и ошибка.
//
// Параметры:
//
// c - указатель на конфигурацию сервиса.
func createTLSConfig(c *udt.Configuration) (*tls.Config, error) {

	cert, err := tls.LoadX509KeyPair(c.TLS.Public, c.TLS.Privae)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки сертификатов: %w", err)
	}

	// Создание новой конфигурации TLS
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

// Создание дочерней директории. Возвращается ошибка.
//
// Параметры:
//
//	subdirName - имя дирекории.
func createSubdirectory(subdirName string) error {

	if _, err := os.Stat(subdirName); os.IsNotExist(err) {

		err := os.Mkdir(subdirName, 0755)
		if err != nil {
			return fmt.Errorf("ошибка при создании директории: %v", err)
		}
		fmt.Printf("Поддиректория <%s> успешно создана.\n", subdirName)
	}
	return nil
}
