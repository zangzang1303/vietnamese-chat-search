package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"vietnamese-chat-search/pkg/tokenizer/coccoc"
)

// Config chứa cấu hình cho phiên benchmark
type Config struct {
	TargetURL   string
	IndexName   string
	DatasetFile string
	BatchSize   int
	Concurrency int
	Mode        string // "vanilla_cgo" hoặc "plugin_raw"
	DictPath    string
	DurationSec int
	MaxDocs     int
}

// Result chứa kết quả benchmark tổng hợp
type Result struct {
	Mode                string        `json:"mode"`
	TargetURL           string        `json:"target_url"`
	TotalDocs           int64         `json:"total_docs"`
	TotalBatches        int64         `json:"total_batches"`
	TotalBytes          int64         `json:"total_bytes"`
	Duration            time.Duration `json:"duration"`
	DocsPerSec          float64       `json:"docs_per_sec"`
	MBPerSec            float64       `json:"mb_per_sec"`
	LatencyP50          float64       `json:"latency_p50_ms"`
	LatencyP90          float64       `json:"latency_p90_ms"`
	LatencyP95          float64       `json:"latency_p95_ms"`
	LatencyP99          float64       `json:"latency_p99_ms"`
	LatencyMax          float64       `json:"latency_max_ms"`
	PreprocessP50       float64       `json:"preprocess_p50_ms"`
	PreprocessP95       float64       `json:"preprocess_p95_ms"`
	PreprocessP99       float64       `json:"preprocess_p99_ms"`
	EndToEndP50         float64       `json:"end_to_end_p50_ms"`
	EndToEndP95         float64       `json:"end_to_end_p95_ms"`
	EndToEndP99         float64       `json:"end_to_end_p99_ms"`
	Errors              int64         `json:"errors"`
	HTTPErrorBatches    int64         `json:"http_error_batches"`
	BulkItemErrors      int64         `json:"bulk_item_errors"`
	TokenizerErrors     int64         `json:"tokenizer_errors"`
	Concurrency         int           `json:"concurrency"`
	BatchSize           int           `json:"batch_size"`
	DurationTargetSec   int           `json:"duration_target_sec,omitempty"`
	JVMHeapUsedMB       float64       `json:"jvm_heap_used_mb"`
	JVMGCCount          int64         `json:"jvm_gc_count"`
	JVMGCTimeMs         int64         `json:"jvm_gc_time_ms"`
	DiskStoreMB         float64       `json:"disk_store_mb"`
	SegmentsCount       int64         `json:"segments_count"`
	FlushTimeMs         int64         `json:"flush_time_ms"`
	MergeTimeMs         int64         `json:"merge_time_ms"`
	WriteRejected       int64         `json:"write_rejected"`
	RequestSamplesMS    []float64     `json:"request_latency_samples_ms,omitempty"`
	PreprocessSamplesMS []float64     `json:"preprocess_samples_ms,omitempty"`
	EndToEndSamplesMS   []float64     `json:"end_to_end_samples_ms,omitempty"`
}

func percentile(sortedValues []float64, q float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(len(sortedValues))*q)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sortedValues) {
		idx = len(sortedValues) - 1
	}
	return sortedValues[idx]
}

// MessageItem định nghĩa struct document đọc từ dataset
type MessageItem struct {
	ID       int64  `json:"id"`
	Tenant   string `json:"tenant"`
	AppID    string `json:"app_id"`
	UserID   string `json:"user_id"`
	ThreadID string `json:"thread_id"`
	Hide     string `json:"hide"`
	CreateAt int64  `json:"create_at"`
	Text     string `json:"text"`
	TextVi   string `json:"text_vi,omitempty"`
	TextRaw  string `json:"text_raw,omitempty"`
	TextEdge string `json:"text_edge,omitempty"`
}

// NodeStats thu thập thông tin JVM và Thread pool từ OpenSearch
func FetchNodeStats(baseURL string) (heapMB float64, gcCount, gcTimeMs int64, bulkQueue, bulkRejected int64) {
	resp, err := http.Get(baseURL + "/_nodes/stats/jvm,thread_pool")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data struct {
		Nodes map[string]struct {
			JVM struct {
				Mem struct {
					HeapUsedInBytes int64 `json:"heap_used_in_bytes"`
				} `json:"mem"`
				GC struct {
					Collectors map[string]struct {
						CollectionCount        int64 `json:"collection_count"`
						CollectionTimeInMillis int64 `json:"collection_time_in_millis"`
					} `json:"collectors"`
				} `json:"gc"`
			} `json:"jvm"`
			ThreadPool struct {
				Write struct {
					Queue    int64 `json:"queue"`
					Rejected int64 `json:"rejected"`
				} `json:"write"`
			} `json:"thread_pool"`
		} `json:"nodes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
		for _, n := range data.Nodes {
			heapMB = float64(n.JVM.Mem.HeapUsedInBytes) / 1024.0 / 1024.0
			for _, coll := range n.JVM.GC.Collectors {
				gcCount += coll.CollectionCount
				gcTimeMs += coll.CollectionTimeInMillis
			}
			bulkQueue += n.ThreadPool.Write.Queue
			bulkRejected += n.ThreadPool.Write.Rejected
			break
		}
	}
	return
}

// FetchIndexStats thu thập thông tin Segment, Disk Store và Merge Time từ index
func FetchIndexStats(baseURL, indexName string) (storeMB float64, segmentsCount int64, flushMs, mergeMs int64) {
	resp, err := http.Get(baseURL + "/" + indexName + "/_stats/store,segments,flush,merge")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data struct {
		Indices map[string]struct {
			Total struct {
				Store struct {
					SizeInBytes int64 `json:"size_in_bytes"`
				} `json:"store"`
				Segments struct {
					Count int64 `json:"count"`
				} `json:"segments"`
				Flush struct {
					TotalTimeInMillis int64 `json:"total_time_in_millis"`
				} `json:"flush"`
				Merges struct {
					TotalTimeInMillis int64 `json:"total_time_in_millis"`
				} `json:"merges"`
			} `json:"total"`
		} `json:"indices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
		if idx, ok := data.Indices[indexName]; ok {
			storeMB = float64(idx.Total.Store.SizeInBytes) / 1024.0 / 1024.0
			segmentsCount = idx.Total.Segments.Count
			flushMs = idx.Total.Flush.TotalTimeInMillis
			mergeMs = idx.Total.Merges.TotalTimeInMillis
		}
	}
	return
}

// LoadDataset nạp toàn bộ dataset vào bộ nhớ để khi chạy benchmark không bị nghẽn Disk I/O
func LoadDataset(filePath string, maxDocs int) ([]MessageItem, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var docs []MessageItem
	scanner := bufio.NewScanner(f)
	// Tăng buffer size cho các dòng dài
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var doc MessageItem
		if err := json.Unmarshal(line, &doc); err == nil {
			docs = append(docs, doc)
			if maxDocs > 0 && len(docs) >= maxDocs {
				break
			}
		}
	}
	return docs, scanner.Err()
}

func formatNumber(n int64) string {
	in := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(in)+len(in)/3)
	for i, c := range in {
		if i > 0 && (len(in)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

// RunBenchmark thực thi phiên đo hiệu năng với số goroutine và batch size chỉ định
func RunBenchmark(cfg Config, docs []MessageItem, tok *coccoc.Tokenizer) Result {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        cfg.Concurrency * 2,
			MaxIdleConnsPerHost: cfg.Concurrency * 2,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 60 * time.Second,
	}

	totalDocsCount := len(docs)
	if totalDocsCount == 0 {
		return Result{}
	}

	// Đọc stats JVM trước khi chạy
	startHeap, startGCCount, startGCTime, _, startRejected := FetchNodeStats(cfg.TargetURL)

	var (
		totalIndexed        int64
		totalBatches        int64
		totalBytes          int64
		totalErrors         int64
		httpErrors          int64
		bulkItemErrors      int64
		tokenizerErrors     int64
		latencies           []float64
		preprocessLatencies []float64
		endToEndLatencies   []float64
		latMu               sync.Mutex
		wg                  sync.WaitGroup
		docIndex            int64
	)

	fmt.Printf("\n🚀 Bắt đầu Benchmark [%s] -> %s (Docs: %s, Batch: %d, Concurrency: %d)\n",
		cfg.Mode, cfg.TargetURL, formatNumber(int64(totalDocsCount)), cfg.BatchSize, cfg.Concurrency)

	startTime := time.Now()
	var deadline time.Time
	if cfg.DurationSec > 0 {
		deadline = startTime.Add(time.Duration(cfg.DurationSec) * time.Second)
	}

	stopProgress := make(chan struct{})
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopProgress:
				return
			case <-ticker.C:
				currentDocs := atomic.LoadInt64(&totalIndexed)
				elapsed := time.Since(startTime)
				if currentDocs > 0 && totalDocsCount > 0 {
					pct := float64(currentDocs) / float64(totalDocsCount) * 100.0
					if cfg.DurationSec > 0 {
						pct = elapsed.Seconds() / float64(cfg.DurationSec) * 100.0
					}
					if pct > 100.0 {
						pct = 100.0
					}
					docsPerSec := float64(currentDocs) / elapsed.Seconds()
					remainingDocs := int64(totalDocsCount) - currentDocs
					if remainingDocs < 0 {
						remainingDocs = 0
					}
					var etaStr string
					if cfg.DurationSec > 0 {
						etaSec := float64(cfg.DurationSec) - elapsed.Seconds()
						if etaSec < 0 {
							etaSec = 0
						}
						etaStr = fmt.Sprintf("%.1fs", etaSec)
					} else if docsPerSec > 0 {
						etaSec := float64(remainingDocs) / docsPerSec
						etaStr = fmt.Sprintf("%.1fs", etaSec)
					} else {
						etaStr = "..."
					}
					barWidth := 20
					completedWidth := int((pct / 100.0) * float64(barWidth))
					if completedWidth > barWidth {
						completedWidth = barWidth
					}
					bar := ""
					for b := 0; b < completedWidth; b++ {
						bar += "="
					}
					if completedWidth < barWidth {
						bar += ">"
						for b := completedWidth + 1; b < barWidth; b++ {
							bar += " "
						}
					}
					if cfg.DurationSec > 0 {
						fmt.Printf("\r   ⏳ [%s] %5.1f%% | %s ops | %6.0f ops/s | Chạy: %4.1fs | Còn: %-6s",
							bar, pct, formatNumber(currentDocs), docsPerSec, elapsed.Seconds(), etaStr)
					} else {
						fmt.Printf("\r   ⏳ [%s] %5.1f%% | %s/%s docs | %6.0f docs/s | Chạy: %4.1fs | Còn: %-6s",
							bar, pct, formatNumber(currentDocs), formatNumber(int64(totalDocsCount)), docsPerSec, elapsed.Seconds(), etaStr)
					}
				}
			}
		}
	}()

	for worker := 0; worker < cfg.Concurrency; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			var buf bytes.Buffer
			batchDocs := make([]MessageItem, 0, cfg.BatchSize)

			for {
				if cfg.DurationSec > 0 && time.Now().After(deadline) {
					break
				}
				absoluteIdx := atomic.AddInt64(&docIndex, int64(cfg.BatchSize)) - int64(cfg.BatchSize)
				if cfg.DurationSec == 0 && absoluteIdx >= int64(totalDocsCount) {
					break
				}
				startIdx := int(absoluteIdx % int64(totalDocsCount))
				endIdx := startIdx + cfg.BatchSize
				if endIdx > totalDocsCount {
					endIdx = totalDocsCount
				}

				batchStart := time.Now()
				batchDocs = batchDocs[:0]
				for i := startIdx; i < endIdx; i++ {
					batchDocs = append(batchDocs, docs[i])
				}

				buf.Reset()
				for _, doc := range batchDocs {
					// Nếu là chế độ Plugin Raw: Gửi text thô cho cả 4 trường để OpenSearch Plugin băm từ bằng JNI
					if cfg.Mode == "plugin_raw" {
						doc.TextVi = doc.Text
						doc.TextRaw = doc.Text
						doc.TextEdge = doc.Text
					} else if cfg.Mode == "vanilla_pretokenized" {
						// Payload đã được tiền xử lý trước trong docs ngoài timing loop
					} else if cfg.Mode == "vanilla_cgo" && tok != nil {
						// Nếu là chế độ Vanilla CGO: Go backend tiền xử lý Cốc Cốc trước
						seg, err := tok.SegmentOriginal(doc.Text)
						if err == nil {
							doc.TextVi = seg
							doc.TextRaw = seg
							doc.TextEdge = seg
						} else {
							atomic.AddInt64(&tokenizerErrors, 1)
							doc.TextVi = doc.Text
							doc.TextRaw = doc.Text
							doc.TextEdge = doc.Text
						}
					}

					// Metadata line
					meta := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%d"}}`+"\n", cfg.IndexName, doc.ID)
					buf.WriteString(meta)
					docBytes, _ := json.Marshal(doc)
					buf.Write(docBytes)
					buf.WriteString("\n")
				}

				payloadBytes := buf.Bytes()
				batchByteLen := int64(len(payloadBytes))
				preprocessDuration := time.Since(batchStart).Seconds() * 1000.0

				reqStart := time.Now()
				resp, err := client.Post(
					fmt.Sprintf("%s/%s/_bulk", cfg.TargetURL, cfg.IndexName),
					"application/x-ndjson",
					bytes.NewReader(payloadBytes),
				)
				reqDuration := time.Since(reqStart).Seconds() * 1000.0 // ms

				if err != nil {
					atomic.AddInt64(&httpErrors, 1)
					atomic.AddInt64(&totalErrors, int64(len(batchDocs)))
				} else {
					bodyBytes, readErr := io.ReadAll(resp.Body)
					resp.Body.Close()

					if readErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
						itemErrors := int64(0)
						var bulkResp struct {
							Errors bool `json:"errors"`
							Items  []map[string]struct {
								Status int             `json:"status"`
								Error  json.RawMessage `json:"error"`
							} `json:"items"`
						}
						if err := json.Unmarshal(bodyBytes, &bulkResp); err != nil {
							itemErrors = int64(len(batchDocs))
						} else if bulkResp.Errors {
							for _, item := range bulkResp.Items {
								for _, action := range item {
									if action.Status < 200 || action.Status >= 300 || len(action.Error) > 0 {
										itemErrors++
									}
								}
							}
						}

						indexedThisBatch := int64(len(batchDocs)) - itemErrors
						atomic.AddInt64(&totalIndexed, indexedThisBatch)
						atomic.AddInt64(&bulkItemErrors, itemErrors)
						atomic.AddInt64(&totalErrors, itemErrors)
						atomic.AddInt64(&totalBatches, 1)
						atomic.AddInt64(&totalBytes, batchByteLen)

						latMu.Lock()
						latencies = append(latencies, reqDuration)
						preprocessLatencies = append(preprocessLatencies, preprocessDuration)
						endToEndLatencies = append(endToEndLatencies, time.Since(batchStart).Seconds()*1000.0)
						latMu.Unlock()
					} else {
						atomic.AddInt64(&httpErrors, 1)
						atomic.AddInt64(&totalErrors, int64(len(batchDocs)))
					}
				}
			}
		}(worker)
	}

	wg.Wait()
	close(stopProgress)
	if cfg.DurationSec > 0 {
		fmt.Printf("\r   ✅ [====================] 100.0%% | %s ops hoàn tất trong %.2fs!                     \n",
			formatNumber(totalIndexed), time.Since(startTime).Seconds())
	} else {
		fmt.Printf("\r   ✅ [====================] 100.0%% | %s/%s docs hoàn tất trong %.2fs!                     \n",
			formatNumber(totalIndexed), formatNumber(int64(totalDocsCount)), time.Since(startTime).Seconds())
	}
	duration := time.Since(startTime)

	// Đọc stats JVM sau khi chạy
	endHeap, endGCCount, endGCTime, _, endRejected := FetchNodeStats(cfg.TargetURL)

	// Tính toán phân vị Latency
	sort.Float64s(latencies)
	var p50, p90, p95, p99, maxLat float64
	if len(latencies) > 0 {
		p50 = percentile(latencies, 0.50)
		p90 = percentile(latencies, 0.90)
		p95 = percentile(latencies, 0.95)
		p99 = percentile(latencies, 0.99)
		maxLat = latencies[len(latencies)-1]
	}
	sort.Float64s(preprocessLatencies)
	sort.Float64s(endToEndLatencies)

	// Trigger refresh để đảm bảo segment và stats được cập nhật đầy đủ
	http.Post(cfg.TargetURL+"/"+cfg.IndexName+"/_refresh", "application/json", nil)
	storeMB, segmentsCount, flushMs, mergeMs := FetchIndexStats(cfg.TargetURL, cfg.IndexName)

	res := Result{
		Mode:                cfg.Mode,
		TargetURL:           cfg.TargetURL,
		Concurrency:         cfg.Concurrency,
		BatchSize:           cfg.BatchSize,
		DurationTargetSec:   cfg.DurationSec,
		TotalDocs:           totalIndexed,
		TotalBatches:        totalBatches,
		TotalBytes:          totalBytes,
		Duration:            duration,
		DocsPerSec:          float64(totalIndexed) / duration.Seconds(),
		MBPerSec:            (float64(totalBytes) / 1024.0 / 1024.0) / duration.Seconds(),
		LatencyP50:          p50,
		LatencyP90:          p90,
		LatencyP95:          p95,
		LatencyP99:          p99,
		LatencyMax:          maxLat,
		PreprocessP50:       percentile(preprocessLatencies, 0.50),
		PreprocessP95:       percentile(preprocessLatencies, 0.95),
		PreprocessP99:       percentile(preprocessLatencies, 0.99),
		EndToEndP50:         percentile(endToEndLatencies, 0.50),
		EndToEndP95:         percentile(endToEndLatencies, 0.95),
		EndToEndP99:         percentile(endToEndLatencies, 0.99),
		Errors:              totalErrors,
		HTTPErrorBatches:    httpErrors,
		BulkItemErrors:      bulkItemErrors,
		TokenizerErrors:     tokenizerErrors,
		JVMHeapUsedMB:       endHeap - startHeap,
		JVMGCCount:          endGCCount - startGCCount,
		JVMGCTimeMs:         endGCTime - startGCTime,
		DiskStoreMB:         storeMB,
		SegmentsCount:       segmentsCount,
		FlushTimeMs:         flushMs,
		MergeTimeMs:         mergeMs,
		WriteRejected:       endRejected - startRejected,
		RequestSamplesMS:    latencies,
		PreprocessSamplesMS: preprocessLatencies,
		EndToEndSamplesMS:   endToEndLatencies,
	}

	fmt.Printf("📊 Kết quả [%s] (Concurrency=%d):\n", cfg.Mode, cfg.Concurrency)
	fmt.Printf("   - Docs indexed: %d docs trong %.2fs\n", res.TotalDocs, duration.Seconds())
	fmt.Printf("   - Thông lượng Throughput: %.2f docs/sec (%.2f MB/sec)\n", res.DocsPerSec, res.MBPerSec)
	fmt.Printf("   - Latency (ms): p50=%.2f | p90=%.2f | p95=%.2f | p99=%.2f | max=%.2f\n",
		res.LatencyP50, res.LatencyP90, res.LatencyP95, res.LatencyP99, res.LatencyMax)
	fmt.Printf("   - End-to-end latency (ms): p50=%.2f | p95=%.2f | p99=%.2f\n",
		res.EndToEndP50, res.EndToEndP95, res.EndToEndP99)
	fmt.Printf("   - JVM GC Delta: %d collections, %d ms pause\n", res.JVMGCCount, res.JVMGCTimeMs)
	fmt.Printf("   - Storage: %.2f MB, Segments: %d, Merge Time: %d ms\n", res.DiskStoreMB, res.SegmentsCount, res.MergeTimeMs)
	if res.Errors > 0 || res.WriteRejected > 0 {
		fmt.Printf("   - ⚠️ Lỗi docs=%d, HTTP batches=%d, bulk items=%d, tokenizer=%d, write rejected=%d\n",
			res.Errors, res.HTTPErrorBatches, res.BulkItemErrors, res.TokenizerErrors, res.WriteRejected)
	}

	return res
}

func main() {
	mode := flag.String("mode", "all", "Chế độ chạy: vanilla, plugin, all, parity")
	batchSize := flag.Int("batch", 1000, "Kích thước batch docs cho mỗi bulk request")
	concurrency := flag.Int("concurrency", 8, "Số lượng worker goroutines đồng thời")
	datasetPath := flag.String("dataset", "benchmarks/dataset_benchmark.jsonl", "Đường dẫn file dataset")
	maxDocs := flag.Int("max_docs", 100000, "Số lượng docs tối đa nạp trong test")
	durationSec := flag.Int("duration_sec", 0, "Thời lượng chạy liên tục; 0 nghĩa là chạy hết max_docs")
	dictPath := flag.String("dict", "data/dicts/coccoc", "Đường dẫn từ điển Cốc Cốc")
	vanillaURL := flag.String("vanilla_url", "http://localhost:9201", "URL OpenSearch Vanilla")
	pluginURL := flag.String("plugin_url", "http://localhost:9202", "URL OpenSearch Plugin")
	outputFile := flag.String("output", "benchmarks/benchmark_results.json", "Đường dẫn file kết quả JSON")
	flag.Parse()

	// Chỉ Model A cần Cốc Cốc Tokenizer. Không fallback âm thầm vì sẽ làm sai workload.
	var tok *coccoc.Tokenizer
	if *mode == "vanilla" || *mode == "vanilla_pretokenized" || *mode == "all" {
		var err error
		tok, err = coccoc.New(*dictPath, false)
		if err != nil {
			fmt.Printf("❌ Không khởi tạo được Cốc Cốc Tokenizer CGO: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("📂 Đang nạp dataset từ %s (tối đa %d docs)...\n", *datasetPath, *maxDocs)
	docs, err := LoadDataset(*datasetPath, *maxDocs)
	if err != nil {
		fmt.Printf("❌ Lỗi nạp dataset: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Đã nạp thành công %d docs vào RAM!\n", len(docs))

	var results []Result

	if *mode == "vanilla_pretokenized" {
		fmt.Printf("⚡ Tiền xử lý Cốc Cốc trước ngoài vùng đo cho %d docs...\n", len(docs))
		preStart := time.Now()
		for i := range docs {
			seg, err := tok.SegmentOriginal(docs[i].Text)
			if err == nil {
				docs[i].TextVi = seg
				docs[i].TextRaw = seg
				docs[i].TextEdge = seg
			} else {
				docs[i].TextVi = docs[i].Text
				docs[i].TextRaw = docs[i].Text
				docs[i].TextEdge = docs[i].Text
			}
		}
		fmt.Printf("✅ Đã tiền xử lý xong %d docs trong %.2fs!\n", len(docs), time.Since(preStart).Seconds())

		cfgVanilla := Config{
			TargetURL:   *vanillaURL,
			IndexName:   "messages",
			BatchSize:   *batchSize,
			Concurrency: *concurrency,
			Mode:        "vanilla_pretokenized",
			DurationSec: *durationSec,
		}
		res := RunBenchmark(cfgVanilla, docs, nil)
		results = append(results, res)
	}

	if *mode == "vanilla" || *mode == "all" {
		cfgVanilla := Config{
			TargetURL:   *vanillaURL,
			IndexName:   "messages",
			BatchSize:   *batchSize,
			Concurrency: *concurrency,
			Mode:        "vanilla_cgo",
			DurationSec: *durationSec,
		}
		res := RunBenchmark(cfgVanilla, docs, tok)
		results = append(results, res)
	}

	if *mode == "plugin" || *mode == "all" {
		cfgPlugin := Config{
			TargetURL:   *pluginURL,
			IndexName:   "messages",
			BatchSize:   *batchSize,
			Concurrency: *concurrency,
			Mode:        "plugin_raw",
			DurationSec: *durationSec,
		}
		res := RunBenchmark(cfgPlugin, docs, tok)
		results = append(results, res)
	}

	// Xuất kết quả ra file JSON
	outJSON, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile(*outputFile, outJSON, 0644)
	fmt.Printf("\n💾 Kết quả chi tiết đã được lưu vào %s\n", *outputFile)
}
