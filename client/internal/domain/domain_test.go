package domain

import (
	"context"
	"database/sql"
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

	// Создание БД через домен.
	dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
	actions, err := NewStorage(dsn)
	require.NoErrorf(t, err, "ошибка конструктора")

	defer func() {
		err := os.Remove("testStorage.db")
		assert.NoErrorf(t, err, "Ошибка удаления БД")
	}()

	// Добавление записи.
	userName := "Foo"
	userPwd := "Bar"
	err = actions.AddUserContext(context.Background(), userName, userPwd)
	require.NoErrorf(t, err, "ошибка добавления пользователя")

	err = actions.Close()
	require.NoErrorf(t, err, "Ошибка закрытия подключения")

	// Запрос данных.
	db, err := sql.Open("sqlite", dsn)
	require.NoErrorf(t, err, "Ошибка подключения")

	defer func() {
		err := db.Close()
		assert.NoErrorf(t, err, "Ошибка закрытия подключения")
	}()

	rxData, err := getAllUsers(db)
	assert.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

	// Проверка результата.
	assert.Equalf(t, 1, len(rxData), "Нет соответствия размера записей")
	assert.Equalf(t, userName, rxData[0].Name, "Нет соответствия имени")
}

//
// --- AuthenticateUserContext ---
//

func TestAuthenticateUserContext(t *testing.T) {

	t.Run("Успешная аутентификация", func(t *testing.T) {

		// Создание БД через домен.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		isAuth, err := actions.AuthenticateUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "Ошибка аутентификации")
		assert.True(t, isAuth, "пользователь не уатентифицирован")
	})

	t.Run("Нет такого пользователя", func(t *testing.T) {

		// Создание БД через домен.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		isAuth, err := actions.AuthenticateUserContext(context.Background(), userName+"A", userPwd)
		require.NoErrorf(t, err, "Ошибка аутентификации")
		assert.Falsef(t, isAuth, "пользователь аутентифицирован")
	})

	t.Run("Неверный пароль", func(t *testing.T) {

		// Создание БД через домен.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		isAuth, err := actions.AuthenticateUserContext(context.Background(), userName, userPwd+"A")
		require.NoErrorf(t, err, "Ошибка аутентификации")
		assert.Falsef(t, isAuth, "пользователь аутентифицирован")
	})

	t.Run("Нет имени", func(t *testing.T) {

		// Создание БД через домен.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		_, err = actions.AuthenticateUserContext(context.Background(), "", userPwd)
		require.Errorf(t, err, "Ошибка аутентификации")
	})

	t.Run("Нет пароля", func(t *testing.T) {

		// Создание БД через домен.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Добавление записи.
		userName := "Foo"
		userPwd := "Bar"
		err = actions.AddUserContext(context.Background(), userName, userPwd)
		require.NoErrorf(t, err, "ошибка добавления пользователя")

		// Аутентификация.
		_, err = actions.AuthenticateUserContext(context.Background(), userName, "")
		require.Errorf(t, err, "Ошибка аутентификации")
	})
}

//
// --- UserExistContext ---
//

func TestUserExistContext(t *testing.T) {

	t.Run("Нет пользователя в системе", func(t *testing.T) {

		// Создание БД через домен.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
			assert.NoErrorf(t, err, "Ошибка удаления БД")
		}()

		// Проверка пользователя.
		isExsist, err := actions.UserExistContext(context.Background())
		require.NoErrorf(t, err, "ошибка добавления пользователя")
		assert.Falsef(t, isExsist, "Есть пользователь")
	})

	t.Run("Есть пользователь в системе", func(t *testing.T) {

		// Создание БД через домен.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err = actions.Close()
			require.NoErrorf(t, err, "Ошибка закрытия подключения")

			err := os.Remove("testStorage.db")
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

		// Создание БД через домен.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err := os.Remove("testStorage.db")
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

		err = actions.Close()
		require.NoErrorf(t, err, "Ошибка закрытия подключения")

		// Запрос данных.
		db, err := sql.Open("sqlite", dsn)
		require.NoErrorf(t, err, "Ошибка подключения")

		defer func() {
			err := db.Close()
			assert.NoErrorf(t, err, "Ошибка закрытия подключения")
		}()

		rxData, err := getAllLoginPassword(db)
		assert.NoErrorf(t, err, "Ошибка запроса:<%v>", err)

		// Проверка результата.
		assert.Equalf(t, 1, len(rxData), "Нет соответствия размера записей")
		assert.Equalf(t, data.Field1, rxData[0].Field1, "Нет соответствия Field1")
		assert.Equalf(t, data.Field2, rxData[0].Field2, "Нет соответствия Field2")
		assert.Equalf(t, data.Field3, rxData[0].Field3, "Нет соответствия Field3")
		assert.Equalf(t, data.CreatedAt, rxData[0].CreatedAt, "Нет соответствия CreatedAt")
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

		// Создание БД через домен.
		dsn := "file:testStorage.db?cache=shared&foreign_keys=on&mode=rwc"
		actions, err := NewStorage(dsn)
		require.NoErrorf(t, err, "ошибка конструктора")

		defer func() {
			err := os.Remove("testStorage.db")
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

// ------------------------------------------------
//
//             Вспомогательные функции
//
// ------------------------------------------------

// Для теста. Прлучение всех записей таблицы пользователей. Возвращается массив и ошибка.
//
// Параметры:
//
//	db - указатель на БД.
func getAllUsers(db *sql.DB) ([]User, error) {

	rows, err := db.Query("SELECT id, user_name, user_password, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Pwd, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Для теста. Прлучение всех записей таблицы логин/пароль. Возвращается массив и ошибка.
//
// Параметры:
//
//	db - указатель на БД.
func getAllLoginPassword(db *sql.DB) ([]DataLoginPassword, error) {

	rows, err := db.Query("SELECT field_1, field_2, field_3, created_at FROM data1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []DataLoginPassword
	for rows.Next() {
		var record DataLoginPassword
		if err := rows.Scan(&record.Field1, &record.Field2, &record.Field3, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}
