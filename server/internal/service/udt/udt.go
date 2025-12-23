package udt

// Подготовительные действия, перед запуском сервиса.
import (
	"sync"

	"github.com/Part001-R/YaPr-GP-2/server/internal/grpc"
	"go.uber.org/zap"
)

// Обеспечение единоразового выполняения для конфигурации сервиса.
var onceConf sync.Once

// Ключи шифрования
type TLSdata struct {
	Public string // Публичный ключ
	Privae string // Приватный ключ
}

// Конфигурация сервиса.
type Configuration struct {
	PtrLogger *zap.Logger           // Указатель на логгер.
	PtrGRPC   *grpc.PasswordManager // Указатель на экземпляр grpc.
	TLS       TLSdata               // Ключи реалзизации GRPCS.
}

// Указатель на конфигурацию сервиса.
var confInst *Configuration

// Создание экземпляра конфигурации сервиса.
func New(l *zap.Logger, g *grpc.PasswordManager, keyPublic, keyPrivate string) *Configuration {
	onceConf.Do(func() {
		confInst = &Configuration{
			PtrLogger: l,
			PtrGRPC:   g,
			TLS: TLSdata{
				Public: keyPublic,
				Privae: keyPrivate,
			},
		}
	})
	return confInst
}

// Проверка содержимого конфигурации. Возвращается ошибка.
func (c Configuration) CheckConf() error {

	if c.PtrLogger == nil {
		return NilPtrArgumentConf
	}
	if c.PtrGRPC == nil {

	}
	return nil
}
