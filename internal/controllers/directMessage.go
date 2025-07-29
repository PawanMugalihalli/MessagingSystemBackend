package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SendDirectMessage handles sending a direct message from one user to another.
// Expects receiver ID in the URL and message content in the JSON body.
func SendDirectMessage(c *gin.Context) {
	receiverIDStr := c.Param("id") // Get receiver ID from the URL path
	if receiverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Receiver ID required"})
		return
	}

	// Convert receiver ID from string to integer
	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
		return
	}

	// SQL: SELECT * FROM users WHERE id = receiverID LIMIT 1;
	// Check if the receiver exists in the users table
	var receiver models.User
	if err := initializers.DB.First(&receiver, receiverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		return
	}

	// Parse the message content from the request body
	var body struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&body); err != nil || body.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message content"})
		return
	}

	// Get the currently authenticated user (sender)
	sender := c.MustGet("user").(models.User)

	// Create a new direct message record
	message := models.DirectMessage{
		SenderID:   sender.Id,
		ReceiverID: receiver.Id,
		Content:    body.Content,
		CreatedAt:  time.Now(),
	}

	// SQL: INSERT INTO direct_messages (sender_id, receiver_id, content, created_at)
	//        VALUES (?, ?, ?, ?);
	// Save the new direct message to the database
	if err := initializers.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

// GetDirectMessage handles retrieving a direct message by its ID.
func GetDirectMessage(c *gin.Context) {
	msgID := c.Param("id") // Get message ID from the URL path
	var msg models.DirectMessage

	// SQL: SELECT * FROM direct_messages WHERE id = msgID LIMIT 1;
	// Fetch the message from the database
	if err := initializers.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Respond with message details in JSON
	c.JSON(http.StatusOK, gin.H{
		"id":         msg.ID,
		"content":    msg.Content,
		"updated_at": msg.UpdatedAt.UTC(), // 👈 ensures UTC
	})
}
