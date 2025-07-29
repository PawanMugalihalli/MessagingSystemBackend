package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ViewDMPreviews(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var results []struct {
		PartnerID uint
		Username  string
		Content   string
		CreatedAt string
	}

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

func ViewGroupPreviews(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var results []struct {
		GroupID   uint
		GroupName string
		Content   string
		CreatedAt string
	}

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

func ViewChatHistory(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	chatType := c.Param("type")
	idParam := c.Param("id") // expecting numeric ID

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if chatType == "dm" {
		var partner models.User
		if err := initializers.DB.First(&partner, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var messages []models.DirectMessage
		initializers.DB.
			Where(`(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)`,
				user.Id, partner.Id, partner.Id, user.Id).
			Order("created_at DESC").
			Limit(10).
			Find(&messages)

		c.JSON(http.StatusOK, messages)
		return
	} else if chatType == "group" {
		var group models.Group
		if err := initializers.DB.First(&group, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}

		var messages []models.GroupMessage
		initializers.DB.
			Where("group_id = ?", group.ID).
			Order("created_at DESC").
			Limit(10).
			Find(&messages)

		c.JSON(http.StatusOK, messages)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat type"})
}
