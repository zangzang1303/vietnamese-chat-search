# AGENT PROMPT: BÁO CÁO KẾT QUẢ & TỔNG HỢP QUÁ TRÌNH THỰC HIỆN TASK 1
## “Đánh giá ảnh hưởng của Plugin tiếng Việt trên OpenSearch 2.19 — Benchmark, Resource Consumption & Search Parity”

Bạn là Senior Performance & Search Engine Engineer. Hãy tiếp quản workspace này để nắm bắt toàn bộ kết quả thực nghiệm, phương pháp luận khoa học, cấu trúc artifacts và các kết luận kỹ thuật đã được công nghiệm đầy đủ cho **Task 1**.

---

## 1. VỊ TRÍ WORKSPACE & CÁC TÀI LIỆU BÁO CÁO CỐT LÕI

* **Thư mục dự án:** `D:\CODE\VSF\vietnamese-chat-search`
* **Môi trường thực thi:** WSL2 Ubuntu 24.04 LTS, Docker 29.1.3 (cgroups v2, cố định 4 vCPU, 3 GiB RAM, JVM Heap 1536 MiB cho mỗi container).
* **Báo cáo hoàn chỉnh chính thức:**
  * [docs/benchmark_opensearch_plugin_report.md](../benchmark_opensearch_plugin_report.md) — *Bản báo cáo khoa học 9 chương đầy đủ số liệu, bảng biểu, phân tích thống kê bắt cặp, bằng chứng pprof, DoD PASS/FAIL và khuyến nghị kiến trúc.*
* **Kế hoạch & Đề bài gốc:**
  * [docs/new plan/plan_task1_opensearch_plugin_benchmark.md](../new%20plan/plan_task1_opensearch_plugin_benchmark.md)
  * [docs/prompts/task1_final_completion_agent_prompt.md](task1_final_completion_agent_prompt.md)
* **Thư mục Artifacts hoàn chỉnh:**
  * Run chính (50 runs $\times$ 1M docs): [benchmarks/final_run/final_20261020_093100/](../../benchmarks/final_run/final_20261020_093100)
  * Run bổ sung (Telemetry, Profiling, Soak, Batch, Search): [benchmarks/final_run/final_20261020_093100/supplement/](../../benchmarks/final_run/final_20261020_093100/supplement)

---

## 2. TỔNG HỢP TOÀN BỘ QUÁ TRÌNH ĐÃ THỰC HIỆN (PHASES A – H)

Quá trình thực nghiệm được triển khai tuần tự, nghiêm ngặt và kiểm chứng bằng dữ liệu máy (empirical data-driven):

### Phase A: Kiểm toán Hệ thống & Smoke Test Môi trường
* Kiểm tra hạn ngạch container bằng `docker inspect`: Cố định chính xác `4.0 vCPU` (`NanoCPUs=4000000000`, `CpuQuota=400000/100000`), `Memory=3 GiB`, `JVM Heap=1536 MiB`.
* Xác thực các phiên bản: Docker `29.1.3`, OpenSearch `2.19.0`, OpenJDK `21.0.6 Temurin`, Go `1.26.0 linux/amd64`.
* Xác nhận tính toàn vẹn: SHA-256 dataset `dataset_real_1m.jsonl` (`b5debf35...`, 367.8 MB), từ điển trích xuất từ plugin `coccoc_plugin_final`.
* Ghi nhận trung thực xuất xứ: Plugin JAR và `libcoccoc_tokenizer_jni.so` có commit provenance là `unknown` (được build sẵn từ Docker image ban đầu).
* Hiệu chỉnh thanh đo tiến trình của `benchmarks/bench_runner.go`: Báo cáo chính xác số `operations` thay vì `unique docs` khi chạy ghi đè ID trong soak test.

### Phase B: Tái Phân Tích Thống Kê Bắt Cặp (50-Run Matrix)
* Thực thi script kiểm định: [benchmarks/analyze_final_statistics.py](../../benchmarks/analyze_final_statistics.py).
* Sử dụng đơn vị thực nghiệm là toàn bộ run (1.000.000 docs/run), thiết kế bắt cặp (paired differences $A - B$ cùng repetition $r \in \{1..5\}$, seed 42).
* Áp dụng song song **Student-t 95% CI** và **Percentile Bootstrap 95% CI quanh Median**.
* **Phát hiện thống kê then chốt:** Do phương sai cao ($CV\% = 13.5\% - 36.8\%$), khoảng tin cậy 95% CI của hiệu số thông lượng $(A - B)$ **đều chứa điểm 0 tại mọi mức concurrency $C \in \{1, 4, 8, 16, 32\}$** (p-value luôn $> 0.05$). Chênh lệch median chỉ mang tính mô tả thực nghiệm, không có ý nghĩa thống kê áp đảo.

### Phase C & D: Resource Telemetry & Đo Kiểm Độ Nhạy Lô (Batch Sensitivity)
* Thu thập chuỗi thời gian telemetry 1 giây đầy đủ: CPU container, Memory RSS, cgroups v2 CPU throttling (`throttled_usec`), JVM Young/Old GC, OpenSearch write queue/rejections, Go client CPU (user/sys qua `getrusage`) và Peak RSS.
* Đo kiểm ma trận 24 runs (3 lần lặp/cấu hình, AB/BA order) gồm:
  * Baseline: $C=1$, 100k docs, batch 500 & 1000.
  * Heavy Stress: $C=32$, 1M docs, batch 1000 & 2000.
* **Phát hiện bão hòa tài nguyên:** Ở $C=32$, batch 2000, Mô hình B gặp hiện tượng nghẽn nặng tại JNI CPU & GC trong OpenSearch engine, dẫn đến **20 bulk batches bị client timeout (60s)** và rớt dữ liệu, trong khi Mô hình A duy trì ổn định không rớt batch nào.

### Phase E: Kiểm Thử Độ Bền Tải Liên Tục 30 Phút (Soak Test)
* Chạy liên tục trọn vẹn **1.800 giây mỗi mô hình** tại $C=16$, batch 1000, ghi nhận telemetry mỗi 1 giây trong [soak_summary.json](../../benchmarks/final_run/final_20261020_093100/supplement/summaries/soak_summary.json):
  * **Mô hình A (Vanilla + CGO):** Nạp **13.629.000 ops** trong 1.810,6s (throughput trung bình **7.565,9 ops/s**). Độ dốc bộ nhớ RSS: **-1,98 MiB/phút** (ổn định). 0 crash, 0 OOMKilled.
  * **Mô hình B (Plugin JNI):** Nạp **14.487.000 ops** trong 1.813,9s (throughput trung bình **8.043,1 ops/s**). Độ dốc bộ nhớ RSS: **+1,09 MiB/phút** (tăng nhẹ không rò rỉ). 0 crash, 0 OOMKilled.
  * **Áp lực GC:** Tổng thời gian JVM Young GC của Mô hình B cao hơn Mô hình A **+19.4%** (35.256 ms vs 29.531 ms).

### Phase F: Tách Biệt Chi Phí (Cost Isolation) & Chứng Minh Điểm Nghẽn Khóa Mutex
* **Go Mutex Profiling (pprof):** Đo kiểm scaling CGO thuần túy với [benchmarks/tokenizer_scaling_profile.go](../../benchmarks/tokenizer_scaling_profile.go) và thu thập `mutex.pprof`.
  * *Bằng chứng thực nghiệm:* **273,50 giây (100% tổng thời gian chờ khóa)** tập trung chính xác tại `sync.(*Mutex).Unlock` trong hàm `coccoc.Tokenizer.SegmentOriginal`.
  * Hiệu suất song song (parallel efficiency) tụt dốc từ **100%** ($C=1$: 75,5k ops/s) xuống còn **2,5%** ($C=32$: 60,8k ops/s). Điểm nghẽn Model A ở tải cao đã được chứng minh bằng định lượng, không còn là giả thuyết.
* **Engine-Only Suite:** Nạp 100k docs ở C=8 và C=16 với payload đã tokenize sẵn trong RAM (ngoài vùng timing) cho Model A vs raw text cho Model B:
  * Tại $C=8$: Vanilla đạt **13.565,6 docs/s** vs Plugin **4.968,1 docs/s** (**Vanilla nhanh gấp 2,73 lần**).
  * Tại $C=16$: Vanilla đạt **13.909,1 docs/s** vs Plugin **3.723,7 docs/s** (**Vanilla nhanh gấp 3,74 lần**).
  * *Ý nghĩa:* Khi loại bỏ hoàn toàn chi phí tokenization ở client, năng lực ghi index Lucene thuần túy của OpenSearch Vanilla nhanh hơn từ 2.7x đến 3.7x.

### Phase G: Đo Kiểm Độ Trễ Tìm Kiếm Thực Tế (Search Latency Scaling)
* Đo kiểm 1.000 câu truy vấn độc lập tại $C \in \{1, 8, 32\}$:
  * Độ trễ p50: ~4,97 ms vs 5,01 ms ($C=1$); ~8,71 ms vs 8,83 ms ($C=8$); ~27,37 ms vs 28,83 ms ($C=32$).
  * Thông lượng tối đa: Đạt ~990 QPS tại C=32.
  * Tỷ lệ lỗi truy vấn: **0,0%** cho cả hai mô hình.

### Phase H: Sinh Hệ Thống 9 Biểu Đồ Chuẩn Hóa & Báo Cáo Cuối
* Cập nhật script [benchmarks/generate_final_charts.py](../../benchmarks/generate_final_charts.py) sinh tự động toàn bộ 9 biểu đồ thực nghiệm và đồng bộ vào `image/`, `benchmark_charts/`, và `supplement/charts/`.
* Hoàn thiện bản báo cáo khoa học [docs/benchmark_opensearch_plugin_report.md](../benchmark_opensearch_plugin_report.md) chuẩn hóa thuật ngữ `GC collection time`, link tương đối portable và bảng DoD 12 gates.

---

## 3. CÁC PHÁT HIỆN KỸ THUẬT QUAN TRỌNG NHẤT (CORE EMPIRICAL FINDINGS)

1. **Cổng Zero-Divergence Parity bị ĐÁNH DẤU FAIL (Trung thực, không che giấu):**
   * Token Parity đạt **95,525%** (358 trường hợp lệch trên 8.000 phép so sánh).
   * Lệch do khác biệt quy tắc Unicode UAX#29 của Lucene `StandardTokenizer` (bỏ `%`, giữ nháy đơn `'`, tách dấu chấm/phẩy số học) so với luật C++ native Cốc Cốc (giữ `95%`, bẻ đôi nháy đơn).
   * Search Parity: 100% trên `text`, 94,8%–95,6% Top-10 trên các trường tiếng Việt, và chỉ **7,1% trên `edge_prefix`**.
2. **Bằng Chứng Mutex Contention:**
   * Thủ phạm khiến Model A bị chững throughput ở tải cao là cơ chế Singleton có `sync.Mutex` ở module Go client (100% thời gian khóa delay đo bằng pprof).
3. **Ưu Thế Ingestion của OpenSearch Vanilla:**
   * OpenSearch Vanilla thuần túy ghi index nhanh hơn gấp **2.7x – 3.7x** và giảm thời gian GC từ **19% đến 40%** do không phải thực thi JNI boundary crossing và cấp phát hàng triệu đối tượng Java `Token` trong Heap.
4. **Độ Tin Cậy Thực Tế:**
   * Ghi nhận **0 lần crash và 0 lần OOMKilled** trong suốt 50 runs chính (50M docs) và 30 phút soak test (>28M ops).

---

## 4. BẢNG TỔNG HỢP ARTIFACTS DỮ LIỆU ĐỂ KIỂM TRA

| Hạng mục | Đường dẫn tệp Artifact |
| :--- | :--- |
| **Phân tích thống kê bắt cặp 50 runs** | `benchmarks/final_run/final_20261020_093100/summaries/statistical_analysis.json` & `paired_throughput.csv` |
| **Hồ sơ xung đột khóa Go Mutex** | `benchmarks/final_run/final_20261020_093100/supplement/tokenizer_profile/mutex.pprof` & `.json` |
| **Kết quả Engine-Only Ingestion** | `benchmarks/final_run/final_20261020_093100/supplement/summaries/engine_only_summary.json` & `.csv` |
| **Kết quả Độ nhạy kích thước lô** | `benchmarks/final_run/final_20261020_093100/supplement/summaries/batch_sensitivity_summary.json` & `.csv` |
| **Kết quả Độ bền 30-phút Soak** | `benchmarks/final_run/final_20261020_093100/supplement/summaries/soak_summary.json` & `.csv` |
| **Kết quả Độ trễ tìm kiếm (Search Latency)**| `benchmarks/final_run/final_20261020_093100/supplement/summaries/search_latency_summary.json` & `.csv` |
| **Kết quả Đối soát Token Parity** | `benchmarks/final_run/final_20261020_093100/parity/full_2000/token_parity_summary.json` |
| **Kết quả Đối soát Search Parity** | `benchmarks/final_run/final_20261020_093100/search/search_parity_summary.json` |
| **Manifest hệ sinh thái & Checksums** | `benchmarks/final_run/final_20261020_093100/manifest.json` |
| **Kế hoạch thực nghiệm bổ sung** | `benchmarks/final_run/final_20261020_093100/supplement/experimental_plan.json` & `README.md` |
| **Thư mục 9 Biểu đồ trực quan hóa** | `image/` và `benchmark_charts/` |

---

## 5. KHUYẾN NGHỊ KIẾN TRÚC CHO AGENT KẾ TIẾP

Khi tư vấn hoặc triển khai cho hệ thống chat tìm kiếm tin nhắn:
1. **Nếu chọn Mô hình A (Khuyến nghị cho Production quy mô lớn):**
   * *Mục tiêu:* Bảo vệ cụm OpenSearch, giảm tải JVM GC, mở rộng ngang (horizontal scale) linh hoạt trên app tier.
   * *Hành động kỹ thuật bắt buộc:* Thay thế cơ chế Singleton `sync.Mutex` bằng **Tokenizer Worker Pool** (mỗi worker goroutine sở hữu riêng một instance Cốc Cốc native) để mở khóa 100% băng thông băm từ ở mức tải cực hạn.
2. **Nếu chọn Mô hình B (In-engine Plugin):**
   * *Mục tiêu:* Đơn giản hóa backend Go (không cần CGO C++), giữ nguyên 100% hành vi phân đoạn từ của Cốc Cốc.
   * *Yêu cầu hạ tầng bắt buộc:* Phải cấp tài nguyên OpenSearch lớn hơn đáng kể (CPU $\ge$ 8 vCPU, JVM Heap $\ge$ 4–8 GiB) và giới hạn kích thước lô bulk ($B \le 1.000$) để tránh hiện tượng client timeout bão hòa JNI đã được chứng minh ở Phase D.
