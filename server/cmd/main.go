package main

import (
	"log"

	"github.com/Part001-R/YaPr-GP-2/server/internal/service"
)

func main() {
	// Запуск сервиса.
	if err := service.Run(); err != nil {
		log.Fatal("сервис завершил работу, по причине: <%w>", err)
	}
	log.Println("сервис остановлен штатно")
}
