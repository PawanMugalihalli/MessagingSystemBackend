package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func SummarizeGroupMessages(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var group models.Group
	if err := initializers.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var messages []models.GroupMessage
	initializers.DB.Where("group_id = ?", group.ID).Order("created_at DESC").Limit(50).Find(&messages)

	if len(messages) == 0 {
		c.JSON(http.StatusOK, gin.H{"summary": "No messages to summarize."})
		return
	}

	type MsgInput struct {
		Sender  string `json:"sender"`
		Content string `json:"content"`
	}

	var msgInputs []MsgInput
	participantsSet := map[string]struct{}{}
	for _, msg := range messages {
		var user models.User
		initializers.DB.First(&user, msg.SenderID)
		msgInputs = append(msgInputs, MsgInput{Sender: user.Username, Content: msg.Content})
		participantsSet[user.Username] = struct{}{}
	}

	prompt := "You are an assistant that summarizes group conversations. Given a list of user messages, return a concise summary.\n\nMessages:\n"
	for _, m := range msgInputs {
		prompt += fmt.Sprintf("%s: %s\n", m.Sender, m.Content)
	}

	type GroqRequest struct {
		Messages []map[string]string `json:"messages"`
		Model    string              `json:"model"`
	}

	groqPayload := GroqRequest{
		Model: "llama3-70b-8192",
		Messages: []map[string]string{
			{"role": "system", "content": "You are a helpful assistant that summarizes group chat discussions."},
			{"role": "user", "content": prompt},
		},
	}

	requestBody, _ := json.Marshal(groqPayload)

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call LLM API"})
		return
	}
	defer resp.Body.Close()

	responseBytes, _ := io.ReadAll(resp.Body)
	fmt.Println("Groq Response:", string(responseBytes))

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(responseBytes, &result)

	if len(result.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No summary returned"})
		return
	}

	summary := strings.TrimSpace(result.Choices[0].Message.Content)

	var participants []string
	for p := range participantsSet {
		participants = append(participants, p)
	}

	c.JSON(http.StatusOK, gin.H{
		"participants": participants,
		"summary_text": summary,
	})
}
