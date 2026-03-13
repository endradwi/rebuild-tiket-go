package routers

import "github.com/gin-gonic/gin"


func InitRouter(router *gin.Engine)  {

	// Serve the uploads directory statically
	router.Static("/uploads", "./uploads")

	AuthRouter(router.Group("/auth"))
	UserRouter(router.Group(""))
	MovieRouter(router.Group(""))
	OrderRouter(router.Group(""))
	AdminRouter(router.Group("/admin"))
	
}
