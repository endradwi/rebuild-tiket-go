package controllers

import (
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {

	

		var user lib.User
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(400, lib.Response{
				Status:  400,
				Message: "bad request",
			})
			return
		}

		result := models.Register(user)


		c.JSON(200, lib.Response{
			Status:  200,
			Message: "success",
			Result:  result,
		})
	
}