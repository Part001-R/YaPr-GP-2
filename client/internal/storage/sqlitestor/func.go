package sqlitestor

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Подключение к БД.
func Connect(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("нет содержимого в аргументе dsn")
	}

	ptrDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: <%w>", err)
	}

	// Установите таймаут для Ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = ptrDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ошибка ping к БД: <%w>", err)
	}

	return ptrDB, nil
}

// Создание таблиц.
func CreateTables(db *sql.DB) error {

	// Создание таблицы пользователей.
	if err := createTableUsers(db); err != nil {
		return fmt.Errorf("функция createTableUsers, вернула ошибку: <%w>", err)
	}

	return nil
}

// Создание таблицы пользователей.
func createTableUsers(db *sql.DB) (err error) {
	// Создание таблицы
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_name TEXT UNIQUE NOT NULL,
        user_password TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    `
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("Ошибка создания таблицы users: <%w>", err)
	}

	// Создание индекса
	query = `
    CREATE INDEX IF NOT EXISTS idx_users_user_name ON users(user_name);
    `
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("Ошибка создания индекса: <%w>", err)
	}

	return nil
}

// Генерация хеша из строки. Возвращается хеш.
//
// Параметры:
//
//	input - входные данные для хеширования.
func generateHash(input string) string {

	// Создание нового хешера
	hasher := sha256.New()

	// Запись строки в хешер
	hasher.Write([]byte(input))

	// Получение хеша в виде байтового среза
	hashBytes := hasher.Sum(nil)

	// Преобразование хеша в строку, в шестнадцатичном представлении.
	return hex.EncodeToString(hashBytes)
}

// Проверка сооответсивя хешей. Возвращается true - соответствие.
//
// Параметры:
//
//	input - входные данные.
//	storedHash - сохранённый хеш.
func isEqualHash(input, storedHash string) bool {
	return generateHash(input) == storedHash
}

// Функция генерирует секретный клю из входной строки. Возвращает секретный клюй ключ.
//
// Параметры:
//
//	input - данные, на основе которых формируется ключ.
func generateSecretKey(input string) string {

	hash := sha256.Sum256([]byte(input)) // генерация 32-х бит

	return hex.EncodeToString(hash[:])
}
