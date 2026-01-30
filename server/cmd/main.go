// Главный пакет сервиса.
package main

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/Part001-R/YaPr-GP-2/server/internal/service"
)

var (
	buildVersion string
	buildDate    string
)

func main() {

	// Перехват паники
	defer func() {
		if r := recover(); r != nil {

			stackTrace := debug.Stack()
			fmt.Printf("Паника приложения. Причина: <%v>. Стек: <%s>\n", r, stackTrace)
		}
	}()

	// Вывод информации о сборке.
	//
	// Пример использования:
	// go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date +%Y-%m-%d)" -o myapp
	log.Printf("Build version: %s", service.GetValueOrDefault(buildVersion))
	log.Printf("Build date: %s", service.GetValueOrDefault(buildDate))

	// Запуск сервиса.
	if err := service.Run(); err != nil {
		log.Fatal("сервис завершил работу, по причине: <%w>", err)
	}
}
