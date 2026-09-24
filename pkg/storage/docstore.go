package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"vietnamese-chat-search/pkg/chat"
)

// DocStore quản lý việc lưu trữ và truy xuất tức thì các tin nhắn gốc trên đĩa cứng
// Sử dụng mô hình 2 file:
// 1. docstore.idx: Bảng mốc byte cố định (DocID * 12 bytes = ByteOffset + DataLength) -> Tra cứu O(1)
// 2. docstore.dat: Chứa chuỗi JSON/Binary nén của từng tin nhắn
type DocStore struct {
	mu         sync.RWMutex
	dir        string
	datFile    *os.File
	idxFile    *os.File
	totalDocs  uint32
	currOffset uint64
}

// OpenDocStore mở hoặc tạo mới kho lưu trữ tài liệu gốc tại thư mục chỉ định
func OpenDocStore(dir string) (*DocStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("không thể tạo thư mục lưu trữ docstore: %w", err)
	}

	datPath := filepath.Join(dir, FileDocStoreDat)
	idxPath := filepath.Join(dir, FileDocStoreIdx)

	datFile, err := os.OpenFile(datPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("không thể mở file docstore.dat: %w", err)
	}

	idxFile, err := os.OpenFile(idxPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		datFile.Close()
		return nil, fmt.Errorf("không thể mở file docstore.idx: %w", err)
	}

	// Xác định kích thước hiện tại của idxFile để tính tổng số document đã có
	idxStat, err := idxFile.Stat()
	if err != nil {
		datFile.Close()
		idxFile.Close()
		return nil, fmt.Errorf("lỗi kiểm tra kích thước file idx: %w", err)
	}

	totalDocs := uint32(idxStat.Size() / int64(DocIndexEntrySize))

	datStat, err := datFile.Stat()
	if err != nil {
		datFile.Close()
		idxFile.Close()
		return nil, fmt.Errorf("lỗi kiểm tra kích thước file dat: %w", err)
	}
	currOffset := uint64(datStat.Size())

	return &DocStore{
		dir:        dir,
		datFile:    datFile,
		idxFile:    idxFile,
		totalDocs:  totalDocs,
		currOffset: currOffset,
	}, nil
}

// AppendMessage ghi nối đuôi một tin nhắn mới vào DocStore
// Trả về DocID nội bộ tương ứng (0-indexed hoặc 1-indexed)
func (ds *DocStore) AppendMessage(msg chat.Message) (uint32, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// 1. Serialize tin nhắn thành mảng byte JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return 0, fmt.Errorf("lỗi serialize message ID %d: %w", msg.ID, err)
	}

	dataLen := uint32(len(data))
	byteOffset := ds.currOffset

	// 2. Ghi dữ liệu vào cuối file docstore.dat
	if _, err := ds.datFile.Seek(int64(byteOffset), io.SeekStart); err != nil {
		return 0, fmt.Errorf("lỗi seek datFile: %w", err)
	}
	if _, err := ds.datFile.Write(data); err != nil {
		return 0, fmt.Errorf("lỗi ghi datFile: %w", err)
	}

	// 3. Ghi mốc offset vào file docstore.idx tại vị trí DocID tương ứng
	docID := uint32(msg.ID)
	idxOffset := int64(docID) * int64(DocIndexEntrySize)

	if _, err := ds.idxFile.Seek(idxOffset, io.SeekStart); err != nil {
		return 0, fmt.Errorf("lỗi seek idxFile: %w", err)
	}

	var idxBuf bytes.Buffer
	if err := WriteUint64(&idxBuf, byteOffset); err != nil {
		return 0, err
	}
	if err := WriteUint32(&idxBuf, dataLen); err != nil {
		return 0, err
	}

	if _, err := ds.idxFile.Write(idxBuf.Bytes()); err != nil {
		return 0, fmt.Errorf("lỗi ghi idxFile: %w", err)
	}

	// Cập nhật trạng thái
	ds.currOffset += uint64(dataLen)
	if docID >= ds.totalDocs {
		ds.totalDocs = docID + 1
	}

	return docID, nil
}

// GetMessage đọc nội dung tin nhắn gốc theo DocID với độ phức tạp O(1)
func (ds *DocStore) GetMessage(docID uint32) (*chat.Message, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	idxOffset := int64(docID) * int64(DocIndexEntrySize)

	// 1. Nhảy thẳng tới vị trí mốc trong file idx
	idxStat, err := ds.idxFile.Stat()
	if err != nil {
		return nil, err
	}
	if idxOffset+int64(DocIndexEntrySize) > idxStat.Size() {
		return nil, fmt.Errorf("không tìm thấy DocID %d trong kho lưu trữ", docID)
	}

	if _, err := ds.idxFile.Seek(idxOffset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("lỗi seek idxFile: %w", err)
	}

	byteOffset, err := ReadUint64(ds.idxFile)
	if err != nil {
		return nil, fmt.Errorf("lỗi đọc ByteOffset: %w", err)
	}

	dataLen, err := ReadUint32(ds.idxFile)
	if err != nil {
		return nil, fmt.Errorf("lỗi đọc DataLength: %w", err)
	}

	if dataLen == 0 {
		return nil, fmt.Errorf("DocID %d đã bị đánh dấu trống hoặc chưa được ghi", docID)
	}

	// 2. Nhảy thẳng tới vị trí byte trong file dat và đọc chính xác dataLen bytes
	buf := make([]byte, dataLen)
	if _, err := ds.datFile.Seek(int64(byteOffset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("lỗi seek datFile: %w", err)
	}

	if _, err := io.ReadFull(ds.datFile, buf); err != nil {
		return nil, fmt.Errorf("lỗi đọc dữ liệu datFile: %w", err)
	}

	// 3. Deserialize dữ liệu
	var msg chat.Message
	if err := json.Unmarshal(buf, &msg); err != nil {
		return nil, fmt.Errorf("lỗi unmarshal message: %w", err)
	}

	return &msg, nil
}

// TotalDocs trả về tổng số tài liệu hiện có trong DocStore
func (ds *DocStore) TotalDocs() uint32 {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.totalDocs
}

// Sync đẩy toàn bộ dữ liệu từ buffer OS xuống đĩa cứng vật lý
func (ds *DocStore) Sync() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if err := ds.datFile.Sync(); err != nil {
		return err
	}
	return ds.idxFile.Sync()
}

// Close đóng các file descriptors an toàn
func (ds *DocStore) Close() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	var errDat, errIdx error
	if ds.datFile != nil {
		errDat = ds.datFile.Close()
	}
	if ds.idxFile != nil {
		errIdx = ds.idxFile.Close()
	}

	if errDat != nil {
		return errDat
	}
	return errIdx
}
