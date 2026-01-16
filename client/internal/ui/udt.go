package ui

import (
	"sync"

	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
)

// Представление приянтых данных логин/пароль
type rxLoginPassword struct {
	name      string // имя записи
	login     string // логин
	password  string // пароль
	createdAt string // дата создания
}

type encrKey struct {
	secretKey [32]byte // секретный ключ
}

// Введённые пользователем данные.
type typeData struct {
	login         string // имя пользователя.
	password1     string // пароль.
	password2     string // пароль (подтверждение).
	ip            string // IP.
	port          string // Port.
	dataFor       string // для чего формируются данные.
	dataLogin     string // логин.
	dataPassword  string // пароль.
	dataText      string // текст.
	dataOwner     string // владелец.
	dataNumb      string // номер.
	dataValidDate string // дата валидности.
	dataCode      string // код.
	dataPathSrc   string // путь, нахождения файла.
	dataPathTrg   string // путь, куда нужно поместить файл.
}

// Признаки выполнения логики
type status struct {
	checkConnectStatus           bool // Результат процедуры проверки связи с сервером.
	checkConnectPassed           bool // Признак, что проверка связи была запущена.
	addUserSUCCESS               bool // Признак успешного добавления пользователя.
	addUserPassed                bool // Признак, что была запущена процедура регистрации пользователя.
	addUserRegBusy               bool // Признак, чот уже есть заргистрированный пользователь
	addLoginPaaswordSUCCESS      bool // Признак успешного добавления пары логин/пароль.
	addLoginPaaswordPassed       bool // Признак, что выполнена процедура добавления пары логин/пароль.
	readLoginPaaswordSUCCESS     bool // Признак, успешного получения данных логин/пароль.
	readLoginPaaswordPassed      bool // Признак, что процедура чтения логин/пароль, пройдена.
	readNameLoginPaaswordSUCCESS bool // Признак, успешного получения имён логин/пароль.
	readNameLoginPaaswordPassed  bool // Признак, что процедура получения имён логин/пароль, пройдена.
	readNameTextSUCCESS          bool // Признак, успешного получения имён текста.
	readNameTextPassed           bool // Признак, что процедура получения имён текста, пройдена.
	delLoginPaaswordSUCCESS      bool // Признак, успешного удаления данных логин/пароль.
	delLoginPaaswordPassed       bool // Признак, что процедура удаления логин/пароль, пройдена.
	addTextSUCCESS               bool // Признак успешного добавления текста.
	addTextPassed                bool // Признак, что выполнена процедура добавления текста.
	readTextSUCCESS              bool // Признак, успешного получения данных текста.
	readTextPassed               bool // Признак, что процедура получения текста, пройдена.
	delTextSUCCESS               bool // Признак, успешного удаления данных текста.
	delTextPassed                bool // Признак, что процедура удаления текста, пройдена.
	addBankCardSUCCESS           bool // Признак успешного добавления карты.
	addBankCardPassed            bool // Признак, что выполнена процедура добавления карты.
	readBankCardSUCCESS          bool // Признак, успешного получения данных карт.
	readBankCardPassed           bool // Признак, что процедура получения данных карт, пройдена.
	readNameBankCardSUCCESS      bool // Признак, успешного получения данных карт.
	readNameBankCardPassed       bool // Признак, что процедура получения данных карт, пройдена.
	readNameFileSUCCESS          bool // Признак, успешного получения данных файлов.
	readNameFilePassed           bool // Признак, что процедура получения данных файлов, пройдена.
	delBankCardSUCCESS           bool // Признак, успешного удаления данных карты.
	delBankCardPassed            bool // Признак, что процедура удаления карты, пройдена.
	addFileSUCCESS               bool // Признак успешного добавления файла.
	addFilePassed                bool // Признак, что выполнена процедура добавления файла.
	readFileSUCCESS              bool // Признак, успешного получения данных файла.
	readFilePassed               bool // Признак, что процедура получения данных файла, пройдена.
	delFileSUCCESS               bool // Признак, успешного удаления файла.
	delFilePassed                bool // Признак, что процедура удаления файла, пройдена.
	extractFileSUCCESS           bool // Признак, успешного извлечения файла.
	extractFilePassed            bool // Признак, что процедура извлечения файла, пройдена.
	restore                      int  // Статус процесса воостановления из резервной копии.
	backUp                       int  // Статус процесса создания резервной копии.
	pushContainer                int  // Статус процесса передачи в контейнер.
	popContainer                 int  // Статус процесса извлечения из контейнера.
	fileTx                       int  // Статус процесса передачи файла на сервер.
	fileRx                       int  // Статус процесса приёма файла от сервера.
	isBusyServer                 bool // Признак занятости сервера.
	statusWDT                    bool // Статус WDT. true - активен.
}

// Для навигации по экранам.
type screens struct {
	activeView   string // название активного экрана
	currentFocus string // на какой элемент установлен фокус
}

// Представление записи логин/пароль
type loginPassword struct {
	name      string
	login     string
	password  string
	createdAt string
}

// Представление записи - текст.
type textData struct {
	name      string
	text      string
	createdAt string
}

// Представление записи - банковская карта.
type bankCard struct {
	name      string
	owner     string
	numb      string
	valid     string
	code      string
	createdAt string
}

// Данные БД.
type data struct {
	encryptLoginPassword []loginPassword // закодированные данные - логин/пароль.
	loginPassword        []loginPassword // данные - логин/пароль.
	encryptTextData      []textData      // закодированные данные - текст.
	textData             []textData      // данные - текст.
	encryptBankCard      []bankCard      // закодированные данные - банковские карты.
	bankCard             []bankCard      // данные - банковские карты.
	files                []string        // файлы
	namesLoginPassword   []string        // имена записей логин/пароль
	namesText            []string        // имена записей текст
	namesBankCard        []string        // имена записей банковские карты
	namesFile            []string        // имена файлов
}

// Индесы.
type indexes struct {
	loginPassword int // текущий индекс для обхода массива - логин/пароль.
	text          int // текущий индекс для обхода массива - текст.
	bankCard      int // текущий индекс для обхода массива - банковские карты.
	file          int // текущий индекс для обхода массива - файлы.
}

// Мьютексы.
type mutex struct {
	restoreBackup       sync.Mutex // для резервного копирования и восстановления.
	processTxRx         sync.Mutex // для данных процесса Tx Rx файлов.
	statusBackUp        sync.Mutex // для статуса процесса передачи.
	statusRestore       sync.Mutex // для статуса процесса приёма.
	statusPushContainer sync.Mutex // для статуса процесса передачи в контейнер.
	statusPopContainer  sync.Mutex // для статуса процесса извлечения из контейнера.
	statusFileTx        sync.Mutex // для статуса процесса передачи файла на сервер.
	statusFileRx        sync.Mutex // для статуса процесса приёма файла от сервера.
	statusIsBusyServer  sync.Mutex // для статуса занятости сервера.
}

// Отправка-приём файлов.
type txrx struct {
	percentTxRx float32 // Процент выполнения процесса передачи файлов.
	totalSizeKB int64   // Передаваемый размер (Байт).
	passedKB    int64   // обработано данных (Байт).
}

// Каналы.
type ch struct {
	resetWDT      chan struct{} // Сброс сторожевого таймера.
	resetWDTClose chan struct{} // Завершение работы WDT.
}

// Общий тип для CLI UI.
type handlerUI struct {
	conf       *udt.Configuration // конфигурация сервиса.
	typed      typeData           // введённые пользователем данные.
	status     status             // признаки сервиса.
	view       screens            // взаимодействие с окнами.
	secret     encrKey            // секретность.
	data       data               // данные.
	index      indexes            // индексы для обхода массивов.
	mutex      mutex              // мьютексы.
	txrx       txrx               // данные по Tx-Rx файлов.
	clientName string             // имя клиента.
	tokenAuth  string             // токен аутентификации.
	ch         ch                 // каналы.
}
