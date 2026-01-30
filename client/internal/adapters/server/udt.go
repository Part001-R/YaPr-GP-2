// Типы данных пакета.
package server

// Представление данных передачи логин/пароль.
type TxLoginPassword struct {
	ID        string // id клиента
	For       string // принадлежность логин/пароль
	Login     string // логин
	Password  string // пароль
	CreatedAt string // дата создания
}

// Представление данных приёма логин/пароль.
type RxLoginPassword struct {
	For       string // принадлежность логин/пароль
	Login     string // логин
	Password  string // пароль
	CreatedAt string // дата создания
}

// Представление данных приёма текста.
type RxText struct {
	For       string // принадлежность текста
	Text      string // текст
	CreatedAt string // дата создания
}

// Представление данных приёма банковской карты.
type RxBankCard struct {
	For       string // принадлежность текста
	Owner     string // владелец
	Numb      string // номер
	Valid     string // валидность
	Code      string // код
	CreatedAt string // дата создания
}

// Представление данных текста.
type TxText struct {
	ID        string // id клиента
	For       string // принадлежность текста
	Text      string // текст
	CreatedAt string // дата создания
}

// Представление данных банковской карты.
type TxBankCard struct {
	ID        string // id клиента
	For       string // принадлежность текста
	Owner     string // владелец
	Numb      string // номер
	ValidData string // валидность
	Code      string // код
	CreatedAt string // дата создания
}

// Передача файла.
type TxFile struct {
	ID       string // id клиента
	FileName string // название файла
	Bytes    []byte // данные файла
}

// Ответ на передачу данных файла (TxFile) .
type TxFileResp struct {
	FileName string // название файла
}

// Представление данных для запроса файла
type DataRequestFile struct {
	FilePath    string   // Путь к файлу.
	FileName    string   // имя файла.
	TokenAuth   string   // Токен аутентификации.
	ClientID    string   // id клиента.
	SizeReqFile int64    // Размер запрашиваемых данных.
	SizePassed  int64    // Обработанный размер.
	SecretKey   [32]byte // Ключ.
}

// Представление данных для передачи файла
type dataSendFile struct {
	filePath     string   // Путь к файлу.
	tokenAuth    string   // Токен аутентификации.
	clientID     string   // id клиента.
	sizeSendFile int64    // Размер передаваемых данных.
	sizePassed   int64    // Обработанный размер.
	secretKey    [32]byte // Ключ.
}

// Представление данных для BackUp
type dataBackUp struct {
	listFiles    []string // Имена файлов.
	tokenAuth    string   // Токен аутентификации.
	clientID     string   // id клиента.
	sizeSendFile int64    // Размер передаваемых данных.
	sizePassed   int64    // Обработанный размер.
	secretKey    [32]byte // Ключ.
}

// Представление данных для Restore
type dataRestore struct {
	listFiles      []InfoByFiles // Имена файлов.
	totalSizeFiles int64         // Общий размер файлов.
	sizePassed     int64         // Обработанный размер.
}

// Информация по файлу.
type InfoByFiles struct {
	Name   string // имя файла.
	Volume int64  // размер файла.
}
