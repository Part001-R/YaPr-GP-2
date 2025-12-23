package service

import (
	"crypto/tls"
	"fmt"

	"github.com/Part001-R/YaPr-GP-2/server/internal/service/udt"
)

// createTLSConfig создание конфигурацию TLS
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
