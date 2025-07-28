package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func EditGroupMessage(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	msgID := c.Param("id")

	var msg models.GroupMessage
	if err := initializers.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Allow edit only if user is sender and within 1 hour
	if msg.SenderID != user.Id {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to edit this message since you are not the user"})
		return
	}
	if time.Since(msg.CreatedAt) > time.Hour{
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to edit this message since the time limit has been exceeded"})
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	msg.Content = body.Content
	initializers.DB.Save(&msg)

	c.JSON(http.StatusOK, gin.H{"success": "Message updated"})
}


func EditDirectMessage(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	msgID := c.Param("id")

	var msg models.DirectMessage
	if err := initializers.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Allow edit only if user is sender and within 1 hour
	if msg.SenderID != user.Id || time.Since(msg.CreatedAt) > time.Hour {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to edit this message"})
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	msg.Content = body.Content
	initializers.DB.Save(&msg)

	c.JSON(http.StatusOK, gin.H{"success": "Message updated"})
}
