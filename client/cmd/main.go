package main

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/Part001-R/YaPr-GP-2/client/internal/service"
	"github.com/Part001-R/YaPr-GP-2/client/internal/service/udt"
)

var (
	buildVersion string // Версия сборки.
	buildDate    string // Дата сборки.
)

func main() {

	// Перехват паники
	defer func() {
		if r := recover(); r != nil {

			stackTrace := debug.Stack()
			fmt.Printf("Паника приложения. Причина: <%v>. Стек: <%s>\n", r, stackTrace)
		}
	}()

	// Прлучение информации по сборке.
	//
	// Пример использования:
	// go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date +%Y-%m-%d)" -o myapp
	udt.UpdateBuildInfo(buildVersion, buildDate)

	// Запуск сервиса.
	if err := service.Run(); err != nil {
		log.Fatal("работа сервиса прервана по ошибке: <%w>", err)
	}
	log.Println("сервис остановлен штатно")
}
