// internal/rag/chunk_test.go
package rag

import (
	"strings"
	"testing"
)

// Counts assume stride = size - overlap and keep-the-tail.
// If you pick different semantics, change `want` and leave a comment saying why.
func TestChunkText_Counts(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		size    int
		overlap int
		want    int
	}{
		{"empty", "", 100, 20, 0},
		{"shorter than size", "hello world", 100, 20, 1},
		{"exact size", strings.Repeat("a", 100), 100, 0, 1},
		{"two clean, no overlap", strings.Repeat("a", 200), 100, 0, 2},
		{"overlap stride 50", strings.Repeat("a", 200), 100, 50, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ChunkText(tt.text, "src", tt.size, tt.overlap)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("got %d chunks, want %d", len(got), tt.want)
			}
		})
	}
}

func TestChunkText_Invariants(t *testing.T) {
	chunks, err := ChunkText(strings.Repeat("a", 250), "postmortem-1", 100, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}
	for i, c := range chunks {
		if c.Source != "postmortem-1" {
			t.Errorf("chunk %d lost source: %q", i, c.Source)
		}
		if c.Index != i {
			t.Errorf("chunk %d has Index %d, want %d", i, c.Index, i)
		}
		if c.Text == "" {
			t.Errorf("chunk %d is empty", i)
		}
	}
}

// Chunking must never split a multi-byte rune.
func TestChunkText_Unicode(t *testing.T) {
	text := strings.Repeat("é", 150) // 2 bytes per rune
	chunks, err := ChunkText(text, "src", 100, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, c := range chunks {
		if strings.Contains(c.Text, "\uFFFD") {
			t.Errorf("chunk %d split a rune", i)
		}
	}
}

// Optional, stronger. Enable once your stride semantics are locked.
// De-overlapped concatenation must equal the original: no characters dropped.
func TestChunkText_NoDataLoss(t *testing.T) {
	text := "the retry storm was caused by an unbounded exponential backoff loop"
	overlap := 5
	chunks, err := ChunkText(text, "src", 20, overlap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var b strings.Builder
	for i, c := range chunks {
		if i == 0 {
			b.WriteString(c.Text)
			continue
		}
		r := []rune(c.Text)
		if overlap < len(r) {
			b.WriteString(string(r[overlap:])) // assumes uniform overlap
		}
	}
	if b.String() != text {
		t.Errorf("data lost:\n got:  %q\n want: %q", b.String(), text)
	}
}
func TestChunkText_UnicodeOddCut(t *testing.T) {
	text := strings.Repeat("é", 100) // 2 bytes each
	chunks, err := ChunkText(text, "src", 33, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, c := range chunks {
		if strings.Contains(c.Text, "\uFFFD") {
			t.Errorf("chunk %d split a rune", i)
		}
	}
}

func TestChunkText_RejectsBadInput(t *testing.T) {
	cases := []struct {
		name    string
		size    int
		overlap int
	}{
		{"overlap equals size", 100, 100},
		{"overlap exceeds size", 100, 150},
		{"negative overlap", 100, -1},
		{"zero size", 0, 0},
		{"negative size", -5, 2},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ChunkText("hello world", "src", tt.size, tt.overlap)
			if err == nil {
				t.Fatalf("size=%d overlap=%d: want error, got nil", tt.size, tt.overlap)
			}
		})
	}
}
