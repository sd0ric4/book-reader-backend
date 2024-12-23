// processor.go
package utils

import (
	"strings"
)

type ContentProcessor struct {
	currentLevel int
	buffer       []StructuredContent
}

// NewContentProcessor 创建新的内容处理器
func NewContentProcessor() *ContentProcessor {
	return &ContentProcessor{
		currentLevel: 0,
		buffer:       make([]StructuredContent, 0),
	}
}

// ProcessContent 将原始内容转换为结构化内容
func (p *ContentProcessor) ProcessContent(text string) []StructuredContent {
	// 重置 buffer，确保每次处理都是全新的开始
	p.buffer = make([]StructuredContent, 0)

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理不同类型的内容
		node := p.classifyAndProcessLine(line)
		if node != nil {
			p.buffer = append(p.buffer, *node)
		}
	}

	// 合并并返回处理后的内容
	result := p.mergeContent()

	// 处理完成后清空 buffer，防止影响下次处理
	p.buffer = make([]StructuredContent, 0)

	return result
}
