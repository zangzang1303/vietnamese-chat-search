package storage

import "time"

// ============================================================================
// HẰNG SỐ ĐỊNH DANH HỆ THỐNG LƯU TRỮ (FILE FORMAT CONSTANTS)
// ============================================================================

const (
	// MagicNumber "VSFS" (Vietnamese Search File System)
	// Byte biểu diễn: 0x56 0x53 0x46 0x53
	MagicNumber uint32 = 0x56534653

	// FormatVersion phiên bản định dạng hiện tại
	FormatVersion uint16 = 1

	// File names chuẩn hóa trong thư mục lưu trữ
	FileSegmentMeta = "segments.meta"
	FileTermsDict   = "terms.dict"
	FilePostingsBin = "postings.bin"
	FileDocStoreDat = "docstore.dat"
	FileDocStoreIdx = "docstore.idx"
	FileTombstone   = "tombstone.del"
	FileWAL         = "wal.log"

	// Kích thước cố định của một bản ghi trong docstore.idx (Offset 8B + Length 4B)
	DocIndexEntrySize = 12
)

// SegmentMeta lưu trữ các thông số thống kê toàn cục phục vụ thuật toán Okapi BM25
type SegmentMeta struct {
	MagicNumber  uint32    `json:"magic_number"`   // Phải luôn bằng 0x56534653
	Version      uint16    `json:"version"`        // Phiên bản file format
	TotalDocs    uint64    `json:"total_docs"`     // Tổng số tài liệu hợp lệ (N trong BM25)
	TotalTokens  uint64    `json:"total_tokens"`   // Tổng số lượng từ trong toàn bộ hệ thống
	AvgDocLength float64   `json:"avg_doc_length"` // Độ dài trung bình của tin nhắn (avgdl)
	LastDocID    uint32    `json:"last_doc_id"`    // DocID lớn nhất đã cấp phát
	CreatedAt    time.Time `json:"created_at"`     // Thời điểm tạo Segment
}

// TermDictEntry đại diện cho một mục từ khóa trong terms.dict
type TermDictEntry struct {
	Term          string `json:"term"`           // Từ khóa (ví dụ: "cà_phê", "học_sinh")
	DocFrequency  uint32 `json:"doc_frequency"`  // Số lượng tài liệu chứa từ này (DF)
	PostingOffset uint64 `json:"posting_offset"` // Vị trí byte bắt đầu trong file postings.bin
	PostingLength uint32 `json:"posting_length"` // Chiều dài vùng byte của danh sách posting
}

// DiskPosting đại diện cho 1 tài liệu chứa từ khóa trong file postings.bin
type DiskPosting struct {
	DocID         uint32   `json:"doc_id"`         // Mã tin nhắn
	TermFrequency uint32   `json:"term_frequency"` // Tần suất xuất hiện (TF)
	Positions     []uint16 `json:"positions"`      // Vị trí các từ trong câu (phục vụ Phrase Search)
}

// DocIndexEntry đại diện cho một bản ghi offset trong file docstore.idx
type DocIndexEntry struct {
	ByteOffset uint64 // Vị trí byte bắt đầu trong file docstore.dat
	DataLength uint32 // Chiều dài byte của tin nhắn đã serialize
}

// TombstoneRecord đại diện cho một bản ghi đánh dấu xóa tin nhắn trong tombstone.del
type TombstoneRecord struct {
	DocID     uint32    `json:"doc_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
