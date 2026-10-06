# KỊCH BẢN THUYẾT TRÌNH CHI TIẾT TỪNG SLIDE (BÁO CÁO KỸ THUẬT)
**Dự án:** Cải Thiện Hệ Thống Tìm Kiếm Tin Nhắn Tiếng Việt (Cốc Cốc Tokenizer & Elasticsearch trong Go)  
**Quy ước xưng hô:** Diễn giả xưng **"em"** – Người nghe xưng **"anh"**.

---

## 📌 SLIDE 1: NGHỊCH LÝ TÌM KIẾM TIẾNG VIỆT TRONG ỨNG DỤNG CHAT

* **Hình ảnh trên slide:** `image/compound_word_trap_demo.jpg` (So sánh giao diện chat: Khung đỏ dính tin rác vs Khung xanh chuẩn xác).
* **Nội dung hiển thị trên slide:**
  * **Nỗi đau thực tế:** Tìm `"học sinh"`, kết quả trả về toàn tin lạc đề: *"sinh nhật"*, *"sinh viên"*, *"hy sinh"*.
  * **Bản chất ngôn ngữ học:** Tiếng Việt là ngôn ngữ đơn lập. Dấu cách chỉ phân tách **âm tiết** (syllable); đơn vị mang nghĩa cốt lõi là **từ ghép** (*học sinh, cà phê*).
  * **Sự thất bại của máy chuẩn:** Bộ tách từ phương Tây bẻ đôi từ ghép theo dấu cách $\rightarrow$ Gây ra **"Bẫy từ ghép"** (Compound Word Trap).
  * **Mục tiêu dự án:** Khóa chặt ngữ nghĩa từ ghép bằng Cốc Cốc NLP, thiết kế chiến lược tìm kiếm đa tầng và phát triển động cơ nhị phân siêu tốc thuần Go.

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Em chào anh! Hôm nay em xin phép được báo cáo chi tiết với anh về dự án Cải thiện hệ thống tìm kiếm tin nhắn tiếng Việt mà em đã hoàn thiện.
> 
> `[Chỉ tay sang hình ảnh viền đỏ bên trái]`  
> Đầu tiên, em mời anh nhìn vào một tình huống cực kỳ quen thuộc trên màn hình: Khi anh gõ từ khóa **'học sinh'** vào ô tìm kiếm của các ứng dụng chat phổ biến hiện nay, kết quả trả về thường bị lẫn lộn những tin nhắn hoàn toàn lạc đề như: 'chúc mừng **sinh** nhật', 'đón tân **sinh** viên', hay 'sự hy **sinh**'.
> 
> `[Dừng 1 giây - Nhấn giọng]`  
> Tại sao lại có sự phi lý này hả anh? Về mặt ngôn ngữ học, tiếng Việt mình là **ngôn ngữ đơn lập**. Dấu cách trong tiếng Việt chỉ dùng để ngắt **âm tiết**, chứ bản thân một âm tiết đơn lẻ nhiều khi không mang đủ nghĩa. Nghĩa trọn vẹn ở đây phải là **từ ghép** — ví dụ như 'học sinh' hay 'cà phê'.
> 
> Các công cụ tìm kiếm chuẩn phương Tây cứ thấy dấu cách là bẻ đôi từ vựng, vô tình đẩy hệ thống vào **'Cái bẫy từ ghép' (Compound Word Trap)** làm loãng kết quả. Đề tài của em sinh ra để giải quyết dứt điểm vấn đề này: Dạy cho cỗ máy hiểu đúng từ ghép tiếng Việt, để anh gõ có dấu, không dấu hay gõ dở thì hệ thống vẫn tìm trúng đích trong tích tắc!"* `[Chuyển Slide 2]`

---

## 📌 SLIDE 2: BẢN ĐỒ KIẾN TRÚC TOÀN HỆ THỐNG (SYSTEM ARCHITECTURE)

* **Hình ảnh trên slide:** `image/Cốc Cốc Search Engine-2026-09-22-065029.png` (Sơ đồ 3 Tầng: Client - Go Brain - Dual Storage).
* **Nội dung hiển thị trên slide:**
  * **Tầng 1 - Client Web Messenger:** Giao diện Dark Mode, tab Đối soát 3 Động cơ song song và thanh Flow Inspector soi token thời gian thực.
  * **Tầng 2 - Bộ não Golang Server (:8080):** Cầu nối CGO C++ Tokenizer, chuẩn hóa Unicode và phân phối dữ liệu Dual Dispatcher.
  * **Tầng 3 - Cụm Lưu trữ Kép (Dual Storage):**
    * 🐘 **Elasticsearch 8.11:** Schema đa trường tối ưu BM25 chuẩn doanh nghiệp.
    * ⚡ **Custom Go Engine:** Động cơ nhị phân độc lập lưu trữ trực tiếp trên đĩa, siêu nhẹ **15MB RAM**, độ trễ kỷ lục **~3ms**.

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Để giải bài toán đó một cách bài bản, em đã thiết kế toàn bộ hệ thống theo kiến trúc 3 tầng phân lớp rõ ràng như trên sơ đồ:
> 
> `[Chỉ tay từ trên xuống dưới theo sơ đồ]`  
> * **Tầng trên cùng là Client Web:** Em xây dựng giao diện Dark Mode lấy cảm hứng từ Facebook Messenger, có sẵn tab Đối soát 3 Động cơ song song và thanh **Flow Inspector** để anh có thể soi trực tiếp từng bước bóc tách dữ liệu ngay lúc gõ phím.
> * **Tầng giữa là Golang Server:** Đây là 'bộ não' điều phối, tích hợp bộ tách từ của Cốc Cốc qua cầu nối **CGO Bridge**, phụ trách chuẩn hóa dữ liệu và điều phối luồng ghi kép.
> * **Tầng dưới cùng là Cụm lưu trữ kép:** Một bên là **Elasticsearch 8.11** đóng vai trò giải pháp doanh nghiệp quy mô lớn; và một bên là **Custom Go Engine** — động cơ nhị phân do chính tay em tự viết bằng Go thuần, lưu thẳng xuống đĩa cứng với mức tiêu thụ RAM siêu nhẹ chỉ **15MB**.*
> 
> *Nhờ mô hình này, hệ thống vừa có thể đối soát thực nghiệm khách quan, vừa làm chủ hoàn toàn công nghệ lưu trữ cấp thấp anh ạ."* `[Chuyển Slide 3]`

---

## 📌 SLIDE 3: TẦNG BÓC TÁCH NGÔN NGỮ TỰ NHIÊN (NLP PREPROCESSING VIA CGO)

* **Hình ảnh trên slide:** Sơ đồ CGO Bridge, Double-Array Trie và Thuật toán Viterbi HMM.
* **Nội dung hiển thị trên slide:**
  * **Double-Array Trie (DAT):** Nén toàn bộ từ điển tiếng Việt vào 2 mảng số nguyên tuyến tính $\rightarrow$ Tra cứu tức thời $O(L)$, dung lượng RAM siêu nhẹ chỉ **45 MB**.
  * **Giải thuật Viterbi HMM:** Giải quyết triệt để hiện tượng nhập nhằng ngữ nghĩa, tìm phương án ngắt từ tối ưu theo thời gian tuyến tính $O(N)$.
  * **CGO Bridge An Toàn Tuyệt Đối:**
    * Đồng bộ đa luồng bằng Mutex, phục vụ hàng ngàn truy vấn song song.
    * Thu hồi bộ nhớ C chủ động (`C.free`), ngăn chặn 100% rò rỉ RAM (Memory Leak).

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Trọng tâm xử lý ngôn ngữ của dự án nằm ở tầng bóc tách từ vựng. Thay vì dùng các thư viện Go thuần có từ điển nghèo nàn, em đã đưa bộ não **Cốc Cốc Tokenizer** viết bằng C++ vào Golang thông qua **CGO**. Ở tầng này có 2 điểm nhấn kỹ thuật quan trọng:
> 
> `[Nhấn mạnh vào 2 luận điểm]`  
> * **Thứ nhất là cấu trúc Double-Array Trie (DAT):** Toàn bộ từ điển từ ghép đồ sộ của Cốc Cốc được nén gọn vào đúng 2 mảng số nguyên tuyến tính. Dung lượng RAM chỉ tốn **45MB**, giúp CPU tra từ vựng tức thì ở độ phức tạp $O(L)$ ngay trên L1/L2 cache.
> * **Thứ hai là Thuật toán Viterbi HMM:** Thuật toán quy hoạch động này giúp cỗ máy giải quyết các câu nhập nhằng và tìm ra cách ngắt từ chuẩn xác nhất trong thời gian tuyến tính $O(N)$.*
> 
> *Về mặt kỹ thuật hệ thống, em đã bọc toàn bộ lời gọi CGO bằng `Mutex` và cơ chế giải phóng vùng nhớ C chủ động ngay sau khi xử lý. Do đó, hệ thống chạy đa luồng cực kỳ ổn định, không bao giờ xảy ra lỗi rò rỉ bộ nhớ anh nhé."* `[Chuyển Slide 4]`

---

## 📌 SLIDE 4: THIẾT KẾ DỮ LIỆU ĐA TRƯỜNG ("CHIẾC TỦ 3 NGĂN")

* **Hình ảnh trên slide:** Sơ đồ phân nhánh phễu: 1 tin nhắn chạy vào 3 trường dữ liệu độc lập.
* **Nội dung hiển thị trên slide:**
  * **Ngăn 1 - `content_tokenized`:** Lưu từ ghép chuẩn nối gạch dưới (`học_sinh`, `cà_phê`) $\rightarrow$ Khóa chặt ngữ nghĩa, ưu tiên số 1 cho người gõ đúng.
  * **Ngăn 2 - `content_unaccented`:** Lưu chuỗi không dấu (`hoc_sinh`, `ca_phe`) $\rightarrow$ Đón đầu người dùng chat nhanh trên điện thoại không gõ dấu.
  * **Ngăn 3 - `content_partial`:** Băm từ ghép thành các lát cắt tiền tố **Edge N-gram** 2-15 ký tự (`học`, `học_s`, `học_si`) $\rightarrow$ Phục vụ tìm kiếm tức thì khi đang gõ dở (Autocomplete).

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Khi chat, mỗi người lại có một thói quen gõ phím khác nhau: có anh gõ chuẩn có dấu, có anh gõ nhanh không dấu, lại có lúc đang gõ dở một nửa từ thì đã muốn thấy kết quả rồi. Để chiều lòng mọi thói quen này, em đã thiết kế cấu trúc dữ liệu như một **'Chiếc tủ 3 ngăn'** cho mỗi tin nhắn:
> 
> * * **Ngăn 1 - `content_tokenized`:** Lưu trữ từ ghép chuẩn mực được nối gạch dưới, ví dụ `học_sinh`, `cà_phê`. Đây là chốt chặn quan trọng nhất để khóa chặt ngữ nghĩa.
> * * **Ngăn 2 - `content_unaccented`:** Lưu chuỗi đã gọt sạch dấu tiếng Việt, ví dụ `hoc_sinh`, `ca_phe`, để người dùng gõ không dấu vẫn tìm thấy ngay.
> * * **Ngăn 3 - `content_partial`:** Băm từ ghép thành các lát cắt tiền tố **Edge N-gram** từ 2 đến 15 ký tự, ví dụ `học`, `học_s`, `học_si`. Đây chính là vũ khí giúp gợi ý tin nhắn ngay khi người dùng vừa đặt tay gõ vài ký tự đầu tiên.*
> 
> *Tách thành 3 ngăn độc lập như thế này chính là chìa khóa mở đường cho luồng tìm kiếm thông minh phía sau anh ạ."* `[Chuyển Slide 5]`

---

## 📌 SLIDE 5: TỔNG QUAN LUỒNG ĐÁNH CHỈ MỤC (INDEXING PIPELINE OVERVIEW)

* **Hình ảnh trên slide:** Sơ đồ dây chuyền 5 trạm xử lý liên hoàn của Luồng Đánh chỉ mục.
* **Nội dung hiển thị trên slide:**
  * **Dây chuyền 5 trạm xử lý liên hoàn:**
    * **Trạm 1:** Tiếp nhận tin nhắn thô qua REST API: `{"content": "Học sinh uống cà phê"}`.
    * **Trạm 2:** Bóc tách ngôn ngữ CGO: dán nhãn từ ghép `học_sinh` và `cà_phê`.
    * **Trạm 3:** Bộ nhân bản tạo chuỗi không dấu `hoc_sinh` và mạng lưới tiền tố N-gram.
    * **Trạm 4:** Bộ điều phối Dual-Write Dispatcher chia dữ liệu thành 2 ngả đường.
    * **Trạm 5:** Nhập kho song song vào Elasticsearch 8.11 và Động cơ nhị phân Custom Go.
  * **Hiệu năng:** Toàn bộ chu trình hoàn tất chỉ trong **1.78 mili-giây/tin nhắn**.

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Bây giờ, em xin trình bày về chiều nạp dữ liệu: Một tin nhắn từ lúc người dùng bấm nút Gửi sẽ đi qua những đâu để được 'nhập kho' tìm kiếm?
> 
> `[Chỉ tay lần lượt theo 5 trạm trên màn hình]`  
> Dữ liệu sẽ đi qua một dây chuyền 5 trạm khép kín:
> * Đầu tiên, Cổng API tiếp nhận gói tin thô.
> * Trạm thứ 2 đưa tin nhắn qua CGO để Cốc Cốc dán nhãn từ ghép thành `học_sinh` và `cà_phê`.
> * Trạm thứ 3 nhân bản biến thể thành chuỗi không dấu và các lát cắt tiền tố N-gram.
> * Trạm thứ 4 là bộ chia luồng Dual Dispatcher.
> * Và Trạm thứ 5 sẽ ghi đồng thời vào cả Elasticsearch lẫn Động cơ nhị phân Custom Go.*
> 
> *Toàn bộ quy trình nạp này chỉ tốn **1.78 mili-giây** cho mỗi tin nhắn. Và để anh thấy rõ từng micro-giây tương tác giữa các tiến trình kỹ thuật bên dưới, em mời anh cùng nhìn sâu vào Sơ đồ Tuần tự của luồng ghi này ở slide tiếp theo!"* `[Chuyển Slide 6]`

---

## 📌 SLIDE 6: DIỄN GIẢI CHI TIẾT SƠ ĐỒ TUẦN TỰ LUỒNG ĐÁNH CHỈ MỤC (THE LIFE OF A WRITE)

* **Hình ảnh trên slide:** Sơ đồ tuần tự chi tiết Luồng Ghi (Sequence Diagram - Dual Storage Ingestion).
* **Nội dung hiển thị trên slide:**
  * **Bước 1-4 (NLP Preprocessing):** Client gửi tin $\rightarrow$ CGO chạy Viterbi HMM $\rightarrow$ Normalizer gọt dấu và sinh Edge N-grams $\rightarrow$ Đóng gói Struct đa trường.
  * **Bước 5 (Khối PAR song song - Dual Storage):**
    * 🐘 **Elasticsearch 8.11:** Bắn gói NDJSON qua cổng 9200. Lucene ghi vào **Translog** chống mất điện và lưu vào RAM Buffer tạo Segment.
    * ⚡ **Custom Go Engine:** Ghi tuần tự vào nhật ký `wal.log` kèm mã băm **CRC32 IEEE**, cập nhật mục lục con trỏ `docstore.idx` cố định **12 bytes ($O(1)$)**, và đưa ngay vào `MemTable` trên RAM.
  * **Bước 6:** Phản hồi 201 Created cho Client, tin nhắn sẵn sàng tìm kiếm tức thì trên cả 2 động cơ.

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Đây là **Sơ đồ tuần tự 'Vòng đời của một thao tác Ghi'**, mô tả chi tiết từng hàm và tiến trình chạy ngầm:
> 
> `[Chỉ vào các bước 1 -> 4]`  
> * Khi người dùng gửi tin nhắn, Go Ingestion Service nhận diện và chuyển vùng nhớ sang C++ Cốc Cốc qua CGO để chạy thuật toán Viterbi. Sau đó, Normalizer gọt sạch dấu và băm các lát cắt Edge N-gram, đóng gói thành Document đa trường hoàn chỉnh.
> * `[Chỉ vào khối PAR song song]` **Điểm đắt giá nhất ở đây là cơ chế ghi song song 2 nhánh:**
>     * Nhánh Elasticsearch: Server bắn gói JSON sang cổng 9200. Lucene ghi ngay vào tệp nhật ký **Translog** để chống mất điện, rồi lưu vào RAM Buffer để định kỳ kết xuất thành Inverted Index.
>     * Nhánh Custom Go Engine: Server ghi tuần tự vào tệp `wal.log` kèm mã kiểm tra toàn vẹn **CRC32 IEEE 4 bytes**, cập nhật bảng mục lục con trỏ `docstore.idx` với kích thước cố định **đúng 12 bytes ($O(1)$)**, và đưa thẳng lên `MemTable` trên RAM.*
> 
> *Nhờ cách thiết kế bất đồng bộ này, tin nhắn vừa gửi xong là đã chốt an toàn xuống đĩa cứng và sẵn sàng tìm thấy ngay lập tức trên cả 2 động cơ mà không hề có độ trễ anh ạ!"* `[Chuyển Slide 7]`

---

## 📌 SLIDE 7: TỔNG QUAN LUỒNG TRUY VẤN & XẾP HẠNG (SEARCH PIPELINE OVERVIEW)

* **Hình ảnh trên slide:** Sơ đồ kiến trúc 3 Động cơ tìm kiếm chạy song song (Parallel Dispatcher).
* **Nội dung hiển thị trên slide:**
  * **Parallel Search Dispatcher:** Tiếp nhận query $\rightarrow$ Phóng 3 Goroutines tìm kiếm đồng thời:
    * 🟢 **Cột 1 - Cốc Cốc ES 8.11:** Bóc tách từ ghép C++ & Kích hoạt cơ chế Boosting 4 tầng điểm số.
    * 🔴 **Cột 2 - Baseline ES 8.11:** Dùng bộ phân tách mặc định theo khoảng trắng làm đối chứng.
    * 🟣 **Cột 3 - Custom Go Engine:** Tìm kiếm nhị phân trên RAM và Segments nhị phân.
  * **Result Aggregator:** Thu thập kết quả 3 bên, bấm giờ chính xác từng mili-giây và hiển thị đối chiếu trực tiếp.

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Ngược lại với luồng ghi, luồng tìm kiếm là nơi thể hiện rõ nhất sự thông minh của thuật toán và tốc độ xử lý của hệ thống.
> 
> `[Chỉ vào khối Dispatcher ở giữa]`  
> Khi người dùng nhập từ khóa tìm kiếm (ví dụ 'hoc sinh'), hệ thống không tra cứu tuần tự từng cái một mà khởi tạo **Parallel Search Dispatcher** để phóng ra 3 nhánh tìm kiếm song song cùng lúc:
> * Nhánh thứ nhất là **Cốc Cốc ES 8.11**: Phân tích từ ghép và kích hoạt bộ chấm điểm Boosting 4 tầng.
> * Nhánh thứ hai là **Baseline ES 8.11**: Chạy cơ chế tách từ mặc định theo dấu cách để làm mẫu đối chứng.
> * Nhánh thứ ba là **Custom Go Engine**: Chạy hoàn toàn bằng code Go thuần trực tiếp trên bộ nhớ.*
> 
> *Tất cả kết quả được gom về bộ tổng hợp, bấm giờ chính xác từng mili-giây để người dùng thấy rõ sự khác biệt giữa 3 cỗ máy. Em mời anh xem tiếp sơ đồ tuần tự chi tiết của luồng này ở slide sau!"* `[Chuyển Slide 8]`

---

## 📌 SLIDE 8: DIỄN GIẢI CHI TIẾT SƠ ĐỒ TUẦN TỰ LUỒNG TRUY VẤN (THE LIFE OF A SEARCH)

* **Hình ảnh trên slide:** Sơ đồ tuần tự chi tiết Luồng Tìm kiếm 3 Động cơ song song (Parallel Search Sequence).
* **Nội dung hiển thị trên slide:**
  * **Bước 1-2:** Nhận query `hoc sinh` $\rightarrow$ CGO nhận diện từ ghép `hoc_sinh` và chuẩn bị các biến thể.
  * **Bước 3 (3 Goroutines song song):**
    * 🟢 **Cốc Cốc ES:** Bool Query 4 tầng (Boost 5.0x có dấu, 4.0x cụm từ, 3.0x không dấu, 1.0x-10.0x n-gram). Lucene chấm điểm BM25 + Highlights (MRR 0.77).
    * 🔴 **Baseline ES:** Standard Match Query (bẻ theo dấu cách, BM25 không boost, dính bẫy từ ghép).
    * 🟣 **Custom Go Engine:** Binary Search $O(\log M)$ trên `terms.dict`, quét `postings.bin` và chấm điểm bằng Pure Go BM25 Ranker trong **2.8ms**.
  * **Bước 4-5:** `sync.WaitGroup` thu thập kết quả, đo độ trễ chi tiết và render bảng đối soát 3 cột lên UI.

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Đây là **Sơ đồ tuần tự 'Vòng đời của một truy vấn Tìm kiếm'**:
> 
> `[Chỉ tay theo các bước tương tác]`  
> * Khi người dùng gõ từ khóa `hoc sinh`, Go Search Service gọi CGO để Cốc Cốc nhận diện ngay đây là từ ghép `học_sinh` (hoặc `hoc_sinh`).
> * `[Chỉ vào 3 nhánh song song]` Sau đó, server dùng cơ chế đa luồng **Goroutines** để phóng 3 yêu cầu song song:
>     * Nhánh Cốc Cốc ES: Gửi câu truy vấn Bool Query đa tầng sang Elasticsearch. Tại đây, cỗ máy áp dụng **4 tầng trọng số Boosting**: Khớp cụm từ nhận Boost 4.0x; Khớp từ ghép có dấu nhận Boost 5.0x; Khớp không dấu nhận Boost 3.0x; và Khớp tiền tố gõ dở nhận Boost từ 1.0x đến 10.0x. Lucene tính điểm BM25 và trả về danh sách ứng viên chuẩn xác kèm vùng bôi sáng từ khóa.
>     * Nhánh Baseline ES: Chạy truy vấn Standard Match thông thường, trả về kết quả đối chứng bị dính đầy tin nhắn rác.
>     * Nhánh Custom Go Engine: Tìm kiếm nhị phân $O(\log M)$ trên tệp từ điển `terms.dict`, đọc danh sách Postings từ `postings.bin` và chấm điểm bằng bộ xếp hạng **Pure Go BM25 Ranker**.*
> * Bộ tổng hợp dùng `sync.WaitGroup` gom đủ 3 kết quả, đo đạc độ trễ và trả về giao diện bảng 3 cột cho người dùng.*
> 
> *Toàn bộ cuộc đua này diễn ra chỉ trong vài phần nghìn giây thôi anh ạ!"* `[Chuyển Slide 9]`

---

## 📌 SLIDE 9: CƠ CHẾ CHẤM ĐIỂM OKAPI BM25 – "BAN GIÁM KHẢO CÔNG TÂM"

* **Hình ảnh trên slide:** Sơ đồ công thức BM25 đơn giản hóa & Cơ chế Boosting 4 tầng.
* **Nội dung hiển thị trên slide:**
  * **Khung công thức trung tâm:** $\text{Final Score} = \text{BM25}(\text{Query}, \text{Message}) \times \text{Boost Weight}$
  * **3 Trụ cột chấm điểm BM25:**
    * **1. Độ quý hiếm (IDF):** Từ xuất hiện khắp nơi (*"và, thì, là"*) bị triệt tiêu điểm về $0$. Từ đặc trưng (*"học_sinh, cà_phê"*) được thưởng điểm rất cao.
    * **2. Hạn mức bão hòa tần suất (TF):** Tham số $k_1 = 1.2$ tạo trần bão hòa $\rightarrow$ Lặp lại 20 lần từ khóa cũng không được cộng thêm điểm $\rightarrow$ **Chống 100% spam từ khóa!**
    * **3. Phạt độ dài tài liệu:** Tham số $b = 0.75$ ưu tiên tin nhắn ngắn gọn, trúng trọng tâm hơn đoạn văn dài lan man.
  * **Hệ số nhân Boosting:** Từ ghép có dấu $\times 5.0$, Cụm từ $\times 4.0$, Không dấu $\times 3.0$, Gõ dở $\times 1.0 \to \times 10.0$.

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Nhiều người thắc mắc: **Tại sao thuật toán Okapi BM25 lại được coi là tiêu chuẩn vàng của ngành tìm kiếm thông tin hiện đại?**
> 
> `[Chỉ vào 3 khối trụ cột trên sơ đồ]`  
> Em hình dung BM25 như một vị **Ban giám khảo chấm thi chí công vô tư** với 3 nguyên tắc khoa học anh ạ:
> * **Nguyên tắc 1 - Độ quý hiếm từ vựng (IDF):** Những từ phổ thông xuất hiện ở mọi nơi như chữ 'và', 'thì', 'là' sẽ bị phạt điểm về 0. Nhưng từ đặc trưng mang ngữ nghĩa cao như 'cà phê', 'học sinh' sẽ được thưởng điểm rất lớn.
> * **Nguyên tắc 2 - Hạn mức bão hòa tần suất (TF):** Nếu có ai cố tình gõ lặp lại 20 lần chữ 'cà phê cà phê...' trong tin nhắn để 'hack' vị trí đầu bảng thì sẽ bị BM25 chặn đứng hoàn toàn. Với tham số bão hòa $k_1 = 1.2$, lặp lại nhiều lần cũng không được cộng thêm điểm $\rightarrow$ Vô hiệu hóa 100% hành vi spam từ khóa!
> * **Nguyên tắc 3 - Phạt công bằng theo độ dài:** Một tin nhắn ngắn 10 từ nói trúng trọng tâm sẽ được ưu tiên điểm cao hơn một đoạn văn dài 500 từ chỉ vô tình nhắc qua từ đó 1 lần.*
> 
> *Kết hợp điểm số BM25 khoa học này với hệ số nhân Boosting 4 tầng, tin nhắn đúng ý người dùng nhất luôn luôn được đưa lên vị trí số 1 tuyệt đối."* `[Chuyển Slide 10]`

---

## 📌 SLIDE 10: XỬ LÝ TRƯỜNG HỢP BIÊN: SỬA LỖI ĐẢO NGƯỢC ĐIỂM KHI GÕ DỞ

* **Hình ảnh trên slide:** `image/score_comparison_chart.jpg` (Biểu đồ cột: 33.36đ vọt lên 93.91đ).
* **Nội dung hiển thị trên slide:**
  * **Hiện tượng lỗi điểm số ngược (Inverse Scoring Anomaly):** Khi gõ dở `"học sin"`, máy tính cũ dùng toán tử `OR` mặc định $\rightarrow$ Tin nhắn chỉ chứa từ đơn `"học"` lặp lại nhiều lần được 37đ, đứng trên tin nhắn chứa đúng từ ghép `"học sinh"` (chỉ được 33.36đ)!
  * **Giải pháp kỹ thuật của em:**
    1. Cấu hình `search_analyzer` độc lập: Ngăn Lucene băm nhỏ query của người dùng.
    2. Ép buộc toán tử `AND`: Bắt buộc tài liệu phải chứa đủ các âm tiết cấu thành.
    3. Nâng hệ số Boosting tiền tố lên **10.0x**.
  * **Kết quả thực nghiệm:** Điểm tin nhắn chứa `"học sinh"` tăng vọt từ **33.36đ lên 93.91đ**, chiếm trọn **Top 1** tuyệt đối!

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Trong quá trình thực nghiệm chuyên sâu, em đã phát hiện và xử lý thành công một ca lỗi rất hóc búa: **Hiện tượng Đảo Ngược Điểm Số khi người dùng gõ dở từ**.
> 
> `[Chỉ vào cột 33.36 điểm]`  
> Khi người dùng mới gõ đến chữ **'học sin'**, máy tính mặc định dùng toán tử `OR` để tìm chữ 'học'. Hậu quả trớ trêu là: Những tin nhắn chỉ chứa từ đơn 'học' lặp đi lặp lại nhiều lần lại đạt **37 điểm** và leo lên đầu; trong khi tin nhắn chứa đúng từ ghép 'học sinh' chỉ được **33.36 điểm** và bị tụt lại phía sau!
> 
> `[Chỉ vào cột 93.91 điểm - Nhấn giọng]`  
> Em đã giải quyết dứt điểm nghịch lý này bằng 3 giải pháp: Thiết lập bộ Search Analyzer độc lập để ngăn Lucene băm nhỏ query, ép buộc toán tử `AND` trên các âm tiết cấu thành, và áp dụng siêu trọng số **Boost 10.0x** cho tiền tố từ ghép. Nhờ đó, điểm số của tin nhắn chứa 'học sinh' đã **tăng vọt từ 33.36 lên 93.91 điểm**, vươn lên chiếm trọn vị trí Top 1 ngay trước mắt người dùng anh ạ!"* `[Chuyển Slide 11]`

---

## 📌 SLIDE 11: ĐỘT PHÁ: ĐỘNG CƠ LƯU TRỮ NHỊ PHÂN THUẦN GO (CUSTOM ENGINE)

* **Hình ảnh trên slide:** `image/The Life of a Write Sequence Diagram.png` (Cấu trúc đĩa nhị phân WAL + DocStore + Segments).
* **Nội dung hiển thị trên slide:**
  * **Mục tiêu:** Thoát ly phụ thuộc Docker/Java, giảm RAM từ 1.200MB xuống **15MB**, độ trễ kỷ lục **~3ms**.
  * **Cấu trúc 4 tệp nhị phân trực tiếp trên ổ đĩa:**
    * **`docstore.idx` & `.dat` (Forward Index):** Bảng mục lục con trỏ cố định **12 bytes/doc** $\rightarrow$ Tra cứu tin nhắn gốc tức thời theo mốc byte $\text{Offset} = \text{DocID} \times 12$ ($O(1)$ tuyệt đối).
    * **`wal.log` (Write-Ahead Log):** Ghi tuần tự kèm mã kiểm tra toàn vẹn **CRC32 IEEE**, tự động Replay khôi phục 100% dữ liệu khi mất điện.
    * **`terms.dict` & `postings.bin` (Inverted Index Bất Biến):** Từ điển sắp xếp Alphabet A-Z tra cứu nhị phân $O(\log M)$.
    * **`tombstone.del` (Xóa mềm):** Cơ chế Soft Delete đảm bảo các luồng đọc song song an toàn không cần Lock (Lock-free).

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Không dừng lại ở việc dùng các thư viện có sẵn, em đặt ra một mục tiêu tham vọng hơn: **Tự tay xây dựng một Động cơ Lưu trữ và Tìm kiếm Nhị phân độc lập thuần Go**, thoát ly hoàn toàn khỏi sự phụ thuộc vào Java và Docker:
> 
> `[Chỉ vào 4 khối file trên sơ đồ]`  
> * **`docstore.idx` & `.dat`:** Em thiết kế bảng mục lục con trỏ cố định **đúng 12 bytes cho mỗi tin nhắn**. Việc tra cứu tin nhắn gốc đạt độ phức tạp **$O(1)$ tuyệt đối**, chỉ cần một phép nhân $\text{Offset} = \text{DocID} \times 12$ là con trỏ hệ điều hành nhảy thẳng tới mốc byte chứa dữ liệu.
> * **`wal.log`:** Mọi thao tác ghi đều tuần tự ghi nhật ký kèm mã niêm phong **CRC32 IEEE**, đảm bảo an toàn dữ liệu 100%, sẵn sàng khôi phục ngay nếu máy chủ bị sập nguồn đột ngột.
> * **`terms.dict` & `postings.bin`:** Từ điển được sắp xếp theo thứ tự bảng chữ cái để tra cứu nhị phân siêu tốc.
> * **`tombstone.del`:** Cơ chế xóa mềm bằng Bitmap giúp các luồng đọc song song không bao giờ bị nghẽn khóa.*
> 
> *Cỗ máy nhị phân này chỉ tiêu tốn **15MB RAM**, mang lại tốc độ phản hồi kỷ lục chỉ **~3 mili-giây** thôi anh ạ!"* `[Chuyển Slide 12]`

---

## 📌 SLIDE 12: ĐỐI CHIẾU KIẾN TRÚC LƯU TRỮ: ELASTICSEARCH 8.11 VS CUSTOM PURE-GO ENGINE

* **Hình ảnh trên slide:** Sơ đồ đối chiếu cơ chế lưu trữ nội tại (Storage Internals Diagram - Khối to, chữ lớn).
* **Nội dung hiển thị trên slide:**
  * **🐘 NỬA TRÁI: 5 THÀNH PHẦN NỘI TẠI CỦA ELASTICSEARCH 8.11 (LUCENE INTERNALS):**
    1. **RAM Indexing Buffer:** Vùng đệm gom tài liệu mới trên bộ nhớ JVM heap trước khi tạo Segment.
    2. **Translog (Write-Ahead Log):** Tệp nhật ký ghi tuần tự xuống đĩa ngay khi nhận tin, bảo vệ toàn vẹn dữ liệu chống mất điện trước khi có commit point.
    3. **Lucene Segments Bất Biến (Immutable Segments):**
       * `.tim` & `.tip` (Terms Dict & FST Index): Từ điển các token tra cứu siêu tốc.
       * `.doc` & `.pos` (Postings List & TF): Danh sách DocID và tần suất từ để tính BM25.
       * `.fdt` & `.fdx` (Stored Fields): Dữ liệu gốc của tin nhắn được nén bằng thuật toán LZ4.
    4. **Merge Policy (Chạy ngầm):** Tiến trình TieredMergePolicy tự động gộp các Segment nhỏ thành Segment lớn, dọn dẹp các bản ghi đã xóa mềm (Purge tombstones).
    5. **Docker Storage Volume (ext4 / XFS):** Điểm chốt Flush & Fsync dữ liệu bền vững xuống ổ cứng vật lý.
  * **⚡ NỬA PHẢI: 5 THÀNH PHẦN NỘI TẠI CỦA CUSTOM GO ENGINE (PURE GO BINARY DISK):**
    1. **`MemTable` trên RAM:** Vùng In-Memory Inverted Index giúp tìm thấy tin nhắn tức thì (0s latency) mà không cần đợi refresh.
    2. **`wal.log` (Write-Ahead Log):** Tệp ghi tuần tự nhị phân có Magic Header `WAL1` và mã băm **CRC32 IEEE** kiểm tra lỗi chống sập nguồn.
    3. **`docstore.idx` & `.dat` (Forward Index $O(1)$):**
       * `docstore.idx`: Bảng con trỏ cố định **12 bytes/doc** (`DocID 4B + Offset 8B`), nhảy trực tiếp bằng phép tính $\text{Offset} = \text{DocID} \times 12$.
       * `docstore.dat`: Lưu payload nội dung tin nhắn gốc.
    4. **`terms.dict` & `postings.bin` (Disk Segments Bất Biến):**
       * `terms.dict`: Từ điển các term sắp xếp A-Z để tra cứu nhị phân $O(\log M)$.
       * `postings.bin`: Danh sách DocID & TF để Pure Go BM25 Ranker tính điểm.
    5. **`tombstone.del` (Bitmap Soft Delete):** Xóa mềm bằng mảng bit, các luồng đọc song song không bị lock (Lock-free).
  * **Bảng so sánh tài nguyên:** Cần Java JVM, ngốn **~1.200MB RAM**, chạy Docker (ES) vs **Zero-Dependency, 15MB RAM**, 1 file thực thi duy nhất (Go Engine).

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Có thể anh sẽ tự hỏi: **Tại sao đã có Elasticsearch rất mạnh rồi mà em vẫn dày công tự làm Động cơ nhị phân thuần Go?**
> 
> `[Chỉ tay vào sơ đồ đối chiếu 2 nửa màn hình]`  
> Bức tranh so sánh trên màn hình sẽ trả lời câu hỏi đó qua việc đối chiếu trực diện từng thành phần nội tại của 2 thế giới lưu trữ:
> 
> * **Ở nửa bên trái - Elasticsearch 8.11:**
>   Đây là cỗ máy tìm kiếm phân tán tiêu chuẩn công nghiệp với 5 khối chức năng phối hợp rất chặt chẽ:
>   * `[Chỉ vào RAM Buffer & Translog]` Đầu tiên, tin nhắn mới gửi vào được gom trong **RAM Indexing Buffer** trên bộ nhớ JVM, đồng thời ghi ngay vào tệp nhật ký **Translog** để chống mất điện.
>   * `[Chỉ vào Lucene Segments]` Cứ mỗi 1 giây (chu kỳ refresh), dữ liệu từ RAM được kết xuất thành các khối **Lucene Segments bất biến** trên đĩa. Trong đó, tệp `.tim` lưu từ điển từ khóa dạng cây FST; tệp `.doc` lưu danh sách Posting để tính điểm BM25; và tệp `.fdt` nén dữ liệu gốc bằng thuật toán LZ4.
>   * `[Chỉ vào Merge Policy & Docker Volume]` Các Segment nhỏ sẽ được tiến trình **Merge Policy** chạy ngầm liên tục gộp lại thành Segment lớn và lưu bền vững xuống **Docker Storage Volume**. Cỗ máy này cực kỳ mạnh mẽ cho hệ thống lớn, nhưng nhược điểm cố hữu là chạy trên nền Java JVM cồng kềnh, ngốn tới hơn **1.2 GB RAM** và phụ thuộc vào Docker.
> 
> * `[Chỉ tay sang nửa bên phải - Nhấn giọng tự hào]`  
> * **Ở nửa bên phải - Custom Go Storage Engine do em tự viết:**
>   Em đã tái tạo lại toàn bộ tinh hoa kiến trúc đó nhưng dưới dạng mã Go thuần siêu nhẹ và tối ưu riêng cho ứng dụng chat:
>   * Thay vì phải đợi chu kỳ refresh 1 giây, em dùng **`MemTable`** trên RAM để tin nhắn vừa gửi xong là tìm thấy ngay tức thì 0 giây.
>   * Thay vì Translog Java cồng kềnh, em dùng tệp **`wal.log`** tuần tự kèm mã băm **CRC32 IEEE**, đảm bảo an toàn dữ liệu 100% khi sập nguồn.
>   * Thay vì Stored Fields LZ4 nén phức tạp, em tự thiết kế **`docstore.idx`** với bản ghi con trỏ cố định **đúng 12 bytes**. Muốn đọc tin nào, con trỏ hệ điều hành chỉ việc nhảy thẳng tới mốc byte bằng một phép nhân duy nhất $\text{Offset} = \text{DocID} \times 12$ đạt tốc độ **$O(1)$ tuyệt đối**.
>   * Và thay vì Lucene Segments nặng nề, em dùng cặp tệp nhị phân **`terms.dict`** sắp xếp A-Z để tra cứu nhị phân $O(\log M)$ và **`postings.bin`** để chấm điểm Pure Go BM25!*
> 
> *Toàn bộ động cơ Go này chỉ gói gọn trong một file thực thi duy nhất, **không phụ thuộc bất kỳ thư viện bên ngoài nào**, chỉ tốn **15MB RAM** và cho tốc độ tìm kiếm kỷ lục chỉ **3 mili-giây**!"* `[Chuyển Slide 13]`

---

## 📌 SLIDE 13: TỔNG KẾT: NHỮNG CON SỐ BIẾT NÓI & LỜI KẾT

* **Hình ảnh trên slide:** `image/benchmark_ir_comparison.jpg` (Biểu đồ 4 chỉ số vàng: MRR, NDCG@10, P@1, Latency).
* **Nội dung hiển thị trên slide:**
  * **Thực nghiệm đo đạc khoa học trên 131 tin nhắn chat thực tế:**
    * 🎯 **Precision@1 (Độ chuẩn xác Top 1):** Đạt **75.0%** (🚀 Tăng vọt **+28.6%** so với Baseline 58.3%).
    * 📈 **MRR (Mean Reciprocal Rank):** Đạt **0.7708** (🚀 Tăng **+23.3%**).
    * 🥇 **NDCG@10 (Chất lượng xếp hạng):** Chạm mức tối ưu **0.9472** (🚀 Tăng **+15.3%**).
    * ⚡ **Độ trễ tìm kiếm:** Custom Engine Go đạt kỷ lục **3.0 mili-giây** (nhanh gấp gần 5 lần Elasticsearch 14.2ms).
  * **Lời kết:** Dự án đã giải quyết trọn vẹn từ gốc rễ ngôn ngữ học tiếng Việt đến việc tự làm chủ công nghệ lưu trữ cấp thấp siêu tốc.

> 🎙️ **LỜI THOẠI THUYẾT TRÌNH:**
> 
> *"Kính thưa anh, tính khoa học và hiệu quả của đề tài được khẳng định qua bộ chỉ số chuẩn ngành Thu nhận Thông tin (IR) trên 131 tin nhắn chat thực tế:
> 
> `[Chỉ vào từng cột chỉ số trên biểu đồ]`  
> * **Độ chuẩn xác Top 1 (Precision@1):** Đạt **75.0%** — tăng vọt **28.6%** so với hệ thống Baseline cũ.
> * **Thứ hạng kết quả đúng (MRR):** Đạt **0.7708** — tăng **23.3%**.
> * **Chất lượng xếp hạng tổng thể (NDCG@10):** Đạt mức gần như lý tưởng **0.9472** — tăng **15.3%**.
> * **Đặc biệt về độ trễ:** Động cơ Custom Go đạt mức không tưởng: chỉ **3.0 mili-giây**, nhanh hơn gần 5 lần so với Elasticsearch!*
> 
> *Dự án của em đã giải quyết trọn vẹn bài toán từ gốc rễ ngôn ngữ học tiếng Việt, tối ưu hóa hai luồng dữ liệu, và chứng minh khả năng tự làm chủ công nghệ lưu trữ cấp thấp. Sau đây, em xin phép được mở màn hình Live Demo trực tiếp hệ thống cho anh xem nhé!"*

---

# 💻 PHẦN 2: KỊCH BẢN THUYẾT TRÌNH LIVE DEMO TRỰC TIẾP (5 PHÚT THỰC CHIẾN)

> **Chuẩn bị trước:** Mở sẵn terminal chạy `chat_server.exe` và trình duyệt vào `http://localhost:8080`.

---

### 🎬 MÀN 1: TỔNG QUAN GIAO DIỆN & TÍNH NĂNG CHAT REALTIME (0:00 - 0:45)
* **Thao tác:** Di chuột quanh màn hình giới thiệu cấu trúc 3 cột của giao diện Messenger Dark Mode.
* 🎙️ **Lời thoại:**
  > *"Em mời anh xem giao diện thực tế của ứng dụng Web Messenger mà em đã xây dựng. Màn hình được thiết kế theo phong cách Dark Mode hiện đại của Facebook Messenger, chia làm 3 khu vực:
  > * Bên trái là danh sách các hội thoại chat nhóm.
  > * Ở giữa là khung chat thời gian thực với các bong bóng tin nhắn gradient.
  > * Và bên phải là bảng điều khiển kỹ thuật với 2 chế độ: **Flow Inspector** để soi luồng bóc tách dữ liệu, và tab **Compare 3 Engines** để so sánh trực diện 3 động cơ tìm kiếm anh nhé."*

---

### 🎬 MÀN 2: DEMO LUỒNG GHI & FLOW INSPECTOR SOI TOKEN (0:45 - 1:45)
* **Thao tác:**
  1. Chọn tab **Flow Inspector** ở cột bên phải.
  2. Tại khung chat giữa, gõ tin nhắn: `Học sinh lớp 12 thích uống cà phê sữa đá`
  3. Bấm **Enter** để gửi.
  4. Chỉ chuột sang cột Flow Inspector xem kết quả bóc tách vừa cập nhật tức thì.
* 🎙️ **Lời thoại:**
  > *"Bây giờ, em xin phép gửi một tin nhắn mới: 'Học sinh lớp 12 thích uống cà phê sữa đá'.  
  > `[Bấm Enter]`  
  > 
  > Anh quan sát thanh Flow Inspector bên phải vừa cập nhật tức thì này:
  > * **Mục 1 (Tokenized):** Cốc Cốc Tokenizer đã dán chặt các từ ghép lại thành `học_sinh`, `cà_phê`, `sữa_đá`.
  > * **Mục 2 (Unaccented):** Hệ thống tự động sinh chuỗi không dấu `hoc_sinh`, `ca_phe`, `sua_da`.
  > * **Mục 3 (Edge N-gram):** Hàng loạt lát cắt tiền tố như `học`, `học_s`, `cà`, `cà_p` đã được sinh ra và ghi đồng thời vào cả Elasticsearch lẫn Custom Go Engine. Toàn bộ quá trình nạp dữ liệu hoàn tất trong **chưa đầy 2 mili-giây** anh ạ!"*

---

### 🎬 MÀN 3: DEMO "BẪY TỪ GHÉP" VÀ ĐỐI SOÁT 3 CHIỀU (1:45 - 3:00) — ⭐ *CAO TRÀO*
* **Thao tác:**
  1. Chuyển cột bên phải sang tab **Compare (Bảng Đối Soát 3 Động Cơ)**.
  2. Bấm vào thanh tìm kiếm ở góc trên, gõ: `học sinh`
  3. Nhấn Tìm kiếm để 3 cột màu Đỏ - Xanh - Tím xuất hiện.
* 🎙️ **Lời thoại:**
  > *"Bây giờ là phần kiểm chứng quan trọng nhất: Chúng ta sẽ đi tìm từ khóa **'học sinh'**.  
  > `[Chỉ con trỏ vào Cột 1 màu Đỏ - Baseline]`  
  > Anh nhìn vào Cột màu đỏ của hệ thống Baseline dùng bộ tách từ chuẩn phương Tây: Kết quả bị dính đầy các tin nhắn rác như: 'chúc mừng **sinh** nhật', 'đón tân **sinh** viên'... Đây chính là 'Cái bẫy từ ghép' làm loãng hoàn toàn thông tin người dùng cần tìm!
  > 
  > `[Chỉ con trỏ sang Cột 2 màu Xanh lá - Cốc Cốc ES]`  
  > Nhưng nhìn sang Cột màu xanh của Cốc Cốc Elasticsearch: 100% tin nhắn rác đã bị lọc sạch! Hệ thống chỉ giữ lại đúng những tin nhắn nói về 'học sinh', với điểm số BM25 được nhân hệ số Boosting $\times 5.0$, xếp trang trọng ở Top 1.
  > 
  > `[Chỉ con trỏ sang Cột 3 màu Tím - Custom Go Engine]`  
  > Và đặc biệt, anh nhìn sang Cột màu tím của Động cơ nhị phân Custom Go do em tự viết: Kết quả trả về chuẩn xác tương đương, nhưng độ trễ tìm kiếm chỉ là **2.8 mili-giây** — nhanh hơn gần 5 lần so với Elasticsearch chạy trên Docker anh nhé!"*

---

### 🎬 MÀN 4: DEMO TÌM KHÔNG DẤU & SỬA LỖI ĐẢO NGƯỢC ĐIỂM KHI GÕ DỞ (3:00 - 4:15)
* **Thao tác:**
  1. Gõ từ khóa không dấu: `uong ca phe` $\rightarrow$ Chỉ vào kết quả chuẩn xác.
  2. Xóa và gõ dở từng chữ từ từ: `h` $\rightarrow$ `học` $\rightarrow$ `học s` $\rightarrow$ `học sin`
  3. Dừng ở chữ `học sin`, chỉ vào điểm số Top 1 đạt `93.91`.
* 🎙️ **Lời thoại:**
  > *"Tiếp theo, em gõ thử một từ khóa hoàn toàn không có dấu: `uong ca phe`.  
  > Như anh thấy, nhờ trường `content_unaccented` với hệ số Boost 3.0x, hệ thống vẫn bắt trúng đích tin nhắn có dấu ban đầu.
  > 
  > Bây giờ, em biểu diễn bài toán gõ dở từng ký tự: `h`... `học`... `học s`... `học sin`.  
  > `[Chỉ vào điểm số 93.91]`  
  > Ở các hệ thống thông thường, khi mới gõ đến 'học sin', tin nhắn chỉ chứa từ 'học' sẽ leo lên đầu do dính toán tử OR. Nhưng tại đây, nhờ giải pháp ép toán tử AND và nâng hệ số Boost 10.0x cho tiền tố từ ghép, tin nhắn chứa 'học sinh' vẫn giữ vững vị trí Top 1 với điểm số áp đảo **93.91 điểm**!"*

---

### 🎬 MÀN 5: DEMO TRẢI NGHIỆM ĐIỀU HƯỚNG TƯƠNG TÁC (4:15 - 5:00)
* **Thao tác:** Nhấp chuột vào một tin nhắn trong bảng kết quả $\rightarrow$ Quan sát khung chat cuộn mượt (Smooth Scroll) và bong bóng chat chớp sáng (Pulse Highlight).
* 🎙️ **Lời thoại:**
  > *"Cuối cùng là trải nghiệm liền mạch dành cho người dùng: Khi em bấm vào bất kỳ kết quả nào trong bảng đối soát...  
  > `[Bấm chuột vào 1 kết quả]`  
  > Khung chat trung tâm sẽ tự động cuộn mượt mà đến đúng vị trí tin nhắn đó trong quá khứ, đồng thời bong bóng chat sẽ nhấp nháy phát sáng để người dùng nắm bắt ngay ngữ cảnh của cuộc trò chuyện mà không mất công tìm kiếm bằng mắt.
  > 
  > Toàn bộ hệ thống — từ thuật toán ngôn ngữ học, kiến trúc lưu trữ nhị phân, đến trải nghiệm người dùng cuối — đã phối hợp nhịp nhàng và chứng minh tính ứng dụng thực tiễn vượt trội.
  > 
  > Em xin kết thúc phần trình bày và demo tại đây. Em rất mong nhận được những câu hỏi và đóng góp quý báu từ anh ạ!"*
