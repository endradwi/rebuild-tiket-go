package routers

import (
	"tiket/controllers"
	"tiket/middleware"

	"github.com/gin-gonic/gin"
)

func UserRouter(router *gin.RouterGroup) {
	userRoutes := router.Group("/users")

	// Allow GET all users
	userRoutes.GET("", controllers.GetAllUsers)
	
	// Protected routes (require valid JWT token)
	protectedRoutes := userRoutes.Group("")
	protectedRoutes.Use(middleware.AuthMiddleware())
	
	protectedRoutes.GET("/:id", controllers.GetUserById)
	// Profile endpoints
	protectedRoutes.GET("/profile", controllers.GetProfile)
	// PATCH method combined with profile image upload capability
	protectedRoutes.PATCH("/profile", controllers.UpdateProfile)

	// Generic DELETE by ID (Can be protected by admin middleware later)
	protectedRoutes.DELETE("/:id", controllers.DeleteUser)
}