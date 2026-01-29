// Тесты пакета.
package logfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест редактирования пути к файлу.
func TestEditDataPath(t *testing.T) {

	t.Run("Корректные данные", func(t *testing.T) {

		appName := "Foo"
		projectName := "YaPr-GP-2"
		fullPath := "/media/user/DATA1/Golang/YandexPr_Adv/Coding/INCREMENTS/10_gradPro_2/YaPr-GP-2/client/internal/ui/ui.go"

		wantResult := "Foo/internal/ui/ui.go"

		rxData, err := editDataPath(appName, projectName, fullPath)
		require.NoErrorf(t, err, "Неожиданная ошибка")
		assert.Equalf(t, wantResult, rxData, "Нет соответствия данных")
	})

	t.Run("Содержится два названия проекта", func(t *testing.T) {

		appName := "Foo"
		projectName := "YaPr-GP-2"
		fullPath := "/media/user/DATA1/Golang/YandexPr_Adv/Coding/INCREMENTS/10_gradPro_2/YaPr-GP-2/YaPr-GP-2/client/internal/ui/ui.go"

		wantResult := "Foo/internal/ui/ui.go"

		rxData, err := editDataPath(appName, projectName, fullPath)
		require.NoErrorf(t, err, "Неожиданная ошибка")
		assert.Equalf(t, wantResult, rxData, "Нет соответствия данных")
	})

	t.Run("Нет части пути после имени проекта", func(t *testing.T) {

		appName := "Foo"
		projectName := "YaPr-GP-2"
		fullPath := "/media/user/DATA1/Golang/YandexPr_Adv/Coding/INCREMENTS/10_gradPro_2/YaPr-GP-2"

		_, err := editDataPath(appName, projectName, fullPath)
		require.Error(t, err, "Ожидается ошибка")
	})

	t.Run("Нет соответствия имени проекта", func(t *testing.T) {

		appName := "Foo"
		projectName := "YaPr-GP-2"
		fullPath := "/media/user/DATA1/Golang/YandexPr_Adv/Coding/INCREMENTS/10_gradPro_2/YaPr-GP-3"

		_, err := editDataPath(appName, projectName, fullPath)
		require.Error(t, err, "Ожидается ошибка")
	})

	t.Run("Нет части пути до имени проекта", func(t *testing.T) {

		appName := "Foo"
		projectName := "YaPr-GP-2"
		fullPath := "YaPr-GP-2/client/internal/ui/ui.go"

		_, err := editDataPath(appName, projectName, fullPath)
		require.NoErrorf(t, err, "Неожиданная ошибка")
	})

	t.Run("Нет имени приложения", func(t *testing.T) {

		appName := ""
		projectName := "YaPr-GP-2"
		fullPath := "YaPr-GP-2/client/internal/ui/ui.go"

		_, err := editDataPath(appName, projectName, fullPath)
		require.Equalf(t, EmptyDataArgumentAppName, err, "Нет соответствия ошибки")
	})

	t.Run("Нет имени проекта", func(t *testing.T) {

		appName := "Foo"
		projectName := ""
		fullPath := "YaPr-GP-2/client/internal/ui/ui.go"

		_, err := editDataPath(appName, projectName, fullPath)
		require.Equalf(t, EmptyDataArgumentProjectName, err, "Нет соответствия ошибки")
	})

	t.Run("Нет пути", func(t *testing.T) {

		appName := "Foo"
		projectName := "YaPr-GP-2"
		fullPath := ""

		_, err := editDataPath(appName, projectName, fullPath)
		require.Equalf(t, EmptyDataArgumentFullPath, err, "Нет соответствия ошибки")
	})
}
