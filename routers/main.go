package routers

import (
	_ "tiket/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)


func InitRouter(router *gin.Engine)  {

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve the uploads directory statically
	router.Static("/uploads", "./uploads")

	AuthRouter(router.Group("/auth"))
	UserRouter(router.Group(""))
	MovieRouter(router.Group(""))
	OrderRouter(router.Group(""))
	AdminRouter(router.Group("/admin"))
	
}
