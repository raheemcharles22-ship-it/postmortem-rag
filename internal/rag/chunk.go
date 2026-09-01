// internal/rag/chunk.go
package rag

import "errors"

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
	var myChunks []Chunk
	index := 0
	start := 0
	stride := size - overlap
	runes := []rune(text)
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
