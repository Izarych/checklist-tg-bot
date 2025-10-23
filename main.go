package main

import (
	"log"

	"checklist-tg-bot/bot"
	"checklist-tg-bot/db"
)

func main() {
	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	log.Println("Подключение к БД успешно!")

	botInstance, err := bot.NewBot(dbConn)
	if err != nil {
		log.Fatalf("Ошибка запуска бота: %v", err)
	}

	botInstance.Start()
}
