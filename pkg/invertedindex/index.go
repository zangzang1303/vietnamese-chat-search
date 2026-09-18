package invertedindex

import (
	"sort"
)

// ============================================================================
// KIẾN THỨC GO CƠ BẢN VỀ CẬP NHẬT CHỈ MỤC & ĐA LUỒNG:
// 1. "sync.RWMutex" (Read-Write Mutex):
//    - Lock(): Dành cho thao tác ghi/sửa (Add/Update/Delete). Chỉ duy nhất 1 goroutine
//      được ghi tại một thời điểm, các luồng khác phải xếp hàng đợi.
//    - RLock(): Dành cho thao tác đọc/tìm kiếm (Search/GetStats). Nhiều luồng có thể
//      đọc song song cùng lúc, giúp tối ưu hiệu năng tối đa.
// 2. Nguyên tắc "Update = Delete + Add" trong Search Engine:
//    - Trong Lucene/Elasticsearch, cập nhật tài liệu không phải là sửa đè tại chỗ.
//    - Hệ thống sẽ đánh dấu xóa (Tombstone) bản ghi cũ, sau đó nạp bản ghi mới.
// ============================================================================

// NewInvertedIndex là hàm khởi tạo để tạo ra một Inverted Index rỗng trong RAM
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		Dictionary:  make(map[string][]Posting),
		Documents:   make(map[int]Document),
		DocLengths:  make(map[int]int),
		TotalTokens: 0,
	}
}

// deleteDocumentInternal xóa một tài liệu khỏi chỉ mục (Hàm private, gọi khi đã có Lock)
func (idx *InvertedIndex) deleteDocumentInternal(docID int) bool {
	// Kiểm tra tài liệu có tồn tại trong hệ thống không
	_, exists := idx.Documents[docID]
	if !exists {
		return false
	}

	// 1. Giảm tổng số token toàn hệ thống và xóa thông tin độ dài
	oldLength := idx.DocLengths[docID]
	idx.TotalTokens -= oldLength
	delete(idx.DocLengths, docID)
	delete(idx.Documents, docID)

	// 2. Dọn dẹp Posting List của tất cả các Term
	// Duyệt qua Dictionary để tìm và gỡ bỏ Posting của docID này
	for term, postings := range idx.Dictionary {
		newPostings := make([]Posting, 0, len(postings))
		for _, p := range postings {
			if p.DocID != docID {
				newPostings = append(newPostings, p)
			}
		}

		if len(newPostings) == 0 {
			// Nếu từ này không còn xuất hiện ở bất kỳ tài liệu nào khác -> xóa hẳn key khỏi Dictionary
			delete(idx.Dictionary, term)
		} else {
			idx.Dictionary[term] = newPostings
		}
	}

	return true
}

// DeleteDocument xóa một tài liệu theo DocID (Thread-safe)
func (idx *InvertedIndex) DeleteDocument(docID int) bool {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return idx.deleteDocumentInternal(docID)
}

// AddDocument thực hiện đánh chỉ mục một tài liệu vào Inverted Index (Hỗ trợ Upsert an toàn)
func (idx *InvertedIndex) AddDocument(doc Document, analyzer Analyzer) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Nếu tài liệu đã từng tồn tại, ta dọn sạch phiên bản cũ trước để tránh trùng lặp Posting
	if _, exists := idx.Documents[doc.ID]; exists {
		idx.deleteDocumentInternal(doc.ID)
	}

	// BƯỚC 1: Bóc tách nội dung câu thành danh sách các từ vựng (tokens)
	tokens := analyzer.Analyze(doc.Content)

	// BƯỚC 2: Lưu lại tài liệu gốc và thống kê độ dài (|D|)
	idx.Documents[doc.ID] = doc
	idx.DocLengths[doc.ID] = len(tokens)
	idx.TotalTokens += len(tokens)

	// BƯỚC 3: Thu thập các vị trí (Positions) của từng token trong câu này
	termPositions := make(map[string][]int)
	for pos, token := range tokens {
		termPositions[token] = append(termPositions[token], pos)
	}

	// BƯỚC 4: Tạo Posting và chèn vào Posting List của từng Term
	for token, positions := range termPositions {
		posting := Posting{
			DocID:         doc.ID,
			TermFrequency: len(positions),
			Positions:     positions,
		}
		idx.Dictionary[token] = append(idx.Dictionary[token], posting)
	}
}

// UpdateDocument cập nhật nội dung tin nhắn và tự động re-index lại
func (idx *InvertedIndex) UpdateDocument(doc Document, analyzer Analyzer) {
	// AddDocument đã có sẵn logic tự xóa bản cũ nếu ID đã tồn tại
	idx.AddDocument(doc, analyzer)
}

// AvgDocLength tính toán độ dài trung bình của tất cả tài liệu trong index (avgdl)
func (idx *InvertedIndex) AvgDocLength() float64 {
	// Lưu ý: Hàm này thường được gọi trong ngữ cảnh đã có RLock
	if len(idx.Documents) == 0 {
		return 0
	}
	return float64(idx.TotalTokens) / float64(len(idx.Documents))
}

// GetStats trả về thống kê tổng quan của Inverted Index
func (idx *InvertedIndex) GetStats() IndexStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	avgdl := 0.0
	if len(idx.Documents) > 0 {
		avgdl = float64(idx.TotalTokens) / float64(len(idx.Documents))
	}

	return IndexStats{
		TotalDocuments: len(idx.Documents),
		TotalTerms:     len(idx.Dictionary),
		TotalTokens:    idx.TotalTokens,
		AvgDocLength:   avgdl,
	}
}

// GetAllDocuments trả về danh sách tất cả Document đã được sắp xếp theo ID
func (idx *InvertedIndex) GetAllDocuments() []Document {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	docs := make([]Document, 0, len(idx.Documents))
	for _, doc := range idx.Documents {
		docs = append(docs, doc)
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].ID < docs[j].ID
	})

	return docs
}

// InspectDocument trả về chi tiết các tokens và postings của một Document cụ thể
func (idx *InvertedIndex) InspectDocument(docID int, analyzer Analyzer) *DocumentIndexDetails {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	doc, exists := idx.Documents[docID]
	if !exists {
		return nil
	}

	tokens := analyzer.Analyze(doc.Content)
	var postingList []TermPostingInfo

	// Thu thập các posting thuộc về docID này
	for term, postings := range idx.Dictionary {
		for _, p := range postings {
			if p.DocID == docID {
				postingList = append(postingList, TermPostingInfo{
					Term:          term,
					TermFrequency: p.TermFrequency,
					Positions:     p.Positions,
				})
			}
		}
	}

	// Sắp xếp các term theo bảng chữ cái để hiển thị đẹp mắt
	sort.Slice(postingList, func(i, j int) bool {
		return postingList[i].Term < postingList[j].Term
	})

	return &DocumentIndexDetails{
		DocID:       docID,
		Content:     doc.Content,
		Length:      idx.DocLengths[docID],
		Tokens:      tokens,
		PostingList: postingList,
	}
}
