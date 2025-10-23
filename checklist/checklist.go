package checklist

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"

	"checklist-tg-bot/models"
)

func CreateChecklist(db *gorm.DB, bot *tgbotapi.BotAPI, chatID int64, userID int64, userName string, title string) {
	checklist := models.Checklist{
		UserID:   userID,
		UserName: userName,
		Title:    title,
	}

	if err := db.Create(&checklist).Error; err != nil {
		log.Printf("Ошибка вставки чеклиста: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при создании чеклиста")
		_, _ = bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Чеклист '%s' создан ✅", title))
		_, _ = bot.Send(msg)
	}
}

func ListChecklists(db *gorm.DB, bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	var checklists []models.Checklist
	db.Select("id, title").Where("user_id = ?", userID).Find(&checklists)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, c := range checklists {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(c.Title, fmt.Sprintf("get_checklist:%d", c.ID)),
		)
		rows = append(rows, row)
	}

	msg := tgbotapi.NewMessage(chatID, "Ваши чеклисты")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения %v", err)
	}
}

func GetChecklist(db *gorm.DB, bot *tgbotapi.BotAPI, chatID int64, callbackID string, data string) {
	idStr := strings.TrimPrefix(data, "get_checklist:")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Ошибка парсинга id: %v", err)
		return
	}

	var checklist models.Checklist
	db.First(&checklist, id)

	callback := tgbotapi.NewCallback(callbackID, fmt.Sprintf("Вы выбрали чеклист: %s", checklist.Title))
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Ошибка при ответе на callback: %v", err)
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы выбрали чеклист: %s", checklist.Title))
	_, _ = bot.Send(msg)
}
