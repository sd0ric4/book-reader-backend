package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type Character struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type BookReviewData struct {
	Title      string      `json:"title"`
	Author     string      `json:"author"`
	Characters []Character `json:"characters"`
	Synopsis   string      `json:"synopsis"`
}

// 新增一个包装类型
type BookReviewDataList struct {
	Items []BookReviewData
}

// 将 Value 方法移动到新的包装类型上
func (rd BookReviewDataList) Value() (driver.Value, error) {
	return json.Marshal(rd.Items)
}

// 将 Scan 方法移动到新的包装类型上
func (rd *BookReviewDataList) Scan(value interface{}) error {
	if value == nil {
		return errors.New("scanning null value")
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source")
	}

	return json.Unmarshal(bytes, &rd.Items)
}

// 修改 Review 结构体
type Review struct {
	ID         uint64             `json:"id" gorm:"primaryKey"`
	UserID     uint64             `json:"user_id"`
	BookID     uint64             `json:"book_id"`
	ReviewData BookReviewDataList `json:"review_data" gorm:"type:json"`
}

func GetReviews(db *gorm.DB) ([]Review, error) {
	var reviews []Review
	if err := db.Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

// 修改 GetReviewByTitle 函数
func GetReviewByTitle(db *gorm.DB, title string) (*Review, error) {
	var review Review
	if err := db.Where("JSON_EXTRACT(review_data, '$[0].title') = ?", title).First(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func UpdateReview(db *gorm.DB, review *Review) error {
	if err := db.Save(review).Error; err != nil {
		return err
	}
	return nil
}

func CreateReview(db *gorm.DB, review *Review) error {
	if err := db.Create(review).Error; err != nil {
		return err
	}
	return nil
}

func DeleteReview(db *gorm.DB, review *Review) error {
	if err := db.Delete(review).Error; err != nil {
		return err
	}
	return nil
}
