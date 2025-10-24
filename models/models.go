package models

import "gorm.io/gorm"

type Checklist struct {
	gorm.Model
	UserID   int64           `gorm:"not null"`
	UserName string          `gorm:"size:255"`
	Title    string          `gorm:"size:255;not null"`
	Tasks    []ChecklistTask `gorm:"foreignkey:ChecklistID;constraint:OnDelete:CASCADE"`
}

type ChecklistTask struct {
	gorm.Model
	ChecklistID uint   `gorm:"not null;index"`
	Order       uint   `gorm:"not null"`
	Name        string `gorm:"size:255;not null"`
	Done        bool   `gorm:"not null;default:false"`

	Checklist Checklist `gorm:"foreignkey:ChecklistID;constraint:OnDelete:CASCADE"`
}

type User struct {
	gorm.Model
	TgUserID uint         `gorm:"not null"`
	Name     string       `gorm:"size:255;not null"`
	Friends  []UserFriend `gorm:"foreignkey:UserID;constraint:OnDelete:CASCADE"`
}

type UserFriend struct {
	gorm.Model
	UserID   uint `gorm:"not null;index"`
	FriendID uint `gorm:"not null;index"`

	User   User `gorm:"foreignkey:UserID;constraint:OnDelete:CASCADE;"`
	Friend User `gorm:"foreignkey:FriendID;constraint:OnDelete:CASCADE;"`
}
