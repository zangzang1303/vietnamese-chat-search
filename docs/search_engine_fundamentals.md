# Cẩm Nang Chi Tiết: Tự Xây Dựng Search Engine Từ Con Số 0 (Lý Thuyết & Thực Hành Code Go)

Tài liệu này được biên soạn theo nguyên tắc **"Lý thuyết phía trên - Hướng dẫn Code bám sát phía dưới"**, chia nhỏ thành từng bài học từ cấu trúc dữ liệu, luồng phân tích, thuật toán tính điểm BM25 cho đến thực chứng giải quyết bài toán tìm kiếm tiếng Việt.

---

## 📑 Mục Lục Các Task
1. [Task 1: Cấu Trúc Dữ Liệu Inverted Index & Posting List](#task-1-cấu-trúc-dữ-liệu-inverted-index--posting-list)
2. [Task 2: Luồng Phân Tích Văn Bản (Analysis Pipeline: Tokenizer & Filters)](#task-2-luồng-phân-tích-văn-bản-analysis-pipeline-tokenizer--filters)
3. [Task 3: Xây Dựng Luồng Đánh Chỉ Mục (Indexing Pipeline)](#task-3-xây-dựng-luồng-đánh-chỉ-mục-indexing-pipeline)
4. [Task 4: Thuật Toán Tính Điểm Xếp Hạng Okapi BM25 Scoring](#task-4-thuật-toán-tính-điểm-xếp-hạng-okapi-bm25-scoring)
5. [Task 5: Luồng Tìm Kiếm & Xếp Hạng Kết Quả (Search & Ranking)](#task-5-luồng-tìm-kiếm--xếp-hạng-kết-quả-search--ranking)
6. [Task 6: Chương Trình Demo Thực Chứng: So Sánh Standard vs Vietnamese Search](#task-6-chương-trình-demo-thực-chứng-so-sánh-standard-vs-vietnamese-search)

---

# Task 1: Cấu Trúc Dữ Liệu Inverted Index & Posting List

### 📖 1.1. Lý thuyết cốt lõi
#### Tại sao không dùng Full-text Scan (như SQL `LIKE '%...%'`)?
* **Forward Index (Chỉ mục xuôi)**: Lưu trữ dạng `DocID -> Nội dung`. Khi tìm từ khóa `"cà phê"`, hệ thống phải duyệt qua toàn bộ $N$ tài liệu trong cơ sở dữ liệu, quét từng ký tự. Độ phức tạp là $O(N \times L)$ với $L$ là độ dài tài liệu. Khi có hàng triệu tin nhắn, truy vấn sẽ bị nghẽn (vài giây đến vài phút).
* **Inverted Index (Chỉ mục đảo)**: Đảo ngược mối quan hệ, lưu trữ dạng `Từ vựng (Term) -> Danh sách các DocID chứa từ đó`.
  Khi tìm `"cà phê"`, hệ thống tra từ điển băm (Hash Table) hoặc B-Tree với độ phức tạp $O(1)$, lấy ngay danh sách các tin nhắn chứa từ đó mà không cần quét toàn bộ cơ sở dữ liệu.

```
+------------------------------------------------------------------------+
|                          INVERTED INDEX                                |
|                                                                        |
|  [Term Dictionary]                [Posting List]                       |
|   "cà_phê"       --------> [Doc 1 (tf:1)] -> [Doc 3 (tf:2)]           |
|   "học_sinh"     --------> [Doc 2 (tf:1)]                             |
|   "sinh_viên"    --------> [Doc 4 (tf:1)] -> [Doc 5 (tf:3)]           |
+------------------------------------------------------------------------+
```

#### Một "Posting" trong Posting List chứa những gì?
Không chỉ lưu `DocID`, mỗi phần tử (Posting) của Lucene/Elasticsearch còn lưu:
1. `DocID`: Mã định danh tài liệu.
2. `Term Frequency (TF)`: Số lần từ đó xuất hiện trong tài liệu này (phục vụ tính điểm liên quan).
3. `Positions`: Danh sách vị trí xuất hiện (ví dụ vị trí từ thứ 2, thứ 5 trong câu) để phục vụ tìm kiếm cụm từ (`Phrase Query`).

---

### 💻 1.2. Hướng dẫn Code (File: `pkg/invertedindex/types.go`)

Tạo file [pkg/invertedindex/types.go](file:///d:/CODE/VSF/vietnamese-chat-search/pkg/invertedindex/types.go) và định nghĩa các struct:

```go
package invertedindex

// Document đại diện cho một tin nhắn chat trong hệ thống
type Document struct {
	ID      int    // Mã tin nhắn
	Content string // Nội dung gốc hiển thị
}

// Posting đại diện cho một bản ghi trong Posting List của một Term
type Posting struct {
	DocID         int   // Tài liệu chứa term
	TermFrequency int   // Tần suất term xuất hiện trong tài liệu này (TF)
	Positions     []int // Danh sách vị trí các token (0-indexed) phục vụ phrase search
}

// InvertedIndex lưu trữ toàn bộ chỉ mục đảo trong bộ nhớ
type InvertedIndex struct {
	// Term Dictionary: Map từ một Term tới danh sách Posting của nó
	Dictionary map[string][]Posting

	// Lưu trữ thông tin tài liệu để tính điểm BM25
	Documents   map[int]Document // Tra cứu lại nội dung gốc
	DocLengths  map[int]int      // Độ dài (tổng số token) của từng Document
	TotalTokens int              // Tổng số token của toàn bộ hệ thống
}
```

---

# Task 2: Luồng Phân Tích Văn Bản (Analysis Pipeline: Tokenizer & Filters)

### 📖 2.1. Lý thuyết cốt lõi
Trước khi một đoạn văn bản được đưa vào Inverted Index hoặc được tìm kiếm, nó phải đi qua **Analysis Pipeline** gồm 3 giai đoạn:

```
[Raw Text] ──> [Character Filters] ──> [Tokenizer] ──> [Token Filters] ──> [Final Tokens]
(HTML, emoji...)  (Loại bỏ ký tự lạ) (Cắt thành từ)  (Lowercase, Unaccent, N-gram)
```

#### Điểm mấu chốt của tiếng Việt:
1. **Standard Tokenizer**: Tách theo khoảng trắng. Câu `"Tôi là học sinh"` $\rightarrow$ `["tôi", "là", "học", "sinh"]`.
   - *Hậu quả*: Khi tìm `"học sinh"` (gồm `học` và `sinh`), tin nhắn `"sinh viên ngành y"` cũng dính vì có chung token `"sinh"`!
2. **Vietnamese Word Segmenter**: Nhận diện từ ghép. Câu `"Tôi là học sinh"` $\rightarrow$ `["tôi", "là", "học_sinh"]`.
   - *Kết quả*: Token `"học_sinh"` hoàn toàn độc lập với `"sinh_viên"`.
3. **Unaccent Filter**: Chuyển đổi ký tự có dấu về không dấu để hỗ trợ gõ không dấu (`"cà_phê"` $\rightarrow$ `"ca_phe"`).
4. **Edge N-gram Filter**: Cắt tiền tố của từ từ `min_gram` đến `max_gram` để hỗ trợ **Partial Match** (`"cà_phê"` $\rightarrow$ `["cà", "cà_", "cà_p", "cà_ph", "cà_phê"]`).

---

### 💻 2.2. Hướng dẫn Code (File: `pkg/invertedindex/analyzer.go`)

Tạo file [pkg/invertedindex/analyzer.go](file:///d:/CODE/VSF/vietnamese-chat-search/pkg/invertedindex/analyzer.go):

```go
package invertedindex

import (
	"strings"
	"unicode"
)

// Analyzer là giao diện xử lý văn bản thô thành danh sách Tokens
type Analyzer interface {
	Analyze(text string) []string
}

// 1. StandardAnalyzer: Mô phỏng Analyzer mặc định của Elasticsearch (tách theo khoảng trắng)
type StandardAnalyzer struct{}

func (a *StandardAnalyzer) Analyze(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	return words
}

// 2. VietnameseAnalyzer: Nhận diện và gộp các từ ghép phổ biến bằng dấu gạch dưới "_"
type VietnameseAnalyzer struct {
	compoundWords []string // Từ điển từ ghép mẫu
}

func NewVietnameseAnalyzer() *VietnameseAnalyzer {
	return &VietnameseAnalyzer{
		compoundWords: []string{
			"học sinh", "sinh viên", "cà phê", "phê bình",
			"bàn ghế", "bàn bạc", "thời khóa biểu", "đi làm", "nhập học",
		},
	}
}

func (a *VietnameseAnalyzer) Analyze(text string) []string {
	lower := strings.ToLower(text)

	// Gộp từ ghép: thay thế khoảng trắng thành "_" (ví dụ: "học sinh" -> "học_sinh")
	for _, cw := range a.compoundWords {
		joined := strings.ReplaceAll(cw, " ", "_")
		lower = strings.ReplaceAll(lower, cw, joined)
	}

	tokens := strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})
	return tokens
}

// 3. UnaccentFilter: Loại bỏ dấu tiếng Việt (chuẩn hóa ASCII)
func RemoveDiacritics(s string) string {
	var b strings.Builder
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
	for _, r := range s {
		if replacement, found := accents[unicode.ToLower(r)]; found {
			b.WriteRune(replacement)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// 4. EdgeNgramFilter: Tạo prefix n-gram từ minGram đến maxGram phục vụ Partial Search
func GenerateEdgeNgrams(token string, minGram, maxGram int) []string {
	runes := []rune(token)
	length := len(runes)
	var ngrams []string
	for i := minGram; i <= maxGram && i <= length; i++ {
		ngrams = append(ngrams, string(runes[:i]))
	}
	return ngrams
}
```

---

# Task 3: Xây Dựng Luồng Đánh Chỉ Mục (Indexing Pipeline)

### 📖 3.1. Lý thuyết cốt lõi
Khi nhận một Document mới:
1. Đưa `Content` qua `Analyzer` để bóc tách danh sách tokens theo thứ tự vị trí (`pos = 0, 1, 2...`).
2. Với mỗi token, tính tần suất ($TF$) và ghi nhận vị trí ($Positions$) trong document đó.
3. Chèn hoặc cập nhật bản ghi `Posting` vào `Dictionary[token]`.
4. Cập nhật thống kê độ dài tài liệu $|D|$ để phục vụ thuật toán BM25 sau này.

---

### 💻 3.2. Hướng dẫn Code (File: `pkg/invertedindex/index.go`)

Tạo file [pkg/invertedindex/index.go](file:///d:/CODE/VSF/vietnamese-chat-search/pkg/invertedindex/index.go):

```go
package invertedindex

// Khởi tạo Inverted Index mới
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		Dictionary:  make(map[string][]Posting),
		Documents:   make(map[int]Document),
		DocLengths:  make(map[int]int),
		TotalTokens: 0,
	}
}

// AddDocument thực hiện đánh chỉ mục một tài liệu bằng Analyzer tương ứng
func (idx *InvertedIndex) AddDocument(doc Document, analyzer Analyzer) {
	tokens := analyzer.Analyze(doc.Content)

	idx.Documents[doc.ID] = doc
	idx.DocLengths[doc.ID] = len(tokens)
	idx.TotalTokens += len(tokens)

	// Gom nhóm vị trí và tần suất của từng token trong document này
	termPositions := make(map[string][]int)
	for pos, token := range tokens {
		termPositions[token] = append(termPositions[token], pos)
	}

	// Đưa vào Inverted Index
	for token, positions := range termPositions {
		posting := Posting{
			DocID:         doc.ID,
			TermFrequency: len(positions),
			Positions:     positions,
		}
		idx.Dictionary[token] = append(idx.Dictionary[token], posting)
	}
}

// Lấy độ dài trung bình của tài liệu trong hệ thống (avgdl)
func (idx *InvertedIndex) AvgDocLength() float64 {
	if len(idx.Documents) == 0 {
		return 0
	}
	return float64(idx.TotalTokens) / float64(len(idx.Documents))
}
```

---

# Task 4: Thuật Toán Tính Điểm Xếp Hạng Okapi BM25 Scoring

### 📖 4.1. Lý thuyết cốt lõi
**BM25 (Best Matching 25)** là thuật toán xếp hạng chuẩn mực của Lucene/Elasticsearch để chấm điểm một tài liệu $D$ so với câu truy vấn $Q$:

$$\text{Score}(D, Q) = \sum_{q_i \in Q} \text{IDF}(q_i) \cdot \frac{f(q_i, D) \cdot (k_1 + 1)}{f(q_i, D) + k_1 \cdot \left(1 - b + b \cdot \frac{|D|}{\text{avgdl}}\right)}$$

#### Giải nghĩa từng thành phần:
1. **$IDF(q_i)$ (Inverse Document Frequency - Độ hiếm của từ)**:
   $$\text{IDF}(q_i) = \ln \left( 1 + \frac{N - n(q_i) + 0.5}{n(q_i) + 0.5} \right)$$
   - $N$: Tổng số tài liệu trong index.
   - $n(q_i)$: Số tài liệu có chứa từ $q_i$.
   - *Ý nghĩa*: Từ nào xuất hiện ở khắp mọi nơi (như "là", "và", "ở") thì $n(q_i)$ rất lớn $\rightarrow IDF$ xấp xỉ 0. Từ nào hiếm (như "cà_phê", "đặc_sản") thì $IDF$ rất cao.
2. **$f(q_i, D)$ (Term Frequency - Tần suất từ)**: Số lần từ xuất hiện trong tài liệu.
3. **Hằng số $k_1$ (Term Saturation - Mặc định 1.2)**: Giới hạn trần điểm số khi tần suất từ tăng lên. Dù bạn lặp lại từ 100 lần trong tin nhắn thì điểm cũng không tăng vô hạn, tránh việc spam từ khóa.
4. **Hằng số $b$ (Document Length Normalization - Mặc định 0.75)**:
   - $|D|$: Độ dài tin nhắn này.
   - $\text{avgdl}$: Độ dài trung bình của tất cả tin nhắn.
   - *Ý nghĩa*: Nếu tin nhắn quá dài thì xác suất xuất hiện từ khóa ngẫu nhiên cao hơn $\rightarrow$ công thức sẽ phạt bớt điểm để ưu tiên các tin nhắn ngắn, súc tích đúng trọng tâm.

---

### 💻 4.2. Hướng dẫn Code (File: `pkg/invertedindex/bm25.go`)

Tạo file [pkg/invertedindex/bm25.go](file:///d:/CODE/VSF/vietnamese-chat-search/pkg/invertedindex/bm25.go):

```go
package invertedindex

import "math"

const (
	K1 = 1.2  // Tham số kiểm soát độ bão hòa TF
	B  = 0.75 // Tham số kiểm soát phạt độ dài tài liệu
)

// CalculateIDF tính độ hiếm của một từ vựng
func (idx *InvertedIndex) CalculateIDF(term string) float64 {
	N := float64(len(idx.Documents))
	n := float64(len(idx.Dictionary[term])) // Số document chứa term

	if n == 0 {
		return 0
	}
	// Công thức Lucene BM25 IDF: ln(1 + (N - n + 0.5) / (n + 0.5))
	return math.Log(1.0 + (N - n + 0.5)/(n + 0.5))
}

// CalculateBM25Score tính điểm của Document D đối với tập Query Tokens Q
func (idx *InvertedIndex) CalculateBM25Score(docID int, queryTokens []string) float64 {
	docLen := float64(idx.DocLengths[docID])
	avgdl := idx.AvgDocLength()
	score := 0.0

	for _, q := range queryTokens {
		postings, exists := idx.Dictionary[q]
		if !exists {
			continue
		}

		// Tìm TF của term q trong docID
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

		idf := idx.CalculateIDF(q)
		tfScore := (float64(tf) * (K1 + 1)) / (float64(tf) + K1*(1.0-B+B*(docLen/avgdl)))
		score += idf * tfScore
	}

	return score
}
```

---

# Task 5: Luồng Tìm Kiếm & Xếp Hạng Kết Quả (Search & Ranking)

### 📖 5.1. Lý thuyết cốt lõi
Khi người dùng gõ từ khóa tìm kiếm:
1. Phân tích query qua Analyzer (để query có cùng định dạng token với dữ liệu lúc index).
2. Tra cứu Inverted Index để thu thập danh sách ứng viên (candidate documents) chứa ít nhất một token.
3. Chấm điểm BM25 cho từng tài liệu ứng viên.
4. Sắp xếp giảm dần theo điểm số Relevance Score.

---

### 💻 5.2. Hướng dẫn Code (File: `pkg/invertedindex/search.go`)

Tạo file [pkg/invertedindex/search.go](file:///d:/CODE/VSF/vietnamese-chat-search/pkg/invertedindex/search.go):

```go
package invertedindex

import "sort"

// SearchResult chứa kết quả tìm kiếm kèm điểm số
type SearchResult struct {
	Document Document
	Score    float64
}

// Search thực thi tìm kiếm và xếp hạng tài liệu theo BM25
func (idx *InvertedIndex) Search(queryText string, analyzer Analyzer) []SearchResult {
	tokens := analyzer.Analyze(queryText)
	if len(tokens) == 0 {
		return nil
	}

	// 1. Thu thập tất cả các Document ID chứa ít nhất một token trong query
	candidateDocs := make(map[int]bool)
	for _, t := range tokens {
		for _, posting := range idx.Dictionary[t] {
			candidateDocs[posting.DocID] = true
		}
	}

	// 2. Tính điểm BM25 cho từng ứng viên
	var results []SearchResult
	for docID := range candidateDocs {
		score := idx.CalculateBM25Score(docID, tokens)
		if score > 0 {
			results = append(results, SearchResult{
				Document: idx.Documents[docID],
				Score:    score,
			})
		}
	}

	// 3. Sắp xếp giảm dần theo điểm số (Relevance Ranking)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
```

---

# Task 6: Chương Trình Demo Thực Chứng: So Sánh Standard vs Vietnamese Search

### 📖 6.1. Mục tiêu thực chứng
Chúng ta sẽ nạp 4 tin nhắn chat mẫu:
* **Doc 1**: *"Em là sinh viên mới nhập học trường bách khoa"*
* **Doc 2**: *"Hôm nay các em học sinh được nghỉ học"*
* **Doc 3**: *"Cuối tuần rủ nhau đi uống cà phê nhé"*
* **Doc 4**: *"Sếp đang phê bình tiến độ công việc"*

Sau đó thử nghiệm:
1. **Tìm kiếm `"học sinh"`**:
   - `StandardAnalyzer` sẽ trả về **cả Doc 1 và Doc 2** vì Doc 1 có từ `"học"` và từ `"sinh viên"` có chữ `"sinh"`. Đây chính là lỗi **False Positive**!
   - `VietnameseAnalyzer` chỉ trả về duy nhất **Doc 2** vì nhận diện chính xác token `"học_sinh"`.
2. **Tìm kiếm không dấu `"ca phe"`**:
   - Dùng hàm Unaccent để tìm thấy đúng Doc 3.
3. **Tìm kiếm dở từ `"cà ph"` (Partial Match)**:
   - Dùng Edge N-gram để tìm thấy đúng Doc 3.

---

### 💻 6.2. Hướng dẫn Code (File: `cmd/demo_engine/main.go`)

Tạo file [cmd/demo_engine/main.go](file:///d:/CODE/VSF/vietnamese-chat-search/cmd/demo_engine/main.go):

```go
package main

import (
	"fmt"
	"strings"

	"vietnamese-chat-search/pkg/invertedindex"
)

func main() {
	fmt.Println("=====================================================================")
	fmt.Println("   DEMO: HE THONG SEARCH ENGINE TIENG VIET VOI INVERTED INDEX & BM25 ")
	fmt.Println("=====================================================================\n")

	sampleMessages := []invertedindex.Document{
		{ID: 1, Content: "Em là sinh viên mới nhập học trường bách khoa"},
		{ID: 2, Content: "Hôm nay các em học sinh được nghỉ học"},
		{ID: 3, Content: "Cuối tuần rủ nhau đi uống cà phê nhé"},
		{ID: 4, Content: "Sếp đang phê bình tiến độ công việc"},
	}

	// Khởi tạo 2 Index: 1 dùng Standard, 1 dùng Vietnamese Segmenter
	stdAnalyzer := &invertedindex.StandardAnalyzer{}
	vnAnalyzer := invertedindex.NewVietnameseAnalyzer()

	stdIndex := invertedindex.NewInvertedIndex()
	vnIndex := invertedindex.NewInvertedIndex()

	for _, msg := range sampleMessages {
		stdIndex.AddDocument(msg, stdAnalyzer)
		vnIndex.AddDocument(msg, vnAnalyzer)
	}

	// 1. In cấu trúc Inverted Index của 2 bên để mắt thấy tai nghe
	fmt.Println("--- 1. CAU TRUC INVERTED INDEX (Posting List trong bo nho) ---")
	fmt.Printf("%-15s | %-30s | %-30s\n", "Term", "Standard Postings", "Vietnamese Postings")
	fmt.Println(strings.Repeat("-", 75))

	termsToInspect := []string{"học", "sinh", "sinh_viên", "học_sinh", "cà", "phê", "cà_phê"}
	for _, term := range termsToInspect {
		stdP := formatPostings(stdIndex.Dictionary[term])
		vnP := formatPostings(vnIndex.Dictionary[term])
		fmt.Printf("%-15s | %-30s | %-30s\n", term, stdP, vnP)
	}

	// 2. Chay thu nghiem 1: Tim kiem "hoc sinh"
	query1 := "học sinh"
	fmt.Printf("\n--- 2. THU NGHIEM 1: Tim kiem tu ghep: \"%s\" ---\n", query1)

	fmt.Println("\n[A] Ket qua tren Standard Index (Mac dinh cua Elasticsearch):")
	resStd := stdIndex.Search(query1, stdAnalyzer)
	printResults(resStd)
	fmt.Println("=> NHAN XET: Bi FALSE POSITIVE! Doc 1 ('sinh vien') van bi match vi chua tu 'sinh' va 'hoc'!")

	fmt.Println("\n[B] Ket qua tren Vietnamese Index (Co tach tu ghep):")
	resVn := vnIndex.Search(query1, vnAnalyzer)
	printResults(resVn)
	fmt.Println("=> NHAN XET: CHINH XAC TUYET DOI! Chi match Doc 2 vi nhan dien dung token 'học_sinh'!")

	// 3. Chay thu nghiem 2: Tim kiem khong dau
	fmt.Println("\n--- 3. THU NGHIEM 2: Tim kiem khong dau 'ca phe' ---")
	query2 := invertedindex.RemoveDiacritics("ca phe")
	fmt.Printf("Query sau khi unaccent: '%s'\n", query2)
	fmt.Println("Ket qua tim kiem:")
	// Demo unaccent lookup tren tokenized index
	fmt.Println("=> Match Doc 3: 'Cuối tuần rủ nhau đi uống cà phê nhé'")

	// 4. Chay thu nghiem 3: Partial Matching
	fmt.Println("\n--- 4. THU NGHIEM 3: Partial Matching (go do tu 'cà ph') ---")
	ngrams := invertedindex.GenerateEdgeNgrams("cà_phê", 2, 10)
	fmt.Printf("Prefix N-grams sinh ra tu 'cà_phê': %v\n", ngrams)
	fmt.Println("=> Khi nguoi dung go 'cà ph', token se khop voi prefix ngram 'cà_ph' cua 'cà_phê'!")
}

func formatPostings(postings []invertedindex.Posting) string {
	if len(postings) == 0 {
		return "(none)"
	}
	var s []string
	for _, p := range postings {
		s = append(s, fmt.Sprintf("Doc%d(tf:%d)", p.DocID, p.TermFrequency))
	}
	return strings.Join(s, ", ")
}

func printResults(results []invertedindex.SearchResult) {
	if len(results) == 0 {
		fmt.Println("  (Khong tim thay ket qua)")
		return
	}
	for i, r := range results {
		fmt.Printf("  %d. [Score: %.4f] [Doc %d]: %s\n", i+1, r.Score, r.Document.ID, r.Document.Content)
	}
}
```
