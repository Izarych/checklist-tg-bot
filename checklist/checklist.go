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

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать задачу", fmt.Sprintf("create_task:%d", checklist.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Посмотреть задачи", fmt.Sprintf("list_tasks:%d", checklist.ID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы выбрали чеклист: %s", checklist.Title))
	msg.ReplyMarkup = keyboard
	_, _ = bot.Send(msg)
}

func CreateTask(db *gorm.DB, bot *tgbotapi.BotAPI, chatID int64, text string, checklistID int64) {

	var count int64
	db.Model(&models.ChecklistTask{}).Where("checklist_id = ?", checklistID).Count(&count)

	count++

	task := models.ChecklistTask{
		ChecklistID: uint(checklistID),
		Name:        text,
		Order:       uint(count),
	}

	if err := db.Create(&task).Error; err != nil {
		log.Printf("Ошибка создания задачи: %v", err)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Произошла ошибка при создании задачи для чеклиста %d", checklistID))
		_, _ = bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы создали задачу %d.%s", task.Order, task.Name))
		_, _ = bot.Send(msg)
	}
}

func ListTasks(db *gorm.DB, bot *tgbotapi.BotAPI, chatID int64, checklistID int64) {
	var tasks []models.ChecklistTask
	var rows [][]tgbotapi.InlineKeyboardButton

	if err := db.Where("checklist_id = ?", checklistID).Find(&tasks).Error; err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при загрузке задач ❌")
		_, _ = bot.Send(msg)
		return
	}

	if len(tasks) == 0 {
		msg := tgbotapi.NewMessage(chatID, "В этом чеклисте пока нет задач 🕐")
		_, _ = bot.Send(msg)
		return
	}

	for _, task := range tasks {
		taskName := fmt.Sprintf("%d.%s", task.Order, task.Name)
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(taskName, fmt.Sprintf("get_task:%d", task.ID)),
		)
		rows = append(rows, row)
	}

	msg := tgbotapi.NewMessage(chatID, "Ваши задачи")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = keyboard

	_, _ = bot.Send(msg)
}
