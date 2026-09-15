package invertedindex

import "sort"

// ============================================================================
// KIẾN THỨC GO CƠ BẢN:
// 1. Pattern tạo "Set" trong Go:
//    - Go không có kiểu dữ liệu `Set` như Python hay Java.
//    - Để lưu các phần tử duy nhất (không trùng lặp), ta dùng `map[int]bool`.
//    - Khi gán `map[id] = true`, nếu id đã tồn tại thì nó chỉ ghi đè lên chứ không bị trùng.
// 2. Sắp xếp mảng bằng "sort.Slice":
//    - Hàm `sort.Slice(slice, func(i, j int) bool)` nhận vào một hàm ẩn danh (anonymous function).
//    - Nếu muốn sắp xếp GIẢM DẦN (điểm cao đứng trước): viết `results[i].Score > results[j].Score`.
//    - Nếu muốn sắp xếp TĂNG DẦN: viết `results[i].Score < results[j].Score`.
// ============================================================================

// SearchResult đại diện cho một kết quả tìm kiếm trả về cho người dùng
type SearchResult struct {
	Document Document // Thông tin tin nhắn gốc
	Score    float64  // Điểm liên quan tính bằng BM25 (càng cao càng khớp)
}

// Search thực hiện luồng tìm kiếm và xếp hạng tài liệu
// Tham số:
// - queryText: Câu người dùng nhập vào ô tìm kiếm (ví dụ: "học sinh")
// - analyzer: Bộ phân tích dùng để bóc tách câu query (phải tương thích với dữ liệu lúc index)
func (idx *InvertedIndex) Search(queryText string, analyzer Analyzer) []SearchResult {
	// BƯỚC 1: Phân tích câu query thành danh sách các token
	// Ví dụ: query "học sinh" qua VietnameseAnalyzer -> ["học_sinh"]
	tokens := analyzer.Analyze(queryText)

	// Nếu câu query rỗng hoặc toàn ký tự đặc biệt bị cắt hết -> trả về nil (rỗng)
	if len(tokens) == 0 {
		return nil
	}

	// BƯỚC 2: Thu thập các tài liệu ứng viên (Candidate Retrieval)
	// Dùng kỹ thuật Boolean OR: tài liệu nào chứa ÍT NHẤT MỘT từ trong câu query đều được đưa vào danh sách ứng viên.
	// Sử dụng map[int]bool như một Set để loại bỏ các DocID bị trùng lặp.
	candidateDocs := make(map[int]bool)
	for _, t := range tokens {
		// Tra từ điển lấy Posting List của từ t
		postings := idx.Dictionary[t]
		for _, posting := range postings {
			candidateDocs[posting.DocID] = true // Đánh dấu tài liệu này là ứng viên
		}
	}

	// BƯỚC 3: Chấm điểm BM25 cho từng tài liệu ứng viên
	var results []SearchResult
	for docID := range candidateDocs {
		score := idx.CalculateBM25Score(docID, tokens)

		// Chỉ giữ lại những tài liệu có điểm số > 0
		if score > 0 {
			results = append(results, SearchResult{
				Document: idx.Documents[docID], // Lấy nội dung gốc từ map Documents
				Score:    score,
			})
		}
	}

	// BƯỚC 4: Xếp hạng kết quả (Ranking)
	// Sắp xếp danh sách kết quả giảm dần theo điểm BM25 (tin nhắn phù hợp nhất lên đầu)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
