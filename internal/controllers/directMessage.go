package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SendDirectMessage(c *gin.Context) {
	receiverUsername := c.Param("username")
	if receiverUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Receiver username required"})
		return
	}

	var receiver models.User
	if err := initializers.DB.Where("username = ?", receiverUsername).First(&receiver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&body); err != nil || body.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message content"})
		return
	}

	sender := c.MustGet("user").(models.User)

	message := models.DirectMessage{
		SenderID:   sender.Id,
		ReceiverID: receiver.Id,
		Content:    body.Content,
		CreatedAt:  time.Now(),
	}

	if err := initializers.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}


func GetDirectMessage(c *gin.Context) {
	msgID := c.Param("id")
	var msg models.DirectMessage

	if err := initializers.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
	"id": msg.ID,
	"content": msg.Content,
	"updated_at": msg.UpdatedAt.UTC(), // 👈 ensures UTC
    })
}