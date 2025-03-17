package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sd0ric4/book-reader-backend/app/database"
	"github.com/sd0ric4/book-reader-backend/app/models"
)

type AiConfig struct {
	BaseURL   string
	ModelName string
	APIKey    string
}

var aiconfig = AiConfig{
	BaseURL:   "http://172.16.26.101:13300/v1",
	ModelName: "SparkDesk-v3.5",
	APIKey:    "sk-cf3Qx1Rm3KnWq0kffEdPhbyNrF3xwYLXtQCycLPGgnzVgpZg", // 请替换为实际的 API Key
}

var PROMPT_TEMPLATE = `
请总结以下书籍的主要人物和主要内容概述，并以JSON格式输出：

{
    "title": "%s",
    "author": "%s",
    "characters": [
        {
            "name": "人物1",
            "role": "描述其在书中的角色或作用。"
        },
        {
            "name": "人物2",
            "role": "描述其在书中的角色或作用。"
        }
        （根据书籍内容添加更多主要人物。）
    ],
    "synopsis": "请简要概述书籍的主要故事情节、主题和核心思想，控制在150字以内。"
}

示例：
{
    "title": "红楼梦",
    "author": "曹雪芹",
    "characters": [
        {
            "name": "贾宝玉",
            "role": "贾府的公子，具有叛逆性格，是书中主要叙事线的中心。"
        },
        {
            "name": "林黛玉",
            "role": "一位聪明且感性的女性，与贾宝玉有深厚感情。"
        },
        {
            "name": "薛宝钗",
            "role": "性格温和，与贾宝玉的关系构成三角恋。"
        }
    ],
    "synopsis": "《红楼梦》以贾宝玉、林黛玉的爱情悲剧为线索，以贾、史、王、薛四大家族的兴衰为背景，着重叙述了贾家荣、宁两府逐渐衰败的过程。广泛地反映了当时的社会现象和各种矛盾，揭露了封建官僚地主家庭的荒淫腐败、虚伪欺诈及其各种罪恶活动，歌颂了贾宝玉、林黛玉的封建叛逆精神，描绘了一些纯洁少女的悲惨遭遇和反抗性格，对封建末期社会进行了剖析和批判，揭示了封建社会必然灭亡的历史趋势。"
}

现在请总结 %s 的《%s》：
`

type RequestData struct {
	BookTitle string `json:"book_title"`
	Author    string `json:"author"`
}

type ClientOption func(*openai.ClientConfig)

func WithBaseURL(url string) ClientOption {
	return func(config *openai.ClientConfig) {
		config.BaseURL = url
	}
}

func NewOpenAIClient(apiKey string, options ...ClientOption) *openai.Client {
	config := openai.DefaultConfig(apiKey)

	for _, option := range options {
		option(&config)
	}

	return openai.NewClientWithConfig(config)
}

// 添加新的验证函数
func isValidBookReview(data map[string]interface{}) bool {
	// 检查必要字段是否存在且不为空
	if title, ok := data["title"].(string); !ok || title == "" {
		return false
	}
	if author, ok := data["author"].(string); !ok || author == "" {
		return false
	}
	if synopsis, ok := data["synopsis"].(string); !ok || synopsis == "" {
		return false
	}

	// 检查characters数组
	characters, ok := data["characters"].([]interface{})
	if !ok || len(characters) == 0 {
		return false
	}

	// 检查每个character的完整性
	for _, char := range characters {
		character, ok := char.(map[string]interface{})
		if !ok {
			return false
		}
		if name, ok := character["name"].(string); !ok || name == "" || name == "人物1" || name == "人物2" {
			return false
		}
		if role, ok := character["role"].(string); !ok || role == "" {
			return false
		}
	}

	return true
}

// 修改 GetBookSummary 函数
func GetBookSummary(c *gin.Context) {
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 先从数据库查询
	db := database.MySQLDB
	review, err := models.GetReviewByTitle(db, requestData.BookTitle)

	// 如果数据库中存在且数据有效，直接返回
	if err == nil && review != nil && len(review.ReviewData.Items) > 0 {
		c.JSON(http.StatusOK, gin.H{"summary": review.ReviewData.Items[0]})
		return
	}

	// 最大重试次数
	maxRetries := 3
	var jsonData map[string]interface{}

	for i := 0; i < maxRetries; i++ {
		// AI 调用逻辑
		client := NewOpenAIClient(aiconfig.APIKey, WithBaseURL(aiconfig.BaseURL))
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: aiconfig.ModelName,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: fmt.Sprintf(PROMPT_TEMPLATE, requestData.BookTitle, requestData.Author, requestData.Author, requestData.BookTitle),
					},
				},
				MaxTokens:   500,
				Temperature: 0.7,
			},
		)

		if err != nil {
			continue
		}

		if len(resp.Choices) > 0 {
			jsonData, err = extractJSON(resp.Choices[0].Message.Content)
			if err != nil {
				continue
			}

			// 验证数据完整性
			if isValidBookReview(jsonData) {
				characters := make([]models.Character, 0)
				// 从 jsonData["characters"] 中提取数据并填充 characters 切片

				for _, char := range jsonData["characters"].([]interface{}) {
					character := char.(map[string]interface{})
					characters = append(characters, models.Character{
						Name: character["name"].(string),
						Role: character["role"].(string),
					})
				}
				reviewData := []models.BookReviewData{{
					Title:      requestData.BookTitle,
					Author:     requestData.Author,
					Characters: characters,
					Synopsis:   jsonData["synopsis"].(string),
				}}

				newReview := &models.Review{
					BookID: 41,
					UserID: 999,
					ReviewData: models.BookReviewDataList{
						Items: reviewData,
					},
				}

				if err := models.CreateReview(db, newReview); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "保存到数据库失败"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"summary": jsonData})
				return
			}
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取有效的书籍摘要"})
}

func extractJSON(summary string) (map[string]interface{}, error) {
	// First try to find JSON within code fences
	re := regexp.MustCompile("(?s)```(?:json)?\\s*({[\\s\\S]*?})\\s*```")
	matches := re.FindStringSubmatch(summary)

	var jsonStr string
	if len(matches) > 1 {
		// Found JSON within code fences
		jsonStr = matches[1]
	} else {
		// Try to find raw JSON without code fences
		re = regexp.MustCompile("(?s){[\\s\\S]*?}")
		matches = re.FindStringSubmatch(summary)
		if len(matches) > 0 {
			jsonStr = matches[0]
		} else {
			return nil, fmt.Errorf("no JSON found in the response")
		}
	}

	// Clean up the JSON string
	jsonStr = strings.TrimSpace(jsonStr)

	// Parse the JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return result, nil
}
