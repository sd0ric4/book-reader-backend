package models

import (
	"time"

	"gorm.io/gorm"
)

// BookChapter 表示书籍章节信息
type BookChapter struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	BookID           uint      `gorm:"index" json:"book_id"`
	ChapterStructure string    `gorm:"type:json;not null" json:"chapter_structure"`
	ChapterName      string    `gorm:"size:255;not null" json:"chapter_name"`
	ChapterContent   string    `gorm:"type:text" json:"chapter_content"`
	ContentPath      string    `gorm:"size:255" json:"content_path"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// 获取所有章节
func GetBookChapters(db *gorm.DB) ([]BookChapter, error) {
	var chapters []BookChapter
	if err := db.Find(&chapters).Error; err != nil {
		return nil, err
	}
	return chapters, nil
}

// 根据书籍ID获取所有章节
func GetChaptersByBookID(db *gorm.DB, bookID uint) ([]BookChapter, error) {
	var chapters []BookChapter
	if err := db.Where("book_id = ?", bookID).Find(&chapters).Error; err != nil {
		return nil, err
	}
	return chapters, nil
}

// 根据ID获取章节
func GetChapterByID(db *gorm.DB, id uint) (*BookChapter, error) {
	var chapter BookChapter
	if err := db.First(&chapter, id).Error; err != nil {
		return nil, err
	}
	return &chapter, nil
}

// 创建章节
func CreateChapter(db *gorm.DB, chapter *BookChapter) error {
	return db.Create(chapter).Error
}

// 批量创建章节
func CreateChapters(db *gorm.DB, chapters []BookChapter) error {
	return db.Create(&chapters).Error
}

// 更新章节
func UpdateChapter(db *gorm.DB, chapter *BookChapter) error {
	return db.Save(chapter).Error
}

// 删除章节
func DeleteChapter(db *gorm.DB, id uint) error {
	return db.Delete(&BookChapter{}, id).Error
}

// 删除书籍的所有章节
func DeleteBookChapters(db *gorm.DB, bookID uint) error {
	return db.Where("book_id = ?", bookID).Delete(&BookChapter{}).Error
}

// 根据书籍ID和章节名称查找章节
func GetChapterByBookIDAndName(db *gorm.DB, bookID uint, chapterName string) (*BookChapter, error) {
	var chapter BookChapter
	if err := db.Where("book_id = ? AND chapter_name = ?", bookID, chapterName).First(&chapter).Error; err != nil {
		return nil, err
	}
	return &chapter, nil
}

// 获取书籍的章节数量
func GetChapterCount(db *gorm.DB, bookID uint) (int64, error) {
	var count int64
	if err := db.Model(&BookChapter{}).Where("book_id = ?", bookID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// 更新章节内容
func UpdateChapterContent(db *gorm.DB, chapterID uint, content string) error {
	return db.Model(&BookChapter{}).Where("id = ?", chapterID).Update("chapter_content", content).Error
}

// 更新章节结构
func UpdateChapterStructure(db *gorm.DB, chapterID uint, structure string) error {
	return db.Model(&BookChapter{}).Where("id = ?", chapterID).Update("chapter_structure", structure).Error
}

// 批量获取指定章节ID的章节
func GetChaptersByIDs(db *gorm.DB, ids []uint) ([]BookChapter, error) {
	var chapters []BookChapter
	if err := db.Where("id IN ?", ids).Find(&chapters).Error; err != nil {
		return nil, err
	}
	return chapters, nil
}

// 检查章节是否存在
func ChapterExists(db *gorm.DB, bookID uint, chapterName string) (bool, error) {
	var count int64
	if err := db.Model(&BookChapter{}).Where("book_id = ? AND chapter_name = ?", bookID, chapterName).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
