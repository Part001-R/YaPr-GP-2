// Вспомогательные функции пакета.
package flags

import "fmt"

// Проверка значений влагов. Возвращается ошибка.
func checkFlags(f Config) error {

	// Проверка режима работы.
	if f.Mode != ModeLocal && f.Mode != ModeRemote {
		return fmt.Errorf("Проверьте режим запуска. Указан: <%s>", f.Mode)
	}

	// Проверка DSN
	if f.Mode == ModeLocal {
		if f.DSN == "" {
			return ErrEmptyDSN
		}
	}

	return nil
}
