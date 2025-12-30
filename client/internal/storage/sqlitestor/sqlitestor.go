package sqlitestor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var once sync.Once

// База данных.
type dataBase struct {
	ptrDB *sql.DB
	mu    *sync.Mutex
}

// Формат записи логин/пароль
type LoginPassword struct {
	Name      string // наименование записи.
	Login     string // логин.
	Password  string // пароль.
	CreatedAt string // время создания/обновления.
}

// Формат записи текстовых данных.
type TextData struct {
	Name      string // наименование записи.
	Text      string // текст.
	CreatedAt string // время создания/обновления.
}

// Формат записи банковской карты.
type BankCard struct {
	Name      string // наименование записи.
	Owner     string // вдажелец.
	Numb      string // номер карты.
	Valid     string // дата валидности.
	Code      string // код.
	CreatedAt string // время создания/обновления.
}

// Интерфейс.
type Actions interface {
	Close() error
	AddUserContext(ctx context.Context, userName, userPwd string) error
	AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error)
	UserExistContext(ctx context.Context) (bool, error)
	AddDataLoginPasswordContext(ctx context.Context, field1, field2, field3, createdAt string) error
	ReadTableLoginPasswordContext(ctx context.Context) (list []LoginPassword, err error)
	DelDataLoginPasswordContext(ctx context.Context, field1 string) error
	AddDataTextContext(ctx context.Context, field1, field2, createdAt string) error
	ReadTableTextContext(ctx context.Context) (list []TextData, err error)
	DelTextContext(ctx context.Context, field1 string) error
	AddDataBankCardContext(ctx context.Context, field1, field2, field3, field4, field5, createdAt string) error
	ReadTableBankCardContext(ctx context.Context) (list []BankCard, err error)
	DelBankCardContext(ctx context.Context, field1 string) error
}

var inst *dataBase

// Конструктор.
func New(dsn string) (Actions, error) {

	var err error

	once.Do(func() {
		db, connErr := connect(dsn)
		if connErr != nil {
			err = fmt.Errorf("функция connect, вернула ошибку: <%v>", connErr)
			return
		}
		if migrationErr := migrationUp(db); migrationErr != nil {
			err = fmt.Errorf("функция migrationUp, вернула ошибку: <%v>", migrationErr)
			return
		}

		inst = &dataBase{
			ptrDB: db,
			mu:    &sync.Mutex{},
		}
	})

	if err != nil {
		return nil, fmt.Errorf("Error: %v", err)
	}

	return inst, nil
}

// Закрытие подключения к БД.
func (d *dataBase) Close() error {

	d.mu.Lock()
	defer d.mu.Unlock()

	return d.ptrDB.Close()
}

// Регистрация нового пользователя.
func (d *dataBase) AddUserContext(ctx context.Context, userName, userPwd string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Формирование хеша из пароля.
	hashPwd := generateHash(userPwd)

	// Подготовка запроса.
	query := `
		INSERT INTO users (user_name, user_password) 
		VALUES (?, ?);`

	// Проверка таймаута контекста
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Выполнение запроса.
	_, err := d.ptrDB.ExecContext(ctx, query, userName, hashPwd)
	if err != nil {
		return fmt.Errorf("ошибка добавления пользователя: <%w>", err)
	}

	return nil
}

// Аутентификация пользователя.
func (d *dataBase) AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error) {

	d.mu.Lock()
	defer d.mu.Unlock()

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
	err := d.ptrDB.QueryRowContext(ctx, query, userName).Scan(&hashPwd)
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
func (d *dataBase) UserExistContext(ctx context.Context) (bool, error) {

	d.mu.Lock()
	defer d.mu.Unlock()

	row := d.ptrDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users")

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

// Добавление пары логин/пароль.
func (d *dataBase) AddDataLoginPasswordContext(ctx context.Context, field1, field2, field3, createdAt string) error {

	// Подготовка SQL-запроса
	query := `INSERT INTO data1 (field_1, field_2, field_3, created_at) VALUES (?, ?, ?, ?)`

	// Запрос.
	_, err := d.ptrDB.ExecContext(ctx, query, field1, field2, field3, createdAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных логин/пароль:<%w>", err)
	}

	return nil
}

// Получение всех записей логин/пароль из БД. Возвращается массив записей и ошибка.
func (d *dataBase) ReadTableLoginPasswordContext(ctx context.Context) (list []LoginPassword, err error) {

	limit := 100
	offset := 0
	query := "SELECT field_1, field_2, field_3, created_at FROM data1 LIMIT ? OFFSET ?"

	// Порционные запросы.
	for {
		rows, err := d.ptrDB.QueryContext(ctx, query, limit, offset)
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

// Удаление пары логин/пароль.
func (d *dataBase) DelDataLoginPasswordContext(ctx context.Context, field1 string) error {

	query := `DELETE FROM data1 WHERE field_1 = ?`

	res, err := d.ptrDB.ExecContext(ctx, query, field1)
	if err != nil {
		return fmt.Errorf("ошибка удаления данных логин/пароль:<%w>", err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения результата удаления:<%w>", err)
	}

	if cnt == 0 {
		return errors.New("удаление не выполнено")
	}

	return nil
}

//
// --- текст ---
//

// Добавление текста.
func (d *dataBase) AddDataTextContext(ctx context.Context, field1, field2, createdAt string) error {

	// Подготовка SQL-запроса
	query := `INSERT INTO data2 (field_1, field_2, created_at) VALUES (?, ?, ?)`

	// Запрос.
	_, err := d.ptrDB.ExecContext(ctx, query, field1, field2, createdAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных логин/пароль:<%w>", err)
	}

	return nil
}

// Получение всех записей текста из БД. Возвращается массив записей и ошибка.
func (d *dataBase) ReadTableTextContext(ctx context.Context) (list []TextData, err error) {

	limit := 100
	offset := 0
	query := "SELECT field_1, field_2, created_at FROM data2 LIMIT ? OFFSET ?"

	// Порционные запросы.
	for {
		rows, err := d.ptrDB.QueryContext(ctx, query, limit, offset)
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

// Удаление текста.
func (d *dataBase) DelTextContext(ctx context.Context, field1 string) error {

	query := `DELETE FROM data2 WHERE field_1 = ?`

	res, err := d.ptrDB.ExecContext(ctx, query, field1)
	if err != nil {
		return fmt.Errorf("ошибка удаления данных текста:<%w>", err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения результата удаления текста:<%w>", err)
	}

	if cnt == 0 {
		return errors.New("удаление не выполнено")
	}

	return nil
}

//
// --- банковские карты ---
//

// Добавление банковской карты.
func (d *dataBase) AddDataBankCardContext(ctx context.Context, field1, field2, field3, field4, field5, createdAt string) error {

	// Подготовка.
	query := `INSERT INTO data4 (field_1, field_2, field_3, field_4, field_5, created_at) VALUES (?, ?, ?, ?, ?, ?)`

	// Запрос.
	_, err := d.ptrDB.ExecContext(ctx, query, field1, field2, field3, field4, field5, createdAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных банковской карты:<%w>", err)
	}

	return nil
}

// Получение всех записей банковских карт из БД. Возвращается массив записей и ошибка.
func (d *dataBase) ReadTableBankCardContext(ctx context.Context) (list []BankCard, err error) {

	limit := 100
	offset := 0
	query := "SELECT field_1, field_2, field_3, field_4, field_5, created_at FROM data4 LIMIT ? OFFSET ?"

	// Порционные запросы.
	for {
		rows, err := d.ptrDB.QueryContext(ctx, query, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("функция db.QueryContext, вернула ошибку: <%w>", err)
		}
		defer rows.Close()

		recordCount := 0 // Счетчик количества прочитанных записей

		for rows.Next() {
			var el BankCard

			err := rows.Scan(&el.Name, &el.Owner, &el.Numb, &el.Valid, &el.Code, &el.CreatedAt)
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

// Удаление банковской карты.
func (d *dataBase) DelBankCardContext(ctx context.Context, field1 string) error {

	query := `DELETE FROM data4 WHERE field_1 = ?`

	res, err := d.ptrDB.ExecContext(ctx, query, field1)
	if err != nil {
		return fmt.Errorf("ошибка удаления данных банковской карты:<%w>", err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения результата удаления банковской карты:<%w>", err)
	}

	if cnt == 0 {
		return errors.New("удаление банковской карты не выполнено")
	}

	return nil
}
