package domain

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// --- TestNewStorage ---
//

func TestNewStorage(t *testing.T) {

	dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
	actions, err := NewStorage(dsn)
	require.NoErrorf(t, err, "ошибка конструктора")

	defer func() {
		err := os.Remove("testStorage.db")
		assert.NoErrorf(t, err, "Ошибка удаления БД")
	}()

	err = actions.Close()
	require.NoErrorf(t, err, "ошибка закрытия подключения")

}

//
// --- AddUserContext ---
//

func TestAddUserContext(t *testing.T) {

	// Конструктор.
	dbName := "testStorage.db"

	dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
	actions, err := NewStorage(dsn)
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

//
// --- AuthenticateUserContext ---
//

func TestAuthenticateUserContext(t *testing.T) {

	t.Run("Успешная аутентификация", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
		var txData DataUser
		txData.Field1 = userName
		txData.Field2 = userPwd
		isAuth, err := actions.AuthenticateUserContext(context.Background(), txData)
		require.NoErrorf(t, err, "Ошибка аутентификации")
		assert.True(t, isAuth, "пользователь не уатентифицирован")
	})

	t.Run("Нет такого пользователя", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
		data.Field1 = userName + "A"
		data.Field2 = userPwd
		isAuth, err := actions.AuthenticateUserContext(context.Background(), data)
		require.NoErrorf(t, err, "Ошибка аутентификации")
		assert.Falsef(t, isAuth, "пользователь аутентифицирован")
	})

	t.Run("Неверный пароль", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
		data.Field2 = userPwd + "A"
		isAuth, err := actions.AuthenticateUserContext(context.Background(), data)
		require.NoErrorf(t, err, "Ошибка аутентификации")
		assert.Falsef(t, isAuth, "пользователь аутентифицирован")
	})

	t.Run("Нет имени", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
		actions, err := NewStorage(dsn)
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
}

//
// --- UserExistContext ---
//

func TestUserExistContext(t *testing.T) {

	t.Run("Нет пользователя в системе", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
		actions, err := NewStorage(dsn)
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

//
// --- AddDataLoginPasswordContext ---
//

func TestAddDataLoginPasswordContext(t *testing.T) {

	t.Run("Успешное добавление данных", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
		assert.Equalf(t, data.Field1, rxData.Name, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData.Login, "Нет соответствия логина")
		assert.Equalf(t, data.Field3, rxData.Password, "Нет соответствия пароля")
		assert.Equalf(t, data.CreatedAt, rxData.CreatedAt, "Нет соответствия даты")
	})

	t.Run("Ошибки", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			data     DataLoginPassword
			wantErr  error
		}{
			{
				testNAme: "Нет Field1",
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
				data: DataLoginPassword{
					Field1:    "A",
					Field2:    "B",
					Field3:    "C",
					CreatedAt: "",
				},
				wantErr: EmptyDataArgumentCreatedAt,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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

				err := actions.AddDataLoginPasswordContext(context.Background(), tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
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
		actions, err := NewStorage(dsn)
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
		assert.Equalf(t, data.Field1, rxData.Name, "Нет соответствия имени записи")
		assert.Equalf(t, data.Field2, rxData.Login, "Нет соответствия логина")
		assert.Equalf(t, data.Field3, rxData.Password, "Нет соответствия пароля")
		assert.Equalf(t, data.CreatedAt, rxData.CreatedAt, "Нет соответствия даты создания")

		// Удаление данных.
		err = actions.DelDataLoginPasswordContext(context.Background(), data.Field1)
		require.NoErrorf(t, err, "ошибка удаления записи логин/пароль")

		// Проверка существования записи.
		_, err = actions.GetLoginPasswordByNameContext(context.Background(), data.Field1)
		require.Equalf(t, "SQlite. Функция GetLoginPasswordByNameContext, вернула ошибку: <отсутствуют данные>", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Ошибки", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			data     string
			wantErr  error
		}{
			{
				testNAme: "Нет имени записи",
				data:     "",
				wantErr:  EmptyDataArgumentField1,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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

				err := actions.DelDataLoginPasswordContext(context.Background(), tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
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
		actions, err := NewStorage(dsn)
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

	t.Run("Ошибки", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			data     DataText
			wantErr  error
		}{
			{
				testNAme: "Нет Field1",
				data: DataText{
					Field1:    "",
					Field2:    "B",
					CreatedAt: "С",
				},
				wantErr: EmptyDataArgumentField1,
			},
			{
				testNAme: "Нет Field2",
				data: DataText{
					Field1:    "A",
					Field2:    "",
					CreatedAt: "С",
				},
				wantErr: EmptyDataArgumentField2,
			},
			{
				testNAme: "Нет CreatedAt",
				data: DataText{
					Field1:    "A",
					Field2:    "B",
					CreatedAt: "",
				},
				wantErr: EmptyDataArgumentCreatedAt,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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

				err := actions.AddDataTextContext(context.Background(), tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
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
		actions, err := NewStorage(dsn)
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

	t.Run("Ошибки", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			data     string
			wantErr  error
		}{
			{
				testNAme: "Нет имени записи",
				data:     "",
				wantErr:  EmptyDataArgumentField1,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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

				err := actions.DelTextContext(context.Background(), tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
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
		actions, err := NewStorage(dsn)
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

	t.Run("Ошибки", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			data     DataBankCard
			wantErr  error
		}{
			{
				testNAme: "Нет Field1",
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
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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

				err := actions.AddDataBankCardContext(context.Background(), tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
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
		actions, err := NewStorage(dsn)
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
		assert.Equalf(t, "SQlite. Функция GetBankCardByNameContext, вернула ошибку: <отсутствуют данные>", err.Error(), "Нет соответствия ошибки")
	})

	t.Run("Ошибки", func(t *testing.T) {

		dataTest := []struct {
			testNAme string
			data     string
			wantErr  error
		}{
			{
				testNAme: "Нет имени записи",
				data:     "",
				wantErr:  EmptyDataArgumentField1,
			},
		}

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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

				err := actions.DelBankCardContext(context.Background(), tt.data)
				require.Equalf(t, tt.wantErr, err, "нет соответствия ошибки")
			})
		}
	})
}

//
// --- ReadNamesTableLoginPasswordContext ---
//

func TestReadNamesTableLoginPasswordContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
}

//
// --- ReadLoginPassworByNameContext ---
//

func TestReadLoginPassworByNameContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
		assert.Equalf(t, data1.Field1, rxData.Name, "Нет соответствия имени записи")
		assert.Equalf(t, data1.Field2, rxData.Login, "Нет соответствия логина записи")
		assert.Equalf(t, data1.Field3, rxData.Password, "Нет соответствия пароля записи")
		assert.Equalf(t, data1.CreatedAt, rxData.CreatedAt, "Нет соответствия даты создания записи")
	})

	t.Run("Ошибки", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Получение данных по имени записи.
		_, err = actions.GetLoginPasswordByNameContext(context.Background(), "")
		require.Equalf(t, "SQlite. Функция GetLoginPasswordByNameContext, вернула ошибку: <нет содержимого в аргументе name>", err.Error(), "Нет соответствия ошибки")
	})
}

//
// --- ReadNamesTableTextContext ---
//

func TestReadNamesTableTextContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
}

//
// --- ReadTextByNameContext ---
//

func TestReadTextByNameContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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

	t.Run("Ошибки", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove(dbName)
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Получение данных по имени записи.
		_, err = actions.GetTextByNameContext(context.Background(), "")
		require.Equalf(t, "SQlite. Функция GetTextByNameContext, вернула ошибку: <нет содержимого в аргументе name>", err.Error(), "Нет соответствия ошибки")
	})
}

//
// --- ReadNamesTableBankCardContext ---
//

func TestReadNamesTableBankCardContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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
}

//
// --- ReadBankCardByNameContext ---
//

func TestReadBankCardByNameContext(t *testing.T) {

	t.Run("Успешное чтение", func(t *testing.T) {

		// Конструктор.
		dbName := "testStorage.db"

		dsn := "file:" + dbName + "?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
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

	t.Run("Ошибки", func(t *testing.T) {

		// Конструктор.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Получение данных по имени записи.
		_, err = actions.GetBankCardByNameContext(context.Background(), "")
		require.Equalf(t, "SQlite. Функция GetBankCardByNameContext, вернула ошибку: <нет содержимого в аргументе name>", err.Error(), "Нет соответствия ошибки")
	})
}
