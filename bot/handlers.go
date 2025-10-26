package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"checklist-tg-bot/checklist"
	"checklist-tg-bot/friends"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var pendingChecklists = make(map[int64]bool)
var pendingFriends = make(map[int64]bool)
var pendingTasks = make(map[int64]int64)

func HandleCallback(b *Bot, query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID
	data := query.Data

	switch {
	case data == "create_checklist":
		pendingChecklists[chatID] = true
		ReplyCallback(b, query.ID)
		SendMessage(b, chatID, "Напиши название нового чеклиста:")

	case data == "list_checklist":
		ReplyCallback(b, query.ID)
		checklist.ListChecklists(b.DB, b.API, chatID, userID)

	case strings.HasPrefix(data, "create_task:"):
		idStr := strings.TrimPrefix(data, "create_task:")
		checklistID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Ошибка парсинга ID: %v", err)
			return
		}

		pendingTasks[chatID] = checklistID
		ReplyCallback(b, query.ID)
		SendMessage(b, chatID, "Напишите название задачи:")

	case strings.HasPrefix(data, "list_tasks:"):
		idStr := strings.TrimPrefix(data, "list_tasks:")
		checklistID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Ошибка парсинга ID: %v", err)
			return
		}

		ReplyCallback(b, query.ID)
		checklist.ListTasks(b.DB, b.API, chatID, checklistID)

	case strings.HasPrefix(data, "get_checklist:"):
		idStr := strings.TrimPrefix(data, "get_checklist:")
		_, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Ошибка парсинга id: %v", err)
			return
		}
		ReplyCallback(b, query.ID)
		checklist.GetChecklist(b.DB, b.API, chatID, query.ID, data)

	case data == "get-id":
		ReplyCallback(b, query.ID)
		SendMessage(b, chatID, fmt.Sprintf("Ваш ID: %d", userID))

	case data == "add-friend":
		pendingFriends[chatID] = true
		ReplyCallback(b, query.ID)
		SendMessage(b, chatID, "Введите ID друга")
	}
}

func HandleMessage(b *Bot, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	userName := msg.From.UserName
	text := strings.TrimSpace(msg.Text)

	if pendingChecklists[chatID] {
		checklist.CreateChecklist(b.DB, b.API, chatID, userID, userName, text)
		delete(pendingChecklists, chatID)
		return
	}

	if pendingFriends[chatID] {
		friends.AddFriend(b.DB, b.API, chatID, userID, userName, text)
		delete(pendingFriends, chatID)
		return
	}

	if checklistID, ok := pendingTasks[chatID]; ok {
		checklist.CreateTask(b.DB, b.API, chatID, text, checklistID)
		delete(pendingTasks, chatID)
		return
	}

	switch text {
	case "/start", "/help":
		SendMainKeyboard(b, chatID, userName)
	}
}
