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
	"vietnamese-chat-search/pkg/customengine"
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

	// 2c. Khởi tạo Custom Search & Binary Disk Storage Engine (Động cơ lưu trữ đĩa riêng tự xây dựng)
	customStorageDir := "data/custom_storage"
	customEngine, err := customengine.Open(customStorageDir, analyzer)
	if err != nil {
		log.Printf("⚠️ Lỗi khởi tạo Custom Engine: %v", err)
	} else {
		fmt.Printf("✅ Đã kích hoạt: Custom Engine (Binary Disk Storage tại '%s')!\n", customStorageDir)
	}

	// 3. Khôi phục toàn bộ tin nhắn từ Custom Engine đĩa (nếu đã có dữ liệu), hoặc nạp mẫu nếu chưa có
	if customEngine != nil && customEngine.Stats().TotalDocs > 0 {
		savedMsgs := customEngine.GetAllMessages()
		if len(savedMsgs) > 0 {
			var msgs []chat.Message
			for _, sm := range savedMsgs {
				msgs = append(msgs, *sm)
			}
			chatManager.LoadMessages(msgs)
			fmt.Printf("💾 Đã khôi phục thành công %d tin nhắn bền vững từ Custom Storage Engine trên đĩa vào ChatManager!\n", len(msgs))

			// Tái sinh Edge N-grams toàn diện cho Custom Engine từ DocStore
			if err := customEngine.RebuildIndexFromDocStore(); err == nil {
				_ = customEngine.Flush()
				fmt.Printf("⚡ [Custom Engine] Đã sinh đầy đủ Edge N-grams & Space forms cho %d tin nhắn trên đĩa!\n", len(msgs))
			}

			// Tự động đồng bộ toàn bộ tin nhắn vào Elasticsearch
			if esClient != nil {
				go func(all []chat.Message) {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := esClient.BulkIndexMessages(ctx, all, analyzer); err != nil {
						log.Printf("⚠️ Lỗi đồng bộ Elasticsearch lúc khởi động: %v", err)
					} else {
						fmt.Printf("🐘 Đã tự động đồng bộ %d tin nhắn vào Elasticsearch (Baseline & Vietnamese)!\n", len(all))
					}
				}(msgs)
			}
		} else {
			seedMessages(chatManager)
		}
	} else {
		seedMessages(chatManager)
		// Nạp dữ liệu vào Custom Engine nếu đĩa còn trống
		if customEngine != nil && customEngine.Stats().TotalDocs == 0 {
			allMsgs := chatManager.ListMessages("")
			for _, m := range allMsgs {
				_ = customEngine.IndexMessage(m)
			}
			_ = customEngine.Flush()
			fmt.Printf("💾 Đã Flush %d tin nhắn mẫu vào Custom Storage Engine trên đĩa!\n", len(allMsgs))
		}
	}

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

			// Lưu trữ tức thì vào Custom Storage Engine trên đĩa
			if customEngine != nil {
				_ = customEngine.IndexMessage(msg)
			}

			// Lưu trữ tự động vào Elasticsearch nếu esClient đã được khởi tạo
			if esClient != nil {
				go func(m chat.Message) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := esClient.IndexSingleMessage(ctx, m, analyzer); err != nil {
						log.Printf("⚠️ [Elasticsearch] Lỗi lưu tự động tin nhắn #ID %d: %v", m.ID, err)
					} else {
						log.Printf("🐘 [Elasticsearch] Đã lưu tự động tin nhắn #ID %d vào 2 index (Baseline & Vietnamese)", m.ID)
					}
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

			// Cập nhật Custom Engine trên đĩa
			if customEngine != nil {
				_ = customEngine.DeleteMessage(id)
				_ = customEngine.IndexMessage(updatedMsg)
			}

			if esClient != nil {
				go func(m chat.Message) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := esClient.IndexSingleMessage(ctx, m, analyzer); err != nil {
						log.Printf("⚠️ [Elasticsearch] Lỗi cập nhật tin nhắn #ID %d: %v", m.ID, err)
					} else {
						log.Printf("🐘 [Elasticsearch] Đã cập nhật tin nhắn #ID %d vào Elasticsearch", m.ID)
					}
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

			// Đánh dấu xóa trên Custom Engine đĩa (Tombstone)
			if customEngine != nil {
				_ = customEngine.DeleteMessage(id)
			}

			if esClient != nil {
				go func(msgId int) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := esClient.DeleteSingleMessage(ctx, msgId); err != nil {
						log.Printf("⚠️ [Elasticsearch] Lỗi xóa tin nhắn #ID %d: %v", msgId, err)
					} else {
						log.Printf("🐘 [Elasticsearch] Đã xóa tin nhắn #ID %d khỏi Elasticsearch", msgId)
					}
				}(id)
			}

			w.Write([]byte(`{"success": true}`))

		default:
			http.Error(w, "Phương thức không hỗ trợ", http.StatusMethodNotAllowed)
		}
	})

	// /api/search : Endpoint tìm kiếm động hỗ trợ tùy chọn Tokenizer (Cốc Cốc vs Standard) và Động cơ (Elasticsearch vs Custom Engine)
	type FlowSearchResultItem struct {
		Message   chat.Message `json:"message"`
		Score     float64      `json:"score"`
		Highlight string       `json:"highlight,omitempty"`
		Tokens    []string     `json:"tokens,omitempty"`
	}

	type FlowSearchResponse struct {
		Query           string                 `json:"query"`
		Tokenizer       string                 `json:"tokenizer"`       // "coccoc" | "standard"
		TokenizerName   string                 `json:"tokenizer_name"`  // "Cốc Cốc Tokenizer (Từ ghép)" | "Standard Tokenizer (Khoảng trắng)"
		Engine          string                 `json:"engine"`          // "es" | "custom"
		EngineName      string                 `json:"engine_name"`     // "Elasticsearch 8.x" | "Custom Engine (Go Binary Disk)"
		Tokens          []string               `json:"tokens"`
		UnaccentedQuery string                 `json:"unaccented_query"`
		EdgeNgrams      []string               `json:"edge_ngrams,omitempty"`
		StorageDetails  string                 `json:"storage_details"`
		ScoringFormula  string                 `json:"scoring_formula"`
		LatencyMs       int64                  `json:"latency_ms"`
		TotalHits       int                    `json:"total_hits"`
		Results         []FlowSearchResultItem `json:"results"`
	}

	http.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query().Get("q")
		room := r.URL.Query().Get("room")
		engineParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("engine")))
		if engineParam == "" {
			engineParam = "es"
		}
		tokenizerParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tokenizer")))
		if tokenizerParam == "" {
			tokenizerParam = "coccoc"
		}

		if strings.TrimSpace(query) == "" {
			json.NewEncoder(w).Encode(FlowSearchResponse{
				Query:   "",
				Results: []FlowSearchResultItem{},
			})
			return
		}

		// Chuẩn bị analyzer phù hợp
		stdAnalyzer := &invertedindex.StandardAnalyzer{}
		var activeAnalyzer invertedindex.Analyzer = analyzer
		tokenizerName := "Cốc Cốc Tokenizer (Double-Array Trie)"
		if tokenizerParam == "standard" {
			activeAnalyzer = stdAnalyzer
			tokenizerName = "Standard Tokenizer (Whitespace / Unigram)"
		}

		tokens := activeAnalyzer.Analyze(query)
		unaccentedQuery := invertedindex.RemoveDiacritics(strings.Join(tokens, " "))

		var (
			results        []FlowSearchResultItem
			latencyMs      int64
			engineName     string
			storageDetails string
			scoringFormula string
		)

		if engineParam == "es" {
			engineName = "Elasticsearch 8.x"
			if esClient != nil && esClient.IsAvailable() {
				if tokenizerParam == "standard" {
					// Tìm trên index baseline của Elasticsearch
					storageDetails = "Lucene Inverted Index (Index: chat_messages_baseline)"
					scoringFormula = "Okapi BM25 Standard (Single-field content, k1=1.2, b=0.75)"
					esRes, lat, err := esClient.SearchBaseline(r.Context(), query, room)
					latencyMs = lat
					if err == nil {
						for _, er := range esRes {
							results = append(results, FlowSearchResultItem{
								Message:   er.Message,
								Score:     er.Score,
								Highlight: er.Highlight,
								Tokens:    tokens,
							})
						}
					}
				} else {
					// Tìm trên index vietnamese của Elasticsearch (Cốc Cốc + Multi-field Boosting)
					storageDetails = "Lucene Inverted Index (Index: chat_messages_vietnamese)"
					scoringFormula = "Okapi BM25 + Boosting (content_tokenized^5.0, unaccented^3.0, phrase^4.0)"
					esRes, lat, _, err := esClient.SearchVietnamese(r.Context(), query, room, activeAnalyzer)
					latencyMs = lat
					if err == nil {
						for _, er := range esRes {
							results = append(results, FlowSearchResultItem{
								Message:   er.Message,
								Score:     er.Score,
								Highlight: er.Highlight,
								Tokens:    tokens,
							})
						}
					}
				}
			} else {
				// Elasticsearch offline -> fallback thông báo rõ
				storageDetails = "Elasticsearch 8.x đang Offline -> Fallback In-Memory Engine"
				scoringFormula = "Okapi BM25 In-Memory"
				start := time.Now()
				ramRes := chatManager.Search(query, room)
				latencyMs = time.Since(start).Milliseconds()
				for _, rr := range ramRes {
					results = append(results, FlowSearchResultItem{
						Message:   rr.Message,
						Score:     rr.Score,
						Tokens:    rr.Tokens,
						Highlight: rr.Message.Content,
					})
				}
			}
		} else {
			// engineParam == "custom"
			engineName = "Custom Storage Engine (Go Disk Segments)"
			storageDetails = "Binary Disk Storage (data/custom_storage: terms.dict, postings.bin, docstore.data)"
			if tokenizerParam == "standard" {
				scoringFormula = "Okapi BM25 Unigram (k1=1.2, b=0.75, standard terms)"
			} else {
				scoringFormula = "Okapi BM25 + Compound Boosting (5.0x Tokenized, 4.0x Phrase, 3.0x Unaccented)"
			}

			if customEngine != nil {
				cRes, lat, _, err := customEngine.SearchWithAnalyzer(query, room, activeAnalyzer)
				latencyMs = lat
				if err == nil {
					for _, cr := range cRes {
						results = append(results, FlowSearchResultItem{
							Message:   cr.Message,
							Score:     cr.Score,
							Highlight: cr.Highlight,
							Tokens:    tokens,
						})
					}
				}
			} else {
				start := time.Now()
				ramRes := chatManager.Search(query, room)
				latencyMs = time.Since(start).Milliseconds()
				for _, rr := range ramRes {
					results = append(results, FlowSearchResultItem{
						Message:   rr.Message,
						Score:     rr.Score,
						Tokens:    rr.Tokens,
						Highlight: rr.Message.Content,
					})
				}
			}
		}

		// Bổ sung danh sách Edge N-grams sinh ra từ tokens
		var allEdgeNgrams []string
		seenNg := make(map[string]bool)
		for _, t := range tokens {
			for _, ng := range invertedindex.GenerateEdgeNgrams(t, 2, 15) {
				if !seenNg[ng] {
					seenNg[ng] = true
					allEdgeNgrams = append(allEdgeNgrams, ng)
				}
				if strings.Contains(ng, "_") {
					spaceNg := strings.ReplaceAll(ng, "_", " ")
					if !seenNg[spaceNg] {
						seenNg[spaceNg] = true
						allEdgeNgrams = append(allEdgeNgrams, spaceNg)
					}
				}
			}
		}

		resp := FlowSearchResponse{
			Query:           query,
			Tokenizer:       tokenizerParam,
			TokenizerName:   tokenizerName,
			Engine:          engineParam,
			EngineName:      engineName,
			Tokens:          tokens,
			UnaccentedQuery: unaccentedQuery,
			EdgeNgrams:      allEdgeNgrams,
			StorageDetails:  storageDetails,
			ScoringFormula:  scoringFormula,
			LatencyMs:       latencyMs,
			TotalHits:       len(results),
			Results:         results,
		}
		json.NewEncoder(w).Encode(resp)
	})

	// /api/search/compare : Endpoint đối soát song song cả 3 động cơ (Custom vs ES Vietnamese vs ES Baseline)
	http.HandleFunc("/api/search/compare", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query().Get("q")
		room := r.URL.Query().Get("room")

		if strings.TrimSpace(query) == "" {
			json.NewEncoder(w).Encode(&es.ComparisonResult{})
			return
		}

		// 1. Thực thi tìm kiếm trên Custom Engine (Go Binary Disk Storage)
		startCustom := time.Now()
		var customResults []es.SearchResult
		if customEngine != nil {
			cRes, _, err := customEngine.Search(query, room)
			if err == nil {
				for _, cr := range cRes {
					customResults = append(customResults, es.SearchResult{
						Message:   cr.Message,
						Score:     cr.Score,
						Highlight: cr.Highlight,
					})
				}
			}
		}
		customLat := time.Since(startCustom).Milliseconds()

		// 2. Thực thi tìm kiếm trên Elasticsearch 8.x nếu có
		if esClient != nil && esClient.IsAvailable() {
			comp, err := esClient.CompareSearch(r.Context(), query, room, analyzer)
			if err == nil {
				comp.CustomResults = customResults
				comp.LatencyCustomMs = customLat
				json.NewEncoder(w).Encode(comp)
				return
			}
			log.Printf("⚠️ Lỗi CompareSearch trên ES: %v", err)
		}

		// 3. Fallback In-memory comparison nếu ES offline
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
			CustomResults:       customResults,
			LatencyBaselineMs:   baseLat,
			LatencyVietnameseMs: vnLat,
			LatencyCustomMs:     customLat,
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
