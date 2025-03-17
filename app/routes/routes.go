package routes

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sd0ric4/book-reader-backend/app/controllers"
)

type TextResponse struct {
	Text string `json:"text"`
}

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
	r.GET("/text", func(c *gin.Context) {
		// 读取文本文件
		content, err := os.ReadFile("../config/book.txt")
		// 如果读取失败，返回错误信息
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		response := TextResponse{
			Text: string(content),
		}
		c.JSON(http.StatusOK, response)
	})
}
