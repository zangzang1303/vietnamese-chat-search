package customengine

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"vietnamese-chat-search/pkg/invertedindex"
	"vietnamese-chat-search/pkg/storage"
)

const (
	BM25_K1 = 1.2
	BM25_B  = 0.75

	// Hệ số Boosting đa tầng chuẩn mực tương thích Elasticsearch
	BoostTokenized  = 5.0 // 🥇 Khớp từ ghép có dấu chuẩn
	BoostPhrase     = 4.0 // 🥈 Khớp cụm từ liên tiếp
	BoostUnaccented = 3.0 // 🥉 Khớp từ ghép không dấu
	BoostPartial    = 1.0 // 🏅 Khớp tiền tố gõ dở
)

// Search thực thi luồng tìm kiếm và chấm điểm Okapi BM25 đa tầng trên Custom Engine
func (e *Engine) Search(query string, room string) ([]SearchResult, int64, error) {
	res, lat, _, err := e.SearchWithAnalyzer(query, room, e.analyzer)
	return res, lat, err
}

// SearchWithAnalyzer thực thi luồng tìm kiếm với Analyzer chỉ định (Cốc Cốc hoặc Standard)
func (e *Engine) SearchWithAnalyzer(query string, room string, an invertedindex.Analyzer) ([]SearchResult, int64, []string, error) {
	startTime := time.Now()

	e.mu.RLock()
	defer e.mu.RUnlock()

	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return []SearchResult{}, 0, nil, nil
	}

	targetAnalyzer := an
	if targetAnalyzer == nil {
		targetAnalyzer = e.analyzer
	}

	// 1. Phân tích câu query
	tokens := targetAnalyzer.Analyze(trimmedQuery)
	unaccentedTokens := strings.Fields(invertedindex.RemoveDiacritics(strings.Join(tokens, " ")))

	// Lấy thống kê toàn cục
	totalDocs := float64(e.totalDocs)
	if totalDocs == 0 {
		totalDocs = 1
	}
	avgDocLen := e.Stats().AvgDocLength
	if avgDocLen == 0 {
		avgDocLen = 10.0
	}

	// 2. Thu thập ứng viên và chấm điểm từng tầng
	// candidateScores: DocID -> Điểm tích lũy
	candidateScores := make(map[uint32]float64)
	docMatches := make(map[uint32]map[string]int) // DocID -> Term -> TF

	// Helper thu thập postings của 1 term từ cả Segment (đĩa) và MemTable (RAM)
	collectPostings := func(term string) (postings []storage.DiskPosting, docFreq int) {
		pMap := make(map[uint32]*storage.DiskPosting)

		// A. Đọc từ Segment đĩa
		if e.segment != nil {
			if entry, ok := e.segment.GetTermEntry(term); ok {
				if diskPosts, err := e.segment.ReadPostings(entry); err == nil {
					for _, dp := range diskPosts {
						pCopy := dp
						pMap[dp.DocID] = &pCopy
					}
				}
			}
		}

		// B. Đọc từ MemTable RAM
		if memPosts, ok := e.memDictionary[term]; ok {
			for _, mp := range memPosts {
				if existing, exists := pMap[mp.DocID]; exists {
					existing.TermFrequency += mp.TermFrequency
					existing.Positions = append(existing.Positions, mp.Positions...)
				} else {
					pMap[mp.DocID] = &storage.DiskPosting{
						DocID:         mp.DocID,
						TermFrequency: mp.TermFrequency,
						Positions:     mp.Positions,
					}
				}
			}
		}

		for _, p := range pMap {
			postings = append(postings, *p)
		}
		return postings, len(postings)
	}

	// Helper tính điểm BM25
	scoreBM25 := func(tf uint32, df int, docLen int) float64 {
		// IDF chuẩn Lucene BM25: ln(1 + (N - n + 0.5) / (n + 0.5))
		numerator := (totalDocs - float64(df) + 0.5)
		denominator := (float64(df) + 0.5)
		idf := math.Log(1.0 + (numerator / denominator))
		if idf < 0 {
			idf = 0.01
		}

		lenNorm := 1.0 - BM25_B + BM25_B*(float64(docLen)/avgDocLen)
		tfComponent := (float64(tf) * (BM25_K1 + 1.0)) / (float64(tf) + BM25_K1*lenNorm)
		return idf * tfComponent
	}

	// Helper lấy độ dài tài liệu
	getDocLength := func(docID uint32) int {
		if l, ok := e.memDocLengths[docID]; ok {
			return l
		}
		return int(avgDocLen)
	}

	// =========================================================================
	// TẦNG 1: Exact Tokenized Match (Boost 5.0)
	// =========================================================================
	for _, t := range tokens {
		postings, df := collectPostings(t)
		if df == 0 {
			continue
		}
		for _, p := range postings {
			if e.tombstones[p.DocID] {
				continue // Đã bị xóa
			}
			docLen := getDocLength(p.DocID)
			s := scoreBM25(p.TermFrequency, df, docLen) * BoostTokenized
			candidateScores[p.DocID] += s

			if docMatches[p.DocID] == nil {
				docMatches[p.DocID] = make(map[string]int)
			}
			docMatches[p.DocID][t] += int(p.TermFrequency)
		}
	}

	// =========================================================================
	// TẦNG 2: Unaccented Match (Boost 3.0)
	// =========================================================================
	for _, ut := range unaccentedTokens {
		postings, df := collectPostings(ut)
		if df == 0 {
			continue
		}
		for _, p := range postings {
			if e.tombstones[p.DocID] {
				continue
			}
			docLen := getDocLength(p.DocID)
			s := scoreBM25(p.TermFrequency, df, docLen) * BoostUnaccented
			candidateScores[p.DocID] += s

			if docMatches[p.DocID] == nil {
				docMatches[p.DocID] = make(map[string]int)
			}
			docMatches[p.DocID][ut] += int(p.TermFrequency)
		}
	}

	// =========================================================================
	// TẦNG 3: Partial Edge N-gram Match (Boost 1.0)
	// =========================================================================
	for _, t := range tokens {
		ngrams := invertedindex.GenerateEdgeNgrams(t, 2, 15)
		for _, ng := range ngrams {
			if ng == t {
				continue
			}
			postings, df := collectPostings(ng)
			if df == 0 {
				continue
			}
			for _, p := range postings {
				if e.tombstones[p.DocID] {
					continue
				}
				docLen := getDocLength(p.DocID)
				s := scoreBM25(p.TermFrequency, df, docLen) * BoostPartial
				candidateScores[p.DocID] += s
			}
		}
	}

	// =========================================================================
	// TẦNG 4: Query Phrase & Edge N-gram Match (cho từ gõ dở "cà p", "ca phe", "học s")
	// =========================================================================
	rawLower := strings.ToLower(trimmedQuery)
	rawUnderscore := strings.ReplaceAll(rawLower, " ", "_")
	rawUnaccented := invertedindex.RemoveDiacritics(rawLower)
	rawUnaccentedUnderscore := strings.ReplaceAll(rawUnaccented, " ", "_")

	rawVariants := []struct {
		term  string
		boost float64
	}{
		{rawLower, BoostPhrase},
		{rawUnderscore, BoostPhrase},
		{rawUnaccented, BoostUnaccented},
		{rawUnaccentedUnderscore, BoostUnaccented},
	}

	for _, v := range rawVariants {
		if v.term == "" {
			continue
		}
		postings, df := collectPostings(v.term)
		if df == 0 {
			continue
		}
		for _, p := range postings {
			if e.tombstones[p.DocID] {
				continue
			}
			docLen := getDocLength(p.DocID)
			s := scoreBM25(p.TermFrequency, df, docLen) * v.boost
			candidateScores[p.DocID] += s

			if docMatches[p.DocID] == nil {
				docMatches[p.DocID] = make(map[string]int)
			}
			docMatches[p.DocID][v.term] += int(p.TermFrequency)
		}
	}

	// 3. Truy xuất nội dung tin nhắn gốc từ DocStore O(1) và lọc phòng
	var results []SearchResult
	for docID, score := range candidateScores {
		if score <= 0 {
			continue
		}

		msg, err := e.docStore.GetMessage(docID)
		if err != nil {
			continue // Không đọc được hoặc doc rỗng
		}

		// Lọc theo phòng chat nếu người dùng yêu cầu
		if room != "" && msg.Room != room {
			continue
		}

		// Tạo highlight đơn giản cho từ khóa
		highlight := highlightText(msg.Content, tokens)

		results = append(results, SearchResult{
			Message:   *msg,
			Score:     math.Round(score*100) / 100,
			Highlight: highlight,
		})
	}

	// 4. Sắp xếp kết quả giảm dần theo điểm số BM25
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Giới hạn Top 30 kết quả
	if len(results) > 30 {
		results = results[:30]
	}

	latencyMs := time.Since(startTime).Milliseconds()
	return results, latencyMs, tokens, nil
}

// highlightText bọc thẻ <em> xung quanh các từ khóa tìm thấy
func highlightText(content string, tokens []string) string {
	res := content
	for _, t := range tokens {
		// Bỏ dấu gạch dưới để highlight từ gốc trong câu
		rawWord := strings.ReplaceAll(t, "_", " ")
		if len(rawWord) > 1 {
			lowerRes := strings.ToLower(res)
			lowerWord := strings.ToLower(rawWord)
			idx := strings.Index(lowerRes, lowerWord)
			if idx != -1 {
				match := res[idx : idx+len(rawWord)]
				res = strings.ReplaceAll(res, match, fmt.Sprintf("<em>%s</em>", match))
			}
		}
	}
	return res
}
