package main

import (
	"checklist-tg-bot/models"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env файл не найден")
	}
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN не задан!")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	err = db.AutoMigrate(&models.Checklist{}, &models.User{}, &models.UserFriend{})
	if err != nil {
		panic(err)
	}

	log.Println("Подключение к БД успешно!")

	var checklists []models.Checklist

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	bot.Debug = true
	log.Printf("Бот авторизован как: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	pendingChecklists := make(map[int64]bool)
	pendingFriends := make(map[int64]bool)

	for update := range updates {

		if update.CallbackQuery != nil {
			chatID := update.CallbackQuery.Message.Chat.ID
			userId := update.CallbackQuery.From.ID
			data := update.CallbackQuery.Data

			if data == "create_checklist" {
				pendingChecklists[chatID] = true
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Хорошо, напиши название чеклиста")
				if _, err := bot.Request(callback); err != nil {
					log.Printf("Ошибка при ответе на callback: %v", err)
				}
				msg := tgbotapi.NewMessage(chatID, "Напиши название нового чеклиста:")
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Ошибка при отправке сообщения: %v", err)
				}
			}
			if data == "list_checklist" {
				var rows [][]tgbotapi.InlineKeyboardButton

				db.Select("id, title").Where("user_id = ?", userId).Find(&checklists)
				for _, c := range checklists {
					row := tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(c.Title, fmt.Sprintf("get_checklist:%d", c.ID)),
					)
					rows = append(rows, row)
				}

				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Ваши чеклисты")
				if _, err := bot.Request(callback); err != nil {
					log.Printf("Ошибка при ответе на callback: %v", err)
				}

				msg := tgbotapi.NewMessage(chatID, "Ваши чеклисты")
				keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
				msg.ReplyMarkup = keyboard

				if _, err := bot.Send(msg); err != nil {
					log.Printf("Ошибка при отправке сообщения %v", err)
				}
			}

			if strings.HasPrefix(data, "get_checklist:") {
				idStr := strings.TrimPrefix(data, "get_checklist:")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					log.Printf("Ошибка парсинга id: %v", err)
					return
				}

				var checklist models.Checklist
				db.First(&checklist, id)

				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, fmt.Sprintf("Вы выбрали чеклист: %s", checklist.Title))
				if _, err := bot.Request(callback); err != nil {
					log.Printf("Ошибка при ответе на callback: %v", err)
				}

				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Вы выбрали чеклист: %s", checklist.Title))
				_, _ = bot.Send(msg)
			}

			if data == "get-id" {
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, fmt.Sprintf("Ваш ID: %d", userId))
				if _, err := bot.Request(callback); err != nil {
					log.Printf("Ошибка при ответе на callback: %v", err)
				}
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ваш ID: %d", userId))
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Ошибка при отправке сообщения: %v", err)
				}
			}

			if data == "add-friend" {
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Введите ID друга")
				if _, err := bot.Request(callback); err != nil {
					log.Printf("Ошибка при ответе на callback: %v", err)
				}
				msg := tgbotapi.NewMessage(chatID, "Введите ID друга")
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Ошибка при отправке сообщения: %v", err)
				}

				pendingFriends[chatID] = true
			}

			continue
		}

		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		userId := update.Message.From.ID
		userName := update.Message.From.UserName
		text := strings.TrimSpace(update.Message.Text)

		if pendingChecklists[chatID] {
			checklist := models.Checklist{
				UserID:   userId,
				UserName: userName,
				Title:    text,
			}
			if err := db.Create(&checklist).Error; err != nil {
				log.Printf("Ошибка вставки чеклиста: %v", err)
				msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при создании чеклиста")
				_, _ = bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Чеклист '%s' создан ✅", text))
				_, _ = bot.Send(msg)
			}
			delete(pendingChecklists, chatID)
			continue
		}

		if pendingFriends[chatID] {
			friendID, err := strconv.ParseUint(text, 10, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Введите корректный числовой ID друга.")
				_, _ = bot.Send(msg)
				continue
			}

			var friend models.User
			result := db.Where("tg_user_id = ?", friendID).First(&friend)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				msg := tgbotapi.NewMessage(chatID, "Такой пользователь еще не был в боте ❌")
				_, _ = bot.Send(msg)
				continue
			}

			var currentUser models.User
			db.FirstOrCreate(&currentUser, models.User{TgUserID: uint(userId), Name: userName})

			userFriend := models.UserFriend{
				UserID:   currentUser.ID,
				FriendID: friend.ID,
			}

			if err := db.Create(&userFriend).Error; err != nil {
				log.Printf("Ошибка при добавлении друга: %v", err)
				msg := tgbotapi.NewMessage(chatID, "Ошибка при добавлении друга.")
				_, _ = bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Пользователь %s успешно добавлен в друзья ✅", friend.Name))
				_, _ = bot.Send(msg)
			}

			delete(pendingFriends, chatID)
			continue
		}

		switch text {
		case "/start":
			msg := tgbotapi.NewMessage(chatID, "Привет! Нажми на кнопку ниже, чтобы создать чеклист.")
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
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения %v", err)
			}
			continue

		case "/help":
			msg := tgbotapi.NewMessage(chatID, "Я могу создавать чеклисты")
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
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Ошибка при отправке сообщения %v", err)
			}
			continue
		}
	}
}
