package es

import (
	"context"
	"fmt"
	"strings"
)

const (
	// IndexBaseline tên chỉ mục đối soát dùng Standard Analyzer mặc định của Elasticsearch
	IndexBaseline = "chat_messages_baseline"

	// IndexVietnamese tên chỉ mục tối ưu tiếng Việt kết hợp Cốc Cốc Tokenizer & Edge N-gram
	IndexVietnamese = "chat_messages_vietnamese"
)

// BaselineMappingDefinition định nghĩa cấu trúc chỉ mục baseline theo chuẩn Standard Analyzer
const BaselineMappingDefinition = `{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "properties": {
      "id": { "type": "integer" },
      "sender": { "type": "keyword" },
      "room": { "type": "keyword" },
      "content": { 
        "type": "text", 
        "analyzer": "standard" 
      },
      "category": { "type": "keyword" },
      "created_at": { "type": "date" },
      "updated_at": { "type": "date" }
    }
  }
}`

// VietnameseMappingDefinition định nghĩa cấu trúc chỉ mục tối ưu tiếng Việt đa tầng
const VietnameseMappingDefinition = `{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "filter": {
        "edge_ngram_filter": {
          "type": "edge_ngram",
          "min_gram": 2,
          "max_gram": 15
        }
      },
      "analyzer": {
        "coccoc_whitespace_analyzer": {
          "type": "custom",
          "tokenizer": "whitespace",
          "filter": ["lowercase"]
        },
        "partial_ngram_analyzer": {
          "type": "custom",
          "tokenizer": "whitespace",
          "filter": ["lowercase", "edge_ngram_filter"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "integer" },
      "sender": { "type": "keyword" },
      "room": { "type": "keyword" },
      "content": { "type": "text" },
      "content_tokenized": {
        "type": "text",
        "analyzer": "coccoc_whitespace_analyzer",
        "similarity": "BM25"
      },
      "content_unaccented": {
        "type": "text",
        "analyzer": "coccoc_whitespace_analyzer",
        "similarity": "BM25"
      },
      "content_partial": {
        "type": "text",
        "analyzer": "partial_ngram_analyzer"
      },
      "category": { "type": "keyword" },
      "created_at": { "type": "date" },
      "updated_at": { "type": "date" }
    }
  }
}`

// EnsureIndices tự động kiểm tra sự tồn tại và tạo mới 2 index nếu chưa có
func (c *Client) EnsureIndices(ctx context.Context) error {
	indices := []struct {
		name    string
		mapping string
	}{
		{name: IndexBaseline, mapping: BaselineMappingDefinition},
		{name: IndexVietnamese, mapping: VietnameseMappingDefinition},
	}

	for _, idx := range indices {
		existsRes, err := c.Typed.Indices.Exists([]string{idx.name}, c.Typed.Indices.Exists.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("lỗi kiểm tra sự tồn tại của index %s: %w", idx.name, err)
		}
		exists := (existsRes.StatusCode == 200)
		existsRes.Body.Close()

		if !exists {
			createRes, err := c.Typed.Indices.Create(
				idx.name,
				c.Typed.Indices.Create.WithBody(strings.NewReader(idx.mapping)),
				c.Typed.Indices.Create.WithContext(ctx),
			)
			if err != nil {
				return fmt.Errorf("lỗi tạo mới index %s: %w", idx.name, err)
			}
			defer createRes.Body.Close()

			if createRes.IsError() {
				return fmt.Errorf("Elasticsearch trả về lỗi khi tạo index %s: %s", idx.name, createRes.Status())
			}
			fmt.Printf("✅ Đã tạo mới thành công Index: %s\n", idx.name)
		} else {
			fmt.Printf("ℹ️ Index %s đã tồn tại trên Elasticsearch.\n", idx.name)
		}
	}

	return nil
}

// RecreateIndices xóa và tạo lại 2 index từ đầu (dùng khi cần reset dữ liệu sạch sẽ)
func (c *Client) RecreateIndices(ctx context.Context) error {
	for _, name := range []string{IndexBaseline, IndexVietnamese} {
		c.Typed.Indices.Delete([]string{name}, c.Typed.Indices.Delete.WithContext(ctx))
	}
	return c.EnsureIndices(ctx)
}
