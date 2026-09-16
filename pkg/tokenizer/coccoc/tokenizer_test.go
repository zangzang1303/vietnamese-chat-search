package coccoc

import (
	"path/filepath"
	"reflect"
	"testing"
)

func getDictPath(t *testing.T) string {
	// Đường dẫn tương đối từ pkg/tokenizer/coccoc đến data/dicts/coccoc
	absPath, err := filepath.Abs("../../../data/dicts/coccoc")
	if err != nil {
		t.Fatalf("Không thể lấy đường dẫn tuyệt đối của thư mục từ điển: %v", err)
	}
	return absPath
}

func TestCoccocTokenizer_SegmentOriginal(t *testing.T) {
	dictPath := getDictPath(t)

	tokenizer, err := New(dictPath, false)
	if err != nil {
		t.Fatalf("Khởi tạo Cốc Cốc Tokenizer thất bại: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Từ ghép học sinh và cà phê",
			input:    "Học sinh đi học hôm nay uống cà phê",
			expected: "Học_sinh đi học hôm_nay uống cà_phê",
		},
		{
			name:     "Có dấu câu phẩy và chấm than",
			input:    "Chào bạn, tôi là học sinh!",
			expected: "Chào bạn, tôi là học_sinh!",
		},
		{
			name:     "Chuỗi rỗng",
			input:    "",
			expected: "",
		},
		{
			name:     "Chuỗi toàn khoảng trắng",
			input:    "   ",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tokenizer.SegmentOriginal(tc.input)
			if err != nil {
				t.Fatalf("SegmentOriginal trả về lỗi: %v", err)
			}
			if got != tc.expected {
				t.Errorf("SegmentOriginal(%q) = %q; kỳ vọng %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestCoccocTokenizer_Tokenize(t *testing.T) {
	dictPath := getDictPath(t)

	tokenizer, err := New(dictPath, false)
	if err != nil {
		t.Fatalf("Khởi tạo Cốc Cốc Tokenizer thất bại: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Tách từ ghép chuẩn",
			input:    "Học sinh đi học hôm nay uống cà phê",
			expected: []string{"học_sinh", "đi", "học", "hôm_nay", "uống", "cà_phê"},
		},
		{
			name:     "Xử lý dấu câu và chữ hoa",
			input:    "Chào bạn, tôi là học sinh!",
			expected: []string{"chào", "bạn", "tôi", "là", "học_sinh"},
		},
		{
			name:     "Phân biệt từ đơn và từ ghép",
			input:    "Học sinh và sinh viên khác nhau",
			expected: []string{"học_sinh", "và", "sinh_viên", "khác", "nhau"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tokenizer.Tokenize(tc.input)
			if err != nil {
				t.Fatalf("Tokenize trả về lỗi: %v", err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("Tokenize(%q) = %v; kỳ vọng %v", tc.input, got, tc.expected)
			}
		})
	}
}

func BenchmarkCoccocTokenizer(b *testing.B) {
	absPath, err := filepath.Abs("../../../data/dicts/coccoc")
	if err != nil {
		b.Fatalf("Lỗi đường dẫn: %v", err)
	}

	tokenizer, err := New(absPath, false)
	if err != nil {
		b.Fatalf("Lỗi khởi tạo: %v", err)
	}

	sampleText := "Chào bạn, hôm nay tôi muốn hẹn bạn đi uống cà phê bàn về thời khóa biểu của học sinh."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tokenizer.Tokenize(sampleText)
	}
}
