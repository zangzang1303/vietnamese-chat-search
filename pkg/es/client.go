package es

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// Client đóng gói elasticsearch.Client chính thức kèm các tiện ích kiểm tra trạng thái
type Client struct {
	Typed *elasticsearch.Client
	URL   string
}

// ClusterInfo chứa thông tin cơ bản về cụm Elasticsearch khi Ping thành công
type ClusterInfo struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	Version     struct {
		Number string `json:"number"`
		Lucene string `json:"lucene_version"`
	} `json:"version"`
	Tagline string `json:"tagline"`
}

// NewClient khởi tạo kết nối tới Elasticsearch với cấu hình tối ưu
func NewClient(addresses []string) (*Client, error) {
	if len(addresses) == 0 {
		addresses = []string{"http://localhost:9200"}
	}

	cfg := elasticsearch.Config{
		Addresses: addresses,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 10 * time.Second,
		},
		MaxRetries: 3,
	}

	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("không thể khởi tạo Elasticsearch Client: %w", err)
	}

	return &Client{
		Typed: esClient,
		URL:   addresses[0],
	}, nil
}

// Ping kiểm tra kết nối tới cụm Elasticsearch và trả về thông tin cụm
func (c *Client) Ping(ctx context.Context) (*ClusterInfo, error) {
	res, err := c.Typed.Info(c.Typed.Info.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối tới Elasticsearch (%s): %w", c.URL, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch trả về mã lỗi: %s", res.Status())
	}

	var info ClusterInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("lỗi giải mã JSON Cluster Info: %w", err)
	}

	return &info, nil
}

// IsAvailable kiểm tra nhanh xem Elasticsearch có đang online hay không (timeout 2s)
func (c *Client) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := c.Ping(ctx)
	return err == nil
}

// PrintClusterInfo in thông tin cụm Elasticsearch ra màn hình
func (c *Client) PrintClusterInfo() {
	info, err := c.Ping(context.Background())
	if err != nil {
		fmt.Printf("⚠️ Elasticsearch chưa sẵn sàng tại %s: %v\n", c.URL, err)
		return
	}

	fmt.Println("================================================================================")
	fmt.Printf("✅ KẾT NỐI ELASTICSEARCH THÀNH CÔNG!\n")
	fmt.Printf("   • Cluster Name   : %s\n", info.ClusterName)
	fmt.Printf("   • Node Name      : %s\n", info.Name)
	fmt.Printf("   • ES Version     : %s (Lucene: %s)\n", info.Version.Number, info.Version.Lucene)
	fmt.Printf("   • URL Endpoint   : %s\n", c.URL)
	fmt.Println("================================================================================")
}

// TrimPrefixes loại bỏ các tiền tố không cần thiết
func cleanString(s string) string {
	return strings.TrimSpace(s)
}
