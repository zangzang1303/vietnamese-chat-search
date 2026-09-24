package customengine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"vietnamese-chat-search/pkg/chat"
	"vietnamese-chat-search/pkg/invertedindex"
	"vietnamese-chat-search/pkg/storage"
)

// Engine đại diện cho một Động cơ Tìm kiếm & Lưu trữ Đĩa Tự Xây Dựng (Custom Search Engine)
type Engine struct {
	mu         sync.RWMutex
	dir        string
	analyzer   invertedindex.Analyzer
	docStore   *storage.DocStore
	wal        *storage.WAL
	segment    *storage.SegmentReader
	tombstones map[uint32]bool

	// MemTable: Bộ nhớ đệm Inverted Index tạm thời trong RAM trước khi Flush
	memDictionary map[string][]storage.MemoryPosting
	memDocLengths map[uint32]int
	totalTokens   uint64
	totalDocs     uint64
	lastDocID     uint32
}

// Open khởi tạo hoặc mở lại Custom Engine từ thư mục chỉ định
func Open(dir string, analyzer invertedindex.Analyzer) (*Engine, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("không thể tạo thư mục cho engine: %w", err)
	}

	// 1. Mở DocStore (Forward Index O(1))
	docStore, err := storage.OpenDocStore(dir)
	if err != nil {
		return nil, fmt.Errorf("khởi tạo docstore thất bại: %w", err)
	}

	// 2. Mở WAL (Write-Ahead Log)
	wal, err := storage.OpenWAL(dir)
	if err != nil {
		docStore.Close()
		return nil, fmt.Errorf("khởi tạo WAL thất bại: %w", err)
	}

	// 3. Nạp danh sách Tombstone (tin nhắn đã xóa)
	tombstones, err := storage.LoadTombstones(dir)
	if err != nil {
		docStore.Close()
		wal.Close()
		return nil, fmt.Errorf("nạp tombstone thất bại: %w", err)
	}

	// 4. Mở Segment nếu đã tồn tại trên đĩa
	var segReader *storage.SegmentReader
	metaPath := filepath.Join(dir, storage.FileSegmentMeta)
	if _, err := os.Stat(metaPath); err == nil {
		segReader, err = storage.OpenSegment(dir)
		if err != nil {
			logErr := fmt.Sprintf("cảnh báo: không thể mở segment hiện tại: %v", err)
			fmt.Println(logErr)
		}
	}

	engine := &Engine{
		dir:           dir,
		analyzer:      analyzer,
		docStore:      docStore,
		wal:           wal,
		segment:       segReader,
		tombstones:    tombstones,
		memDictionary: make(map[string][]storage.MemoryPosting),
		memDocLengths: make(map[uint32]int),
	}

	// Thiết lập các chỉ số từ Segment nếu có
	if segReader != nil {
		meta := segReader.Meta()
		engine.totalDocs = meta.TotalDocs
		engine.totalTokens = meta.TotalTokens
		engine.lastDocID = meta.LastDocID
	}

	// 5. Crash Recovery: Replay các tin nhắn từ WAL chưa kịp flush
	replayedCount := 0
	err = wal.Replay(func(msg chat.Message) error {
		engine.indexToMemTable(msg, false) // nạp vào MemTable, không ghi lại vào WAL
		replayedCount++
		return nil
	})
	if err != nil {
		fmt.Printf("⚠️ Lỗi khi replay WAL: %v\n", err)
	} else if replayedCount > 0 {
		fmt.Printf("🔄 [Crash Recovery] Đã khôi phục thành công %d tin nhắn từ WAL!\n", replayedCount)
	}

	return engine, nil
}

// IndexMessage đánh chỉ mục một tin nhắn mới: ghi WAL -> ghi DocStore -> nạp MemTable
func (e *Engine) IndexMessage(msg chat.Message) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Ghi vào WAL chống mất dữ liệu khi mất điện/crash
	if err := e.wal.Append(msg); err != nil {
		return fmt.Errorf("lỗi ghi WAL: %w", err)
	}

	// 2. Ghi vào DocStore trên đĩa để tra cứu O(1)
	if _, err := e.docStore.AppendMessage(msg); err != nil {
		return fmt.Errorf("lỗi ghi DocStore: %w", err)
	}

	// 3. Đánh chỉ mục từ vựng đa tầng vào MemTable
	e.indexToMemTable(msg, true)

	return nil
}

// indexToMemTable xử lý bóc tách ngôn ngữ và nạp vào cấu trúc chỉ mục trong RAM
func (e *Engine) indexToMemTable(msg chat.Message, updateCounters bool) {
	docID := uint32(msg.ID)

	// Phân tích từ ghép bằng Cốc Cốc Tokenizer
	tokens := e.analyzer.Analyze(msg.Content)
	docLen := len(tokens)

	e.memDocLengths[docID] = docLen
	if updateCounters {
		e.totalTokens += uint64(docLen)
		e.totalDocs++
		if docID > e.lastDocID {
			e.lastDocID = docID
		}
	}

	// Gom tần suất và vị trí cho từng term
	type termStat struct {
		tf        uint32
		positions []uint16
	}
	stats := make(map[string]*termStat)

	addStat := func(term string, pos int) {
		term = strings.TrimSpace(strings.ToLower(term))
		if len([]rune(term)) == 0 {
			return
		}
		st, ok := stats[term]
		if !ok {
			st = &termStat{}
			stats[term] = st
		}
		hasPos := false
		for _, p := range st.positions {
			if int(p) == pos {
				hasPos = true
				break
			}
		}
		if !hasPos {
			st.positions = append(st.positions, uint16(pos))
			st.tf++
		}
	}

	// A. Tầng từ ghép chuẩn có dấu (Cốc Cốc) & Edge N-grams & Space forms
	for pos, t := range tokens {
		t = strings.ToLower(t)
		addStat(t, pos)

		// Từ ghép chứa "_"
		if strings.Contains(t, "_") {
			addStat(strings.ReplaceAll(t, "_", " "), pos)
			for _, sub := range strings.Split(t, "_") {
				addStat(sub, pos)
			}
		}

		// Edge N-grams (2 đến 15 ký tự)
		ngrams := invertedindex.GenerateEdgeNgrams(t, 2, 15)
		for _, ng := range ngrams {
			addStat(ng, pos)
			if strings.Contains(ng, "_") {
				addStat(strings.ReplaceAll(ng, "_", " "), pos)
			}
		}
	}

	// B. Tầng không dấu
	tokenizedStr := strings.Join(tokens, " ")
	unaccentedStr := invertedindex.RemoveDiacritics(tokenizedStr)
	unaccentedTokens := strings.Fields(unaccentedStr)

	for pos, ut := range unaccentedTokens {
		ut = strings.ToLower(ut)
		addStat(ut, pos)

		if strings.Contains(ut, "_") {
			addStat(strings.ReplaceAll(ut, "_", " "), pos)
			for _, sub := range strings.Split(ut, "_") {
				addStat(sub, pos)
			}
		}

		ngrams := invertedindex.GenerateEdgeNgrams(ut, 2, 15)
		for _, ng := range ngrams {
			addStat(ng, pos)
			if strings.Contains(ng, "_") {
				addStat(strings.ReplaceAll(ng, "_", " "), pos)
			}
		}
	}

	// Nạp vào MemDictionary
	for term, st := range stats {
		e.memDictionary[term] = append(e.memDictionary[term], storage.MemoryPosting{
			DocID:         docID,
			TermFrequency: st.tf,
			Positions:     st.positions,
		})
	}
}

// Flush tuần tự hóa toàn bộ MemTable xuống file Segment nhị phân trên đĩa
func (e *Engine) Flush() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Gom dữ liệu từ Segment cũ (nếu có) và MemTable mới
	combinedDict := make(map[string][]storage.MemoryPosting)

	// 1. Đọc dữ liệu từ Segment cũ nếu có
	if e.segment != nil {
		for term, entry := range e.segment.GetTermEntryMap() {
			diskPostings, err := e.segment.ReadPostings(entry)
			if err == nil {
				memPosts := make([]storage.MemoryPosting, len(diskPostings))
				for i, dp := range diskPostings {
					memPosts[i] = storage.MemoryPosting{
						DocID:         dp.DocID,
						TermFrequency: dp.TermFrequency,
						Positions:     dp.Positions,
					}
				}
				combinedDict[term] = memPosts
			}
		}
	}

	// 2. Gộp MemTable vào
	for term, memPosts := range e.memDictionary {
		combinedDict[term] = append(combinedDict[term], memPosts...)
	}

	avgDocLen := 0.0
	if e.totalDocs > 0 {
		avgDocLen = float64(e.totalTokens) / float64(e.totalDocs)
	}

	segmentData := &storage.MemoryInvertedData{
		Dictionary:   combinedDict,
		TotalDocs:    e.totalDocs,
		TotalTokens:  e.totalTokens,
		AvgDocLength: avgDocLen,
		LastDocID:    e.lastDocID,
	}

	// Đóng segment reader cũ trước khi ghi đè
	if e.segment != nil {
		_ = e.segment.Close()
		e.segment = nil
	}

	// 3. Ghi file Segment mới
	if err := storage.WriteSegment(e.dir, segmentData); err != nil {
		return fmt.Errorf("ghi segment thất bại: %w", err)
	}

	// 4. Mở lại SegmentReader mới
	segReader, err := storage.OpenSegment(e.dir)
	if err != nil {
		return fmt.Errorf("mở lại segment sau khi flush thất bại: %w", err)
	}
	e.segment = segReader

	// 5. Xóa trắng MemTable và Truncate WAL
	e.memDictionary = make(map[string][]storage.MemoryPosting)
	e.memDocLengths = make(map[uint32]int)
	if err := e.wal.Truncate(); err != nil {
		fmt.Printf("⚠️ Cảnh báo không thể truncate WAL: %v\n", err)
	}

	return nil
}

// DeleteMessage đánh dấu xóa tin nhắn qua Tombstone
func (e *Engine) DeleteMessage(docID int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	uDocID := uint32(docID)
	e.tombstones[uDocID] = true

	// Ghi vào file tombstone.del trên đĩa
	return storage.AppendTombstone(e.dir, uDocID)
}

// Stats xuất các chỉ số vận hành của Engine
func (e *Engine) Stats() EngineStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	avgLen := 0.0
	if e.totalDocs > 0 {
		avgLen = float64(e.totalTokens) / float64(e.totalDocs)
	}

	segDocs := uint64(0)
	if e.segment != nil {
		segDocs = e.segment.Meta().TotalDocs
	}

	return EngineStats{
		TotalDocs:        e.totalDocs,
		TotalTokens:      e.totalTokens,
		AvgDocLength:     avgLen,
		MemTableDocs:     len(e.memDocLengths),
		SegmentDocs:      segDocs,
		DeletedDocsCount: len(e.tombstones),
		StorageDir:       e.dir,
	}
}

// GetAllMessages đọc toàn bộ tin nhắn hợp lệ từ DocStore trên đĩa
func (e *Engine) GetAllMessages() []*chat.Message {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var messages []*chat.Message
	total := e.docStore.TotalDocs()
	for id := uint32(1); id <= total; id++ {
		if e.tombstones[id] {
			continue // Đã bị xóa
		}
		msg, err := e.docStore.GetMessage(id)
		if err == nil && msg != nil && msg.ID > 0 {
			messages = append(messages, msg)
		}
	}
	return messages
}

// RebuildIndexFromDocStore đọc lại toàn bộ tin nhắn từ DocStore để nạp lại MemTable với đầy đủ Edge N-grams
func (e *Engine) RebuildIndexFromDocStore() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.memDictionary = make(map[string][]storage.MemoryPosting)
	e.memDocLengths = make(map[uint32]int)
	e.totalTokens = 0
	e.totalDocs = 0
	e.lastDocID = 0

	total := e.docStore.TotalDocs()
	for id := uint32(1); id <= total; id++ {
		if e.tombstones[id] {
			continue
		}
		msg, err := e.docStore.GetMessage(id)
		if err == nil && msg != nil && msg.ID > 0 {
			e.indexToMemTable(*msg, true)
		}
	}

	return nil
}


// Close giải phóng toàn bộ tài nguyên của Engine
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.segment != nil {
		_ = e.segment.Close()
	}
	if e.docStore != nil {
		_ = e.docStore.Close()
	}
	if e.wal != nil {
		_ = e.wal.Close()
	}
	return nil
}
