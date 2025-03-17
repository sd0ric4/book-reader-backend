package services

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sd0ric4/book-reader-backend/app/config"
	"github.com/sd0ric4/book-reader-backend/app/database"
)

const testURL = "http://192.168.0.118:5244/d/library/%E6%81%B6%E6%84%8F%E4%BB%A3%E7%A0%81.pdf?sign=ouDTiH2AaZBAdeneIwz6Ob-IQay-IopATnoEA6XzXnU=:0" // 替换为实际的测试 URL
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

func TestExtractPDFFromURL(t *testing.T) {
	// 下载 PDF 文件
	pdfPath, err := DownloadPDF(testURL)
	if err != nil {
		t.Fatalf("下载 PDF 文件失败: %v", err)
	}
	defer os.Remove(pdfPath) // 测试完成后删除临时文件

	// 测试提取内容
	t.Run("测试章节提取", func(t *testing.T) {
		chapters, err := ExtractAndSaveChapters(pdfPath, 1)
		if err != nil {
			t.Errorf("提取章节失败: %v", err)
			return
		}

		// 可以继续对 chapters 进行断言测试
		if len(chapters) == 0 {
			t.Log("未提取到章节")
		}
	})

	t.Run("测试特定页面提取", func(t *testing.T) {
		content, err := ExtractSpecificPages(pdfPath, 0, 1)
		t.Log(content)
		if err != nil {
			t.Errorf("提取特定页面失败: %v", err)
		}
		if content == "" {
			t.Error("提取的内容为空")
		}
	})
}

func TestDownloadPDF(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "有效的 PDF URL",
			url:     testURL,
			wantErr: false,
		},
		{
			name:    "无效的 URL",
			url:     "https://invalid-url.com/not-exists.pdf",
			wantErr: true,
		},
		{
			name:    "非 PDF URL",
			url:     "https://example.com/not-pdf.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdfPath, err := DownloadPDF(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Error("期望得到错误，但没有错误")
				}
				return
			}

			if err != nil {
				t.Errorf("下载失败: %v", err)
				return
			}

			// 检查文件是否存在且大小不为零
			info, err := os.Stat(pdfPath)
			if err != nil {
				t.Errorf("无法获取文件信息: %v", err)
				return
			}

			if info.Size() == 0 {
				t.Error("下载的文件大小为零")
			}

			// 清理临时文件
			os.Remove(pdfPath)
		})
	}
}
