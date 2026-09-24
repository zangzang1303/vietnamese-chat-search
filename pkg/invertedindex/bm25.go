package invertedindex

import "math"

// ============================================================================
// KIẾN THỨC GO CƠ BẢN:
// 1. "const" trong Go:
//    - Dùng để định nghĩa các hằng số không thể thay đổi giá trị trong quá trình chạy.
// 2. Package "math":
//    - math.Log(x): Tính logarit tự nhiên (ln) của x.
//    - Mọi hàm trong package "math" đều nhận và trả về kiểu "float64".
// ============================================================================

const (
	// K1 kiểm soát mức độ bão hòa của tần suất từ (Term Frequency Saturation).
	// Giá trị chuẩn của Apache Lucene và Elasticsearch là 1.2.
	// - Nếu K1 = 0: Tần suất từ không có ý nghĩa (xuất hiện 1 lần hay 100 lần điểm như nhau).
	// - K1 càng lớn: Tần suất từ càng đóng góp nhiều điểm hơn.
	K1 = 1.2

	// B kiểm soát mức độ phạt tài liệu dài (Document Length Normalization).
	// Giá trị chuẩn của Lucene/Elasticsearch là 0.75.
	// - Nếu B = 1: Phạt độ dài tài liệu tối đa (tin nhắn càng dài điểm càng bị giảm mạnh).
	// - Nếu B = 0: Không phạt độ dài tài liệu (tin nhắn dài hay ngắn không bị trừ điểm).
	B = 0.75
)

// CalculateIDF tính toán nghịch đảo tần suất tài liệu (Inverse Document Frequency)
// Công thức chuẩn của Lucene BM25:
//
//	IDF(q) = ln( 1 + (N - n + 0.5) / (n + 0.5) )
//
// Trong đó:
// - N: Tổng số tài liệu trong index (len(idx.Documents))
// - n: Số tài liệu có chứa từ q (len(idx.Dictionary[q]))
func (idx *InvertedIndex) CalculateIDF(term string) float64 {
	// Ép kiểu tổng số document sang float64
	N := float64(len(idx.Documents))

	// Tra từ điển xem có bao nhiêu document chứa term này
	// idx.Dictionary[term] trả về slice các Posting, độ dài slice chính là số document (n)
	n := float64(len(idx.Dictionary[term]))

	// Nếu từ này không xuất hiện trong bất kỳ tài liệu nào, điểm độ hiếm = 0
	if n == 0 {
		return 0
	}

	// Áp dụng công thức logarit:
	return math.Log(1.0 + (N - n + 0.5)/(n + 0.5))
}

// CalculateBM25Score tính điểm số liên quan giữa Document ID và danh sách các từ khóa tìm kiếm (Query Tokens)
// Công thức tổng thể:
//
//	Score(D, Q) = Tổng[ IDF(q) * (TF * (K1 + 1)) / (TF + K1 * (1 - B + B * (|D| / avgdl))) ]
func (idx *InvertedIndex) CalculateBM25Score(docID int, queryTokens []string) float64 {
	// Lấy độ dài của tài liệu này (|D|)
	docLen := float64(idx.DocLengths[docID])

	// Lấy độ dài trung bình của toàn bộ tài liệu trong hệ thống (avgdl)
	avgdl := idx.AvgDocLength()

	// Biến tích lũy tổng điểm
	score := 0.0

	// Duyệt qua từng từ khóa trong câu query của người dùng
	for _, q := range queryTokens {
		// Kiểm tra từ khóa q có tồn tại trong từ điển Inverted Index không
		// Cú pháp Go "val, exists := map[key]":
		// - "postings": giá trị trả về nếu tìm thấy
		// - "exists": boolean (true nếu key có trong map, false nếu không)
		postings, exists := idx.Dictionary[q]
		if !exists {
			continue // Từ khóa này không có trong bất kỳ tài liệu nào -> bỏ qua
		}

		// Tìm xem trong tài liệu "docID" này, từ "q" xuất hiện bao nhiêu lần (TF)
		tf := 0
		for _, p := range postings {
			if p.DocID == docID {
				tf = p.TermFrequency
				break // Đã tìm thấy tài liệu này trong posting list -> dừng vòng lặp
			}
		}

		// Nếu tài liệu này không chứa từ q -> điểm đóng góp của từ q = 0
		if tf == 0 {
			continue
		}

		// 1. Tính độ hiếm của từ (IDF)
		idf := idx.CalculateIDF(q)

		// 2. Tính thành phần điểm tần suất (TF Score) có chuẩn hóa theo độ dài:
		// Tử số: TF * (k1 + 1)
		numerator := float64(tf) * (K1 + 1.0)

		// Mẫu số: TF + k1 * (1 - b + b * (docLen / avgdl))
		denominator := float64(tf) + K1*(1.0-B+B*(docLen/avgdl))

		tfScore := numerator / denominator

		// 3. Cộng dồn vào tổng điểm: Score = Score + (IDF * TF_Score)
		score += idf * tfScore
	}

	return score
}

// CalculateBM25WeightedScore tính điểm BM25 với trọng số đa tầng cho từng nhóm Term (Exact, Phrase, Unaccented, Edge N-gram)
func (idx *InvertedIndex) CalculateBM25WeightedScore(docID int, weightedTerms map[string]float64) float64 {
	docLen := float64(idx.DocLengths[docID])
	avgdl := idx.AvgDocLength()
	if avgdl == 0 {
		avgdl = 10.0
	}

	score := 0.0

	for term, weight := range weightedTerms {
		postings, exists := idx.Dictionary[term]
		if !exists {
			continue
		}

		tf := 0
		for _, p := range postings {
			if p.DocID == docID {
				tf = p.TermFrequency
				break
			}
		}

		if tf == 0 {
			continue
		}

		idf := idx.CalculateIDF(term)
		numerator := float64(tf) * (K1 + 1.0)
		denominator := float64(tf) + K1*(1.0-B+B*(docLen/avgdl))
		tfScore := numerator / denominator

		score += idf * tfScore * weight
	}

	return score
}

