package chat

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"vietnamese-chat-search/pkg/invertedindex"
)

// ChatManager quản lý luồng tin nhắn và đồng bộ tức thì với Inverted Index
type ChatManager struct {
	mu          sync.RWMutex
	messages    map[int]Message
	nextID      int
	index       *invertedindex.InvertedIndex
	analyzer    invertedindex.Analyzer
}

// NewChatManager khởi tạo hệ thống quản lý chat mới
func NewChatManager(analyzer invertedindex.Analyzer) *ChatManager {
	if analyzer == nil {
		analyzer = invertedindex.NewVietnameseAnalyzer()
	}

	return &ChatManager{
		messages: make(map[int]Message),
		nextID:   1,
		index:    invertedindex.NewInvertedIndex(),
		analyzer: analyzer,
	}
}

// PostMessage tạo tin nhắn mới và tự động lập chỉ mục thời gian thực (Real-time Indexing)
func (cm *ChatManager) PostMessage(sender, room, content string) Message {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if strings.TrimSpace(sender) == "" {
		sender = "Ẩn danh"
	}
	if strings.TrimSpace(room) == "" {
		room = "Chung"
	}

	now := time.Now()
	msg := Message{
		ID:        cm.nextID,
		Sender:    sender,
		Room:      room,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	cm.nextID++

	// 1. Lưu tin nhắn vào kho lưu trữ
	cm.messages[msg.ID] = msg

	// 2. Tự động chuyển đổi và nạp vào Inverted Index
	doc := invertedindex.Document{
		ID:      msg.ID,
		Content: msg.Content,
	}
	cm.index.AddDocument(doc, cm.analyzer)

	return msg
}

// UpdateMessage cập nhật nội dung tin nhắn và tự động re-index (Dọn posting cũ, nạp posting mới)
func (cm *ChatManager) UpdateMessage(id int, newContent string) (Message, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	msg, exists := cm.messages[id]
	if !exists {
		return Message{}, errors.New("không tìm thấy tin nhắn với ID chỉ định")
	}

	// 1. Cập nhật thông tin tin nhắn
	msg.Content = newContent
	msg.UpdatedAt = time.Now()
	cm.messages[id] = msg

	// 2. Re-index tài liệu trong Inverted Index
	doc := invertedindex.Document{
		ID:      msg.ID,
		Content: msg.Content,
	}
	cm.index.UpdateDocument(doc, cm.analyzer)

	return msg, nil
}

// DeleteMessage xóa tin nhắn và gỡ bỏ toàn bộ Postings khỏi chỉ mục
func (cm *ChatManager) DeleteMessage(id int) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.messages[id]; !exists {
		return false
	}

	// 1. Xóa khỏi kho tin nhắn
	delete(cm.messages, id)

	// 2. Xóa khỏi Inverted Index
	cm.index.DeleteDocument(id)

	return true
}

// GetMessage lấy thông tin một tin nhắn
func (cm *ChatManager) GetMessage(id int) (Message, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	msg, exists := cm.messages[id]
	return msg, exists
}

// ListMessages lấy danh sách tin nhắn theo phòng (hoặc toàn bộ nếu room = "")
func (cm *ChatManager) ListMessages(room string) []Message {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []Message
	for _, msg := range cm.messages {
		if room == "" || strings.EqualFold(msg.Room, room) {
			result = append(result, msg)
		}
	}

	// Sắp xếp theo ID tăng dần (theo dòng thời gian)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

// Search tìm kiếm tin nhắn bằng thuật toán Okapi BM25 trong Inverted Index
func (cm *ChatManager) Search(query string, roomFilter string) []ChatSearchResult {
	// 1. Thực hiện tìm kiếm trên Inverted Index
	searchResults := cm.index.Search(query, cm.analyzer)
	tokens := cm.analyzer.Analyze(query)

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var chatResults []ChatSearchResult
	for _, sr := range searchResults {
		msg, exists := cm.messages[sr.Document.ID]
		if !exists {
			continue
		}

		// Lọc theo phòng nếu có yêu cầu
		if roomFilter != "" && !strings.EqualFold(msg.Room, roomFilter) {
			continue
		}

		chatResults = append(chatResults, ChatSearchResult{
			Message: msg,
			Score:   sr.Score,
			Tokens:  tokens,
		})
	}

	return chatResults
}

// InspectMessageIndex soi các tokens và postings của tin nhắn đang lưu trong RAM
func (cm *ChatManager) InspectMessageIndex(id int) *invertedindex.DocumentIndexDetails {
	return cm.index.InspectDocument(id, cm.analyzer)
}

// GetStats trả về thống kê tổng thể
func (cm *ChatManager) GetStats() ChatStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	idxStats := cm.index.GetStats()

	roomSet := make(map[string]bool)
	for _, msg := range cm.messages {
		roomSet[msg.Room] = true
	}

	var rooms []string
	for r := range roomSet {
		rooms = append(rooms, r)
	}
	sort.Strings(rooms)

	return ChatStats{
		TotalMessages: len(cm.messages),
		TotalTerms:    idxStats.TotalTerms,
		TotalTokens:   idxStats.TotalTokens,
		AvgDocLength:  idxStats.AvgDocLength,
		ActiveRooms:   rooms,
	}
}
