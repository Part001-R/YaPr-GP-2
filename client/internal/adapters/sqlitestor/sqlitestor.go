// Обработчика пакета.
package sqlitestor

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// База данных.
type dataBase struct {
	storage *sql.DB     // Указатель на БД.
	mu      *sync.Mutex // Мьютекс доступа.
}

type dataBaseI interface {
	Close() error
	AddUserContext(ctx context.Context, userName, userPwd string) error
	AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error)
	UserExistContext(ctx context.Context) (bool, error)
	AddDataLoginPasswordContext(ctx context.Context, data DataLoginPassword) error
	DelDataLoginPasswordContext(ctx context.Context, field1 string) error
	AddDataTextContext(ctx context.Context, data DataText) error
	DelTextContext(ctx context.Context, field1 string) error
	AddDataBankCardContext(ctx context.Context, data DataBankCard) error
	DelBankCardContext(ctx context.Context, field1 string) error
	ReadNamesTableLoginPasswordContext(ctx context.Context) (names []string, err error)
	ReadLoginPassworByNameContext(ctx context.Context, name string) (data LoginPassword, err error)
	ReadNamesTableTextContext(ctx context.Context) (names []string, err error)
	ReadTextByNameContext(ctx context.Context, name string) (data TextData, err error)
	ReadNamesTableBankCardContext(ctx context.Context) (names []string, err error)
	ReadBankCardByNameContext(ctx context.Context, name string) (data BankCard, err error)
}

// Интерфейс.
type Actions interface {
	dataBaseI
}

// Экземпляр.
var inst *dataBase

// Конструктор. Возвращается интерфейс и ошибка.
//
// Параметры:
//
//	dsn - строка подключения к БД.
func New(dsn string) (Actions, error) {

	db, connErr := connect(dsn)
	if connErr != nil {
		return nil, fmt.Errorf("функция connect, вернула ошибку: <%v>", connErr)

	}
	if migrationErr := migrationUp(db); migrationErr != nil {
		return nil, fmt.Errorf("функция migrationUp, вернула ошибку: <%v>", migrationErr)

	}

	inst = &dataBase{
		storage: db,
		mu:      &sync.Mutex{},
	}

	return inst, nil
}

// Закрытие подключения к БД. Возвращается ошибка.
func (d *dataBase) Close() error {

	d.mu.Lock()
	defer d.mu.Unlock()

	return d.storage.Close()
}

//
// --- пользователь ---
//

// Регистрация нового пользователя. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	userName - имя пользователя.
//	userPwd - пароль пользователя.
func (d *dataBase) AddUserContext(ctx context.Context, userName, userPwd string) error {

	d.mu.Lock()
	defer d.mu.Unlock()

	// Проверка аргументов.
	if userName == "" {
		return EmptyDataArgumentName
	}
	if userPwd == "" {
		return EmptyDataArgumentPwd
	}

	// Формирование хэша из пароля.
	hashPwd := generateHash(userPwd)

	// Подготовка запроса.
	query := `
		INSERT INTO users (user_name, user_password) 
		VALUES (?, ?);`

	// Проверка таймаута контекста
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Выполнение запроса.
	_, err := d.storage.ExecContext(ctx, query, userName, hashPwd)
	if err != nil {
		return fmt.Errorf("ошибка добавления пользователя: <%w>", err)
	}

	return nil
}

// Аутентификация пользователя. Возвращается true - если пользователь аутентифицирован и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	userName - имя пользователя.
//	userPwd - пароль пользователя.
func (d *dataBase) AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error) {

	d.mu.Lock()
	defer d.mu.Unlock()

	// Проверка аргументов.
	if userName == "" {
		return false, EmptyDataArgumentName
	}
	if userPwd == "" {
		return false, EmptyDataArgumentPwd
	}

	// Подготовка запроса.
	query := `
		SELECT user_password 
		FROM users 
		WHERE user_name = ?;`

	// Проверка таймаута контекста.
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Выполнение запроса.
	var hashPwd string
	err := d.storage.QueryRowContext(ctx, query, userName).Scan(&hashPwd)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Пользователь не найден.
		}
		return false, fmt.Errorf("ошибка выполнения запроса: <%w>", err)
	}

	// Сравнение хешей паролей.
	if !isEqualHash(userPwd, hashPwd) {
		return false, nil // Пароли не совпадают.
	}

	return true, nil // Аутентификация успешна.
}

// Проверка, что уже есть зарегистрированный пользователь. Возвращается true - если уже есть запись пользователя и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) UserExistContext(ctx context.Context) (bool, error) {

	d.mu.Lock()
	defer d.mu.Unlock()

	row := d.storage.QueryRowContext(ctx, "SELECT COUNT(*) FROM users")

	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("при выполнении row.Scan, возникла ошибка: <%w>", err)
	}

	// Ответ
	if count > 0 {
		return true, nil
	}

	return false, nil
}

//
// --- логин/пароль ---
//

// Добавление пары логин/пароль. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	data - данные.
func (d *dataBase) AddDataLoginPasswordContext(ctx context.Context, data DataLoginPassword) error {

	// Проверка аргументов.
	if data.Field1 == "" {
		return EmptyDataArgumentField1
	}
	if data.Field2 == "" {
		return EmptyDataArgumentField2
	}
	if data.Field3 == "" {
		return EmptyDataArgumentField3
	}
	if data.CreatedAt == "" {
		return EmptyDataArgumentCreatedAt
	}

	// Подготовка SQL-запроса
	query := `INSERT INTO data1 (field_1, field_2, field_3, created_at) VALUES (?, ?, ?, ?)`

	// Запрос.
	_, err := d.storage.ExecContext(ctx, query, data.Field1, data.Field2, data.Field3, data.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных логин/пароль:<%w>", err)
	}

	return nil
}

// Чтение всех имен записей. Возвращается массив имён и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) ReadNamesTableLoginPasswordContext(ctx context.Context) (names []string, err error) {

	// Запрос
	query := "SELECT field_1 FROM data1"
	rows, err := d.storage.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обработка ответа.
	for rows.Next() {
		var field1 string
		if err := rows.Scan(&field1); err != nil {
			return nil, err
		}
		names = append(names, field1)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

// Чтение данных логин/пароль по имени записи. Возвращаются данные записи и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	name - имя записи.
func (d *dataBase) ReadLoginPassworByNameContext(ctx context.Context, name string) (data LoginPassword, err error) {

	// Проверка аргументов.
	if name == "" {
		return LoginPassword{}, EmptyDataArgumentName
	}

	// Запрос.
	query := "SELECT field_1, field_2, field_3, created_at FROM data1 WHERE field_1 = ?"
	row := d.storage.QueryRowContext(ctx, query, name)

	// Обработка ответа.
	err = row.Scan(&data.Name, &data.Login, &data.Password, &data.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return LoginPassword{}, MissingData
		}
		return LoginPassword{}, err
	}

	return data, nil
}

// Удаление пары логин/пароль. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
func (d *dataBase) DelDataLoginPasswordContext(ctx context.Context, field1 string) error {

	// Проверка аргументов.
	if field1 == "" {
		return EmptyDataArgumentField1
	}

	// Логика.
	query := `DELETE FROM data1 WHERE field_1 = ?`

	res, err := d.storage.ExecContext(ctx, query, field1)
	if err != nil {
		return fmt.Errorf("ошибка удаления данных логин/пароль:<%w>", err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения результата удаления:<%w>", err)
	}

	if cnt == 0 {
		return FaultDelete
	}

	return nil
}

//
// --- текст ---
//

// Добавление текста. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	data - данные.
func (d *dataBase) AddDataTextContext(ctx context.Context, data DataText) error {

	// Проверка аргументов.
	if data.Field1 == "" {
		return EmptyDataArgumentField1
	}
	if data.Field2 == "" {
		return EmptyDataArgumentField2
	}
	if data.CreatedAt == "" {
		return EmptyDataArgumentCreatedAt
	}

	// Подготовка SQL-запроса
	query := `INSERT INTO data2 (field_1, field_2, created_at) VALUES (?, ?, ?)`

	// Запрос.
	_, err := d.storage.ExecContext(ctx, query, data.Field1, data.Field2, data.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных логин/пароль:<%w>", err)
	}

	return nil
}

// Чтение всех имен записей текста. Возвращается массив имён и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) ReadNamesTableTextContext(ctx context.Context) (names []string, err error) {

	// Запрос
	query := "SELECT field_1 FROM data2"
	rows, err := d.storage.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обработка ответа.
	for rows.Next() {
		var field1 string
		if err := rows.Scan(&field1); err != nil {
			return nil, err
		}
		names = append(names, field1)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

// Чтение данных текста по имени записи. Возвращаются данные записи и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	name - имя записи.
func (d *dataBase) ReadTextByNameContext(ctx context.Context, name string) (data TextData, err error) {

	// Проверка аргументов.
	if name == "" {
		return TextData{}, EmptyDataArgumentName
	}

	query := "SELECT field_1, field_2, created_at FROM data2 WHERE field_1 = ?"
	row := d.storage.QueryRowContext(ctx, query, name)

	// Обработка ответа.
	err = row.Scan(&data.Name, &data.Text, &data.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return TextData{}, nil
		}
		return TextData{}, err
	}

	return data, nil
}

// Удаление текста. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
func (d *dataBase) DelTextContext(ctx context.Context, field1 string) error {

	// Проверка аргументов
	if field1 == "" {
		return EmptyDataArgumentField1
	}

	// Логика.
	query := `DELETE FROM data2 WHERE field_1 = ?`

	res, err := d.storage.ExecContext(ctx, query, field1)
	if err != nil {
		return fmt.Errorf("ошибка удаления данных текста:<%w>", err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения результата удаления текста:<%w>", err)
	}

	if cnt == 0 {
		return FaultDelete
	}

	return nil
}

//
// --- банковские карты ---
//

// Добавление банковской карты. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	data - данные.
func (d *dataBase) AddDataBankCardContext(ctx context.Context, data DataBankCard) error {

	// Проверка аргументов.
	if data.Field1 == "" {
		return EmptyDataArgumentField1
	}
	if data.Field2 == "" {
		return EmptyDataArgumentField2
	}
	if data.Field3 == "" {
		return EmptyDataArgumentField3
	}
	if data.Field4 == "" {
		return EmptyDataArgumentField4
	}
	if data.Field5 == "" {
		return EmptyDataArgumentField5
	}
	if data.CreatedAt == "" {
		return EmptyDataArgumentCreatedAt
	}

	// Подготовка.
	query := `INSERT INTO data4 (field_1, field_2, field_3, field_4, field_5, created_at) VALUES (?, ?, ?, ?, ?, ?)`

	// Запрос.
	_, err := d.storage.ExecContext(ctx, query, data.Field1, data.Field2, data.Field3, data.Field4, data.Field5, data.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных банковской карты:<%w>", err)
	}

	return nil
}

// Удаление банковской карты. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
func (d *dataBase) DelBankCardContext(ctx context.Context, field1 string) error {

	// Проверка аргументов.
	if field1 == "" {
		return EmptyDataArgumentField1
	}

	// Логика.
	query := `DELETE FROM data4 WHERE field_1 = ?`

	res, err := d.storage.ExecContext(ctx, query, field1)
	if err != nil {
		return fmt.Errorf("ошибка удаления данных банковской карты:<%w>", err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения результата удаления банковской карты:<%w>", err)
	}

	if cnt == 0 {
		return FaultDelete
	}

	return nil
}

// Чтение всех имен записей банковских карт. Возвращается массив имён и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) ReadNamesTableBankCardContext(ctx context.Context) (names []string, err error) {

	// Запрос
	query := "SELECT field_1 FROM data4"
	rows, err := d.storage.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обработка ответа.
	for rows.Next() {
		var field1 string
		if err := rows.Scan(&field1); err != nil {
			return nil, err
		}
		names = append(names, field1)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

// Чтение данных банковской карты по имени записи. Возвращаются данные записи и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	name - имя записи.
func (d *dataBase) ReadBankCardByNameContext(ctx context.Context, name string) (data BankCard, err error) {

	// Проверка аргументов.
	if name == "" {
		return BankCard{}, EmptyDataArgumentName
	}

	// Запрос.
	query := "SELECT field_1, field_2, field_3, field_4, field_5, created_at FROM data4 WHERE field_1 = ?"
	row := d.storage.QueryRowContext(ctx, query, name)

	// Обработка ответа.
	err = row.Scan(&data.Name, &data.Owner, &data.Numb, &data.Valid, &data.Code, &data.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return BankCard{}, MissingData
		}
		return BankCard{}, err
	}

	return data, nil
}
