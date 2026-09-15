package invertedindex

import (
	"strings"
	"unicode"
)

// ============================================================================
// KIẾN THỨC GO CƠ BẢN:
// 1. "interface" trong Go:
//    - Go không có từ khóa "implements". Nếu một struct có các hàm với tên và kiểu
//      khớp với interface thì nó TỰ ĐỘNG được coi là thỏa mãn interface đó (Duck typing).
// 2. "Receiver" trong Go:
//    - Cú pháp `func (a *StandardAnalyzer) Analyze(...)`: phần `(a *StandardAnalyzer)`
//      được gọi là "method receiver", tương đương với con trỏ `this` trong Java/C++
//      hoặc `self` trong Python.
// 3. Kiểu "rune" trong Go:
//    - Một string trong Go là một chuỗi các byte (mã hóa UTF-8).
//    - Ký tự tiếng Việt có dấu (như 'à', 'ê', 'đ') chiếm từ 2 đến 3 bytes.
//    - "rune" (thực chất là số nguyên int32) đại diện cho 1 ký tự Unicode trọn vẹn.
//    - Khi thao tác cắt chuỗi tiếng Việt, LUÔN LUÔN phải chuyển sang `[]rune`
//      để tránh bị cắt đứt nửa chừng một byte gây lỗi font (ký tự hình thoi )!
// ============================================================================

// Analyzer là giao diện (interface) chung cho mọi bộ phân tích văn bản.
// Bất kỳ struct nào có hàm Analyze(text string) []string đều được coi là một Analyzer.
type Analyzer interface {
	Analyze(text string) []string
}

// ----------------------------------------------------------------------------
// 1. StandardAnalyzer: Mô phỏng bộ phân tích mặc định của Elasticsearch
// ----------------------------------------------------------------------------

// StandardAnalyzer là một struct rỗng, đóng vai trò như một class tiện ích.
type StandardAnalyzer struct{}

// Analyze thực hiện:
// 1. Chuyển toàn bộ chữ thành chữ thường (lowercase)
// 2. Cắt các từ theo khoảng trắng và dấu câu (chỉ giữ lại chữ cái và chữ số)
func (a *StandardAnalyzer) Analyze(text string) []string {
	// strings.ToLower là hàm có sẵn trong thư viện "strings" của Go
	text = strings.ToLower(text)

	// strings.FieldsFunc nhận vào chuỗi và một hàm kiểm tra từng ký tự (rune).
	// Nếu hàm trả về true, ký tự đó sẽ được dùng làm ranh giới để cắt từ.
	words := strings.FieldsFunc(text, func(r rune) bool {
		// Dùng package "unicode": nếu KHÔNG phải chữ cái và KHÔNG phải chữ số thì cắt bỏ
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	return words
}

// ----------------------------------------------------------------------------
// 2. VietnameseAnalyzer: Phân tích từ vựng có ghép từ tiếng Việt
// ----------------------------------------------------------------------------

// VietnameseAnalyzer chứa danh sách các từ ghép mẫu để nhận diện
type VietnameseAnalyzer struct {
	compoundWords []string // Slice chứa các từ ghép tiếng Việt (từ điển nhỏ)
}

// NewVietnameseAnalyzer là "Constructor pattern" phổ biến trong Go
// (Go không có constructor tự động, ta quy ước viết hàm New... để khởi tạo struct)
func NewVietnameseAnalyzer() *VietnameseAnalyzer {
	return &VietnameseAnalyzer{
		compoundWords: []string{
			"học sinh", "sinh viên", "cà phê", "phê bình",
			"bàn ghế", "bàn bạc", "thời khóa biểu", "đi làm", "nhập học",
		},
	}
}

// Analyze nhận diện các cụm từ ghép và nối chúng lại bằng dấu "_"
// Ví dụ: "em là học sinh" -> ["em", "là", "học_sinh"]
func (a *VietnameseAnalyzer) Analyze(text string) []string {
	lower := strings.ToLower(text)

	// Duyệt qua danh sách từ ghép trong từ điển
	// Vòng lặp "for _, cw := range ...":
	// - "_" là blank identifier (bỏ qua biến chỉ số index vì không dùng tới)
	// - "cw" nhận giá trị từng phần tử trong slice
	for _, cw := range a.compoundWords {
		// Nối từ ghép bằng gạch dưới, ví dụ: "học sinh" -> "học_sinh"
		joined := strings.ReplaceAll(cw, " ", "_")
		// Thay thế trong toàn bộ câu
		lower = strings.ReplaceAll(lower, cw, joined)
	}

	// Cắt từ: giữ lại chữ cái, chữ số VÀ dấu gạch dưới "_" (để không bị tách từ ghép ra lại)
	tokens := strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})

	return tokens
}

// ----------------------------------------------------------------------------
// 3. UnaccentFilter: Loại bỏ dấu tiếng Việt (chuẩn hóa về chữ không dấu)
// ----------------------------------------------------------------------------

// RemoveDiacritics chuyển các ký tự có dấu về không dấu (ví dụ: "cà_phê" -> "ca_phe")
func RemoveDiacritics(s string) string {
	// strings.Builder là công cụ tối ưu bộ nhớ trong Go để ghép chuỗi (thay vì dùng s += "a" chậm)
	var b strings.Builder

	// Bảng tra cứu ánh xạ ký tự có dấu -> ký tự không dấu (map giữa các rune)
	accents := map[rune]rune{
		'à': 'a', 'á': 'a', 'ả': 'a', 'ã': 'a', 'ạ': 'a',
		'ă': 'a', 'ằ': 'a', 'ắ': 'a', 'ẳ': 'a', 'ẵ': 'a', 'ặ': 'a',
		'â': 'a', 'ầ': 'a', 'ấ': 'a', 'ẩ': 'a', 'ẫ': 'a', 'ậ': 'a',
		'è': 'e', 'é': 'e', 'ẻ': 'e', 'ẽ': 'e', 'ẹ': 'e',
		'ê': 'e', 'ề': 'e', 'ế': 'e', 'ể': 'e', 'ễ': 'e', 'ệ': 'e',
		'ì': 'i', 'í': 'i', 'ỉ': 'i', 'ĩ': 'i', 'ị': 'i',
		'ò': 'o', 'ó': 'o', 'ỏ': 'o', 'õ': 'o', 'ọ': 'o',
		'ô': 'o', 'ồ': 'o', 'ố': 'o', 'ổ': 'o', 'ỗ': 'o', 'ộ': 'o',
		'ơ': 'o', 'ờ': 'o', 'ớ': 'o', 'ở': 'o', 'ỡ': 'o', 'ợ': 'o',
		'ù': 'u', 'ú': 'u', 'ủ': 'u', 'ũ': 'u', 'ụ': 'u',
		'ư': 'u', 'ừ': 'u', 'ứ': 'u', 'ử': 'u', 'ữ': 'u', 'ự': 'u',
		'ỳ': 'y', 'ý': 'y', 'ỷ': 'y', 'ỹ': 'y', 'ỵ': 'y',
		'đ': 'd',
	}

	// Duyệt qua từng ký tự (rune) trong chuỗi đầu vào
	for _, r := range s {
		// "found" là biến boolean cho biết ký tự r có nằm trong map accents hay không
		if replacement, found := accents[unicode.ToLower(r)]; found {
			b.WriteRune(replacement) // Thay thế bằng ký tự không dấu
		} else {
			b.WriteRune(r) // Giữ nguyên ký tự gốc (như a, b, c, 1, 2, _)
		}
	}

	return b.String()
}

// ----------------------------------------------------------------------------
// 4. EdgeNgramFilter: Sinh tiền tố phục vụ tìm kiếm dở từ (Partial Matching)
// ----------------------------------------------------------------------------

// GenerateEdgeNgrams sinh các chuỗi tiền tố từ độ dài minGram đến maxGram
// Ví dụ: token = "cà_phê", minGram = 2, maxGram = 5
// Kết quả trả về: ["cà", "cà_", "cà_p", "cà_ph"]
func GenerateEdgeNgrams(token string, minGram, maxGram int) []string {
	// Ép kiểu string sang []rune để cắt theo từng chữ cái tiếng Việt an toàn!
	runes := []rune(token)
	length := len(runes)

	// Khai báo một slice chuỗi rỗng
	var ngrams []string

	// Vòng lặp for kinh điển trong Go
	for i := minGram; i <= maxGram && i <= length; i++ {
		// Cắt mảng con (slice operator): runes[:i] lấy từ vị trí 0 đến trước i
		// Sau đó ép kiểu ngược lại từ []rune sang string
		ngrams = append(ngrams, string(runes[:i]))
	}

	return ngrams
}
