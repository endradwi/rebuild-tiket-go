package controllers

import (
	"net/http"
	"strconv"
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

// GetProfile gets the current logged-in user's profile
// GetProfile godoc
// @Summary      Get user profile
// @Description  Retrieve the profile of the currently logged-in user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  lib.Response{result=lib.UserProfile}
// @Failure      401   {object}  lib.Response
// @Failure      404   {object}  lib.Response
// @Router       /users/profile [get]
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

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update the profile of the currently logged-in user
// @Tags         users
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        fullName      formData  string  false  "Full Name"
// @Param        email         formData  string  false  "Email"
// @Param        phoneNumber   formData  string  false  "Phone Number"
// @Param        image         formData  file    false  "Profile Image"
// @Success      200   {object}  lib.Response{result=lib.UserProfile}
// @Failure      400   {object}  lib.Response
// @Failure      401   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /users/profile [patch]
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

	// Handle the file upload manually if present
	file, err := c.FormFile("image")
	if err == nil {
		// 1. Validate File Size (e.g., max 5MB)
		const maxFileSize = 5 << 20 // 5 MB
		if file.Size > maxFileSize {
			c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "File size exceeds 5MB limit"})
			return
		}

		// 2. Validate File Type (must be an image)
		// We open the file to read its first 512 bytes for content-type detection
		openedFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to process image"})
			return
		}
		defer openedFile.Close()

		buffer := make([]byte, 512)
		_, err = openedFile.Read(buffer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to read image"})
			return
		}

		// Reset the read pointer
		openedFile.Seek(0, 0)

		contentType := http.DetectContentType(buffer)
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
			c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid file type. Only JPEG, PNG, and WebP are allowed"})
			return
		}

		// create uploads directory if not exists
		// Create a unique filename
		filename := "uploads/profile-" + strconv.Itoa(userId) + "-" + file.Filename
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to save image"})
			return
		}
		
		// Set the Image URL in request, use a pointer
		imageUrl := "/" + filename
		req.Image = &imageUrl
	}

	updatedProfile, err := models.UpdateUserProfile(userId, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: updatedProfile})
}

// GetAllUsers gets all users (Admin or general CRUD)
// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieve a list of all users (admin access required)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  lib.Response{result=[]lib.UserProfile}
// @Failure      500   {object}  lib.Response
// @Router       /users [get]
func GetAllUsers(c *gin.Context) {
	users, err := models.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: users})
}

// GetUserById godoc
// @Summary      Get user by ID
// @Description  Retrieve a user profile by their ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int     true  "User ID"
// @Success      200   {object}  lib.Response{result=lib.UserProfile}
// @Failure      400   {object}  lib.Response
// @Failure      404   {object}  lib.Response
// @Router       /users/{id} [get]
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

// DeleteUser godoc
// @Summary      Delete user
// @Description  Remove a user account by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int     true  "User ID"
// @Success      200   {object}  lib.Response
// @Failure      400   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /users/{id} [delete]
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

// GetUserOrders godoc
// @Summary      Get user orders
// @Description  Retrieve the order history of the currently logged-in user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  lib.Response{result=[]lib.Order}
// @Failure      401   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /users/history [get]
func GetUserOrders(c *gin.Context) {
	userIdAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, lib.Response{Status: 401, Message: "Unauthorized"})
		return
	}

	userId := int(userIdAny.(float64))
	orders, err := models.GetUserOrders(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to fetch orders: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: orders})
}
