package routers

import (
	"tiket/controllers"

	"github.com/gin-gonic/gin"
)

func MovieRouter(router *gin.RouterGroup) {
	movieRoutes := router.Group("/movies")

	// Allow public access to getting all movies or getting a specific movie
	movieRoutes.GET("", controllers.GetAllMovies)
	movieRoutes.GET("/:id", controllers.GetMovieById)

	// Admin protected routes can be added later, for now we keep it standard
	// You might want to wrap these in `middleware.AuthMiddleware()` if needed
	movieRoutes.POST("", controllers.CreateMovie)
	movieRoutes.PATCH("/:id", controllers.UpdateMovie)
	movieRoutes.DELETE("/:id", controllers.DeleteMovie)
}
