package main

import (
	"MessagingSystemBackend/internal/controllers"
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Initialize environment variables, database connection, and perform DB migrations
func init() {
	initializers.LoadEnv()      // Load environment variables from .env file
	initializers.ConnectToDb()  // Connect to the database using GORM
	initializers.SyncDatabase() // Auto-migrate all models to the database
}

func main() {
	r := gin.Default() // Initialize a Gin router with default middleware (logger and recovery)

	// Authentication routes
	r.POST("/signup", controllers.Signup)                            // User registration
	r.POST("/login", controllers.Login)                              // User login
	r.GET("/validate", middleware.RequireAuth, controllers.Validate) // Authenticated route to verify user session
	r.GET("/logout", middleware.RequireAuth, controllers.Logout)     // Logout user by clearing auth token

	// Direct message (DM) routes
	dmRoutes := r.Group("/dm")
	dmRoutes.Use(middleware.RequireAuth)                // Require authentication for all DM routes
	dmRoutes.POST(":id", controllers.SendDirectMessage) // Send a direct message to a user by ID
	dmRoutes.GET(":id", controllers.GetDirectMessage)   // Get direct messages with a specific user

	// Group-related routes
	groupRoutes := r.Group("/groups")
	groupRoutes.Use(middleware.RequireAuth)                             // Require authentication for all group routes
	groupRoutes.POST("/create", controllers.CreateGroup)                // Create a new group
	groupRoutes.POST("/:id/message", controllers.SendGroupMessage)      // Send a message to a group
	groupRoutes.POST("/:id/add-member", controllers.AddGroupMember)     // Add a new member to a group
	groupRoutes.POST("/:id/add-admin", controllers.AddAdmin)            // Promote a member to group admin
	groupRoutes.GET("/:id/summary", controllers.SummarizeGroupMessages) // Summarize group chat using NLP
	groupRoutes.GET("/:id", controllers.GetGroupMessage)                // Retrieve all messages from a group

	// Routes for viewing message previews and chat history
	viewRoutes := r.Group("/view")
	viewRoutes.Use(middleware.RequireAuth)                         // Require authentication for all view routes
	viewRoutes.GET("/dms", controllers.ViewDMPreviews)             // View DM conversation previews
	viewRoutes.GET("/groups", controllers.ViewGroupPreviews)       // View group conversation previews
	viewRoutes.GET("/chat/:type/:id", controllers.ViewChatHistory) // View full chat history (DM/group)

	// Edit message routes
	groupRoutes.PUT("/message/:id", controllers.EditGroupMessage) // Edit a group message by ID
	dmRoutes.PUT("/message/:id", controllers.EditDirectMessage)   // Edit a direct message by ID

	// Start the Gin server on default port 8080
	r.Run()
}
