package initializers

import "MessagingSystemBackend/internal/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{}, &models.Group{}, &models.GroupMember{}, &models.GroupMessage{})
}
