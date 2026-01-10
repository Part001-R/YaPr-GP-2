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

// Представление данных банковской карты.
type TxBankCard struct {
	TxID        string // id клиента
	TxFor       string // принадлежность текста
	TxOwner     string // владелец
	TxNumb      string // номер
	TxValidData string // валидность
	TxCode      string // код
	TxCreatedAt string // дата создания
}

// Передача файла.
type TxFile struct {
	TxID       string // id клиента
	TxFileName string // название файла
	TxBytes    []byte // данные файла
}

// Ответ на передачу данных файла (TxFile) .
type TxFileResp struct {
	TxFileName string // название файла
}
