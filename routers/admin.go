package routers

import (
	"tiket/controllers"
	"tiket/middleware"

	"github.com/gin-gonic/gin"
)

func AdminRouter(r *gin.RouterGroup) {
	adminRoutes := r.Group("")
	adminRoutes.Use(middleware.AuthMiddleware())
	adminRoutes.Use(middleware.RoleMiddleware("ADMIN"))

	adminRoutes.GET("/stats", controllers.GetDashboardStats)
}
