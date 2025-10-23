package bot

import (
	"checklist-tg-bot/models"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func SendMainKeyboard(b *Bot, chatID int64, userName string) {
	createUser(b.DB, chatID, userName)
	msg := tgbotapi.NewMessage(chatID, "Привет! Нажми на кнопку ниже:")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать чеклист", "create_checklist"),
			tgbotapi.NewInlineKeyboardButtonData("Все чеклисты", "list_checklist"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Получить свой ID", "get-id"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить друга", "add-friend"),
		),
	)

	msg.ReplyMarkup = keyboard
	if _, err := b.API.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func SendMessage(b *Bot, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.API.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func ReplyCallback(b *Bot, callbackID string) {
	if _, err := b.API.Request(tgbotapi.NewCallback(callbackID, "")); err != nil {
		log.Printf("Ошибка при ответе на callback: %v", err)
	}
}

func createUser(db *gorm.DB, userID int64, userName string) {
	var user models.User

	if err := db.FirstOrCreate(&user, models.User{
		TgUserID: uint(userID),
		Name:     userName,
	}).Error; err != nil {
		log.Printf("Ошибка при создании пользователя: %v", err)
	}
}
