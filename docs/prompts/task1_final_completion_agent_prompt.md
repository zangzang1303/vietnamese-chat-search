# PROMPT THỰC THI HOÀN TẤT TASK 1

Bạn là Lead Performance & Search Engine Engineer. Hãy tiếp quản workspace và hoàn tất có kiểm chứng Task 1: **“Đánh giá ảnh hưởng của Plugin tiếng Việt trên OpenSearch 2.19 — Benchmark, Resource Consumption & Search Parity”**.

## 1. Workspace và nguyên tắc bảo toàn

- Workspace Windows: `D:\CODE\VSF\vietnamese-chat-search`
- Môi trường chạy: WSL2 distro `Ubuntu`, Docker chạy bên trong WSL.
- Lệnh Docker: `wsl.exe -u root -d Ubuntu -- docker ...`
- Worktree đang dirty và chứa thay đổi của người dùng/agent trước. Trước tiên chạy `git status --short`.
- Không `git reset`, không `git checkout --`, không xóa hay ghi đè artifact cũ.
- Không chạy hai OpenSearch container đồng thời vì WSL chỉ có khoảng 7,6 GiB RAM.
- Dùng một WSL session sống liên tục hoặc keepalive; kiểm tra container uptime/restart trước và sau mỗi suite.
- Gửi progress update ít nhất mỗi 5–10 phút trong các bài chạy dài.

## 2. Tài liệu và code bắt buộc đọc trước

1. `docs/new plan/plan_task1_opensearch_plugin_benchmark.md`
2. `docs/benchmark_opensearch_plugin_report.md`
3. `benchmarks/bench_runner.go`
4. `benchmarks/run_final_matrix_orchestrator.py`
5. `benchmarks/run_final_parity.py`
6. `benchmarks/run_final_search_parity.py`
7. `benchmarks/generate_final_charts.py`
8. `benchmarks/docker-compose.benchmark.yml`
9. `benchmarks/mapping_vanilla.json`
10. `benchmarks/mapping_plugin.json`
11. Toàn bộ artifact tại `benchmarks/final_run/final_20261020_093100/`

Các file mới đã được chuẩn bị nhưng **chưa được xác nhận bằng full run**, vì vậy phải review/test trước khi dùng:

- `benchmarks/analyze_final_statistics.py`
- `benchmarks/run_completion_supplement.py`
- `benchmarks/tokenizer_scaling_profile.go`
- Thay đổi duration mode trong `benchmarks/bench_runner.go`

Không được mặc định rằng code chuẩn bị sẵn là đúng.

## 3. Trạng thái đã xác minh của run chính

Run chính: `benchmarks/final_run/final_20261020_093100/`

- Có 50 measured runs: 2 mô hình × C={1,4,8,16,32} × 5 repetitions.
- Mỗi run đã index 1.000.000 operations/documents, không có HTTP error, bulk-item error, tokenizer error hoặc write rejection.
- Thiết kế paired/randomized AB/BA, seed 42; có hai warm-up runs.
- Model A dùng binary Linux CGO thật và dictionary đồng bộ từ plugin.
- Token parity: 7.642/8.000 = 95,525%; không đạt zero divergence.
- Search parity: `text` đạt 100%; các field tiếng Việt không đạt 100%; `edge_prefix` trong multi-field lệch nghiêm trọng.
- 50 telemetry file cũ chỉ có JVM/thread-pool. Container CPU/RSS bị thiếu vì collector cũ timeout và nuốt exception.

Các kết luận thống kê đã biết:

- CV throughput khoảng 13,5%–36,8%.
- Paired 95% CI của chênh lệch throughput A−B chứa 0 tại mọi concurrency.
- Vì vậy không được gọi các chênh lệch median là “A thắng” hoặc “B thắng” có ý nghĩa thống kê.
- Báo cáo hiện ghi “Median [95% CI]”, nhưng CI cũ được tính quanh mean; đây là lỗi phải sửa.
- `collection_time_in_millis` là **GC collection time**, không được gọi là STW GC pause nếu không có GC log/JFR.
- Global `sync.Mutex` trong Model A là giả thuyết contention hợp lý, chưa phải root cause đã chứng minh nếu chưa có profile riêng.

## 4. Mục tiêu hoàn tất

Kết thúc công việc chỉ khi có đủ:

1. Thống kê đúng trên 50 run gốc.
2. Resource consumption thật cho OpenSearch container và Go client.
3. Batch sensitivity theo kế hoạch gốc.
4. Duration stability/soak đủ 30 phút mỗi mô hình.
5. Engine-only measurement và tokenizer-only scaling profile để tách chi phí.
6. Search latency theo concurrency, ngoài search parity hiện có.
7. Manifest/provenance được sửa và artifact được kiểm tra máy.
8. Charts sinh từ dữ liệu, không hard-code.
9. Báo cáo cuối không còn overclaim và có bảng completion PASS/FAIL.

## 5. Trình tự thực thi bắt buộc

### Phase A — Audit và smoke test

1. Ghi lại `git status --short`; phân biệt thay đổi có sẵn với thay đổi của bạn.
2. Kiểm tra Docker/WSL, disk free, RAM, CPU, container status, `OOMKilled`, restart count và uptime.
3. Xác minh resource limit bằng `docker inspect`:
   - 4 vCPU/container.
   - 3 GiB RAM/container.
   - JVM heap 1536 MiB.
4. Xác minh OpenSearch 2.19.0, Java, Go, Docker version; sửa lỗi collector hiện chỉ lưu `Client:` và bỏ trống Java version.
5. Xác minh checksum dataset, mappings, dictionary, runner, plugin JAR và native `.so`.
6. Không tự bịa commit plugin/native library. Nếu không truy được provenance, ghi `unknown` và giới hạn reproducibility tương ứng.
7. Review duration mode trong `bench_runner.go`; chạy unit/smoke 5–10 giây trên từng model. Duration mode phải:
   - chạy gần đúng thời lượng yêu cầu;
   - dừng sạch;
   - không có lỗi bulk/tokenizer;
   - không báo progress/count sai;
   - không dùng count unique document làm số operation khi test ghi đè ID.
8. Review collector mới. Không được tiếp tục full suite nếu CPU/RSS vẫn `null`, bằng 0 bất thường, hoặc sample interval không gần 1 giây.

### Phase B — Sửa và tái tạo thống kê của 50 run

Chạy và kiểm tra `benchmarks/analyze_final_statistics.py`.

Yêu cầu output:

- `summaries/statistical_analysis.json`
- `summaries/paired_throughput.csv`
- Median và percentile-bootstrap 95% CI phải được ghi đúng tên.
- Mean và Student-t CI phải là cột riêng.
- Với A/B: dùng paired difference theo cùng repetition, báo mean difference, paired 95% CI và paired effect size (Cohen's dz hoặc tương đương).
- Ghi rõ không concurrency nào có kết luận thắng nếu paired CI chứa 0.
- Không dùng từng bulk request làm sample thống kê; experimental unit là full run.
- Nếu CV >10%, đánh dấu `unstable/high variance`. Không bắt buộc chạy lại 50 run nếu báo cáo công khai giới hạn này; tuyệt đối không cherry-pick best run.

### Phase C — Resource telemetry hợp lệ

Sửa/test `benchmarks/run_completion_supplement.py` trước khi chạy. Telemetry 1 giây phải có:

- Container CPU%.
- Container memory working set/RSS.
- CPU throttling từ cgroup (`nr_periods`, `nr_throttled`, `throttled_usec` hoặc tương đương).
- JVM heap used.
- Young/old GC count và collection time.
- OpenSearch process CPU.
- Write active/queue/rejected.
- Indexing pressure nếu endpoint hỗ trợ.
- Disk/block read-write hoặc cgroup I/O.
- Merge/flush/segments/store đầu-cuối.
- Go client: user CPU seconds, system CPU seconds, CPU%, max RSS.
- Tổng CPU-seconds ước lượng của client + container, với công thức và giới hạn ghi rõ.
- Timestamp và actual sample interval; báo tỷ lệ sample missing.

Không nuốt exception. Collector error phải làm quality gate fail hoặc được ghi rõ trong artifact.

### Phase D — Batch sensitivity

Chạy theo đúng matrix trong kế hoạch gốc, paired AB/BA, ba repetitions/config:

- Baseline: C=1, 100.000 docs, batch 500 và 1.000.
- Heavy stress: C=32, 1.000.000 docs, batch 1.000 và 2.000.
- Cả Model A và Model B.
- Warm-up không tính kết quả.
- Recreate index cho mỗi measured run.

Mỗi run phải qua gate:

- HTTP errors = 0.
- Bulk item errors = 0.
- Tokenizer errors = 0.
- Write rejections = 0, hoặc nếu có thì phải được xem là kết quả saturation chứ không được bỏ qua.
- Với fixed-doc run, `/_count` phải đúng expected docs.
- Container không restart và không OOMKilled.
- Telemetry container và client không rỗng.

### Phase E — Soak/duration stability

Chạy tuần tự 30 phút/model tại C=16, batch=1.000. Nếu đủ thời gian có thể nâng lên 60 phút, nhưng 30 phút là điều kiện hoàn thành theo kế hoạch gốc.

- Model A trước/B sau hoặc random hóa thứ tự và lưu seed.
- Dùng duration runner trong một process liên tục, không loop khởi động runner nhiều lần.
- Có thể ghi đè cùng 1M IDs để giới hạn disk, nhưng báo rõ metric `successful indexing operations` khác `unique documents`.
- Thu telemetry liên tục mỗi giây.
- Chia timeline thành các cửa sổ 5 phút và tính:
  - throughput từng cửa sổ;
  - throughput drift đầu-cuối;
  - RSS/heap slope MiB/phút;
  - GC collection rate/time;
  - merge/flush;
  - queue/rejection/errors;
  - restart/OOM/crash.
- Chỉ được viết “không quan sát crash trong 30 phút và N operations”, không suy ra xác suất crash bằng 0.

### Phase F — Tách chi phí và kiểm chứng giả thuyết

1. Chạy `tokenizer_scaling_profile.go` ở C={1,4,8,16,32}, tối thiểu 3 repetitions/concurrency, cùng corpus và dictionary final.
2. Báo tokenizer ops/s, speedup, parallel efficiency và saturation.
3. Nếu có thể, thu Go mutex profile/pprof. Chỉ kết luận mutex là bottleneck khi profile thể hiện thời gian chờ khóa đáng kể; nếu không, gọi là giả thuyết được hỗ trợ bởi code inspection/tokenizer-only scaling.
4. Thực hiện engine-only suite:
   - Chuẩn bị payload ngoài vùng timing.
   - Model A dùng corpus đã CGO-tokenize trước; Model B dùng raw corpus.
   - Bulk sender/timing logic giống nhau.
   - Chạy C=8 và C=16, batch=1.000, tối thiểu 3 repetitions/model.
   - Báo rõ engine-only không phải total-system cost và payload/analyzer giữa hai kiến trúc vẫn khác về bản chất.

### Phase G — Search latency bổ sung

- Dùng ít nhất 1.000 unique queries hoặc tập 1.000 queries hiện có.
- Concurrency tìm kiếm: 1, 8, 32.
- Warm-up riêng và chỉ công bố warm-cache nếu không có phương pháp cold-cache hợp lệ.
- Random hóa model/query order.
- Báo p50/p95/p99, error rate và throughput queries/s.
- Giữ search parity và search performance thành hai kết luận riêng.

### Phase H — Aggregate, charts và report

1. Sinh `supplement/summary.json`, CSV tương ứng và README mô tả schema.
2. Sinh lại chart từ raw/aggregate. Ít nhất phải có:
   - Throughput scaling với CI đúng nhãn.
   - Paired throughput differences với CI quanh zero.
   - Latency percentiles.
   - Container + client CPU/RSS.
   - GC collection metrics.
   - Batch sensitivity.
   - Soak timeline/drift.
   - Search latency theo concurrency.
3. Cập nhật `docs/benchmark_opensearch_plugin_report.md`:
   - Đổi mọi “GC pause/STW” chưa chứng minh thành “GC collection time”.
   - Xóa tuyên bố dual-regime/winner nếu paired CI không hỗ trợ.
   - Đổi nguyên nhân mutex/JNI allocation thành hypothesis nếu không có profile chứng minh.
   - Không nói reproducible 100% khi provenance còn unknown.
   - Không nói telemetry cũ lấy mẫu mỗi giây nếu interval thực tế không đạt.
   - Dùng relative links, không dùng `file:///d:/...`.
   - Mô tả dataset là mixed corpus, không phải 1M chat messages thuần.
   - Tách rõ observed evidence, statistical inference và qualitative architectural risk.
   - Recommendation phải dựa trên trade-off; không hard-code Model A từ trước.
   - Ghi rõ zero-divergence gate đã FAIL và hệ quả với khả năng thay thế tương đương.
4. Sửa manifest để chứa đúng Docker/Java version, resource limits, checksums và provenance.

## 6. Cấu trúc artifact bổ sung

Không ghi đè 50 run gốc. Lưu phần mới tại:

```text
benchmarks/final_run/final_20261020_093100/supplement/
├── README.md
├── experimental_plan.json
├── raw/
├── telemetry/
├── client/
├── engine_only/
├── tokenizer_profile/
├── search_latency/
├── summaries/
├── charts/
└── logs/
```

Mỗi artifact phải có timestamp, model, concurrency, batch, docs/duration, repetition, command hoặc config, và checksum nếu phù hợp.

## 7. Definition of Done

Chỉ đánh dấu Task 1 hoàn tất khi tất cả mục sau có trạng thái rõ ràng:

| Gate | Điều kiện PASS |
|---|---|
| Primary matrix integrity | 50/50 raw runs hợp lệ và truy vết được |
| Statistical validity | CI đúng nhãn; có paired analysis; không overclaim |
| Token parity | Kết quả định lượng + mismatch evidence; zero-divergence được đánh FAIL nếu <100% |
| Search parity | 1.000 queries, field-level + production query evidence |
| Resource consumption | Container và client CPU/RSS hợp lệ, có total-system accounting |
| Batch sensitivity | Đủ matrix kế hoạch gốc × 3 reps × 2 models |
| Soak | ≥30 phút/model, không thiếu timeline telemetry |
| Cost isolation | Tokenizer-only + engine-only evidence |
| Search performance | C=1/8/32, p50/p95/p99 và errors |
| Reliability wording | Chỉ mô tả observation window, không suy xác suất crash |
| Reproducibility | Manifest/commands/checksums đầy đủ; unknown provenance được công khai |
| Report consistency | Mọi số trong report truy được về artifact; link portable |

Gate nghiên cứu có thể FAIL mà Task vẫn hoàn thành, ví dụ parity <100%. “Hoàn thành” nghĩa là đã đo và kết luận trung thực, không phải tất cả giả thuyết đều đúng.

## 8. Kết quả phải trả lại

Khi hoàn tất, trả lời bằng tiếng Việt, ngắn gọn nhưng có bằng chứng:

1. Verdict: Task 1 đã hoàn tất hay chưa.
2. Bảng PASS/FAIL cho toàn bộ Definition of Done.
3. Danh sách file đã tạo/sửa.
4. Số liệu chính, kèm đường dẫn artifact nguồn.
5. Các limitation còn lại và tác động của chúng.
6. Recommendation kiến trúc sau cùng, tách facts khỏi judgment.
7. Lệnh tái lập chính xác.

Nếu bị block, không tự tuyên bố hoàn tất. Ghi chính xác lệnh, lỗi, artifact liên quan và phần nào vẫn có thể tiếp tục độc lập.
