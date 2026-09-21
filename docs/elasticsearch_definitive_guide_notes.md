# Tóm Tắt & Ứng Dụng Sách "Elasticsearch: The Definitive Guide" Vào Dự Án Vietnamese Chat Search

> **Tài liệu tổng hợp và đúc kết chuyên sâu** từ cuốn sách kinh điển *Elasticsearch: The Definitive Guide* (Clinton Gormley & Zachary Tong, O'Reilly).  
> Tài liệu này **chắt lọc chính xác các chương và kỹ thuật cốt lõi có liên quan trực tiếp** đến bài toán tìm kiếm tin nhắn chat tiếng Việt: Cơ chế hoạt động của Lucene Shard, Multifield Search, Edge N-gram Partial Matching, Thuật toán BM25, và Xử lý chuẩn hóa ngôn ngữ (Accented/Unaccented).

---

## 📌 BẢN ĐỒ KIẾN THỨC SÁCH ÁP DỤNG VÀO DỰ ÁN

```
                                    ELASTICSEARCH: THE DEFINITIVE GUIDE
                                                    │
        ┌───────────────────┬───────────────────────┼───────────────────────┬───────────────────┐
        ▼                   ▼                       ▼                       ▼                   ▼
  [ CHƯƠNG 6 & 11 ]    [ CHƯƠNG 14 ]           [ CHƯƠNG 16 ]           [ CHƯƠNG 17 ]       [ CHƯƠNG 20 ]
Inside a Shard &    Multifield Search &     Partial Matching &      Controlling Relevance   Normalizing Tokens
Inverted Index      Field Boosting          Edge N-gram             & Okapi BM25            & Diacritics
─────────────────   ────────────────────    ───────────────────     ─────────────────────   ──────────────────
• Inverted Index    • Single vs Multi Query • Nhược điểm Prefix     • Hạn chế TF-IDF        • Unicode NFC/NFD
• Immutability      • multi_match types     • Index-time N-grams    • Bão hòa TF (k1=1.2)   • Accented vs
• Deletes & Merges  • Multifield Mapping    • Edge N-gram (2-15)    • Phạt độ dài (b=0.75)    Unaccented
• NRT & Translog    • Boosting (^5, ^3, ^1) • Search-as-you-type    • Lucene BM25 Scorer    • Token Filters
```

---

## PHẦN 1. BẢN CHẤT INVERTED INDEX & CƠ CHẾ HOẠT ĐỘNG CỦA SHARD (CHƯƠNG 6 & 11)

### 1.1. Cấu Trúc Inverted Index Trong Lucene & Elasticsearch
Theo Chương 6, một Inverted Index bao gồm hai thành phần cơ bản:
1. **Term Dictionary**: Danh sách toàn bộ các từ khóa (tokens) duy nhất đã được sắp xếp theo thứ tự từ điển, cho phép tìm kiếm nhị phân $O(\log N)$.
2. **Postings List**: Danh sách liên kết ghi nhận:
   - ID của các tài liệu chứa từ khóa đó.
   - Tần suất từ (Term Frequency - TF) trong tài liệu.
   - Vị trí xuất hiện (Positions) của từ khóa (phục vụ tìm kiếm cụm từ `match_phrase` và Proximity search).

### 1.2. Tính Bất Biến (Immutability) Của Chỉ Mục Đảo
Một nguyên lý thiết kế sống còn được sách nhấn mạnh trong Chương 11 (*Inside a Shard*):
> *"An inverted index, once written to disk, is immutable: it never changes."* (Chỉ mục đảo một khi đã ghi xuống đĩa thì không bao giờ bị sửa đổi).

**Ưu điểm của tính bất biến:**
- **Không cần Lock đồng thời**: Vì chỉ mục chỉ đọc (read-only), hàng triệu luồng tìm kiếm có thể truy cập song song mà không sợ xung đột dữ liệu.
- **Tận dụng tối đa OS File System Cache**: Hệ điều hành giữ toàn bộ chỉ mục trên bộ nhớ đệm RAM mà không bao giờ phải kiểm tra xem file có bị ghi đè hay không.

### 1.3. Cơ Chế Xóa & Cập Nhật (Deletes & Updates via Tombstones)
Nếu chỉ mục là bất biến, làm sao Elasticsearch có thể cập nhật hoặc xóa tài liệu?
- **Khi Xóa một tài liệu**: Elasticsearch **không xóa dữ liệu** khỏi segment. Thay vào đó, nó ghi ID của tài liệu vào một file riêng gọi là **`.del` file (Tombstone)**. Trong quá trình tìm kiếm, tài liệu vẫn được khớp nhưng bị bộ lọc loại bỏ trước khi trả về kết quả.
- **Khi Cập Nhật một tài liệu**: Elasticsearch thực hiện chuỗi thao tác nguyên tử:
  $$\text{Update} = \text{Delete (Đánh dấu vào .del)} + \text{Index Document Mới (vào Segment mới)}$$
- **Segment Merging (Tiến trình dọn rác ngầm)**:
  - Khi số lượng segments nhỏ tăng lên, tiến trình ngầm sẽ gộp chúng lại thành các segments lớn hơn.
  - Lúc này, các tài liệu bị đánh dấu trong `.del` mới thực sự bị xóa bỏ vĩnh viễn khỏi ổ cứng.
- **Liên hệ với Go In-memory Engine**: Cơ chế này chính là cảm hứng để chúng ta cài đặt hàm `DeleteDocument` và cơ chế **Atomic Upsert** trong `pkg/invertedindex/index.go`, giải quyết triệt để lỗi duplicate postings.

### 1.4. Near Real-Time (NRT) Search & Độ Bền Vững (Translog)
- **`refresh_interval` (Mặc định 1 giây)**: Tài liệu mới nạp được ghi vào buffer bộ nhớ, sau đó đẩy sang OS cache (gọi là `refresh`) để có thể tìm kiếm được ngay sau 1 giây mà không cần ghi thẳng xuống đĩa cứng (FSync).
- **Translog (Transaction Log)**: Để chống mất dữ liệu khi mất điện hoặc crash ứng dụng, mọi thao tác ghi đều được append tuần tự vào Translog. Khi khởi động lại, Elasticsearch đọc Translog để phục hồi dữ liệu hoàn toàn.

---

## PHẦN 2. CHIẾN LƯỢC TÌM KIẾM ĐA TRƯỜNG & BOOSTING (CHƯƠNG 14)

### 2.1. Tại Sao Tin Nhắn Chat Tiếng Việt Bắt Buộc Phải Dùng Multifield?
Chương 14 (*Multifield Search*) chỉ ra rằng: Hiếm khi một tài liệu chỉ được tìm kiếm bằng một trường đơn lẻ. Đặc biệt với tiếng Việt, người dùng có nhiều thói quen gõ khác nhau:
- Gõ chuẩn có dấu: *"học sinh"*
- Gõ nhanh không dấu: *"hoc sinh"*
- Gõ dở từ / tiền tố: *"học s"* hoặc *"ca ph"*

Nếu ta chỉ lưu một trường duy nhất, ta sẽ rơi vào thế "tiến thoái lưỡng nan": Khử dấu thì mất độ chính xác của từ ghép; giữ nguyên dấu thì không tìm được khi gõ không dấu; dùng N-gram thì làm chậm tốc độ tìm kiếm từ khóa chính xác.

**Giải pháp của Elasticsearch**: Sử dụng **Multifield Mapping** – một trường văn bản `content` được phân tích thành nhiều trường con với các Analyzer khác nhau:
```
content (Text hiển thị)
  ├── content.tokenized   (Cốc Cốc Tokenizer: giữ từ ghép có dấu "học_sinh")
  ├── content.unaccented  (Chuyển không dấu: "hoc_sinh")
  └── content.partial     (Edge N-gram: "họ", "học", "học_", "học_s", ...)
```

### 2.2. Chiến Lược Truy Vấn `multi_match` Với `most_fields`
Sách giải thích sự khác biệt giữa hai chế độ quan trọng của `multi_match`:
- **`best_fields` (dùng `dis_max`)**: Tìm một trường duy nhất khớp tốt nhất với query. Thích hợp khi tìm tiêu đề sách hoặc tên người.
- **`most_fields` (Khuyên dùng cho Chat tiếng Việt)**: Tìm kiếm trên tất cả các trường và **cộng dồn điểm số (Combine Scores)** từ các trường khớp được. Tài liệu nào vừa khớp từ ghép có dấu, vừa khớp không dấu, vừa khớp tiền tố sẽ nhận được tổng điểm cao nhất.

### 2.3. Trọng Số Tăng Cường (Per-Field Relevance Boosting)
Cú pháp tăng cường trọng số trường của Elasticsearch sử dụng toán tử `^boost`:
```json
{
  "multi_match": {
    "query": "học sinh",
    "type": "most_fields",
    "fields": [
      "content.tokenized^5",
      "content.unaccented^3",
      "content.partial^1"
    ]
  }
}
```
* **Boost `^5` cho `content.tokenized`**: Đảm bảo tin nhắn khớp trúng từ ghép tiếng Việt chuẩn luôn đứng Top 1.
* **Boost `^3` cho `content.unaccented`**: Giúp người dùng gõ không dấu vẫn tìm thấy tin nhắn với điểm số cao nhì.
* **Boost `^1` cho `content.partial`**: Dự phòng cho các trường hợp gõ dở từ hoặc tìm kiếm tiền tố.

---

## PHẦN 3. TÌM KIẾM TIỀN TỐ & GÕ DỞ TỪ BẰNG EDGE N-GRAM (CHƯƠNG 16)

### 3.1. Nhược Điểm Chí Mạng Của `prefix` và `wildcard` Query
Chương 16 (*Partial Matching*) cảnh báo nghiêm khắc:
- Truy vấn `prefix` (ví dụ `"học*"`) hoặc `wildcard` (ví dụ `"*học*"`) phải **quét tuần tự qua toàn bộ Term Dictionary** của Lucene để tìm các từ bắt đầu bằng tiền tố đó.
- Khi dữ liệu tăng lên hàng triệu tin nhắn, truy vấn dạng này gây suy giảm hiệu năng nghiêm trọng (Heavy CPU & Disk I/O).

### 3.2. Giải Pháp: Index-Time Edge N-grams
Thay vì để lúc tìm kiếm mới đi quét tiền tố, ta chuyển chi phí tính toán về **thời điểm nạp dữ liệu (Index-Time)** bằng kỹ thuật **Edge N-gram Token Filter**:
- **N-gram thông thường**: Tách mọi chuỗi con ở mọi vị trí (sinh ra số lượng token khổng lồ, gây nhiễu).
- **Edge N-gram**: Chỉ neo và cắt chuỗi con **từ đầu từ (Anchor from the beginning)**.

**Ví dụ thực tế với từ ghép `cà_phê`:**
$$\text{cà\_phê} \xrightarrow{\text{edge\_ngram (min: 2, max: 10)}} [\text{"cà"}, \text{"cà\_"}, \text{"cà\_p"}, \text{"cà\_ph"}, \text{"cà\_phê"}]$$

Khi người dùng gõ `"cà ph"`, Elasticsearch chỉ cần thực hiện phép tìm kiếm chính xác $O(1)$ vào token `"cà_ph"` trong Inverted Index mà không cần bất kỳ toán tử regex hay wildcard nào!

---

## PHẦN 4. THUẬT TOÁN XẾP HẠNG OKAPI BM25 (CHƯƠNG 17)

### 4.1. Tại Sao BM25 Vượt Trội Hơn TF/IDF Kinh Điển?
Chương 17 (*Controlling Relevance*) phân tích sâu lý do Lucene chuyển đổi thuật toán mặc định từ TF/IDF sang **Okapi BM25**:

```
Điểm số (Score)
  │                                 /─── TF/IDF (Điểm tăng vô hạn khi lặp từ)
  │                               /
  │                    ──────────/────── Ngưỡng bão hòa của BM25 (Trần điểm số)
  │                  /
  │                /
  │              /
  │            /
  │          /
  │        /
  └───────┴─────────────────────────────── Tần suất từ trong tài liệu (TF)
```

1. **Bão Hòa Tần Suất Từ ($k_1 = 1.2$)**:
   - Trong TF-IDF, nếu một từ xuất hiện 20 lần trong tin nhắn rác, điểm của nó sẽ gấp nhiều lần tin nhắn xuất hiện 1 lần.
   - Trong BM25, khi TF vượt qua một ngưỡng nhất định, điểm đóng góp sẽ tiệm cận ngưỡng trần bão hòa $\approx (k_1 + 1) = 2.2$. Kẻ xấu không thể spam lặp từ để đẩy thứ hạng tin nhắn lên đầu.
2. **Chuẩn Hóa Độ Dài Tài Liệu ($b = 0.75$)**:
   - BM25 so sánh độ dài $|D|$ của tin nhắn với độ dài trung bình toàn hệ thống ($avgdl$).
   - Tin nhắn ngắn, súc tích chứa từ khóa sẽ nhận điểm cao hơn nhiều so với một đoạn văn bản dài lê thê vô tình chứa từ khóa đó.

---

## PHẦN 5. CHUẨN HÓA DẤU THANH TIẾNG VIỆT (CHƯƠNG 20)

### 5.1. Thách Thức Của Ngôn Ngữ Có Dấu (Diacritics)
Chương 20 (*Normalizing Tokens - You Have an Accent*) phân tích:
- Tiếng Việt sử dụng hệ thống ký tự Latinh mở rộng với nhiều dấu thanh phức tạp (sắc, huyền, hỏi, ngã, nặng) và dấu mũ/móc (ă, â, đ, ê, ô, ơ, ư).
- Người dùng chat thường xuyên gõ không dấu để tiết kiệm thời gian, đặc biệt trên điện thoại di động.

### 5.2. Chuẩn Hóa Unicode (NFC vs NFD)
- Tiếng Việt tồn tại 2 dạng biểu diễn Unicode:
  - **Dạng dựng sẵn (NFC - Canonical Composition)**: Ký tự `á` được lưu bằng 1 mã Unicode duy nhất (`U+00E1`).
  - **Dạng tổ hợp (NFD - Canonical Decomposition)**: Ký tự `á` được ghép từ `a` (`U+0061`) và dấu sắc `´` (`U+0301`).
- Nếu không chuẩn hóa về cùng định dạng **NFC**, Elasticsearch sẽ coi 2 từ viết giống hệt nhau là 2 từ khác biệt hoàn toàn.

### 5.3. Chiến Lược Lập Chỉ Mục Kép
Sách khuyến nghị không nên xóa hẳn dấu thanh (vì làm mất ngữ nghĩa: *"bán hàng"* khác *"bàn ghế"*), mà phải lưu giữ **cả 2 phiên bản song song**:
- Phiên bản có dấu phục vụ xếp hạng ưu tiên cao nhất.
- Phiên bản không dấu (sau khi lọc qua unaccent filter) phục vụ độ phủ (Recall) khi người dùng gõ nhanh.

---

## PHẦN 6. BẢN THIẾT KẾ MAPPING ELASTICSEARCH CHUẨN MỰC CHO DỰ ÁN

Dựa trên toàn bộ các nguyên lý từ cuốn sách, dưới đây là bản thiết kế Mapping Elasticsearch 8.x hoàn chỉnh cho hệ thống chat tin nhắn tiếng Việt:

### 6.1. Chỉ Mục Đối Soát Mặc Định: `chat_messages_baseline`
```json
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "properties": {
      "id": { "type": "integer" },
      "sender": { "type": "keyword" },
      "room": { "type": "keyword" },
      "content": { 
        "type": "text", 
        "analyzer": "standard" 
      },
      "created_at": { "type": "date" }
    }
  }
}
```

### 6.2. Chỉ Mục Tối Ưu Tiếng Việt: `chat_messages_vietnamese`
```json
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "filter": {
        "edge_ngram_filter": {
          "type": "edge_ngram",
          "min_gram": 2,
          "max_gram": 15
        }
      },
      "analyzer": {
        "coccoc_whitespace_analyzer": {
          "type": "custom",
          "tokenizer": "whitespace",
          "filter": ["lowercase"]
        },
        "partial_ngram_analyzer": {
          "type": "custom",
          "tokenizer": "whitespace",
          "filter": ["lowercase", "edge_ngram_filter"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "integer" },
      "sender": { "type": "keyword" },
      "room": { "type": "keyword" },
      "content": { "type": "text" },
      "content_tokenized": {
        "type": "text",
        "analyzer": "coccoc_whitespace_analyzer",
        "similarity": "BM25"
      },
      "content_unaccented": {
        "type": "text",
        "analyzer": "coccoc_whitespace_analyzer",
        "similarity": "BM25"
      },
      "content_partial": {
        "type": "text",
        "analyzer": "partial_ngram_analyzer"
      },
      "created_at": { "type": "date" }
    }
  }
}
```

### 6.3. Cấu Trúc Truy Vấn Tìm Kiếm Tối Ưu (Query DSL)
```json
{
  "query": {
    "bool": {
      "must": [
        {
          "multi_match": {
            "query": "học sinh",
            "type": "most_fields",
            "fields": [
              "content_tokenized^5.0",
              "content_unaccented^3.0",
              "content_partial^1.0"
            ]
          }
        }
      ],
      "filter": [
        { "term": { "room": "Team Dự Án Search Engine" } }
      ]
    }
  }
}
```

---

## 🎯 TỔNG KẾT BÀI HỌC ÁP DỤNG

1. **Hiểu bản chất Segment & Tombstone**: Giúp chúng ta tự tin xây dựng và kiểm soát cơ chế Atomic Re-index không bao giờ bị duplicate postings.
2. **Kỹ thuật Multifield Mapping**: Cho phép một tin nhắn chat phục vụ trọn vẹn 3 nhu cầu tìm kiếm (có dấu chính xác, không dấu nhanh, và gõ dở từ) trên cùng một văn bản.
3. **Edge N-gram thay thế Wildcard**: Bảo đảm tính năng Autocomplete và Partial Matching có thời gian phản hồi dưới $10\text{ms}$ ngay cả khi kho dữ liệu mở rộng lớn.
4. **Trọng số Boosting kết hợp BM25**: Đảm bảo trải nghiệm người dùng luôn nhận được kết quả chuẩn xác nhất ở vị trí Top 1.
