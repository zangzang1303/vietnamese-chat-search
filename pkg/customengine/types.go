package customengine

import (
	"vietnamese-chat-search/pkg/chat"
)

// SearchResult kết quả tìm kiếm trả về từ Custom Engine
type SearchResult struct {
	Message   chat.Message `json:"message"`
	Score     float64      `json:"score"`
	Highlight string       `json:"highlight,omitempty"`
}

// EngineStats thống kê trạng thái của Custom Engine
type EngineStats struct {
	TotalDocs        uint64  `json:"total_docs"`
	TotalTokens      uint64  `json:"total_tokens"`
	AvgDocLength     float64 `json:"avg_doc_length"`
	MemTableDocs     int     `json:"memtable_docs"`
	SegmentDocs      uint64  `json:"segment_docs"`
	DeletedDocsCount int     `json:"deleted_docs_count"`
	StorageDir       string  `json:"storage_dir"`
}
