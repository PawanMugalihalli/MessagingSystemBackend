package models

import "time"

type Group struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"uniqueIndex;not null"` // Fast lookup by name
	CreatedBy uint      `gorm:"not null;index"`       // For joins
	Creator   User      `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"index"` // If sorting or filtering by date
}

// CREATE TABLE groups (
//     id SERIAL PRIMARY KEY,
//     name VARCHAR(255) NOT NULL UNIQUE,
//     created_at TIMESTAMP
// );

// if you sort/filter by recent groups:
// CREATE INDEX idx_groups_created_at ON groups(created_at);

type GroupMember struct {
	ID uint `gorm:"primaryKey"`

	GroupID uint  `gorm:"index;uniqueIndex:idx_group_user"` // Used in WHERE and JOIN
	Group   Group `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`

	UserID uint `gorm:"index;uniqueIndex:idx_group_user"` // Used in WHERE and JOIN
	User   User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	IsAdmin  bool      `gorm:"not null;default:false;index"` // Filtering (admin checks)
	JoinedAt time.Time `gorm:"index"`                        // Optional, for sorting by join time
}

// CREATE TABLE group_members (
//     id SERIAL PRIMARY KEY,
//     user_id INTEGER NOT NULL,
//     group_id INTEGER NOT NULL,
//     is_admin BOOLEAN DEFAULT false,
//     joined_at TIMESTAMP,
//     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
//     FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
//     UNIQUE (group_id, user_id)  -- Enforce one entry per user per group
// );
// -- Index to support admin-count checks:
// CREATE INDEX idx_group_admin ON group_members(group_id, is_admin);

type GroupMessage struct {
	ID uint `gorm:"primaryKey"`

	GroupID uint  `gorm:"index"` // Frequently used in WHERE clauses
	Group   Group `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`

	SenderID uint `gorm:"index"` // Used to filter messages sent by user
	Sender   User `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`

	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"index"` // Sorting messages
	UpdatedAt time.Time
}

// CREATE TABLE group_messages (
//     id SERIAL PRIMARY KEY,
//     group_id INTEGER NOT NULL,
//     sender_id INTEGER NOT NULL,
//     content TEXT NOT NULL,
//     created_at TIMESTAMP,
//     updated_at TIMESTAMP,
//     CONSTRAINT fk_group_msg_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
//     CONSTRAINT fk_group_msg_sender FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
// );

// CREATE INDEX idx_group_messages_group_id ON group_messages(group_id);
// CREATE INDEX idx_group_messages_sender_id ON group_messages(sender_id);
// CREATE INDEX idx_group_messages_created_at ON group_messages(created_at);
