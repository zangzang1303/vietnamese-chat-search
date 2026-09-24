package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

type LexiconTokenizer struct {
	compounds map[string]bool
	unaccentMap map[string]string // unaccent -> original or just unaccent set
}

func NewLexiconTokenizer(dictPath string) (*LexiconTokenizer, error) {
	tok := &LexiconTokenizer{
		compounds: make(map[string]bool, 50000),
		unaccentMap: make(map[string]string, 50000),
	}

	file, err := os.Open(dictPath)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			w := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if w != "" && strings.Contains(w, " ") {
				tok.compounds[w] = true
				unacc := removeDiacritics(w)
				tok.unaccentMap[unacc] = w
			}
		}
	}

	return tok, nil
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

func (tok *LexiconTokenizer) Segment(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}

	// Tách các từ và giữ dấu câu
	words := strings.Fields(text)
	var result []string
	n := len(words)
	i := 0

	for i < n {
		matched := false
		// Thử ghép tối đa 4 từ xuống 2 từ (Maximum Matching)
		for win := 4; win >= 2; win-- {
			if i+win <= n {
				candidateWords := make([]string, win)
				for k := 0; k < win; k++ {
					candidateWords[k] = cleanWord(words[i+k])
				}
				candidate := strings.ToLower(strings.Join(candidateWords, " "))
				candidateUnacc := removeDiacritics(candidate)

				if tok.compounds[candidate] || tok.unaccentMap[candidateUnacc] != "" {
					// Nối các từ lại bằng dấu '_'
					joined := strings.Join(words[i:i+win], "_")
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

	return strings.Join(result, " ")
}

func cleanWord(w string) string {
	return strings.TrimFunc(w, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func (tok *LexiconTokenizer) Tokenize(text string) []string {
	seg := tok.Segment(text)
	lower := strings.ToLower(seg)
	return strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})
}

func main() {
	tok, err := NewLexiconTokenizer("data/dicts/vietnamese_compound_words.txt")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded %d compound words.\n", len(tok.compounds))

	tests := []string{
		"học sinh",
		"hoc sinh",
		"trường học y sinh",
		"Chào các bạn học sinh mới vào trường",
		"uống cà phê",
		"sinh viên đi làm thêm quán cà phê",
	}

	for _, t := range tests {
		fmt.Printf("\nInput: %q\n", t)
		fmt.Printf("  Segment:  %q\n", tok.Segment(t))
		fmt.Printf("  Tokenize: %v\n", tok.Tokenize(t))
	}
}
