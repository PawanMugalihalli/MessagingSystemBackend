package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ViewDMPreviews returns the latest direct message preview for each unique conversation
func ViewDMPreviews(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var results []struct {
		PartnerID uint
		Username  string
		Content   string
		CreatedAt string
	}

	// SQL:
	// SELECT DISTINCT ON (LEAST(sender_id, receiver_id), GREATEST(sender_id, receiver_id))
	//   CASE WHEN sender_id = {userId} THEN receiver_id ELSE sender_id END AS partner_id,
	//   username, content, created_at
	// FROM direct_messages
	// JOIN users ON users.id = CASE WHEN sender_id = {userId} THEN receiver_id ELSE sender_id END
	// WHERE sender_id = {userId} OR receiver_id = {userId}
	// ORDER BY conversation_id, created_at DESC
	initializers.DB.Raw(`
		SELECT DISTINCT ON (LEAST(sender_id, receiver_id), GREATEST(sender_id, receiver_id))
			CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END AS partner_id,
			u.username,
			dm.content,
			dm.created_at
		FROM direct_messages dm
		JOIN users u ON u.id = CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END
		WHERE sender_id = ? OR receiver_id = ?
		ORDER BY LEAST(sender_id, receiver_id), GREATEST(sender_id, receiver_id), dm.created_at DESC
		LIMIT 10
	`, user.Id, user.Id, user.Id, user.Id).Scan(&results)

	c.JSON(http.StatusOK, results)
}

// ViewGroupPreviews returns the latest message previews from the groups the user is a member of
func ViewGroupPreviews(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var results []struct {
		GroupID   uint
		GroupName string
		Content   string
		CreatedAt string
	}

	// SQL:
	// SELECT g.id, g.name, gm.content, gm.created_at
	// FROM group_members
	// JOIN groups ON group_members.group_id = groups.id
	// LEFT JOIN LATERAL (
	//     SELECT * FROM group_messages WHERE group_id = groups.id ORDER BY created_at DESC LIMIT 1
	// ) gm ON true
	// WHERE group_members.user_id = {userId}
	// ORDER BY gm.created_at DESC
	// LIMIT 10;
	initializers.DB.Raw(`
		SELECT g.id as group_id, g.name as group_name, gm.content, gm.created_at
		FROM group_members m
		JOIN groups g ON m.group_id = g.id
		LEFT JOIN LATERAL (
			SELECT * FROM group_messages gm2 WHERE gm2.group_id = g.id ORDER BY created_at DESC LIMIT 1
		) gm ON true
		WHERE m.user_id = ?
		ORDER BY gm.created_at DESC
		LIMIT 10
	`, user.Id).Scan(&results)

	c.JSON(http.StatusOK, results)
}

// ViewChatHistory returns the latest 10 messages in a DM or group chat based on the type and id
func ViewChatHistory(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	chatType := c.Param("type")
	idParam := c.Param("id")

	// Parse chat ID from string to int
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if chatType == "dm" {
		// Check if user exists
		var partner models.User
		if err := initializers.DB.First(&partner, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var messages []models.DirectMessage

		// SQL:
		// SELECT * FROM direct_messages
		// WHERE (sender_id = {user.Id} AND receiver_id = {partner.Id})
		//    OR (sender_id = {partner.Id} AND receiver_id = {user.Id})
		// ORDER BY created_at DESC LIMIT 10;
		initializers.DB.
			Where(`(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)`,
				user.Id, partner.Id, partner.Id, user.Id).
			Order("created_at DESC").
			Limit(10).
			Find(&messages)

		// Return only necessary fields
		var resp []gin.H
		for _, msg := range messages {
			resp = append(resp, gin.H{
				"id":          msg.ID,
				"sender_id":   msg.SenderID,
				"receiver_id": msg.ReceiverID,
				"content":     msg.Content,
				"created_at":  msg.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, resp)
		return
	}

	if chatType == "group" {
		// Check if group exists
		var group models.Group
		if err := initializers.DB.First(&group, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}

		var messages []models.GroupMessage

		// SQL:
		// SELECT * FROM group_messages
		// WHERE group_id = {group.ID}
		// ORDER BY created_at DESC LIMIT 10;
		initializers.DB.
			Where("group_id = ?", group.ID).
			Order("created_at DESC").
			Limit(10).
			Find(&messages)

		// Return only necessary fields
		var resp []gin.H
		for _, msg := range messages {
			resp = append(resp, gin.H{
				"id":         msg.ID,
				"sender_id":  msg.SenderID,
				"group_id":   msg.GroupID,
				"content":    msg.Content,
				"created_at": msg.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, resp)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat type"})
}
