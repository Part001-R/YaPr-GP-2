package sqlitestor

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// --- AddUserContext ---
//

func TestAddUserContext(t *testing.T) {

	db, mock, err := sqlmock.New()
	require.NoErrorf(t, err, "ошибка создания мок БД")
	defer db.Close()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("Нет имени", func(t *testing.T) {
		err = d.AddUserContext(context.Background(), "", "Bar")
		assert.Equalf(t, EmptyDataArgumentName, err, "Нет соответствия ошибки")
	})

	t.Run("Нет пароля", func(t *testing.T) {
		err = d.AddUserContext(context.Background(), "Foo", "")
		assert.Equalf(t, EmptyDataArgumentPwd, err, "Нет соответствия ошибки")
	})

	t.Run("Успешное добавление", func(t *testing.T) {
		userName := "Foo"
		userPwd := "Bar"
		hashPwd := generateHash(userPwd)

		mock.ExpectExec(`INSERT INTO users`).
			WithArgs(userName, hashPwd).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = d.AddUserContext(context.Background(), userName, userPwd)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

//
// --- AuthenticateUserContext ---
//

func TestAuthenticateUserContext(t *testing.T) {

	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() {
		_ = db.Close()
	}()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("пустое имя пользователя", func(t *testing.T) {
		ok, err := d.AuthenticateUserContext(context.Background(), "", "Bar")
		assert.False(t, ok)
		assert.Equal(t, EmptyDataArgumentName, err)
	})

	t.Run("пустой пароль", func(t *testing.T) {
		ok, err := d.AuthenticateUserContext(context.Background(), "Foo", "")
		assert.False(t, ok)
		assert.Equal(t, EmptyDataArgumentPwd, err)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		mock.ExpectQuery(`SELECT user_password FROM users WHERE user_name = \?`).
			WithArgs("Foo").
			WillReturnError(sql.ErrNoRows)

		ok, err := d.AuthenticateUserContext(context.Background(), "Foo", "Bar")

		assert.NoError(t, mock.ExpectationsWereMet(), "не все ожидания мок-объекта были выполнены")

		assert.False(t, ok)
		assert.NoError(t, err)
	})

	t.Run("ошибка выполнения запроса", func(t *testing.T) {
		mock.ExpectQuery(`SELECT user_password FROM users WHERE user_name = \?`).
			WithArgs("Foo").
			WillReturnError(context.DeadlineExceeded)

		ok, err := d.AuthenticateUserContext(context.Background(), "Foo", "Bar")

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.False(t, ok)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка выполнения запроса")
	})

	t.Run("пароль не совпадает", func(t *testing.T) {
		mock.ExpectQuery(`SELECT user_password FROM users WHERE user_name = \?`).
			WithArgs("Foo").
			WillReturnRows(sqlmock.NewRows([]string{"user_password"}).AddRow("wrong_hash"))

		ok, err := d.AuthenticateUserContext(context.Background(), "Foo", "Bar")

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.False(t, ok)
		assert.NoError(t, err)
	})

	t.Run("успешная аутентификация", func(t *testing.T) {
		correctHash := generateHash("Bar")

		mock.ExpectQuery(`SELECT user_password FROM users WHERE user_name = \?`).
			WithArgs("Foo").
			WillReturnRows(sqlmock.NewRows([]string{"user_password"}).AddRow(correctHash))

		ok, err := d.AuthenticateUserContext(context.Background(), "Foo", "Bar")

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.True(t, ok)
		assert.NoError(t, err)
	})
}

//
// --- UserExistContext ---
//

func TestUserExistContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() {
		_ = db.Close()
	}()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("Пользователи существуют", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").WillReturnRows(rows)

		ok, err := d.UserExistContext(context.Background())
		require.NoError(t, err, "ошибка вызова UserExistContext")
		assert.True(t, ok)
	})

	t.Run("Пользователи отсутствуют", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").WillReturnRows(rows)

		ok, err := d.UserExistContext(context.Background())
		require.NoError(t, err, "ошибка вызова UserExistContext")
		assert.False(t, ok)
	})

	t.Run("Ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").WillReturnError(errors.New("ошибка"))

		ok, err := d.UserExistContext(context.Background())

		assert.Error(t, err, "ожидалась ошибка вызова UserExistContext")
		assert.False(t, ok)

		assert.Contains(t, err.Error(), "ошибка", "сообщение об ошибке должно содержать 'ошибка'")
	})
}

//
// --- DelDataLoginPasswordContext ---
//

func TestDelDataLoginPasswordContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("пустой field1", func(t *testing.T) {
		err := d.DelDataLoginPasswordContext(context.Background(), "")
		assert.ErrorIs(t, err, EmptyDataArgumentField1)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("успешное удаление (1 строка затронута)", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM data1 WHERE field_1 = \?`).
			WithArgs("valid_field1").
			WillReturnResult(sqlmock.NewResult(1, 1)) // 1 строка удалена

		err := d.DelDataLoginPasswordContext(context.Background(), "valid_field1")

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("удаление не затронуло строки (0 строк)", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM data1 WHERE field_1 = \?`).
			WithArgs("no_rows_field1").
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк удалено

		err := d.DelDataLoginPasswordContext(context.Background(), "no_rows_field1")

		assert.ErrorIs(t, err, FaultDelete)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка выполнения ExecContext", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM data1 WHERE field_1 = \?`).
			WithArgs("error_field1").
			WillReturnError(errors.New("db exec error"))

		err := d.DelDataLoginPasswordContext(context.Background(), "error_field1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка удаления данных логин/пароль")
		assert.Contains(t, err.Error(), "db exec error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

//
// --- AddDataTextContext ---
//

func TestAddDataTextContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	// Общий тестовый объект DataText
	validData := DataText{
		Field1:    "valid_field1",
		Field2:    "valid_field2",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	t.Run("пустой Field1", func(t *testing.T) {
		data := DataText{
			Field1:    "",
			Field2:    validData.Field2,
			CreatedAt: validData.CreatedAt,
		}

		err := d.AddDataTextContext(context.Background(), data)
		assert.ErrorIs(t, err, EmptyDataArgumentField1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пустой Field2", func(t *testing.T) {
		data := DataText{
			Field1:    validData.Field1,
			Field2:    "",
			CreatedAt: validData.CreatedAt,
		}

		err := d.AddDataTextContext(context.Background(), data)
		assert.ErrorIs(t, err, EmptyDataArgumentField2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пустой CreatedAt", func(t *testing.T) {
		data := DataText{
			Field1:    validData.Field1,
			Field2:    validData.Field2,
			CreatedAt: "",
		}

		err := d.AddDataTextContext(context.Background(), data)
		assert.ErrorIs(t, err, EmptyDataArgumentCreatedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("успешное добавление данных", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO data2 \(field_1, field_2, created_at\) VALUES \(\?, \?, \?\)`).
			WithArgs(validData.Field1, validData.Field2, validData.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := d.AddDataTextContext(context.Background(), validData)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка выполнения ExecContext", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO data2 \(field_1, field_2, created_at\) VALUES \(\?, \?, \?\)`).
			WithArgs(validData.Field1, validData.Field2, validData.CreatedAt).
			WillReturnError(errors.New("db exec error"))

		err := d.AddDataTextContext(context.Background(), validData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка добавления данных логин/пароль")
		assert.Contains(t, err.Error(), "db exec error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

//
// --- DelTextContext ---
//

func TestDelTextContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("пустой field1", func(t *testing.T) {
		err := d.DelTextContext(context.Background(), "")
		assert.ErrorIs(t, err, EmptyDataArgumentField1)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("успешное удаление (1 строка затронута)", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM data2 WHERE field_1 = \?`).
			WithArgs("valid_field1").
			WillReturnResult(sqlmock.NewResult(1, 1)) // 1 строка удалена

		err := d.DelTextContext(context.Background(), "valid_field1")

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("удаление не затронуло строки (0 строк)", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM data2 WHERE field_1 = \?`).
			WithArgs("no_rows_field1").
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк удалено

		err := d.DelTextContext(context.Background(), "no_rows_field1")

		assert.ErrorIs(t, err, FaultDelete)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка выполнения ExecContext", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM data2 WHERE field_1 = \?`).
			WithArgs("error_field1").
			WillReturnError(errors.New("db exec error"))

		err := d.DelTextContext(context.Background(), "error_field1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка удаления данных текста")
		assert.Contains(t, err.Error(), "db exec error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

//
// --- AddDataBankCardContext ---
//

func TestAddDataBankCardContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	validData := DataBankCard{
		Field1:    "field1",
		Field2:    "field2",
		Field3:    "field3",
		Field4:    "field4",
		Field5:    "field5",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	t.Run("пустой Field1", func(t *testing.T) {
		data := validData
		data.Field1 = ""

		err := d.AddDataBankCardContext(context.Background(), data)
		assert.ErrorIs(t, err, EmptyDataArgumentField1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пустой Field2", func(t *testing.T) {
		data := validData
		data.Field2 = ""

		err := d.AddDataBankCardContext(context.Background(), data)
		assert.ErrorIs(t, err, EmptyDataArgumentField2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пустой Field3", func(t *testing.T) {
		data := validData
		data.Field3 = ""

		err := d.AddDataBankCardContext(context.Background(), data)
		assert.ErrorIs(t, err, EmptyDataArgumentField3)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пустой Field4", func(t *testing.T) {
		data := validData
		data.Field4 = ""

		err := d.AddDataBankCardContext(context.Background(), data)
		assert.ErrorIs(t, err, EmptyDataArgumentField4)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пустой Field5", func(t *testing.T) {
		data := validData
		data.Field5 = ""

		err := d.AddDataBankCardContext(context.Background(), data)
		assert.ErrorIs(t, err, EmptyDataArgumentField5)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пустой CreatedAt", func(t *testing.T) {
		data := validData
		data.CreatedAt = ""

		err := d.AddDataBankCardContext(context.Background(), data)
		assert.ErrorIs(t, err, EmptyDataArgumentCreatedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("успешное добавление данных", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO data4 \(field_1, field_2, field_3, field_4, field_5, created_at\) VALUES \(\?, \?, \?, \?, \?, \?\)`).
			WithArgs(
				validData.Field1,
				validData.Field2,
				validData.Field3,
				validData.Field4,
				validData.Field5,
				validData.CreatedAt,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := d.AddDataBankCardContext(context.Background(), validData)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка выполнения ExecContext", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO data4 \(field_1, field_2, field_3, field_4, field_5, created_at\) VALUES \(\?, \?, \?, \?, \?, \?\)`).
			WithArgs(
				validData.Field1,
				validData.Field2,
				validData.Field3,
				validData.Field4,
				validData.Field5,
				validData.CreatedAt,
			).
			WillReturnError(errors.New("db exec error"))

		err := d.AddDataBankCardContext(context.Background(), validData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка добавления данных банковской карты")
		assert.Contains(t, err.Error(), "db exec error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

//
// --- DelBankCardContext ---
//

func TestDelBankCardContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("пустой field1", func(t *testing.T) {
		err := d.DelBankCardContext(context.Background(), "")
		assert.ErrorIs(t, err, EmptyDataArgumentField1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("успешное удаление (1 строка затронута)", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM data4 WHERE field_1 = \?`).
			WithArgs("valid_field1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := d.DelBankCardContext(context.Background(), "valid_field1")

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("удаление не затронуло строки (0 строк)", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM data4 WHERE field_1 = \?`).
			WithArgs("no_rows_field1").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := d.DelBankCardContext(context.Background(), "no_rows_field1")

		assert.ErrorIs(t, err, FaultDelete)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка выполнения ExecContext", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM data4 WHERE field_1 = \?`).
			WithArgs("error_field1").
			WillReturnError(errors.New("db exec error"))

		err := d.DelBankCardContext(context.Background(), "error_field1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка удаления данных банковской карты")
		assert.Contains(t, err.Error(), "db exec error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

//
// --- ReadNamesTableLoginPasswordContext ---
//

func TestReadNamesTableLoginPasswordContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("успешное чтение (несколько записей)", func(t *testing.T) {

		mock.ExpectQuery("SELECT field_1 FROM data1").
			WillReturnRows(
				sqlmock.NewRows([]string{"field_1"}).
					AddRow("name1").
					AddRow("name2").
					AddRow("name3"),
			)

		names, err := d.ReadNamesTableLoginPasswordContext(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, []string{"name1", "name2", "name3"}, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("успешное чтение (пустая таблица)", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1 FROM data1").
			WillReturnRows(sqlmock.NewRows([]string{"field_1"}))

		names, err := d.ReadNamesTableLoginPasswordContext(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, names) // пустой срез
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка выполнения QueryContext", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1 FROM data1").
			WillReturnError(errors.New("db query error"))

		names, err := d.ReadNamesTableLoginPasswordContext(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db query error")
		assert.Nil(t, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка сканирования строки", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"field_1"}).
			AddRow("name1").
			AddRow(sql.NullString{Valid: false})

		mock.ExpectQuery("SELECT field_1 FROM data1").WillReturnRows(rows)

		names, err := d.ReadNamesTableLoginPasswordContext(context.Background())

		assert.Error(t, err)
		assert.Nil(t, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

//
// --- ReadLoginPassworByNameContext ---
//

func TestReadLoginPassworByNameContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("пустой name", func(t *testing.T) {
		data, err := d.ReadLoginPassworByNameContext(context.Background(), "")

		assert.ErrorIs(t, err, EmptyDataArgumentName)
		assert.Equal(t, LoginPassword{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("запись найдена", func(t *testing.T) {
		expectedData := LoginPassword{
			Name:      "test_user",
			Login:     "user_login",
			Password:  "secure_password",
			CreatedAt: "2023-01-01T00:00:00Z",
		}

		mock.ExpectQuery("SELECT field_1, field_2, field_3, created_at FROM data1 WHERE field_1 = \\?").
			WithArgs("test_user").
			WillReturnRows(
				sqlmock.NewRows([]string{"field_1", "field_2", "field_3", "created_at"}).
					AddRow(expectedData.Name, expectedData.Login, expectedData.Password, expectedData.CreatedAt),
			)

		data, err := d.ReadLoginPassworByNameContext(context.Background(), "test_user")

		assert.NoError(t, err)
		assert.Equal(t, expectedData, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("запись не найдена (sql.ErrNoRows)", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1, field_2, field_3, created_at FROM data1 WHERE field_1 = \\?").
			WithArgs("non_existent_user").
			WillReturnError(sql.ErrNoRows)

		data, err := d.ReadLoginPassworByNameContext(context.Background(), "non_existent_user")

		assert.NoError(t, err)
		assert.Equal(t, LoginPassword{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("общая ошибка БД при запросе", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1, field_2, field_3, created_at FROM data1 WHERE field_1 = \\?").
			WithArgs("db_error_user").
			WillReturnError(errors.New("database query failed"))

		data, err := d.ReadLoginPassworByNameContext(context.Background(), "db_error_user")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database query failed")
		assert.Equal(t, LoginPassword{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

//
// --- ReadNamesTableTextContext ---
//

func TestReadNamesTableTextContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("успешное чтение (несколько записей)", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1 FROM data2").
			WillReturnRows(
				sqlmock.NewRows([]string{"field_1"}).
					AddRow("text1").
					AddRow("text2").
					AddRow("text3"),
			)

		names, err := d.ReadNamesTableTextContext(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, []string{"text1", "text2", "text3"}, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("успешное чтение (пустая таблица)", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1 FROM data2").
			WillReturnRows(sqlmock.NewRows([]string{"field_1"}))

		names, err := d.ReadNamesTableTextContext(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка выполнения QueryContext", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1 FROM data2").
			WillReturnError(errors.New("db query error"))

		names, err := d.ReadNamesTableTextContext(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db query error")
		assert.Nil(t, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка сканирования строки", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"field_1"}).
			AddRow("text1").
			// Следующая строка вызовет ошибку при сканировании
			AddRow(sql.NullString{Valid: false})

		mock.ExpectQuery("SELECT field_1 FROM data2").WillReturnRows(rows)

		names, err := d.ReadNamesTableTextContext(context.Background())

		assert.Error(t, err)
		assert.Nil(t, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

//
// --- ReadTextByNameContext ---
//

func TestReadTextByNameContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("пустой name", func(t *testing.T) {
		data, err := d.ReadTextByNameContext(context.Background(), "")

		assert.ErrorIs(t, err, EmptyDataArgumentName)
		assert.Equal(t, TextData{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("запись найдена", func(t *testing.T) {
		expectedData := TextData{
			Name:      "sample_name",
			Text:      "sample_text",
			CreatedAt: "2023-01-01T00:00:00Z",
		}

		mock.ExpectQuery("SELECT field_1, field_2, created_at FROM data2 WHERE field_1 = \\?").
			WithArgs("sample_name").
			WillReturnRows(
				sqlmock.NewRows([]string{"field_1", "field_2", "created_at"}).
					AddRow(expectedData.Name, expectedData.Text, expectedData.CreatedAt),
			)

		data, err := d.ReadTextByNameContext(context.Background(), "sample_name")

		assert.NoError(t, err)
		assert.Equal(t, expectedData, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("запись не найдена (sql.ErrNoRows)", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1, field_2, created_at FROM data2 WHERE field_1 = \\?").
			WithArgs("non_existent_name").
			WillReturnError(sql.ErrNoRows)

		data, err := d.ReadTextByNameContext(context.Background(), "non_existent_name")

		assert.NoError(t, err)
		assert.Equal(t, TextData{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка сканирования данных", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1, field_2, created_at FROM data2 WHERE field_1 = \\?").
			WithArgs("bad_data_name").
			WillReturnRows(
				sqlmock.NewRows([]string{"field_1", "field_2", "created_at"}).
					AddRow("valid_name", nil, "2023-01-01T00:00:00Z"),
			)

		data, err := d.ReadTextByNameContext(context.Background(), "bad_data_name")

		assert.Error(t, err)
		assert.Equal(t, TextData{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("общая ошибка БД при запросе", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1, field_2, created_at FROM data2 WHERE field_1 = \\?").
			WithArgs("db_error_name").
			WillReturnError(errors.New("database query failed"))

		data, err := d.ReadTextByNameContext(context.Background(), "db_error_name")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database query failed")
		assert.Equal(t, TextData{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

//
// --- ReadNamesTableBankCardContext ---
//

func TestReadNamesTableBankCardContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("успешное чтение (несколько записей)", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1 FROM data4").
			WillReturnRows(
				sqlmock.NewRows([]string{"field_1"}).
					AddRow("card1").
					AddRow("card2").
					AddRow("card3"),
			)

		names, err := d.ReadNamesTableBankCardContext(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, []string{"card1", "card2", "card3"}, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("успешное чтение (пустая таблица)", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1 FROM data4").
			WillReturnRows(sqlmock.NewRows([]string{"field_1"}))

		names, err := d.ReadNamesTableBankCardContext(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка выполнения QueryContext", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1 FROM data4").
			WillReturnError(errors.New("db query error"))

		names, err := d.ReadNamesTableBankCardContext(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db query error")
		assert.Nil(t, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка сканирования строки", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"field_1"}).
			AddRow("card1").
			// Следующая строка вызовет ошибку при сканировании
			AddRow(sql.NullString{Valid: false})

		mock.ExpectQuery("SELECT field_1 FROM data4").WillReturnRows(rows)

		names, err := d.ReadNamesTableBankCardContext(context.Background())

		assert.Error(t, err)
		assert.Nil(t, names)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

//
// --- ReadBankCardByNameContext ---
//

func TestReadBankCardByNameContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "ошибка создания мок БД")
	defer func() { _ = db.Close() }()

	d := &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	t.Run("пустой name", func(t *testing.T) {
		data, err := d.ReadBankCardByNameContext(context.Background(), "")

		assert.ErrorIs(t, err, EmptyDataArgumentName)
		assert.Equal(t, BankCard{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("запись найдена", func(t *testing.T) {
		expectedData := BankCard{
			Name:      "card_123",
			Owner:     "A",
			Numb:      "4111111111111111",
			Valid:     "B",
			Code:      "C",
			CreatedAt: "2023-01-01T00:00:00Z",
		}

		mock.ExpectQuery("SELECT field_1, field_2, field_3, field_4, field_5, created_at FROM data4 WHERE field_1 = \\?").
			WithArgs("card_123").
			WillReturnRows(
				sqlmock.NewRows([]string{"field_1", "field_2", "field_3", "field_4", "field_5", "created_at"}).
					AddRow(
						expectedData.Name,
						expectedData.Owner,
						expectedData.Numb,
						expectedData.Valid,
						expectedData.Code,
						expectedData.CreatedAt,
					),
			)

		data, err := d.ReadBankCardByNameContext(context.Background(), "card_123")

		assert.NoError(t, err)
		assert.Equal(t, expectedData, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("запись не найдена (sql.ErrNoRows)", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1, field_2, field_3, field_4, field_5, created_at FROM data4 WHERE field_1 = \\?").
			WithArgs("non_existent_card").
			WillReturnError(sql.ErrNoRows)

		data, err := d.ReadBankCardByNameContext(context.Background(), "non_existent_card")

		assert.NoError(t, err)
		assert.Equal(t, BankCard{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка сканирования данных", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1, field_2, field_3, field_4, field_5, created_at FROM data4 WHERE field_1 = \\?").
			WithArgs("bad_data_card").
			WillReturnRows(
				sqlmock.NewRows([]string{"field_1", "field_2", "field_3", "field_4", "field_5", "created_at"}).
					AddRow("valid_name", "owner", nil, "12/25", "123", "2023-01-01T00:00:00Z"),
			)

		data, err := d.ReadBankCardByNameContext(context.Background(), "bad_data_card")

		assert.Error(t, err)
		assert.Equal(t, BankCard{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("общая ошибка БД при запросе", func(t *testing.T) {
		mock.ExpectQuery("SELECT field_1, field_2, field_3, field_4, field_5, created_at FROM data4 WHERE field_1 = \\?").
			WithArgs("db_error_card").
			WillReturnError(errors.New("database query failed"))

		data, err := d.ReadBankCardByNameContext(context.Background(), "db_error_card")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database query failed")
		assert.Equal(t, BankCard{}, data)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
