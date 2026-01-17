// Вспомогательные функции пакета.
package sqlitestor

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Подключение к БД. Возвращается указатель на БД и ошибка.
//
// Параметры:
//
//	dsn - строка подключения к БД.
func connect(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, EmptyDataArgumentDSN
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

// Генерация хеша из строки. Возвращается хэш.
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
