package main

import (
	"tiket/routers"

	"github.com/gin-gonic/gin"
)


func main() {
	router := gin.Default()

	routers.InitRouter(router)

	router.Run(":8888")
}
