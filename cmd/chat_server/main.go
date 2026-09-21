package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"vietnamese-chat-search/pkg/chat"
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

	// 3. Nạp sẵn các tin nhắn mẫu thực tế
	seedMessages(chatManager)

	// 4. Đăng ký các HTTP Handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		stats := chatManager.GetStats()
		json.NewEncoder(w).Encode(stats)
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
			json.NewEncoder(w).Encode(updatedMsg)

		case http.MethodDelete:
			deleted := chatManager.DeleteMessage(id)
			if !deleted {
				http.Error(w, `{"error": "Không tìm thấy tin nhắn để xóa"}`, http.StatusNotFound)
				return
			}
			log.Printf("🗑️ [Xóa & Evict] Đã gỡ tin nhắn #ID %d khỏi cơ sở dữ liệu và chỉ mục", id)
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
	fmt.Println("👉 Mở trình duyệt web của bạn và truy cập địa chỉ trên để bắt đầu thử nghiệm gửi tin, sửa tin và tìm kiếm!")
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
