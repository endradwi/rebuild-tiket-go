package routers

import (
	"tiket/controllers"
	"tiket/middleware"

	"github.com/gin-gonic/gin"
)

func OrderRouter(router *gin.RouterGroup) {
	orderRoutes := router.Group("/orders")

	// Public routes for data retrieval
	orderRoutes.GET("/seats", controllers.GetSeats)

	// Protected routes (optional, can be public for now to match simplicity of other routers)
	// But usually creating an order should be authenticated
	protectedRoutes := orderRoutes.Group("")
	protectedRoutes.Use(middleware.AuthMiddleware())

	protectedRoutes.POST("", controllers.CreateOrder)
	protectedRoutes.GET("/:id", controllers.GetOrderDetails)
	protectedRoutes.POST("/:id/payment", controllers.ProcessPayment)
}
