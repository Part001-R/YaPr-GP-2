// Тесты пакета.
package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тест подготовки
func TestPrepare(t *testing.T) {

	_, err := prepare()
	assert.NoErrorf(t, err, "Ошибка подготовки")

	os.Remove("log.txt")
	os.Remove("localStorage.db")
}
