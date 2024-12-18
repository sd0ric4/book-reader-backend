package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sd0ric4/book-reader-backend/app/controllers"
)

func SetupRoutes(r *gin.Engine) {
	// Upload
	r.POST("/books/upload", controllers.UploadBook)
	r.POST("/books/summarize", controllers.SummarizeBook)
	// Recommendation
	r.POST("/books/recommend", controllers.RecommendBooksHandler)
	// Book related routes
	r.GET("/books/list", controllers.GetBooks)
	r.GET("/books/:id", controllers.GetBookByID)
	r.PUT("/books/:id", controllers.UpdateBook)
	r.DELETE("/books/:id", controllers.DeleteBook)

	// User related routes
	r.POST("/users/register", controllers.Register)
	r.POST("/users/login", controllers.Login)
	r.PUT("/users/change-password", controllers.ChangePassword)

}
