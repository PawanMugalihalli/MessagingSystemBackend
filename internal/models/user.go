package models

import "time"

type User struct {
	Id        uint   `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;not null"` // Index for lookup
	Password  string `gorm:"not null"`
	CreatedAt time.Time
}

// CREATE TABLE users (
//     id SERIAL PRIMARY KEY,
//     username VARCHAR(255) NOT NULL UNIQUE,
//     password VARCHAR(255) NOT NULL,
//     created_at TIMESTAMP
// );
