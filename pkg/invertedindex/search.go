package invertedindex

import (
	"sort"
	"strings"
)

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

// Search thực hiện luồng tìm kiếm và xếp hạng tài liệu hỗ trợ từ ghép, gõ dở (Edge N-grams) và không dấu
// Tham số:
// - queryText: Câu người dùng nhập vào ô tìm kiếm (ví dụ: "học sinh", "cà p", "ca phe")
// - analyzer: Bộ phân tích dùng để bóc tách câu query (phải tương thích với dữ liệu lúc index)
func (idx *InvertedIndex) Search(queryText string, analyzer Analyzer) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	rawLower := strings.ToLower(strings.TrimSpace(queryText))
	if rawLower == "" {
		return nil
	}

	// BƯỚC 1: Phân tích câu query thành danh sách các token
	tokens := analyzer.Analyze(rawLower)

	// Xây dựng tập tra cứu và trọng số đa tầng (Boosting)
	// Exact Token (5.0) > Phrase Space (4.5) > Edge N-gram (3.5) > Unaccented (3.0) > Sub-token (1.5)
	weightedTerms := make(map[string]float64)

	// 1. Tokenized match
	for _, t := range tokens {
		weightedTerms[t] = 5.0
		ut := RemoveDiacritics(t)
		if ut != t {
			if _, exists := weightedTerms[ut]; !exists {
				weightedTerms[ut] = 3.0
			}
		}
	}

	// 2. Chuỗi query thô (hỗ trợ trường hợp người dùng gõ cụm từ có dấu cách hoặc gõ dở "cà p", "cà phe")
	if _, exists := weightedTerms[rawLower]; !exists {
		weightedTerms[rawLower] = 4.5
	}
	underscoreRaw := strings.ReplaceAll(rawLower, " ", "_")
	if _, exists := weightedTerms[underscoreRaw]; !exists {
		weightedTerms[underscoreRaw] = 4.5
	}

	// 3. Biến thể không dấu của câu query thô ("ca p", "ca phe")
	unaccentedRaw := RemoveDiacritics(rawLower)
	if _, exists := weightedTerms[unaccentedRaw]; !exists {
		weightedTerms[unaccentedRaw] = 3.5
	}
	unaccentedUnderscore := strings.ReplaceAll(unaccentedRaw, " ", "_")
	if _, exists := weightedTerms[unaccentedUnderscore]; !exists {
		weightedTerms[unaccentedUnderscore] = 3.5
	}

	// BƯỚC 2: Thu thập các tài liệu ứng viên (Candidate Retrieval)
	candidateDocs := make(map[int]bool)
	for term := range weightedTerms {
		postings := idx.Dictionary[term]
		for _, posting := range postings {
			candidateDocs[posting.DocID] = true
		}
	}

	// BƯỚC 3: Chấm điểm BM25 đa tầng cho từng tài liệu ứng viên
	var results []SearchResult
	for docID := range candidateDocs {
		score := idx.CalculateBM25WeightedScore(docID, weightedTerms)

		// Chỉ giữ lại những tài liệu có điểm số > 0
		if score > 0 {
			results = append(results, SearchResult{
				Document: idx.Documents[docID],
				Score:    score,
			})
		}
	}

	// BƯỚC 4: Xếp hạng kết quả (Ranking)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

