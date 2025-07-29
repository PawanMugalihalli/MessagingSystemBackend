package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateGroup creates a new group with the current user as the creator.
func CreateGroup(c *gin.Context) {
	var body struct {
		Name string
	}
	if err := c.Bind(&body); err != nil || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group name"})
		return
	}

	// Get the authenticated user from the context
	user := c.MustGet("user").(models.User)

	// Check if group name already exists
	// SQL: SELECT * FROM groups WHERE name = ? LIMIT 1
	var existing models.Group
	initializers.DB.Where("name = ?", body.Name).First(&existing)
	if existing.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name already taken"})
		return
	}

	// Create the group
	// SQL: INSERT INTO groups (name, created_by, created_at) VALUES (?, ?, ?)
	group := models.Group{
		Name:      body.Name,
		CreatedBy: user.Id,
		CreatedAt: time.Now(),
	}
	if err := initializers.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not create group"})
		return
	}

	// Promote the creator to admin
	// (Assumes PromoteToAdmin internally performs an INSERT or UPDATE on group_members or similar table)
	if err := PromoteToAdmin(group.ID, user.Id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send the created group ID in response
	c.JSON(http.StatusOK, gin.H{"group_id": group.ID})
}

// SendGroupMessage allows a group member to send a message to the group.
func SendGroupMessage(c *gin.Context) {
	groupIDParam := c.Param("id")
	if groupIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	groupID, err := strconv.Atoi(groupIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID must be a number"})
		return
	}

	// SQL: SELECT * FROM groups WHERE id = ? LIMIT 1;
	var group models.Group
	if err := initializers.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&body); err != nil || body.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message"})
		return
	}

	user := c.MustGet("user").(models.User)

	// SQL: SELECT * FROM group_members WHERE group_id = ? AND user_id = ? LIMIT 1;
	var member models.GroupMember
	result := initializers.DB.Where("group_id = ? AND user_id = ?", group.ID, user.Id).First(&member)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not a member of this group"})
		return
	}

	// Create the group message
	msg := models.GroupMessage{
		GroupID:   group.ID,
		SenderID:  user.Id,
		Content:   body.Content,
		CreatedAt: time.Now(),
	}

	// SQL: INSERT INTO group_messages (group_id, sender_id, content, created_at) VALUES (?, ?, ?, ?);
	initializers.DB.Create(&msg)

	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

// CanAddGroupMember returns true if the group has fewer than 25 members.
func CanAddGroupMember(groupID uint) bool {
	var count int64
	// SQL: SELECT COUNT(*) FROM group_members WHERE group_id = ?;
	initializers.DB.Model(&models.GroupMember{}).Where("group_id = ?", groupID).Count(&count)
	return count < 25
}

// CanAddAdmin returns true if there are fewer than 2 admins in the group.
func CanAddAdmin(groupID uint) bool {
	var adminCount int64
	// SQL: SELECT COUNT(*) FROM group_members WHERE group_id = ? AND is_admin = true;
	initializers.DB.Model(&models.GroupMember{}).Where("group_id = ? AND is_admin = true", groupID).Count(&adminCount)
	return adminCount < 2
}

// PromoteToAdmin promotes a user to admin if allowed.
func PromoteToAdmin(groupID uint, userID uint) error {
	// SQL: SELECT * FROM group_members WHERE group_id = ? AND user_id = ?;
	var member models.GroupMember
	err := initializers.DB.
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error

	if err != nil {
		return fmt.Errorf("user must be a group member to be promoted")
	}

	if member.IsAdmin {
		return fmt.Errorf("user is already an admin")
	}

	// Check max admin limit
	if !CanAddAdmin(groupID) {
		return fmt.Errorf("group already has 2 admins")
	}

	member.IsAdmin = true
	return initializers.DB.Save(&member).Error
}

// IsGroupAdmin checks if a user is an admin of a group.
func IsGroupAdmin(groupID uint, userID uint) bool {
	var gm models.GroupMember
	// SQL: SELECT * FROM group_members WHERE group_id = ? AND user_id = ? AND is_admin = true LIMIT 1;
	err := initializers.DB.
		Where("group_id = ? AND user_id = ? AND is_admin = true", groupID, userID).
		First(&gm).Error
	return err == nil
}

// AddAdmin promotes another user to admin, if the requester is an admin.
func AddAdmin(c *gin.Context) {
	groupIDParam := c.Param("id")
	if groupIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing group ID"})
		return
	}

	groupID, err := strconv.Atoi(groupIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// SQL: SELECT * FROM groups WHERE id = ?;
	var group models.Group
	if err := initializers.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	currentUser := c.MustGet("user").(models.User)

	if !IsGroupAdmin(group.ID, currentUser.Id) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Only an admin can promote members"})
		return
	}

	var body struct {
		Username string `json:"username"`
	}
	if err := c.Bind(&body); err != nil || body.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username required"})
		return
	}

	// SQL: SELECT * FROM users WHERE username = ?;
	var targetUser models.User
	if err := initializers.DB.Where("username = ?", body.Username).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if IsGroupAdmin(group.ID, targetUser.Id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already an admin"})
		return
	}

	if !CanAddAdmin(group.ID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group already has 2 admins"})
		return
	}

	if err := PromoteToAdmin(group.ID, targetUser.Id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User promoted to admin"})
}

// AddGroupMember adds a new member to a group, optionally as an admin.
func AddGroupMember(c *gin.Context) {
	groupIDParam := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDParam, 10, 64)
	if err != nil || groupID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// SQL: SELECT * FROM groups WHERE id = ?;
	var group models.Group
	if err := initializers.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group not found"})
		return
	}

	currentUser := c.MustGet("user").(models.User)

	if !IsGroupAdmin(group.ID, currentUser.Id) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Only admins can add members to this group"})
		return
	}

	if !CanAddGroupMember(group.ID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group already has 25 members"})
		return
	}

	var body struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"is_admin"`
	}
	if err := c.Bind(&body); err != nil || body.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username"})
		return
	}

	// SQL: SELECT * FROM users WHERE username = ?;
	var user models.User
	if err := initializers.DB.Where("username = ?", body.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
		return
	}

	// SQL: SELECT * FROM group_members WHERE group_id = ? AND user_id = ?;
	var existing models.GroupMember
	initializers.DB.Where("group_id = ? AND user_id = ?", group.ID, user.Id).First(&existing)
	if existing.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already a member"})
		return
	}

	// Add as admin or normal member
	if body.IsAdmin {
		if err := PromoteToAdmin(group.ID, user.Id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		member := models.GroupMember{
			UserID:   user.Id,
			GroupID:  group.ID,
			IsAdmin:  false,
			JoinedAt: time.Now(),
		}
		// SQL: INSERT INTO group_members (...) VALUES (...);
		initializers.DB.Create(&member)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to group"})
}

// GetGroupMessage fetches a group message by its ID.
func GetGroupMessage(c *gin.Context) {
	msgID := c.Param("id")
	var msg models.GroupMessage

	// SQL: SELECT * FROM group_messages WHERE id = ?;
	if err := initializers.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	fmt.Println(msg.UpdatedAt)

	c.JSON(http.StatusOK, gin.H{
		"id":         msg.ID,
		"content":    msg.Content,
		"updated_at": msg.UpdatedAt.UTC(), // 👈 ensures UTC
	})
}
