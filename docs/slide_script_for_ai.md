# BẢN KỊCH BẢN THUYẾT TRÌNH TRỰC QUAN (VISUAL-FIRST SLIDE DECK)
# Định dạng: 50% HÌNH ẢNH / SƠ ĐỒ ĐỒ HỌA + 50% GẠCH ĐẦU DÒNG CÔ ĐỌNG (TRÁNH BẪY NHIỀU CHỮ)
# Hướng dẫn: Copy toàn bộ nội dung từ "=== BẮT ĐẦU COPY ===" dán vào Gamma.app hoặc Pitch.com

=== BẮT ĐẦU COPY ===

# Slide 1: Nghịch Lý Tìm Kiếm Tiếng Việt Trong Chat
## Hiện tượng "Bẫy từ ghép" khiến công cụ tìm kiếm trả về kết quả rác
- [Layout: 50% Ảnh Infographic bên Trái - 50% Chữ bên Phải]
- [Hình ảnh Slide 1: `image/compound_word_trap_demo.jpg`]
- **Nỗi đau thực tế:** Tìm `học sinh` nhưng nhận về toàn tin lạc đề: *sinh nhật*, *sinh viên*, *hy sinh*.
- **Nguyên nhân cốt lõi:** Tiếng Việt là ngôn ngữ đơn lập, khoảng trắng chỉ ngắt âm tiết, không ngắt từ mang nghĩa.
- **Sự thất bại của máy chuẩn:** Bộ tách từ Standard bẻ đôi `học sinh` thành 2 chữ rời rạc, làm mất 100% ngữ nghĩa gốc.
- **Mục tiêu dự án:** Bảo toàn nguyên vẹn từ ghép tiếng Việt, gõ có dấu, không dấu hay gõ dở đều tìm trúng đích!

---

# Slide 2: Kiến Trúc Hệ Thống Tổng Thể (System Architecture)
## Mô hình lai 3 tầng kết hợp NLP C++, Go Brain và Cụm lưu trữ kép
- [Layout: Sơ đồ Kiến trúc toàn màn hình (Image Hero)]
- [Hình ảnh Slide 2: `image/Cốc Cốc Search Engine-2026-09-22-065029.png`]
- **Tầng 1 - Client Web Messenger:** Giao diện Dark Mode 3 cột, hiển thị Inspector luồng xử lý và tab đối soát 3 chiều.
- **Tầng 2 - Go Brain (:8080):** Cầu nối CGO C++ Tokenizer, bộ chuẩn hóa Unicode, và bộ điều phối ghi song song (Dual Dispatcher).
- **Tầng 3 - Storage Cluster:** 
  - 🐘 *Elasticsearch 8.11:* Lưu trữ phân tán, mapping đa trường, xếp hạng Okapi BM25.
  - 💾 *Custom Go Binary Engine:* Động cơ lưu trữ nhị phân trực tiếp trên đĩa, siêu nhẹ 15MB RAM.

---

# Slide 3: Tầng Bóc Tách Ngôn Ngữ Tự Nhiên (NLP via CGO)
## Bí quyết từ điển Double-Array Trie và Quy hoạch động Viterbi
- [Layout: 2 Columns - Sơ đồ Đồ thị bên Trái, Giải thuật bên Phải]
- [Gợi ý đồ họa: Sơ đồ mạng phân từ Word Lattice nối các chữ 'Bác' - 'sĩ' - 'khám' - 'bệnh']
- **Double-Array Trie (DAT):** Nén toàn bộ từ điển tiếng Việt vào 2 mảng số nguyên `BASE` và `CHECK` $\rightarrow$ Tra cứu tốc độ tức thời $O(L)$, dung lượng RAM chỉ vỏn vẹn **45 MB**.
- **Thuật toán Viterbi HMM:** Giải quyết bài toán nhập nhằng từ vựng bằng cách tìm đường đi ngắn nhất trên đồ thị có hướng trong thời gian tuyến tính $O(N)$.
- **CGO Bridge Thread-Safe:** Cầu nối trung gian giữa Go và C++, bọc khóa Mutex chống xung đột luồng và giải phóng vùng nhớ C thủ công chống rò rỉ RAM.

---

# Slide 4: Thiết Kế "Chiếc Tủ 3 Ngăn" (Multi-Field Mapping)
## Phân nhánh dữ liệu đón đầu mọi thói quen gõ phím của người dùng
- [Layout: 3 Cards Đồ Họa Trực Quan - Đại diện 3 Ngăn Tủ]
- [Gợi ý đồ họa: Hình minh họa 1 tin nhắn chạy qua phễu phân tách thành 3 trường song song]
- **Ngăn 1 - `content_tokenized` (Chữ chuẩn có dấu):**
  - Lưu từ ghép nối dấu gạch dưới: `uống cà_phê sáng`.
  - Giữ nguyên vẹn 100% ngữ nghĩa tiếng Việt.
- **Ngăn 2 - `content_unaccented` (Chữ bỏ hết dấu):**
  - Lưu chuỗi không dấu: `uong ca_phe sang`.
  - Phục vụ người dùng chat nhanh không bật bộ gõ tiếng Việt.
- **Ngăn 3 - `content_partial` (Lát cắt tiền tố Edge N-gram):**
  - Băm nhỏ từ ghép từ 2 đến 15 ký tự: `cà`, `cà_p`, `cà_ph`, `ca_p`.
  - Phục vụ tìm kiếm dở từ / gợi ý tức thì (Autocomplete).

---

# Slide 5: LUỒNG 1: QUY TRÌNH ĐÁNH CHỈ MỤC (INDEXING PIPELINE)
## Sơ đồ tuần tự 5 bước đưa tin nhắn thô vào Lucene Inverted Index
- [Layout: Sơ Đồ Tuần Tự (Sequence Diagram) Trung Tâm]
- [Hình ảnh Slide 5: `image/Cốc Cốc Search Engine-2026-09-18-072911.png`]
- **Bước 1 (Nhận tin):** REST API tiếp nhận tin nhắn thô: `"Uống cà phê sáng nhé"`.
- **Bước 2 (CGO Bridge):** Cốc Cốc đóng gói từ ghép thành: `["Uống", "cà_phê", "sáng", "nhé"]`.
- **Bước 3 (Biến thể):** Sinh chuỗi không dấu `ca_phe` và mạng lưới tiền tố `cà p`, `cà_p`.
- **Bước 4 (Bulk API):** Đóng gói tài liệu đa trường bắn sang Elasticsearch theo lô.
- **Bước 5 (Ghi đĩa):** Lưu Translog chống mất điện $\rightarrow$ Kết xuất thành Segment chỉ mục ngược bất biến.

---

# Slide 6: LUỒNG 2: QUY TRÌNH TRUY VẤN & XẾP HẠNG (SEARCH PIPELINE)
## Cơ chế chấm điểm Okapi BM25 kết hợp chiến lược Boosting đa tầng
- [Layout: Sơ Đồ Tuần Tự bên Trái - Thang Điểm Boosting bên Phải]
- [Hình ảnh Slide 6: `image/Cốc Cốc Search Engine-2026-09-18-073414.png`]
- **Sơ đồ luồng:** Query người dùng $\rightarrow$ Bóc tách đa tầng $\rightarrow$ Xây dựng Bool Query $\rightarrow$ Quét Inverted Index $\rightarrow$ Xếp hạng BM25.
- **Thang điểm trao giải (Relevance Boosting):**
  - 🥇 **Boost 5.0x:** Khớp chính xác từ ghép có dấu (`content_tokenized`).
  - 🥈 **Boost 4.0x:** Khớp đúng cụm từ liền kề (`match_phrase`).
  - 🥉 **Boost 3.0x:** Khớp từ ghép không dấu (`content_unaccented`).
  - 🏅 **Boost 10.0x:** Khớp tiền tố gõ dở (`content_partial` kèm toán tử `AND`).

---

# Slide 7: Xử Lý Trường Hợp Biên: Giải Quyết Bài Toán Gõ Dở
## Kỹ thuật tinh chỉnh Search Analyzer giúp từ gõ dở nhảy vọt lên Top 1
- [Layout: 2 Cards Before vs After - Kèm Biểu Đồ Cột Nhỏ]
- [Gợi ý đồ họa: Biểu đồ cột so sánh điểm số 33.36đ vs 93.91đ khi gõ 'học sin']
- **Vấn đề ban đầu (Điểm số bị ngược):**
  - Gõ dở `"học sin"`: Máy bẻ vụn thành `học` và `sin`. Do dùng toán tử OR mặc định, tin nhắn chỉ chứa từ đơn `"học"` bị ăn điểm cao hơn bài có đúng từ ghép `"học sinh"`.
- **Giải pháp xử lý:**
  - Tách riêng `search_analyzer: coccoc_whitespace_analyzer` (không băm nhỏ câu query).
  - Ép buộc toán tử `AND` (bắt buộc phải có đủ các mẩu cấu thành).
- **Kết quả:** Điểm tin nhắn chứa `"học sinh"` tăng vọt từ **33.36 lên 93.91**, chiếm trọn **Top 1**!

---

# Slide 8: Đột Phá: Động Cơ Lưu Trữ Nhị Phân Thuần Go (Custom Engine)
## Kiến trúc LSM-Tree cấp thấp: Forward Index O(1) và Nhật ký WAL chống sập nguồn
- [Layout: Sơ Đồ Kiến Trúc Tệp Nhị Phân (Low-level Architecture)]
- [Hình ảnh Slide 8: `image/The Life of a Write Sequence Diagram.png`]
- **`docstore.dat` & `docstore.idx` (Forward Index):** Bản ghi index cố định **12 bytes/doc**. Đọc tin nhắn gốc tức thời theo mốc byte con trỏ: $\text{Offset} = \text{DocID} \times 12$ ($O(1)$ tuyệt đối).
- **`wal.log` (Write-Ahead Log):** Ghi tuần tự kèm mã kiểm tra toàn vẹn **CRC32 IEEE**, tự động Replay khôi phục 100% dữ liệu khi sập nguồn.
- **`terms.dict` & `postings.bin` (Inverted Index Bất Biến):** Từ điển sắp xếp Alphabet A-Z tra cứu nhị phân $O(\log M)$.
- **Hiệu năng vượt trội:** Không cần JVM/Docker, chỉ tốn **15 MB RAM**, tốc độ tìm kiếm chỉ **~3 ms**!

---

# Slide 9: Trực Quan Hóa: Web Messenger Đối Soát 3 Chiều
## Trải nghiệm thực tế tương tác thời gian thực và kiểm chứng trực quan
- [Layout: Mockup Giao Diện Web 3 Cột Thời Gian Thực]
- [Gợi ý đồ họa: Ảnh chụp màn hình giao diện Messenger Dark Mode với 3 cột kết quả màu sắc khác nhau]
- **Khung Chat Messenger:** Giao diện Dark Mode chuẩn Facebook Messenger, chat bong bóng, hỗ trợ phòng chat nhóm.
- **Thanh Flow Inspector:** Bóc tách trực quan theo thời gian thực: `Chữ thô` $\rightarrow$ `Tokens Cốc Cốc` $\rightarrow$ `Không dấu` $\rightarrow$ `Edge N-grams`.
- **Bảng đối soát 3 chiều cạnh nhau:**
  - 🔴 *Cột 1 (Baseline ES):* Dính lỗi bẫy từ ghép (trả về cả tin nhắn rác).
  - 🟢 *Cột 2 (Cốc Cốc ES):* Lọc sạch rác, chỉ giữ đúng từ ghép chuẩn (~14ms).
  - 🟣 *Cột 3 (Custom Go Engine):* Chuẩn xác tuyệt đối với tốc độ tia chớp **~3ms**.

---

# Slide 10: Bảng Vàng Thành Tích (Benchmark IR) & Tổng Kết
## Những con số định lượng khoa học chứng minh hiệu quả vượt trội
- [Layout: 50% Ảnh Infographic Biểu Đồ Cột - 50% Điểm Nhấn Tổng Kết]
- [Hình ảnh Slide 10: `image/benchmark_ir_comparison.jpg`]
- **🎯 MRR (Mean Reciprocal Rank):** Đạt **0.7708** (🚀 Tăng +23.3% so với 0.6250).
- **📈 NDCG@10 (Chất lượng xếp hạng):** Đạt mức tối ưu **0.9472** (🚀 Tăng +15.3%).
- **🥇 Precision@1 (Độ chuẩn Top 1):** Đạt **75.0%** (🚀 Tăng +28.6% so với 58.3%).
- **⚡ Tốc độ tìm kiếm:** Cốc Cốc ES đạt 14.2ms, Custom Engine Go đạt kỷ lục **3.0ms** (gấp 5 lần ES).
- **Thông điệp kết luận:** Kết hợp Tokenizer tiếng Việt chuyên biệt với kiến trúc schema đa trường và động cơ nhị phân là giải pháp toàn diện nhất cho bài toán tìm kiếm tin nhắn chat tiếng Việt.

=== KẾT THÚC COPY ===
