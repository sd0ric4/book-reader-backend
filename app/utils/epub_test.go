package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sd0ric4/book-reader-backend/app/config"
	"github.com/sd0ric4/book-reader-backend/app/services"
	"github.com/stretchr/testify/require"
)

func TestEpubUtils(t *testing.T) {
	// 1. 初始化配置和服务
	config.LoadConfig("../../config/config.yaml")

	client, err := services.NewMinioClient(config.Config.S3)
	if err != nil {
		t.Fatalf("初始化 Minio 客户端失败: %s", err)
	}

	s3Service := services.NewS3Service(client, config.Config.S3.BucketName)

	// 测试用例列表
	tests := []struct {
		name         string
		fileName     string
		wantErr      bool
		expectedMeta map[string]string
	}{
		{
			name:     "傲慢与偏见",
			fileName: "Pride and Prejudice (Austen Jane) (Z-Library).epub",
			wantErr:  false,
			expectedMeta: map[string]string{
				"title":  "Pride and Prejudice",
				"author": "Jane Austen",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 2. 获取文件URL
			url, err := s3Service.GetFileURL(tt.fileName)
			if err != nil {
				t.Fatalf("获取文件URL失败: %s", err)
			}
			t.Logf("文件URL: %s", url)

			// 3. 创建临时目录
			tempDir := t.TempDir()
			epubPath := filepath.Join(tempDir, tt.fileName)
			outputDir := filepath.Join(tempDir, "covers")

			// 4. 下载EPUB文件
			err = downloadFile(url, epubPath)
			require.NoError(t, err, "下载EPUB文件失败")

			// 5. 测试封面提取 (go-fitz方式)
			t.Run("Test_Fitz_Cover", func(t *testing.T) {
				coverPath, err := ExtractEpubCoverWithFitz(epubPath, outputDir)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.NotEmpty(t, coverPath)
				validateImage(t, coverPath)
			})

			// 6. 测试传统方式提取封面
			t.Run("Test_Traditional_Cover", func(t *testing.T) {
				coverPath, err := ExtractEpubCover(epubPath, outputDir)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.NotEmpty(t, coverPath)
				validateImage(t, coverPath)
			})

			// 7. 测试内容提取
			t.Run("Test_Content", func(t *testing.T) {
				chapters, err := ExtractEpubContent(epubPath)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				if len(chapters) > 0 {
					for _, chapter := range chapters {
						require.NotEmpty(t, chapter.Title, "章节标题不应为空")
						if len(chapter.Content) > 0 {
							t.Logf("章节: %s, 内容项数: %d", chapter.Title, len(chapter.Content))
							for i, content := range chapter.Content {
								require.NotEmpty(t, content.Text,
									"章节 %s 的第 %d 个内容项不应为空", chapter.Title, i)
							}
						} else {
							t.Logf("警告: 章节 %s 没有内容", chapter.Title)
						}
					}
				} else {
					t.Log("警告: 未提取到任何章节内容")
				}
			})
			t.Run("Test_Content_With_processor", func(t *testing.T) {
				// 提取内容
				chapters, err := ExtractEpubContent(epubPath)
				if err != nil {
					log.Fatal(err)
				}

				// 处理结构化内容
				for _, chapter := range chapters {
					// 访问结构化内容
					for _, content := range chapter.Structured {
						switch content.Type {
						case Heading:
							fmt.Printf("发现标题: %s (级别 %d)\n", content.Content, content.Level)
						case List:
							fmt.Printf("列表项: %s (缩进级别 %d)\n", content.Content, content.Level)
						case TextBlock:
							fmt.Printf("正文: %s\n", content.Content)
						}
					}
				}
			})
			// 8. 测试章节列表提取
			t.Run("Test_ChapterList", func(t *testing.T) {
				chapters, err := ExtractChapterListWithFitz(epubPath)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				if len(chapters) > 0 {
					for _, chapter := range chapters {
						require.NotEmpty(t, chapter.Title, "章节标题不应为空")
						require.NotEmpty(t, chapter.Href, "章节链接不应为空")
						t.Logf("发现章节: %s", chapter.Title)
					}
				} else {
					t.Log("警告: 未发现章节列表")
				}
			})

			// 9. 测试元数据提取
			t.Run("Test_Metadata", func(t *testing.T) {
				metadata, err := GetEpubMetadata(epubPath)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.NotEmpty(t, metadata, "元数据不应为空")

				for key, expectedValue := range tt.expectedMeta {
					actualValue, exists := metadata[key]
					if !exists {
						t.Logf("警告: 未找到期望的元数据键: %s", key)
						continue
					}

					// 清理和比较值
					expectedValue = strings.TrimSpace(expectedValue)
					actualValue = strings.TrimSpace(actualValue)
					require.Equal(t, expectedValue, actualValue,
						"元数据值不匹配 key: %s, 期望: %s, 实际: %s",
						key, expectedValue, actualValue)
				}

				// 输出所有找到的元数据
				t.Log("找到的所有元数据:")
				for key, value := range metadata {
					t.Logf("%s: %s", key, value)
				}
			})
		})
	}
}

// 辅助函数：下载文件
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// 辅助函数：验证图片
func validateImage(t *testing.T, imagePath string) {
	file, err := os.Open(imagePath)
	require.NoError(t, err, "应能打开图片文件")
	defer file.Close()

	// 读取文件头
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	require.NoError(t, err, "应能读取图片文件头")

	// 验证文件类型
	contentType := http.DetectContentType(buffer)
	require.True(t, strings.HasPrefix(contentType, "image/"),
		"文件应为图片格式，实际类型为: %s", contentType)

	// 验证文件大小
	fileInfo, err := file.Stat()
	require.NoError(t, err, "应能获取文件信息")
	require.Greater(t, fileInfo.Size(), int64(0), "图片文件不应为空")

	t.Logf("成功验证图片: %s, 类型: %s, 大小: %d bytes",
		imagePath, contentType, fileInfo.Size())
}
