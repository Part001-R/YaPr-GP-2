// Типы данных и константы.
package flags

const (
	ModeLocal          = "local"                                                      // Режим локальной работы с сервером.
	ModeRemote         = "remote"                                                     // Работа на сервере.
	DSN                = "file:localStorage.db?cache=shared&foreign_keys=on&mode=rwc" // Строка подключения к БД.
	LocalNameDB        = "localDB.db"                                                 // Имя локальной БД
	LocalNameContainer = "localContainer.data"                                        // Имя локального контейнера
)
