package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	// Import package nội bộ trong dự án theo tên module khai báo ở go.mod
	"vietnamese-chat-search/pkg/invertedindex"
)

// ============================================================================
// KIẾN THỨC GO CƠ BẢN VỀ NHẬP DỮ LIỆU TỪ BÀN PHÍM:
// 1. "bufio.NewScanner(os.Stdin)":
//    - os.Stdin đại diện cho Standard Input (luồng dữ liệu từ bàn phím).
//    - bufio.Scanner là công cụ đọc dữ liệu theo từng dòng (Line by Line) an toàn nhất,
//      đặc biệt hỗ trợ đọc tốt chuỗi tiếng Việt có dấu và có chứa khoảng trắng.
// 2. "scanner.Scan()":
//    - Tạm dừng chương trình và đợi người dùng gõ phím rồi nhấn ENTER.
//    - Trả về true nếu đọc thành công, false nếu gặp lỗi hoặc kết thúc luồng.
// 3. "scanner.Text()":
//    - Lấy chuỗi ký tự mà người dùng vừa gõ vào.
// 4. "strings.TrimSpace()":
//    - Cắt bỏ khoảng trắng thừa và ký tự xuống dòng (\r, \n) ở 2 đầu chuỗi.
// 5. Vòng lặp vô tận "for { ... }":
//    - Go không có vòng lặp "while". Để lặp vô tận, ta chỉ cần viết `for { ... }`.
//    - Lệnh "break" dùng để thoát khỏi vòng lặp khi người dùng gõ "exit".
// ============================================================================

func main() {
	fmt.Println("================================================================================")
	fmt.Println("    CONG CU TIM KIEM TIN NHAN TIENG VIET TUONG TAC (INTERACTIVE SEARCH REPL)    ")
	fmt.Println("================================================================================")
	fmt.Println()

	// 1. Tạo tập dữ liệu tin nhắn chat mẫu đa dạng
	sampleMessages := []invertedindex.Document{
		{ID: 1, Content: "Em là sinh viên mới nhập học trường bách khoa"},
		{ID: 2, Content: "Hôm nay các em học sinh được nghỉ học bài"},
		{ID: 3, Content: "Cuối tuần rủ nhau đi uống cà phê nói chuyện nhé"},
		{ID: 4, Content: "Sếp đang phê bình tinh thần làm việc của cả nhóm"},
		{ID: 5, Content: "Phòng họp mới mua bộ bàn ghế gỗ rất đẹp"},
		{ID: 6, Content: "Mọi người tập trung bàn bạc kế hoạch dự án tuần tới"},
		{ID: 7, Content: "Đã có thời khóa biểu học kỳ mới rồi nhé"},
		{ID: 8, Content: "Hôm nay mình bận đi làm thêm từ sáng đến tối"},
	}

	// 2. Khởi tạo 2 bộ phân tích từ vựng (Analyzer)
	stdAnalyzer := &invertedindex.StandardAnalyzer{}
	vnAnalyzer := invertedindex.NewVietnameseAnalyzer()

	// 3. Khởi tạo 2 Inverted Index trong RAM
	stdIndex := invertedindex.NewInvertedIndex()
	vnIndex := invertedindex.NewInvertedIndex()

	// 4. Nạp dữ liệu vào cả 2 Index
	for _, msg := range sampleMessages {
		stdIndex.AddDocument(msg, stdAnalyzer)
		vnIndex.AddDocument(msg, vnAnalyzer)
	}

	// Hiển thị danh sách tin nhắn hiện có trong hệ thống
	printDatabase(sampleMessages)

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("HUONG DAN:")
	fmt.Println("- Nhap bat ky tu khoa nao de tim kiem (vi du: 'học sinh', 'ca phe', 'cà ph', 'bàn ghế')...")
	fmt.Println("- Go 'list' de xem lai toan bo tin nhan trong database.")
	fmt.Println("- Go 'exit' hoac 'q' de thoat chuong trinh.")
	fmt.Println("--------------------------------------------------------------------------------")

	// Tạo bộ đọc dữ liệu từ bàn phím
	scanner := bufio.NewScanner(os.Stdin)

	// Vòng lặp tương tác (REPL Loop)
	for {
		fmt.Print("\n🔍 Nhập từ khóa tìm kiếm: ")

		// Chờ người dùng nhập xong và bấm Enter
		if !scanner.Scan() {
			break
		}

		// Lấy chuỗi vừa nhập và làm sạch khoảng trắng
		input := strings.TrimSpace(scanner.Text())

		// Nếu người dùng chỉ bấm Enter mà không gõ gì -> bỏ qua hỏi lại
		if input == "" {
			continue
		}

		// Kiểm tra lệnh thoát
		if input == "exit" || input == "q" {
			fmt.Println("\n👋 Đã thoát chương trình tìm kiếm. Tạm biệt!")
			break
		}

		// Kiểm tra lệnh xem danh sách tin nhắn
		if input == "list" {
			printDatabase(sampleMessages)
			continue
		}

		// --------------------------------------------------------------------
		// THỰC THI TÌM KIẾM VÀ SO SÁNH
		// --------------------------------------------------------------------
		fmt.Printf("\n--- KẾT QUẢ TÌM KIẾM CHO TỪ KHÓA: \"%s\" ---\n", input)

		// 1. Tìm trên Standard Index (Mặc định của Elasticsearch)
		fmt.Println("\n[A] STANDARD ANALYZER (Tách theo khoảng trắng - Mặc định ES):")
		tokensStd := stdAnalyzer.Analyze(input)
		fmt.Printf("    -> Tokens được tạo ra: %v\n", tokensStd)
		resStd := stdIndex.Search(input, stdAnalyzer)
		printResults(resStd)

		// 2. Tìm trên Vietnamese Index (Tách từ ghép - Cốc Cốc Tokenizer)
		fmt.Println("\n[B] VIETNAMESE ANALYZER (Tách từ ghép tiếng Việt):")
		tokensVn := vnAnalyzer.Analyze(input)
		fmt.Printf("    -> Tokens được tạo ra: %v\n", tokensVn)
		resVn := vnIndex.Search(input, vnAnalyzer)
		printResults(resVn)

		// 3. Tìm không dấu (Unaccent Search)
		unaccentQuery := invertedindex.RemoveDiacritics(input)
		fmt.Printf("\n[C] UNACCENT SEARCH (Tìm kiếm không dấu): \"%s\"\n", unaccentQuery)
		matchCount := 0
		for _, doc := range sampleMessages {
			docUnaccent := invertedindex.RemoveDiacritics(doc.Content)
			if strings.Contains(strings.ToLower(docUnaccent), strings.ToLower(unaccentQuery)) {
				matchCount++
				fmt.Printf("    ✓ [Khớp không dấu] [Doc %d]: %s\n", doc.ID, doc.Content)
			}
		}
		if matchCount == 0 {
			fmt.Println("    (Không có tin nhắn nào khớp không dấu)")
		}

		// 4. Đưa ra nhận xét tự động so sánh
		printComparisonAnalysis(input, resStd, resVn)
	}
}

// printDatabase in ra bảng danh sách tin nhắn hiện có trong hệ thống
func printDatabase(messages []invertedindex.Document) {
	fmt.Println("\nDANH SÁCH TIN NHẮN TRONG HỆ THỐNG (Database):")
	fmt.Println(strings.Repeat("-", 75))
	for _, m := range messages {
		fmt.Printf("  [Doc %d]: %s\n", m.ID, m.Content)
	}
	fmt.Println(strings.Repeat("-", 75))
}

// printResults in danh sách kết quả kèm điểm BM25
func printResults(results []invertedindex.SearchResult) {
	if len(results) == 0 {
		fmt.Println("    (Không tìm thấy tin nhắn nào)")
		return
	}
	for i, r := range results {
		fmt.Printf("    %d. [Điểm BM25: %.4f] [Doc %d]: %s\n", i+1, r.Score, r.Document.ID, r.Document.Content)
	}
}

// printComparisonAnalysis đưa ra nhận định tại sao kết quả lại khác biệt
func printComparisonAnalysis(query string, stdRes, vnRes []invertedindex.SearchResult) {
	fmt.Println("\n💡 PHÂN TÍCH SO SÁNH:")
	if len(stdRes) > len(vnRes) {
		fmt.Printf("  -> Cảnh báo: Standard Search trả về %d kết quả, trong khi Vietnamese Search chỉ trả về %d kết quả.\n", len(stdRes), len(vnRes))
		fmt.Println("  -> Lý do: Standard Analyzer bị hiện tượng FALSE POSITIVE do bẻ gãy từ ghép tiếng Việt thành các từ đơn lẻ!")
	} else if len(vnRes) > 0 {
		fmt.Println("  -> Vietnamese Search nhận diện chính xác cấu trúc từ vựng tiếng Việt, điểm số tập trung đúng tài liệu mục tiêu.")
	} else {
		fmt.Println("  -> Không có kết quả trực tiếp từ index có dấu, hãy xem kết quả ở mục [C] Unaccent Search.")
	}
	fmt.Println(strings.Repeat("-", 75))
}
