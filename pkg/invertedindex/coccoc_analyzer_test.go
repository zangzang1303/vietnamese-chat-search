package invertedindex

import (
	"path/filepath"
	"testing"
)

func TestCoccocAnalyzer_SearchAccuracyComparison(t *testing.T) {
	dictPath, err := filepath.Abs("../../data/dicts/coccoc")
	if err != nil {
		t.Fatalf("Lỗi đường dẫn: %v", err)
	}

	// 1. Khởi tạo 2 bộ Index: một dùng Standard, một dùng Cốc Cốc
	stdAnalyzer := &StandardAnalyzer{}
	stdIndex := NewInvertedIndex()

	coccocAnalyzer, err := NewCoccocAnalyzer(dictPath, false)
	if err != nil {
		t.Fatalf("Không thể khởi tạo CoccocAnalyzer: %v", err)
	}
	coccocIndex := NewInvertedIndex()

	// 2. Dữ liệu thử nghiệm (các tin nhắn chat chứa từ "sinh")
	docs := []struct {
		id      int
		message string
	}{
		{id: 1, message: "Chào các bạn học sinh mới vào trường"},
		{id: 2, message: "Sinh viên năm nhất đi làm thêm quán cà phê"},
		{id: 3, message: "Người lính đã hy sinh anh dũng vì tổ quốc"},
	}

	for _, d := range docs {
		stdIndex.AddDocument(Document{ID: d.id, Content: d.message}, stdAnalyzer)
		coccocIndex.AddDocument(Document{ID: d.id, Content: d.message}, coccocAnalyzer)
	}

	// 3. Tìm kiếm với truy vấn "học sinh"
	query := "học sinh"

	// Kết quả với Standard Analyzer
	stdResults := stdIndex.Search(query, stdAnalyzer)
	t.Logf("--- Kết quả tìm '%s' với Standard Analyzer ---", query)
	for _, r := range stdResults {
		t.Logf("DocID: %d | Score: %.4f | Content: %s", r.Document.ID, r.Score, r.Document.Content)
	}

	// Kết quả với Cốc Cốc Analyzer
	coccocResults := coccocIndex.Search(query, coccocAnalyzer)
	t.Logf("--- Kết quả tìm '%s' với Cốc Cốc Analyzer ---", query)
	for _, r := range coccocResults {
		t.Logf("DocID: %d | Score: %.4f | Content: %s", r.Document.ID, r.Score, r.Document.Content)
	}

	// Kiểm tra độ chính xác:
	// Cốc Cốc PHẢI CHỈ trả về đúng Doc 1 (chứa từ ghép "học_sinh")
	// Doc 2 ("sinh_viên") và Doc 3 ("hy_sinh") KHÔNG được xuất hiện!
	if len(coccocResults) != 1 {
		t.Fatalf("Kỳ vọng Cốc Cốc trả về đúng 1 kết quả, nhưng nhận được %d kết quả", len(coccocResults))
	}

	if coccocResults[0].Document.ID != 1 {
		t.Errorf("Kỳ vọng DocID=1, nhận được DocID=%d", coccocResults[0].Document.ID)
	}

	// Trong khi đó, Standard Analyzer sẽ match cả 3 documents (do đều chứa từ đơn "sinh")
	if len(stdResults) < 2 {
		t.Errorf("Standard Analyzer kỳ vọng trả về nhiều false positives (ít nhất 2 docs), nhưng chỉ có %d", len(stdResults))
	}
}
