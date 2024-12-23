package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// mergeContent 合并相邻的相同类型内容
func (p *ContentProcessor) mergeContent() []StructuredContent {
	if len(p.buffer) == 0 {
		return nil
	}

	result := make([]StructuredContent, 0)
	current := p.buffer[0]

	for i := 1; i < len(p.buffer); i++ {
		next := p.buffer[i]

		// 如果当前和下一个是相同类型的文本块，合并它们
		if current.Type == TextBlock && next.Type == TextBlock {
			current.Content += "\n" + next.Content
		} else {
			result = append(result, current)
			current = next
		}
	}

	result = append(result, current)
	return result
}

// Reset 重置处理器状态
func (p *ContentProcessor) Reset() {
	p.currentLevel = 0
	p.buffer = make([]StructuredContent, 0)
}

// classifyAndProcessLine 识别并处理单行内容
func (p *ContentProcessor) classifyAndProcessLine(line string) *StructuredContent {
	// 识别代码块
	if isCode := p.detectCodeBlock(line); isCode {
		return &StructuredContent{
			Type:    Code,
			Content: line,
			Metadata: map[string]string{
				"language": p.detectCodeLanguage(line),
			},
		}
	}

	// 识别图片
	if isImage := p.detectImage(line); isImage {
		metadata := p.extractImageMetadata(line)
		return &StructuredContent{
			Type:     Image,
			Content:  metadata["alt"],
			Metadata: metadata,
		}
	}

	// 识别标题
	if isHeading := p.detectHeading(line); isHeading {
		level := p.detectHeadingLevel(line)
		return &StructuredContent{
			Type:    Heading,
			Level:   level,
			Content: strings.TrimSpace(removeHeadingMarkers(line)),
		}
	}

	// 识别列表项
	if isList := p.detectListItem(line); isList {
		level := p.detectListLevel(line)
		return &StructuredContent{
			Type:    List,
			Level:   level,
			Content: strings.TrimSpace(removeListMarkers(line)),
		}
	}

	// 识别引用
	if isQuote := strings.HasPrefix(line, ">"); isQuote {
		return &StructuredContent{
			Type:    Quote,
			Content: strings.TrimSpace(strings.TrimPrefix(line, ">")),
		}
	}

	// 识别表格
	if isTable := p.detectTable(line); isTable {
		return &StructuredContent{
			Type:    Table,
			Content: line,
		}
	}

	// 默认作为文本块处理
	return &StructuredContent{
		Type:    TextBlock,
		Content: line,
	}
}

// 内容类型检测函数
func (p *ContentProcessor) detectCodeBlock(line string) bool {
	return strings.HasPrefix(line, "```")
}

func (p *ContentProcessor) detectCodeLanguage(line string) string {
	if !strings.HasPrefix(line, "```") {
		return ""
	}
	line = strings.TrimPrefix(line, "```")
	if idx := strings.Index(line, "\n"); idx != -1 {
		return line[:idx]
	}
	return line
}

func (p *ContentProcessor) detectImage(line string) bool {
	return regexp.MustCompile(`!\[.*?\]\(.*?\)`).MatchString(line)
}

func (p *ContentProcessor) extractImageMetadata(line string) map[string]string {
	altRegex := regexp.MustCompile(`!\[(.*?)\]`)
	urlRegex := regexp.MustCompile(`\((.*?)\)`)

	metadata := make(map[string]string)

	if alt := altRegex.FindStringSubmatch(line); len(alt) > 1 {
		metadata["alt"] = alt[1]
	}

	if url := urlRegex.FindStringSubmatch(line); len(url) > 1 {
		metadata["url"] = url[1]
	}

	return metadata
}

func (p *ContentProcessor) detectTable(line string) bool {
	return strings.Contains(line, "|") &&
		(strings.HasPrefix(line, "|") || strings.HasSuffix(line, "|"))
}

func (p *ContentProcessor) detectHeading(line string) bool {
	return strings.HasPrefix(line, "#") ||
		(len(line) > 0 && (strings.HasSuffix(line, "===") || strings.HasSuffix(line, "---")))
}

func (p *ContentProcessor) detectHeadingLevel(line string) int {
	if strings.HasPrefix(line, "#") {
		return strings.IndexFunc(line, func(r rune) bool {
			return r != '#'
		})
	}
	if strings.HasSuffix(line, "===") {
		return 1
	}
	if strings.HasSuffix(line, "---") {
		return 2
	}
	return 1
}

func (p *ContentProcessor) detectListItem(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "- ") ||
		strings.HasPrefix(trimmed, "* ") ||
		matchesNumberedList(trimmed)
}

func (p *ContentProcessor) detectListLevel(line string) int {
	return (strings.IndexFunc(line, func(r rune) bool {
		return !unicode.IsSpace(r)
	}) / 2) + 1
}

// 辅助函数
func matchesNumberedList(line string) bool {
	return regexp.MustCompile(`^\d+\.\s`).MatchString(line)
}

func removeHeadingMarkers(line string) string {
	if strings.HasPrefix(line, "#") {
		return strings.TrimLeft(line, "# ")
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "==="), "---")
}

func removeListMarkers(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return strings.TrimSpace(line[2:])
	}
	if matchesNumberedList(line) {
		parts := strings.SplitN(line, ". ", 2)
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	return line
}

// ConvertToStructuredContent 将原有的 ContentNode 转换为新的结构化格式
func ConvertToStructuredContent(nodes []ContentNode) []StructuredContent {
	processor := NewContentProcessor()
	var result []StructuredContent

	for _, node := range nodes {
		content := processor.ProcessContent(node.Text)
		result = append(result, content...)
	}

	return result
}

// BuildChapterContent 构建章节内容
func BuildChapterContent(title string, content []ContentNode) ChapterContent {
	structuredContent := ConvertToStructuredContent(content)

	return ChapterContent{
		Title:      title,
		Content:    content,
		Structured: structuredContent,
	}
}

// Marshal 将结构化内容转换为 JSON
func (c *ChapterContent) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

// String 返回章节内容的字符串表示
func (c *ChapterContent) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("== %s ==\n", c.Title))

	for _, content := range c.Structured {
		switch content.Type {
		case Heading:
			builder.WriteString(strings.Repeat("#", content.Level))
			builder.WriteString(" ")
			builder.WriteString(content.Content)
		case TextBlock:
			builder.WriteString(content.Content)
		case List:
			builder.WriteString(strings.Repeat("  ", content.Level-1))
			builder.WriteString("- ")
			builder.WriteString(content.Content)
		case Quote:
			builder.WriteString("> ")
			builder.WriteString(content.Content)
		case Code:
			builder.WriteString("```")
			if lang, ok := content.Metadata["language"]; ok {
				builder.WriteString(lang)
			}
			builder.WriteString("\n")
			builder.WriteString(content.Content)
			builder.WriteString("\n```")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
