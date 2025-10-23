package bot

import (
	"log"

	"gorm.io/gorm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API *tgbotapi.BotAPI
	DB  *gorm.DB
}

func NewBot(db *gorm.DB) (*Bot, error) {
	token := ""

	botAPI, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		return nil, err
	}

	botAPI.Debug = true
	log.Printf("Бот авторизован как: %s", botAPI.Self.UserName)

	return &Bot{
		API: botAPI,
		DB:  db,
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.API.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			HandleCallback(b, update.CallbackQuery)
			continue
		}
		if update.Message != nil {
			HandleMessage(b, update.Message)
		}
	}
}
