package models

import "time"

type DirectMessage struct {
	ID       uint `gorm:"primaryKey"`
	SenderID uint `gorm:"not null;index"` // Filtering by sender
	Sender   User `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`

	ReceiverID uint `gorm:"not null;index"` // Filtering by receiver
	Receiver   User `gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE"`

	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"index"` // For ordering by time
	UpdatedAt time.Time
}

// CREATE TABLE direct_messages (
//     id SERIAL PRIMARY KEY,
//     sender_id INTEGER NOT NULL,
//     receiver_id INTEGER NOT NULL,
//     content TEXT NOT NULL,
//     created_at TIMESTAMP,
//     updated_at TIMESTAMP,
//     CONSTRAINT fk_dm_sender FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
//     CONSTRAINT fk_dm_receiver FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
// );

// CREATE INDEX idx_direct_messages_sender_id ON direct_messages(sender_id);
// CREATE INDEX idx_direct_messages_receiver_id ON direct_messages(receiver_id);
// CREATE INDEX idx_direct_messages_created_at ON direct_messages(created_at);
