package models

import "time"

type DirectMessage struct {
	ID         uint `gorm:"primaryKey"`
	SenderID   uint `gorm:"index"`
	ReceiverID uint `gorm:"index"`
	Content    string
	CreatedAt  time.Time
}
