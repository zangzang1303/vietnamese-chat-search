package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"vietnamese-chat-search/pkg/chat"
)

func TestDocStore_AppendAndGet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_docstore_*")
	if err != nil {
		t.Fatalf("Lỗi tạo temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ds, err := OpenDocStore(tempDir)
	if err != nil {
		t.Fatalf("Lỗi OpenDocStore: %v", err)
	}
	defer ds.Close()

	msg1 := chat.Message{
		ID:        1,
		Sender:    "Tuan",
		Room:      "general",
		Content:   "Học sinh uống cà phê tại quán",
		CreatedAt: time.Now(),
	}

	msg2 := chat.Message{
		ID:        2,
		Sender:    "Nam",
		Room:      "general",
		Content:   "Chào mừng các bạn tân sinh viên",
		CreatedAt: time.Now(),
	}

	// 1. Test Append
	docID1, err := ds.AppendMessage(msg1)
	if err != nil {
		t.Fatalf("AppendMessage msg1 thất bại: %v", err)
	}
	if docID1 != 1 {
		t.Errorf("Kỳ vọng DocID 1, nhận %d", docID1)
	}

	docID2, err := ds.AppendMessage(msg2)
	if err != nil {
		t.Fatalf("AppendMessage msg2 thất bại: %v", err)
	}
	if docID2 != 2 {
		t.Errorf("Kỳ vọng DocID 2, nhận %d", docID2)
	}

	// 2. Test GetMessage O(1)
	got1, err := ds.GetMessage(1)
	if err != nil {
		t.Fatalf("GetMessage(1) lỗi: %v", err)
	}
	if got1.Content != msg1.Content || got1.Sender != msg1.Sender {
		t.Errorf("Dữ liệu msg1 không khớp: got %+v", got1)
	}

	got2, err := ds.GetMessage(2)
	if err != nil {
		t.Fatalf("GetMessage(2) lỗi: %v", err)
	}
	if got2.Content != msg2.Content || got2.Sender != msg2.Sender {
		t.Errorf("Dữ liệu msg2 không khớp: got %+v", got2)
	}

	// 3. Test Reopen DocStore
	ds.Close()
	dsReopened, err := OpenDocStore(tempDir)
	if err != nil {
		t.Fatalf("Reopen DocStore thất bại: %v", err)
	}
	defer dsReopened.Close()

	if dsReopened.TotalDocs() < 2 {
		t.Errorf("Reopened TotalDocs kỳ vọng >= 2, got %d", dsReopened.TotalDocs())
	}

	gotReopened, err := dsReopened.GetMessage(1)
	if err != nil || gotReopened.Content != msg1.Content {
		t.Errorf("Đọc lại sau khi reopen thất bại: %v", err)
	}
}

func TestWAL_AppendAndReplay(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_wal_*")
	if err != nil {
		t.Fatalf("Lỗi tạo temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wal, err := OpenWAL(tempDir)
	if err != nil {
		t.Fatalf("OpenWAL lỗi: %v", err)
	}

	msg1 := chat.Message{ID: 10, Sender: "A", Content: "Test WAL 1"}
	msg2 := chat.Message{ID: 20, Sender: "B", Content: "Test WAL 2"}

	if err := wal.Append(msg1); err != nil {
		t.Fatalf("WAL Append msg1 lỗi: %v", err)
	}
	if err := wal.Append(msg2); err != nil {
		t.Fatalf("WAL Append msg2 lỗi: %v", err)
	}
	wal.Close()

	// Replay khôi phục
	walReopened, err := OpenWAL(tempDir)
	if err != nil {
		t.Fatalf("Reopen WAL lỗi: %v", err)
	}
	defer walReopened.Close()

	var replayed []chat.Message
	err = walReopened.Replay(func(m chat.Message) error {
		replayed = append(replayed, m)
		return nil
	})

	if err != nil {
		t.Fatalf("Replay lỗi: %v", err)
	}
	if len(replayed) != 2 {
		t.Fatalf("Kỳ vọng 2 tin nhắn replayed, got %d", len(replayed))
	}
	if replayed[0].Content != "Test WAL 1" || replayed[1].Content != "Test WAL 2" {
		t.Errorf("Nội dung replayed không khớp")
	}

	// Test Truncate
	if err := walReopened.Truncate(); err != nil {
		t.Fatalf("Truncate lỗi: %v", err)
	}

	replayedAfterTrunc := 0
	_ = walReopened.Replay(func(m chat.Message) error {
		replayedAfterTrunc++
		return nil
	})
	if replayedAfterTrunc != 0 {
		t.Errorf("Sau khi truncate kỳ vọng 0 tin nhắn, got %d", replayedAfterTrunc)
	}
}

func TestSegment_WriteAndRead(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_segment_*")
	if err != nil {
		t.Fatalf("Lỗi tạo temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	data := &MemoryInvertedData{
		Dictionary: map[string][]MemoryPosting{
			"cà_phê": {
				{DocID: 1, TermFrequency: 2, Positions: []uint16{3, 7}},
				{DocID: 5, TermFrequency: 1, Positions: []uint16{2}},
			},
			"học_sinh": {
				{DocID: 1, TermFrequency: 1, Positions: []uint16{0}},
				{DocID: 3, TermFrequency: 1, Positions: []uint16{1}},
			},
		},
		TotalDocs:    10,
		TotalTokens:  120,
		AvgDocLength: 12.0,
		LastDocID:    5,
	}

	// 1. Ghi Segment xuống đĩa
	if err := WriteSegment(tempDir, data); err != nil {
		t.Fatalf("WriteSegment lỗi: %v", err)
	}

	// Kiểm tra các file đã được tạo
	expectedFiles := []string{FileSegmentMeta, FileTermsDict, FilePostingsBin}
	for _, ef := range expectedFiles {
		p := filepath.Join(tempDir, ef)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("File %s không tồn tại trên đĩa", ef)
		}
	}

	// 2. Đọc lại Segment từ đĩa
	reader, err := OpenSegment(tempDir)
	if err != nil {
		t.Fatalf("OpenSegment lỗi: %v", err)
	}
	defer reader.Close()

	meta := reader.Meta()
	if meta.TotalDocs != 10 || meta.AvgDocLength != 12.0 {
		t.Errorf("Meta không khớp: %+v", meta)
	}

	// Kiểm tra term "cà_phê"
	entry, ok := reader.GetTermEntry("cà_phê")
	if !ok {
		t.Fatalf("Không tìm thấy term 'cà_phê' trong terms.dict")
	}
	if entry.DocFrequency != 2 {
		t.Errorf("Kỳ vọng DocFrequency=2, got %d", entry.DocFrequency)
	}

	postings, err := reader.ReadPostings(entry)
	if err != nil {
		t.Fatalf("ReadPostings lỗi: %v", err)
	}
	if len(postings) != 2 {
		t.Fatalf("Kỳ vọng 2 postings, got %d", len(postings))
	}
	if postings[0].DocID != 1 || postings[0].TermFrequency != 2 || len(postings[0].Positions) != 2 {
		t.Errorf("Posting[0] không khớp: %+v", postings[0])
	}

	// 3. Test Tombstone
	if err := AppendTombstone(tempDir, 1); err != nil {
		t.Fatalf("AppendTombstone lỗi: %v", err)
	}
	tombstones, err := LoadTombstones(tempDir)
	if err != nil {
		t.Fatalf("LoadTombstones lỗi: %v", err)
	}
	if !tombstones[1] {
		t.Errorf("Kỳ vọng DocID 1 có trong tombstones")
	}
	if tombstones[5] {
		t.Errorf("DocID 5 không được có trong tombstones")
	}
}
