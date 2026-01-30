// Тесты пакета.
package server

import (
	"os"
	"testing"
	"time"

	"net"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// Connect
func TestConnect(t *testing.T) {
	t.Run("Пустые ip или port", func(t *testing.T) {
		conn, client, err := connect("", "8080")
		assert.Nil(t, conn)
		assert.Nil(t, client)
		assert.Nil(t, err)

		conn, client, err = connect("127.0.0.1", "")
		assert.Nil(t, conn)
		assert.Nil(t, client)
		assert.Nil(t, err)
	})

	t.Run("Успешное подключение", func(t *testing.T) {
		certPath := "tls/server.crt"

		_, err := os.Stat(certPath)
		require.NoError(t, err, "Сертификат %s должен существовать для теста", certPath)

		// Запуск тестового gRPC-сервера.
		listener, err := net.Listen("tcp", "localhost:50051")
		require.NoError(t, err)

		server := grpc.NewServer()

		go func() {
			if err := server.Serve(listener); err != nil {
				t.Logf("Сервер завершился с ошибкой: %v", err)
			}
		}()

		defer func() {
			server.GracefulStop()
			listener.Close()
		}()

		// Ожидание готовности сервера.
		time.Sleep(100 * time.Millisecond)

		// Подключение к тестовому серверу.
		conn, client, err := connect("localhost", "50051")
		require.NoError(t, err)
		require.NotNil(t, conn)
		require.NotNil(t, client)

		// Закрытие подключенния.
		err = conn.Close()
		assert.NoError(t, err)
	})
}
