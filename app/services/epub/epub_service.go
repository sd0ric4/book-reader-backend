package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/sd0ric4/book-reader-backend/app/models"
)

// ExtractAndSaveChapters 从PDF文件提取章节内容并保存
func ExtractAndSaveChapters(pdfPath string, bookID uint) ([]models.BookChapter, error) {
	// 打开PDF文件
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("打开PDF文件失败: %v", err)
	}
	defer doc.Close()

	// 获取总页数
	pageCount := doc.NumPage()

	// 用于存储所有章节
	var chapters []models.BookChapter

	// 当前章节内容
	var currentChapter strings.Builder
	var currentChapterName string
	var chapterStartPage int

	// 遍历所有页面
	for pageNum := 0; pageNum < pageCount; pageNum++ {
		// 提取当前页面的文本
		text, err := doc.Text(pageNum)
		if err != nil {
			return nil, fmt.Errorf("提取页面 %d 文本失败: %v", pageNum, err)
		}

		// 检测是否是新章节的开始
		if isChapterStart(text) {
			// 如果已经有当前章节，保存它
			if currentChapterName != "" {
				chapter := createChapter(
					bookID,
					currentChapterName,
					currentChapter.String(),
					chapterStartPage,
					pageNum-1,
				)
				chapters = append(chapters, chapter)
			}

			// 开始新章节
			currentChapter.Reset()
			currentChapterName = extractChapterName(text)
			chapterStartPage = pageNum
		}

		// 将当前页面内容添加到章节
		currentChapter.WriteString(text)
		currentChapter.WriteString("\n")
	}

	// 保存最后一个章节
	if currentChapterName != "" {
		chapter := createChapter(
			bookID,
			currentChapterName,
			currentChapter.String(),
			chapterStartPage,
			pageCount-1,
		)
		chapters = append(chapters, chapter)
	}

	// 调用 models 包函数保存所有章节
	return chapters, nil
}

// isChapterStart 判断文本是否是章节开始
func isChapterStart(text string) bool {
	// 这里需要根据实际PDF的格式来调整判断逻辑
	// 例如，可以检查是否包含"第X章"或者其他章节标记
	return strings.Contains(text, "第") && strings.Contains(text, "章")
}

// extractChapterName 从文本中提取章节名称
func extractChapterName(text string) string {
	// 这里需要根据实际PDF的格式来调整提取逻辑
	// 示例实现，需要根据实际情况修改
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.Contains(line, "章") {
			return strings.TrimSpace(line)
		}
	}
	return "未命名章节"
}

// createChapter 创建章节对象
func createChapter(bookID uint, name, content string, startPage, endPage int) models.BookChapter {
	// 创建章节结构信息
	structure := map[string]interface{}{
		"startPage": startPage,
		"endPage":   endPage,
	}
	structureJSON, _ := json.Marshal(structure)

	return models.BookChapter{
		BookID:           bookID,
		ChapterName:      name,
		ChapterContent:   content,
		ChapterStructure: string(structureJSON),
		ContentPath:      "", // 如果需要，可以设置内容路径
	}
}

// ExtractSpecificPages 提取指定页面范围的内容
func ExtractSpecificPages(pdfPath string, startPage, endPage int) (string, error) {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("打开PDF文件失败: %v", err)
	}
	defer doc.Close()

	if startPage < 0 || endPage >= doc.NumPage() || startPage > endPage {
		return "", fmt.Errorf("页面范围无效")
	}

	var content strings.Builder
	for pageNum := startPage; pageNum <= endPage; pageNum++ {
		text, err := doc.Text(pageNum)
		if err != nil {
			return "", fmt.Errorf("提取页面 %d 文本失败: %v", pageNum, err)
		}
		content.WriteString(text)
		content.WriteString("\n")
	}

	return content.String(), nil
}

// UpdateChapterFromPDF 更新指定章节的内容
func UpdateChapterFromPDF(chapterID uint, pdfPath string, startPage, endPage int) error {
	content, err := ExtractSpecificPages(pdfPath, startPage, endPage)
	if err != nil {
		return err
	}

	return models.UpdateChapterContent(nil, chapterID, content)
}

// DownloadPDF 从 URL 下载 PDF 文件并返回临时文件路径
func DownloadPDF(url string) (string, error) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "download-*.pdf")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer tmpFile.Close()

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("下载文件失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 检查 Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && contentType != "application/pdf" {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("不是有效的 PDF 文件: %s", contentType)
	}

	// 将响应内容写入临时文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("保存文件失败: %v", err)
	}

	return tmpFile.Name(), nil
}

// ExtractPDFFromURL 从 URL 下载 PDF 并提取内容
func ExtractPDFFromURL(url string, bookID uint) ([]models.BookChapter, error) {
	// 下载 PDF 文件
	pdfPath, err := DownloadPDF(url)
	if err != nil {
		return nil, fmt.Errorf("下载 PDF 失败: %v", err)
	}
	defer os.Remove(pdfPath) // 使用完后删除临时文件

	// 提取内容
	chapters, err := ExtractAndSaveChapters(pdfPath, bookID)
	if err != nil {
		return nil, fmt.Errorf("提取章节失败: %v", err)
	}

	return chapters, nil
}

// ExtractSpecificPagesFromURL 从 URL 下载 PDF 并提取指定页面
func ExtractSpecificPagesFromURL(url string, startPage, endPage int) (string, error) {
	// 下载 PDF 文件
	pdfPath, err := DownloadPDF(url)
	if err != nil {
		return "", fmt.Errorf("下载 PDF 失败: %v", err)
	}
	defer os.Remove(pdfPath) // 使用完后删除临时文件

	// 提取指定页面
	return ExtractSpecificPages(pdfPath, startPage, endPage)
}

// 在现有代码后添加新函数

// ExtractAllPages 提取PDF中的所有页面内容
func ExtractAllPages(pdfPath string) (string, error) {
	// 打开PDF文件
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("打开PDF文件失败: %v", err)
	}
	defer doc.Close()

	// 获取总页数
	pageCount := doc.NumPage()

	// 用于存储所有页面内容
	var content strings.Builder

	// 遍历所有页面
	for pageNum := 0; pageNum < pageCount; pageNum++ {
		// 提取当前页面的文本
		text, err := doc.Text(pageNum)
		if err != nil {
			return "", fmt.Errorf("提取页面 %d 文本失败: %v", pageNum, err)
		}
		content.WriteString(text)
		content.WriteString("\n")
	}

	return content.String(), nil
}

// ExtractAllPagesFromURL 从URL下载PDF并提取所有页面内容
func ExtractAllPagesFromURL(url string) (string, error) {
	// 下载 PDF 文件
	pdfPath, err := DownloadPDF(url)
	if err != nil {
		return "", fmt.Errorf("下载 PDF 失败: %v", err)
	}
	defer os.Remove(pdfPath) // 使用完后删除临时文件

	// 提取所有页面
	return ExtractAllPages(pdfPath)
}
