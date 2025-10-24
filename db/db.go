package db

import (
	"fmt"
	"log"
	"os"

	"checklist-tg-bot/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Println(".env файл не найден")
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
		return nil, err
	}

	if err := db.AutoMigrate(&models.Checklist{}, &models.User{}, &models.UserFriend{}, &models.ChecklistTask{}); err != nil {
		return nil, err
	}

	return db, nil
}
