package models

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Book struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"not null;size:255;index" json:"title"`
	Author      string    `gorm:"not null;size:100;index" json:"author"`
	BookURL     string    `gorm:"not null;size:255" json:"book_url"`
	Description string    `gorm:"type:text" json:"description"`
	CoverURL    string    `gorm:"size:255" json:"cover_url"`
	Format      string    `gorm:"size:50" json:"format"`
	Tags        string    `gorm:"type:json" json:"tags"`
	Score       float64   `json:"score,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type RecommendationRequest struct {
	UserID    int    `json:"user_id"`
	UserBooks []Book `json:"user_books"`
}

// 获取书籍列表
func GetBooks(db *gorm.DB) ([]Book, error) {
	var books []Book
	if err := db.Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}

// 根据ID获取书籍
func GetBookByID(db *gorm.DB, id uint) (*Book, error) {
	var book Book
	if err := db.First(&book, id).Error; err != nil {
		return nil, err
	}
	return &book, nil
}

// 更新书籍
func UpdateBook(db *gorm.DB, book *Book) error {
	if err := db.Save(book).Error; err != nil {
		return err
	}
	return nil
}

// 创建书籍
func CreateBook(db *gorm.DB, book *Book) error {
	if err := db.Create(book).Error; err != nil {
		return err
	}
	return nil
}

// 删除书籍
func DeleteBook(db *gorm.DB, id uint) error {
	if err := db.Delete(&Book{}, id).Error; err != nil {
		return err
	}
	return nil
}

// 获取书籍简要信息
func GetBookBriefs(db *gorm.DB) ([]Book, error) {
	var books []Book
	if err := db.Find(&books).Error; err != nil {
		return nil, err
	}

	var bookBriefs []Book
	for _, book := range books {
		var tags []string
		if err := json.Unmarshal([]byte(book.Tags), &tags); err != nil {
			return nil, err
		}
		bookBriefs = append(bookBriefs, Book{
			ID:          book.ID,
			Title:       book.Title,
			Author:      book.Author,
			Description: book.Description,
			CoverURL:    book.CoverURL,
			Format:      book.Format,
			Tags:        strings.Join(tags, " "),
			CreatedAt:   book.CreatedAt,
			UpdatedAt:   book.UpdatedAt,
		})
	}

	return bookBriefs, nil
}

// 更新标签的方法
func UpdateBookTags(db *gorm.DB, bookID uint, newTags []string) error {
	// 将新标签转换为 JSON 字符串
	tagsJSON, err := json.Marshal(newTags)
	if err != nil {
		return err
	}

	// 更新数据库中的标签
	result := db.Model(&Book{}).Where("id = ?", bookID).Update("tags", string(tagsJSON))
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// 添加标签的方法
func AddBookTag(db *gorm.DB, bookID uint, newTag string) error {
	// 先获取当前书籍
	var book Book
	if err := db.First(&book, bookID).Error; err != nil {
		return err
	}

	// 解析当前标签
	var tags []string
	if err := json.Unmarshal([]byte(book.Tags), &tags); err != nil {
		return err
	}

	// 检查标签是否已存在
	for _, tag := range tags {
		if tag == newTag {
			return nil // 标签已存在，无需重复添加
		}
	}

	// 添加新标签
	tags = append(tags, newTag)

	// 将更新后的标签转换为 JSON
	updatedTagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}

	// 更新数据库
	result := db.Model(&Book{}).Where("id = ?", bookID).Update("tags", string(updatedTagsJSON))
	return result.Error
}

// 删除特定标签的方法
func RemoveBookTag(db *gorm.DB, bookID uint, tagToRemove string) error {
	// 先获取当前书籍
	var book Book
	if err := db.First(&book, bookID).Error; err != nil {
		return err
	}

	// 解析当前标签
	var tags []string
	if err := json.Unmarshal([]byte(book.Tags), &tags); err != nil {
		return err
	}

	// 创建新的标签切片（不包含要删除的标签）
	var updatedTags []string
	for _, tag := range tags {
		if tag != tagToRemove {
			updatedTags = append(updatedTags, tag)
		}
	}

	// 将更新后的标签转换为 JSON
	updatedTagsJSON, err := json.Marshal(updatedTags)
	if err != nil {
		return err
	}

	// 更新数据库
	result := db.Model(&Book{}).Where("id = ?", bookID).Update("tags", string(updatedTagsJSON))
	return result.Error
}
