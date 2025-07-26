package initializers

import "MessagingSystemBackend/internal/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
}
