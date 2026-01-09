package server

// Представление данных передачи логин/пароль.
type TxLoginPassword struct {
	TxID        string // id клиента
	TxFor       string // принадлежность логин/пароль
	TxLogin     string // логин
	TxPassword  string // пароль
	TxCreatedAt string // дата создания
}

// Представление данных текста.
type TxText struct {
	TxID        string // id клиента
	TxFor       string // принадлежность текста
	TxText      string // текст
	TxCreatedAt string // дата создания
}
