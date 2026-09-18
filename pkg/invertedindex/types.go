package invertedindex

import "sync"

// ============================================================================
// KIẾN THỨC GO CƠ BẢN:
// - "type ... struct": Trong Go không có class như Java/C++, thay vào đó dùng "struct"
//   để gom nhóm các trường dữ liệu (fields) lại với nhau.
// - Viết hoa chữ cái đầu (VD: ID, Content, Document) có nghĩa là "Exported" (public),
//   các package bên ngoài có thể truy cập được.
// - Viết thường chữ cái đầu (VD: docLength) là private trong nội bộ package.
// ============================================================================

// Document đại diện cho một tin nhắn chat trong hệ thống
type Document struct {
	ID      int    `json:"id"`      // Mã định danh duy nhất của tin nhắn (Document ID)
	Content string `json:"content"` // Nội dung văn bản gốc của tin nhắn (ví dụ: "Tôi là sinh viên")
}

// Posting đại diện cho một bản ghi trong Posting List của một Term (từ vựng)
// Khi một từ xuất hiện trong một tài liệu, ta tạo ra một Posting để ghi nhận.
type Posting struct {
	DocID         int   `json:"doc_id"`         // Tài liệu nào chứa từ này?
	TermFrequency int   `json:"term_frequency"` // Từ này xuất hiện bao nhiêu lần trong tài liệu đó? (TF - Term Frequency)
	Positions     []int `json:"positions"`      // Từ này nằm ở vị trí thứ mấy trong câu? (0, 1, 2...).
}

// InvertedIndex đại diện cho cấu trúc Chỉ mục đảo hoàn chỉnh được lưu trong RAM.
type InvertedIndex struct {
	mu sync.RWMutex // Khóa đồng bộ đa luồng an toàn (Read-Write Mutex)

	// 1. Term Dictionary & Posting Lists:
	Dictionary map[string][]Posting `json:"dictionary"`

	// 2. Dữ liệu bổ trợ để phục vụ hiển thị và thuật toán xếp hạng BM25:
	Documents   map[int]Document `json:"documents"`    // Tra cứu nhanh: DocID -> Nội dung gốc Document
	DocLengths  map[int]int      `json:"doc_lengths"`  // Lưu độ dài (tổng số từ) của từng tin nhắn: DocID -> Số lượng token
	TotalTokens int              `json:"total_tokens"` // Tổng số từ của toàn bộ tin nhắn trong hệ thống
}

// TermPostingInfo dùng để xuất thông tin chi tiết một Term trong Document phục vụ thanh tra
type TermPostingInfo struct {
	Term          string `json:"term"`
	TermFrequency int    `json:"term_frequency"`
	Positions     []int  `json:"positions"`
}

// DocumentIndexDetails chứa đầy đủ thông tin bóc tách của một Document trong Index
type DocumentIndexDetails struct {
	DocID       int               `json:"doc_id"`
	Content     string            `json:"content"`
	Length      int               `json:"length"`
	Tokens      []string          `json:"tokens"`
	PostingList []TermPostingInfo `json:"posting_list"`
}

// IndexStats thống kê tổng quan về Inverted Index
type IndexStats struct {
	TotalDocuments int     `json:"total_documents"`
	TotalTerms     int     `json:"total_terms"`
	TotalTokens    int     `json:"total_tokens"`
	AvgDocLength   float64 `json:"avg_doc_length"`
}

