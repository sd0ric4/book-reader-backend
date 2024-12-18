package services

import (
	"math"
	"strings"

	"gonum.org/v1/gonum/mat"
)

// ComputeTFIDF 计算TF-IDF矩阵
func ComputeTFIDF(documents []string) *mat.Dense {
	// 词频计算
	wordFreq := make(map[string][]int)
	totalDocs := len(documents)

	// 统计单词在文档中出现的频率
	for docIndex, doc := range documents {
		words := strings.Fields(doc)
		for _, word := range words {
			if wordFreq[word] == nil {
				wordFreq[word] = make([]int, totalDocs)
			}
			wordFreq[word][docIndex]++
		}
	}

	// 创建TF-IDF矩阵
	rows := len(documents)
	cols := len(wordFreq)
	tfidfMatrix := mat.NewDense(rows, cols, nil)

	wordList := make([]string, 0, len(wordFreq))
	for word := range wordFreq {
		wordList = append(wordList, word)
	}

	for j, word := range wordList {
		// 计算逆文档频率
		docsWithWord := 0
		for _, freq := range wordFreq[word] {
			if freq > 0 {
				docsWithWord++
			}
		}
		idf := math.Log(float64(totalDocs) / (1 + float64(docsWithWord)))

		for i := 0; i < rows; i++ {
			// 计算词频
			tf := float64(wordFreq[word][i]) / float64(len(strings.Fields(documents[i])))
			tfidfMatrix.Set(i, j, tf*idf)
		}
	}

	return tfidfMatrix
}

// CosineSimilarity 计算余弦相似度
func CosineSimilarity(a, b *mat.VecDense) float64 {
	dotProduct := mat.Dot(a, b)
	normA := mat.Norm(a, 2)
	normB := mat.Norm(b, 2)

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}
