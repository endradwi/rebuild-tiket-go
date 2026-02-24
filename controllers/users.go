package controllers

import (
	"net/http"
	"strconv"
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

// GetProfile gets the current logged-in user's profile
func GetProfile(c *gin.Context) {
	userIdAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, lib.Response{Status: 401, Message: "Unauthorized"})
		return
	}

	userId := int(userIdAny.(float64))
	profile, err := models.GetUserProfile(userId)
	if err != nil {
		c.JSON(http.StatusNotFound, lib.Response{Status: 404, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: profile})
}

// UpdateProfile updates the current logged-in user's profile (using PATCH method approach)
func UpdateProfile(c *gin.Context) {
	userIdAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, lib.Response{Status: 401, Message: "Unauthorized"})
		return
	}

	userId := int(userIdAny.(float64))

	var req lib.ProfileUpdateRequest
	// Use ShouldBind to bind form data (multipart/form-data support for PATCH/POST)
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Bad Request: " + err.Error()})
		return
	}

	updatedProfile, err := models.UpdateUserProfile(userId, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: updatedProfile})
}

// GetAllUsers gets all users (Admin or general CRUD)
func GetAllUsers(c *gin.Context) {
	users, err := models.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: users})
}

// GetUserById gets a user profile by their ID param
func GetUserById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid user ID"})
		return
	}

	profile, err := models.GetUserProfile(id)
	if err != nil {
		c.JSON(http.StatusNotFound, lib.Response{Status: 404, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: profile})
}


// DeleteUser deletes a user by ID
func DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid user ID"})
		return
	}

	err = models.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "User deleted successfully"})
}
