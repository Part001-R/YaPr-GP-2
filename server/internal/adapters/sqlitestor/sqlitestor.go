// Обработчики пакета.
package sqlitestor

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

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
	AuthenticateUserContext(ctx context.Context, data DataUser) (bool, error)
	UserExistContext(ctx context.Context) (bool, error)
	AddDataLoginPasswordContext(ctx context.Context, data DataLoginPassword) error
	ReadTableLoginPasswordContext(ctx context.Context) (list []LoginPassword, err error)
	DelDataLoginPasswordContext(ctx context.Context, field1 string) error
	AddDataTextContext(ctx context.Context, data DataText) error
	ReadTableTextContext(ctx context.Context) (list []TextData, err error)
	DelTextContext(ctx context.Context, field1 string) error
	AddDataBankCardContext(ctx context.Context, data DataBankCard) error
	ReadTableBankCardContext(ctx context.Context) (list []BankCard, err error)
	DelBankCardContext(ctx context.Context, field1 string) error
	GetNamesLoginPasswordContext(ctx context.Context) ([]string, error)
	GetLoginPasswordByNameContext(ctx context.Context, name string) (data DataLoginPassword, err error)
	GetNamesTextContext(ctx context.Context) ([]string, error)
	GetTextByNameContext(ctx context.Context, name string) (data TextData, err error)
	GetNamesBankCardContext(ctx context.Context) ([]string, error)
	GetBankCardByNameContext(ctx context.Context, name string) (data BankCardData, err error)
	ResetForTest() error
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

	// Проверка аргументов.
	if dsn == "" {
		return nil, EmptyDataArgumentDSN
	}

	// Логика.
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

	// Проверка.
	if d.storage == nil {
		return NilPtrDB
	}

	return d.storage.Close()
}

// Регистрация нового пользователя. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	userName - имя пользователя.
//	userPwd - пароль пользователя.
func (d *dataBase) AddUserContext(ctx context.Context, userName, userPwd string) error {

	// Проверка.
	if ctx == nil {
		return EmptyDataArgumentCtx
	}
	if userName == "" {
		return EmptyDataArgumentName
	}
	if userPwd == "" {
		return EmptyDataArgumentPwd
	}
	if d.storage == nil {
		return NilPtrDB
	}
	if d.mu == nil {
		return NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Формирование хеша из пароля.
	hashPwd := generateHash(userPwd)

	// Подготовка запроса.
	query := `
		INSERT INTO users (user_name, user_password) 
		VALUES (?, ?);`

	// Выполнение запроса.
	_, err := d.storage.ExecContext(ctx, query, userName, hashPwd)
	if err != nil {
		return fmt.Errorf("ошибка добавления пользователя: <%w>", err)
	}

	return nil
}

// Аутентификация пользователя. Возвращается true - если успешно и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	data - данные.
func (d *dataBase) AuthenticateUserContext(ctx context.Context, data DataUser) (bool, error) {

	// Проверка аргументов.
	if ctx == nil {
		return false, EmptyDataArgumentCtx
	}
	if data.Field1 == "" {
		return false, EmptyDataArgumentField1
	}
	if data.Field2 == "" {
		return false, EmptyDataArgumentField2
	}
	if d.storage == nil {
		return false, NilPtrDB
	}
	if d.mu == nil {
		return false, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Подготовка запроса.
	query := `
		SELECT user_password 
		FROM users 
		WHERE user_name = ?;`

	// Выполнение запроса.
	var hashPwd string
	err := d.storage.QueryRowContext(ctx, query, data.Field1).Scan(&hashPwd)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Пользователь не найден.
		}
		return false, fmt.Errorf("ошибка выполнения запроса: <%w>", err)
	}

	// Сравнение хешей паролей.
	if !isEqualHash(data.Field2, hashPwd) {
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

	// Проверка.
	if ctx == nil {
		return false, EmptyDataArgumentCtx
	}
	if d.storage == nil {
		return false, NilPtrDB
	}
	if d.mu == nil {
		return false, NilPtrMutex
	}

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

	// Проверка аргументов
	if ctx == nil {
		return EmptyDataArgumentCtx
	}
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
	if d.storage == nil {
		return NilPtrDB
	}
	if d.mu == nil {
		return NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Подготовка SQL-запроса
	query := `INSERT INTO data1 (field_1, field_2, field_3, created_at) VALUES (?, ?, ?, ?)`

	// Запрос.
	_, err := d.storage.ExecContext(ctx, query, data.Field1, data.Field2, data.Field3, data.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных логин/пароль:<%w>", err)
	}

	return nil
}

// Получение всех записей логин/пароль из БД. Возвращается массив записей и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) ReadTableLoginPasswordContext(ctx context.Context) (list []LoginPassword, err error) {

	// Проверка аргументов.
	if ctx == nil {
		return nil, EmptyDataArgumentCtx
	}
	if d.storage == nil {
		return nil, NilPtrDB
	}
	if d.mu == nil {
		return nil, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Логика.
	limit := 100
	offset := 0
	query := "SELECT field_1, field_2, field_3, created_at FROM data1 LIMIT ? OFFSET ?"

	// Порционные запросы.
	for {
		rows, err := d.storage.QueryContext(ctx, query, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("функция db.QueryContext, вернула ошибку: <%w>", err)
		}
		defer rows.Close()

		recordCount := 0 // Счетчик количества прочитанных записей

		for rows.Next() {
			var el LoginPassword

			err := rows.Scan(&el.Name, &el.Login, &el.Password, &el.CreatedAt)
			if err != nil {
				log.Fatalf("Ошибка при считывании строки: %v", err)
			}
			list = append(list, el)
			recordCount++
		}

		if err := rows.Err(); err != nil {
			log.Fatalf("Ошибка при обработке строк: %v", err)
		}

		// Если меньше, значит записей больше нет.
		if recordCount < limit {
			break
		}
		// Изменение смещения.
		offset += limit
	}

	// Результат.
	return list, nil
}

// Удаление пары логин/пароль. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
func (d *dataBase) DelDataLoginPasswordContext(ctx context.Context, field1 string) error {

	// Проверка аргументов.
	if ctx == nil {
		return EmptyDataArgumentCtx
	}
	if field1 == "" {
		return EmptyDataArgumentField1
	}
	if d.storage == nil {
		return NilPtrDB
	}
	if d.mu == nil {
		return NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

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

// Получение имён записей для логин/пароль. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) GetNamesLoginPasswordContext(ctx context.Context) ([]string, error) {

	// Проверка аргументов.
	if ctx == nil {
		return nil, EmptyDataArgumentCtx
	}
	if d.storage == nil {
		return nil, NilPtrDB
	}
	if d.mu == nil {
		return nil, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Запрос к базе данных
	query := "SELECT field_1 FROM data1"
	rows, err := d.storage.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: <%w>", err)
	}
	defer rows.Close()

	// Массив для хранения значений field_1
	var records []string

	// Обработка
	for rows.Next() {
		var field1 string
		if err := rows.Scan(&field1); err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи: <%w>", err)
		}
		records = append(records, field1)
	}

	// Проверка на ошибки после обработки всех строк
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по записям: <%w>", err)
	}

	return records, nil
}

// Получение данных логин/пароль по имени записи. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) GetLoginPasswordByNameContext(ctx context.Context, name string) (data DataLoginPassword, err error) {

	// Провера аргументов.
	if ctx == nil {
		return DataLoginPassword{}, EmptyDataArgumentCtx
	}
	if name == "" {
		return DataLoginPassword{}, EmptyDataArgumentName
	}
	if d.storage == nil {
		return DataLoginPassword{}, NilPtrDB
	}
	if d.mu == nil {
		return DataLoginPassword{}, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	query := "SELECT field_1, field_2, field_3, created_at FROM data1 WHERE field_1 = ?"
	row := d.storage.QueryRowContext(ctx, query, name)

	err = row.Scan(&data.Field1, &data.Field2, &data.Field3, &data.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return DataLoginPassword{}, MissingData // нет записи по указанному имени
		}
		return DataLoginPassword{}, err
	}

	return data, nil
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
	if ctx == nil {
		return EmptyDataArgumentCtx
	}
	if data.Field1 == "" {
		return EmptyDataArgumentField1
	}
	if data.Field2 == "" {
		return EmptyDataArgumentField2
	}
	if data.CreatedAt == "" {
		return EmptyDataArgumentCreatedAt
	}
	if d.storage == nil {
		return NilPtrDB
	}
	if d.mu == nil {
		return NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Подготовка SQL-запроса
	query := `INSERT INTO data2 (field_1, field_2, created_at) VALUES (?, ?, ?)`

	// Запрос.
	_, err := d.storage.ExecContext(ctx, query, data.Field1, data.Field2, data.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных логин/пароль:<%w>", err)
	}

	return nil
}

// Получение всех записей текста из БД. Возвращается массив записей и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) ReadTableTextContext(ctx context.Context) (list []TextData, err error) {

	// Проверка аргументов.
	if ctx == nil {
		return nil, EmptyDataArgumentCtx
	}
	if d.storage == nil {
		return nil, NilPtrDB
	}
	if d.mu == nil {
		return nil, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Логика.
	limit := 100
	offset := 0
	query := "SELECT field_1, field_2, created_at FROM data2 LIMIT ? OFFSET ?"

	// Порционные запросы.
	for {
		rows, err := d.storage.QueryContext(ctx, query, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("функция db.QueryContext, вернула ошибку: <%w>", err)
		}
		defer rows.Close()

		recordCount := 0 // Счетчик количества прочитанных записей

		for rows.Next() {
			var el TextData

			err := rows.Scan(&el.Name, &el.Text, &el.CreatedAt)
			if err != nil {
				log.Fatalf("Ошибка при считывании строки: %v", err)
			}
			list = append(list, el)
			recordCount++
		}

		if err := rows.Err(); err != nil {
			log.Fatalf("Ошибка при обработке строк: %v", err)
		}

		// Если меньше, значит записей больше нет.
		if recordCount < limit {
			break
		}
		// Изменение смещения.
		offset += limit
	}

	// Результат.
	return list, nil
}

// Удаление текста. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
func (d *dataBase) DelTextContext(ctx context.Context, field1 string) error {

	// Провера аргументов.
	if ctx == nil {
		return EmptyDataArgumentCtx
	}
	if field1 == "" {
		return EmptyDataArgumentField1
	}
	if d.storage == nil {
		return NilPtrDB
	}
	if d.mu == nil {
		return NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

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

// Получение имён записей для текст. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) GetNamesTextContext(ctx context.Context) ([]string, error) {

	// Проверка аргументов.
	if ctx == nil {
		return nil, EmptyDataArgumentCtx
	}
	if d.storage == nil {
		return nil, NilPtrDB
	}
	if d.mu == nil {
		return nil, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Запрос к базе данных
	query := "SELECT field_1 FROM data2"
	rows, err := d.storage.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: <%w>", err)
	}
	defer rows.Close()

	// Массив для хранения значений field_1
	var records []string

	// Обработка
	for rows.Next() {
		var field1 string
		if err := rows.Scan(&field1); err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи: <%w>", err)
		}
		records = append(records, field1)
	}

	// Проверка на ошибки после обработки всех строк
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по записям: <%w>", err)
	}

	return records, nil
}

// Получение данных текста по имени записи. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) GetTextByNameContext(ctx context.Context, name string) (data TextData, err error) {

	// Проверка аргументов.
	if ctx == nil {
		return TextData{}, EmptyDataArgumentCtx
	}
	if name == "" {
		return TextData{}, EmptyDataArgumentName
	}
	if d.storage == nil {
		return TextData{}, NilPtrDB
	}
	if d.mu == nil {
		return TextData{}, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	query := "SELECT field_1, field_2, created_at FROM data2 WHERE field_1 = ?"
	row := d.storage.QueryRowContext(ctx, query, name)

	err = row.Scan(&data.Name, &data.Text, &data.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return TextData{}, nil // нет записи по указанному имени
		}
		return TextData{}, err
	}

	return data, nil
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
	if ctx == nil {
		return EmptyDataArgumentCtx
	}
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
	if d.storage == nil {
		return NilPtrDB
	}
	if d.mu == nil {
		return NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Подготовка.
	query := `INSERT INTO data4 (field_1, field_2, field_3, field_4, field_5, created_at) VALUES (?, ?, ?, ?, ?, ?)`

	// Запрос.
	_, err := d.storage.ExecContext(ctx, query, data.Field1, data.Field2, data.Field3, data.Field4, data.Field5, data.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных банковской карты:<%w>", err)
	}

	return nil
}

// Получение всех записей банковских карт из БД. Возвращается массив записей и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) ReadTableBankCardContext(ctx context.Context) (list []BankCard, err error) {

	// Проверка аргументов.
	if ctx == nil {
		return nil, EmptyDataArgumentCtx
	}
	if d.storage == nil {
		return nil, NilPtrDB
	}
	if d.mu == nil {
		return nil, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	limit := 100
	offset := 0
	query := "SELECT field_1, field_2, field_3, field_4, field_5, created_at FROM data4 LIMIT ? OFFSET ?"

	// Порционные запросы.
	for {
		rows, err := d.storage.QueryContext(ctx, query, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("функция db.QueryContext, вернула ошибку: <%w>", err)
		}
		defer rows.Close()

		recordCount := 0 // Счетчик количества прочитанных записей

		for rows.Next() {
			var el BankCard

			err := rows.Scan(&el.Field1, &el.Field2, &el.Field3, &el.Field4, &el.Field5, &el.CreatedAt)
			if err != nil {
				log.Fatalf("Ошибка при считывании строки: %v", err)
			}
			list = append(list, el)
			recordCount++
		}

		if err := rows.Err(); err != nil {
			log.Fatalf("Ошибка при обработке строк: %v", err)
		}

		// Если меньше, значит записей больше нет.
		if recordCount < limit {
			break
		}
		// Изменение смещения.
		offset += limit
	}

	// Результат.
	return list, nil
}

// Удаление банковской карты. Возвращается ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	field1 - поле имени записи.
func (d *dataBase) DelBankCardContext(ctx context.Context, field1 string) error {

	// Проверка аргументов.
	if ctx == nil {
		return EmptyDataArgumentCtx
	}
	if field1 == "" {
		return EmptyDataArgumentField1
	}
	if d.storage == nil {
		return NilPtrDB
	}
	if d.mu == nil {
		return NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

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

// Получение имён записей для банковских карт. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
func (d *dataBase) GetNamesBankCardContext(ctx context.Context) ([]string, error) {

	// Проверка аргументов.
	if ctx == nil {
		return nil, EmptyDataArgumentCtx
	}
	if d.storage == nil {
		return nil, NilPtrDB
	}
	if d.mu == nil {
		return nil, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Запрос к базе данных
	query := "SELECT field_1 FROM data4"
	rows, err := d.storage.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: <%w>", err)
	}
	defer rows.Close()

	// Массив для хранения значений field_1
	var records []string

	// Обработка
	for rows.Next() {
		var field1 string
		if err := rows.Scan(&field1); err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи: <%w>", err)
		}
		records = append(records, field1)
	}

	// Проверка на ошибки после обработки всех строк
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по записям: <%w>", err)
	}

	return records, nil
}

// Получение данных банковской карты по имени записи. Возвращаются данные и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	name - имя записи.
func (d *dataBase) GetBankCardByNameContext(ctx context.Context, name string) (data BankCardData, err error) {

	// Проверка аргументов.
	if ctx == nil {
		return BankCardData{}, EmptyDataArgumentCtx
	}
	if name == "" {
		return BankCardData{}, EmptyDataArgumentName
	}
	if d.storage == nil {
		return BankCardData{}, NilPtrDB
	}
	if d.mu == nil {
		return BankCardData{}, NilPtrMutex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	query := "SELECT field_1, field_2, field_3, field_4, field_5, created_at FROM data4 WHERE field_1 = ?"
	row := d.storage.QueryRowContext(ctx, query, name)

	err = row.Scan(&data.Name, &data.Owner, &data.Numb, &data.Valid, &data.Code, &data.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return BankCardData{}, MissingData // нет записи по указанному имени
		}
		return BankCardData{}, err
	}

	return data, nil
}

// ========================================================================================

// Сброс экземпляра, для тестов.
func (d *dataBase) ResetForTest() error {

	_, filePath, _, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("Не удалось получить информацию о вызове")
	}

	if !strings.HasSuffix(filepath.Base(filePath), "_test.go") {
		return fmt.Errorf("Эта функция может быть вызвана только из тестов")
	}

	// Сброс
	d.storage = nil
	d.mu = nil

	return nil
}
