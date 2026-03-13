package controllers

import (
	"net/http"
	"strconv"
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

// CreateMovie godoc
// @Summary      Create a new movie
// @Description  Insert a new movie into the database with its many-to-many relations and showtimes
// @Tags         movies
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        title           formData  string  true   "Movie Title"
// @Param        released_at     formData  string  false  "Release Date (RFC3339)"
// @Param        recommendation  formData  bool    false  "Is Recommended"
// @Param        duration        formData  string  false  "Duration"
// @Param        synopsis        formData  string  false  "Synopsis"
// @Param        director_name   formData  string  false  "Director Name"
// @Param        genre_ids       formData  []int   false  "Genre IDs"
// @Param        caster_ids      formData  []int   false  "Caster IDs"
// @Param        cinema_ids      formData  []int   false  "Cinema IDs"
// @Param        image           formData  file    false  "Movie Poster"
// @Success      201   {object}  lib.Response{result=lib.Movie}
// @Failure      400   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /movies [post]
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

// GetAllMovies godoc
// @Summary      Get all movies
// @Description  Retrieve a paginated list of movies with optional filters
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        page    query     int     false  "Page number"
// @Param        limit   query     int     false  "Items per page"
// @Param        search  query     string  false  "Search by title"
// @Param        sort    query     string  false  "Sort order (asc/desc)"
// @Param        month   query     int     false  "Filter by release month"
// @Param        year    query     int     false  "Filter by release year"
// @Success      200   {object}  lib.ListResponse{result=[]lib.Movie}
// @Failure      400   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /movies [get]
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

	c.JSON(http.StatusOK, lib.ListResponse{
		Status:   200,
		Message:  "success",
		Result:   movies,
		PageInfo: &pageInfo,
	})
}

// GetMovieById godoc
// @Summary      Get movie by ID
// @Description  Retrieve detailed information about a single movie with consolidated and filtered showtimes
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        id           path      int     true   "Movie ID"
// @Param        location_id  query     int     false  "Filter showtimes by Location ID"
// @Param        date         query     string  false  "Filter showtimes by Date (YYYY-MM-DD)"
// @Param        time         query     string  false  "Filter showtimes by minimum Time (HH:MM:SS)"
// @Success      200   {object}  lib.Response{result=lib.Movie}
// @Failure      400   {object}  lib.Response
// @Failure      404   {object}  lib.Response
// @Router       /movies/{id} [get]
func GetMovieById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid movie ID"})
		return
	}

	var params lib.MovieDetailParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid filter parameters"})
		return
	}

	movie, err := models.GetMovieById(id, params)
	if err != nil {
		c.JSON(http.StatusNotFound, lib.Response{Status: 404, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: movie})
}

// UpdateMovie godoc
// @Summary      Update movie
// @Description  Update movie details by ID
// @Tags         movies
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id              path      int     true   "Movie ID"
// @Param        title           formData  string  false  "Movie Title"
// @Param        released_at     formData  string  false  "Release Date (RFC3339)"
// @Param        recommendation  formData  bool    false  "Is Recommended"
// @Param        duration        formData  string  false  "Duration"
// @Param        synopsis        formData  string  false  "Synopsis"
// @Param        director_name   formData  string  false  "Director Name"
// @Param        genre_ids       formData  []int   false  "Genre IDs"
// @Param        caster_ids      formData  []int   false  "Caster IDs"
// @Param        cinema_ids      formData  []int   false  "Cinema IDs"
// @Param        image           formData  file    false  "Movie Poster"
// @Success      200   {object}  lib.Response{result=lib.Movie}
// @Failure      400   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /movies/{id} [patch]
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

// DeleteMovie godoc
// @Summary      Delete movie
// @Description  Remove a movie from the database by ID
// @Tags         movies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int     true  "Movie ID"
// @Success      200   {object}  lib.Response
// @Failure      400   {object}  lib.Response
// @Failure      404   {object}  lib.Response
// @Router       /movies/{id} [delete]
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

