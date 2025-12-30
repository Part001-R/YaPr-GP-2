package sqlitestor

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Реализация Up миграции.
func migrationUp(db *sql.DB) error {

	// Проверка аргументов.
	if db == nil {
		return errors.New("в аргументе db, нет указателя")
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
