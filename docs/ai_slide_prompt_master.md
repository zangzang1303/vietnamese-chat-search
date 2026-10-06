Hãy đóng vai một chuyên gia thiết kế thuyết trình công nghệ hàng đầu (Senior Tech Presentation Designer). Hãy tạo cho tôi một bộ slide thuyết trình gồm đúng 10 SLIDE theo phong cách Tech Startup hiện đại, giao diện Dark Mode (nền xám đen than chì Graphite, điểm nhấn xanh neon và tím neon).

Yêu cầu thiết kế quan trọng:
- Bố cục Visual-First: Mỗi slide chia 50% diện tích cho [KHUNG ẢNH TRỐNG] để tôi tự tải ảnh kiến trúc từ máy lên, và 50% diện tích cho nội dung chữ.
- Nội dung chữ: Cực kỳ cô đọng, súc tích (chỉ 3-4 gạch đầu dòng logic, không viết đoạn văn dài, không dùng mã code vụn vặt).
- Ở mỗi slide, bắt buộc phải tạo một khung chứa ảnh trống (Image Placeholder) chuyên nghiệp ở đúng vị trí chỉ định.

Dưới đây là nội dung chi tiết của 10 slide:

---
SLIDE 1: KIẾN TRÚC HỆ THỐNG TỔNG THỂ (THE BIG PICTURE)
- Bố cục: 2 Cột (Cột trái: Khung ảnh kiến trúc; Cột phải: 3 Tầng hệ thống).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải sơ đồ kiến trúc tổng thể lên)]
- Nội dung:
  • Tầng 1 (Client): Giao diện Facebook Messenger Dark Mode 3 cột thời gian thực, tích hợp Flow Inspector.
  • Tầng 2 (Go Brain :8080): Cầu nối CGO C++ Tokenizer, bộ chuẩn hóa Unicode, và bộ điều phối ghi song song (Dual Dispatcher).
  • Tầng 3 (Dual Storage): Cụm lưu trữ kép gồm Elasticsearch 8.11 phân tán và Động cơ nhị phân thuần Go lưu đĩa siêu nhẹ 15MB RAM.

---
SLIDE 2: PHÂN TÍCH CHUYÊN SÂU: CỐC CỐC TOKENIZER & CGO
- Bố cục: 2 Cột (Cột trái: Khung ảnh giải thuật; Cột phải: Phân tích kỹ thuật).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải sơ đồ kết nối CGO & Viterbi lên)]
- Nội dung:
  • Giải quyết bẫy từ ghép: Khóa chặt từ ghép tiếng Việt (cà_phê, học_sinh), ngăn chặn 100% hiện tượng ăn lan sang từ đơn lẻ.
  • Double-Array Trie (DAT): Nén toàn bộ từ điển tiếng Việt vào 2 mảng số nguyên -> Tra cứu tức thời O(L), RAM chỉ 45MB.
  • Giải thuật Viterbi HMM: Tìm chuỗi phân đoạn từ tối ưu nhất trên đồ thị có hướng theo thời gian tuyến tính O(N).
  • CGO Bridge an toàn: Cơ chế giải phóng bộ nhớ thủ công chống rò rỉ RAM và bọc Mutex bảo vệ đa luồng.

---
SLIDE 3: DỮ LIỆU ĐƯỢC LƯU TRỮ NHƯ THẾ NÀO? (STORAGE INTERNALS)
- Bố cục: 2 Cột (Cột trái: Khung ảnh cấu trúc đĩa; Cột phải: 2 Thế giới lưu trữ).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải sơ đồ cấu trúc lưu trữ nhị phân lên)]
- Nội dung:
  • Phía Elasticsearch (Chiếc tủ 3 ngăn): Phân nhánh 3 trường dữ liệu: chữ chuẩn có dấu, chữ không dấu và lát cắt Edge N-gram 2-15 ký tự.
  • Phía Custom Go Engine (Lưu đĩa trực tiếp):
    - docstore.idx & dat: Bảng con trỏ cố định 12 bytes/doc -> Tra cứu tin nhắn gốc O(1) tuyệt đối.
    - wal.log: Sổ nhật ký ghi trước có mã kiểm tra CRC32 IEEE chống sập nguồn.
    - terms.dict: Chỉ mục ngược bất biến sắp xếp A-Z tra cứu nhị phân O(log M).

---
SLIDE 4: TỔNG QUAN LUỒNG ĐÁNH CHỈ MỤC (INDEXING FLOW OVERVIEW)
- Bố cục: 2 Cột (Cột trái: Khung ảnh sơ đồ khối luồng Index; Cột phải: Dây chuyền 5 trạm).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải sơ đồ khối luồng Indexing lên)]
- Nội dung:
  • Trạm 1 & 2: Nhận tin nhắn thô qua REST API -> Cốc Cốc Tokenizer đóng gói khóa chặt từ ghép.
  • Trạm 3: Normalizer nhân bản dữ liệu thành bản không dấu và mạng lưới lát cắt tiền tố Edge N-gram.
  • Trạm 4: Dual-Write Dispatcher điều phối ghi đồng thời sang cả 2 động cơ.
  • Trạm 5: Ghi nhật ký an toàn và kết xuất thành các khối chỉ mục ngược bất biến trên đĩa trong 1.78 mili-giây!

---
SLIDE 5: DIỄN GIẢI SƠ ĐỒ TUẦN TỰ LUỒNG INDEXING (SEQUENCE BREAKDOWN)
- Bố cục: 2 Cột (Cột trái: Khung ảnh Sequence Diagram; Cột phải: Các bước tương tác thời gian).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải Sequence Diagram Indexing lên)]
- Nội dung:
  • Bước 1 & 2: Client gửi HTTP POST -> Go Server gọi C++ Cốc Cốc qua CGO Bridge để nhận mảng token.
  • Bước 3 & 4: Sinh biến thể Unicode và đóng gói toàn bộ vào hồ sơ tài liệu đa trường.
  • Bước 5: Bắn gói Bulk NDJSON sang Elasticsearch:
    - Ghi tuần tự vào Translog chống mất điện.
    - Nạp vào bộ đệm RAM, định kỳ kết xuất thành Segment chỉ mục ngược bất biến.

---
SLIDE 6: TỔNG QUAN LUỒNG SEARCH (SEARCH FLOW OVERVIEW)
- Bố cục: 2 Cột (Cột trái: Khung ảnh sơ đồ khối luồng Search; Cột phải: 4 Giai đoạn tìm kiếm).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải sơ đồ khối luồng Search lên)]
- Nội dung:
  • Giai đoạn 1: Bóc tách và chuẩn hóa câu truy vấn của người dùng qua cùng bộ Tokenizer như lúc lưu tin.
  • Giai đoạn 2: Xây dựng câu truy vấn đa tầng gắn các hệ số thưởng điểm tương quan (Boosting).
  • Giai đoạn 3: Parallel Search Dispatcher kích hoạt chạy đua cùng lúc 3 động cơ độc lập.
  • Giai đoạn 4: Tổng hợp kết quả, đo đạc thời gian mili-giây và hiển thị đối soát song song.

---
SLIDE 7: DIỄN GIẢI SƠ ĐỒ TUẦN TỰ LUỒNG SEARCH (SEQUENCE BREAKDOWN)
- Bố cục: 2 Cột (Cột trái: Khung ảnh Sequence Diagram; Cột phải: Thang điểm Boosting & BM25).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải Sequence Diagram Search lên)]
- Nội dung:
  • Chiến lược phân bổ trọng số Boosting:
    - 🥇 Boost 5.0x: Khớp chính xác từ ghép có dấu (content_tokenized - Ưu tiên cao nhất).
    - 🥈 Boost 4.0x: Khớp đúng vị trí từ liền kề (match_phrase).
    - 🥉 Boost 3.0x: Khớp từ ghép không dấu (content_unaccented).
    - 🏅 Boost 1.0x - 2.0x: Khớp tiền tố gõ dở thông thường.
  • Thuật toán Okapi BM25: Hãm trần bão hòa tần suất chống spam từ và phạt công bằng tin nhắn quá dài.

---
SLIDE 8: TINH CHỈNH BÀI TOÁN BIÊN: XỬ LÝ GÕ DỞ TỪ GHÉP ('HỌC SIN')
- Bố cục: 2 Cột (Cột trái: Khung ảnh so sánh điểm số; Cột phải: 3 Bước giải quyết triệt để).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải biểu đồ điểm số nhảy vọt lên)]
- Nội dung:
  • Lỗi điểm số ngược: Khi gõ dở 'học sin', máy cũ dùng toán tử OR khiến bài chỉ có từ đơn 'học' ăn điểm cao hơn bài có 'học sinh'.
  • 3 Bước xử lý kỹ thuật:
    1. Cấu hình Search Analyzer độc lập (không băm nhỏ câu query khi tra cứu).
    2. Ép buộc toán tử AND (bắt buộc phải có đủ các âm tiết cấu thành).
    3. Thưởng nóng nhánh gõ dở có gạch dưới lên tới 10.0x để bù đắp điểm thiếu hụt.
  • Kết quả: Điểm số nhảy vọt từ 33.36 lên 93.91, đưa tin nhắn đúng lên chiếm trọn Top 1!

---
SLIDE 9: SO SÁNH KẾT QUẢ & BENCHMARK ĐỊNH LƯỢNG (EVALUATION)
- Bố cục: 2 Cột (Cột trái: Khung ảnh biểu đồ cột Benchmark; Cột phải: 4 Chỉ số vàng).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải biểu đồ Benchmark IR lên)]
- Nội dung:
  • 🎯 MRR (Mean Reciprocal Rank): Đạt 0.7708 (🚀 Tăng +23.3% so với Baseline 0.6250).
  • 📈 NDCG@10 (Chất lượng xếp hạng): Đạt mức tối ưu 0.9472 (🚀 Tăng +15.3%).
  • 🥇 Precision@1 (Độ chuẩn Top 1): Đạt 75.0% (🚀 Tăng +28.6% so với Baseline 58.3%).
  • ⚡ Độ trễ tìm kiếm: Cốc Cốc ES đạt 14.2ms; Custom Go Engine đạt kỷ lục 3.0ms (nhanh gấp 5 lần ES, RAM chỉ 15MB).

---
SLIDE 10: HIỆN THỰC HÓA ỨNG DỤNG: WEB MESSENGER ĐỐI SOÁT 3 CHIỀU & TỔNG KẾT
- Bố cục: 2 Cột (Cột trái: Khung ảnh giao diện Web Messenger; Cột phải: Đúc kết giá trị).
- [KHUNG ẢNH TRỐNG: (Để trống khung ảnh ở đây để tôi tải ảnh chụp giao diện Messenger lên)]
- Nội dung:
  • Giao diện Facebook Messenger Dark Mode: Khung chat thời gian thực, thanh Flow Inspector soi token và bảng đối soát 3 chiều.
  • Trực quan hóa đối soát: Cột Đỏ (Baseline dính rác) vs Cột Xanh (Cốc Cốc ES chuẩn 14ms) vs Cột Tím (Custom Go siêu tốc 3ms).
  • Tương tác mượt mà: Bấm vào kết quả -> Khung chat tự động cuộn (Smooth Scroll) và chớp sáng (Pulse Highlight).
  • Lời kết: Dự án giải quyết trọn vẹn từ bản chất ngôn ngữ học tiếng Việt, tối ưu 2 luồng dữ liệu, đến sản phẩm thực tế chạy mượt mà.
