// Типы данных пакета.
package grpc

import (
	pb "github.com/Part001-R/YaPr-GP-2/proto"
	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/server/internal/utils/flags"
	"go.uber.org/zap"
)

const (
	stageNotActive = 0
	stageActive    = 1
)

// Статусы процессов.
type statusSrv struct {
	backUp  int32 // Признак активности процесса BackUp (клиент передаёт данные).
	restore int32 // Признак активности процесса Restore (клиент принимает данные).
	rxFile  int32 // Признак активности процесса приёма файла от клиента.
	txFile  int32 // Признак активности процесса передаче клиенту файла.
}

// Принятые данные регистрации.
type registrationRX struct {
	userName      string // Имя пользователя.
	userPwd       string // Пароль пользователя.
	userPwdRepeat string // Подтверждение пароля.
}

// Принятые данные аутентификации.
type authenticationRX struct {
	userName string // Имя пользователя.
	userPwd  string // Пароль пользователя.
}

// Представление токена.
type tokenData struct {
	name  string // имя
	token string // токен
}

// Принятые данные логин/пароль.
type RxLoginPassword struct {
	ID        string // id клиента
	For       string // принадлежность логин/пароль
	Login     string // логин
	Password  string // пароль
	CreatedAt string // дата создания
}

// Передаваемые данные логин/пароль.
type TxLoginPassword struct {
	For       string // принадлежность логин/пароль
	Login     string // логин
	Password  string // пароль
	CreatedAt string // дата создания
}

// Передаваемые данные текста.
type TxText struct {
	For       string // принадлежность текста
	Text      string // текст
	CreatedAt string // дата создания
}

// Передаваемые данные банковской карты.
type TxBankCard struct {
	For       string // принадлежность текста
	Owner     string // владелец
	Numb      string // номер
	Valid     string // валидность
	Code      string // код
	CreatedAt string // дата создания
}

// Принятые данные текста.
type RxText struct {
	ID        string // id клиента
	For       string // принадлежность текста
	Text      string // текст
	CreatedAt string // дата создания
}

// Принятые данные банковской карты.
type RxBankCard struct {
	ID        string // id клиента
	For       string // принадлежность текста
	Owner     string // владелец
	Numb      string // номер
	ValidData string // валидность
	Code      string // код
	CreatedAt string // дата создания
}

// Приянтые данные для запроса логин/пароль по имени записи.
type RxReqLoginPasswordByName struct {
	ClientID string // id клиента
	Name     string // имя записи логин/пароль
}

// Приянтые данные для запроса текста по имени записи.
type RxReqTextByName struct {
	ClientID string // id клиента
	Name     string // имя записи текста
}

// Приянтые данные для запроса банковской карты по имени записи.
type RxReqBankCardByName struct {
	ClientID string // id клиента
	Name     string // имя записи банковской карты
}

// Информация по файлу.
type fileInfo struct {
	name string // имя
	hash string // хэш
	size int64  // размер
}

// ==============================

// Представленние сервиса.
type TestManager struct {
	pb.UnimplementedPasswordManagerServer
	logger    *zap.Logger    // Логгер.
	status    statusSrv      // Статусы.
	storage   domain.DomainI // База данных.
	token     string         // Выданный токен
	secretKey string         // Секретный ключ
	flag      *flags.Config  // Флаги
}

// Ключи шифрования
type TLSdata struct {
	Public string // Публичный ключ
	Privae string // Приватный ключ
}

// Конфигурация сервиса.
type Configuration struct {
	Lgr     *zap.Logger    // Указатель на логгер.
	Srv     *Manager       // Указатель на экземпляр grpc.
	TLS     TLSdata        // Ключи реалзизации GRPCS.
	Storage domain.DomainI // БД
	Flag    *flags.Config  // Флаги.
}
