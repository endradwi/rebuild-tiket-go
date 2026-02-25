package controllers

import (
	"net/http"
	"strconv"
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

// CreateMovie handles creating a new movie
func CreateMovie(c *gin.Context) {
	var req lib.MovieCreateRequest
	
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

		openedFile.Seek(0, 0)

		contentType := http.DetectContentType(buffer)
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
			c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid file type. Only JPEG, PNG, and WebP are allowed"})
			return
		}

		// Create a unique filename
		filename := "uploads/movie-" + strconv.FormatInt(file.Size, 10) + "-" + file.Filename
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to save image"})
			return
		}
		
		imageUrl := "/" + filename
		req.Image = &imageUrl
	}

	movie, err := models.CreateMovie(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to create movie: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lib.Response{Status: 201, Message: "Movie created successfully", Result: movie})
}

// GetAllMovies handles retrieving all movies
func GetAllMovies(c *gin.Context) {
	var params lib.MovieQueryParams

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid query parameters"})
		return
	}

	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	movies, pageInfo, err := models.GetAllMovies(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to fetch movies: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{
		Status:   200,
		Message:  "success",
		Result:   movies,
		PageInfo: &pageInfo,
	})
}

// GetMovieById handles retrieving a single movie by ID
func GetMovieById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid movie ID"})
		return
	}

	movie, err := models.GetMovieById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, lib.Response{Status: 404, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: movie})
}

// UpdateMovie handles updating a movie using PATCH fields
func UpdateMovie(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid movie ID"})
		return
	}

	var req lib.MovieUpdateRequest
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

		openedFile.Seek(0, 0)

		contentType := http.DetectContentType(buffer)
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
			c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid file type. Only JPEG, PNG, and WebP are allowed"})
			return
		}

		// Create a unique filename
		filename := "uploads/movie-" + strconv.FormatInt(file.Size, 10) + "-" + file.Filename
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to save image"})
			return
		}
		
		imageUrl := "/" + filename
		req.Image = &imageUrl
	}

	updatedMovie, err := models.UpdateMovie(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to update movie: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "Movie updated successfully", Result: updatedMovie})
}

// DeleteMovie handles deleting a movie by ID
func DeleteMovie(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid movie ID"})
		return
	}

	err = models.DeleteMovie(id)
	if err != nil {
		c.JSON(http.StatusNotFound, lib.Response{Status: 404, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "Movie deleted successfully"})
}