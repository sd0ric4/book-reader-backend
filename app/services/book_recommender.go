package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sd0ric4/book-reader-backend/app/database"
	"github.com/sd0ric4/book-reader-backend/app/models"
	"gonum.org/v1/gonum/mat"
)

// RecommendBooks 推荐书籍的核心逻辑
func RecommendBooks(req models.RecommendationRequest) ([]models.Book, error) {
	// 从数据库中获取书籍简要信息
	bookBriefs, err := models.GetBookBriefs(database.MySQLDB)
	if err != nil {
		return nil, fmt.Errorf("failed to get book briefs: %v", err)
	}

	// 检查用户书籍是否为空
	if len(req.UserBooks) == 0 {
		return nil, fmt.Errorf("no user books provided for recommendation")
	}

	// 提取用户阅读书籍的标签
	userTags := make([]string, len(req.UserBooks))
	for i, book := range req.UserBooks {
		userTags[i] = book.Tags
	}
	userTagString := strings.Join(userTags, " ")

	// 准备所有书籍的标签（包括用户标签）
	allTagDocuments := make([]string, len(bookBriefs)+1)
	for i, book := range bookBriefs {
		allTagDocuments[i] = book.Tags
	}
	allTagDocuments[len(bookBriefs)] = userTagString

	// 计算TF-IDF矩阵
	tfidfMatrix := ComputeTFIDF(allTagDocuments)

	// 获取用户标签向量（最后一行）
	userVector := mat.NewVecDense(len(tfidfMatrix.RawRowView(len(bookBriefs))), tfidfMatrix.RawRowView(len(bookBriefs)))

	// 计算相似度
	similarities := make([]models.Book, len(bookBriefs))
	for i, book := range bookBriefs {
		bookVector := mat.NewVecDense(len(tfidfMatrix.RawRowView(i)), tfidfMatrix.RawRowView(i))
		similarity := CosineSimilarity(userVector, bookVector)
		similarities[i] = models.Book{
			ID:          uint(book.ID),
			Title:       book.Title,
			Author:      book.Author,
			Description: book.Description,
			Tags:        book.Tags,
			CoverURL:    book.CoverURL,
			Score:       similarity,
			CreatedAt:   book.CreatedAt,
			UpdatedAt:   book.UpdatedAt,
		}
	}

	// 按相似度排序
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Score > similarities[j].Score
	})

	// 返回前5个推荐
	return similarities[:5], nil
}
