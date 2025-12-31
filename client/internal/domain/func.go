// Вспомогательные функции пакета.
package domain

import "strings"

// Извлечение префикса из строки dsn. Возвращается префикс.
//
// Параметры:
//
//	input - строка для обработки.
func extractPrefixDSN(input string) string {

	parts := strings.Split(input, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
