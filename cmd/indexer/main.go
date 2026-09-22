package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"vietnamese-chat-search/pkg/chat"
	"vietnamese-chat-search/pkg/es"
	"vietnamese-chat-search/pkg/invertedindex"
)

func findDictDir() string {
	candidates := []string{
		"data/dicts/coccoc",
		"../../data/dicts/coccoc",
		"../data/dicts/coccoc",
	}
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err == nil {
			if _, err := os.Stat(filepath.Join(abs, "multiterm_trie.dump")); err == nil {
				return abs
			}
		}
	}
	return "data/dicts/coccoc"
}

func main() {
	fmt.Println("================================================================================")
	fmt.Println("       CÔNG CỤ NẠP DỮ LIỆU CHUẨN VÀO ELASTICSEARCH (GO INDEXER CLI)            ")
	fmt.Println("             ĐỒNG BỘ ĐỒNG THỜI VÀO CẢ 2 CHỈ MỤC: BASELINE & VIETNAMESE          ")
	fmt.Println("================================================================================")

	// 1. Khởi tạo Elasticsearch Client
	client, err := es.NewClient([]string{"http://localhost:9200"})
	if err != nil {
		log.Fatalf("❌ Lỗi khởi tạo Elasticsearch Client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.Ping(ctx)
	if err != nil {
		log.Fatalf("❌ Không thể kết nối tới Elasticsearch tại http://localhost:9200.\n   👉 Vui lòng đảm bảo cụm Elasticsearch đã được khởi chạy bằng lệnh: docker compose up -d\n   Chi tiết lỗi: %v", err)
	}

	fmt.Printf("✅ Đã kết nối cụm Elasticsearch: %s (v%s)\n", info.ClusterName, info.Version.Number)

	// 2. Khởi tạo Analyzer (Cốc Cốc Tokenizer hoặc Fallback)
	dictPath := findDictDir()
	var analyzer invertedindex.Analyzer

	coccocAnalyzer, err := invertedindex.NewCoccocAnalyzer(dictPath, false)
	if err == nil {
		analyzer = coccocAnalyzer
		fmt.Println("✅ Đã kích hoạt: Cốc Cốc Tokenizer (C++ CGO & Double-Array Trie) cho quá trình Indexing!")
	} else {
		analyzer = invertedindex.NewVietnameseAnalyzer()
		fmt.Println("ℹ️ Chạy chế độ: Vietnamese Analyzer (Dự phòng từ ghép nội bộ).")
	}

	// 3. Khởi tạo cấu trúc Mapping cho 2 index
	fmt.Println("\n--- BƯỚC 1: KHỞI TẠO MAPPING CHỈ MỤC ---")
	if err := client.EnsureIndices(ctx); err != nil {
		log.Fatalf("❌ Lỗi tạo mapping index: %v", err)
	}

	// 4. Đọc dữ liệu từ file data/sample_messages.json
	fmt.Println("\n--- BƯỚC 2: ĐỌC TẬP DỮ LIỆU MẪU 131 TIN NHẮN ---")
	samplePath := "data/sample_messages.json"
	data, err := os.ReadFile(samplePath)
	if err != nil {
		// Thử tìm ở thư mục cha nếu chạy từ cmd/indexer
		samplePath = "../../data/sample_messages.json"
		data, err = os.ReadFile(samplePath)
		if err != nil {
			log.Fatalf("❌ Không tìm thấy file dữ liệu mẫu tại data/sample_messages.json: %v", err)
		}
	}

	var rawMessages []struct {
		ID        int       `json:"id"`
		Sender    string    `json:"sender"`
		Room      string    `json:"room"`
		Content   string    `json:"content"`
		Category  string    `json:"category"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	if err := json.Unmarshal(data, &rawMessages); err != nil {
		log.Fatalf("❌ Lỗi parse JSON dữ liệu mẫu: %v", err)
	}

	roomMap := map[string]string{
		"general":       "Hội Cà Phê & Đời Sống",
		"engineering":   "Team Dự Án Search Engine",
		"random":        "Góc Tán Gẫu IT",
		"announcements": "Kênh Thông Báo Toàn Công Ty",
		"hr-admin":      "Hành Chính & Nhân Sự",
	}

	var messages []chat.Message
	for _, m := range rawMessages {
		roomName := m.Room
		if friendly, ok := roomMap[m.Room]; ok {
			roomName = friendly
		}
		messages = append(messages, chat.Message{
			ID:        m.ID,
			Sender:    m.Sender,
			Room:      roomName,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}

	fmt.Printf("✅ Đã nạp thành công %d tin nhắn mẫu từ file %s\n", len(messages), samplePath)

	// 5. Nạp hàng loạt vào Elasticsearch bằng Bulk API
	fmt.Println("\n--- BƯỚC 3: NẠP DỮ LIỆU HÀNG LOẠT VÀO ELASTICSEARCH QUA BULK API ---")
	startTime := time.Now()
	if err := client.BulkIndexMessages(ctx, messages, analyzer); err != nil {
		log.Fatalf("❌ Lỗi Bulk Index: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("⚡ Thời gian hoàn thành Bulk Index: %v (Trung bình: %.2f ms / tin nhắn)\n",
		duration, float64(duration.Milliseconds())/float64(len(messages)))

	fmt.Println("\n================================================================================")
	fmt.Println("🎉 QUÁ TRÌNH NẠP DỮ LIỆU VÀO ELASTICSEARCH ĐÃ HOÀN TẤT THÀNH CÔNG 100%!")
	fmt.Println("👉 Dữ liệu hiện đã được lưu trữ vĩnh viễn trên Elasticsearch Volume.")
	fmt.Println("👉 Bạn có thể mở Kibana tại http://localhost:5601 để soi trực tiếp index và documents.")
	fmt.Println("================================================================================")
}
