package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"tiket/lib"
	"tiket/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func Login(c *gin.Context) {
	var user lib.User
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(400, lib.Response{
			Status: 400,
			Message: "bad request",
		})
	}

	dbUser, err := models.FindEmail(user)
	if err != nil {
		c.JSON(404, lib.Response{
			Status:  404,
			Message: "Email not found",
		})
		return
	}

	match, err := lib.GenerateToken(user.Password, dbUser.Password)
	if err != nil {
		fmt.Printf("Error comparing password: %v\n", err)
		c.JSON(500, lib.Response{
			Status:  500,
			Message: "internal server error",
		})
		return
	}

	if !match {
		c.JSON(400, lib.Response{
			Status:  400,
			Message: "password not match",
		})
		return
	}

	tokenJWT, err := lib.GenerateTokenJwt(map[string]interface{}{
		"userId": dbUser.Id,
	})
	if err != nil {
		c.JSON(500, lib.Response{
			Status:  500,
			Message: "failed to generate token",
		})
		return
	}

	c.JSON(200, lib.Response{
		Status:  200,
		Message: "success",
		Result: gin.H{
			"token": tokenJWT,
		},
	})

}

func ForgotPassword(c *gin.Context) {
	var user lib.User
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(400, lib.Response{
			Status:  400,
			Message: "bad request",
		})
		return
	}

	dbUser, err := models.FindEmail(user)
	if err != nil {
		c.JSON(404, lib.Response{
			Status:  404,
			Message: "Email not found",
		})
		return
	}

	token := uuid.NewString()
	
	hashToken := sha256.Sum256([]byte(token))
	hashTokenEncode := hex.EncodeToString(hashToken[:])

	resetPassword := lib.ResetPassword{
		ProfileId: dbUser.Id,
		TokenHash: hashTokenEncode,
		ExpiredAt: time.Now().Add(time.Hour * 24),
	}

	resetLink := fmt.Sprintf("http://localhost:8080/auth/reset-password?token=%s", hashTokenEncode)

	err = lib.SendResetPassword(dbUser.Email, resetLink)
	if err != nil {
		c.JSON(500, lib.Response{
			Status:  500,
			Message: "failed to send email",
		})
		return
	}

	err = models.CreateResetPassword(resetPassword)
	if err != nil {
		c.JSON(500, lib.Response{
			Status:  500,
			Message: "internal server error",
		})
		return
	}

	c.JSON(200, lib.Response{
		Status:  200,
		Message: "success",
	})
	
}

func ValidatePasswordReset(c *gin.Context) {
	var resetReq lib.ResetPasswordRequest
	
	if err := c.ShouldBind(&resetReq); err != nil {
		c.JSON(400, lib.Response{
			Status:  400,
			Message: "bad request",
		})
		return
	}

	if resetReq.Token == "" {
		resetReq.Token = c.Query("token")
	}

	resetPassword, err := models.FindResetPassword(resetReq.Token)
	if err != nil {
		c.JSON(404, lib.Response{
			Status:  404,
			Message: "reset password not found",
		})
		return
	}

	hashPassword, err := lib.HashPassword(resetReq.Password)

	if err != nil {
		c.JSON(500, lib.Response{
			Status:  500,
			Message: "internal server error",
		})
		return
	}
	
	err = models.UpdatePassword(resetPassword.ProfileId, hashPassword)
	if err != nil {
		c.JSON(500, lib.Response{
			Status:  500,
			Message: "internal server error",
		})
		return
	}

	// err = models.DeleteResetPassword(resetPassword.TokenHash)
	// if err != nil {
	// 	c.JSON(500, lib.Response{
	// 		Status:  500,
	// 		Message: "internal server error",
	// 	})
	// 	return
	// }

	c.JSON(200, lib.Response{
		Status:  200,
		Message: "success",
	})
	
}