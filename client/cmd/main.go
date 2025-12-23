package main

import (
	"log"

	"github.com/Part001-R/YaPr-GP-2/client/internal/service"
)

func main() {

	if err := service.Run(); err != nil {
		log.Fatal("работа сервиса прервана по ошибке: <%w>", err)
	}
	log.Println("сервис остановлен штатно")
}
