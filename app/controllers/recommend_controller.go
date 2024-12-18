package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sd0ric4/book-reader-backend/app/models"
	"github.com/sd0ric4/book-reader-backend/app/services"
)

// RecommendBooksHandler 处理书籍推荐的HTTP请求
func RecommendBooksHandler(c *gin.Context) {
	// 解析请求数据
	var reqBody models.RecommendationRequest

	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(reqBody.UserBooks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User books are required"})
		return
	}

	// 调用推荐服务
	recommendations, err := services.RecommendBooks(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Recommendation failed"})
		return
	}

	// 返回推荐结果
	c.JSON(http.StatusOK, gin.H{
		"user_id":         reqBody.UserID,
		"recommendations": recommendations,
	})
}
