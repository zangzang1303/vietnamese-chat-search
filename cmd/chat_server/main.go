package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"vietnamese-chat-search/pkg/chat"
	"vietnamese-chat-search/pkg/es"
	"vietnamese-chat-search/pkg/invertedindex"
)

//go:embed web/index.html
var indexHTML []byte

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
	fmt.Println("   HỆ THỐNG TIN NHẮN THỜI GIAN THỰC & CẬP NHẬT CHỈ MỤC TỨC THÌ (CHAT SERVER)    ")
	fmt.Println("       TỰ ĐỘNG RE-INDEX VỚI THUẬT TOÁN BM25 & PHÂN ĐOẠN TỪ TIẾNG VIỆT          ")
	fmt.Println("================================================================================")

	// 1. Khởi tạo Analyzer phù hợp
	dictPath := findDictDir()
	var analyzer invertedindex.Analyzer

	coccocAnalyzer, err := invertedindex.NewCoccocAnalyzer(dictPath, false)
	if err == nil {
		analyzer = coccocAnalyzer
		fmt.Println("✅ Đã kích hoạt: Cốc Cốc Tokenizer (CGO & Double-Array Trie) cho hệ thống chat!")
	} else {
		analyzer = invertedindex.NewVietnameseAnalyzer()
		fmt.Println("ℹ️ Đang chạy chế độ: Vietnamese Analyzer (Dự phòng từ ghép nội bộ).")
	}

	// 2. Khởi tạo ChatManager
	chatManager := chat.NewChatManager(analyzer)

	// 2b. Khởi tạo Elasticsearch Client (Persistent Storage & Dual-search Engine)
	esClient, err := es.NewClient([]string{"http://localhost:9200"})
	if err == nil && esClient.IsAvailable() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		esClient.PrintClusterInfo()
		if err := esClient.EnsureIndices(ctx); err != nil {
			log.Printf("⚠️ Lỗi kiểm tra Elasticsearch mapping: %v", err)
		}
	} else {
		fmt.Println("ℹ️ Elasticsearch đang offline -> Chạy chế độ In-Memory Engine (Go thuần).")
	}

	// 3. Nạp sẵn các tin nhắn mẫu thực tế
	seedMessages(chatManager)

	// 4. Đăng ký các HTTP Handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Đọc trực tiếp từ đĩa để hot-reload HTML ngay khi sửa
		diskPaths := []string{
			"cmd/chat_server/web/index.html",
			"web/index.html",
		}
		for _, p := range diskPaths {
			if data, err := os.ReadFile(p); err == nil {
				w.Write(data)
				return
			}
		}
		w.Write(indexHTML)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		stats := chatManager.GetStats()
		statsMap := map[string]interface{}{
			"total_messages": stats.TotalMessages,
			"total_terms":    stats.TotalTerms,
			"total_tokens":   stats.TotalTokens,
			"avg_doc_length": stats.AvgDocLength,
			"active_rooms":   stats.ActiveRooms,
			"es_online":      esClient != nil && esClient.IsAvailable(),
		}
		json.NewEncoder(w).Encode(statsMap)
	})

	http.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			room := r.URL.Query().Get("room")
			messages := chatManager.ListMessages(room)
			json.NewEncoder(w).Encode(messages)

		case http.MethodPost:
			var req chat.CreateMessageRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error": "JSON không hợp lệ"}`, http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(req.Content) == "" {
				http.Error(w, `{"error": "Nội dung tin nhắn không được để trống"}`, http.StatusBadRequest)
				return
			}

			msg := chatManager.PostMessage(req.Sender, req.Room, req.Content)
			log.Printf("📩 [Tin mới] #ID %d từ '%s' trong '%s': %s", msg.ID, msg.Sender, msg.Room, msg.Content)

			// Lưu trữ bền vững vào Elasticsearch nếu online
			if esClient != nil && esClient.IsAvailable() {
				go func(m chat.Message) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					esClient.IndexSingleMessage(ctx, m, analyzer)
				}(msg)
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(msg)

		default:
			http.Error(w, "Phương thức không hỗ trợ", http.StatusMethodNotAllowed)
		}
	})

	// /api/messages/{id}
	http.HandleFunc("/api/messages/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		idStr := strings.TrimPrefix(r.URL.Path, "/api/messages/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error": "ID không hợp lệ"}`, http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			msg, exists := chatManager.GetMessage(id)
			if !exists {
				http.Error(w, `{"error": "Không tìm thấy tin nhắn"}`, http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(msg)

		case http.MethodPut:
			var req chat.UpdateMessageRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error": "JSON không hợp lệ"}`, http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(req.Content) == "" {
				http.Error(w, `{"error": "Nội dung không được để trống"}`, http.StatusBadRequest)
				return
			}

			updatedMsg, err := chatManager.UpdateMessage(id, req.Content)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusNotFound)
				return
			}
			log.Printf("✏️ [Re-index] Đã sửa tin nhắn #ID %d: %s", id, updatedMsg.Content)

			if esClient != nil && esClient.IsAvailable() {
				go func(m chat.Message) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					esClient.IndexSingleMessage(ctx, m, analyzer)
				}(updatedMsg)
			}

			json.NewEncoder(w).Encode(updatedMsg)

		case http.MethodDelete:
			deleted := chatManager.DeleteMessage(id)
			if !deleted {
				http.Error(w, `{"error": "Không tìm thấy tin nhắn để xóa"}`, http.StatusNotFound)
				return
			}
			log.Printf("🗑️ [Xóa & Evict] Đã gỡ tin nhắn #ID %d khỏi cơ sở dữ liệu và chỉ mục", id)

			if esClient != nil && esClient.IsAvailable() {
				go func(msgId int) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					esClient.DeleteSingleMessage(ctx, msgId)
				}(id)
			}

			w.Write([]byte(`{"success": true}`))

		default:
			http.Error(w, "Phương thức không hỗ trợ", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query().Get("q")
		room := r.URL.Query().Get("room")

		if strings.TrimSpace(query) == "" {
			json.NewEncoder(w).Encode([]chat.ChatSearchResult{})
			return
		}

		results := chatManager.Search(query, room)
		json.NewEncoder(w).Encode(results)
	})

	// /api/search/compare : Endpoint đối soát A/B song song
	http.HandleFunc("/api/search/compare", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query().Get("q")
		room := r.URL.Query().Get("room")

		if strings.TrimSpace(query) == "" {
			json.NewEncoder(w).Encode(&es.ComparisonResult{})
			return
		}

		if esClient != nil && esClient.IsAvailable() {
			comp, err := esClient.CompareSearch(r.Context(), query, room, analyzer)
			if err == nil {
				json.NewEncoder(w).Encode(comp)
				return
			}
			log.Printf("⚠️ Lỗi CompareSearch trên ES: %v", err)
		}

		// Fallback In-memory comparison
		startVN := time.Now()
		tokens := analyzer.Analyze(query)
		vnResults := chatManager.Search(query, room)
		vnLat := time.Since(startVN).Milliseconds()

		// Mock Baseline: Standard whitespace analyzer
		startBase := time.Now()
		stdAnalyzer := &invertedindex.StandardAnalyzer{}
		stdTokens := stdAnalyzer.Analyze(query)
		var baseResults []es.SearchResult
		for _, msg := range chatManager.ListMessages(room) {
			matchCount := 0
			lowerContent := strings.ToLower(msg.Content)
			for _, st := range stdTokens {
				if strings.Contains(lowerContent, st) {
					matchCount++
				}
			}
			if matchCount > 0 {
				score := float64(matchCount) * 1.15
				baseResults = append(baseResults, es.SearchResult{
					Message:   msg,
					Score:     score,
					Highlight: msg.Content,
				})
			}
		}
		baseLat := time.Since(startBase).Milliseconds()

		var convertedVN []es.SearchResult
		for _, r := range vnResults {
			convertedVN = append(convertedVN, es.SearchResult{
				Message:   r.Message,
				Score:     r.Score,
				Highlight: r.Message.Content,
			})
		}

		json.NewEncoder(w).Encode(&es.ComparisonResult{
			Query:               query,
			Tokens:              tokens,
			BaselineResults:     baseResults,
			VietnameseResults:   convertedVN,
			LatencyBaselineMs:   baseLat,
			LatencyVietnameseMs: vnLat,
		})
	})

	http.HandleFunc("/api/inspect/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		idStr := strings.TrimPrefix(r.URL.Path, "/api/inspect/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error": "ID không hợp lệ"}`, http.StatusBadRequest)
			return
		}

		details := chatManager.InspectMessageIndex(id)
		if details == nil {
			http.Error(w, `{"error": "Không tìm thấy tin nhắn"}`, http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(details)
	})

	port := 8080
	fmt.Printf("\n🚀 Máy chủ đã sẵn sàng phục vụ tại: http://localhost:%d\n", port)
	fmt.Println("👉 Mở trình duyệt web của bạn và truy cập địa chỉ trên để bắt đầu thử nghiệm gửi tin, sửa tin và tìm kiếm đối soát A/B!")
	fmt.Println("--------------------------------------------------------------------------------")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func seedMessages(cm *chat.ChatManager) {
	candidates := []string{
		"data/sample_messages.json",
		"../../data/sample_messages.json",
		"../data/sample_messages.json",
	}

	roomMap := map[string]string{
		"general":       "Hội Cà Phê & Đời Sống",
		"engineering":   "Team Dự Án Search Engine",
		"random":        "Góc Tán Gẫu IT",
		"announcements": "Kênh Thông Báo Toàn Công Ty",
		"hr-admin":      "Hành Chính & Nhân Sự",
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			var samples []struct {
				Sender   string `json:"sender"`
				Room     string `json:"room"`
				Content  string `json:"content"`
				Category string `json:"category"`
			}
			if err := json.Unmarshal(data, &samples); err == nil && len(samples) > 0 {
				for _, s := range samples {
					roomName := s.Room
					if friendly, ok := roomMap[s.Room]; ok {
						roomName = friendly
					}
					cm.PostMessage(s.Sender, roomName, s.Content)
				}
				fmt.Printf("✅ Đã nạp tự động %d tin nhắn mẫu thực tế từ %s vào Inverted Index!\n", len(samples), path)
				return
			}
		}
	}

	// Fallback nếu không tìm thấy file json
	cm.PostMessage("Nguyễn Văn An", "Team Dự Án Search Engine", "Hôm nay nhóm mình tập trung bàn bạc kế hoạch dự án tuần tới nhé")
	cm.PostMessage("Trần Thị Mai", "Team Dự Án Search Engine", "Em đang là sinh viên mới nhập học trường bách khoa, mong anh chị giúp đỡ")
	cm.PostMessage("Lê Văn C", "Hội Cà Phê & Đời Sống", "Cuối tuần rủ nhau đi uống cà phê nói chuyện phiếm đi mọi người")
	cm.PostMessage("Phạm Tuấn D", "Hội Cà Phê & Đời Sống", "Quán cà phê mới mở gần trường học có view rất đẹp")
	cm.PostMessage("Hoàng Mai E", "Hội Cà Phê & Đời Sống", "Các em học sinh chuẩn bị bài cho tiết học sáng mai")
	cm.PostMessage("Đỗ Hùng F", "Hội Cà Phê & Đời Sống", "Phòng họp mới sắm bộ bàn ghế gỗ rất sang trọng")
}
