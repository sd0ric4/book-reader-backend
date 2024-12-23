package utils

import (
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
)

// MobiError represents custom error types for MOBI processing
type MobiError struct {
	Message string
	Err     error
}

func (e *MobiError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// MobiChapter represents a chapter in the MOBI file
type MobiChapter struct {
	Title      string              `json:"title"`
	Content    string              `json:"content"`
	Level      int                 `json:"level"`
	Structured []StructuredContent `json:"structured"`
}

// ExtractMobiContent extracts content from MOBI file
func ExtractMobiContent(mobiPath string) ([]MobiChapter, error) {
	doc, err := fitz.New(mobiPath)
	if err != nil {
		return nil, &MobiError{
			Message: "无法打开MOBI文件",
			Err:     err,
		}
	}
	defer doc.Close()

	// 获取目录结构
	toc, err := doc.ToC()
	if err != nil {
		return nil, &MobiError{
			Message: "无法提取目录",
			Err:     err,
		}
	}

	var chapters []MobiChapter
	processor := NewContentProcessor()

	// 处理每个章节
	for i, item := range toc {
		startPage := item.Page - 1
		endPage := doc.NumPage()
		if i < len(toc)-1 {
			endPage = toc[i+1].Page - 1
		}

		// 提取章节内容
		var content strings.Builder
		for page := startPage; page < endPage; page++ {
			text, err := doc.Text(page)
			if err != nil {
				continue
			}
			content.WriteString(text)
		}

		// 处理内容
		contentText := content.String()
		structuredContent := processor.ProcessContent(contentText)

		chapter := MobiChapter{
			Title:      cleanMobiTitle(item.Title),
			Content:    contentText,
			Level:      item.Level + 1,
			Structured: structuredContent,
		}

		chapters = append(chapters, chapter)
	}

	return chapters, nil
}

// ExtractMobiChapterList extracts the chapter list from MOBI file
func ExtractMobiChapterList(mobiPath string) ([]ChapterStructure, error) {
	doc, err := fitz.New(mobiPath)
	if err != nil {
		return nil, &MobiError{
			Message: "无法打开MOBI文件",
			Err:     err,
		}
	}
	defer doc.Close()

	toc, err := doc.ToC()
	if err != nil {
		return nil, &MobiError{
			Message: "无法提取目录",
			Err:     err,
		}
	}

	var chapters []ChapterStructure
	for _, item := range toc {
		chapter := ChapterStructure{
			Title: cleanMobiTitle(item.Title),
			Level: item.Level,
			Href:  item.URI,
		}
		chapters = append(chapters, chapter)
	}

	return chapters, nil
}

// ExtractMobiCover extracts the cover image from MOBI file
func ExtractMobiCover(mobiPath, outputDir string) (string, error) {
	doc, err := fitz.New(mobiPath)
	if err != nil {
		return "", &MobiError{
			Message: "无法打开MOBI文件",
			Err:     err,
		}
	}
	defer doc.Close()

	// 获取第一页作为封面
	img, err := doc.Image(0)
	if err != nil {
		return "", &MobiError{
			Message: "无法提取封面图片",
			Err:     err,
		}
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", &MobiError{
			Message: "无法创建输出目录",
			Err:     err,
		}
	}

	// 生成输出文件路径
	fileName := filepath.Base(mobiPath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	outputPath := filepath.Join(outputDir, fileNameWithoutExt+"_cover.jpg")

	// 创建输出文件
	output, err := os.Create(outputPath)
	if err != nil {
		return "", &MobiError{
			Message: "无法创建封面文件",
			Err:     err,
		}
	}
	defer output.Close()

	// 保存为JPEG
	if err := jpeg.Encode(output, img, &jpeg.Options{Quality: 95}); err != nil {
		return "", &MobiError{
			Message: "无法保存封面图片",
			Err:     err,
		}
	}

	return outputPath, nil
}

// GetMobiMetadata extracts metadata from MOBI file
func GetMobiMetadata(mobiPath string) (map[string]string, error) {
	doc, err := fitz.New(mobiPath)
	if err != nil {
		return nil, &MobiError{
			Message: "无法打开MOBI文件",
			Err:     err,
		}
	}
	defer doc.Close()

	metadata := make(map[string]string)
	meta := doc.Metadata()

	// 清理并保存元数据
	for k, v := range meta {
		cleanValue := strings.TrimRight(v, "\x00")
		if cleanValue != "" {
			metadata[k] = cleanValue
		}
	}

	return metadata, nil
}

// Helper functions
func cleanMobiTitle(title string) string {
	title = strings.TrimSpace(title)
	// 处理常见的章节标题格式
	if strings.HasPrefix(strings.ToLower(title), "chapter") {
		parts := strings.Fields(title)
		if len(parts) >= 2 {
			return fmt.Sprintf("第%s章", parts[1])
		}
	}
	return title
}
