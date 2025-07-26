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

	err = AddAdmin(uint(group.ID), user.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group_id": group.ID})
}

func SendGroupMessage(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
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
	result := initializers.DB.Where("group_id = ? AND user_id = ?", groupID, user.Id).First(&member)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not a member of this group"})
		return
	}

	msg := models.GroupMessage{
		GroupID:   uint(groupID),
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

func AddAdmin(groupID uint, userID uint) error {
	if !CanAddAdmin(groupID) {
		return fmt.Errorf("Group already has 2 admins")
	}
	admin := models.GroupMember{
		UserID:   userID,
		GroupID:  groupID,
		IsAdmin:  true,
		JoinedAt: time.Now(),
	}
	return initializers.DB.Create(&admin).Error
}

func AddGroupMember(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("id"))
	fmt.Println("groupID:", groupID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	if !CanAddGroupMember(uint(groupID)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group already has 25 members"})
		return
	}

	var body struct {
		UserID  uint `json:"user_id"`
		IsAdmin bool `json:"is_admin"`
	}
	if err := c.Bind(&body); err != nil || body.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var user models.User
	if err := initializers.DB.First(&user, body.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
		return
	}
	
    fmt.Printf("Parsed body: %+v\n", body)

	fmt.Println(groupID, body.UserID)
	// Optional: check if user already a member
	var existing models.GroupMember
	initializers.DB.Where("group_id = ? AND user_id = ?", groupID, body.UserID).First(&existing)
	if existing.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already a member"})
		return
	}

	if body.IsAdmin {
		err := AddAdmin(uint(groupID), body.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		member := models.GroupMember{
			UserID:   body.UserID,
			GroupID:  uint(groupID),
			IsAdmin:  false,
			JoinedAt: time.Now(),
		}
		initializers.DB.Create(&member)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to group"})
}
