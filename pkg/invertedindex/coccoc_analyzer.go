package invertedindex

import (
	"vietnamese-chat-search/pkg/tokenizer/coccoc"
)

// ==============================================================================
// CoccocAnalyzer: Bộ phân tích văn bản sử dụng Cốc Cốc Tokenizer (CGO)
// Thỏa mãn interface `Analyzer`:
//     type Analyzer interface {
//         Analyze(text string) []string
//     }
// ==============================================================================

// CoccocAnalyzer bao bọc coccoc.Tokenizer để tích hợp trực tiếp vào InvertedIndex
type CoccocAnalyzer struct {
	tokenizer *coccoc.Tokenizer
	unaccent  bool // Có chuẩn hóa bỏ dấu tiếng Việt hay không
}

// NewCoccocAnalyzer khởi tạo CoccocAnalyzer với đường dẫn từ điển
// unaccent: nếu true, các tokens sau khi tách từ ghép sẽ được bỏ dấu (ví dụ: "cà_phê" -> "ca_phe")
func NewCoccocAnalyzer(dictPath string, unaccent bool) (*CoccocAnalyzer, error) {
	tok, err := coccoc.New(dictPath, false)
	if err != nil {
		return nil, err
	}

	return &CoccocAnalyzer{
		tokenizer: tok,
		unaccent:  unaccent,
	}, nil
}

// Analyze thực hiện:
// 1. Phân tách từ ghép và dấu câu bằng Cốc Cốc Tokenizer
// 2. (Tùy chọn) Bỏ dấu tiếng Việt nếu unaccent = true
func (a *CoccocAnalyzer) Analyze(text string) []string {
	tokens, err := a.tokenizer.Tokenize(text)
	if err != nil {
		// Trong trường hợp lỗi không mong muốn, fallback về StandardAnalyzer
		std := &StandardAnalyzer{}
		return std.Analyze(text)
	}

	if a.unaccent {
		for i, tok := range tokens {
			tokens[i] = RemoveDiacritics(tok)
		}
	}

	return tokens
}
