package main

import (
	"tiket/routers"

	"github.com/gin-gonic/gin"
)


// @title           Tikitz API
// @version         1.0
// @description     API for Movie Ticket Booking System
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8888
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	router := gin.Default()

	routers.InitRouter(router)

	router.Run(":8888")
}
