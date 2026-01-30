// Вспомогательные функции пакета.
package sqlitestor

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

// Подключение к БД. Возвращается указатель на БД и ошибка.
//
// Параметры:
//
//	dsn - dsn БД.
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

// Генерация хэша из строки. Возвращается хэш.
//
// Параметры:
//
//	input - входные данные для хэширования.
func generateHash(input string) string {

	hasher := sha256.New()
	hasher.Write([]byte(input))

	// Получение хэша в виде байтового среза
	hashBytes := hasher.Sum(nil)

	// Преобразование хэша в строку, в шестнадцатичном представлении.
	return hex.EncodeToString(hashBytes)
}

// Проверка сооответсивя хэшей. Возвращается true - соответствие.
//
// Параметры:
//
//	input - входные данные.
//	storedHash - сохранённый хеш.
func isEqualHash(input, storedHash string) bool {
	return generateHash(input) == storedHash
}

// Реализация Up миграции. Возвращается ошибка.
//
// Параметры:
//
//	db - указательна БД.
func migrationUp(db *sql.DB) error {

	// Проверка аргументов.
	if db == nil {
		return NilPtrArgumentDB
	}

	// Подготовка к миграции.
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("ошибка создания драйвера: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "sqlite", driver)
	if err != nil {
		return fmt.Errorf("ошибка создания экземпляра миграции: %v", err)
	}
	// Up миграция
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		} else {
			return fmt.Errorf("ошибка выполнения Up миграции: %v", err)
		}
	}

	return nil
}
