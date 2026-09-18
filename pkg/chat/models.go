package chat

import "time"

// Message đại diện cho một tin nhắn chat trong hệ thống thực tế
type Message struct {
	ID        int       `json:"id"`
	Sender    string    `json:"sender"`
	Room      string    `json:"room"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatSearchResult đại diện cho tin nhắn tìm thấy kèm điểm BM25 và từ khóa khớp
type ChatSearchResult struct {
	Message Message  `json:"message"`
	Score   float64  `json:"score"`
	Tokens  []string `json:"tokens"`
}

// CreateMessageRequest dữ liệu gửi lên khi tạo tin nhắn mới
type CreateMessageRequest struct {
	Sender  string `json:"sender"`
	Room    string `json:"room"`
	Content string `json:"content"`
}

// UpdateMessageRequest dữ liệu gửi lên khi sửa tin nhắn
type UpdateMessageRequest struct {
	Content string `json:"content"`
}

// ChatStats thông kê tổng quan về hệ thống tin nhắn và chỉ mục
type ChatStats struct {
	TotalMessages int      `json:"total_messages"`
	TotalTerms    int      `json:"total_terms"`
	TotalTokens   int      `json:"total_tokens"`
	AvgDocLength  float64  `json:"avg_doc_length"`
	ActiveRooms   []string `json:"active_rooms"`
}
