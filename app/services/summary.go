package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
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
请总结以下书籍的主要人物和主要内容概述，按照固定格式回答：

书名：《%s》
作者：%s
主要人物：
- 人物1：描述其在书中的角色或作用。
- 人物2：描述其在书中的角色或作用。
（根据书籍内容添加更多主要人物。）

主要内容概述：
请简要概述书籍的主要故事情节、主题和核心思想，控制在150字以内。

示例：
书名：《红楼梦》
主要人物：
- 贾宝玉：贾府的公子，具有叛逆性格，是书中主要叙事线的中心。
- 林黛玉：一位聪明且感性的女性，与贾宝玉有深厚感情。
- 薛宝钗：性格温和，与贾宝玉的关系构成三角恋。

主要内容概述：
《红楼梦》以贾宝玉、林黛玉的爱情悲剧为线索，以贾、史、王、薛四大家族的兴衰为背景，着重叙述了贾家荣、宁两府逐渐衰败的过程。广泛地反映了当时的社会现象和各种矛盾，揭露了封建官僚地主家庭的荒淫腐败、虚伪欺诈及其各种罪恶活动，歌颂了贾宝玉、林黛玉的封建叛逆精神，描绘了一些纯洁少女的悲惨遭遇和反抗性格，对封建末期社会进行了剖析和批判，揭示了封建社会必然灭亡的历史趋势。

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

func GetBookSummary(c *gin.Context) {
	API_KEY := aiconfig.APIKey
	BASE_URL := aiconfig.BaseURL
	MODEL_NAME := aiconfig.ModelName
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 准备提示词
	prompt := fmt.Sprintf(
		PROMPT_TEMPLATE,
		requestData.BookTitle,
		requestData.Author,
		requestData.Author,
		requestData.BookTitle,
	)

	// 创建 OpenAI 客户端，允许自定义 BaseURL
	client := NewOpenAIClient(
		API_KEY,
		WithBaseURL(BASE_URL),
	)

	// 准备 ChatCompletion 请求
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: MODEL_NAME,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   500, // 根据需要调整
			Temperature: 0.7, // 控制随机性
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("API 调用失败: %v", err),
		})
		return
	}

	// 检查响应
	if len(resp.Choices) > 0 {
		summary := resp.Choices[0].Message.Content
		c.JSON(http.StatusOK, gin.H{"summary": summary})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "未能获取书籍摘要",
		})
	}
}
