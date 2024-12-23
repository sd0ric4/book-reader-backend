// types.go
package utils

import (
	"encoding/xml"
	"fmt"
)

// Common error type for ebook processing
type EBookError struct {
	Message string
	Err     error
}

func (e *EBookError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Common types for book content
type ChapterStructure struct {
	ID       string             `json:"id"`
	Title    string             `json:"title"`
	Level    int                `json:"level"`
	Href     string             `json:"href"`
	Children []ChapterStructure `json:"children,omitempty"`
}

type ContentNode struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	Level    int           `json:"level,omitempty"`
	Children []ContentNode `json:"children,omitempty"`
}

type ChapterContent struct {
	Title      string              `json:"title"`
	Content    []ContentNode       `json:"content"`    // 原始内容
	Structured []StructuredContent `json:"structured"` // 结构化内容
	Level      int                 `json:"level"`      // 章节层级
}

// ContentType 定义内容的类型
type ContentType string

const (
	TextBlock ContentType = "text"    // 普通文本块
	Heading   ContentType = "heading" // 标题
	Image     ContentType = "image"   // 图片
	Table     ContentType = "table"   // 表格
	List      ContentType = "list"    // 列表
	Quote     ContentType = "quote"   // 引用
	Code      ContentType = "code"    // 代码块
)

// StructuredContent 表示结构化的内容
type StructuredContent struct {
	Type     ContentType         `json:"type"`
	Level    int                 `json:"level,omitempty"`    // 用于标题层级或列表嵌套层级
	Content  string              `json:"content"`            // 主要内容
	Metadata map[string]string   `json:"metadata,omitempty"` // 额外的元数据，如图片URL、样式等
	Children []StructuredContent `json:"children,omitempty"` // 子内容，用于嵌套结构
}

// Package represents the OPF package document
type Package struct {
	XMLName  xml.Name `xml:"package"`
	Metadata Metadata `xml:"metadata"`
	Manifest Manifest `xml:"manifest"`
	Spine    Spine    `xml:"spine"`
}

// Container represents the META-INF/container.xml file
type Container struct {
	XMLName   xml.Name   `xml:"container"`
	RootFiles []RootFile `xml:"rootfiles>rootfile"`
}

type RootFile struct {
	FullPath  string `xml:"full-path,attr"`
	MediaType string `xml:"media-type,attr"`
}

type Metadata struct {
	Meta []Meta `xml:"meta"`
}

type Meta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type Manifest struct {
	Items []Item `xml:"item"`
}

type Item struct {
	ID         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr,omitempty"`
}

type Spine struct {
	Items []SpineItem `xml:"itemref"`
}

type SpineItem struct {
	IDRef string `xml:"idref,attr"`
}

// Chapter represents a chapter in the EPUB
type Chapter struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Href     string    `json:"href"`
	Order    int       `json:"order"`
	Level    int       `json:"level"`
	Children []Chapter `json:"children,omitempty"`
}
