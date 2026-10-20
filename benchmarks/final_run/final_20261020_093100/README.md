# THƯ MỤC BẰNG CHỨNG THỰC NGHIỆM GỐC (`final_run/final_20261020_093100/`)

Thư mục này lưu trữ **toàn bộ dữ liệu máy gốc (Raw Evidence)** được đo đạc tự động trong quá trình thực thi Task 1:

---

## 1. Dữ Liệu Ma Trận 50 Run Chính (1.000.000 Documents/Run)
* **`raw/` (50 tệp JSON):** Kết quả chi tiết của từng lượt chạy độc lập (2 mô hình $\times$ 5 mức tải $C \in \{1, 4, 8, 16, 32\} \times 5$ lần lặp AB/BA).
  * Ví dụ: `rep1_vanilla_c8.json`, `rep1_plugin_c8.json`...
* **`telemetry/` (50 tệp JSON):** Chuỗi thời gian lấy mẫu mỗi giây ghi nhận:
  * JVM Heap used, young/old GC time & count, write queue peak, write rejections, indexing ops/s.
* **`summaries/`:**
  * **[aggregate.json](summaries/aggregate.json)** & **[aggregate.csv](summaries/aggregate.csv)**: Bảng số liệu tổng hợp Median, Mean, 95% CI cho toàn bộ ma trận 50 runs.
  * **[statistical_analysis.json](summaries/statistical_analysis.json)** & **[paired_throughput.csv](summaries/paired_throughput.csv)**: **Bằng chứng phân tích bắt cặp (Paired Difference $A - B$)**, chứng minh toàn bộ khoảng tin cậy 95% CI đều chứa điểm 0.

---

## 2. Dữ Liệu Thực Nghiệm Bổ Sung (`supplement/`)
Xem chi tiết tại [supplement/README.md](supplement/README.md):
* **`tokenizer_profile/mutex.pprof`:** Bằng chứng Go pprof chứng minh **100% thời gian khóa delay (273,5s)** tập trung tại `sync.(*Mutex).Unlock` trong `SegmentOriginal`.
* **`summaries/engine_only_summary.json` & `.csv`:** Bằng chứng OpenSearch Vanilla nạp dữ liệu thuần túy nhanh gấp **2.7x – 3.7x** so với Plugin khi đã băm từ sẵn.
* **`summaries/batch_sensitivity_summary.json` & `.csv`:** Bằng chứng độ nhạy lô và 20 batch timeouts ở Model B tại C=32, batch 2000.
* **`summaries/soak_summary.json` & `.csv`:** Bằng chứng kiểm thử độ bền tải 30 phút liên tục (>28M operations, độ dốc bộ nhớ ổn định, 0 crashes).
* **`summaries/search_latency_summary.json` & `.csv`:** Bằng chứng đo độ trễ 1.000 truy vấn tìm kiếm ở C=1, 8, 32 (0.0% lỗi).
* **`parity/full_2000/token_parity_summary.json`:** Bằng chứng đối soát từ vựng trên 2.000 câu (95,525% parity, 358 mismatches).
* **`search/search_parity_summary.json`:** Bằng chứng đối soát truy vấn 1.000 câu trên 5 trường dữ liệu.

---

## 3. Thông Tin Hệ Thống & Kiểm Chứng Tính Toàn Vẹn
* **[manifest.json](manifest.json)**: Ghi nhận Git SHA commit, phiên bản Docker (`29.1.3`), phiên bản Java Temurin (`21.0.6`), hạn ngạch 4 vCPU, 3 GiB RAM và mã băm SHA-256 của toàn bộ dataset và bộ từ điển.
