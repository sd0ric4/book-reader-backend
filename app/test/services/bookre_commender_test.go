package services_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sd0ric4/book-reader-backend/app/config"
	"github.com/sd0ric4/book-reader-backend/app/database"
	"github.com/sd0ric4/book-reader-backend/app/models"
	"github.com/sd0ric4/book-reader-backend/app/services"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/mat"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.DebugMode)
	config.Config = &config.ConfigStruct{
		MySQL: config.MySQLConfig{
			User:     "root",
			Password: "root",
			Host:     "192.168.0.118",
			Port:     3306,
			DBName:   "bookdb",
		},
	}
	database.InitMySQL()
	m.Run()
}

// 测试推荐书籍功能
func TestRecommendBooks(t *testing.T) {
	testCases := []struct {
		name           string
		userBooks      []models.Book
		expectedLength int
		expectedError  bool
		errorMessage   string
	}{
		{
			name: "Recommend based on fiction mystery",
			userBooks: []models.Book{
				{ID: 1, Tags: "fiction  mystery"},
			},
			expectedLength: 5,
			expectedError:  false,
		},
		{
			name: "Recommend with multiple user books",
			userBooks: []models.Book{
				{ID: 1, Tags: "fiction adventure mystery"},
				{ID: 2, Tags: "romance drama"},
				{ID: 3, Tags: "fiction  mystery"},
			},
			expectedLength: 5,
			expectedError:  false,
		},
		{
			name:           "Empty user books should fail",
			userBooks:      []models.Book{},
			expectedLength: 0,
			expectedError:  true,
			errorMessage:   "no user books provided for recommendation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := models.RecommendationRequest{
				UserID:    123,
				UserBooks: tc.userBooks,
			}

			recommendations, err := services.RecommendBooks(req)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorMessage != "" {
					assert.EqualError(t, err, tc.errorMessage)
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, recommendations, tc.expectedLength)
			// 验证推荐结果按相似度降序排列
			for i := 1; i < len(recommendations); i++ {
				t.Logf("Rank %d: %s (%f) ", i-1, recommendations[i-1].Title, recommendations[i-1].Score)
				assert.GreaterOrEqual(t, recommendations[i-1].Score, recommendations[i].Score)
			}
		})
	}
}

// 测试 ComputeTFIDF 函数
func TestComputeTFIDF(t *testing.T) {
	testCases := []struct {
		name      string
		documents []string
		minRows   int
		minCols   int
	}{
		{
			name: "Basic TF-IDF computation",
			documents: []string{
				"fiction adventure mystery",
				"romance drama",
				"history biography",
			},
			minRows: 3,
			minCols: 5,
		},
		{
			name: "Single document",
			documents: []string{
				"fiction adventure mystery",
			},
			minRows: 1,
			minCols: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tfidfMatrix := services.ComputeTFIDF(tc.documents)

			assert.NotNil(t, tfidfMatrix)
			assert.GreaterOrEqual(t, tfidfMatrix.RawMatrix().Rows, tc.minRows)
			assert.GreaterOrEqual(t, tfidfMatrix.RawMatrix().Cols, tc.minCols)
		})
	}
}

// 测试余弦相似度计算
func TestCosineSimilarity(t *testing.T) {
	testCases := []struct {
		name           string
		vectorA        []float64
		vectorB        []float64
		expectedResult float64
	}{
		{
			name:           "Identical vectors",
			vectorA:        []float64{1, 2, 3},
			vectorB:        []float64{1, 2, 3},
			expectedResult: 1.0,
		},
		{
			name:           "Orthogonal vectors",
			vectorA:        []float64{1, 0, 0},
			vectorB:        []float64{0, 1, 0},
			expectedResult: 0.0,
		},
		{
			name:           "Opposite vectors",
			vectorA:        []float64{1, 2, 3},
			vectorB:        []float64{-1, -2, -3},
			expectedResult: -1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vecA := mat.NewVecDense(len(tc.vectorA), tc.vectorA)
			vecB := mat.NewVecDense(len(tc.vectorB), tc.vectorB)

			similarity := services.CosineSimilarity(vecA, vecB)
			assert.InDelta(t, tc.expectedResult, similarity, 0.0001)
		})
	}
}

// 基准测试 RecommendBooks 性能
func BenchmarkRecommendBooks(b *testing.B) {
	req := models.RecommendationRequest{
		UserID: 123,
		UserBooks: []models.Book{
			{ID: 1, Tags: "fiction adventure mystery"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = services.RecommendBooks(req)
	}
}
