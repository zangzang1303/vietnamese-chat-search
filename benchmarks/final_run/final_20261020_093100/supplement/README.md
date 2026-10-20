# Supplement Benchmark Artifacts Directory

Tập hợp toàn bộ empirical artifacts bổ sung cho **Task 1: Đánh giá ảnh hưởng của Plugin tiếng Việt trên OpenSearch 2.19**.

## Cấu trúc thư mục

```text
supplement/
├── README.md                     # Tài liệu này (mô tả schema và cấu trúc)
├── experimental_plan.json        # Kế hoạch thực nghiệm bổ sung và tham số cấu hình
├── raw/                          # JSON output từ bench_runner cho từng run
├── telemetry/                    # 1-second telemetry (cgroups v2, docker stats, OS node stats)
├── client/                       # Client resource metrics (getrusage, CPU user/sys, max RSS)
├── engine_only/                  # Kết quả đo tách riêng chi phí indexing của OpenSearch engine
├── tokenizer_profile/            # Kết quả đo scaling tokenizer CGO và file profile mutex contention (pprof)
├── search_latency/               # Kết quả benchmark query latency trên 1.000 queries thực tế
├── summaries/                    # Bảng tổng hợp JSON & CSV cho từng suite
├── charts/                       # Biểu đồ trực quan hóa dữ liệu thực nghiệm
└── logs/                         # Toàn bộ console logs chi tiết của các lượt chạy
```

## Schema tệp trong `summaries/`

### 1. `batch_sensitivity_summary.json` (.csv)
- `model`: `vanilla` (Model A) hoặc `plugin` (Model B).
- `concurrency`: Mức tải đồng thời (`1` hoặc `32`).
- `batch_size`: Kích thước lô (`500`, `1000`, hoặc `2000`).
- `docs`: Số tài liệu index (`100000` hoặc `1000000`).
- `repetitions`: Số lần chạy lặp lại (`3`).
- `throughput_median`: Median throughput đạt được (docs/s).
- `throughput_mean`: Mean throughput (docs/s).
- `latency_p95_median_ms`: Median thời gian phản hồi p95 (ms).
- `container_cpu_avg_pct_median`: Median OpenSearch container CPU usage (%).
- `container_rss_peak_mib_median`: Median RSS đỉnh của OpenSearch container (MiB).
- `container_cgroup_throttled_sec_median`: Median thời gian cgroup CPU bị throttle (giây).
- `client_cpu_sec_median`: Median CPU time tiêu thụ bởi Go client process (giây).
- `client_max_rss_mib_median`: Median peak memory của Go client process (MiB).
- `total_system_cpu_sec_median`: Tổng CPU tiêu thụ toàn hệ thống (Client CPU sec + Container CPU sec).
- `jvm_gc_collection_time_ms_median`: Median tổng thời gian JVM thực hiện Garbage Collection (ms).

### 2. `engine_only_summary.json` (.csv)
- `concurrency`: Mức đồng thời (`8`, `16`).
- `batch_size`: `1000`.
- `docs`: `100000`.
- `vanilla_pretokenized_throughput_median`: Throughput của OpenSearch Vanilla khi nhận payload đã tokenize sẵn trong RAM (docs/s).
- `plugin_raw_throughput_median`: Throughput của OpenSearch Plugin khi nhận payload text thô (docs/s).
- `engine_throughput_ratio_vanilla_vs_plugin`: Tỷ số tăng tốc pure ingestion của Vanilla so với Plugin.
- `vanilla_container_cpu_median_pct` & `plugin_container_cpu_median_pct`: Container CPU%.
- `vanilla_gc_time_ms_median` & `plugin_gc_time_ms_median`: JVM GC collection time (ms).

### 3. `search_latency_summary.json` (.csv)
- `model`: `vanilla` hoặc `plugin`.
- `concurrency`: `1`, `8`, `32`.
- `total_queries`: Tổng số query test (`1000`).
- `successful_queries`: Số query thành công (`1000`).
- `errors`: Số lỗi (`0`).
- `wall_time_sec`: Tổng thời gian chạy test (giây).
- `throughput_qps`: Throughput tìm kiếm (Queries Per Second).
- `latency_p50_ms`, `latency_p90_ms`, `latency_p95_ms`, `latency_p99_ms`, `latency_max_ms`: Các mốc phân vị độ trễ (ms).

### 4. `soak_summary.json` (.csv) *(được sinh khi hoàn tất Phase E)*
- `model`: `vanilla` hoặc `plugin`.
- `duration_target_sec`: Thời lượng mục tiêu (`1800` giây / 30 phút).
- `duration_actual_sec`: Thời lượng thực tế chạy.
- `total_operations`: Tổng số operations hoàn tất.
- `overall_throughput_ops_sec`: Throughput trung bình cả đợt (ops/s).
- `throughput_drift_pct`: Tỷ lệ sụt giảm/tăng throughput giữa cửa sổ 5 phút đầu và 5 phút cuối (%).
- `rss_slope_mib_per_min`: Độ dốc tăng trưởng bộ nhớ RSS container (MiB/phút).
- `jvm_heap_slope_mib_per_min`: Độ dốc tăng trưởng JVM Heap (MiB/phút).
- `total_gc_collection_time_sec`: Tổng thời gian JVM GC collection trong suốt 30 phút (giây).
- `cgroup_throttled_sec`: Tổng thời gian cgroup CPU throttling (giây).
- `crashes_observed`: Số lần crash quan sát được (`0`).
