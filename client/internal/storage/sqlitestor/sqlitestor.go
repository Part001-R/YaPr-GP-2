package sqlitestor

import (
	"context"
	"database/sql"
	"fmt"
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

// Интерфейс.
type Actions interface {
	Close() error
	AddUserContext(ctx context.Context, userName, userPwd string) error
	AuthenticateUserContext(ctx context.Context, userName, userPwd string) (bool, error)
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
