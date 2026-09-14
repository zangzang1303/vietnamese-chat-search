# Vietnamese Chat Message Search (Go + Cốc Cốc Tokenizer + Elasticsearch)

Dự án nghiên cứu và phát triển giải pháp **tối ưu hóa tìm kiếm tin nhắn tiếng Việt** cho hệ thống chat nhóm (Group Chat). Giải pháp kết hợp bộ tách từ vựng tiếng Việt **[Cốc Cốc Tokenizer](https://github.com/coccoc/coccoc-tokenizer)** được nhúng vào ứng dụng **Go** thông qua **CGO**, lưu trữ và đánh chỉ mục trên **Elasticsearch**.

> 📌 **Tài liệu liên quan**:
> - [REQUIREMENTS.md](REQUIREMENTS.md): Đề bài và yêu cầu gốc của bài toán.
> - [TASKS.md](TASKS.md): Bảng phân rã chi tiết toàn bộ các task công việc (WBS & Checklist).
> - [daily_log/](daily_log/): Thư mục nhật ký công việc và tiến độ theo từng ngày.
> - [docs/references.md](docs/references.md): Danh mục tài liệu đọc và nghiên cứu chuyên sâu.

---

## 1. Bối Cảnh & Vấn Đề (Problem Statement)

* **Hiện trạng**: Chat 1-1 được mã hóa đầu cuối (E2EE), server không can thiệp. Chat nhóm không mã hóa, server lưu trữ và cung cấp tính năng tìm kiếm tin nhắn.
* **Vấn đề**: Khi sử dụng bộ phân tích mặc định (`standard analyzer`) của Elasticsearch (dựa trên tách từ theo khoảng trắng của tiếng Anh), kết quả tìm kiếm tiếng Việt cho độ chính xác (Precision) rất thấp:
  * Tiếng Việt có từ ghép đa âm tiết (*"học sinh"*, *"sinh viên"*, *"cà phê"*...). Khi tách rời theo khoảng trắng, người dùng tìm kiếm cụm từ `"học sinh"` sẽ bị lẫn hàng loạt kết quả chứa từ `"sinh viên"` hoặc `"hy sinh"` do chung âm tiết `"sinh"`.
  * Khó khăn khi người dùng gõ **không dấu** (*"uong ca phe"*) hoặc **gõ dở từ/tiền tố** (*"cà ph"*, *"sinh v"*).
* **Mục tiêu**: Cải thiện độ chính xác và chất lượng tìm kiếm tiếng Việt với:
  1. **Word-level Matching**: Nhận diện chính xác ranh giới từ ghép tiếng Việt bằng Cốc Cốc Tokenizer.
  2. **Accented & Unaccented Search**: Hỗ trợ tìm kiếm cả tiếng Việt có dấu và không dấu.
  3. **Partial Matching**: Hỗ trợ tìm kiếm một phần từ / tiền tố (autocomplete / partial match).

---

## 2. Kiến Trúc Giải Pháp (Architecture)

```
[Người dùng / Client]
         │
         ▼ (Gõ từ khóa: "hoc sinh" / "ca ph")
┌─────────────────────────────────────────────────────────────┐
│                       GO SEARCH SERVICE                     │
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │               CGO Tokenizer Bridge                  │   │
│   │   Go Code  <──────(CGO)──────>  libcoccoc_tokenizer │   │
│   │                                 + Dict (sys.dic)    │   │
│   └─────────────────────────────────────────────────────┘   │
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │    Text Normalizer (Lowercase + Unaccent folding)   │   │
│   └─────────────────────────────────────────────────────┘   │
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │    Elasticsearch Query Builder (Multi-Match/Bool)   │   │
│   └─────────────────────────────────────────────────────┘   │
└──────────────────────────────┬──────────────────────────────┘
                               │ HTTP / JSON
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    ELASTICSEARCH CLUSTER                    │
│                                                             │
│   Index: chat_messages_vietnamese                           │
│   ├── content (Văn bản gốc hiển thị)                        │
│   ├── content_tokenized (Từ ghép có dấu: "học_sinh")        │
│   ├── content_unaccented (Từ ghép không dấu: "hoc_sinh")    │
│   └── content_partial (Edge N-gram tokens phục vụ partial)  │
└─────────────────────────────────────────────────────────────┘
```

### Nguyên lý hai luồng xử lý:
1. **Luồng Indexing**:
   - Go nhận tin nhắn thô $\rightarrow$ Gọi thư viện C++ Cốc Cốc qua CGO để tokenize thành chuỗi từ ghép (nối gạch dưới, ví dụ: `học_sinh`).
   - Go chuẩn hóa không dấu và đẩy dữ liệu vào các trường tương ứng của Elasticsearch index.
2. **Luồng Search & Ranking**:
   - Query tìm kiếm được Go tokenize bằng Cốc Cốc $\rightarrow$ Sinh truy vấn `bool` query đa tầng có trọng số điểm (Relevance Boosting):
     - **Match từ ghép có dấu** (Boost 5.0) $\rightarrow$ Ưu tiên cao nhất.
     - **Match từ ghép không dấu** (Boost 3.0) $\rightarrow$ Ưu tiên nhì.
     - **Match tiền tố / Partial N-gram** (Boost 1.0) $\rightarrow$ Đảm bảo tìm được khi gõ dở từ.

---

## 3. Cấu Trúc Thư Mục Dự Án (Project Structure)

```text
vietnamese-chat-search/
├── cmd/
│   ├── indexer/                # Công cụ nạp dữ liệu mẫu vào Elasticsearch
│   └── searcher/               # Ứng dụng CLI/API thực thi tìm kiếm & so sánh
├── pkg/
│   ├── tokenizer/              # Go CGO binding giao tiếp với Cốc Cốc C++
│   │   ├── coccoc.go           # Cgo wrapper & API tách từ
│   │   ├── unaccent.go         # Hàm chuẩn hóa loại bỏ dấu tiếng Việt
│   │   └── coccoc/             # Mã nguồn C++ Cốc Cốc tokenizer + sys.dic
│   └── es/                     # Client Elasticsearch, schema mapping & query builder
├── data/
│   └── sample_messages.json    # Tập dữ liệu mẫu tin nhắn chat tiếng Việt
├── docker/
│   ├── Dockerfile              # Môi trường build Go + GCC/CMake/CGO đa nền tảng
│   └── docker-compose.yml      # Cụm Elasticsearch + Kibana
├── docs/                       # Tài liệu nghiên cứu lý thuyết & báo cáo so sánh
│   ├── search_fundamentals.md  # Báo cáo Inverted Index, Lucene & luồng xử lý
│   └── evaluation_results.md   # Kết quả đánh giá so sánh (có vs không có tokenizer)
├── daily_log/                  # Thư mục lưu nhật ký công việc theo từng ngày
│   ├── TEMPLATE.md             # Mẫu nhật ký để copy cho ngày mới
│   ├── 2026-09-14.md           # Nhật ký ngày 1
│   └── README.md               # Mục lục tổng hợp các ngày
├── Makefile                    # Lệnh tiện ích: build, up, seed, test, bench
├── REQUIREMENTS.md             # Đề bài & yêu cầu ban đầu
├── TASKS.md                    # Bảng phân rã nhiệm vụ (WBS)
└── README.md                   # Tài liệu chính của dự án
```

---

## 4. Công Nghệ Sử Dụng (Tech Stack)

* **Ngôn ngữ**: [Go](https://go.dev/) (Golang $\ge$ 1.21).
* **Giao tiếp C/C++**: [CGO](https://pkg.go.dev/cmd/cgo) để gọi native dynamic/static library của Cốc Cốc.
* **Thư viện Tokenizer**: [coccoc-tokenizer](https://github.com/coccoc/coccoc-tokenizer) (C++11, hiệu năng cao, dựa trên từ điển).
* **Search Engine**: [Elasticsearch](https://www.elastic.co/elasticsearch) 8.x / 7.17 (chạy trên Docker).
* **Orchestration**: Docker & Docker Compose.

---

## 5. Hướng Dẫn Cài Đặt & Chạy Mẫu (Quickstart)

### Yêu cầu tiên quyết:
* Máy đã cài đặt **Docker Desktop** (hoặc Linux/WSL2 có Docker & Docker Compose).

### Bước 1: Khởi động Elasticsearch & Kibana
```bash
docker compose -f docker/docker-compose.yml up -d
```
* Elasticsearch: `http://localhost:9200`
* Kibana: `http://localhost:5601`

### Bước 2: Build ứng dụng & nạp dữ liệu mẫu (Seed Data)
```bash
# Build và chạy nạp dữ liệu mẫu
docker compose -f docker/docker-compose.yml run --rm app-indexer
```

### Bước 3: Chạy thử nghiệm và so sánh kết quả tìm kiếm
```bash
# Chạy script so sánh song song 2 luồng (Baseline vs Cốc Cốc)
docker compose -f docker/docker-compose.yml run --rm app-searcher --query="học sinh"
```

---

## 6. Kế Hoạch & Tiến Độ

Xem chi tiết danh sách checklist công việc tại [TASKS.md](TASKS.md).
