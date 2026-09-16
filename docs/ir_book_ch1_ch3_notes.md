# Tóm Tắt Chi Tiết: Nền Tảng Truy Vấn Thông Tin (Information Retrieval) - Chương 1, 2, 3

> **Tài liệu gốc:** *An Introduction to Information Retrieval* (Christopher D. Manning, Prabhakar Raghavan, Hinrich Schütze - Cambridge University Press)

---

## CHƯƠNG 1: Boolean Retrieval (Mô Hình Truy Vấn Boolean)

### 1. Bài Toán Truy Vấn Thông Tin & Hạn Chế Của Grep
* **Khái niệm IR:** Tìm kiếm các tài liệu có bản chất phi cấu trúc (văn bản) đáp ứng nhu cầu thông tin từ tập bộ sưu tập quy mô lớn.
* **Hạn chế của Grep (Quét tuyến tính):**
  * Tốc độ quá chậm: $O(N)$ thời gian cho $N$ tài liệu khi tập dữ liệu lớn.
  * Không hỗ trợ toán tử vị trí/khoảng cách (ví dụ: `Romans NEAR countrymen`).
  * Không xếp hạng (rank) kết quả theo độ liên quan.
* **Ma trận Xuất hiện (Term-Document Incidence Matrix):**
  * Ánh xạ giá trị nhị phân `1` (có xuất hiện) hoặc `0` (không xuất hiện) giữa thuật ngữ (term) và tài liệu (document).
  * **Đặc tính thưa (Sparsity):** Trong thực tế, ma trận có độ thưa rất cao (thường mỏng hơn 99.8% toàn số 0).

#### Sơ đồ Ma trận Xuất hiện Term-Document:
```text
                  Antony and Cleopatra   Julius Caesar   Hamlet   Othello   Macbeth   Tempest
Brutus                     1                   1            0        1         0         0
Caesar                     1                   1            0        1         1         1
Calpurnia                  0                   1            0        0         0         0
```

#### Xử lý logic TRUY VẤN BOOLEAN (Bitwise Operation):
Để xử lý truy vấn: `Brutus AND Caesar AND NOT Calpurnia`, hệ thống lấy các vectơ hàng tương ứng và thực hiện phép toán bitwise:
```text
  Vectơ Brutus:      1  1  0  1  0  0
  Vectơ Caesar:      1  1  0  1  1  1
AND NOT Calpurnia:   1  0  1  1  1  1   (Đảo bit của 0 1 0 0 0 0)
----------------------------------------
KẾT QUẢ:            1  0  0  1  0  0   => Khớp với: Antony and Cleopatra & Hamlet
```

---

### 2. Cấu Trúc Chỉ Mục Đảo (Inverted Index) & Các Bước Khởi Tạo
Vì ma trận xuất hiện rất thưa, hệ thống IR chỉ lưu trữ các vị trí có giá trị `1` thông qua cấu trúc **Chỉ mục đảo**.

#### Cấu trúc gồm 2 thành phần chính:
1. **Dictionary (Từ điển / Lexicon):** Tập hợp tất cả các thuật ngữ duy nhất, thường được lưu trên bộ nhớ RAM.
2. **Postings List:** Danh sách chứa các `docID` mà thuật ngữ xuất hiện, được sắp xếp tăng dần và lưu trên đĩa (Disk).

```text
DICTIONARY (RAM)             POSTINGS LISTS (DISK)
+-----------+--------+      +---+   +---+   +---+   +----+   +----+
| Term      |  df    | ---> | 1 |-> | 2 |-> | 4 |-> | 11 |-> | 31 | ...
+-----------+--------+      +---+   +---+   +---+   +----+   +----+
| Brutus    |   8    | --->
| Caesar    |   8    | --->
| Calpurnia |   4    | --->
+-----------+--------+
```

#### Quy trình 4 bước xây dựng chỉ mục đảo:
1. **Collect Documents:** Thu thập các văn bản đầu vào.
2. **Tokenize:** Tách chuỗi văn bản thành danh sách các token.
3. **Linguistic Preprocessing:** Tiền xử lý ngôn ngữ (chuyển chữ thường, chuẩn hóa) tạo ra các Term.
4. **Sort & Group:** Sắp xếp danh sách cặp `(term, docID)` theo bảng chữ cái, gộp các docID trùng và xây dựng Postings List.

---

### 3. Xử Lý Truy Vấn Boolean & Tối Ưu Hóa (Query Optimization)

#### A. Thuật toán Giao hai danh sách Postings (Postings Intersection)
Sử dụng phương pháp hai con trỏ duyệt đồng thời hai danh sách postings đã được sắp xếp tăng dần.

```python
INTERSECT(p1, p2):
    answer = []
    while p1 != NIL and p2 != NIL:
        if docID(p1) == docID(p2):
            ADD(answer, docID(p1))
            p1 = next(p1)
            p2 = next(p2)
        elif docID(p1) < docID(p2):
            p1 = next(p1)
        else:
            p2 = next(p2)
    return answer
```
* **Độ phức tạp thời gian:** $O(x + y)$ với $x, y$ lần lượt là độ dài của hai danh sách postings.

#### B. Tối ưu hóa Truy vấn (Query Optimization)
* **Heuristic cốt lõi:** Luôn xử lý các term theo **thứ tự tần suất xuất hiện tài liệu (Document Frequency - $df$) tăng dần** (từ thuật ngữ ít xuất hiện nhất đến phổ biến nhất).
* **Lợi ích:** Giúp danh sách kết quả trung gian trong bộ nhớ luôn có kích thước nhỏ nhất, giảm tối đa số phép so sánh.

```python
INTERSECT(<t1, t2, ..., tn>):
    terms = SORTBYINCREASINGFREQUENCY(<t1, t2, ..., tn>)
    result = postings(first(terms))
    terms = rest(terms)
    
    while terms != NIL and result != NIL:
        result = INTERSECT(result, postings(first(terms)))
        terms = rest(terms)
        
    return result
```

> **💬 Lời bình của tôi:** Chỗ sắp xếp theo tần suất xuất hiện $df$ tăng dần này rất thông minh. Giả sử tìm `thời_khóa_biểu AND học`, từ `thời_khóa_biểu` hiếm hơn nhiều nên kết quả trung gian lọc ra ngay chỉ 1-2 tài liệu, không mất công duyệt hàng nghìn tin nhắn chứa từ `học`. Hôm qua mình code chưa có bước sắp xếp này, cần lưu ý tối ưu sau.

---

## CHƯƠNG 2: The Term Vocabulary and Postings Lists (Tập Từ Vựng Và Danh Sách Postings)

### 1. Xác Định Tập Từ Vựng (Determining the Vocabulary)
* **Indexing Granularity (Độ mịn chỉ mục):** Chọn đơn vị tài liệu (cả cuốn sách, một chương, hay một đoạn văn) tạo ra sự đánh đổi giữa **Precision** và **Recall**.
* **Phân biệt Token, Type và Term:**
  * **Token:** Một thể hiện chuỗi ký tự cụ thể trong tài liệu.
  * **Type:** Lớp chứa tất cả các token có cùng chuỗi ký tự.
  * **Term:** Một type đã qua chuẩn hóa và đưa vào từ điển của hệ thống.
* **Stop Words (Từ dừng):** Loại bỏ các từ cực kỳ phổ biến nhưng ít giá trị phân biệt thông tin dựa trên **Collection Frequency** (ví dụ: *the, and, is*).
* **Normalization, Stemming & Lemmatization:**
  * **Normalization:** Tạo lớp tương đương cho các dạng viết khác nhau (ví dụ: `U.S.A.` $\rightarrow$ `USA`, lowercasing).
  * **Stemming:** Sử dụng quy tắc thô chặt bỏ đuôi từ (ví dụ: *Porter Stemmer* biến *automation, automatic* $\rightarrow$ *automat*).
  * **Lemmatization:** Sử dụng từ điển và phân tích hình thái học để đưa từ về dạng nguyên thể chuẩn (Lemma) (ví dụ: *saw* $\rightarrow$ *see*).

> **💬 Lời bình của tôi:** Đoạn này giải thích chuẩn bài toán mình đang gặp: tiếng Anh chỉ cần tách theo space là ra từ đơn, nhưng tiếng Việt nếu coi mỗi space là 1 token thì chết chắc ở từ ghép. Thảo nào Elasticsearch mặc định không dùng được cho chat tiếng Việt nếu thiếu tokenizer chuyên dụng như Cốc Cốc.

---

### 2. Con Trỏ Nhảy (Skip Pointers)
Nhằm vượt qua độ phức tạp tuyến tính $O(m + n)$ khi thực hiện phép giao AND, **Skip Pointers** là các con trỏ "nhảy tắt" được bổ sung vào danh sách postings ngay từ thời điểm đánh chỉ mục.

#### Sơ đồ Danh sách Postings có Skip Pointers:
```text
Brutus: ------16-----> ------28-----> ------72----->
           |                   |                  |                   |
           v                   v                  v                   v
          (2) -> (4) -> (8) -> (16) -> (19) -> (23) -> (28) -> (43) -> (72)

Caesar: ------5------> ------51-----> ------98----->
           |                  |                  |                   |
           v                  v                  v                   v
          (1) -> (2) -> (3) -> (5) -> (8) -> (41) -> (51) -> (60) -> (71) -> (98)
```

#### Mã giả Thuật toán Giao Postings với Skip Pointers:
```python
INTERSECTWITHSKIPS(p1, p2):
    answer = []
    while p1 != NIL and p2 != NIL:
        if docID(p1) == docID(p2):
            ADD(answer, docID(p1))
            p1 = next(p1)
            p2 = next(p2)
        elif docID(p1) < docID(p2):
            if hasSkip(p1) and docID(skip(p1)) <= docID(p2):
                while hasSkip(p1) and docID(skip(p1)) <= docID(p2):
                    p1 = skip(p1)
            else:
                p1 = next(p1)
        else:
            if hasSkip(p2) and docID(skip(p2)) <= docID(p1):
                while hasSkip(p2) and docID(skip(p2)) <= docID(p1):
                    p2 = skip(p2)
            else:
                p2 = next(p2)
    return answer
```
* **Quy tắc đặt con trỏ:** Với danh sách postings độ dài $P$, đặt $\sqrt{P}$ con trỏ nhảy cách đều nhau.

---

### 3. Truy Vấn Cụm Từ (Phrase Queries) & Xử Lý Vị Trí

#### A. Biword Indexes (Chỉ mục Từ đôi)
* Coi mỗi cặp từ liên tiếp là một term (ví dụ: *"Friends, Romans, Countrymen"* $\rightarrow$ biwords: `friends romans`, `romans countrymen`).
* Hạn chế: Kích thước từ điển bùng nổ và có thể sinh ra dương tính giả đối với các cụm từ dài hơn 2 từ.

#### B. Positional Indexes (Chỉ mục Vị trí)
* Với mỗi term, danh sách postings lưu thêm thông tin vị trí các token xuất hiện trong tài liệu dạng `docID: <pos1, pos2, ...>`.

```text
DICTIONARY          POSTINGS LIST WITH POSITIONS
+-------+-----+     +-------------------------------------------------------+
| to    | 993 | --> | doc 1: <7, 18, 33, 72, 86, 231>;                     |
| be    | 178 | --> | doc 4: <17, 191, 291, 430, 434>;                      |
+-------+-----+     +-------------------------------------------------------+
```

* **Proximity Search (`/k` operator):** Tìm kiếm hai từ xuất hiện cách nhau trong khoảng $k$ vị trí. Positional Index giải quyết triệt để bài toán này với độ phức tạp thời gian $\Theta(T)$ ($T$ là tổng số token trong toàn bộ tập tài liệu).

> **💬 Lời bình của tôi:** Lúc làm `pkg/invertedindex` hôm trước mình có lưu mảng `Positions: []int` trong Posting mà chưa hình dung hết ứng dụng ngoài việc demo. Đọc đến đây mới sáng tỏ: nhờ có vị trí từng từ mà hệ thống mới hỗ trợ tìm cụm từ chính xác dạng `"cà phê"` hoặc tìm 2 từ cách nhau tối đa $k$ chữ.

---

## CHƯƠNG 3: Dictionaries and Tolerant Retrieval (Từ Điển Và Tìm Kiếm Dung Lỗi)

### 1. Cấu Trúc Dữ Liệu Tra Cứu Từ Điển
* **Bảng Băm (Hashing):** Tra cứu trung bình $O(1)$, nhưng không hỗ trợ tìm kiếm tiền tố (prefix search) hay truy vấn khoảng.
* **Cây Tìm Kiếm (B-Tree / Search Trees):** Hỗ trợ tra cứu theo thứ tự bảng chữ cái, tìm kiếm tiền tố và khoảng. B-Tree tối ưu hóa cho lưu trữ đĩa bằng cách gộp nhiều mức của cây nhị phân vào một nút đĩa.

---

### 2. Truy Vấn Ký Tự Đại Diện (Wildcard Queries)

#### A. Permuterm Index (Chỉ mục Xoay)
* Thêm ký tự `$`. Tạo ra tất cả các bản xoay vòng của từ sao cho ký tự đại diện `*` luôn ở cuối chuỗi.
* *Ví dụ với từ `hello$`:*
```text
PERMUTERM VOCABULARY               ORIGINAL TERM
+------------+
|  hello$    | --------------------->
|  ello$h    | --------------------->   hello
|  llo$he    | --------------------->
|  lo$hel    | --------------------->
+------------+
```
* Biến đổi truy vấn **`m*n`** $\rightarrow$ xoay thành **`n$m*`** $\rightarrow$ tra cứu tiền tố trên Permuterm B-Tree.

#### B. k-gram Index cho Wildcard
* Tách từ thành các chuỗi $k$ ký tự liên tiếp có ký hiệu `$`. Ví dụ 3-gram của `castle$` là `$ca`, `cas`, `ast`, `stl`, `tle`, `le$`.
* Để xử lý `re*ve`, hệ thống thực hiện truy vấn Boolean trên 3-gram index: **`$re AND ve$`**.
* **Post-filtering (Lọc sau):** Kiểm tra lại tập kết quả ứng viên với mẫu chuỗi gốc `re*ve` để loại bỏ các kết quả dương tính giả (false positives).

> **💬 Lời bình của tôi:** Khái niệm $k$-gram ở đây chính là nền tảng cho Edge N-gram trong Elasticsearch mà tài liệu yêu cầu dùng cho tính năng tìm kiếm dở từ (Partial matching: gõ `cà ph` ra `cà phê`).

---

### 3. Sửa Lỗi Chính Tả (Spelling Correction)

#### A. Khoảng cách Chỉnh sửa (Edit Distance / Levenshtein Distance)
Số phép toán tối thiểu bao gồm **Chèn (Insertion)**, **Xóa (Deletion)**, **Thay thế (Substitution)** để chuyển chuỗi $s_1$ thành $s_2$.

```python
EDITDISTANCE(s1, s2):
    m = matrix of size (|s1| + 1) x (|s2| + 1)
    for i from 1 to |s1|: m[i, 0] = i
    for j from 1 to |s2|: m[0, j] = j
    
    for i from 1 to |s1|:
        for j from 1 to |s2|:
            cost = 0 if s1[i] == s2[j] else 1
            m[i, j] = min(
                m[i-1, j-1] + cost,  # Thay thế (Substitution)
                m[i-1, j] + 1,       # Xóa (Deletion)
                m[i, j-1] + 1        # Chèn (Insertion)
            )
    return m[|s1|, |s2|]
```

#### B. Hệ số Jaccard (Jaccard Coefficient) trên k-gram Index
Để tìm các từ ứng viên nhanh chóng trước khi tính Edit Distance, hệ thống đo độ tương đồng k-gram bằng **Hệ số Jaccard**:

$$\text{Jaccard}(q, t) = \frac{|A \cap B|}{|A \cup B|} = \frac{|A \cap B|}{|A| + |B| - |A \cap B|}$$

*Trong đó $|A|$ và $|B|$ là số lượng k-gram của truy vấn $q$ và thuật ngữ $t$.*

#### C. Sửa lỗi theo Ngữ cảnh (Context-Sensitive Spelling Correction)
Xử lý các trường hợp các từ đứng riêng lẻ đều đúng chính tả nhưng sai ngữ cảnh (ví dụ: *"flew form Heathrow"* $\rightarrow$ *"flew from Heathrow"*). Sử dụng thống kê tần suất biword/ngram trong **Query Logs** hoặc toàn bộ tập văn bản để chọn phương án sửa phù hợp nhất.

> **💬 Lời bình của tôi:** Phần sửa lỗi này liên quan trực tiếp đến cái lỗi `hocj sinh` mình vừa test chiều nay! Nếu áp dụng khoảng cách Levenshtein hoặc quy tắc chuyển đổi phím Telex thì hệ thống hoàn toàn có thể tự gợi ý hoặc sửa từ `hocj` thành `học`.

---

### 4. Sửa Lỗi Theo Ngữ Âm (Soundex Algorithm)
Ánh xạ các từ có âm đọc tương tự nhau (đặc biệt là tên riêng) về cùng một **mã 4 ký tự** (1 chữ cái + 3 chữ số).

#### Các bước chuyển đổi Soundex:
1. Giữ nguyên chữ cái đầu tiên của từ.
2. Chuyển các chữ cái `A, E, I, O, U, H, W, Y` thành số `0`.
3. Chuyển các chữ cái còn lại thành số:
   * `B, F, P, V` $\rightarrow$ `1`
   * `C, G, J, K, Q, S, X, Z` $\rightarrow$ `2`
   * `D, T` $\rightarrow$ `3`
   * `L` $\rightarrow$ `4`
   * `M, N` $\rightarrow$ `5`
   * `R` $\rightarrow$ `6`
4. Bỏ các chữ số trùng lặp đứng cạnh nhau.
5. Bỏ tất cả số `0`, đệm thêm số `0` ở cuối cho đủ 4 ký tự.

* **Ví dụ:** Từ `Hermann` $\rightarrow$ `H` | `e(0)` `r(6)` `m(5)` `a(0)` `n(5)` `n(5)` $\rightarrow$ `H6555` $\rightarrow$ gộp lặp $\rightarrow$ Mã Soundex = **`H655`**.
