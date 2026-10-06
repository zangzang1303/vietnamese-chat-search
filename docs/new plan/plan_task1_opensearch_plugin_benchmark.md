# TÀI LIỆU KẾ HOẠCH CHI TIẾT: TASK 1
## ĐÁNH GIÁ ẢNH HƯỞNG CỦA PLUGIN TRÊN OPENSEARCH 2.19 (BENCHMARK & SEARCH PARITY)

---

## 1. TỔNG QUAN & BỐI CẢNH DỰ ÁN

### 1.1. Bối cảnh kỹ thuật
Hệ thống tìm kiếm tin nhắn chat tiếng Việt đòi hỏi khả năng xử lý ngôn ngữ tự nhiên đặc thù: từ ghép (compound words, ví dụ: "học sinh", "cà phê", "Việt Nam"), tiếng Việt không dấu (unaccented), và tìm kiếm tiền tố tức thời (prefix / edge n-gram cho trải nghiệm search-as-you-type).

Để đáp ứng yêu cầu trên trên nền tảng **OpenSearch 2.19.0**, hiện có 2 hướng tiếp cận kiến trúc chính cần được đánh giá định lượng:
1. **Mô hình A (App-side CGO Pre-tokenization)**:
   - Cụm OpenSearch 2.19.0 chỉ cài đặt plugin chuẩn `analysis-icu`.
   - Backend Go đóng vai trò tiền xử lý: tích hợp Cốc Cốc Tokenizer qua CGO (hoặc wrapper C++) để nhận diện từ ghép (ghép bằng dấu `_`), chuẩn hóa và sinh dữ liệu các trường trước khi nạp qua OpenSearch Bulk API.
2. **Mô hình B (In-engine Plugin Tokenization)**:
   - Cụm OpenSearch 2.19.0 cài đặt trực tiếp plugin `opensearch-analysis-vietnamese` (fork bởi Duy Đỗ) kèm theo native C++ JNI library (`libcoccoc_tokenizer_jni.so`) và các tệp từ điển Cốc Cốc.
   - Backend Go chỉ gửi nội dung tin nhắn thô; OpenSearch node tự động thực hiện phân đoạn từ vựng thông qua analyzer `vi_analyzer` / `vi_tokenizer` trực tiếp bên trong engine.

### 1.2. Mục tiêu trọng tâm
1. **Định lượng hiệu năng Indexing**: So sánh Throughput (docs/sec, MB/sec) và Latency ($p50, p95, p99$) khi nạp khối lượng dữ liệu lớn (500.000 đến 1.000.000 tin nhắn).
2. **Đánh giá mức tiêu thụ tài nguyên OpenSearch**: Đo lường % CPU utilization, Memory RSS, JVM Garbage Collection pauses, và Threadpool Queue/Rejection của container OpenSearch trong suốt quá trình indexing liên tục.
3. **Giả lập tải đa luồng (Concurrent Stress Testing)**: Thử nghiệm với nhiều luồng đồng thời (concurrency: 4, 8, 16, 32 workers) trong thời gian đủ dài (15 – 30 phút).
4. **Bảo đảm Search Parity (Zero Divergence)**: Khẳng định rằng cả 2 mô hình đều tạo ra bộ terms inverted index tương đương 100%, kết quả tìm kiếm không có bất kỳ sự sai lệch nào trên cả 4 trường dữ liệu (`text`, `text_vi`, `text_raw`, `text_edge`).

---

## 2. THIẾT KẾ MAPPINGS & ĐỒNG NHẤT HÓA TOKEN (SEARCH PARITY)

### 2.1. Cấu trúc Index Messages & Yêu cầu Token Output
Index `messages` có schema tiêu chuẩn như sau:
* Các trường định danh: `id` (long, sort desc), `tenant` (keyword), `app_id` (keyword), `user_id` (keyword), `thread_id` (keyword), `hide` (keyword), `create_at` (long, index: false).
* 4 trường nội dung tìm kiếm:
  1. `text`: Text thô phân tích bằng default analyzer của OpenSearch.
  2. `text_vi`: Text có ghép từ vựng tiếng Việt Cốc Cốc.
  3. `text_raw`: Text có ghép từ vựng + loại bỏ dấu tiếng Việt (ICU folding).
  4. `text_edge`: Text hỗ trợ search tiền tố (Edge N-gram 3 đến 10 ký tự, search_analyzer: `default`).

#### Bảng so sánh kỳ vọng Token Output cho cụm từ: *"Tôi yêu Việt Nam"*

| Trường tìm kiếm | Analyzer áp dụng | Kỳ vọng Tokens sinh ra trong Inverted Index |
| :--- | :--- | :--- |
| **`text`** | Default / Standard Analyzer | `["tôi", "yêu", "việt", "nam"]` |
| **`text_vi`** | Cốc Cốc Vietnamese Analyzer | `["tôi", "yêu", "việt_nam"]` |
| **`text_raw`** | ICU Unaccent + Cốc Cốc | `["toi", "yeu", "viet_nam"]` |
| **`text_edge`** | Custom Edge N-gram (3-10) | `["toi", "yeu", "vie", "viet", "viet_n", "viet_na", "viet_nam"]` |

### 2.2. Phân tích kỹ thuật & Đồng bộ Analyzer giữa 2 Mô hình

#### Cơ chế của Mô hình A (Go Backend CGO Tokenizer):
1. Go backend gọi hàm `coccoc.SegmentOriginal("Tôi yêu Việt Nam")` -> Chuỗi sinh ra: `"Tôi yêu Việt_Nam"`.
2. Backend chuẩn bị tài liệu:
   * `text`: `"Tôi yêu Việt Nam"`
   * `text_vi`: `"Tôi yêu Việt_Nam"`
   * `text_raw`: `"Tôi yêu Việt_Nam"`
   * `text_edge`: `"Tôi yêu Việt_Nam"`
3. Trong OpenSearch:
   * Với `text_vi`: Sử dụng analyzer có `tokenizer: standard` và `filter: [lowercase]`. Do trong chuẩn Lucene Standard Tokenizer, dấu gạch dưới `_` được tính là ký tự nối từ (`\w`), nên `"Việt_Nam"` được giữ nguyên thành token đơn `"việt_nam"`.
   * Với `text_raw`: Analyzer gồm `tokenizer: standard`, filter: `[lowercase, icu_folding]`. Ký tự có dấu được chuyển thành không dấu: `"việt_nam"` -> `"viet_nam"`.
   * Với `text_edge`: Analyzer gồm `tokenizer: standard`, filter: `[lowercase, icu_folding, custom_edge_ngram]`. `viet_nam` được băm thành: `vie`, `viet`, `viet_n`, `viet_na`, `viet_nam`.

#### Cơ chế của Mô hình B (OpenSearch Plugin `opensearch-analysis-vietnamese`):
1. Go backend chỉ gửi văn bản thô `"Tôi yêu Việt Nam"` cho cả 4 trường.
2. Trong OpenSearch:
   * `text`: Dùng analyzer tiêu chuẩn (`standard`).
   * `text_vi`: Dùng `vi_analyzer` (sử dụng `vi_tokenizer` gọi JNI C++ Cốc Cốc). Cần cấu hình để token xuất ra có dấu gạch dưới `_` thay vì dấu cách, đảm bảo đồng nhất với Mô hình A.
   * `text_raw`: Sử dụng `vi_tokenizer` kết hợp `[lowercase, icu_folding]`.
   * `text_edge`: Sử dụng `vi_tokenizer` kết hợp `[lowercase, icu_folding, custom_edge_ngram]`.

### 2.3. Quy trình Kiểm chứng Search Parity (Zero Diff Verification)
Trước khi bước vào giai đoạn benchmark hiệu năng, bắt buộc phải thực hiện bước kiểm chứng tính tương đồng:
1. **Analyze API Verification**:
   Chạy lệnh `POST /_analyze` trên cả 2 cluster với 100 câu mẫu đa dạng (tiếng Việt có dấu, không dấu, từ ghép 2-4 âm tiết, tên riêng, tiếng lóng, ký tự đặc biệt).
   So sánh danh sách tokens, start_offset, end_offset và position.
2. **Search Hits Parity Test**:
   Index 10.000 documents vào cả 2 cluster.
   Thực thi 1.000 queries ngẫu nhiên thuộc 3 nhóm:
   * Match từ đơn: `"tôi"`, `"yêu"`.
   * Match cụm từ ghép: `"việt nam"`, `"cà phê"`, `"học sinh"`.
   * Match tiền tố edge: `"vie"`, `"viet_n"`, `"hoc_s"`.
   * Match không dấu: `"viet nam"`, `"ca phe"`.
   Yêu cầu: Kết quả `total_hits` và thứ tự danh sách ID trả về giữa 2 bên phải **trùng khớp 100%**.

---

## 3. THIẾT KẾ MÔI TRƯỜNG THỰC NGHIỆM & KIỂM SOÁT BIẾN SỐ

### 3.1. Cấu hình Hạ tầng & Docker Compose

Để loại bỏ hoàn toàn các sai số do biến động tài nguyên ngoại cảnh, 2 môi trường được đóng gói và cấu hình hoàn toàn độc lập với Resource Quotas nghiêm ngặt.

```yaml
# docker-compose.benchmark.yml
version: '3.8'

services:
  # -------------------------------------------------------------
  # MÔ HÌNH A: OpenSearch Vanilla + Analysis-ICU
  # -------------------------------------------------------------
  opensearch-vanilla:
    image: opensearchproject/opensearch:2.19.0
    container_name: os_vanilla_bench
    environment:
      - cluster.name=bench-vanilla-cluster
      - node.name=node-vanilla
      - discovery.type=single-node
      - bootstrap.memory_lock=true
      - "OPENSEARCH_JAVA_OPTS=-Xms3g -Xmx3g -XX:+UseG1GC"
      - DISABLE_SECURITY_PLUGIN=true
      - DISABLE_INSTALL_DEMO_CONFIG=true
    ulimits:
      memlock:
        soft: -1
        hard: -1
      nofile:
        soft: 65536
        hard: 65536
    deploy:
      resources:
        limits:
          cpus: '4.0'
          memory: 6144M
        reservations:
          cpus: '4.0'
          memory: 6144M
    ports:
      - "9201:9200"

  # -------------------------------------------------------------
  # MÔ HÌNH B: OpenSearch + Vietnamese Plugin + CocCoc JNI
  # -------------------------------------------------------------
  opensearch-vietnamese:
    build:
      context: .
      dockerfile: Dockerfile.opensearch-vi
      args:
        OS_VERSION: 2.19.0
    container_name: os_vietnamese_bench
    environment:
      - cluster.name=bench-vi-cluster
      - node.name=node-vietnamese
      - discovery.type=single-node
      - bootstrap.memory_lock=true
      - "OPENSEARCH_JAVA_OPTS=-Xms3g -Xmx3g -XX:+UseG1GC"
      - DISABLE_SECURITY_PLUGIN=true
      - DISABLE_INSTALL_DEMO_CONFIG=true
    ulimits:
      memlock:
        soft: -1
        hard: -1
      nofile:
        soft: 65536
        hard: 65536
    deploy:
      resources:
        limits:
          cpus: '4.0'
          memory: 6144M
        reservations:
          cpus: '4.0'
          memory: 6144M
    ports:
      - "9202:9200"
```

### 3.2. Cấu hình Index tối ưu cho Bulk Ingestion
Theo đúng thiết lập thực tế từ đề bài:
* `refresh_interval`: `"-1"` (Tắt hoàn toàn auto-refresh trong khi index, chỉ refresh thủ công khi kết thúc bài test để đo lường thuần túy tốc độ nạp dữ liệu).
* `number_of_shards`: `4` (Phù hợp với 4 CPU cores được cấp phát).
* `number_of_replicas`: `0` (Không replica trong quá trình bulk nạp lớn).
* `sort.field`: `"id"`, `sort.order`: `"desc"` (Mô phỏng cơ chế index sorting của hệ thống chat thực tế).

---

## 4. BỘ DỮ LIỆU THỬ NGHIỆM & KỊCH BẢN STRESS-TEST ĐA LUỒNG

### 4.1. Bộ dữ liệu tổng hợp thực tế (Dataset Generator)
Quy mô dữ liệu: **500.000 đến 1.000.000 tin nhắn**.
Dữ liệu được sinh bằng script Go chuyên dụng, mô phỏng chính xác phân phối tin nhắn chat thực tế:
* **70% Tin nhắn ngắn (1 – 10 từ)**: Chào hỏi, xác nhận, từ cảm thán, emoji ("ok bạn", "chúc mừng sinh nhật", "hẹn gặp lại ở quán cà phê nhé").
* **20% Tin nhắn trung bình (10 – 40 từ)**: Thảo luận công việc, trao đổi dự án, có xen lẫn số liệu, từ ghép chuyên ngành ("Hôm nay nhóm mình họp lúc 9 giờ sáng để bàn về tiến độ tích hợp Cốc Cốc Tokenizer vào OpenSearch").
* **10% Tin nhắn dài (40 – 150 từ)**: Biên bản họp, thông báo nội bộ, văn bản trích dẫn ("Kế hoạch triển khai hệ thống tìm kiếm tin nhắn cho toàn bộ người dùng trong quý 4 bao gồm các giai đoạn chuẩn bị hạ tầng, đánh giá hiệu năng...").
* **Đặc tính ngôn ngữ**: Chứa 60% từ ghép tiếng Việt chuẩn, 20% tiếng Việt không dấu (teencode), 10% chứa từ mượn tiếng Anh, 10% chứa ký tự đặc biệt, URL, email.

### 4.2. Ma trận Tham số Benchmark (Testing Matrix)

| Test Suite | Khối lượng Docs | Batch Size (Docs/Request) | Mức Concurrency (Workers) | Thời gian chạy ước tính |
| :--- | :--- | :--- | :--- | :--- |
| **Suite 1: Baseline Single-thread** | 100.000 | 500, 1.000 | 1 worker | 3 – 5 phút |
| **Suite 2: Concurrency Scaling** | 500.000 | 1.000 | 4, 8, 16 workers | 10 – 15 phút |
| **Suite 3: Heavy Stress & Saturation**| 1.000.000 | 1.000, 2.000 | 32 workers | 25 – 35 phút |
| **Suite 4: Duration Stability Test** | Continuous | 1.000 | 16 workers | 30 phút liên tục |

### 4.3. Kiến trúc Công cụ Đo kiểm (Go Benchmark Client Engine)
Xây dựng một công cụ command-line chuyên dụng viết bằng Go:
* Khởi chạy $N$ goroutines độc lập (tương ứng với số concurrency).
* Mỗi goroutine sở hữu một HTTP client tái sử dụng connection pool (`keep-alive`, `MaxIdleConnsPerHost = 100`).
* Client hỗ trợ 2 chế độ nạp:
  * **Mode A (Go CGO Pre-process)**: Goroutine gọi Cốc Cốc CGO để tách từ câu văn bản, định dạng JSON tài liệu theo mapping và gửi Bulk API lên `http://localhost:9201`.
  * **Mode B (OpenSearch Plugin Raw)**: Goroutine chỉ đóng gói văn bản thô vào JSON và gửi Bulk API lên `http://localhost:9202`.
* Ghi nhận chính xác timestamp bắt đầu và kết thúc của từng request với độ chính xác microsecond ($\mu s$).

---

## 5. BỘ CHỈ SỐ ĐO LƯỜNG & PHƯƠNG PHÁP THU THẬP METRICS

### 5.1. Bảng chỉ số Hiệu năng & Tài nguyên cần Thu thập

```
+-----------------------------------------------------------------------------------+
|                            METRICS COLLECTION SUITE                               |
+------------------------------------+----------------------------------------------+
| 1. INDEXING THROUGHPUT & LATENCY   | 2. OPENSEARCH CONTAINER RESOURCES            |
| - Docs Per Second (docs/s)         | - CPU Usage (%) & Throttling (cpu.stat)      |
| - Data Ingestion Rate (MB/s)       | - Memory RSS & Working Set (bytes)           |
| - Request Latency: p50, p90, p95,  | - JVM Heap Used vs Max Heap                  |
|   p99, Max (ms)                    | - JVM Garbage Collection:                    |
| - Bulk Rejections & Failures       |   * Young Gen GC Count & Total Pause Time    |
|                                    |   * Old Gen GC Count & Total Pause Time      |
|                                    | - Thread Pool: bulk queue & active threads   |
+------------------------------------+----------------------------------------------+
| 3. CLIENT APPLICATION (GO)         | 4. SEARCH PARITY & POST-INDEXING             |
| - Client CPU Usage (%)             | - Total Indexed Documents Count              |
| - Client Memory Allocation (B/op)  | - Storage Size on Disk (MB)                  |
| - CGO Call Latency (ns/op)         | - Segment Count & Merge Time                 |
|                                    | - Query Latency (Warm vs Cold)               |
+------------------------------------+----------------------------------------------+
```

### 5.2. Công cụ & Phương pháp Ghi nhận Metrics
1. **Metrics Engine (Prometheus + Go Exporter)**:
   * Thu thập mỗi 1 giây một lần từ OpenSearch REST API:
     * `GET /_nodes/stats/jvm,process,thread_pool,indices/bulk`
   * Thu thập cgroup metrics trực tiếp từ Docker container:
     * `/sys/fs/cgroup/cpu/docker/<id>/cpuacct.usage`
     * `/sys/fs/cgroup/memory/docker/<id>/memory.usage_in_bytes`
2. **Xuất báo cáo tự động (Artifact Export)**:
   * Dữ liệu raw được xuất ra file CSV / JSON lines: `timestamp, docs_indexed, latency_ms, cpu_percent, memory_mb, jvm_heap_mb, gc_pause_ms`.
   * Tạo script vẽ đồ thị trực quan (matplotlib / Chart.js) so sánh đường biểu diễn hiệu năng của 2 mô hình theo thời gian thực.

---

## 6. TIÊU CHÍ ĐÁNH GIÁ, TRADE-OFFS & KHUYẾN NGHỊ KIẾN TRÚC

Bản báo cáo đánh giá cuối cùng sẽ phân tích sâu sắc các khía cạnh kỹ thuật sau:

1. **Hiệu suất Thực tế (Throughput vs Latency)**:
   * Mô hình nào đạt thông lượng cao hơn khi số lượng luồng tăng lên?
   * Khi nào OpenSearch node bắt đầu xuất hiện tình trạng `EsRejectedExecutionException` (nghẽn hàng đợi bulk)?
2. **Cân đối Tải Tính toán (Compute Offloading vs Engine Saturation)**:
   * Việc đẩy tác vụ tách từ sang Go Backend có giúp giảm tải đáng kể cho OpenSearch Cluster, cho phép OpenSearch tập trung toàn bộ I/O và CPU vào việc ghi segment Lucene và merge hay không?
   * Đánh giá chi phí CPU bỏ ra ở Go Client (CGO) so với chi phí JNI C++ chạy trực tiếp bên trong JVM của OpenSearch.
3. **Độ ổn định & Cô lập Rủi ro (Fault Tolerance & Stability)**:
   * Rủi ro khi C++ native tokenizer gặp lỗi segmentation fault (segfault) hoặc memory leak:
     * Ở Mô hình B: Nếu native lib bị crash, **toàn bộ process OpenSearch JVM sẽ sập (Crash Cluster Node)**, ảnh hưởng đến toàn bộ dữ liệu và các dịch vụ khác.
     * Ở Mô hình A: Nếu native lib gặp sự cố, chỉ có worker pod của Go bị crash và Kubernetes có thể restart tức thì mà không ảnh hưởng tới OpenSearch cluster.
4. **Khả năng Mở rộng Ngang (Horizontal Scalability)**:
   * Việc mở rộng Go App Servers (stateless pods) dễ dàng và tiết kiệm chi phí gấp nhiều lần so với việc thêm node cho OpenSearch Cluster (stateful data nodes).
5. **Tính linh hoạt & Nâng cấp Hệ thống (Maintainability & Upgradability)**:
   * Plugin `opensearch-analysis-vietnamese` bị gắn cứng (`hardcode`) với exact OpenSearch version trong file `plugin-descriptor.properties`. Mỗi khi nâng cấp OpenSearch (ví dụ từ 2.19 lên 2.20), bắt buộc phải recompile và kiểm thử lại plugin.
   * Mô hình Go CGO cho phép nâng cấp OpenSearch hoàn toàn độc lập, không phụ thuộc vào vòng đời của plugin bên thứ ba.

---

## 7. KẾ HOẠCH TRIỂN KHAI THEO TỪNG GIAI ĐOẠN (ROADMAP)

| Giai đoạn | Nội dung công việc cụ thể | Thời gian dự kiến | Sản phẩm đầu ra (Deliverable) |
| :--- | :--- | :--- | :--- |
| **Phase 1** | - Khởi tạo Docker Compose cho 2 cluster OpenSearch 2.19.<br>- Thiết lập chính xác index mappings và analyzers.<br>- Viết test script kiểm tra Search Parity và đối soát 100% token output. | Ngày 1 | `docker-compose.benchmark.yml`, Mappings JSON, Parity Test Report |
| **Phase 2** | - Xây dựng bộ sinh dữ liệu 500k - 1M documents tiếng Việt chuẩn chat.<br>- Viết công cụ benchmark đa luồng bằng Go (hỗ trợ cả 2 mode A và B).<br>- Tích hợp collector đo CPU, RAM, GC pauses, Threadpool. | Ngày 2 | `generator.go`, `benchmark_runner.go`, Metrics Collector Scripts |
| **Phase 3** | - Thực thi toàn bộ Ma trận Benchmark (Single thread, 4, 8, 16, 32 workers).<br>- Chạy stress-test 30 phút liên tục ghi nhận dữ liệu steady-state.<br>- Ghi nhận metrics chi tiết ra CSV/JSON. | Ngày 3 | Bộ raw data benchmark CSV/JSON của 2 mô hình |
| **Phase 4** | - Tổng hợp số liệu, vẽ biểu đồ so sánh Throughput, Latency, CPU, RAM.<br>- Phân tích chi tiết ưu nhược điểm (Trade-offs) và rủi ro vận hành.<br>- Hoàn thiện Bản báo cáo đánh giá hiệu năng chính thức (Markdown/PDF). | Ngày 4 | File `docs/benchmark_opensearch_plugin_report.md` hoàn chỉnh |
