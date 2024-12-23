package utils

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
)

// Common errors
var (
	ErrNoCover = errors.New("no cover image found in EPUB")
)

// ExtractEpubCover 尝试多种方式提取封面
func ExtractEpubCover(epubPath, outputDir string) (string, error) {
	// 首先尝试使用 go-fitz
	coverPath, err := ExtractEpubCoverWithFitz(epubPath, outputDir)
	if err == nil {
		return coverPath, nil
	}

	// 如果 fitz 失败，回退到传统方法
	return extractEpubCoverTraditional(epubPath, outputDir)
}

// ExtractEpubCoverWithFitz 使用 go-fitz 提取封面
func ExtractEpubCoverWithFitz(epubPath string, outputDir string) (string, error) {
	doc, err := fitz.New(epubPath)
	if err != nil {
		return "", fmt.Errorf("无法打开EPUB文件: %w", err)
	}
	defer doc.Close()

	img, err := doc.Image(0)
	if err != nil {
		return "", fmt.Errorf("无法提取封面图片: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	epubName := filepath.Base(epubPath)
	epubNameWithoutExt := strings.TrimSuffix(epubName, filepath.Ext(epubName))
	outputPath := filepath.Join(outputDir, epubNameWithoutExt+"_cover.jpg")

	output, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer output.Close()

	if err := jpeg.Encode(output, img, &jpeg.Options{Quality: 95}); err != nil {
		return "", fmt.Errorf("保存封面图片失败: %w", err)
	}

	return outputPath, nil
}

// extractEpubCoverTraditional 使用传统方法提取封面
func extractEpubCoverTraditional(epubPath string, outputDir string) (string, error) {
	zipReader, err := zip.OpenReader(epubPath)
	if err != nil {
		return "", fmt.Errorf("failed to open epub: %w", err)
	}
	defer zipReader.Close()

	containerFile, err := findFileInZip(&zipReader.Reader, "META-INF/container.xml")
	if err != nil {
		return "", fmt.Errorf("failed to find container.xml: %w", err)
	}

	container, err := readContainer(containerFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse container.xml: %w", err)
	}

	if len(container.RootFiles) == 0 {
		return "", errors.New("no root file found in container.xml")
	}

	opfPath := container.RootFiles[0].FullPath
	opfFile, err := findFileInZip(&zipReader.Reader, opfPath)
	if err != nil {
		return "", fmt.Errorf("failed to find OPF file: %w", err)
	}

	var pkg Package
	if err := xml.NewDecoder(opfFile).Decode(&pkg); err != nil {
		return "", fmt.Errorf("failed to parse OPF file: %w", err)
	}

	coverPath := findCoverPath(&pkg)
	if coverPath == "" {
		return "", ErrNoCover
	}

	opfDir := filepath.Dir(opfPath)
	fullCoverPath := filepath.Join(opfDir, coverPath)
	fullCoverPath = filepath.ToSlash(fullCoverPath)

	coverFile, err := findFileInZip(&zipReader.Reader, fullCoverPath)
	if err != nil {
		return "", fmt.Errorf("failed to find cover file: %w", err)
	}
	defer coverFile.Close()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, filepath.Base(coverPath))
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, coverFile); err != nil {
		return "", fmt.Errorf("failed to save cover image: %w", err)
	}

	return outputPath, nil
}

// findCoverPath 在包中查找封面路径
func findCoverPath(pkg *Package) string {
	// Method 1: Look for cover in metadata
	coverID := ""
	for _, meta := range pkg.Metadata.Meta {
		if meta.Name == "cover" {
			coverID = meta.Content
			break
		}
	}

	if coverID != "" {
		for _, item := range pkg.Manifest.Items {
			if item.ID == coverID {
				return item.Href
			}
		}
	}

	// Method 2: Look for cover-image property
	for _, item := range pkg.Manifest.Items {
		if strings.Contains(item.Properties, "cover-image") {
			return item.Href
		}
	}

	// Method 3: Look for image with "cover" in name
	for _, item := range pkg.Manifest.Items {
		if strings.HasPrefix(item.MediaType, "image/") &&
			(strings.Contains(strings.ToLower(item.Href), "cover") ||
				strings.Contains(strings.ToLower(item.ID), "cover")) {
			return item.Href
		}
	}

	return ""
}

// ExtractEpubContent 提取 EPUB 内容
// ExtractEpubContent 按章节提取 EPUB 内容
func ExtractEpubContent(epubPath string) ([]ChapterContent, error) {
	doc, err := fitz.New(epubPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开EPUB文件: %w", err)
	}
	defer doc.Close()

	// 首先获取章节结构
	toc, err := doc.ToC()
	if err != nil {
		return nil, fmt.Errorf("提取目录失败: %w", err)
	}

	var chapters []ChapterContent
	processor := NewContentProcessor()

	// 按章节提取内容
	for i, item := range toc {
		// 确定章节的页面范围
		startPage := item.Page - 1
		endPage := doc.NumPage()
		if i < len(toc)-1 {
			endPage = toc[i+1].Page - 1
		}

		// 提取该章节的所有页面内容
		var chapterText strings.Builder
		for page := startPage; page < endPage; page++ {
			text, err := doc.Text(page)
			if err != nil {
				continue
			}
			chapterText.WriteString(text)
		}

		// 使用内容处理器处理文本
		structuredContent := processor.ProcessContent(chapterText.String())

		chapter := ChapterContent{
			Title:      item.Title,
			Content:    parseContent(chapterText.String()),
			Structured: structuredContent,
			Level:      (item.Level + 1),
		}

		chapters = append(chapters, chapter)
	}

	return chapters, nil
}

// ExtractChapterListWithFitz 提取章节列表
func ExtractChapterListWithFitz(epubPath string) ([]ChapterStructure, error) {
	doc, err := fitz.New(epubPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开EPUB文件: %w", err)
	}
	defer doc.Close()

	toc, err := doc.ToC()
	if err != nil {
		return nil, fmt.Errorf("提取目录失败: %w", err)
	}

	var chapters []ChapterStructure
	for _, item := range toc {
		chapter := ChapterStructure{
			Title: item.Title,
			Level: item.Level,
			Href:  item.URI,
		}
		chapters = append(chapters, chapter)
	}

	return chapters, nil
}

// GetEpubMetadata 获取 EPUB 元数据
func GetEpubMetadata(epubPath string) (map[string]string, error) {
	doc, err := fitz.New(epubPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开EPUB文件: %w", err)
	}
	defer doc.Close()

	metadata := make(map[string]string)
	// 从文档获取元数据
	meta := doc.Metadata()

	// 将元数据复制到结果map中，并清理空字节
	for k, v := range meta {
		metadata[k] = strings.TrimRight(v, "\x00")
	}

	return metadata, nil
}

// Helper functions
func findFileInZip(reader *zip.Reader, path string) (io.ReadCloser, error) {
	for _, f := range reader.File {
		if f.Name == path {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func readContainer(r io.Reader) (*Container, error) {
	var container Container
	if err := xml.NewDecoder(r).Decode(&container); err != nil {
		return nil, err
	}
	return &container, nil
}

func parseContent(text string) []ContentNode {
	var nodes []ContentNode

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		node := ContentNode{
			Type: "paragraph",
			Text: line,
		}

		nodes = append(nodes, node)
	}

	return nodes
}
