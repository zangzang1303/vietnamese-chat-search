package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"vietnamese-chat-search/pkg/tokenizer/coccoc"
)

// MessageDoc đại diện cho tài liệu chat trong OpenSearch
type MessageDoc struct {
	ID       int64  `json:"id"`
	Tenant   string `json:"tenant"`
	AppID    string `json:"app_id"`
	UserID   string `json:"user_id"`
	ThreadID string `json:"thread_id"`
	Hide     string `json:"hide"`
	CreateAt int64  `json:"create_at"`
	Text     string `json:"text"`
	TextVi   string `json:"text_vi,omitempty"`
	TextRaw  string `json:"text_raw,omitempty"`
	TextEdge string `json:"text_edge,omitempty"`
}

// Corpus từ vựng và câu mẫu tiếng Việt thực tế trong Chat
var (
	tenants = []string{"tenant_vn_enterprise", "tenant_global_chat", "tenant_fintech_sg"}
	apps    = []string{"zalo_connect", "tele_chat", "internal_web_chat", "mobile_ios", "mobile_android"}
	
	shortPhrases = []string{
		"Chào buổi sáng", "Ok bạn nha", "Alo có ai ở đây không", "Hẹn gặp lại ở quán cà phê",
		"Chúc mừng sinh nhật", "Gửi lại file tài liệu giúp mình", "Chuẩn luôn nhé", "Tuyệt vời",
		"Đã nhận thông tin", "Cảm ơn bạn nhiều", "Tôi yêu Việt Nam", "Sinh viên trường này rất năng động",
		"Học sinh được nghỉ học hôm nay", "Bàn ghế văn phòng mới", "Đi làm đúng giờ", "Nhập học đại học",
		"Trà sữa ngon quá", "Check mail giúp anh", "Báo cáo tiến độ dự án", "Bàn bạc kỹ lại nha",
	}

	mediumPhrases = []string{
		"Hôm nay nhóm mình họp lúc 9 giờ sáng để bàn về tiến độ tích hợp Cốc Cốc Tokenizer vào hệ thống OpenSearch.",
		"Công nghệ thông tin và trí tuệ nhân tạo đang làm thay đổi căn bản cách thức tìm kiếm tin nhắn chat trong doanh nghiệp.",
		"Bạn xem lại bảng thời khóa biểu và danh sách học sinh sinh viên đăng ký học bổng kỳ này đã đầy đủ chưa nhé.",
		"Văn phòng mới trang bị đầy đủ bàn ghế, máy lạnh và phục vụ trà cà phê miễn phí cho toàn thể nhân viên công ty.",
		"Hệ thống tìm kiếm toàn văn bản cần đáp ứng độ trễ dưới 50ms cho hàng triệu người dùng trực tuyến cùng thời điểm.",
		"Kế hoạch nâng cấp OpenSearch lên phiên bản 2.19 cần được kiểm thử hiệu năng chịu tải cẩn thận trước khi lên production.",
		"Bảo hiểm xã hội và các chế độ đãi ngộ dành cho cán bộ công nhân viên được cập nhật chi tiết trong sổ tay nội bộ.",
		"Dự thảo quy chế tuyển sinh đại học năm nay có nhiều điểm mới thuận lợi cho học sinh vùng sâu vùng xa tiếp cận giáo dục.",
	}

	longPhrases = []string{
		"Báo cáo tổng kết quý 3: Doanh số mảng giải pháp phần mềm tăng trưởng 25% so với cùng kỳ năm trước. Đội ngũ kỹ sư đã hoàn thành nghiên cứu và phát triển bộ tách từ vựng tiếng Việt Cốc Cốc kết hợp chỉ mục đảo OpenSearch, giảm thiểu tối đa tình trạng tràn bộ nhớ RAM và tối ưu hóa thời gian index cho hàng trăm triệu tin nhắn mỗi ngày. Kính mời các thành viên tham gia buổi thuyết trình chi tiết vào sáng thứ Hai tuần tới.",
		"Thông báo lịch bảo trì định kỳ hệ thống: Vào lúc 0h00 ngày Chủ Nhật tới đây, đội ngũ quản trị hạ tầng sẽ tiến hành nâng cấp cụm máy chủ OpenSearch và cập nhật plugin phân tích ngôn ngữ tiếng Việt. Trong thời gian bảo trì dự kiến kéo dài 2 tiếng, tính năng tìm kiếm lịch sử chat có thể bị gián đoạn cục bộ. Rất mong quý khách hàng và các đối tác thông cảm cho sự bất tiện này.",
		"Tài liệu hướng dẫn tối ưu hóa hiệu năng cơ sở dữ liệu chat: Khi khối lượng tin nhắn vượt mốc 10 triệu bản ghi, việc phân vùng chỉ mục (sharding) kết hợp tiền xử lý tokenization ở tầng ứng dụng Go sẽ giúp giảm tải đáng kể cho cụm OpenSearch. Đặc biệt, việc sử dụng Edge N-gram từ 3 đến 10 ký tự cho phép người dùng gõ tìm kiếm gợi ý tức thời (instant search) mà không gây quá tải CPU của máy chủ.",
	}

	punctuation = []string{".", "!", "?", "...", " :)", " ^^", " <3", ""}
)

// GenerateRandomMessage sinh ngẫu nhiên một tin nhắn chat phản ánh phân phối thực tế
func GenerateRandomMessage(r *rand.Rand) string {
	dist := r.Float64()
	var base string
	if dist < 0.70 { // 70% ngắn
		base = shortPhrases[r.Intn(len(shortPhrases))]
	} else if dist < 0.90 { // 20% trung bình
		base = mediumPhrases[r.Intn(len(mediumPhrases))]
	} else { // 10% dài
		base = longPhrases[r.Intn(len(longPhrases))]
	}
	punct := punctuation[r.Intn(len(punctuation))]
	return base + punct
}

// GenerateDataset tạo ra tập dữ liệu n documents và lưu ra file NDJSON để phục vụ benchmark
func GenerateDataset(count int, outputPath string, tok *coccoc.Tokenizer) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	r := rand.New(rand.NewSource(42)) // Seed cố định để reproducible

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

	for i := 1; i <= count; i++ {
		text := GenerateRandomMessage(r)
		var textVi string
		if tok != nil {
			seg, err := tok.SegmentOriginal(text)
			if err == nil {
				textVi = seg
			} else {
				textVi = text
			}
		}

		doc := MessageDoc{
			ID:       int64(i),
			Tenant:   tenants[r.Intn(len(tenants))],
			AppID:    apps[r.Intn(len(apps))],
			UserID:   fmt.Sprintf("user_%04d", r.Intn(500)+1),
			ThreadID: fmt.Sprintf("thread_%05d", r.Intn(2000)+1),
			Hide:     "false",
			CreateAt: baseTime + int64(i*1000),
			Text:     text,
			TextVi:   textVi,
			TextRaw:  textVi,
			TextEdge: textVi,
		}

		data, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		f.Write(data)
		f.WriteString("\n")

		if i%100000 == 0 {
			fmt.Printf("   ... Đã sinh %d/%d documents\n", i, count)
		}
	}

	return nil
}

func main() {
	dictPath, _ := filepath.Abs("data/dicts/coccoc")
	tok, err := coccoc.New(dictPath, false)
	if err != nil {
		fmt.Printf("⚠️ Cảnh báo: Không thể khởi tạo Cốc Cốc Tokenizer: %v (sử dụng raw text fallback)\n", err)
	}

	count := 100000
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &count)
	}

	outPath := "benchmarks/dataset_benchmark.jsonl"
	if len(os.Args) > 2 {
		outPath = os.Args[2]
	}
	fmt.Printf("🚀 Đang sinh %d documents mẫu vào %s...\n", count, outPath)
	if err := GenerateDataset(count, outPath, tok); err != nil {
		fmt.Printf("❌ Lỗi sinh dataset: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Đã hoàn thành sinh dataset %d documents thành công!\n", count)
}
