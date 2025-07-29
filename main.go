package main

import (
	"MessagingSystemBackend/internal/controllers"
	"MessagingSystemBackend/internal/initializers"
	"MessagingSystemBackend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnv()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {
	r := gin.Default()

	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)
	r.GET("/validate", middleware.RequireAuth, controllers.Validate)
	r.GET("/logout", middleware.RequireAuth, controllers.Logout)

	// Direct message endpoint
	dmRoutes := r.Group("/dm")
	dmRoutes.Use(middleware.RequireAuth)
	dmRoutes.POST(":id", controllers.SendDirectMessage)
	dmRoutes.GET(":id", controllers.GetDirectMessage)

	groupRoutes := r.Group("/groups")
	groupRoutes.Use(middleware.RequireAuth)
	groupRoutes.POST("/create", controllers.CreateGroup)
	groupRoutes.POST("/:id/message", controllers.SendGroupMessage)
	groupRoutes.POST("/:id/add-member", controllers.AddGroupMember)
	groupRoutes.POST("/:id/add-admin", controllers.AddAdmin)
	groupRoutes.GET("/:id/summary", controllers.SummarizeGroupMessages)
	groupRoutes.GET("/:id", controllers.GetGroupMessage)

	viewRoutes := r.Group("/view")
	viewRoutes.Use(middleware.RequireAuth)
	viewRoutes.GET("/dms", controllers.ViewDMPreviews)
	viewRoutes.GET("/groups", controllers.ViewGroupPreviews)
	viewRoutes.GET("/chat/:type/:id", controllers.ViewChatHistory)

	groupRoutes.PUT("/message/:id", controllers.EditGroupMessage)
	dmRoutes.PUT("/message/:id", controllers.EditDirectMessage)

	r.Run()
}
