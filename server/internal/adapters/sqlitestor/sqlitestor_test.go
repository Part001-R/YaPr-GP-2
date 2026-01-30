package sqlitestor

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// --- NewStorage ---
//

func TestNewStorage_SUCCESS(t *testing.T) {

	dbName := "testStorage.db"

	dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
	actions, err := New(dsn)
	require.NoErrorf(t, err, "ошибка конструктора")

	err = actions.Close()
	require.NoErrorf(t, err, "Ошибка закрытия подключения")

	err = os.Remove(dbName)
	assert.NoErrorf(t, err, "Ошибка удаления БД")
}

func TestNewStorage_FAULT(t *testing.T) {

	_, err := New("")
	require.Equalf(t, EmptyDataArgumentDSN, err, "нет соответствия ошибки")
}

//
// --- Close ---
//

func TestClose(t *testing.T) {

	dbName := "testStorage.db"

	dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
	actions, err := New(dsn)
	require.NoErrorf(t, err, "ошибка конструктора")

	err = actions.Close()
	require.NoErrorf(t, err, "Ошибка закрытия подключения")

	err = os.Remove(dbName)
	assert.NoErrorf(t, err, "Ошибка удаления БД")
}

//
// --- AddUserContext ---
//

func TestAddUserContext_SUCCESS(t *testing.T) {

	// Конструктор.
	dbName := "testStorage.db"

	dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
	actions, err := New(dsn)
	require.NoErrorf(t, err, "ошибка конструктора")

	defer func() {
		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err := os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")
	}()

	// Добавление записи.
	userName := "Foo"
	userPwd := "Bar"
	err = actions.AddUserContext(context.Background(), userName, userPwd)
	require.NoErrorf(t, err, "ошибка добавления пользователя")
}

func TestAddUserContext_FAULT(t *testing.T) {

	t.Run("Ошибки в аргуентах", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Данные для теста.
		dataTest := []struct {
			nameTest string
			ctx      context.Context
			name     string
			pwd      string
			wantErr  error
		}{
			{
				nameTest: "Нет контекста",
				ctx:      nil,
				name:     "Foo",
				pwd:      "Bar",
				wantErr:  EmptyDataArgumentCtx,
			},
			{
				nameTest: "Нет имени",
				ctx:      context.Background(),
				name:     "",
				pwd:      "Bar",
				wantErr:  EmptyDataArgumentName,
			},
			{
				nameTest: "Нет пароля",
				ctx:      context.Background(),
				name:     "Foo",
				pwd:      "",
				wantErr:  EmptyDataArgumentPwd,
			},
		}

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nameTest, func(t *testing.T) {

				err = actions.AddUserContext(tt.ctx, tt.name, tt.pwd)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}

	})

	t.Run("Коллизия", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.Errorf(t, err, "ожидается ошибка")
	})

	t.Run("Длительность отработки", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		time.Sleep(11 * time.Millisecond)

		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(ctx, userName, userPwd)
		require.Errorf(t, err, "ожидается ошибка")

	})

	t.Run("Нет указателя на БД и мьютекс", func(t *testing.T) {

		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "Ошибка сброса экземпляра")

		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.Equalf(t, NilPtrDB, err, "нет соответствия ошибки")

	})
}

//
// --- AuthenticateUserContext ---
//

func TestAuthenticateUserContext(t *testing.T) {

	t.Run("Успешная аутентификация", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"

		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		var data DataUser
		data.Field1 = userName
		data.Field2 = userPwd
		isAuth, err := actions.AuthenticateUserContext(context.Background(), data)
		require.NoErrorf(t, err, "Ошибка аутентификации")
		assert.True(t, isAuth, "пользователь не уатентифицирован")
	})

	t.Run("Нет такого пользователя", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		var data DataUser
		data.Field1 = userName + "1"
		data.Field2 = userPwd
		isAuth, err := actions.AuthenticateUserContext(context.Background(), data)
		require.NoErrorf(t, err, "Ошибка аутентификации")
		assert.Falsef(t, isAuth, "пользователь аутентифицирован")
	})

	t.Run("Неверный пароль", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		var data DataUser
		data.Field1 = userName
		data.Field2 = userPwd + "1"
		isAuth, err := actions.AuthenticateUserContext(context.Background(), data)
		require.NoErrorf(t, err, "Ошибка аутентификации")
		assert.Falsef(t, isAuth, "пользователь аутентифицирован")
	})

	t.Run("Нет имени", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		var data DataUser
		data.Field1 = ""
		data.Field2 = userPwd
		_, err = actions.AuthenticateUserContext(context.Background(), data)
		require.Errorf(t, err, "Ошибка аутентификации")
	})

	t.Run("Нет пароля", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		var data DataUser
		data.Field1 = userName
		data.Field2 = ""
		_, err = actions.AuthenticateUserContext(context.Background(), data)
		require.Errorf(t, err, "Ошибка аутентификации")
	})

	t.Run("Нет контекста", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		var data DataUser
		data.Field1 = userName
		data.Field2 = userPwd

		var ctx context.Context
		ctx = nil

		_, err = actions.AuthenticateUserContext(ctx, data)
		require.Errorf(t, err, "Ошибка аутентификации")
	})

	t.Run("Длительность отработки", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		time.Sleep(11 * time.Millisecond)

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"

		err = actions.AddUserContext(ctx, userName, userPwd)
		require.Errorf(t, err, "ожидается ошибка")
	})

	t.Run("Ошибка в конфигурации", func(t *testing.T) {

		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "ошибка сброса")

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"

		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.Equalf(t, NilPtrDB, err, "нет соответствия ошибки")
	})
}

//
// --- UserExistContext ---
//

func TestUserExistContext_SUCCESS(t *testing.T) {

	t.Run("Нет пользователя в системе", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Проверка пользователя.
		isExsist, err := actions.UserExistContext(context.Background())
		require.NoErrorf(t, err, "ошибка добавления пользователя")
		assert.Falsef(t, isExsist, "Есть пользователь")
	})

	t.Run("Есть пользователь в системе", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Проверка пользователя.
		isExsist, err := actions.UserExistContext(context.Background())
		require.NoErrorf(t, err, "ошибка добавления пользователя")
		assert.Truef(t, isExsist, "Нет пользователь")
	})

}

func TestUserExistContext_FAULT(t *testing.T) {

	t.Run("Ошибки в аргументах", func(t *testing.T) {

		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Данные для теста.
		dataTest := []struct {
			nemeTest string
			ctx      context.Context
			wantErr  error
		}{
			{
				nemeTest: "Нет контекста",
				ctx:      nil,
				wantErr:  EmptyDataArgumentCtx,
			},
		}

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.nemeTest, func(t *testing.T) {

				_, err = actions.UserExistContext(tt.ctx)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
	})

	t.Run("Ошибки в конфигурации", func(t *testing.T) {

		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		assert.NoErrorf(t, err, "Ошибка сброса")

		// Проверка пользователя.
		_, err = actions.UserExistContext(context.Background())
		assert.Equalf(t, NilPtrDB, err, "нет соответствия ошибки")
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		time.Sleep(11 * time.Millisecond)

		_, err = actions.UserExistContext(ctx)
		assert.Errorf(t, err, "ожидается ошибка")
	})
}

//
// --- AddDataLoginPasswordContext ---
//

func TestAddDataLoginPasswordContext(t *testing.T) {

	t.Run("Успешное добавление данных", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data DataLoginPassword
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.CreatedAt = "D"
		err = actions.AddDataLoginPasswordContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		// Проверка добавления.
		rxData, err := actions.GetLoginPasswordByNameContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "ошибка чтения записи логин/пароль")
		assert.Equalf(t, data.Field1, rxData.Field1, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData.Field2, "Нет соответствия логина")
		assert.Equalf(t, data.Field3, rxData.Field3, "Нет соответствия пароля")
		assert.Equalf(t, data.CreatedAt, rxData.CreatedAt, "Нет соответствия даты")
	})

	t.Run("Ошибки аргументов", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			ctx      context.Context
			data     DataLoginPassword
			wantErr  error
		}{
			{
				testNAme: "Нет Field1",
				ctx:      context.Background(),
				data: DataLoginPassword{
					Field1:    "",
					Field2:    "B",
					Field3:    "C",
					CreatedAt: "D",
				},
				wantErr: EmptyDataArgumentField1,
			},
			{
				testNAme: "Нет Field2",
				ctx:      context.Background(),
				data: DataLoginPassword{
					Field1:    "A",
					Field2:    "",
					Field3:    "C",
					CreatedAt: "D",
				},
				wantErr: EmptyDataArgumentField2,
			},
			{
				testNAme: "Нет Field3",
				ctx:      context.Background(),
				data: DataLoginPassword{
					Field1:    "A",
					Field2:    "B",
					Field3:    "",
					CreatedAt: "D",
				},
				wantErr: EmptyDataArgumentField3,
			},
			{
				testNAme: "Нет CreatedAt",
				ctx:      context.Background(),
				data: DataLoginPassword{
					Field1:    "A",
					Field2:    "B",
					Field3:    "C",
					CreatedAt: "",
				},
				wantErr: EmptyDataArgumentCreatedAt,
			},
			{
				testNAme: "Нет контекста",
				ctx:      nil,
				data: DataLoginPassword{
					Field1:    "A",
					Field2:    "B",
					Field3:    "C",
					CreatedAt: "D",
				},
				wantErr: EmptyDataArgumentCtx,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.testNAme, func(t *testing.T) {

				err := actions.AddDataLoginPasswordContext(tt.ctx, tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
	})

	t.Run("Коллизия", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data DataLoginPassword
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.CreatedAt = "D"
		err = actions.AddDataLoginPasswordContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		err = actions.AddDataLoginPasswordContext(context.Background(), data)
		require.Errorf(t, err, "ожидается ошибка")
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		time.Sleep(11 * time.Millisecond)

		// Добавление данных.
		var data DataLoginPassword
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.CreatedAt = "D"
		err = actions.AddDataLoginPasswordContext(ctx, data)
		require.Errorf(t, err, "ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "ошибка сброса")

		// Добавление данных.
		var data DataLoginPassword
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.CreatedAt = "D"
		err = actions.AddDataLoginPasswordContext(context.Background(), data)
		require.Equalf(t, NilPtrDB, err, "нет соответствия ошибки")

	})
}

//
// --- ReadTableLoginPasswordContext ---
//

func TestReadTableLoginPasswordContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data DataLoginPassword
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.CreatedAt = "D"
		err = actions.AddDataLoginPasswordContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		// Проверка добавления.
		rxData, err := actions.ReadTableLoginPasswordContext(context.Background())
		require.NoErrorf(t, err, "ошибка чтения записи логин/пароль")
		require.Equalf(t, 1, len(rxData), "Нет соответствия размера массива")

		assert.Equalf(t, data.Field1, rxData[0].Name, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData[0].Login, "Нет соответствия логина")
		assert.Equalf(t, data.Field3, rxData[0].Password, "Нет соответствия пароля")
	})

	t.Run("Длительный запрос", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Проверка добавления.
		_, err = actions.ReadTableLoginPasswordContext(ctx)
		require.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "Ошибка сброса")

		// Проверка добавления.
		_, err = actions.ReadTableLoginPasswordContext(context.Background())
		require.Equalf(t, NilPtrDB, err, "Нет соответствия ошибки")

	})
}

//
// --- DelDataLoginPasswordContext ---
//

func TestDelDataLoginPasswordContext(t *testing.T) {

	t.Run("Успешное удаление данных", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data DataLoginPassword
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.CreatedAt = "D"

		err = actions.AddDataLoginPasswordContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		// Проверка существования записи.
		rxData, err := actions.GetLoginPasswordByNameContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "ошибка удаления записи логин/пароль по имени")
		assert.Equalf(t, data.Field1, rxData.Field1, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData.Field2, "Нет соответствия логина")
		assert.Equalf(t, data.Field3, rxData.Field3, "Нет соответствия пароля")
		assert.Equalf(t, data.CreatedAt, rxData.CreatedAt, "Нет соответствия даты создания")

		// Удаление данных.
		err = actions.DelDataLoginPasswordContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "ошибка удаления записи логин/пароль")

		// Проверка существования записи.
		_, err = actions.GetLoginPasswordByNameContext(context.Background(), data.Field1)
		require.Equalf(t, MissingData, err, "Нет соответствия ошибки")
	})

	t.Run("Ошибки аргументов", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			ctx      context.Context
			data     string
			wantErr  error
		}{
			{
				testNAme: "Нет имени записи",
				ctx:      context.Background(),
				data:     "",
				wantErr:  EmptyDataArgumentField1,
			},
			{
				testNAme: "Нет контекста",
				ctx:      nil,
				data:     "Foo",
				wantErr:  EmptyDataArgumentCtx,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.testNAme, func(t *testing.T) {

				err := actions.DelDataLoginPasswordContext(tt.ctx, tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Добавление данных.
		var data DataLoginPassword
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.CreatedAt = "D"
		err = actions.AddDataLoginPasswordContext(ctx, data)
		require.Errorf(t, err, "ожидается ошибка")
	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "ошибка сброса")

		// Добавление данных.
		var data DataLoginPassword
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.CreatedAt = "D"
		err = actions.AddDataLoginPasswordContext(context.Background(), data)
		require.Errorf(t, err, "ожидается ошибка")
	})
}

//
// --- AddDataTextContext ---
//

func TestAddDataTextContext(t *testing.T) {

	t.Run("Успешное добавление", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		var data DataText
		data.Field1 = "A"
		data.Field2 = "B"
		data.CreatedAt = "C"
		err = actions.AddDataTextContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления текста")

		// Проверка добавления.
		rxData, err := actions.GetTextByNameContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "Ошибка чтения данных")
		assert.Equalf(t, data.Field1, rxData.Name, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData.Text, "Нет соответствия текста")
		assert.Equalf(t, data.CreatedAt, rxData.CreatedAt, "Нет соответствия даты создания")
	})

	t.Run("Ошибки аргументов", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			ctx      context.Context
			data     DataText
			wantErr  error
		}{
			{
				testNAme: "Нет Field1",
				ctx:      context.Background(),
				data: DataText{
					Field1:    "",
					Field2:    "B",
					CreatedAt: "С",
				},
				wantErr: EmptyDataArgumentField1,
			},
			{
				testNAme: "Нет Field2",
				ctx:      context.Background(),
				data: DataText{
					Field1:    "A",
					Field2:    "",
					CreatedAt: "С",
				},
				wantErr: EmptyDataArgumentField2,
			},
			{
				testNAme: "Нет CreatedAt",
				ctx:      context.Background(),
				data: DataText{
					Field1:    "A",
					Field2:    "B",
					CreatedAt: "",
				},
				wantErr: EmptyDataArgumentCreatedAt,
			},
			{
				testNAme: "Нет контекста",
				ctx:      nil,
				data: DataText{
					Field1:    "A",
					Field2:    "B",
					CreatedAt: "C",
				},
				wantErr: EmptyDataArgumentCtx,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.testNAme, func(t *testing.T) {

				err := actions.AddDataTextContext(tt.ctx, tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
	})

	t.Run("Коллизия", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		var data DataText
		data.Field1 = "A"
		data.Field2 = "B"
		data.CreatedAt = "C"
		err = actions.AddDataTextContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления текста")

		err = actions.AddDataTextContext(context.Background(), data)
		require.Errorf(t, err, "ожидается ошибка")
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Добавление записи.
		var data DataText
		data.Field1 = "A"
		data.Field2 = "B"
		data.CreatedAt = "C"
		err = actions.AddDataTextContext(ctx, data)
		require.Errorf(t, err, "ожидается ошибка")

	})

	t.Run("Ошибка в конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "Ошибка сброса")

		// Добавление записи.
		var data DataText
		data.Field1 = "A"
		data.Field2 = "B"
		data.CreatedAt = "C"
		err = actions.AddDataTextContext(context.Background(), data)
		require.Equalf(t, NilPtrDB, err, "нет соответствия ошибки")

	})
}

//
// --- ReadTableTextContext ---
//

func TestReadTableTextContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		var data DataText
		data.Field1 = "A"
		data.Field2 = "B"
		data.CreatedAt = "C"
		err = actions.AddDataTextContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления текста")

		// Проверка добавления.
		rxData, err := actions.ReadTableTextContext(context.Background())
		require.NoErrorf(t, err, "Ошибка чтения данных")
		require.Equalf(t, 1, len(rxData), "нет соответствия размера")

		assert.Equalf(t, data.Field1, rxData[0].Name, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData[0].Text, "Нет соответствия текста")
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Проверка добавления.
		_, err = actions.ReadTableTextContext(ctx)
		require.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		assert.NoErrorf(t, err, "Ошибка сброса")

		// Проверка добавления.
		_, err = actions.ReadTableTextContext(context.Background())
		require.Equalf(t, NilPtrDB, err, "Нет соответствия ошибки")

	})
}

//
// --- DelTextContext ---
//

func TestDelTextContext(t *testing.T) {

	t.Run("Успешное удаление данных", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data DataText
		data.Field1 = "A"
		data.Field2 = "B"
		data.CreatedAt = "C"
		err = actions.AddDataTextContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		// Чтение данных.
		rxData, err := actions.GetNamesTextContext(context.Background())
		require.NoErrorf(t, err, "Ошибка получения данных по имени")
		assert.Equalf(t, data.Field1, rxData[0], "Нет соответствия имени записи")

		// Удаление записи.
		err = actions.DelTextContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "Ошибка удаления записи")

		// Чтение данных.
		rxData, err = actions.GetNamesTextContext(context.Background())
		require.NoErrorf(t, err, "Ошибка получения данных по имени")
		assert.Equalf(t, 0, len(rxData), "Нет соответствия размера ответа")
	})

	t.Run("Ошибки аргументов", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			ctx      context.Context
			data     string
			wantErr  error
		}{
			{
				testNAme: "Нет имени записи",
				ctx:      context.Background(),
				data:     "",
				wantErr:  EmptyDataArgumentField1,
			},
			{
				testNAme: "Нет контекста",
				ctx:      nil,
				data:     "Foo",
				wantErr:  EmptyDataArgumentCtx,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.testNAme, func(t *testing.T) {

				err := actions.DelTextContext(tt.ctx, tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		err = actions.DelTextContext(ctx, "Foo")
		require.Errorf(t, err, "ожидается ошибка")
	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "Ошибка сброса")

		err = actions.DelTextContext(context.Background(), "Foo")
		require.Errorf(t, err, "ожидается ошибка")
	})
}

//
// --- AddDataBankCardContext ---
//

func TestAddDataBankCardContext(t *testing.T) {

	t.Run("Успешное добавление", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		var data DataBankCard
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.Field4 = "D"
		data.Field5 = "E"
		data.CreatedAt = "F"
		err = actions.AddDataBankCardContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления текста")

		// Проверка добавления.
		rxData, err := actions.GetBankCardByNameContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "ошибка добавления текста")
		assert.Equalf(t, data.Field1, rxData.Name, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData.Owner, "Нет соответствия имени владельца")
		assert.Equalf(t, data.Field3, rxData.Numb, "Нет соответствия номера карты")
		assert.Equalf(t, data.Field4, rxData.Valid, "Нет соответствия даты валидности")
		assert.Equalf(t, data.Field5, rxData.Code, "Нет соответствия кода")
	})

	t.Run("Ошибки аргументов", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			ctx      context.Context
			data     DataBankCard
			wantErr  error
		}{
			{
				testNAme: "Нет Field1",
				ctx:      context.Background(),
				data: DataBankCard{
					Field1:    "",
					Field2:    "B",
					Field3:    "C",
					Field4:    "D",
					Field5:    "E",
					CreatedAt: "F",
				},
				wantErr: EmptyDataArgumentField1,
			},
			{
				testNAme: "Нет Field2",
				ctx:      context.Background(),
				data: DataBankCard{
					Field1:    "A",
					Field2:    "",
					Field3:    "C",
					Field4:    "D",
					Field5:    "E",
					CreatedAt: "F",
				},
				wantErr: EmptyDataArgumentField2,
			},
			{
				testNAme: "Нет Field3",
				ctx:      context.Background(),
				data: DataBankCard{
					Field1:    "A",
					Field2:    "B",
					Field3:    "",
					Field4:    "D",
					Field5:    "E",
					CreatedAt: "F",
				},
				wantErr: EmptyDataArgumentField3,
			},
			{
				testNAme: "Нет Field4",
				ctx:      context.Background(),
				data: DataBankCard{
					Field1:    "A",
					Field2:    "B",
					Field3:    "C",
					Field4:    "",
					Field5:    "E",
					CreatedAt: "F",
				},
				wantErr: EmptyDataArgumentField4,
			},
			{
				testNAme: "Нет Field5",
				ctx:      context.Background(),
				data: DataBankCard{
					Field1:    "A",
					Field2:    "B",
					Field3:    "C",
					Field4:    "D",
					Field5:    "",
					CreatedAt: "F",
				},
				wantErr: EmptyDataArgumentField5,
			},
			{
				testNAme: "Нет CreatedAt",
				ctx:      context.Background(),
				data: DataBankCard{
					Field1:    "A",
					Field2:    "B",
					Field3:    "C",
					Field4:    "D",
					Field5:    "E",
					CreatedAt: "",
				},
				wantErr: EmptyDataArgumentCreatedAt,
			},
			{
				testNAme: "Нет контекста",
				ctx:      nil,
				data: DataBankCard{
					Field1:    "A",
					Field2:    "B",
					Field3:    "C",
					Field4:    "D",
					Field5:    "E",
					CreatedAt: "F",
				},
				wantErr: EmptyDataArgumentCtx,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.testNAme, func(t *testing.T) {

				err := actions.AddDataBankCardContext(tt.ctx, tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
	})

	t.Run("Коллизия", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		var data DataBankCard
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.Field4 = "D"
		data.Field5 = "E"
		data.CreatedAt = "F"
		err = actions.AddDataBankCardContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления текста")

		err = actions.AddDataBankCardContext(context.Background(), data)
		require.Errorf(t, err, "ожидается ошибка")

	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Добавление записи.
		var data DataBankCard
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.Field4 = "D"
		data.Field5 = "E"
		data.CreatedAt = "F"
		err = actions.AddDataBankCardContext(ctx, data)
		require.Errorf(t, err, "ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "Ошибка сброса")

		// Добавление записи.
		var data DataBankCard
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.Field4 = "D"
		data.Field5 = "E"
		data.CreatedAt = "F"
		err = actions.AddDataBankCardContext(context.Background(), data)
		require.Errorf(t, err, "ожидается ошибка")

	})
}

//
// --- ReadTableBankCardContext ---
//

func TestReadTableBankCardContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		var data DataBankCard
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.Field4 = "D"
		data.Field5 = "E"
		data.CreatedAt = "F"
		err = actions.AddDataBankCardContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления текста")

		// Проверка добавления.
		rxData, err := actions.ReadTableBankCardContext(context.Background())
		require.NoErrorf(t, err, "ошибка добавления текста")
		require.Equalf(t, 1, len(rxData), "нет соответствия размера")

		assert.Equalf(t, data.Field1, rxData[0].Field1, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData[0].Field2, "Нет соответствия имени владельца")
		assert.Equalf(t, data.Field3, rxData[0].Field3, "Нет соответствия номера карты")
		assert.Equalf(t, data.Field4, rxData[0].Field4, "Нет соответствия даты валидности")
		assert.Equalf(t, data.Field5, rxData[0].Field5, "Нет соответствия кода")
	})

	t.Run("Длительный запрос", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Проверка добавления.
		_, err = actions.ReadTableBankCardContext(ctx)
		require.Errorf(t, err, "ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "Ошибка сброса")

		// Проверка добавления.
		_, err = actions.ReadTableBankCardContext(context.Background())
		require.Equalf(t, NilPtrDB, err, "Нет соответствия ошибки")

	})
}

//
// --- DelBankCardContext ---
//

func TestDelBankCardContext(t *testing.T) {

	t.Run("Успешное удаление данных", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data DataBankCard
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.Field4 = "D"
		data.Field5 = "E"
		data.CreatedAt = "F"
		err = actions.AddDataBankCardContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления записи банковской карты")

		// Удаление данных.
		err = actions.DelBankCardContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "ошибка удаления записи банковской карты")

		// Проверка отсутствия.
		_, err = actions.GetBankCardByNameContext(context.Background(), data.Field1)
		assert.Equalf(t, MissingData, err, "Нет соответствия ошибки")
	})

	t.Run("Ошибки аргументов", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			ctx      context.Context
			data     string
			wantErr  error
		}{
			{
				testNAme: "Нет имени записи",
				ctx:      context.Background(),
				data:     "",
				wantErr:  EmptyDataArgumentField1,
			},
			{
				testNAme: "Нет контекста",
				ctx:      nil,
				data:     "Foo",
				wantErr:  EmptyDataArgumentCtx,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Тесты.
		for _, tt := range dataTest {
			t.Run(tt.testNAme, func(t *testing.T) {

				err := actions.DelBankCardContext(tt.ctx, tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data DataBankCard
		data.Field1 = "A"
		data.Field2 = "B"
		data.Field3 = "C"
		data.Field4 = "D"
		data.Field5 = "E"
		data.CreatedAt = "F"
		err = actions.AddDataBankCardContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления записи банковской карты")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Удаление данных.
		err = actions.DelBankCardContext(ctx, data.Field1)
		require.Errorf(t, err, "ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "Ошибка сброса")

		// Удаление данных.
		err = actions.DelBankCardContext(context.Background(), "Foo")
		require.Errorf(t, err, "ожидается ошибка")

	})
}

//
// --- GetNamesLoginPasswordContext ---
//

func TestGetNamesLoginPasswordContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data1 DataLoginPassword
		data1.Field1 = "A1"
		data1.Field2 = "B1"
		data1.Field3 = "C1"
		data1.CreatedAt = "D1"
		err = actions.AddDataLoginPasswordContext(context.Background(), data1)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		var data2 DataLoginPassword
		data2.Field1 = "A2"
		data2.Field2 = "B2"
		data2.Field3 = "C2"
		data2.CreatedAt = "D2"
		err = actions.AddDataLoginPasswordContext(context.Background(), data2)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		// Получение имён записей.
		rxData, err := actions.GetNamesLoginPasswordContext(context.Background())
		require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

		// Проверка результата.
		assert.Equalf(t, 2, len(rxData), "Нет соответствия размера записей")
		assert.Equalf(t, data1.Field1, rxData[0], "Нет соответствия имени первой записи")
		assert.Equalf(t, data2.Field1, rxData[1], "Нет соответствия имени второй записи")
	})

	t.Run("Нет контекста", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		var ctx context.Context
		ctx = nil

		// Запрос.
		_, err = actions.GetNamesLoginPasswordContext(ctx)
		require.Equalf(t, EmptyDataArgumentCtx, err, "Нет соответствия ошибки")

	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data1 DataLoginPassword
		data1.Field1 = "A1"
		data1.Field2 = "B1"
		data1.Field3 = "C1"
		data1.CreatedAt = "D1"
		err = actions.AddDataLoginPasswordContext(context.Background(), data1)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Получение имён записей.
		_, err = actions.GetNamesLoginPasswordContext(ctx)
		require.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		assert.NoErrorf(t, err, "Ошибка сброса")

		// Получение имён записей.
		_, err = actions.GetNamesLoginPasswordContext(context.Background())
		require.Equalf(t, NilPtrDB, err, "нет соответствия ошибки")

	})
}

//
// --- GetLoginPasswordByNameContext ---
//

func TestGetLoginPasswordByNameContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data1 DataLoginPassword
		data1.Field1 = "A1"
		data1.Field2 = "B1"
		data1.Field3 = "C1"
		data1.CreatedAt = "D1"
		err = actions.AddDataLoginPasswordContext(context.Background(), data1)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		// Получение данных по имени записи.
		rxData, err := actions.GetLoginPasswordByNameContext(context.Background(), data1.Field1)
		require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

		// Проверка результата.
		assert.Equalf(t, data1.Field1, rxData.Field1, "Нет соответствия имени записи")
		assert.Equalf(t, data1.Field2, rxData.Field2, "Нет соответствия логина записи")
		assert.Equalf(t, data1.Field3, rxData.Field3, "Нет соответствия пароля записи")
		assert.Equalf(t, data1.CreatedAt, rxData.CreatedAt, "Нет соответствия даты создания записи")
	})

	t.Run("нет имени записи", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Получение данных по имени записи.
		_, err = actions.GetLoginPasswordByNameContext(context.Background(), "")
		require.Equalf(t, EmptyDataArgumentName, err, "Нет соответствия ошибки")
	})

	t.Run("нет контекста", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		var ctx context.Context
		ctx = nil

		// Получение данных по имени записи.
		_, err = actions.GetLoginPasswordByNameContext(ctx, "Foo")
		require.Equalf(t, EmptyDataArgumentCtx, err, "Нет соответствия ошибки")
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Получение данных по имени записи.
		_, err = actions.GetLoginPasswordByNameContext(ctx, "Foo")
		require.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		require.NoErrorf(t, err, "Ошибка сброса")

		// Получение данных по имени записи.
		_, err = actions.GetLoginPasswordByNameContext(context.Background(), "Foo")
		require.Equalf(t, NilPtrDB, err, "Нет соответствия ошибки")

	})
}

//
// --- GetNamesTextContext ---
//

func TestGetNamesTextContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data1 DataText
		data1.Field1 = "A1"
		data1.Field2 = "B1"
		data1.CreatedAt = "C1"
		err = actions.AddDataTextContext(context.Background(), data1)
		require.NoErrorf(t, err, "ошибка добавления записи текста")

		var data2 DataText
		data2.Field1 = "A2"
		data2.Field2 = "B2"
		data2.CreatedAt = "C2"
		err = actions.AddDataTextContext(context.Background(), data2)
		require.NoErrorf(t, err, "ошибка добавления записи текста")

		// Получение имён записей.
		rxData, err := actions.GetNamesTextContext(context.Background())
		require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

		// Проверка результата.
		assert.Equalf(t, 2, len(rxData), "Нет соответствия размера записей")
		assert.Equalf(t, data1.Field1, rxData[0], "Нет соответствия имени первой записи")
		assert.Equalf(t, data2.Field1, rxData[1], "Нет соответствия имени второй записи")
	})

	t.Run("Нет контекста", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		var ctx context.Context
		ctx = nil

		// Получение имён записей.
		_, err = actions.GetNamesTextContext(ctx)
		require.Equalf(t, EmptyDataArgumentCtx, err, "Нет соответствия ошибки")
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Получение имён записей.
		_, err = actions.GetNamesTextContext(ctx)
		require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		assert.NoErrorf(t, err, "Ошибка сброса")

		// Получение имён записей.
		_, err = actions.GetNamesTextContext(context.Background())
		require.Equalf(t, NilPtrDB, err, "Нет соответствия ошибки")

	})
}

//
// --- GetTextByNameContext ---
//

func TestGetTextByNameContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data DataText
		data.Field1 = "A1"
		data.Field2 = "B1"
		data.CreatedAt = "C1"
		err = actions.AddDataTextContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления записи логин/пароль")

		// Получение данных по имени записи.
		rxData, err := actions.GetTextByNameContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

		// Проверка результата.
		assert.Equalf(t, data.Field1, rxData.Name, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData.Text, "Нет соответствия текста записи")
		assert.Equalf(t, data.CreatedAt, rxData.CreatedAt, "Нет соответствия даты создания записи")
	})

	t.Run("Нет имени", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Получение данных по имени записи.
		_, err = actions.GetTextByNameContext(context.Background(), "")
		require.Equalf(t, EmptyDataArgumentName, err, "Нет соответствия ошибки")
	})

	t.Run("Нет контекста", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		var ctx context.Context
		ctx = nil

		// Получение данных по имени записи.
		_, err = actions.GetTextByNameContext(ctx, "Foo")
		require.Equalf(t, EmptyDataArgumentCtx, err, "Нет соответствия ошибки")
	})

	t.Run("Длительная отработка", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Получение данных по имени записи.
		_, err = actions.GetTextByNameContext(ctx, "Foo")
		require.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		assert.NoErrorf(t, err, "Ошибка сброса")

		// Получение данных по имени записи.
		_, err = actions.GetTextByNameContext(context.Background(), "Foo")
		require.Equalf(t, NilPtrDB, err, "Нет соответствия ошибки")

	})
}

//
// --- GetNamesBankCardContext ---
//

func TestGetNamesBankCardContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data1 DataBankCard
		data1.Field1 = "A1"
		data1.Field2 = "B1"
		data1.Field3 = "C1"
		data1.Field4 = "D1"
		data1.Field5 = "E1"
		data1.CreatedAt = "F1"
		err = actions.AddDataBankCardContext(context.Background(), data1)
		require.NoErrorf(t, err, "ошибка добавления записи банковской карты")

		var data2 DataBankCard
		data2.Field1 = "A2"
		data2.Field2 = "B2"
		data2.Field3 = "C2"
		data2.Field4 = "D2"
		data2.Field5 = "E2"
		data2.CreatedAt = "F2"
		err = actions.AddDataBankCardContext(context.Background(), data2)
		require.NoErrorf(t, err, "ошибка добавления записи банковской карты")

		// Получение имён записей.
		rxData, err := actions.GetNamesBankCardContext(context.Background())
		require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

		// Проверка результата.
		assert.Equalf(t, 2, len(rxData), "Нет соответствия размера записей")
		assert.Equalf(t, data1.Field1, rxData[0], "Нет соответствия имени первой записи")
		assert.Equalf(t, data2.Field1, rxData[1], "Нет соответствия имени второй записи")
	})

	t.Run("Нет контекста", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		var ctx context.Context
		ctx = nil

		// Получение имён записей.
		_, err = actions.GetNamesBankCardContext(ctx)
		require.Equalf(t, EmptyDataArgumentCtx, err, "Нет соответствия ошибки")
	})

	t.Run("Длительный запрос", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Получение имён записей.
		_, err = actions.GetNamesBankCardContext(ctx)
		require.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		assert.NoErrorf(t, err, "Ошибка сброса")

		// Получение имён записей.
		_, err = actions.GetNamesBankCardContext(context.Background())
		require.Equalf(t, NilPtrDB, err, "Нет соответствия ошибки")

	})
}

//
// --- GetBankCardByNameContext ---
//

func TestGetBankCardByNameContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление данных.
		var data DataBankCard
		data.Field1 = "A1"
		data.Field2 = "B1"
		data.Field3 = "C1"
		data.Field4 = "D1"
		data.Field5 = "E1"
		data.CreatedAt = "F1"
		err = actions.AddDataBankCardContext(context.Background(), data)
		require.NoErrorf(t, err, "ошибка добавления записи банковской карты")

		// Получение данных по имени записи.
		rxData, err := actions.GetBankCardByNameContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

		// Проверка результата.
		assert.Equalf(t, data.Field1, rxData.Name, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData.Owner, "Нет соответствия владельца карты")
		assert.Equalf(t, data.Field3, rxData.Numb, "Нет соответствия номера карты")
		assert.Equalf(t, data.Field4, rxData.Valid, "Нет соответствия даты валидности карты")
		assert.Equalf(t, data.Field5, rxData.Code, "Нет соответствия кода карты")
		assert.Equalf(t, data.CreatedAt, rxData.CreatedAt, "Нет соответствия даты создания записи")
	})

	t.Run("Нет имени", func(t *testing.T) {

		// Конструктор.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Получение данных по имени записи.
		_, err = actions.GetBankCardByNameContext(context.Background(), "")
		require.Equalf(t, EmptyDataArgumentName, err, "Нет соответствия ошибки")
	})

	t.Run("Нет контекста", func(t *testing.T) {

		// Конструктор.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		var ctx context.Context
		ctx = nil

		// Получение данных по имени записи.
		_, err = actions.GetBankCardByNameContext(ctx, "Foo")
		require.Equalf(t, EmptyDataArgumentCtx, err, "Нет соответствия ошибки")
	})

	t.Run("Длительный запрос", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		time.Sleep(6 * time.Millisecond)

		// Получение данных по имени записи.
		_, err = actions.GetBankCardByNameContext(ctx, "Foo")
		require.Errorf(t, err, "Ожидается ошибка")

	})

	t.Run("Ошибка конфигурации", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := New(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		err = actions.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")

		err = os.Remove(dbName)
		assert.NoErrorf(t, err, "Ошибка удаления БД")

		err = actions.ResetForTest()
		assert.NoErrorf(t, err, "Ошибка сброса")

		// Получение данных по имени записи.
		_, err = actions.GetBankCardByNameContext(context.Background(), "Foo")
		require.Equalf(t, NilPtrDB, err, "Ожидается ошибка")

	})
}
