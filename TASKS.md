# Danh Sách Nhiệm Vụ Triển Khai (Task Breakdown)

Dự án: **Cải thiện kết quả tìm kiếm tin nhắn tiếng Việt với Cốc Cốc Tokenizer & Elasticsearch trong Go**  
Tham chiếu đề bài gốc: [`REQUIREMENTS.md`](file:///d:/CODE/VSF/vietnamese-chat-search/REQUIREMENTS.md)

---

## 📌 Bảng Tiến Độ Tổng Quan

| Giai đoạn | Nội dung chính | Trạng thái | Ước lượng |
| :--- | :--- | :---: | :---: |
| **Phase 1** | Nghiên cứu lý thuyết Search Engine, Lucene & Elasticsearch | ⏳ To Do | 1-2 ngày |
| **Phase 2** | Thiết lập môi trường Docker (ES, Kibana) & Seed Data | ⏳ To Do | 1 ngày |
| **Phase 3** | Xây dựng CGO Wrapper cho Cốc Cốc Tokenizer trong Go | ⏳ To Do | 2-3 ngày |
| **Phase 4** | Thiết kế Elasticsearch Index Mapping & Pipeline Indexing | ⏳ To Do | 1-2 ngày |
| **Phase 5** | Xây dựng Search Engine Service & Multi-match Query Builder | ⏳ To Do | 1-2 ngày |
| **Phase 6** | Kiểm thử so sánh chất lượng search & Benchmark hiệu năng | ⏳ To Do | 1-2 ngày |
| **Phase 7** | Hoàn thiện tài liệu báo cáo & tổng kết | ⏳ To Do | 1 ngày |

---

## Chi Tiết Các Nhiệm Vụ (Task Checklist)

### 🔹 Phase 1: Nghiên cứu lý thuyết Search Engine & Kiến trúc
- [ ] **TASK-1.1: Nghiên cứu cơ chế Inverted Index của Apache Lucene**
  - [ ] Tìm hiểu cấu trúc dữ liệu: *Term Dictionary*, *Posting List*, *Term Frequencies*, *Positions*, *Doc Values*.
  - [ ] Hiểu cách Lucene lưu trữ và tra cứu từ khóa trong bộ nhớ / ổ đĩa.
- [ ] **TASK-1.2: Nghiên cứu hai luồng cốt lõi của Search Engine**
  - [ ] **Luồng Indexing**: `Character Filters` $\rightarrow$ `Tokenizer` $\rightarrow$ `Token Filters` $\rightarrow$ `Inverted Index` $\rightarrow$ `Storage`.
  - [ ] **Luồng Search & Ranking**: `Query Tokenizer` $\rightarrow$ `Filtering` $\rightarrow$ `Inverted Index Lookup` $\rightarrow$ `BM25 Scoring Algorithm`.
- [ ] **TASK-1.3: Phân tích bài toán đặc thù của tiếng Việt**
  - [ ] Phân tích vấn đề từ đơn vs từ ghép (ví dụ: *"học sinh"* vs *"sinh viên"*).
  - [ ] Phân tích vấn đề tìm kiếm không dấu (*"ca phe"* tìm *"cà phê"*).
  - [ ] Phân tích bài toán tìm kiếm tiền tố / một phần từ (*Partial Matching*).
  - [ ] Viết tài liệu tổng hợp lý thuyết tại `docs/search_engine_fundamentals.md`.

---

### 🔹 Phase 2: Môi trường phát triển & Dữ liệu mẫu (Data Seeding)
- [ ] **TASK-2.1: Khởi tạo cụm Elasticsearch & Kibana với Docker Compose**
  - [ ] Tạo file `docker-compose.yml` gồm Elasticsearch 8.x (hoặc 7.17) và Kibana.
  - [ ] Cấu hình single-node, tắt SSL/Security cho môi trường local để đơn giản hóa giao tiếp HTTP.
  - [ ] Thiết lập healthcheck đảm bảo ES sẵn sàng trước khi app kết nối.
- [ ] **TASK-2.2: Thiết lập môi trường Build CGO & Go**
  - [ ] Cấu hình `Dockerfile` hỗ trợ build đa tầng (multi-stage build) có sẵn `gcc`, `g++`, `cmake`, `git`, `go`.
  - [ ] Cấu hình `Makefile` với các lệnh thông dụng (`make up`, `make down`, `make build`, `make seed`, `make test`).
- [ ] **TASK-2.3: Xây dựng bộ dữ liệu mẫu tin nhắn chat tiếng Việt**
  - [ ] Tạo file `data/sample_messages.json` chứa 100 - 500 tin nhắn chat nhóm thực tế.
  - [ ] Đảm bảo bao phủ các trường hợp đặc biệt:
    - Cặp từ ghép trùng âm tiết: *"học sinh"*, *"sinh viên"*, *"hy sinh"*, *"sinh nhật"*.
    - Cặp từ đa nghĩa: *"cà phê"*, *"phê bình"*, *"bàn ghế"*, *"bàn luận"*.
    - Tin nhắn viết không dấu, viết tắt, telex.
    - Tin nhắn ngắn, tin nhắn dài, văn phong chat thường ngày.

---

### 🔹 Phase 3: Tích hợp Cốc Cốc Tokenizer trong Go qua CGO
- [ ] **TASK-3.1: Chuẩn bị mã nguồn C++ Cốc Cốc Tokenizer**
  - [ ] Tích hợp mã nguồn [coccoc/coccoc-tokenizer](https://github.com/coccoc/coccoc-tokenizer) và từ điển phân đoạn từ `sys.dic`.
  - [ ] Xây dựng thư viện liên kết động (`libcoccoc_tokenizer.so`) hoặc thư viện tĩnh (`.a`).
- [ ] **TASK-3.2: Xây dựng Go CGO Wrapper (`pkg/tokenizer`)**
  - [ ] Định nghĩa cgo header và các cflags/ldflags để liên kết C++ code vào Go.
  - [ ] Viết struct `Tokenizer` và hàm khởi tạo tải từ điển `sys.dic`.
  - [ ] Cung cấp hàm `Tokenize(text string) []string`: phân tách từ tiếng Việt và nối từ ghép bằng dấu gạch dưới (ví dụ: `["tôi", "uống", "cà_phê"]`).
- [ ] **TASK-3.3: Xây dựng Module chuẩn hóa không dấu (`pkg/tokenizer/unaccent.go`)**
  - [ ] Hàm chuẩn hóa Unicode NFC / NFD.
  - [ ] Hàm chuyển đổi chuỗi tiếng Việt có dấu sang không dấu (`"uống cà_phê"` $\rightarrow$ `"uong ca_phe"`).
- [ ] **TASK-3.4: Kiểm thử Unit Test cho Tokenizer**
  - [ ] Viết test cases kiểm tra độ chính xác của tokenizer với các câu tiếng Việt phức tạp.
  - [ ] Đo đạc benchmark tốc độ xử lý tách từ của CGO wrapper (`go test -bench`).

---

### 🔹 Phase 4: Thiết kế Elasticsearch Mapping & Xây dựng Indexer
- [ ] **TASK-4.1: Thiết kế Elasticsearch Mapping (Index Schema)**
  - [ ] Tạo index `chat_messages_baseline` (dùng `standard` analyzer mặc định của ES).
  - [ ] Tạo index `chat_messages_vietnamese` với multi-field mapping:
    - `content`: Văn bản gốc dùng để hiển thị kết quả.
    - `content_tokenized`: Lưu các token đã qua Cốc Cốc tokenizer (analyzer tách bằng khoảng trắng `whitespace` + `lowercase`).
    - `content_unaccented`: Lưu token không dấu (`whitespace` + `lowercase`).
    - `content_partial`: Lưu token kết hợp custom token filter `edge_ngram` (min_gram: 2, max_gram: 15) phục vụ tìm kiếm dở từ / autocomplete.
- [ ] **TASK-4.2: Xây dựng công cụ nạp dữ liệu (Go Indexer CLI)**
  - [ ] Tạo CLI `cmd/indexer` kết nối Elasticsearch qua thư viện Go client (`elastic/go-elasticsearch`).
  - [ ] Đọc dữ liệu từ `data/sample_messages.json`.
  - [ ] Xử lý tokenize nội dung qua Cốc Cốc Tokenizer và gửi Bulk Index vào cả 2 chỉ mục (baseline & vietnamese).
  - [ ] Đo thời gian hoàn thành quá trình index.

---

### 🔹 Phase 5: Xây dựng Search Service & Query Builder
- [ ] **TASK-5.1: Xây dựng CLI / HTTP Search Service (`cmd/searcher`)**
  - [ ] Tạo giao diện dòng lệnh (CLI) hoặc REST endpoint (`GET /api/search?q=...&mode=...`) cho phép test linh hoạt.
- [ ] **TASK-5.2: Xây dựng Search Query cho luồng Baseline**
  - [ ] Dùng truy vấn `match` chuẩn của Elasticsearch trên index `chat_messages_baseline`.
- [ ] **TASK-5.3: Xây dựng Search Query cho luồng Vietnamese Tokenizer**
  - [ ] Bước tiền xử lý: Query của người dùng được tách từ qua Cốc Cốc Tokenizer và chuẩn hóa không dấu.
  - [ ] Xây dựng truy vấn `bool` query đa tầng kết hợp độ ưu tiên (Relevance Boosting):
    1. **Tầng 1 (Exact Tokenized Match - Boost 5.0)**: Khớp chính xác từ ghép có dấu (`match_phrase` trên `content_tokenized`).
    2. **Tầng 2 (Unaccented Token Match - Boost 3.0)**: Khớp từ ghép không dấu (`match` trên `content_unaccented`).
    3. **Tầng 3 (Partial N-gram Match - Boost 1.0)**: Khớp một phần từ đang gõ (`match` trên `content_partial`).

---

### 🔹 Phase 6: Đánh giá, So sánh & Benchmark
- [ ] **TASK-6.1: So sánh chất lượng tìm kiếm (Relevance & Accuracy)**
  - [ ] Soạn danh sách các test case đối chứng (ít nhất 10 kịch bản tìm kiếm phổ biến):
    - Tìm từ ghép: `"học sinh"` (kiểm tra xem có lẫn *"sinh viên"* không).
    - Tìm từ ghép: `"cà phê"` (kiểm tra xem có lẫn *"phê bình"* không).
    - Tìm không dấu: `"hoc sinh"`, `"ca phe"`.
    - Tìm partial / dở từ: `"học s"`, `"cà ph"`, `"sinh v"`.
  - [ ] Chạy song song cả 2 luồng và xuất bảng so sánh kết quả (Top kết quả, điểm score, tỷ lệ False Positives).
- [ ] **TASK-6.2 (Optional): Benchmark hiệu năng hệ thống**
  - [ ] So sánh thông lượng (throughput - docs/sec) khi index 10.000 tin nhắn (Baseline vs Cốc Cốc CGO).
  - [ ] So sánh thời gian phản hồi truy vấn (Search Latency p50, p95, p99) giữa 2 phương thức.
  - [ ] Phân tích overhead của CGO call trong Go.

---

### 🔹 Phase 7: Hoàn thiện Báo cáo & Tài liệu
- [ ] **TASK-7.1: Viết tài liệu báo cáo lý thuyết hệ thống tìm kiếm** (`docs/search_architecture.md`).
- [ ] **TASK-7.2: Viết báo cáo so sánh kết quả chi tiết kèm dẫn chứng** (`docs/evaluation_results.md`).
- [ ] **TASK-7.3: Hoàn thiện `README.md` với hướng dẫn cài đặt, cấu hình và chạy mẫu**.
