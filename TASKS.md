# Danh Sách Nhiệm Vụ Triển Khai (Task Breakdown)

Dự án: **Cải thiện kết quả tìm kiếm tin nhắn tiếng Việt với Cốc Cốc Tokenizer & Elasticsearch trong Go**  
Tham chiếu đề bài gốc: [`REQUIREMENTS.md`](REQUIREMENTS.md)

---

## 📌 Bảng Tiến Độ Tổng Quan

| Giai đoạn | Nội dung chính | Trạng thái | Ước lượng |
| :--- | :--- | :---: | :---: |
| **Phase 1** | Nghiên cứu lý thuyết Search Engine, Lucene & Elasticsearch | ✅ Hoàn thành | 1-2 ngày |
| **Phase 2** | Thiết lập môi trường Docker (ES, Kibana) & Seed Data | ✅ Hoàn thành | 1 ngày |
| **Phase 3** | Xây dựng CGO Wrapper cho Cốc Cốc Tokenizer trong Go | ✅ Hoàn thành | 2-3 ngày |
| **Phase 4** | Thiết kế Elasticsearch Index Mapping & Pipeline Indexing | ✅ Hoàn thành | 1-2 ngày |
| **Phase 5** | Xây dựng Search Engine Service & Multi-match Query Builder | ✅ Hoàn thành | 1-2 ngày |
| **Phase 6** | Kiểm thử so sánh chất lượng search & Benchmark hiệu năng (MRR, NDCG) | ✅ Hoàn thành | 1-2 ngày |
| **Phase 7** | Hoàn thiện tài liệu báo cáo & tổng kết | 🔄 Đang hoàn thiện | 1 ngày |

---

## Chi Tiết Các Nhiệm Vụ (Task Checklist)

### 🔹 Phase 1: Nghiên cứu lý thuyết Search Engine & Kiến trúc
- [x] **TASK-1.1: Nghiên cứu cơ chế Inverted Index của Apache Lucene**
  - [x] Tìm hiểu cấu trúc dữ liệu: *Term Dictionary*, *Posting List*, *Term Frequencies*, *Positions*, *Doc Values*.
  - [x] Hiểu cách Lucene lưu trữ và tra cứu từ khóa trong bộ nhớ / ổ đĩa (Segment Immutability, `.del` Tombstone).
- [x] **TASK-1.2: Nghiên cứu hai luồng cốt lõi của Search Engine**
  - [x] **Luồng Indexing**: `Character Filters` $\rightarrow$ `Tokenizer` $\rightarrow$ `Token Filters` $\rightarrow$ `Inverted Index` $\rightarrow$ `Storage`.
  - [x] **Luồng Search & Ranking**: `Query Tokenizer` $\rightarrow$ `Filtering` $\rightarrow$ `Inverted Index Lookup` $\rightarrow$ `BM25 Scoring Algorithm`.
- [x] **TASK-1.3: Phân tích bài toán đặc thù của tiếng Việt**
  - [x] Phân tích vấn đề từ đơn vs từ ghép (ví dụ: *"học sinh"* vs *"sinh viên"*).
  - [x] Phân tích vấn đề tìm kiếm không dấu (*"ca phe"* tìm *"cà phê"*).
  - [x] Phân tích bài toán tìm kiếm tiền tố / một phần từ (*Partial Matching* với Edge N-gram).
  - [x] Viết tài liệu tổng hợp lý thuyết tại `docs/search_engine_fundamentals.md` và `docs/elasticsearch_definitive_guide_notes.md`.

---

### 🔹 Phase 2: Môi trường phát triển & Dữ liệu mẫu (Data Seeding)
- [x] **TASK-2.1: Khởi tạo cụm Elasticsearch & Kibana với Docker Compose**
  - [x] Tạo file `docker/docker-compose.yml` gồm Elasticsearch 8.11.0 và Kibana.
  - [x] Cấu hình single-node, tắt SSL/Security cho môi trường local để tối ưu hóa hiệu năng và kết nối HTTP.
  - [x] Thiết lập healthcheck và volume `es_data` đảm bảo ES sẵn sàng và lưu trữ bền vững trên đĩa cứng.
- [ ] **TASK-2.2: Thiết lập môi trường Build CGO & Go**
  - [ ] Cấu hình `Dockerfile` hỗ trợ build đa tầng (multi-stage build) có sẵn `gcc`, `g++`, `cmake`, `git`, `go`.
  - [ ] Cấu hình `Makefile` với các lệnh thông dụng (`make up`, `make down`, `make build`, `make seed`, `make test`).
- [x] **TASK-2.3: Xây dựng bộ dữ liệu mẫu tin nhắn chat tiếng Việt**
  - [x] Tạo file `data/sample_messages.json` chứa 131 tin nhắn chat nhóm thực tế bao phủ đầy đủ các bẫy từ ghép.

---

### 🔹 Phase 3: Tích hợp Cốc Cốc Tokenizer trong Go qua CGO
- [x] **TASK-3.1: Chuẩn bị mã nguồn C++ Cốc Cốc Tokenizer**
  - [x] Nghiên cứu mã nguồn [coccoc/coccoc-tokenizer](https://github.com/coccoc/coccoc-tokenizer) và từ điển phân đoạn từ `sys.dic`.
  - [x] Thiết kế cầu nối C-Interface bridge (`coccoc_bridge.h`, `coccoc_bridge.cpp`).
- [x] **TASK-3.2: Xây dựng Go CGO Wrapper (`pkg/tokenizer`)**
  - [x] Định nghĩa CGO header và cflags/ldflags để liên kết thư viện C++ vào Go.
  - [x] Xây dựng stub tokenizer và wrapper thread-safe với `sync.Mutex`.
  - [x] Cung cấp hàm `Tokenize(text string) []string`: phân tách từ tiếng Việt và nối từ ghép bằng dấu gạch dưới (ví dụ: `["tôi", "uống", "cà_phê"]`).
- [x] **TASK-3.3: Xây dựng Module chuẩn hóa không dấu (`pkg/tokenizer/unaccent.go`)**
  - [x] Hàm chuẩn hóa Unicode NFC / NFD.
  - [x] Hàm chuyển đổi chuỗi tiếng Việt có dấu sang không dấu (`"uống cà_phê"` $\rightarrow$ `"uong ca_phe"`).
- [x] **TASK-3.4: Kiểm thử Unit Test cho Tokenizer**
  - [x] Viết test cases kiểm tra độ chính xác của tokenizer với các câu tiếng Việt phức tạp (`pkg/tokenizer/coccoc/stub_test.go`).

---

### 🔹 Phase 4: Thiết kế Elasticsearch Mapping & Xây dựng Indexer
- [x] **TASK-4.1: Thiết kế Elasticsearch Mapping (Index Schema)**
  - [x] Tạo index `chat_messages_baseline` (dùng `standard` analyzer mặc định của ES).
  - [x] Tạo index `chat_messages_vietnamese` với multi-field mapping:
    - `content`: Văn bản gốc dùng để hiển thị kết quả.
    - `content_tokenized`: Lưu các token đã qua Cốc Cốc tokenizer (`coccoc_whitespace_analyzer`).
    - `content_unaccented`: Lưu token không dấu (`asciifolding`).
    - `content_partial`: Lưu token kết hợp custom token filter `edge_ngram` (min_gram: 2, max_gram: 15) phục vụ tìm kiếm dở từ / autocomplete.
- [x] **TASK-4.2: Xây dựng công cụ nạp dữ liệu (Go Indexer CLI)**
  - [x] Tạo CLI `cmd/indexer` kết nối Elasticsearch qua thư viện Go client (`elastic/go-elasticsearch/v8`).
  - [x] Đọc dữ liệu từ `data/sample_messages.json`.
  - [x] Xử lý tokenize nội dung qua Cốc Cốc Tokenizer và gửi Bulk Index vào cả 2 chỉ mục (baseline & vietnamese).
  - [x] Đo thời gian hoàn thành quá trình index (233ms cho toàn bộ 131 tin nhắn).

---

### 🔹 Phase 5: Xây dựng Search Service & Query Builder
- [x] **TASK-5.1: Xây dựng CLI / HTTP Search Service (`cmd/searcher`, `cmd/chat_server`)**
  - [x] Tạo REST endpoint (`GET /api/search?q=...&room=...` và `GET /api/search/compare?q=...`).
  - [x] Xây dựng giao diện Facebook Messenger Dark Mode 3 cột với tab A/B đối soát song song trực quan.
  - [x] Tích hợp cơ chế Hot-reload giao diện đọc trực tiếp từ đĩa.
- [x] **TASK-5.2: Xây dựng Search Query cho luồng Baseline**
  - [x] Dùng truy vấn `match` chuẩn của Elasticsearch trên index `chat_messages_baseline`.
- [x] **TASK-5.3: Xây dựng Search Query cho luồng Vietnamese Tokenizer**
  - [x] Bước tiền xử lý: Query của người dùng được tách từ qua Cốc Cốc Tokenizer và chuẩn hóa không dấu.
  - [x] Xây dựng truy vấn `bool` query đa tầng kết hợp độ ưu tiên (Relevance Boosting):
    1. **Tầng 1 (Exact Tokenized Match - Boost 5.0 + match_phrase Boost 4.0)**: Khớp chính xác từ ghép có dấu (`content_tokenized`).
    2. **Tầng 2 (Unaccented Token Match - Boost 3.0)**: Khớp từ ghép không dấu (`content_unaccented`).
    3. **Tầng 3 (Partial N-gram Match - Boost 1.0)**: Khớp một phần từ đang gõ (`content_partial`).

---

### 🔹 Phase 6: Đánh giá, So sánh & Benchmark
- [x] **TASK-6.1: So sánh chất lượng tìm kiếm (Relevance & Accuracy)**
  - [x] Soạn danh sách Test Suite đối chứng (12 kịch bản tìm kiếm có dấu, không dấu, bẫy từ ghép).
  - [x] Chạy đo đạc thực tế tự động và xuất báo cáo tại `docs/benchmark_report.md` với các chỉ số IR kinh điển:
    - **MRR (Mean Reciprocal Rank)**: **0.7708** (+23.3% so với Baseline 0.6250).
    - **NDCG@10**: Đạt mức tối ưu **0.9472**.
    - **Precision@1 (P@1)**: **75.0%** (+28.6% so với Baseline 58.3%).
    - **Precision@5 (P@5)**: **56.7%** (+30.8% so với Baseline 43.3%).
- [x] **TASK-6.2: Benchmark hiệu năng hệ thống (Latency & Throughput)**
  - [x] Đo đạc thời gian nạp Bulk API NDJSON: 233ms cho 131 tin nhắn (1.78ms/tin).
  - [x] Đo đạc Search Latency: Cốc Cốc trung bình **14.2ms** (nhanh hơn 25.4% so với Baseline 19.0ms).

---

### 🔹 Phase 7: Hoàn thiện Báo cáo & Tài liệu
- [x] **TASK-7.1: Viết tài liệu báo cáo lý thuyết hệ thống tìm kiếm** (`docs/search_architecture.md`).
- [x] **TASK-7.2: Viết báo cáo so sánh kết quả chi tiết kèm dẫn chứng** (`docs/benchmark_report.md`).
- [x] **TASK-7.3: Hoàn thiện `README.md` với hướng dẫn cài đặt, cấu hình và chạy mẫu**.

---

### 🔹 Phase 8: Tự Xây Dựng Động Cơ Lưu Trữ & Tìm Kiếm Nhị Phân Độc Lập (Custom Engine thuần Go)
- [x] **TASK-8.1: Thiết kế cấu trúc nhị phân & Tài liệu lý thuyết đặc tả**
  - [x] Viết tài liệu đặc tả kiến trúc: [`docs/custom_storage_engine_spec.md`](docs/custom_storage_engine_spec.md).
  - [x] Viết tài liệu luồng ghi tin nhắn & cấu trúc lưu trữ đĩa: [`docs/write_flow_and_storage_internals.md`](docs/write_flow_and_storage_internals.md).
- [x] **TASK-8.2: Xây dựng Hạ tầng Lưu trữ Nhị phân (`pkg/storage/`)**
  - [x] Bộ I/O nhị phân LittleEndian tối ưu (`pkg/storage/binary_io.go`).
  - [x] Forward Index tra cứu $O(1)$ tin nhắn gốc: `docstore.dat` & `docstore.idx` (12 bytes/doc cố định) (`pkg/storage/docstore.go`).
  - [x] Write-Ahead Log chống sập nguồn với kiểm tra toàn vẹn CRC32 & Replay Crash Recovery (`pkg/storage/wal.go`).
  - [x] Disk Segments bất biến: `terms.dict` (sắp xếp Alphabet A-Z), `postings.bin` (con trỏ vị trí & TF), `segments.meta` (`pkg/storage/segment.go`).
  - [x] Cơ chế xóa mềm Tombstone: `tombstone.del`.
  - [x] Viết Unit Test và chạy kiểm thử thành công 100% (`pkg/storage/storage_test.go`).
- [x] **TASK-8.3: Xây dựng Động cơ Tìm kiếm & Chỉ mục MemTable (`pkg/customengine/`)**
  - [x] Điều phối MemTable trong RAM, bóc tách tiếng Việt Cốc Cốc đa tầng (chuẩn, không dấu, edge n-gram) (`pkg/customengine/engine.go`).
  - [x] Thuật toán chấm điểm Okapi BM25 đa tầng (Boost 5x tokenized, 4x phrase, 3x unaccented, 1x ngram) (`pkg/customengine/search.go`).
  - [x] Viết Unit Test vòng đời End-to-End pass 100% (`pkg/customengine/engine_test.go`).
- [x] **TASK-8.4: Tích hợp Bộ điều phối kép Dual-Engine vào Backend & Web UI Messenger**
  - [x] Cập nhật `cmd/chat_server/main.go` để ghi đồng thời vào Custom Engine và Elasticsearch khi POST/PUT/DELETE.
  - [x] Cập nhật `/api/search/compare` thực thi song song và đo độ trễ cả 3 động cơ.
  - [x] Nâng cấp giao diện Web Messenger sang chế độ đối soát 3 chiều:
    - Cột 1: Cốc Cốc Tokenizer (Elasticsearch 8.x)
    - Cột 2: Standard Baseline (Elasticsearch 8.x)
    - Cột 3: Custom Engine (Go Binary Disk Persistence, 0% Docker, ~3ms Latency).
