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

	// создание таблицы для хранения логин/пароль.
	if err := createTableLoginPassword(db); err != nil {
		return fmt.Errorf("функция createTableLoginPassword, вернула ошибку: <%w>", err)
	}

	return nil
}

// Создание таблицы пользователей.
func createTableUsers(db *sql.DB) (err error) {

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

	return nil
}

// Создание таблицы для хранения логи/пароль.
func createTableLoginPassword(db *sql.DB) (err error) {

	// В запросе маскируется принадлежность данных.
	//
	// field_1 - наименование ресурса, к которому сопоставляется логин/пароль.
	// field_2 - логин.
	// field_3 - пароль.
	query := `
    CREATE TABLE IF NOT EXISTS data1 (  
        field_1 TEXT UNIQUE NOT NULL,
        field_2 TEXT NOT NULL,
		field_3 TEXT NOT NULL,
        created_at DATETIME NOT NULL
    );
    `
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("Ошибка создания таблицы data1: <%w>", err)
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

// Получение количества записей в таблице логин/пароль (data1). Возвращается количество записей и ошибка.
//
// Параметры:
//
//	ctx - контекст.
//	db - указатель на БД.
func getRecordLoginPwdContext(ctx context.Context, db *sql.DB) (int, error) {

	// Подготовка запроса.
	query := "SELECT COUNT(*) FROM data1"

	// Запрос.
	var count int

	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("функция db.QueryRowContext, вернула ошибку: <%w>", err)
	}

	// Результат.
	return count, nil
}
