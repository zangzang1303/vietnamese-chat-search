//go:build !cgo

package coccoc

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

// Tokenizer đóng vai trò Pure Go Tokenizer khi môi trường không bật CGO (như Windows mặc định)
// Sử dụng thuật toán Maximum Matching (Longest Match Forward) trên từ điển từ ghép tiếng Việt Cốc Cốc
type Tokenizer struct {
	mu          sync.RWMutex
	dictPath    string
	compounds   map[string]bool
	unaccentMap map[string]string
}

var (
	globalInstance *Tokenizer
	globalOnce     sync.Once
)

// New khởi tạo Pure Go Tokenizer với từ điển từ ghép tiếng Việt chuẩn
func New(dictPath string, loadNontone bool) (*Tokenizer, error) {
	globalOnce.Do(func() {
		tok := &Tokenizer{
			dictPath:    dictPath,
			compounds:   make(map[string]bool, 50000),
			unaccentMap: make(map[string]string, 50000),
		}

		// Nạp từ điển các từ ghép tiếng Việt
		tok.loadDictionary(dictPath)
		globalInstance = tok
	})

	return globalInstance, nil
}

func (t *Tokenizer) loadDictionary(dictPath string) {
	candidates := []string{
		filepath.Join(dictPath, "vietnamese_compound_words.txt"),
		filepath.Join(filepath.Dir(dictPath), "vietnamese_compound_words.txt"),
		"data/dicts/vietnamese_compound_words.txt",
		"../../data/dicts/vietnamese_compound_words.txt",
		"../../../data/dicts/vietnamese_compound_words.txt",
	}

	for _, cand := range candidates {
		if file, err := os.Open(cand); err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				w := strings.ToLower(strings.TrimSpace(scanner.Text()))
				if w != "" && strings.Contains(w, " ") {
					t.compounds[w] = true
					unacc := removeDiacritics(w)
					t.unaccentMap[unacc] = w
				}
			}
			file.Close()
			if len(t.compounds) > 0 {
				break
			}
		}
	}

	// Bổ sung các từ ghép cốt lõi đảm bảo luôn có
	coreCompounds := []string{
		"học sinh", "sinh viên", "cà phê", "phê bình", "phê duyệt", "phê chuẩn",
		"bàn ghế", "bàn bạc", "bàn luận", "bàn thảo", "bàn tán",
		"trà sữa", "trà đá", "sữa chua", "sửa chữa", "sửa sang",
		"trường học", "y sinh", "hy sinh", "sinh nhật", "khuyến học",
		"học bổng", "học phí", "đi làm", "nhập học", "thời sự", "kinh tế",
	}
	for _, w := range coreCompounds {
		t.compounds[w] = true
		t.unaccentMap[removeDiacritics(w)] = w
	}
}

// SegmentOriginal phân đoạn văn bản và ghép các từ ghép bằng dấu "_"
// Ví dụ: "học sinh trường học y sinh" -> "học_sinh trường_học y_sinh"
func (t *Tokenizer) SegmentOriginal(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	words := strings.Fields(text)
	var result []string
	n := len(words)
	i := 0

	for i < n {
		matched := false
		// Thuật toán Maximum Matching: Ưu tiên ghép từ dài nhất (từ 4 từ xuống 2 từ)
		for win := 4; win >= 2; win-- {
			if i+win <= n {
				// Kiểm tra: Không được ghép nếu các từ ở giữa chứa dấu câu ở đuôi (ví dụ "bạn," không ghép với "tôi")
				hasPunctuation := false
				for k := 0; k < win-1; k++ {
					w := words[i+k]
					runes := []rune(w)
					if len(runes) > 0 {
						lastChar := runes[len(runes)-1]
						if !unicode.IsLetter(lastChar) && !unicode.IsNumber(lastChar) {
							hasPunctuation = true
							break
						}
					}
				}
				if hasPunctuation {
					continue
				}

				candidateWords := make([]string, win)
				for k := 0; k < win; k++ {
					candidateWords[k] = cleanWord(words[i+k])
				}
				candidate := strings.ToLower(strings.Join(candidateWords, " "))
				candidateUnacc := removeDiacritics(candidate)

				isUnaccentedInput := (candidate == candidateUnacc)
				isMatched := false
				if t.compounds[candidate] {
					isMatched = true
				} else if isUnaccentedInput && t.unaccentMap[candidate] != "" {
					isMatched = true
				}

				if isMatched {
					// Lưu lại dấu câu ở từ cuối cùng nếu có
					lastOriginal := words[i+win-1]
					cleanedLast := cleanWord(lastOriginal)
					punctSuffix := ""
					if len(lastOriginal) > len(cleanedLast) {
						punctSuffix = lastOriginal[len(cleanedLast):]
					}

					joinedWords := make([]string, win)
					for k := 0; k < win; k++ {
						joinedWords[k] = cleanWord(words[i+k])
					}
					joined := strings.Join(joinedWords, "_") + punctSuffix

					result = append(result, joined)
					i += win
					matched = true
					break
				}
			}
		}
		if !matched {
			result = append(result, words[i])
			i++
		}
	}

	return strings.Join(result, " "), nil
}

// Tokenize nhận vào một câu văn bản và trả về danh sách các tokens đã được ghép từ
func (t *Tokenizer) Tokenize(text string) ([]string, error) {
	segmented, err := t.SegmentOriginal(text)
	if err != nil {
		return nil, err
	}

	lower := strings.ToLower(segmented)
	tokens := strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})
	return tokens, nil
}

func cleanWord(w string) string {
	return strings.TrimFunc(w, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func removeDiacritics(s string) string {
	var b strings.Builder
	accents := map[rune]rune{
		'à': 'a', 'á': 'a', 'ả': 'a', 'ã': 'a', 'ạ': 'a',
		'ă': 'a', 'ằ': 'a', 'ắ': 'a', 'ẳ': 'a', 'ẵ': 'a', 'ặ': 'a',
		'â': 'a', 'ầ': 'a', 'ấ': 'a', 'ẩ': 'a', 'ẫ': 'a', 'ậ': 'a',
		'đ': 'd',
		'è': 'e', 'é': 'e', 'ẻ': 'e', 'ẽ': 'e', 'ẹ': 'e',
		'ê': 'e', 'ề': 'e', 'ế': 'e', 'ể': 'e', 'ễ': 'e', 'ệ': 'e',
		'ì': 'i', 'í': 'i', 'ỉ': 'i', 'ĩ': 'i', 'ị': 'i',
		'ò': 'o', 'ó': 'o', 'ỏ': 'o', 'õ': 'o', 'ọ': 'o',
		'ô': 'o', 'ồ': 'o', 'ố': 'o', 'ổ': 'o', 'ỗ': 'o', 'ộ': 'o',
		'ơ': 'o', 'ờ': 'o', 'ớ': 'o', 'ở': 'o', 'ỡ': 'o', 'ợ': 'o',
		'ù': 'u', 'ú': 'u', 'ủ': 'u', 'ũ': 'u', 'ụ': 'u',
		'ư': 'u', 'ừ': 'u', 'ứ': 'u', 'ử': 'u', 'ữ': 'u', 'ự': 'u',
		'ỳ': 'y', 'ý': 'y', 'ỷ': 'y', 'ỹ': 'y', 'ỵ': 'y',
	}
	for _, r := range strings.ToLower(s) {
		if mapped, ok := accents[r]; ok {
			b.WriteRune(mapped)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
