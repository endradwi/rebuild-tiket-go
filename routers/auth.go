package routers

import (
	"tiket/controllers"

	"github.com/gin-gonic/gin"
)

func AuthRouter(router *gin.RouterGroup)  {
	router.POST("/register", controllers.Register)	
}