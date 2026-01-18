// Главный пакет сервиса.
package main

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/Part001-R/YaPr-GP-2/server/internal/service"
)

func main() {

	// Перехват паники
	defer func() {
		if r := recover(); r != nil {

			stackTrace := debug.Stack()
			fmt.Printf("Паника приложения. Причина: <%v>. Стек: <%s>\n", r, stackTrace)
		}
	}()

	// Запуск сервиса.
	if err := service.Run(); err != nil {
		log.Fatal("сервис завершил работу, по причине: <%w>", err)
	}
	log.Println("сервис остановлен штатно")
}
