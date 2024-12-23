// metadata.go
package utils

import (
	"strings"

	"github.com/gen2brain/go-fitz"
)

func GetBookMetadata(bookPath string) (map[string]string, error) {
	doc, err := fitz.New(bookPath)
	if err != nil {
		return nil, &EBookError{
			Message: "Unable to open book file",
			Err:     err,
		}
	}
	defer doc.Close()

	metadata := make(map[string]string)
	meta := doc.Metadata()

	for k, v := range meta {
		cleanValue := strings.TrimRight(v, "\x00")
		if cleanValue != "" {
			metadata[k] = cleanValue
		}
	}

	return metadata, nil
}
