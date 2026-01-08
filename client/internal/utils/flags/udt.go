package flags

const (
	ModeLocal  = "local"                                                 // Режим локальной работы с сервером.
	ModeRemote = "remote"                                                // Работа на сервере.
	DSN        = "file:manager.db?cache=shared&foreign_keys=on&mode=rwc" // Строка подключения к БД.
)
