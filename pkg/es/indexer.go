package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"vietnamese-chat-search/pkg/chat"
	"vietnamese-chat-search/pkg/invertedindex"
)

// MessageDoc đại diện cho cấu trúc tài liệu được lưu trữ trong Elasticsearch
type MessageDoc struct {
	ID                int       `json:"id"`
	Sender            string    `json:"sender"`
	Room              string    `json:"room"`
	Content           string    `json:"content"`
	ContentTokenized  string    `json:"content_tokenized,omitempty"`
	ContentUnaccented string    `json:"content_unaccented,omitempty"`
	ContentPartial    string    `json:"content_partial,omitempty"`
	Category          string    `json:"category,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// PrepareDocs chuẩn bị đồng thời 2 bản ghi: cho Baseline Index và cho Vietnamese Multi-field Index
func PrepareDocs(msg chat.Message, analyzer invertedindex.Analyzer) (MessageDoc, MessageDoc) {
	// 1. Bản ghi cho Baseline Index (chỉ chứa text thô, để Standard Analyzer của ES tự tách từ)
	baseDoc := MessageDoc{
		ID:        msg.ID,
		Sender:    msg.Sender,
		Room:      msg.Room,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
	}

	// 2. Bản ghi cho Vietnamese Index: Tách từ ghép qua Cốc Cốc Tokenizer và chuẩn hóa không dấu
	var tokenizedStr, unaccentedStr string
	if analyzer != nil {
		tokens := analyzer.Analyze(msg.Content)
		tokenizedStr = strings.Join(tokens, " ")
		unaccentedStr = invertedindex.RemoveDiacritics(tokenizedStr)
	} else {
		tokenizedStr = msg.Content
		unaccentedStr = invertedindex.RemoveDiacritics(msg.Content)
	}

	spaceStr := strings.ReplaceAll(tokenizedStr, "_", " ")
	spaceUnaccentedStr := strings.ReplaceAll(unaccentedStr, "_", " ")
	partialStr := strings.Join([]string{tokenizedStr, spaceStr, unaccentedStr, spaceUnaccentedStr}, " ")

	vnDoc := MessageDoc{
		ID:                msg.ID,
		Sender:            msg.Sender,
		Room:              msg.Room,
		Content:           msg.Content,
		ContentTokenized:  tokenizedStr,
		ContentUnaccented: unaccentedStr,
		ContentPartial:    partialStr, // Chứa cả dạng gạch dưới, khoảng trắng và không dấu để Edge N-gram sinh đầy đủ
		CreatedAt:         msg.CreatedAt,
		UpdatedAt:         msg.UpdatedAt,
	}

	return baseDoc, vnDoc
}

// BulkIndexMessages nạp hàng loạt danh sách tin nhắn vào cả 2 index sử dụng Bulk API (NDJSON)
func (c *Client) BulkIndexMessages(ctx context.Context, messages []chat.Message, analyzer invertedindex.Analyzer) error {
	if len(messages) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, msg := range messages {
		baseDoc, vnDoc := PrepareDocs(msg, analyzer)
		docID := strconv.Itoa(msg.ID)

		// 1. Thao tác nạp vào Baseline Index
		metaBaseline := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%s"}}`+"\n", IndexBaseline, docID)
		docBaseJSON, err := json.Marshal(baseDoc)
		if err != nil {
			return fmt.Errorf("lỗi serialize baseDoc ID %d: %w", msg.ID, err)
		}
		buf.WriteString(metaBaseline)
		buf.Write(docBaseJSON)
		buf.WriteString("\n")

		// 2. Thao tác nạp vào Vietnamese Index
		metaVN := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%s"}}`+"\n", IndexVietnamese, docID)
		docVNJSON, err := json.Marshal(vnDoc)
		if err != nil {
			return fmt.Errorf("lỗi serialize vnDoc ID %d: %w", msg.ID, err)
		}
		buf.WriteString(metaVN)
		buf.Write(docVNJSON)
		buf.WriteString("\n")
	}

	res, err := c.Typed.Bulk(
		bytes.NewReader(buf.Bytes()),
		c.Typed.Bulk.WithContext(ctx),
		c.Typed.Bulk.WithRefresh("true"), // Refresh tức thời để tìm kiếm được ngay
	)
	if err != nil {
		return fmt.Errorf("lỗi gửi Bulk API lên Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch Bulk API trả về lỗi: %s", res.Status())
	}

	// Đọc kết quả response để kiểm tra có document nào bị lỗi cục bộ không
	var rawResp struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			Error map[string]interface{} `json:"error"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&rawResp); err != nil {
		return fmt.Errorf("lỗi đọc response từ Bulk API: %w", err)
	}

	if rawResp.Errors {
		return fmt.Errorf("phát hiện lỗi trong một số phần tử của Bulk Index request")
	}

	fmt.Printf("✅ Đã nạp thành công %d tin nhắn vào cả 2 index (Baseline & Vietnamese)!\n", len(messages))
	return nil
}

// IndexSingleMessage nạp hoặc cập nhật một tin nhắn đơn lẻ vào Elasticsearch
func (c *Client) IndexSingleMessage(ctx context.Context, msg chat.Message, analyzer invertedindex.Analyzer) error {
	baseDoc, vnDoc := PrepareDocs(msg, analyzer)
	docID := strconv.Itoa(msg.ID)

	// 1. Nạp vào Baseline
	baseJSON, _ := json.Marshal(baseDoc)
	resBase, err := c.Typed.Index(
		IndexBaseline,
		bytes.NewReader(baseJSON),
		c.Typed.Index.WithDocumentID(docID),
		c.Typed.Index.WithContext(ctx),
		c.Typed.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("lỗi index message ID %d vào %s: %w", msg.ID, IndexBaseline, err)
	}
	resBase.Body.Close()

	// 2. Nạp vào Vietnamese
	vnJSON, _ := json.Marshal(vnDoc)
	resVN, err := c.Typed.Index(
		IndexVietnamese,
		bytes.NewReader(vnJSON),
		c.Typed.Index.WithDocumentID(docID),
		c.Typed.Index.WithContext(ctx),
		c.Typed.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("lỗi index message ID %d vào %s: %w", msg.ID, IndexVietnamese, err)
	}
	resVN.Body.Close()

	return nil
}

// UpdateSingleMessage cập nhật nội dung tin nhắn và tự động phân tích lại các trường tokens
func (c *Client) UpdateSingleMessage(ctx context.Context, id int, sender, room, newContent string, analyzer invertedindex.Analyzer) error {
	msg := chat.Message{
		ID:        id,
		Sender:    sender,
		Room:      room,
		Content:   newContent,
		UpdatedAt: time.Now(),
	}
	return c.IndexSingleMessage(ctx, msg, analyzer)
}

// DeleteSingleMessage xóa vĩnh viễn tin nhắn khỏi cả 2 index trên Elasticsearch
func (c *Client) DeleteSingleMessage(ctx context.Context, id int) error {
	docID := strconv.Itoa(id)

	for _, indexName := range []string{IndexBaseline, IndexVietnamese} {
		res, err := c.Typed.Delete(
			indexName,
			docID,
			c.Typed.Delete.WithContext(ctx),
			c.Typed.Delete.WithRefresh("true"),
		)
		if err != nil {
			return fmt.Errorf("lỗi xóa document ID %d khỏi %s: %w", id, indexName, err)
		}
		res.Body.Close()
	}

	return nil
}
