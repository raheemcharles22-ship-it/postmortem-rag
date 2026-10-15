// internal/rag/chunk.go
package rag

import (
	"errors"
	"strings"
)

type Chunk struct {
	Text   string
	Index  int    // position in source doc
	Source string // doc id, carried through for citations
}

func ChunkText(text, source string, size, overlap int) ([]Chunk, error) {
	if size <= 0 {
		return nil, errors.New("size must be positive")
	}
	if overlap < 0 || overlap >= size {
		return nil, errors.New("overlap must be non-negative and less than siz")
	}
	clean_text := StripFrontmatter(text)

	var myChunks []Chunk
	index := 0
	start := 0
	stride := size - overlap
	runes := []rune(clean_text)
	for start < len(runes) {
		end := min(start+size, len(runes))
		chunk := Chunk{Text: string(runes[start:end]), Index: index, Source: source}
		start += stride
		index++
		myChunks = append(myChunks, chunk)
		if end == len(runes) {
			break
		}
	}

	return myChunks, nil
}

func StripFrontmatter(text string) string {
	if !strings.HasPrefix(text, "---") {
		return text
	}
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 {
		return text
	}
	return strings.TrimLeft(parts[2], "\n")
}
