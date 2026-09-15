package invertedindex

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
	ID      int    // Mã định danh duy nhất của tin nhắn (Document ID)
	Content string // Nội dung văn bản gốc của tin nhắn (ví dụ: "Tôi là sinh viên")
}

// Posting đại diện cho một bản ghi trong Posting List của một Term (từ vựng)
// Khi một từ xuất hiện trong một tài liệu, ta tạo ra một Posting để ghi nhận.
type Posting struct {
	DocID         int   // Tài liệu nào chứa từ này?
	TermFrequency int   // Từ này xuất hiện bao nhiêu lần trong tài liệu đó? (TF - Term Frequency)
	Positions     []int // Từ này nằm ở vị trí thứ mấy trong câu? (0, 1, 2...). Dùng []int là slice (mảng động trong Go).
}

// InvertedIndex đại diện cho cấu trúc Chỉ mục đảo hoàn chỉnh được lưu trong RAM.
type InvertedIndex struct {
	// 1. Term Dictionary & Posting Lists:
	// "map[K]V" trong Go là kiểu dữ liệu bảng băm (Hash Map):
	// - Key: string (chính là Term - từ vựng, ví dụ: "học_sinh", "cà_phê")
	// - Value: []Posting (một mảng động/slice chứa danh sách các Posting liên quan)
	// Tra cứu map[term] trong Go có độ phức tạp trung bình là O(1) - cực kỳ nhanh!
	Dictionary map[string][]Posting

	// 2. Dữ liệu bổ trợ để phục vụ hiển thị và thuật toán xếp hạng BM25:
	Documents   map[int]Document // Tra cứu nhanh: DocID -> Nội dung gốc Document
	DocLengths  map[int]int      // Lưu độ dài (tổng số từ) của từng tin nhắn: DocID -> Số lượng token
	TotalTokens int              // Tổng số từ của toàn bộ tin nhắn trong hệ thống (phục vụ tính avgdl - độ dài trung bình)
}
