//go:build !cgo

package coccoc

import (
	"strings"
	"unicode"
)

// Tokenizer đóng vai trò fallback khi biên dịch trên môi trường không bật CGO (như Windows mặc định)
type Tokenizer struct {
	dictPath string
}

// New khởi tạo stub Tokenizer an toàn
func New(dictPath string, loadNontone bool) (*Tokenizer, error) {
	return &Tokenizer{dictPath: dictPath}, nil
}

// SegmentOriginal trả về văn bản gốc nếu không có CGO
func (t *Tokenizer) SegmentOriginal(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}
	return text, nil
}

// Tokenize phân tách từ vựng cơ bản dự phòng
func (t *Tokenizer) Tokenize(text string) ([]string, error) {
	lower := strings.ToLower(text)
	tokens := strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})
	return tokens, nil
}
