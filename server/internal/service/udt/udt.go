package udt

// Подготовительные действия, перед запуском сервиса.
import (
	"sync"

	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
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
	Lgr     *zap.Logger     // Указатель на логгер.
	Srv     *grpc.Manager   // Указатель на экземпляр grpc.
	TLS     TLSdata         // Ключи реалзизации GRPCS.
	Storage domain.StorageI // БД
}

// Указатель на конфигурацию сервиса.
var confInst *Configuration

// Конструктор.
func New(l *zap.Logger, g *grpc.Manager, keyPublic, keyPrivate string, storage domain.StorageI) *Configuration {
	onceConf.Do(func() {
		confInst = &Configuration{
			Lgr: l,
			Srv: g,
			TLS: TLSdata{
				Public: keyPublic,
				Privae: keyPrivate,
			},
			Storage: storage,
		}
	})
	return confInst
}

// Проверка содержимого конфигурации. Возвращается ошибка.
func (c Configuration) CheckConf() error {

	if c.Lgr == nil {
		return NilPtrArgumentConf
	}
	if c.Srv == nil {

	}
	return nil
}
