package routers

import "github.com/gin-gonic/gin"


func InitRouter(router *gin.Engine)  {

	AuthRouter(router.Group("/auth"))
	
}
