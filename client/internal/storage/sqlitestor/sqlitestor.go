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
type DataBase struct {
	PtrDB *sql.DB
	mu    *sync.Mutex
}

// Формат записи логин/пароль
type LoginPassword struct {
	Name      string // наименование записи.
	Login     string // логин.
	Password  string // пароль.
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
}

var inst *DataBase

// Конструктор.
func New(db *sql.DB) Actions {
	once.Do(func() {
		inst = &DataBase{
			PtrDB: db,
			mu:    &sync.Mutex{},
		}
	})
	return inst
}

// Закрытие подключения к БД.
func (d *DataBase) Close() error {

	d.mu.Lock()
	defer d.mu.Unlock()

	return d.PtrDB.Close()
}

// Регистрация нового пользователя.
func (d *DataBase) AddUserContext(ctx context.Context, userName, userPwd string) error {
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
	_, err := d.PtrDB.ExecContext(ctx, query, userName, hashPwd)
	if err != nil {
		return fmt.Errorf("ошибка добавления пользователя: <%w>", err)
	}

	return nil
}

// Аутентификация пользователя.
func (d *DataBase) AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error) {

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
	err := d.PtrDB.QueryRowContext(ctx, query, userName).Scan(&hashPwd)
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
func (d *DataBase) UserExistContext(ctx context.Context) (bool, error) {

	d.mu.Lock()
	defer d.mu.Unlock()

	row := d.PtrDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users")

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

// Добавление пары логин/пароль.
func (d *DataBase) AddDataLoginPasswordContext(ctx context.Context, field1, field2, field3, createdAt string) error {

	// Подготовка SQL-запроса
	query := `INSERT INTO data1 (field_1, field_2, field_3, created_at) VALUES (?, ?, ?, ?)`

	// Запрос.
	_, err := d.PtrDB.ExecContext(ctx, query, field1, field2, field3, createdAt)
	if err != nil {
		return fmt.Errorf("ошибка добавления данных логин/пароль:<%w>", err)
	}

	return nil
}

// Получение всех записей логин/пароль из БД. Возвращается массив записей и ошибка.
func (d *DataBase) ReadTableLoginPasswordContext(ctx context.Context) (list []LoginPassword, err error) {

	limit := 100
	offset := 0
	query := "SELECT field_1, field_2, field_3, created_at FROM data1 LIMIT ? OFFSET ?"

	// Порционные запросы.
	for {
		rows, err := d.PtrDB.QueryContext(ctx, query, limit, offset)
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
func (d *DataBase) DelDataLoginPasswordContext(ctx context.Context, field1 string) error {

	query := `DELETE FROM data1 WHERE field_1 = ?`

	res, err := d.PtrDB.ExecContext(ctx, query, field1)
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
