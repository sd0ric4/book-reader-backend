package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sd0ric4/book-reader-backend/app/config"
	"github.com/sd0ric4/book-reader-backend/app/database"
	"github.com/sd0ric4/book-reader-backend/app/models"
	"github.com/sd0ric4/book-reader-backend/app/services"
	"github.com/sd0ric4/book-reader-backend/app/utils"
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
	book.Tags = c.DefaultPostForm("tags", "")

	// 将tags转为json数组
	if book.Tags != "" {
		book.Tags = fmt.Sprintf(`["%s"]`, strings.Join(strings.Split(book.Tags, " "), `","`))
	}
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

	// 初始化 S3 客户端
	cfg := config.Config
	minioClient, err := services.NewMinioClient(cfg.S3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to S3"})
		return
	}

	// 上传文件到 S3
	objectName := file.Filename
	bucketName := cfg.S3.BucketName
	err = services.UploadFile(minioClient, bucketName, objectName, fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to S3"})
		return
	}

	// 如果文件是 EPUB，则提取封面
	if filepath.Ext(file.Filename) == ".epub" {
		// 创建临时文件来存储EPUB
		tempFile, err := os.CreateTemp("", "epub-*")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
			return
		}
		defer os.Remove(tempFile.Name()) // 清理临时文件

		// 写入EPUB内容到临时文件
		if _, err := tempFile.Write(fileBytes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write temporary file"})
			return
		}
		tempFile.Close()

		// 创建临时目录用于存储封面
		tempCoverDir, err := os.MkdirTemp("", "covers-*")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary directory"})
			return
		}
		defer os.RemoveAll(tempCoverDir) // 清理临时目录

		// 提取封面
		coverPath, err := utils.ExtractEpubCover(tempFile.Name(), tempCoverDir)
		if err != nil && err != utils.ErrNoCover {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to extract cover: %v", err)})
			return
		}

		// 如果成功提取了封面，上传封面到 S3
		if coverPath != "" {
			coverData, err := os.ReadFile(coverPath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read cover file"})
				return
			}

			// 生成封面的对象名称
			coverObjectName := fmt.Sprintf("covers/%s_cover%s",
				strings.TrimSuffix(objectName, filepath.Ext(objectName)),
				filepath.Ext(coverPath))

			// 上传封面到 S3
			err = services.UploadFile(minioClient, bucketName, coverObjectName, coverData)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload cover to S3"})
				return
			}

			// 设置封面URL
			book.CoverURL = fmt.Sprintf("http://%s/%s/%s", cfg.S3.Endpoint, bucketName, coverObjectName)
		}
	}

	// 如果有封面URL，但没有封面，则设置封面URL为空
	if book.CoverURL != "" && book.CoverURL == c.DefaultPostForm("cover_url", "") {
		book.CoverURL = ""
	}
	// 设置书籍的URL
	book.BookURL = fmt.Sprintf("http://%s/%s/%s", cfg.S3.Endpoint, bucketName, objectName)

	// 将书籍信息保存到数据库
	if err := models.CreateBook(database.MySQLDB, &book); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save book to database"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Book uploaded successfully",
		"book":    book,
	})
}

func SummarizeBook(c *gin.Context) {
	services.GetBookSummary(c)
}
