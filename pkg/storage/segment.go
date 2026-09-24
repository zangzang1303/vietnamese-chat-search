package storage

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ============================================================================
// GHI FILE SEGMENT XUỐNG ĐĨA (SEGMENT SERIALIZATION)
// ============================================================================

// MemoryPosting đại diện cho posting trong bộ nhớ RAM trước khi đẩy xuống đĩa
type MemoryPosting struct {
	DocID         uint32
	TermFrequency uint32
	Positions     []uint16
}

// MemoryInvertedData chứa toàn bộ dữ liệu cần Flush xuống Segment
type MemoryInvertedData struct {
	Dictionary   map[string][]MemoryPosting // Term -> Postings
	TotalDocs    uint64
	TotalTokens  uint64
	AvgDocLength float64
	LastDocID    uint32
}

// WriteSegment ghi toàn bộ chỉ mục trong RAM xuống các file Segment nhị phân trên đĩa
func WriteSegment(dir string, data *MemoryInvertedData) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục segment: %w", err)
	}

	termsPath := filepath.Join(dir, FileTermsDict)
	postingsPath := filepath.Join(dir, FilePostingsBin)
	metaPath := filepath.Join(dir, FileSegmentMeta)

	termsFile, err := os.OpenFile(termsPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("lỗi mở file terms.dict: %w", err)
	}
	defer termsFile.Close()

	postingsFile, err := os.OpenFile(postingsPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("lỗi mở file postings.bin: %w", err)
	}
	defer postingsFile.Close()

	// 1. Sắp xếp các Term theo thứ tự từ điển A-Z để tối ưu hóa tìm kiếm nhị phân
	sortedTerms := make([]string, 0, len(data.Dictionary))
	for term := range data.Dictionary {
		sortedTerms = append(sortedTerms, term)
	}
	sort.Strings(sortedTerms)

	var postingCurrOffset uint64 = 0

	// 2. Ghi song song terms.dict và postings.bin
	for _, term := range sortedTerms {
		postings := data.Dictionary[term]
		docFreq := uint32(len(postings))
		startOffset := postingCurrOffset

		// A. Ghi Block Postings vào file postings.bin
		var postBuf bytes.Buffer
		// Ghi số lượng posting (uint32)
		if err := WriteUint32(&postBuf, docFreq); err != nil {
			return err
		}

		for _, p := range postings {
			// DocID (uint32)
			if err := WriteUint32(&postBuf, p.DocID); err != nil {
				return err
			}
			// TermFrequency (uint32)
			if err := WriteUint32(&postBuf, p.TermFrequency); err != nil {
				return err
			}
			// Positions Count (uint16)
			posCount := uint16(len(p.Positions))
			if err := WriteUint16(&postBuf, posCount); err != nil {
				return err
			}
			// Positions array
			for _, pos := range p.Positions {
				if err := WriteUint16(&postBuf, pos); err != nil {
					return err
				}
			}
		}

		postBlockBytes := postBuf.Bytes()
		blockLen := uint32(len(postBlockBytes))

		if _, err := postingsFile.Write(postBlockBytes); err != nil {
			return fmt.Errorf("lỗi ghi block posting cho term '%s': %w", term, err)
		}
		postingCurrOffset += uint64(blockLen)

		// B. Ghi Term Entry vào file terms.dict
		// [Term string] + [DocFreq uint32] + [PostingOffset uint64] + [PostingLength uint32]
		if err := WriteString(termsFile, term); err != nil {
			return err
		}
		if err := WriteUint32(termsFile, docFreq); err != nil {
			return err
		}
		if err := WriteUint64(termsFile, startOffset); err != nil {
			return err
		}
		if err := WriteUint32(termsFile, blockLen); err != nil {
			return err
		}
	}

	// 3. Ghi Header File: segments.meta
	metaFile, err := os.OpenFile(metaPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("lỗi mở file segments.meta: %w", err)
	}
	defer metaFile.Close()

	if err := WriteUint32(metaFile, MagicNumber); err != nil {
		return err
	}
	if err := WriteUint16(metaFile, FormatVersion); err != nil {
		return err
	}
	if err := WriteUint64(metaFile, data.TotalDocs); err != nil {
		return err
	}
	if err := WriteUint64(metaFile, data.TotalTokens); err != nil {
		return err
	}
	if err := WriteFloat64(metaFile, data.AvgDocLength); err != nil {
		return err
	}
	if err := WriteUint32(metaFile, data.LastDocID); err != nil {
		return err
	}
	if err := WriteUint64(metaFile, uint64(time.Now().UnixMilli())); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// ĐỌC FILE SEGMENT TỪ ĐĨA (SEGMENT DESERIALIZATION & SEARCH)
// ============================================================================

// SegmentReader chịu trách nhiệm đọc và truy vấn trên các file Segment đĩa cứng
type SegmentReader struct {
	dir          string
	meta         SegmentMeta
	dictionary   map[string]TermDictEntry // Term -> Thông tin vị trí trong postings.bin
	sortedTerms  []string                 // Danh sách các term đã sắp xếp để Binary Search
	postingsFile *os.File
}

// OpenSegment mở một phân đoạn Segment từ thư mục chỉ định
func OpenSegment(dir string) (*SegmentReader, error) {
	metaPath := filepath.Join(dir, FileSegmentMeta)
	termsPath := filepath.Join(dir, FileTermsDict)
	postingsPath := filepath.Join(dir, FilePostingsBin)

	// 1. Đọc và xác thực Header: segments.meta
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return nil, fmt.Errorf("không thể mở segments.meta: %w", err)
	}
	defer metaFile.Close()

	magic, err := ReadUint32(metaFile)
	if err != nil || magic != MagicNumber {
		return nil, fmt.Errorf("file segment bị hỏng hoặc sai định dạng Magic Number (0x%X)", magic)
	}

	version, err := ReadUint16(metaFile)
	if err != nil || version != FormatVersion {
		return nil, fmt.Errorf("phiên bản Segment không hỗ trợ (v%d)", version)
	}

	totalDocs, _ := ReadUint64(metaFile)
	totalTokens, _ := ReadUint64(metaFile)
	avgDocLen, _ := ReadFloat64(metaFile)
	lastDocID, _ := ReadUint32(metaFile)
	createdAtMs, _ := ReadUint64(metaFile)

	meta := SegmentMeta{
		MagicNumber:  magic,
		Version:      version,
		TotalDocs:    totalDocs,
		TotalTokens:  totalTokens,
		AvgDocLength: avgDocLen,
		LastDocID:    lastDocID,
		CreatedAt:    time.UnixMilli(int64(createdAtMs)),
	}

	// 2. Nạp toàn bộ terms.dict vào bộ nhớ RAM
	termsFile, err := os.Open(termsPath)
	if err != nil {
		return nil, fmt.Errorf("không thể mở terms.dict: %w", err)
	}
	defer termsFile.Close()

	dict := make(map[string]TermDictEntry)
	var sortedTerms []string

	for {
		term, err := ReadString(termsFile)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("lỗi đọc term từ terms.dict: %w", err)
		}

		docFreq, err := ReadUint32(termsFile)
		if err != nil {
			return nil, err
		}
		postingOffset, err := ReadUint64(termsFile)
		if err != nil {
			return nil, err
		}
		postingLen, err := ReadUint32(termsFile)
		if err != nil {
			return nil, err
		}

		entry := TermDictEntry{
			Term:          term,
			DocFrequency:  docFreq,
			PostingOffset: postingOffset,
			PostingLength: postingLen,
		}

		dict[term] = entry
		sortedTerms = append(sortedTerms, term)
	}

	// 3. Mở file postings.bin sẵn sàng cho các lệnh đọc ngẫu nhiên
	postingsFile, err := os.Open(postingsPath)
	if err != nil {
		return nil, fmt.Errorf("không thể mở postings.bin: %w", err)
	}

	return &SegmentReader{
		dir:          dir,
		meta:         meta,
		dictionary:   dict,
		sortedTerms:  sortedTerms,
		postingsFile: postingsFile,
	}, nil
}

// Meta trả về thông tin thống kê của Segment
func (sr *SegmentReader) Meta() SegmentMeta {
	return sr.meta
}

// HasTerm kiểm tra xem từ khóa có tồn tại trong từ điển không
func (sr *SegmentReader) HasTerm(term string) bool {
	_, ok := sr.dictionary[term]
	return ok
}

// GetTermEntry lấy thông tin offset của từ khóa
func (sr *SegmentReader) GetTermEntry(term string) (TermDictEntry, bool) {
	entry, ok := sr.dictionary[term]
	return entry, ok
}

// GetTermEntryMap trả về toàn bộ map từ điển TermDictEntry
func (sr *SegmentReader) GetTermEntryMap() map[string]TermDictEntry {
	return sr.dictionary
}

// ReadPostings đọc danh sách Posting của 1 từ trực tiếp từ file postings.bin trong thời gian cực nhanh
func (sr *SegmentReader) ReadPostings(entry TermDictEntry) ([]DiskPosting, error) {
	if _, err := sr.postingsFile.Seek(int64(entry.PostingOffset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("lỗi seek postings.bin tại offset %d: %w", entry.PostingOffset, err)
	}

	// Đọc toàn bộ block vào RAM để parse nhanh
	blockBuf := make([]byte, entry.PostingLength)
	if _, err := io.ReadFull(sr.postingsFile, blockBuf); err != nil {
		return nil, fmt.Errorf("lỗi đọc block byte posting: %w", err)
	}

	r := bytes.NewReader(blockBuf)

	postCount, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}

	postings := make([]DiskPosting, postCount)
	for i := uint32(0); i < postCount; i++ {
		docID, err := ReadUint32(r)
		if err != nil {
			return nil, err
		}
		tf, err := ReadUint32(r)
		if err != nil {
			return nil, err
		}
		posCount, err := ReadUint16(r)
		if err != nil {
			return nil, err
		}

		positions := make([]uint16, posCount)
		for j := uint16(0); j < posCount; j++ {
			pos, err := ReadUint16(r)
			if err != nil {
				return nil, err
			}
			positions[j] = pos
		}

		postings[i] = DiskPosting{
			DocID:         docID,
			TermFrequency: tf,
			Positions:     positions,
		}
	}

	return postings, nil
}

// Close đóng file postings
func (sr *SegmentReader) Close() error {
	if sr.postingsFile != nil {
		return sr.postingsFile.Close()
	}
	return nil
}

// ============================================================================
// QUẢN LÝ FILE TOMBSTONE (XÓA BẢN GHI)
// ============================================================================

// AppendTombstone ghi nhận một DocID đã bị xóa vào file tombstone.del
func AppendTombstone(dir string, docID uint32) error {
	path := filepath.Join(dir, FileTombstone)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("không thể mở tombstone.del: %w", err)
	}
	defer file.Close()

	if err := WriteUint32(file, docID); err != nil {
		return err
	}
	if err := WriteUint64(file, uint64(time.Now().UnixMilli())); err != nil {
		return err
	}

	return file.Sync()
}

// LoadTombstones nạp danh sách các DocID đã bị xóa vào map để tra cứu O(1)
func LoadTombstones(dir string) (map[uint32]bool, error) {
	path := filepath.Join(dir, FileTombstone)
	deletedMap := make(map[uint32]bool)

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return deletedMap, nil // Chưa có file tombstone nào -> chưa có ai bị xóa
	}
	if err != nil {
		return nil, fmt.Errorf("không thể đọc tombstone.del: %w", err)
	}
	defer file.Close()

	for {
		docID, err := ReadUint32(file)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// Đọc timestamp xóa (8B)
		_, err = ReadUint64(file)
		if err != nil {
			return nil, err
		}

		deletedMap[docID] = true
	}

	return deletedMap, nil
}
