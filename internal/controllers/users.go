package controllers

import (
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Signup handles user registration
func Signup(c *gin.Context) {
	// Get the username and password from the request body
	var body struct {
		Username string
		Password string
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Hash the password using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Create a new user record with the hashed password
	user := models.User{Username: body.Username, Password: string(hash)}

	// SQL: INSERT INTO users (username, password) VALUES (?, ?);
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{"success": "Registered Successfully"})
}

// Login handles user authentication and JWT creation
func Login(c *gin.Context) {
	// Get the username and password from the request body
	var body struct {
		Username string
		Password string
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Look up the user by username
	// SQL: SELECT * FROM users WHERE username = ? LIMIT 1;
	var user models.User
	initializers.DB.First(&user, "username = ?", body.Username)

	if user.Id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email",
		})
		return
	}

	// Compare provided password with stored password hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid password",
		})
		return
	}

	// Generate a JWT token with user ID as subject and 30 days expiry
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Id,                                    // subject claim
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(), // expiry time
	})

	// Sign the token using the SECRET from environment variables
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create a jwt token",
		})
		return
	}

	// Set the JWT as a cookie in the response
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"success": "Logged in Successfully",
	})
}

// Validate confirms the user is logged in and the token is valid
func Validate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "I'm logged in",
	})
}

// Logout clears the Authorization cookie to log the user out
func Logout(c *gin.Context) {
	// Invalidate the cookie by setting it with an immediate expiry
	c.SetCookie("Authorization", "", -1, "", "", false, true)

	// Respond with logout success
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
