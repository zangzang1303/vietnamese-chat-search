package storage

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
	"vietnamese-chat-search/pkg/chat"
)

const (
	WALMagicNumber uint16 = 0x5741 // "WA" - Write-Ahead
)

// WAL quản lý Write-Ahead Log để đảm bảo tính bền vững (Durability) tuyệt đối
// Mọi tin nhắn mới đều được ghi vào WAL trước khi cập nhật vào MemTable trong RAM.
type WAL struct {
	mu   sync.Mutex
	path string
	file *os.File
}

// OpenWAL mở hoặc tạo mới file Write-Ahead Log tại thư mục chỉ định
func OpenWAL(dir string) (*WAL, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("không thể tạo thư mục cho WAL: %w", err)
	}

	path := filepath.Join(dir, FileWAL)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("không thể mở file wal.log: %w", err)
	}

	return &WAL{
		path: path,
		file: file,
	}, nil
}

// Append ghi nối đuôi một tin nhắn mới vào WAL và ép buộc flush xuống đĩa vật lý
// Cấu trúc bản ghi:
// [Magic uint16 (2B)] + [PayloadLength uint32 (4B)] + [JSON Payload (NB)] + [Checksum CRC32 uint32 (4B)]
func (w *WAL) Append(msg chat.Message) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("lỗi serialize msg cho WAL: %w", err)
	}

	checksum := crc32.ChecksumIEEE(data)
	payloadLen := uint32(len(data))

	// Chuẩn bị buffer
	totalSize := 2 + 4 + len(data) + 4
	buf := make([]byte, totalSize)

	// Ghi Magic
	binary.LittleEndian.PutUint16(buf[0:2], WALMagicNumber)
	// Ghi Length
	binary.LittleEndian.PutUint32(buf[2:6], payloadLen)
	// Ghi Data
	copy(buf[6:6+len(data)], data)
	// Ghi Checksum
	binary.LittleEndian.PutUint32(buf[6+len(data):totalSize], checksum)

	if _, err := w.file.Write(buf); err != nil {
		return fmt.Errorf("lỗi ghi vào file WAL: %w", err)
	}

	// Ép buộc hệ điều hành ghi xuống đĩa cứng để chống mất mát dữ liệu
	return w.file.Sync()
}

// Replay đọc tuần tự từ đầu file WAL và gọi callback handler cho từng tin nhắn hợp lệ
// Thường dùng khi khởi động server để khôi phục lại các tin nhắn chưa kịp Flush xuống Segment
func (w *WAL) Replay(handler func(msg chat.Message) error) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Mở một file descriptor riêng để đọc tuần tự từ đầu
	readFile, err := os.Open(w.path)
	if err != nil {
		return fmt.Errorf("lỗi mở file WAL để đọc replay: %w", err)
	}
	defer readFile.Close()

	var header [6]byte // 2B magic + 4B length

	for {
		_, err := io.ReadFull(readFile, header[:])
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break // Đã đọc hết log
		}
		if err != nil {
			return fmt.Errorf("lỗi đọc header WAL: %w", err)
		}

		magic := binary.LittleEndian.Uint16(header[0:2])
		if magic != WALMagicNumber {
			return fmt.Errorf("bản ghi WAL bị hỏng (magic mismatch: 0x%X)", magic)
		}

		payloadLen := binary.LittleEndian.Uint32(header[2:6])
		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(readFile, payload); err != nil {
			return fmt.Errorf("lỗi đọc payload WAL: %w", err)
		}

		var checksumBytes [4]byte
		if _, err := io.ReadFull(readFile, checksumBytes[:]); err != nil {
			return fmt.Errorf("lỗi đọc checksum WAL: %w", err)
		}
		expectedChecksum := binary.LittleEndian.Uint32(checksumBytes[:])
		actualChecksum := crc32.ChecksumIEEE(payload)

		if expectedChecksum != actualChecksum {
			return fmt.Errorf("phát hiện bản ghi WAL bị lỗi Checksum (CRC mismatch)")
		}

		var msg chat.Message
		if err := json.Unmarshal(payload, &msg); err != nil {
			return fmt.Errorf("lỗi parse tin nhắn từ WAL: %w", err)
		}

		if err := handler(msg); err != nil {
			return fmt.Errorf("callback handler xử lý tin nhắn WAL thất bại: %w", err)
		}
	}

	return nil
}

// Truncate xóa sạch nội dung file WAL sau khi đã Flush toàn bộ MemTable xuống đĩa an toàn
func (w *WAL) Truncate() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		_ = w.file.Close()
	}

	// Mở lại file với cờ O_TRUNC để xóa sạch nội dung trên đĩa một cách an toàn trên mọi hệ điều hành
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("lỗi truncate WAL: %w", err)
	}
	w.file = file

	return nil
}

// Close đóng file WAL an toàn
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
