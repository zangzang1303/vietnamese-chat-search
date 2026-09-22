package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"vietnamese-chat-search/pkg/chat"
	"vietnamese-chat-search/pkg/invertedindex"
)

// SearchResult chứa thông tin một tin nhắn tìm thấy cùng điểm số BM25 và trường khớp
type SearchResult struct {
	Message       chat.Message `json:"message"`
	Score         float64      `json:"score"`
	MatchedFields []string     `json:"matched_fields,omitempty"`
	Highlight     string       `json:"highlight,omitempty"`
}

// ComparisonResult chứa kết quả đối soát song song giữa Baseline (Standard) và Vietnamese (Cốc Cốc)
type ComparisonResult struct {
	Query               string         `json:"query"`
	Tokens              []string       `json:"tokens"`
	BaselineResults     []SearchResult `json:"baseline_results"`
	VietnameseResults   []SearchResult `json:"vietnamese_results"`
	LatencyBaselineMs   int64          `json:"latency_baseline_ms"`
	LatencyVietnameseMs int64          `json:"latency_vietnamese_ms"`
}

// RawSearchResponse cấu trúc response nội bộ trả về từ Elasticsearch
type rawSearchResponse struct {
	Took int `json:"took"`
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Index     string                 `json:"_index"`
			ID        string                 `json:"_id"`
			Score     float64                `json:"_score"`
			Source    MessageDoc             `json:"_source"`
			Highlight map[string][]string    `json:"highlight"`
		} `json:"hits"`
	} `json:"hits"`
}

// SearchBaseline tìm kiếm trên index baseline sử dụng Standard Analyzer mặc định của Elasticsearch
func (c *Client) SearchBaseline(ctx context.Context, query string, room string) ([]SearchResult, int64, error) {
	startTime := time.Now()

	// Xây dựng câu truy vấn bool query đơn giản
	queryMap := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"content": query,
						},
					},
				},
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"content": map[string]interface{}{},
			},
		},
		"size": 30,
	}

	// Thêm bộ lọc phòng nếu có
	if room != "" {
		boolMap := queryMap["query"].(map[string]interface{})["bool"].(map[string]interface{})
		boolMap["filter"] = []interface{}{
			map[string]interface{}{
				"term": map[string]interface{}{
					"room": room,
				},
			},
		}
	}

	bodyJSON, _ := json.Marshal(queryMap)
	res, err := c.Typed.Search(
		c.Typed.Search.WithContext(ctx),
		c.Typed.Search.WithIndex(IndexBaseline),
		c.Typed.Search.WithBody(bytes.NewReader(bodyJSON)),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("lỗi thực thi SearchBaseline: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("Elasticsearch trả về lỗi: %s", res.Status())
	}

	var rawResp rawSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&rawResp); err != nil {
		return nil, 0, fmt.Errorf("lỗi parse response Baseline: %w", err)
	}

	latency := time.Since(startTime).Milliseconds()
	var results []SearchResult

	for _, hit := range rawResp.Hits.Hits {
		highlightText := hit.Source.Content
		if hl, ok := hit.Highlight["content"]; ok && len(hl) > 0 {
			highlightText = hl[0]
		}

		results = append(results, SearchResult{
			Message: chat.Message{
				ID:        hit.Source.ID,
				Sender:    hit.Source.Sender,
				Room:      hit.Source.Room,
				Content:   hit.Source.Content,
				CreatedAt: hit.Source.CreatedAt,
				UpdatedAt: hit.Source.UpdatedAt,
			},
			Score:     hit.Score,
			Highlight: highlightText,
		})
	}

	return results, latency, nil
}

// SearchVietnamese tìm kiếm trên index tối ưu tiếng Việt kết hợp Cốc Cốc Tokenizer & Relevance Boosting
func (c *Client) SearchVietnamese(ctx context.Context, query string, room string, analyzer invertedindex.Analyzer) ([]SearchResult, int64, []string, error) {
	startTime := time.Now()

	// 1. Phân tích query qua Cốc Cốc Tokenizer
	var tokens []string
	if analyzer != nil {
		tokens = analyzer.Analyze(query)
	} else {
		tokens = strings.Fields(strings.ToLower(query))
	}

	tokenizedQuery := strings.Join(tokens, " ")
	unaccentedQuery := invertedindex.RemoveDiacritics(tokenizedQuery)

	// 2. Xây dựng truy vấn bool query đa tầng với Boosting
	queryMap := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query": tokenizedQuery,
							"type":  "most_fields",
							"fields": []string{
								"content_tokenized^5.0",
								"content_unaccented^3.0",
								"content_partial^1.0",
							},
						},
					},
				},
				"should": []interface{}{
					// Match phrase để ưu tiên tuyệt đối cụm từ ghép liền kề
					map[string]interface{}{
						"match_phrase": map[string]interface{}{
							"content_tokenized": map[string]interface{}{
								"query": tokenizedQuery,
								"boost": 4.0,
							},
						},
					},
					// Match trường không dấu nếu người dùng gõ không dấu
					map[string]interface{}{
						"match": map[string]interface{}{
							"content_unaccented": map[string]interface{}{
								"query": unaccentedQuery,
								"boost": 2.0,
							},
						},
					},
				},
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"content": map[string]interface{}{},
			},
		},
		"size": 30,
	}

	// Thêm bộ lọc phòng nếu có
	if room != "" {
		boolMap := queryMap["query"].(map[string]interface{})["bool"].(map[string]interface{})
		boolMap["filter"] = []interface{}{
			map[string]interface{}{
				"term": map[string]interface{}{
					"room": room,
				},
			},
		}
	}

	bodyJSON, _ := json.Marshal(queryMap)
	res, err := c.Typed.Search(
		c.Typed.Search.WithContext(ctx),
		c.Typed.Search.WithIndex(IndexVietnamese),
		c.Typed.Search.WithBody(bytes.NewReader(bodyJSON)),
	)
	if err != nil {
		return nil, 0, tokens, fmt.Errorf("lỗi thực thi SearchVietnamese: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, tokens, fmt.Errorf("Elasticsearch trả về lỗi: %s", res.Status())
	}

	var rawResp rawSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&rawResp); err != nil {
		return nil, 0, tokens, fmt.Errorf("lỗi parse response Vietnamese: %w", err)
	}

	latency := time.Since(startTime).Milliseconds()
	var results []SearchResult

	for _, hit := range rawResp.Hits.Hits {
		highlightText := hit.Source.Content
		if hl, ok := hit.Highlight["content"]; ok && len(hl) > 0 {
			highlightText = hl[0]
		}

		results = append(results, SearchResult{
			Message: chat.Message{
				ID:        hit.Source.ID,
				Sender:    hit.Source.Sender,
				Room:      hit.Source.Room,
				Content:   hit.Source.Content,
				CreatedAt: hit.Source.CreatedAt,
				UpdatedAt: hit.Source.UpdatedAt,
			},
			Score:     hit.Score,
			Highlight: highlightText,
		})
	}

	return results, latency, tokens, nil
}

// CompareSearch thực thi đồng thời cả 2 luồng tìm kiếm và trả về đối soát chất lượng
func (c *Client) CompareSearch(ctx context.Context, query string, room string, analyzer invertedindex.Analyzer) (*ComparisonResult, error) {
	baseRes, baseLat, err := c.SearchBaseline(ctx, query, room)
	if err != nil {
		return nil, fmt.Errorf("lỗi luồng baseline: %w", err)
	}

	vnRes, vnLat, tokens, err := c.SearchVietnamese(ctx, query, room, analyzer)
	if err != nil {
		return nil, fmt.Errorf("lỗi luồng vietnamese: %w", err)
	}

	return &ComparisonResult{
		Query:               query,
		Tokens:              tokens,
		BaselineResults:     baseRes,
		VietnameseResults:   vnRes,
		LatencyBaselineMs:   baseLat,
		LatencyVietnameseMs: vnLat,
	}, nil
}
