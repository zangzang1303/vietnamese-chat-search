# Tóm Tắt Chi Tiết: Nền Tảng Truy Vấn Thông Tin (Information Retrieval) - Chương 6, 7

> **Tài liệu gốc:** *An Introduction to Information Retrieval* (Christopher D. Manning, Prabhakar Raghavan, Hinrich Schütze - Cambridge University Press)

---

## CHƯƠNG 6: Scoring, Term Weighting and the Vector Space Model (Chấm Điểm, Trọng Số Thuật Ngữ Và Mô Hình Không Gian Vector)

### 1. Tại Sao Cần Ranked Retrieval (Tìm Kiếm Xếp Hạng)?
* **Hạn chế của Boolean Retrieval:**
  * **Hiện tượng Feast or Famine:** Câu lệnh `AND` quá chặt dễ trả về 0 kết quả; câu lệnh `OR` quá lỏng dễ trả về hàng chục nghìn kết quả.
  * **Không phân biệt mức độ liên quan:** Mọi tài liệu thỏa mãn logic đều có giá trị như nhau.
* **Mục tiêu của Ranked Retrieval:** Nhận câu truy vấn ngôn ngữ tự nhiên (Free-text query) và trả về danh sách **Top $K$** tài liệu có điểm số phù hợp (Relevance Score) cao nhất.

---

### 2. Trọng Số Tần Suất Thuật Ngữ (Term Frequency - TF)
* **Mô hình Túi Đựng Từ (Bag of Words Model):** Bỏ qua hoàn toàn thứ tự từ ngữ và ngữ pháp, chỉ quan tâm đến số lần xuất hiện của thuật ngữ trong tài liệu ($tf_{t,d}$).
* **Log-frequency Weighting:** Mức độ liên quan không tăng tuyến tính theo số lần xuất hiện (tài liệu xuất hiện 10 lần một từ không có nghĩa là liên quan gấp 10 lần tài liệu xuất hiện 1 lần). Vì vậy sử dụng hàm logarit:
  $$w_{t,d} = \begin{cases} 1 + \log_{10}(tf_{t,d}) & \text{nếu } tf_{t,d} > 0 \\ 0 & \text{ngược lại} \end{cases}$$

---

### 3. Nghịch Đảo Tần Suất Tài Liệu (Inverse Document Frequency - IDF)
* **Document Frequency ($df_t$):** Số lượng tài liệu trong toàn bộ kho chứa thuật ngữ $t$.
  * Phân biệt với $cf_t$ (Collection Frequency - tổng số lần $t$ xuất hiện trong kho). $df_t$ phản ánh độ phân biệt thông tin chính xác hơn nhiều.
* **Công thức IDF:**
  $$idf_t = \log_{10}\left(\frac{N}{df_t}\right)$$
  *(Trong đó $N$ là tổng số tài liệu của toàn hệ thống).*
* **Ý nghĩa:**
  * Từ quá phổ biến (như stop words, xuất hiện ở mọi tài liệu $df_t = N$) $\rightarrow idf_t = \log(1) = 0$.
  * Từ hiếm (chỉ xuất hiện ở số ít tài liệu) $\rightarrow idf_t$ cao, đóng góp điểm số quyết định.

---

### 4. Trọng Số TF-IDF
Kết hợp cả tần suất trong tài liệu và độ hiếm toàn cục:
$$w_{t,d} = (1 + \log_{10} tf_{t,d}) \times \log_{10}\left(\frac{N}{df_t}\right)$$
* Trọng số đạt cực đại khi từ xuất hiện nhiều lần trong một tài liệu ngắn và hiếm gặp trong toàn bộ kho tài liệu.

---

### 5. Mô Hình Không Gian Vector (Vector Space Model) & Cosine Similarity
* **Biểu diễn Vector:** Mỗi tài liệu $d$ và truy vấn $q$ được xem là một vector trong không gian $|V|$ chiều ($|V|$ là kích thước từ điển):
  $$\vec{V}(d) = (w_{t_1, d}, w_{t_2, d}, \dots, w_{t_{|V|}, d})$$
* **Chuẩn Hóa Độ Dài (Length Normalization - $L_2$ Norm):**
  Chia vector cho độ dài Euclid để loại bỏ sự thiên vị đối với các tài liệu dài:
  $$\vec{v} = \frac{\vec{V}}{\|\vec{V}\|} = \frac{\vec{V}}{\sqrt{\sum_{i=1}^{|V|} w_i^2}}$$
* **Độ Tương Đồng Cosine (Cosine Similarity):**
  Tính góc $\theta$ giữa hai vector. Góc càng nhỏ $\rightarrow \cos(\theta)$ càng tiến gần 1 $\rightarrow$ Mức độ tương đồng càng cao:
  $$\text{Sim}(q, d) = \cos(\theta) = \frac{\vec{V}(q) \cdot \vec{V}(d)}{\|\vec{V}(q)\| \|\vec{V}(d)\|} = \sum_{t \in q \cap d} \frac{w_{t,q}}{\|\vec{V}(q)\|} \cdot \frac{w_{t,d}}{\|\vec{V}(d)\|}$$

---

### 6. Hệ Thống Ký Hiệu SMART (SMART Notation)
Quy ước 3 ký tự `ddd.qqq` thể hiện cách tính trọng số:
* Ký tự 1: Biến thể TF (`n`: raw, `l`: log, `a`: augmented).
* Ký tự 2: Biến thể IDF (`n`: none, `t`: idf).
* Ký tự 3: Biến thể Normalization (`n`: none, `c`: cosine).
* Ví dụ: `ltc.lnc` (Tài liệu dùng log-tf, idf, cosine; Truy vấn dùng log-tf, no-idf, cosine).

---

## CHƯƠNG 7: Computing Scores in a Complete Search System (Tính Toán Điểm Số Trong Hệ Thống Hoàn Chỉnh)

### 1. Nút Thắt Hiệu Năng & Bài Toán "Inexact Top K"
* **Thách thức:** Với kho hàng triệu tài liệu, việc tính điểm Cosine chính xác cho mọi tài liệu chứa ít nhất một từ trong query là bất khả thi về mặt thời gian phản hồi (dưới 100ms).
* **Bản chất thực tế:** Người dùng chỉ quan tâm đến **Top 10 / Top 20** kết quả đầu tiên.
* **Chiến lược:** Sử dụng các kỹ thuật xấp xỉ (Inexact Top K Heuristics) để lọc nhanh các ứng viên sáng giá nhất mà không cần quét toàn bộ tập dữ liệu.

---

### 2. Các Kỹ Thuật Tối Ưu Hóa & Lấy Top K Hiệu Năng Cao

#### A. Champion Lists (Danh Sách Tinh Hoa)
* Với mỗi term $t$, lưu trước một danh sách phụ gồm $r$ tài liệu có trọng số $w_{t,d}$ cao nhất ($r \approx 100 - 1000$).
* Khi tìm kiếm, chỉ cần duyệt hợp các Champion Lists của các query terms. Chỉ khi không gom đủ $K$ kết quả mới quét đến Posting List đầy đủ.

#### B. Static Quality Scores & Ordered Postings (Điểm Uy Tín Cố Định)
* Mỗi tài liệu có điểm chất lượng tĩnh $g(d)$ độc lập với truy vấn (PageRank, lượt thích, độ tin cậy, thời gian đăng).
* Hàm điểm kết hợp:
  $$\text{NetScore}(d, q) = g(d) + \alpha \cdot \text{Cosine}(d, q)$$
* **Early Termination (Dừng sớm):** Sắp xếp Postings List theo $g(d)$ giảm dần thay vì theo DocID. Khi duyệt, nếu $g(d)$ giảm xuống dưới ngưỡng, dừng duyệt ngay lập tức vì các tài liệu phía sau không có cơ hội lọt vào Top $K$.

#### C. Impact-Ordered Postings
* Sắp xếp Postings List theo tần suất từ $tf_{t,d}$ giảm dần.
* Dừng duyệt khi $tf_{t,d}$ giảm xuống mức thấp.

#### D. Cluster Pruning (Tỉa Cụm)
1. Chọn ngẫu nhiên $\sqrt{N}$ tài liệu làm **Leaders** (Trưởng cụm).
2. Phân chia $N - \sqrt{N}$ tài liệu còn lại (**Followers**) vào Leader gần nhất.
3. Khi có query $q$: So sánh $q$ với $\sqrt{N}$ Leaders để tìm cụm phù hợp nhất, sau đó chỉ chấm điểm các tài liệu trong cụm đó. Giảm độ phức tạp từ $O(N)$ xuống $O(\sqrt{N})$.

#### E. Tiered Indexes (Chỉ Mục Phân Tầng)
* Tách Inverted Index thành nhiều tầng: Tier 1 (Postings có TF cao), Tier 2 (TF trung bình), Tier 3 (TF thấp).
* Tìm kiếm tại Tier 1 trước. Nếu đủ $K$ kết quả thì trả về ngay, không cần đọc Tier 2 và 3.

---

### 3. Khoảng Cách Gần Giữa Các Từ Truy Vấn (Query-Term Proximity)
* Vector Space Model bỏ qua vị trí giữa các từ.
* Trong thực tế, tài liệu chứa các từ query nằm sát cạnh nhau (cùng 1 câu) luôn liên quan hơn tài liệu chứa các từ nằm rải rác ở đầu và cuối trang.
* Hệ thống IR tính thêm điểm thưởng dựa trên kích thước cửa sổ (Proximity Window) nhỏ nhất chứa toàn bộ các từ của query.

---

### 4. Kiến Trúc Hoàn Chỉnh Của Một Search Engine Hiện Đại
1. **Query Parser:** Bóc tách cú pháp, kiểm tra chính tả, mở rộng từ đồng nghĩa.
2. **Tiered Index Lookup:** Tra cứu phân tầng, khai thác Champion Lists.
3. **Scoring & Ranking Engine:** Kết hợp BM25/Vector Space + Proximity + Static Quality Score ($g(d)$).
4. **Snippet Generation & Highlighting:** Trích xuất đoạn văn chứa từ khóa và bôi đậm.
5. **Distributed Coordination:** Phân tán dữ liệu qua nhiều Shard và tổng hợp kết quả (Scatter-Gather).
