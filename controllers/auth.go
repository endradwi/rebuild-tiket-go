package controllers

import (
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
		var user lib.UserRole
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(400, lib.Response{
				Status:  400,
				Message: "bad request",
			})
			return
		}

	err := models.Register(user)
	if err != nil {
		status := 500
		message := "internal server error"

		if err.Error() == "email already exists" {
			status = 400
			message = "email already exists"
		}

		c.JSON(status, lib.Response{
			Status:  status,
			Message: message,
		})
		return
	}

	c.JSON(200, lib.Response{
		Status:  200,
		Message: "success",
	})
	
}