package routes

import (
	"day/handlers"
	"day/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	public := r.Group("/api")

	{
		public.GET("/books", handlers.GetBooks)
		public.GET("/books/pages/:count", handlers.PaginateBooks)
		public.GET("/books/:id", handlers.GetBookByID)
		public.POST("/register", handlers.Register)
		public.POST("/login", handlers.Login)
		public.GET("/profile/:id", handlers.GetProfile)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/create-book", handlers.CreateBook)
		api.DELETE("/books/:id", handlers.DeleteBook)
		api.PUT("/profile", handlers.UpdateProfile)
	}

}
