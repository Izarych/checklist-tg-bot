package models

import "gorm.io/gorm"

type Checklist struct {
	gorm.Model
	UserID   int64  `gorm:"not null"`
	UserName string `gorm:"size:255"`
	Title    string `gorm:"size:255;not null"`
}
