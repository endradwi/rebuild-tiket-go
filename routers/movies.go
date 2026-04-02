package routers

import (
	"tiket/controllers"
	"tiket/middleware"

	"github.com/gin-gonic/gin"
)

func MovieRouter(router *gin.RouterGroup) {
	movieRoutes := router.Group("/movies")

	// Allow public access to getting all movies or getting a specific movie
	movieRoutes.GET("", controllers.GetAllMovies)
	movieRoutes.GET("/:id", controllers.GetMovieById)
	movieRoutes.GET("/genres", controllers.GetGenres)
	movieRoutes.GET("/casters", controllers.GetCasters)
	movieRoutes.GET("/cinemas", controllers.GetCinemas)

	// Admin protected routes
	protectedRoutes := movieRoutes.Group("")
	protectedRoutes.Use(middleware.AuthMiddleware())
	protectedRoutes.POST("", controllers.CreateMovie)
	protectedRoutes.PATCH("/:id", controllers.UpdateMovie)
	protectedRoutes.DELETE("/:id", controllers.DeleteMovie)
}
