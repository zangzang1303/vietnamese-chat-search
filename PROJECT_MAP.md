# BẢN ĐỒ ĐIỀU HƯỚNG DỰ ÁN (PROJECT REPOSITORY MAP)
## HỆ THỐNG PHÂN LOẠI: BÁO CÁO GIẢI THÍCH, BẰNG CHỨNG CHẠY & SƠ ĐỒ TRỰC QUAN

Tài liệu này đóng vai trò là **Bản đồ chỉ mục trung tâm (Master Index)** giúp kỹ sư, giảng viên hoặc AI Agent dễ dàng định vị chính xác:
1. 📘 **Đâu là Báo cáo giải thích & Tài liệu kỹ thuật**
2. 🔬 **Đâu là Bằng chứng chạy thực tế & Dữ liệu máy (Raw Evidence)**
3. 📊 **Đâu là Sơ đồ kiến trúc & Biểu đồ thực nghiệm**
4. ⚙️ **Đâu là Mã nguồn & Kịch bản thực thi**

---

```
                                      DỰ ÁN: VIETNAMESE CHAT SEARCH
                                                    │
         ┌──────────────────────────┬───────────────┴───────────────┬──────────────────────────┐
         ▼                          ▼                               ▼                          ▼
 📘 1. BÁO CÁO & GIẢI THÍCH   🔬 2. BẰNG CHỨNG CHẠY        📊 3. SƠ ĐỒ & BIỂU ĐỒ        ⚙️ 4. MÃ NGUỒN & RUNNERS
 (Thư mục: docs/)            (Thư mục: benchmarks/)       (Thư mục: image/, charts)   (cmd/, pkg/, benchmarks/)
 ├─ Task 1 Report            ├─ 50 Raw Runs JSON          ├─ 9 Benchmark Charts       ├─ Go Tokenizer Module
 ├─ Task 2 Report            ├─ Telemetry 1s cgroups      ├─ Architecture Diagrams    ├─ Matrix Orchestrator
 ├─ Giải thích Biểu đồ       ├─ Go Mutex pprof            └─ Write Flow Sequence      └─ Bench Runner (Linux)
 └─ Kịch bản thuyết trình    └─ Soak & Batch Summaries
```

---

## 📘 KHU VỰC 1: BÁO CÁO GIẢI THÍCH & PHÂN TÍCH KỸ THUẬT (REPORTS & DOCS)

Toàn bộ các tài liệu phân tích sâu, báo cáo khoa học và kịch bản thuyết trình nằm tại thư mục `docs/`:

| Tên Tài Liệu | Đường Dẫn Tệp | Nội Dung & Mục Đích |
| :--- | :--- | :--- |
| **Báo cáo Hoàn tất Task 1 (Master)** | [docs/benchmark_opensearch_plugin_report.md](docs/benchmark_opensearch_plugin_report.md) | **Báo cáo cốt lõi của Task 1:** 9 chương phân tích hiệu năng OpenSearch Vanilla vs Plugin, thống kê bắt cặp 50 runs, bằng chứng pprof mutex, engine-only 2.7x–3.7x, batch sensitivity, 30-min soak test, search latency và bảng tiêu chí DoD 12 gates. |
| **Báo cáo Hoàn tất Task 2** | [docs/task2_coccoc_tokenizer_go_module_report.md](docs/task2_coccoc_tokenizer_go_module_report.md) | **Báo cáo cốt lõi của Task 2:** Thiết kế module `coccoc-tokenizer-go` độc lập, kiến trúc Dual-Mode (CGO C++ ~1.2µs & Pure-Go fallback ~45µs), kiểm thử an toàn bộ nhớ 0 leak, API idiomatic và CI/CD. |
| **Báo cáo Giải thích Biểu đồ** | [docs/bao_cao_giai_thich_bieu_do_thuc_nghiem.md](docs/bao_cao_giai_thich_bieu_do_thuc_nghiem.md) | Phân tích chi tiết từng mốc dữ liệu trên các đồ thị benchmark: tại sao C=1 đến C=4 Model A thắng, tại sao C=16 Model B vượt lên, phân tích điểm nghẽn Go mutex và chi phí GC. |
| **Kiến trúc Tìm kiếm Hệ thống** | [docs/search_architecture.md](docs/search_architecture.md) | Thiết kế kiến trúc tổng thể của hệ thống chat search tiếng Việt, luồng index, query pipeline và analyzer settings. |
| **Đặc tả Engine Lưu trữ Tùy biến** | [docs/custom_storage_engine_spec.md](docs/custom_storage_engine_spec.md) | Đặc tả kỹ thuật về cấu trúc docstore, index file và cơ chế tối ưu I/O. |
| **Luồng Ghi & Cơ chế Lưu trữ** | [docs/write_flow_and_storage_internals.md](docs/write_flow_and_storage_internals.md) | Giải thích chi tiết cơ chế hoạt động của commit, flush, write-ahead log (translog) và segment merge trong Lucene. |
| **Kịch bản Thuyết trình Slide** | [docs/kich_ban_thuyet_trinh_tung_slide.md](docs/kich_ban_thuyet_trinh_tung_slide.md) & [docs/presentation_10_slides.md](docs/presentation_10_slides.md) | Dàn ý và lời thoại thuyết trình 10 slides báo cáo kết quả trước hội đồng/team. |
| **Hướng dẫn Tích hợp CGO** | [docs/coccoc_tokenizer_and_cgo_guide.md](docs/coccoc_tokenizer_and_cgo_guide.md) | Hướng dẫn cấu hình CGO, quản lý bộ nhớ native C++ và đóng gói Docker container. |

---

## 🔬 KHU VỰC 2: BẰNG CHỨNG CHẠY THỰC TẾ & DỮ LIỆU MÁY (RUN EVIDENCE & RAW DATA)

Toàn bộ dữ liệu đo kiểm thực nghiệm nguyên bản (không qua chỉnh sửa thủ công) nằm tại thư mục `benchmarks/final_run/final_20261020_093100/`:

### 1. Dữ liệu Ma trận 50 Run Chính (1.000.000 Docs/Run, 5 Repetitions)
* **50 Tệp Kết Quả Raw JSON:** Nằm tại thư mục `benchmarks/final_run/final_20261020_093100/raw/`
  * Cấu trúc tên: `rep{R}_{vanilla|plugin}_c{C}.json` (với $R \in \{1..5\}$, $C \in \{1, 4, 8, 16, 32\}$).
  * Chứa đầy đủ: wall time, throughput docs/s, latency percentiles (p50, p90, p95, p99), HTTP status, bulk errors.
* **50 Tệp Telemetry Mỗi Giây:** Nằm tại thư mục `benchmarks/final_run/final_20261020_093100/telemetry/`
  * Cấu trúc tên: `rep{R}_{vanilla|plugin}_c{C}_telemetry.json`.
  * Chứa: JVM Heap used, young/old GC count & time, write queue, write rejected, indexing operations rate.
* **Bảng Tổng Hợp Thống Kê & Kiểm Định:**
  * [summaries/aggregate.json](benchmarks/final_run/final_20261020_093100/summaries/aggregate.json) & [aggregate.csv](benchmarks/final_run/final_20261020_093100/summaries/aggregate.csv): Thống kê Median, Mean, CI 95% cho từng cấu hình.
  * [summaries/statistical_analysis.json](benchmarks/final_run/final_20261020_093100/summaries/statistical_analysis.json) & [paired_throughput.csv](benchmarks/final_run/final_20261020_093100/summaries/paired_throughput.csv): **Bằng chứng phân tích bắt cặp (Paired difference $A - B$)**, chứng minh 95% CI đều chứa 0.

### 2. Dữ liệu Thực Nghiệm Bổ Sung (Supplement Evidence)
Nằm tại thư mục `benchmarks/final_run/final_20261020_093100/supplement/`:
* **Hồ Sơ Xung Đột Khóa Mutex (Go Mutex Profile):**
  * File pprof nhị phân: [tokenizer_profile/mutex.pprof](benchmarks/final_run/final_20261020_093100/supplement/tokenizer_profile/mutex.pprof) — *Chứng minh 100% (273,5s) thời gian nghẽn tại `sync.(*Mutex).Unlock`.*
  * Số liệu scaling CGO: [tokenizer_profile/tokenizer_scaling_profile.json](benchmarks/final_run/final_20261020_093100/supplement/tokenizer_profile/tokenizer_scaling_profile.json).
* **Kết Quả Đo Tách Riêng Engine (Engine-Only Suite):**
  * [summaries/engine_only_summary.json](benchmarks/final_run/final_20261020_093100/supplement/summaries/engine_only_summary.json) & `.csv` — *Chứng minh Vanilla ghi index nhanh gấp 2.7x – 3.7x khi dữ liệu đã băm sẵn.*
* **Kết Quả Đo Độ Nhạy Kích Thước Lô (Batch Sensitivity):**
  * [summaries/batch_sensitivity_summary.json](benchmarks/final_run/final_20261020_093100/supplement/summaries/batch_sensitivity_summary.json) & `.csv` — *Bằng chứng 20 batches bị client timeout ở Model B tại C=32, batch 2000.*
* **Kết Quả Kiểm Thử Độ Bền Tải 30 Phút (Soak Test):**
  * [summaries/soak_summary.json](benchmarks/final_run/final_20261020_093100/supplement/summaries/soak_summary.json) & `.csv` — *Ghi nhận nạp >28 triệu ops, độ dốc bộ nhớ -1.98 vs +1.09 MiB/min, 0 crash.*
* **Kết Quả Độ Trễ Tìm Kiếm (Search Latency):**
  * [summaries/search_latency_summary.json](benchmarks/final_run/final_20261020_093100/supplement/summaries/search_latency_summary.json) & `.csv` — *1.000 queries tại C=1, 8, 32; p50 từ 4.9ms đến 28.8ms, 0% lỗi.*
* **Kết Quả Đối Soát Tính Tương Đồng (Parity):**
  * [parity/full_2000/token_parity_summary.json](benchmarks/final_run/final_20261020_093100/parity/full_2000/token_parity_summary.json) — *Chi tiết 358 lỗi lệch / 8.000 so sánh (95.525% parity).*
  * [search/search_parity_summary.json](benchmarks/final_run/final_20261020_093100/search/search_parity_summary.json) — *Đối soát 1.000 queries trên 5 trường dữ liệu.*
* **Thông Tin Môi Trường & Checksums:**
  * [manifest.json](benchmarks/final_run/final_20261020_093100/manifest.json) — *Ghi nhận commit SHA, Docker 29.1.3, Temurin 21, cpus=4.0, mem=3GiB, SHA-256 dataset & dicts.*

---

## 📊 KHU VỰC 3: SƠ ĐỒ KIẾN TRÚC & BIỂU ĐỒ TRỰC QUAN (DIAGRAMS & CHARTS)

Toàn bộ đồ thị trực quan hóa và sơ đồ luồng hoạt động nằm tại thư mục `image/` và `benchmark_charts/`:

### 1. Hệ Thống 9 Biểu Đồ Thực Nghiệm Chuẩn Hóa (Benchmark Charts)
Được sinh tự động bằng `benchmarks/generate_final_charts.py` từ dữ liệu thực:

| Tên Biểu Đồ | File Ảnh trong `image/` | Ý Nghĩa Trực Quan |
| :--- | :--- | :--- |
| **1. Throughput Scaling Curve** | [image/benchmark_throughput_scaling.png](image/benchmark_throughput_scaling.png) | Đường cong thông lượng docs/s từ C=1 đến C=32 kèm khoảng tin cậy Bootstrap 95% CI quanh Median. |
| **2. Paired Differences ($A - B$)** | [image/benchmark_paired_throughput_difference.png](image/benchmark_paired_throughput_difference.png) | Chênh lệch thông lượng bắt cặp từng repetition với khoảng tin cậy quanh điểm 0 (bác bỏ overclaim winner). |
| **3. Latency Distribution (C=8)** | [image/benchmark_latency_distribution.png](image/benchmark_latency_distribution.png) | Phân bố độ trễ request (p50, p90, p95, p99, max) tại mức tải chuẩn C=8. |
| **4. JVM GC Collection Stats** | [image/benchmark_jvm_gc_stats.png](image/benchmark_jvm_gc_stats.png) | So sánh thời gian tiến trình GC làm việc (Model B cao hơn từ +19% đến +40%). |
| **5. Storage & Segments** | [image/benchmark_storage_segments.png](image/benchmark_storage_segments.png) | Dung lượng Lucene index trên đĩa (550MB – 640MB) và số lượng segment files. |
| **6. Batch Sensitivity Matrix** | [image/benchmark_batch_sensitivity.png](image/benchmark_batch_sensitivity.png) | So sánh thông lượng và độ trễ ở các kích thước lô: batch 500, 1000, 2000. |
| **7. Cost Isolation & Lock Contention** | [image/benchmark_cost_isolation.png](image/benchmark_cost_isolation.png) | Panel đôi: Ingestion thuần túy của engine (Vanilla 2.7x–3.7x) & Độ sụt giảm parallel efficiency của tokenizer CGO. |
| **8. Search Latency vs Concurrency** | [image/benchmark_search_latency.png](image/benchmark_search_latency.png) | Độ trễ tìm kiếm p50/p95/p99 và thông lượng QPS trên 1.000 queries tại C=1, 8, 32. |
| **9. 30-Minute Soak Timeline** | [image/benchmark_soak_timeline.png](image/benchmark_soak_timeline.png) | Diễn biến Working Set RSS và JVM Heap qua các cửa sổ 5 phút trong suốt 30 phút soak test. |

### 2. Hệ Thống Sơ Đồ Kiến Trúc & Luồng Hoạt Động (Architecture & Sequence Diagrams)

| Tên Sơ Đồ | File Ảnh trong `image/` | Mô Tả Kỹ Thuật |
| :--- | :--- | :--- |
| **Luồng Ghi Dữ Liệu Tuần Tự (Write Sequence)** | [image/The Life of a Write Sequence Diagram.png](image/The%20Life%20of%20a%20Write%20Sequence%20Diagram.png) | Sơ đồ tương tác tuần tự từ lúc Client gửi `_bulk` qua OpenSearch Node, Memory Buffer, Translog đến Lucene Segment. |
| **Kiến Trúc Tổng Thể Hệ Thống (Architecture)** | [image/Hình 2.1.png](image/Hình%202.1.png) & [image/Hình 3.1.png](image/Hình%203.1.png) | Mô hình các tầng: Client App, Messaging Queue, Worker, Storage Engine và Search Cluster. |
| **Mô Hình Bộ Tách Từ Cốc Cốc (Tokenizer Model)** | [image/Cốc Cốc Search Engine-2026-09-18-072911.png](image/Cốc%20Cốc%20Search%20Engine-2026-09-18-072911.png) | Cấu trúc Double-Array Trie (DATrie) và giải thuật phân đoạn từ vựng tiếng Việt. |
| **So Sánh Điểm Số Tìm Kiếm (BM25 vs Compound)** | [image/score_comparison_chart.jpg](image/score_comparison_chart.jpg) & [image/compound_word_trap_demo.jpg](image/compound_word_trap_demo.jpg) | Minh họa bẫy từ ghép tiếng Việt (ví dụ: "cà phê", "học sinh") khi không có bộ tách từ chuẩn. |

---

## ⚙️ KHU VỰC 4: MÃ NGUỒN & KỊCH BẢN THỰC THI (CODE & SCRIPTS)

| Thành Phần | Thư Mục / Tệp Mã Nguồn | Chức Năng |
| :--- | :--- | :--- |
| **Go Tokenizer Module (Task 2)** | `coccoc-tokenizer-go/` | Package Go độc lập `github.com/vntk/coccoc-tokenizer-go` (Dual-mode CGO C++ & Pure-Go fallback). |
| **Orchestrator Ma Trận 50 Runs** | [benchmarks/run_final_matrix_orchestrator.py](benchmarks/run_final_matrix_orchestrator.py) | Điều phối chạy tự động 50 runs AB/BA, quản lý warmup và container lifecycle. |
| **Supplement Runner** | [benchmarks/run_completion_supplement.py](benchmarks/run_completion_supplement.py) | Điều phối chạy Batch sensitivity, Engine-only, Search latency và Soak test 30 phút. |
| **Runner Benchmark Tiền Xử Lý (CGO)** | [benchmarks/bench_runner.go](benchmarks/bench_runner.go) | Mã nguồn Go đo tốc độ nạp dữ liệu, hỗ trợ cả `vanilla`, `plugin` và `vanilla_pretokenized`. |
| **Công Cụ Profiling Go Mutex** | [benchmarks/tokenizer_scaling_profile.go](benchmarks/tokenizer_scaling_profile.go) | Chương trình đo scaling CGO tokenizer và xuất file `mutex.pprof`. |
| **Công Cụ Đo Kiểm Tính Tương Đồng** | [benchmarks/run_final_parity.py](benchmarks/run_final_parity.py) & [benchmarks/run_final_search_parity.py](benchmarks/run_final_search_parity.py) | Đo kiểm Token Parity trên 2.000 câu và Search Parity trên 1.000 queries. |
| **Kịch Bản Sinh Biểu Đồ** | [benchmarks/generate_final_charts.py](benchmarks/generate_final_charts.py) | Script đọc các summary JSON để vẽ 9 biểu đồ PNG. |
| **Docker Compose Benchmark** | [benchmarks/docker-compose.benchmark.yml](benchmarks/docker-compose.benchmark.yml) | Cấu hình giới hạn container: 4 vCPU, 3 GiB RAM, JVM Heap 1536m. |
