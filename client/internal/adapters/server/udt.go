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
type dataRequestFile struct {
	fileName    string
	tokenAuth   string
	clientID    string
	sizeReqFile int64
	sizePassed  int64
	secretKey   [32]byte
}

// Представление данных для передачи файла
type dataSendFile struct {
	filePath     string
	tokenAuth    string
	clientID     string
	sizeSendFile int64
	sizePassed   int64
	secretKey    [32]byte
}

// Представление данных для BackUp
type dataBackUp struct {
	listFiles    []string
	tokenAuth    string
	clientID     string
	sizeSendFile int64
	sizePassed   int64
	secretKey    [32]byte
}

// Представление данных для Restore
type dataRestore struct {
	listFiles      []InfoByFiles
	totalSizeFiles int64
	sizePassed     int64
}

// Информация по файлу.
type InfoByFiles struct {
	Name   string // имя файла.
	Volume int64  // размер файла.
}
