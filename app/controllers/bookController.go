package controllers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sd0ric4/book-reader-backend/app/config"
	"github.com/sd0ric4/book-reader-backend/app/database"
	"github.com/sd0ric4/book-reader-backend/app/models"
	"github.com/sd0ric4/book-reader-backend/app/services"
)

func GetBooks(c *gin.Context) {
	books, err := models.GetBooks(database.MySQLDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch books"})
		return
	}
	c.JSON(http.StatusOK, books)
}

func GetBookByID(c *gin.Context) {
	id := c.Param("id")
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}
	book, err := models.GetBookByID(database.MySQLDB, uint(uintID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	c.JSON(http.StatusOK, book)
}

func UpdateBook(c *gin.Context) {
	id := c.Param("id")
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	// 获取表单数据
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取书籍信息
	bookInDB, err := models.GetBookByID(database.MySQLDB, uint(uintID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// 更新书籍信息
	if book.Title != "" {
		bookInDB.Title = book.Title
	}
	if book.Author != "" {
		bookInDB.Author = book.Author
	}
	if book.Description != "" {
		bookInDB.Description = book.Description
	}
	if book.CoverURL != "" {
		bookInDB.CoverURL = book.CoverURL
	}

	// 保存更新后的书籍信息
	if err := models.UpdateBook(database.MySQLDB, bookInDB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "Book updated successfully", "book": bookInDB})
}

func DeleteBook(c *gin.Context) {
	id := c.Param("id")
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	// 删除书籍
	if err := models.DeleteBook(database.MySQLDB, uint(uintID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete book"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
}

func UploadBook(c *gin.Context) {
	// 绑定表单字段到book结构体
	var book models.Book

	// 获取表单数据并绑定到book结构体
	book.Title = c.DefaultPostForm("title", "")
	book.Author = c.DefaultPostForm("author", "")
	book.Description = c.DefaultPostForm("description", "")
	book.CoverURL = c.DefaultPostForm("cover_url", "")

	// 判断必填字段是否为空
	if book.Title == "" || book.Author == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and Author are required"})
		return
	}

	// 获取文件数据
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// 打开文件
	fileData, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileData.Close()

	// 读取文件内容
	fileBytes, err := io.ReadAll(fileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// 上传文件到S3
	cfg := config.Config
	minioClient, err := services.NewMinioClient(cfg.S3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to S3"})
		return
	}

	objectName := file.Filename
	bucketName := cfg.S3.BucketName
	err = services.UploadFile(minioClient, bucketName, objectName, fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to S3"})
		return
	}

	// 设置书籍的URL
	book.BookURL = fmt.Sprintf("https://%s/%s/%s", cfg.S3.Endpoint, bucketName, objectName)

	// 如果成功，将书籍信息保存到数据库，否则返回错误响应
	if err := models.CreateBook(database.MySQLDB, &book); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload book"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "Book uploaded successfully", "book": book})
}

func SummarizeBook(c *gin.Context) {
	services.GetBookSummary(c)
}
