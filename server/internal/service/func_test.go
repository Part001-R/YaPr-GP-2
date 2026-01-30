// Тесты пакета.
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тест GetValueOrDefault.
func TestGetValueOrDefault(t *testing.T) {

	t.Run("Корректные данные", func(t *testing.T) {

		data := "Foo"
		assert.Equalf(t, data, GetValueOrDefault(data), "Нет соответствия значения")
	})

	t.Run("Нет значения", func(t *testing.T) {

		data := ""
		wantData := "N/A"
		assert.Equalf(t, wantData, GetValueOrDefault(data), "Нет соответствия значения")
	})
}
