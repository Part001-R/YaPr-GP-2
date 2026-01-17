// Создание экземпляра пакета.
package udt

// Подготовительные действия, перед запуском сервиса.
import (
	"sync"

	"github.com/Part001-R/YaPr-GP-2/server/internal/domain"
	"github.com/Part001-R/YaPr-GP-2/server/internal/grpc"
	"github.com/Part001-R/YaPr-GP-2/server/internal/utils/flags"
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
	Lgr     *zap.Logger    // Указатель на логгер.
	Srv     *grpc.Manager  // Указатель на экземпляр grpc.
	TLS     TLSdata        // Ключи реалзизации GRPCS.
	Storage domain.DomainI // БД
	Flag    *flags.Config  // Флаги.
}

// Указатель на конфигурацию сервиса.
var confInst *Configuration

// Конструктор. Возвращается указатель на экземпляр сервиса.
//
// Параметры:
//
//	l - логгер.
//	g - указатель на grpc.
//	keyPublic - публичный ключ.
//	keyPrivate - приватный ключ.
//	storage - указатель на домен.
//	flag - указатель на флаги.
func New(l *zap.Logger, g *grpc.Manager, keyPublic, keyPrivate string, storage domain.DomainI, flag *flags.Config) *Configuration {
	onceConf.Do(func() {
		confInst = &Configuration{
			Lgr: l,
			Srv: g,
			TLS: TLSdata{
				Public: keyPublic,
				Privae: keyPrivate,
			},
			Storage: storage,
			Flag:    flag,
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
