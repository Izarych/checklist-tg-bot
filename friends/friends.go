package friends

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"checklist-tg-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func AddFriend(db *gorm.DB, bot *tgbotapi.BotAPI, chatID int64, userID int64, userName string, text string) {
	friendID, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		sendMessage(bot, chatID, "Введите корректный числовой ID друга.", nil)
		return
	}

	if uint64(userID) == friendID {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Добавить друга", "add-friend"),
			),
		)

		sendMessage(bot, chatID, "Нельзя добавить в друзья самого себя ❌", &keyboard)
		return
	}

	var friend models.User
	result := db.Where("tg_user_id = ?", friendID).First(&friend)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		sendMessage(bot, chatID, "Такой пользователь еще не был в боте ❌", nil)
		return
	}

	var currentUser models.User
	db.FirstOrCreate(&currentUser, models.User{TgUserID: uint(userID), Name: userName})

	var existing models.UserFriend
	if err := db.Where("user_id = ? AND friend_id = ?", currentUser.ID, friend.ID).First(&existing).Error; err == nil {
		sendMessage(bot, chatID, fmt.Sprintf("Пользователь %s уже есть в друзьях ✅", friend.Name), nil)
		return
	}

	userFriend := models.UserFriend{
		UserID:   currentUser.ID,
		FriendID: friend.ID,
	}

	if err := db.Create(&userFriend).Error; err != nil {
		log.Printf("Ошибка при добавлении друга: %v", err)
		sendMessage(bot, chatID, "Ошибка при добавлении друга.", nil)
		return
	}

	sendMessage(bot, chatID, fmt.Sprintf("Пользователь %s успешно добавлен в друзья ✅", friend.Name), nil)

	notifyText := fmt.Sprintf("👋 Вас добавил в друзья пользователь %s!", currentUser.Name)
	sendMessage(bot, int64(friend.TgUserID), notifyText, nil)
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}
