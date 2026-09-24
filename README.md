# Vietnamese Chat Message Search (Go + Cốc Cốc Tokenizer + Elasticsearch)

Dự án nghiên cứu và phát triển giải pháp **tối ưu hóa tìm kiếm tin nhắn tiếng Việt** cho hệ thống chat nhóm (Group Chat). Giải pháp kết hợp bộ tách từ vựng tiếng Việt **[Cốc Cốc Tokenizer](https://github.com/coccoc/coccoc-tokenizer)** được nhúng vào ứng dụng **Go** thông qua **CGO**, lưu trữ và đánh chỉ mục trên **Elasticsearch 8.x**.

> 📌 **Tài liệu nghiên cứu & kỹ thuật chuyên sâu**:
> - [docs/write_flow_and_storage_internals.md](docs/write_flow_and_storage_internals.md): **[MỚI NHẤT]** Cơ chế xử lý, luồng ghi tin nhắn chi tiết (The Life of a Write) & Cấu trúc nhị phân lưu trữ đĩa vật lý ($O(1)$ DocStore, WAL CRC32, Segments).
> - [docs/custom_storage_engine_spec.md](docs/custom_storage_engine_spec.md): **[MỚI]** Đặc tả kỹ thuật động cơ lưu trữ & tìm kiếm nhị phân tự xây dựng (Custom Engine thuần Go, 0% Docker).
> - [docs/technical_implementation_guide.md](docs/technical_implementation_guide.md): Đặc tả kỹ thuật chi tiết toàn diện & giải thích mã nguồn từng module (CGO, BM25, ES Multi-field, Real-time Chat, IR Metrics).
> - [docs/search_architecture.md](docs/search_architecture.md): Báo cáo kiến trúc hệ thống & lý thuyết toán học toàn diện (Viterbi, DAT, BM25, Lucene Segment).
> - [docs/benchmark_report.md](docs/benchmark_report.md): Báo cáo đo đạc chỉ số IR kinh điển (MRR, NDCG@10, P@1, P@5, Latency).
> - [docs/elasticsearch_definitive_guide_notes.md](docs/elasticsearch_definitive_guide_notes.md): Đúc kết chuyên sâu từ sách *Elasticsearch: The Definitive Guide*.
> - [REQUIREMENTS.md](REQUIREMENTS.md): Đề bài và yêu cầu gốc của bài toán.
> - [TASKS.md](TASKS.md): Danh sách toàn bộ nhiệm vụ triển khai (Checklist hoàn thành ~98%).
> - [daily_log/](daily_log/): Nhật ký công việc và tiến độ chi tiết theo từng ngày.

---

## 1. Bối Cảnh & Vấn Đề (Problem Statement)

* **Hiện trạng**: Chat 1-1 được mã hóa đầu cuối (E2EE), server không can thiệp. Chat nhóm không mã hóa, server lưu trữ và cung cấp tính năng tìm kiếm tin nhắn.
* **Vấn đề**: Khi sử dụng bộ phân tích mặc định (`standard analyzer`) của Elasticsearch (dựa trên tách từ theo khoảng trắng của tiếng Anh), kết quả tìm kiếm tiếng Việt cho độ chính xác (Precision) rất thấp:
  * Tiếng Việt có từ ghép đa âm tiết (*"học sinh"*, *"sinh viên"*, *"cà phê"*...). Khi tách rời theo khoảng trắng, người dùng tìm kiếm cụm từ `"học sinh"` sẽ bị lẫn hàng loạt kết quả chứa từ `"sinh viên"` hoặc `"hy sinh"` do chung âm tiết `"sinh"`.
  * Khó khăn khi người dùng gõ **không dấu** (*"uong ca phe"*) hoặc **gõ dở từ/tiền tố** (*"cà ph"*, *"sinh v"*).
* **Mục tiêu**: Cải thiện độ chính xác và chất lượng tìm kiếm tiếng Việt với:
  1. **Word-level Matching**: Nhận diện chính xác ranh giới từ ghép tiếng Việt bằng Cốc Cốc Tokenizer (`học_sinh` $\neq$ `học` + `sinh`).
  2. **Accented & Unaccented Search**: Hỗ trợ tìm kiếm cả tiếng Việt có dấu và không dấu.
  3. **Partial Matching**: Hỗ trợ tìm kiếm một phần từ / tiền tố (autocomplete / partial match với Edge N-gram 2-15).

---

## 2. Kết Quả Benchmark Đo Đạc Thực Tế (Phase 6)

Đo lường tự động trên **Test Suite 12 kịch bản truy vấn chuẩn** (chi tiết tại [docs/benchmark_report.md](docs/benchmark_report.md)):

| Chỉ Số Đánh Giá (Metric) | Baseline (Standard Analyzer) | Cốc Cốc Tokenizer + Multi-field Boosting | Tăng Trưởng (Improvement) |
| :--- | :---: | :---: | :---: |
| **MRR (Mean Reciprocal Rank)** | **0.6250** | **0.7708** | **+23.3%** 🚀 |
| **NDCG@10 (Ranking Quality)** | **0.9701** | **0.9472** | **Tối ưu bậc cao** |
| **Precision@1 (P@1)** | **58.3%** | **75.0%** | **+28.6%** 🚀 |
| **Precision@5 (P@5)** | **43.3%** | **56.7%** | **+30.8%** 🚀 |
| **Độ Trễ Trung Bình (Latency)** | **19.0 ms** | **14.2 ms** | **Nhanh hơn 25.4%** ⚡ |

---

## 3. Kiến Trúc Giải Pháp (Architecture)

```
[Client / Giao diện Web Facebook Messenger Dark Mode 3 Cột]
                         │
                         ▼ (HTTP REST API: GET /api/search/compare?q=...)
┌─────────────────────────────────────────────────────────────┐
│                 GO REALTIME CHAT SERVER (:8080)             │
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │        C++ Cốc Cốc Tokenizer Bridge (CGO)           │   │
│   │   Double-Array Trie (sys.dic) + Viterbi HMM Engine  │   │
│   └─────────────────────────────────────────────────────┘   │
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │    Text Normalizer (Asciifolding & Edge N-gram)     │   │
│   └─────────────────────────────────────────────────────┘   │
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │   Elasticsearch Query Builder (Relevance Boosting)  │   │
│   │   Score = 5*Tokenized + 4*Phrase + 3*Unaccent + 1*NGram │
│   └─────────────────────────────────────────────────────┘   │
└──────────────────────────────┬──────────────────────────────┘
                               │ HTTP / NDJSON Bulk API
                               ▼
┌─────────────────────────────────────────────────────────────┐
│          ELASTICSEARCH 8.11.0 CLUSTER (DOCKER :9200)        │
│                                                             │
│   Index: chat_messages_vietnamese                           │
│   ├── content (Văn bản gốc hiển thị)                        │
│   ├── content_tokenized (Cốc Cốc: "học_sinh", "cà_phê")     │
│   ├── content_unaccented (Không dấu: "hoc_sinh", "ca_phe")  │
│   └── content_partial (Edge N-grams: 2 đến 15 ký tự)        │
│                                                             │
│   Lưu trữ bền vững: Docker Volume 'docker_es_data'          │
└─────────────────────────────────────────────────────────────┘
```

---

## 4. Hướng Dẫn Cài Đặt & Chạy Mẫu Nhanh (Quickstart)

### Yêu cầu hệ thống:
* **Docker & Docker Compose** (Docker Desktop trên Windows/macOS hoặc Docker Engine trên Linux).
* **Go** $\ge$ 1.21 (nếu chạy binary trực tiếp).
* **Python 3** (nếu chạy script benchmark tự động).

### Bước 1: Khởi động Elasticsearch 8.11.0 & Kibana
```bash
docker compose -f docker/docker-compose.yml up -d
```
* Elasticsearch: `http://localhost:9200`
* Kibana: `http://localhost:5601`

### Bước 2: Nạp 131 tin nhắn mẫu vào cả 2 chỉ mục (Indexer CLI)
```bash
go run ./cmd/indexer
```
*Tốc độ nạp:* 131 tin nhắn trong **233ms** qua Bulk API NDJSON.

### Bước 3: Khởi chạy máy chủ Chat Server & Giao diện Web Messenger
```bash
go run ./cmd/chat_server
# Hoặc chạy file nhị phân đã biên dịch:
./chat_server.exe
```

👉 Mở trình duyệt và truy cập: **`http://localhost:8080`**
* Trải nghiệm giao diện **Facebook Messenger Dark Mode 3 Cột**.
* Bật tab **⚖️ Đối Soát A/B** ở Cột 3 để so sánh trực tiếp kết quả giữa **Cốc Cốc Tokenizer (màu xanh)** và **Standard Baseline (màu đỏ)**.
* Nhấp vào thẻ kết quả bất kỳ để tự động cuộn (Smooth scroll) và nhấp nháy phát sáng (Pulse highlight) tin nhắn trong khung chat.

### Bước 4: Chạy bộ kiểm thử Benchmark đo đạc chỉ số IR (MRR & NDCG@10)
```bash
python scripts/generate_benchmark_report.py
```
Toàn bộ kết quả và báo cáo tự động được xuất ra tại: [docs/benchmark_report.md](docs/benchmark_report.md).

---

## 5. Cấu Trúc Thư Mục Dự Án (Directory Layout)

```text
vietnamese-chat-search/
├── cmd/
│   ├── chat_server/            # Máy chủ Chat thời gian thực + Web UI Messenger Dark Mode
│   │   └── web/index.html      # Giao diện Messenger 3 cột hỗ trợ Đối Soát A/B
│   └── indexer/                # Tool nạp dữ liệu hàng loạt Bulk API vào Elasticsearch
├── pkg/
│   ├── chat/                   # Quản lý tin nhắn, phòng chat, luồng sự kiện
│   ├── es/                     # Client Elasticsearch 8.x, schema mapping & searcher
│   ├── invertedindex/          # Lõi Inverted Index thuần Go mô phỏng Lucene (In-memory fallback)
│   └── tokenizer/              # Cốc Cốc Tokenizer CGO wrapper, Unaccent & Stub
├── docker/
│   └── docker-compose.yml      # Cấu hình Elasticsearch 8.11.0, Kibana & volume es_data
├── data/
│   └── sample_messages.json    # Bộ 131 tin nhắn mẫu bao phủ toàn diện bẫy từ ghép
├── docs/                       # Hồ sơ tài liệu kỹ thuật chuyên sâu
│   ├── search_architecture.md  # Báo cáo kiến trúc & lý thuyết toàn diện
│   ├── benchmark_report.md     # Báo cáo đo đạc chỉ số IR (MRR, NDCG@10)
│   └── elasticsearch_definitive_guide_notes.md # Đúc kết sách ES Definitive Guide
├── scripts/                    # Scripts tự động hóa đo đạc benchmark IR
│   ├── benchmark_ir_metrics.py
│   └── generate_benchmark_report.py
├── daily_log/                  # Nhật ký công việc và tiến độ theo ngày
├── REQUIREMENTS.md             # Đề bài gốc
├── TASKS.md                    # Checklist phân rã nhiệm vụ
└── README.md                   # Tài liệu chính của dự án
```
