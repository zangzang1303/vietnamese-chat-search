package invertedindex

// ============================================================================
// KIẾN THỨC GO CƠ BẢN:
// 1. Hàm "make()":
//    - Map và Slice trong Go là kiểu tham chiếu (reference type).
//    - Nếu chỉ khai báo `var m map[string]int` thì giá trị mặc định là `nil`.
//      Nếu bạn ghi dữ liệu vào một map nil (m["a"] = 1), Go sẽ PANIC (crash app)!
//    - Vì vậy, luôn luôn dùng `make(map[...])` để cấp phát vùng nhớ trước khi dùng.
// 2. Con trỏ (Pointer) trong Receiver:
//    - `(idx *InvertedIndex)`: Dấu `*` có nghĩa là hàm này nhận một con trỏ tới struct.
//    - Nhờ con trỏ, hàm có thể SỬA ĐỔI TRỰC TIẾP dữ liệu bên trong `idx`.
//    - Nếu không dùng `*`, Go sẽ tạo một BẢN SAO của struct, mọi thao tác sửa đổi
//      sẽ bị hủy bỏ khi hàm chạy xong!
// 3. Ép kiểu tường minh (Explicit Casting):
//    - Go cực kỳ nghiêm ngặt về kiểu dữ liệu: không tự động đổi int thành float!
//    - Để chia lấy số thập phân: bắt buộc viết `float64(a) / float64(b)`.
// ============================================================================

// NewInvertedIndex là hàm khởi tạo để tạo ra một Inverted Index rỗng trong RAM
func NewInvertedIndex() *InvertedIndex {
	// Trả về địa chỉ (&) của struct mới được cấp phát vùng nhớ
	return &InvertedIndex{
		// Cấp phát bộ nhớ cho các map bằng hàm make()
		Dictionary:  make(map[string][]Posting),
		Documents:   make(map[int]Document),
		DocLengths:  make(map[int]int),
		TotalTokens: 0,
	}
}

// AddDocument thực hiện đánh chỉ mục một tài liệu vào Inverted Index
// Tham số:
// - doc: tin nhắn cần index (chứa ID và Content)
// - analyzer: bộ phân tích từ vựng muốn dùng (Standard hoặc Vietnamese)
func (idx *InvertedIndex) AddDocument(doc Document, analyzer Analyzer) {
	// BƯỚC 1: Bóc tách nội dung câu thành danh sách các từ vựng (tokens)
	// Ví dụ: "em học sinh đi học" -> ["em", "học_sinh", "đi", "học"]
	tokens := analyzer.Analyze(doc.Content)

	// BƯỚC 2: Lưu lại tài liệu gốc và thống kê độ dài (|D|)
	idx.Documents[doc.ID] = doc
	idx.DocLengths[doc.ID] = len(tokens) // Số lượng token trong câu này
	idx.TotalTokens += len(tokens)       // Cộng dồn vào tổng số token toàn hệ thống

	// BƯỚC 3: Thu thập các vị trí (Positions) của từng token trong câu này
	// Ví dụ với từ "học": xuất hiện ở vị trí thứ 1 và thứ 3 -> termPositions["học"] = [1, 3]
	termPositions := make(map[string][]int)
	for pos, token := range tokens {
		// append() tự động mở rộng mảng khi thêm phần tử mới
		termPositions[token] = append(termPositions[token], pos)
	}

	// BƯỚC 4: Tạo Posting và chèn vào Posting List của từng Term trong Inverted Index
	for token, positions := range termPositions {
		posting := Posting{
			DocID:         doc.ID,
			TermFrequency: len(positions), // Tần suất từ (TF) chính là số lần xuất hiện
			Positions:     positions,      // Danh sách vị trí
		}

		// idx.Dictionary[token] là slice các Posting.
		// append() sẽ thêm posting của tài liệu này vào cuối danh sách của từ đó!
		idx.Dictionary[token] = append(idx.Dictionary[token], posting)
	}
}

// AvgDocLength tính toán độ dài trung bình của tất cả tài liệu trong index (avgdl)
// Công thức: avgdl = Tổng số token của tất cả tài liệu / Tổng số tài liệu
func (idx *InvertedIndex) AvgDocLength() float64 {
	// Nếu chưa có tài liệu nào, trả về 0 để tránh lỗi chia cho 0 (Divide by Zero)
	if len(idx.Documents) == 0 {
		return 0
	}

	// Ép kiểu sang float64 để thực hiện phép chia lấy số thực chính xác
	return float64(idx.TotalTokens) / float64(len(idx.Documents))
}
