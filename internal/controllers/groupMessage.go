package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateGroup(c *gin.Context) {
	var body struct {
		Name string
	}
	if err := c.Bind(&body); err != nil || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group name"})
		return
	}

	user := c.MustGet("user").(models.User)

	// Check group name uniqueness (optional)
	var existing models.Group
	initializers.DB.Where("name = ?", body.Name).First(&existing)
	if existing.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name already taken"})
		return
	}

	group := models.Group{Name: body.Name}
	err := initializers.DB.Create(&group).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not create group"})
		return
	}

	// Make creator an admin
	err = PromoteToAdmin(group.ID, user.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group_id": group.ID})
}

func SendGroupMessage(c *gin.Context) {
	groupName := c.Param("name")
	if groupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group name"})
		return
	}

	var group models.Group
	if err := initializers.DB.Where("name = ?", groupName).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var body struct {
		Content string
	}
	if err := c.Bind(&body); err != nil || body.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message"})
		return
	}

	user := c.MustGet("user").(models.User)

	// Check membership
	var member models.GroupMember
	result := initializers.DB.Where("group_id = ? AND user_id = ?", group.ID, user.Id).First(&member)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not a member of this group"})
		return
	}

	msg := models.GroupMessage{
		GroupID:   group.ID,
		SenderID:  user.Id,
		Content:   body.Content,
		CreatedAt: time.Now(),
	}
	initializers.DB.Create(&msg)
	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

func CanAddGroupMember(groupID uint) bool {
	var count int64
	initializers.DB.Model(&models.GroupMember{}).Where("group_id = ?", groupID).Count(&count)
	return count < 25
}

func CanAddAdmin(groupID uint) bool {
	var adminCount int64
	initializers.DB.Model(&models.GroupMember{}).Where("group_id = ? AND is_admin = true", groupID).Count(&adminCount)
	return adminCount < 2
}

func PromoteToAdmin(groupID uint, userID uint) error {
	// Check if already admin
	var existing models.GroupMember
	if err := initializers.DB.
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&existing).Error; err == nil {
		if existing.IsAdmin {
			return fmt.Errorf("user is already an admin")
		}
		existing.IsAdmin = true
		return initializers.DB.Save(&existing).Error
	}

	if !CanAddAdmin(groupID) {
		return fmt.Errorf("group already has 2 admins")
	}
	admin := models.GroupMember{
		UserID:   userID,
		GroupID:  groupID,
		IsAdmin:  true,
		JoinedAt: time.Now(),
	}
	return initializers.DB.Create(&admin).Error
}

func IsGroupAdmin(groupID uint, userID uint) bool {
	var gm models.GroupMember
	err := initializers.DB.
		Where("group_id = ? AND user_id = ? AND is_admin = true", groupID, userID).
		First(&gm).Error
	return err == nil
}

func AddAdmin(c *gin.Context) {
	groupName := c.Param("name")
	if groupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing group name"})
		return
	}

	var group models.Group
	if err := initializers.DB.Where("name = ?", groupName).First(&group).Error; err != nil {
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

func AddGroupMember(c *gin.Context) {
	groupName := c.Param("name")
	if groupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing group name"})
		return
	}

	var group models.Group
	if err := initializers.DB.Where("name = ?", groupName).First(&group).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group not found"})
		return
	}

	currentUser := c.MustGet("user").(models.User)

	// Check if current user is an admin of the group
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

	var user models.User
	if err := initializers.DB.Where("username = ?", body.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
		return
	}

	// Check if already a member
	var existing models.GroupMember
	initializers.DB.Where("group_id = ? AND user_id = ?", group.ID, user.Id).First(&existing)
	if existing.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already a member"})
		return
	}

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
		initializers.DB.Create(&member)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to group"})
}

func GetGroupMessage(c *gin.Context) {
	msgID := c.Param("id")
	var msg models.GroupMessage

	if err := initializers.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}
	fmt.Println(msg.UpdatedAt)
	c.JSON(http.StatusOK, gin.H{
	"id": msg.ID,
	"content": msg.Content,
	"updated_at": msg.UpdatedAt.UTC(), // 👈 ensures UTC
    })
}
