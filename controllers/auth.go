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

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      lib.RegisterRequest  true  "User Registration Details"
// @Success      200   {object}  lib.Response
// @Failure      400   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /auth/register [post]
func Register(c *gin.Context) {
	var user lib.RegisterRequest
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

// Login godoc
// @Summary      Login user
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      lib.RegisterRequest  true  "User Credentials"
// @Success      200   {object}  lib.Response{result=map[string]string} "Success response with token"
// @Failure      400   {object}  lib.Response
// @Failure      404   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /auth/login [post]
func Login(c *gin.Context) {
	var user lib.RegisterRequest
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(400, lib.Response{
			Status: 400,
			Message: "bad request",
		})
	}

	dbUser, err := models.FindEmail(user.Email)
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
		"role":   dbUser.RoleName,
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
			"token":   tokenJWT,
			"role_id": dbUser.RoleId,
		},
	})

}

// ForgotPassword godoc
// @Summary      Request password reset
// @Description  Send a reset password link to the user's email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      lib.User  true  "User Email"
// @Success      200   {object}  lib.Response
// @Failure      400   {object}  lib.Response
// @Failure      404   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /auth/forgot-password [post]
func ForgotPassword(c *gin.Context) {
	var user lib.User
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(400, lib.Response{
			Status:  400,
			Message: "bad request",
		})
		return
	}

	dbUser, err := models.FindEmail(user.Email)
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

// ValidatePasswordReset godoc
// @Summary      Reset password
// @Description  Verify token and update user password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token   query     string  false  "Reset Token"
// @Param        reset   body      lib.ResetPasswordRequest  true  "New Password Details"
// @Success      200   {object}  lib.Response
// @Failure      400   {object}  lib.Response
// @Failure      404   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /auth/reset-password [post]
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