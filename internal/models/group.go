package models

import "time"

type Group struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null;unique"`
	CreatedAt time.Time
	Members   []GroupMember
	Messages  []GroupMessage
}

type GroupMember struct {
	ID       uint `gorm:"primaryKey"`
	UserID   uint
	GroupID  uint
	IsAdmin  bool
	JoinedAt time.Time
}

type GroupMessage struct {
	ID        uint `gorm:"primaryKey"`
	GroupID   uint
	SenderID  uint
	Content   string `gorm:"not null"`
	CreatedAt time.Time
}
