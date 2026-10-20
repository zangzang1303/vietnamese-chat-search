# DANH MỤC SƠ ĐỒ KIẾN TRÚC & BIỂU ĐỒ TRỰC QUAN (`image/`)

Thư mục này lưu trữ toàn bộ các hình ảnh đồ thị đo kiểm thực nghiệm, sơ đồ kiến trúc hệ thống và luồng dữ liệu tuần tự.

---

## 1. Hệ Thống 9 Biểu Đồ Đo Kiểm Thực Nghiệm (Benchmark Charts)
Toàn bộ biểu đồ dưới đây được sinh tự động bằng script `benchmarks/generate_final_charts.py` từ dữ liệu thực nghiệm đã đo:

1. **[benchmark_throughput_scaling.png](benchmark_throughput_scaling.png)**: Đường cong thông lượng nạp dữ liệu (docs/s) từ C=1 đến C=32 kèm khoảng tin cậy Percentile Bootstrap 95% CI quanh Median.
2. **[benchmark_paired_throughput_difference.png](benchmark_paired_throughput_difference.png)**: Hiệu số thông lượng bắt cặp từng repetition $(A - B)$ với khoảng tin cậy 95% CI quanh điểm 0 (chứng minh không có mô hình nào thắng áp đảo về mặt thống kê).
3. **[benchmark_latency_distribution.png](benchmark_latency_distribution.png)**: Phân bố các phân vị độ trễ (p50, p90, p95, p99, max) tại mức tải chuẩn C=8.
4. **[benchmark_jvm_gc_stats.png](benchmark_jvm_gc_stats.png)**: So sánh tổng thời gian JVM Garbage Collection làm việc (Mô hình B cao hơn từ +19% đến +40%).
5. **[benchmark_storage_segments.png](benchmark_storage_segments.png)**: Dung lượng lưu trữ chỉ mục Lucene trên đĩa và số lượng tập tin segment.
6. **[benchmark_batch_sensitivity.png](benchmark_batch_sensitivity.png)**: So sánh hiệu năng ở các kích thước lô: batch 500, 1000 và 2000 (phát hiện hiện tượng rớt batch ở Model B tại C=32, b2000).
7. **[benchmark_cost_isolation.png](benchmark_cost_isolation.png)**: Bảng đôi: Tốc độ nạp dữ liệu thuần túy của Engine (Vanilla 2.7x–3.7x) & Sự sụt giảm parallel efficiency của tokenizer CGO do nghẽn mutex.
8. **[benchmark_search_latency.png](benchmark_search_latency.png)**: Độ trễ tìm kiếm (p50/p95/p99) và thông lượng truy vấn QPS trên 1.000 queries thực tế ở C=1, 8, 32.
9. **[benchmark_soak_timeline.png](benchmark_soak_timeline.png)**: Diễn biến bộ nhớ Working Set RSS và JVM Heap qua các cửa sổ 5 phút trong suốt bài kiểm tra độ bền 30 phút.

---

## 2. Hệ Thống Sơ Đồ Kiến Trúc & Luồng Dữ Liệu (Architecture & Sequence Diagrams)
1. **[The Life of a Write Sequence Diagram.png](The%20Life%20of%20a%20Write%20Sequence%20Diagram.png)**: Sơ đồ tương tác tuần tự luồng ghi dữ liệu (Write Flow) từ Client qua Memory Buffer, Translog, Index Inverted List đến Disk Flush.
2. **[Hình 2.1.png](Hình%202.1.png)** & **[Hình 3.1.png](Hình%203.1.png)**: Sơ đồ kiến trúc tổng thể của hệ thống chat search tiếng Việt.
3. **[Cốc Cốc Search Engine-2026-09-18-072911.png](Cốc%20Cốc%20Search%20Engine-2026-09-18-072911.png)**: Mô hình hoạt động của bộ băm từ Cốc Cốc Tokenizer sử dụng cấu trúc Double-Array Trie (DATrie).
4. **[score_comparison_chart.jpg](score_comparison_chart.jpg)** & **[compound_word_trap_demo.jpg](compound_word_trap_demo.jpg)**: Minh họa sự sai lệch điểm số BM25 và "bẫy từ ghép" tiếng Việt khi không có bộ tách từ chuẩn hóa.
