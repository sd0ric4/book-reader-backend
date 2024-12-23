package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

func cleanTitle(title string) string {
	title = strings.TrimSpace(title)
	if strings.HasPrefix(strings.ToLower(title), "chapter") {
		parts := strings.Fields(title)
		if len(parts) >= 2 {
			return fmt.Sprintf("第%s章", parts[1])
		}
	}
	return title
}

func generateCoverPath(bookPath, outputDir string) string {
	fileName := filepath.Base(bookPath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return filepath.Join(outputDir, fileNameWithoutExt+"_cover.jpg")
}
