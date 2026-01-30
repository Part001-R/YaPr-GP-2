package service

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Run
func TestRun(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		conf, err := prepare()
		require.NoErrorf(t, err, "Ошибка подготовки")
		conf.Flag.Port = "50101" // Для исключения коллизии в тестах

		err = Run() // Запуск сервиса.
		assert.NoErrorf(t, err, "Ошибка сервиса:<%v>", err)
	}()

	time.Sleep(5 * time.Second) // выдержка на запуск.

	//  Ctrl+C
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("не удалось найти процесс: %v", err)
	}

	// Сигнал Ctrl+C
	err = proc.Signal(os.Interrupt)
	if err != nil {
		t.Fatalf("не удалось отправить сигнал SIGINT: %v", err)
	}

	wg.Wait()

	// Очистка следов.
	os.RemoveAll("backup")
	os.Remove("remote.db")
	os.RemoveAll("files")
}
