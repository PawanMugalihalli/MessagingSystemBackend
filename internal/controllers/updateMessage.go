package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// EditGroupMessage handles editing of a group message by its sender
func EditGroupMessage(c *gin.Context) {
	// Get the authenticated user from context
	user := c.MustGet("user").(models.User)

	// Get message ID from the URL path
	msgID := c.Param("id")

	var msg models.GroupMessage

	// Retrieve the group message from the database by ID
	// SQL equivalent: SELECT * FROM group_messages WHERE id = ? LIMIT 1;
	if err := initializers.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Ensure that only the sender of the message can edit it
	if msg.SenderID != user.Id {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to edit this message since you are not the user"})
		return
	}

	// Enforce 1-hour time limit for editing messages
	if time.Since(msg.CreatedAt) > time.Hour {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to edit this message since the time limit has been exceeded"})
		return
	}

	// Struct to bind request body
	var body struct {
		Content   string    `json:"content"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	// Parse and validate JSON body from request
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check optimistic locking: compare if the message has changed since last fetch
	if !msg.UpdatedAt.Equal(body.UpdatedAt) {
		c.JSON(http.StatusConflict, gin.H{"error": "Message has been updated elsewhere. Please refresh and try again."})
		return
	}

	// Update message content
	msg.Content = body.Content

	// Save the updated message to the database
	// SQL equivalent: UPDATE group_messages SET content = ? WHERE id = ?;
	if err := initializers.DB.Save(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"success": "Message updated"})
}

// EditDirectMessage handles editing of a direct (private) message
func EditDirectMessage(c *gin.Context) {
	// Get the authenticated user from context
	user := c.MustGet("user").(models.User)

	// Get message ID from the URL path
	msgID := c.Param("id")

	var msg models.DirectMessage

	// Retrieve the direct message from the database by ID
	// SQL equivalent: SELECT * FROM direct_messages WHERE id = ? LIMIT 1;
	if err := initializers.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Check if the user is the sender and within 1 hour time window
	if msg.SenderID != user.Id || time.Since(msg.CreatedAt) > time.Hour {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to edit this message"})
		return
	}

	// Struct to bind request body
	var body struct {
		Content   string    `json:"content"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	// Parse and validate JSON body from request
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Optimistic locking check to avoid editing stale data
	if !msg.UpdatedAt.Equal(body.UpdatedAt) {
		c.JSON(http.StatusConflict, gin.H{"error": "Message has been updated elsewhere. Please refresh and try again."})
		return
	}

	// Apply content update
	msg.Content = body.Content

	// Save the updated message to the database
	// SQL equivalent: UPDATE direct_messages SET content = ? WHERE id = ?;
	if err := initializers.DB.Save(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"success": "Message updated"})
}
