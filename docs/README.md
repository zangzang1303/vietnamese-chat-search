# THƯ MỤC BÁO CÁO GIẢI THÍCH & TÀI LIỆU KỸ THUẬT (`docs/`)

Thư mục này chứa toàn bộ các văn bản giải thích, báo cáo thực nghiệm, tài liệu học thuật và kịch bản thuyết trình của dự án.

## 1. Báo Cáo Thực Nghiệm Cốt Lõi (Primary Engineering Reports)
* 🌟 **[benchmark_opensearch_plugin_report.md](benchmark_opensearch_plugin_report.md)**: **Báo cáo chung cuộc Task 1 (9 chương)** — Đánh giá toàn diện OpenSearch Vanilla vs Plugin, thống kê bắt cặp 50 runs, phân tích Go mutex contention qua pprof, tốc độ nạp engine-only 2.7x–3.7x, batch sensitivity, độ bền 30-phút soak test, search latency và bảng tiêu chí hoàn thành DoD 12 gates.
* 🌟 **[task2_coccoc_tokenizer_go_module_report.md](task2_coccoc_tokenizer_go_module_report.md)**: **Báo cáo chung cuộc Task 2** — Đóng gói Cốc Cốc Tokenizer thành Go Module độc lập (`github.com/vntk/coccoc-tokenizer-go`), cơ chế Dual-Mode (CGO C++ ~1.2µs & Pure-Go fallback ~45µs), kiểm thử an toàn bộ nhớ 0 leak, idiomatic API và CI/CD.

## 2. Báo Cáo Phân Tích & Giải Thích Chi Tiết
* **[bao_cao_giai_thich_bieu_do_thuc_nghiem.md](bao_cao_giai_thich_bieu_do_thuc_nghiem.md)**: Giải thích chi tiết cơ chế hoạt động đằng sau các mốc đồ thị thực nghiệm (tại sao C=1, C=4 Model A dẫn trước; tại sao C=16 Model B vượt lên; phân tích điểm nghẽn Go Mutex và chi phí GC).
* **[search_architecture.md](search_architecture.md)**: Thiết kế kiến trúc tìm kiếm hệ thống, cấu hình analyzer và mapping đa trường tiếng Việt.
* **[custom_storage_engine_spec.md](custom_storage_engine_spec.md)**: Đặc tả kỹ thuật động cơ lưu trữ tùy biến cho tin nhắn chat.
* **[write_flow_and_storage_internals.md](write_flow_and_storage_internals.md)**: Giải thích chi tiết cơ chế tầng thấp của Lucene (Segment, Translog, Flush, Merge).
* **[coccoc_tokenizer_and_cgo_guide.md](coccoc_tokenizer_and_cgo_guide.md)**: Hướng dẫn tích hợp CGO và quản lý vùng nhớ C++.

## 3. Kịch Bản Thuyết Trình & Trình Chiếu (Presentation & Slides)
* **[presentation_10_slides.md](presentation_10_slides.md)**: Cấu trúc 10 slides chuẩn báo cáo trước hội đồng chuyên môn.
* **[kich_ban_thuyet_trinh_tung_slide.md](kich_ban_thuyet_trinh_tung_slide.md)**: Kịch bản thuyết trình chi tiết kèm lời thoại từng slide.
* **[kich_ban_thuyet_trinh_mobile.html](kich_ban_thuyet_trinh_mobile.html)**: Giao diện web tra cứu nhanh kịch bản trên điện thoại.
* **[render_diagrams.html](render_diagrams.html)**: Công cụ web kết xuất trực quan các sơ đồ Mermaid.

## 4. Tài Liệu Kế Hoạch & Prompt Bàn Giao
* **[new plan/plan_task1_opensearch_plugin_benchmark.md](new%20plan/plan_task1_opensearch_plugin_benchmark.md)**: Kế hoạch thực nghiệm gốc cho Task 1.
* **[prompts/task1_final_completion_agent_prompt.md](prompts/task1_final_completion_agent_prompt.md)**: Đề bài và ràng buộc thực thi Task 1.
* **[prompts/task1_handoff_and_results_summary_prompt.md](prompts/task1_handoff_and_results_summary_prompt.md)**: Prompt bàn giao kết quả và quá trình đã thực hiện cho AI Agent tiếp theo.
