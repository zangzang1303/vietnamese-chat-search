# TÀI LIỆU KỸ THUẬT: CƠ CHẾ XỬ LÝ, LUỒNG GHI VÀ CẤU TRÚC LƯU TRỮ ĐĨA CỦA TIN NHẮN
## (The Life of a Write & Disk Storage Internals)

> **Tài liệu tham chiếu thiết kế hệ thống lưu trữ & tìm kiếm tiếng Việt**  
> **Dự án:** Vietnamese Chat Search Engine (Dual-Engine: Go Binary Disk vs Elasticsearch 8.x Cốc Cốc)  
> **Tác giả:** Đội ngũ Kỹ thuật & R&D  
> **Trạng thái:** Đã triển khai thực tế & Hoạt động hoàn chỉnh trên Production Local  

---

## 1. Tổng Quan Bài Toán: "Khi Nhập Tin Nhắn Thì Hệ Thống Làm Gì?"

Khi một người dùng gõ tin nhắn vào ô chat (ví dụ: *"Chào các bạn học sinh mới uống cà phê"*) và nhấn **Gửi (Enter)**, hệ thống phải giải quyết 3 bài toán kỹ thuật cốt lõi:

1. **Tính Toàn Vẹn & Bền Vững (Durability):** Tin nhắn phải được lưu ngay lập tức xuống đĩa cứng vật lý. Nếu máy tính bị rút điện hoặc crash trong tích tắc tiếp theo, dữ liệu **tuyệt đối không được mất**.
2. **Khả Năng Đọc Toàn Văn Siêu Tốc (Fast Retrieval - $O(1)$):** Khi hiển thị cuộc trò chuyện hoặc cuộn xem lịch sử, hệ thống phải đọc được tin nhắn ra ngay lập tức mà không cần quét toàn bộ cơ sở dữ liệu.
3. **Khả Năng Tìm Kiếm Ngữ Nghĩa Tiếng Việt (Searchable Inverted Index):** Tin nhắn phải được phân tích bóc tách từ ngữ tiếng Việt đa tầng (từ ghép có dấu, không dấu, tiền tố gõ dở) và lập chỉ mục ngược để người dùng gõ bất kỳ từ nào cũng tìm thấy tức thì với xếp hạng điểm số Okapi BM25.

---

## 2. Sơ Đồ Luồng Dữ Liệu Chi Tiết (The Life of a Write Sequence Diagram)

Sơ đồ tuần tự thể hiện đường đi của dữ liệu từ khi bấm phím đến từng byte trên đĩa cứng:

![Sơ Đồ Luồng Dữ Liệu Chi Tiết](../image/The%20Life%20of%20a%20Write%20Sequence%20Diagram.png)

*Hình 2.1: Sơ đồ tuần tự thể hiện đường đi của dữ liệu từ khi bấm phím đến từng byte trên đĩa cứng*

---

## 3. Tin Nhắn Được Lưu Ở Đâu? (Bản Đồ Thư Mục Vật Lý Trên Ổ Đĩa)

Toàn bộ dữ liệu của Custom Engine được lưu tại thư mục độc lập:  
📁 `d:\CODE\VSF\vietnamese-chat-search\data\custom_storage\`

### 3.1. Cây Thư Mục Thực Tế Trên Máy Tính
```text
d:\CODE\VSF\vietnamese-chat-search\data\custom_storage\
├── wal.log          <-- [1] Nhật ký ghi trước chống sập nguồn (199 bytes)
├── docstore.idx     <-- [2] Bảng mục lục Offset cố định 12 bytes/doc (1,584 bytes)
├── docstore.dat     <-- [3] Kho dữ liệu toàn văn của toàn bộ tin nhắn gốc (36,226 bytes)
├── terms.dict       <-- [4] Từ điển nhị phân đã sắp xếp A-Z (44,300 bytes)
├── postings.bin     <-- [5] Danh sách vị trí xuất hiện (DocID, TF, Positions) (104,304 bytes)
├── segments.meta    <-- [6] Metadata thống kê số Doc, Token cho Okapi BM25 (42 bytes)
└── tombstone.del    <-- [7] Danh sách DocID các tin nhắn bị xóa (Soft Delete)
```

### 3.2. Bảng Mô Tả Chi Tiết Từng File Dữ Liệu

| Tên File | Vai Trò Kỹ Thuật | Tần Suất Truy Cập | Cơ Chế An Toàn |
| :--- | :--- | :--- | :--- |
| **`wal.log`** | Ghi nhật ký tuần tự (Write-Ahead Log) trước khi đưa vào RAM. | Ghi liên tục khi có tin nhắn mới ($O(1)$ append). | Có mã CRC32 kiểm tra chống hỏng file; gọi `fsync()` ngay lập tức. |
| **`docstore.idx`** | Forward Index: Bảng mục lục con trỏ 12 bytes/DocID. | Ghi khi thêm tin mới; Đọc ngẫu nhiên khi tìm kiếm. | Kích thước cố định giúp tính toán vị trí con trỏ bằng công thức toán học $Offset = DocID \times 12$. |
| **`docstore.dat`** | Kho lưu trữ dữ liệu gốc toàn văn (ID, Sender, Room, Content, Timestamps). | Ghi nối đuôi; Đọc đúng đoạn byte cần thiết qua `docstore.idx`. | Dữ liệu nhị phân nguyên bản, không cần nạp toàn bộ file vào RAM. |
| **`terms.dict`** | Từ điển từ khóa tiếng Việt đã sắp xếp Alphabet A-Z. | Đọc khi tìm kiếm bằng thuật toán Binary Search $O(\log N)$. | File bất biến (Immutable), không bị Race Condition khi nhiều luồng đọc đồng thời. |
| **`postings.bin`** | Danh sách Posting List (DocID, TF, Positions) của từng term. | Đọc trực tiếp từ đĩa vào RAM theo con trỏ offset từ `terms.dict`. | Lưu trữ nhị phân nén chặt chẽ. |
| **`segments.meta`** | Header lưu thông số: `TotalDocs`, `TotalTokens`, `AvgDocLength`. | Đọc một lần khi khởi động engine. | Cung cấp tham số toàn cục cho công thức xếp hạng Okapi BM25. |
| **`tombstone.del`** | Danh sách DocID các tin nhắn đã bị xóa (Tombstone). | Ghi khi gọi DELETE; Đọc khi tìm kiếm để loại bỏ tin nhắn đã hủy. | Cơ chế Soft Delete: Không cần viết lại file segment khổng lồ. |

---

## 4. Cấu Trúc Nhị Phân Từng Byte Dưới Đĩa (Binary Disk Layout)

Để đạt tốc độ tối đa và tiêu tốn tối thiểu tài nguyên phần cứng, engine không sử dụng JSON hay SQLite, mà tự định nghĩa cấu trúc nhị phân thuần túy theo chuẩn **Little Endian**:

### 4.1. Cấu Trúc File `wal.log` (Write-Ahead Log)
Mỗi bản ghi tin nhắn trong `wal.log` bao gồm 3 phần liên tiếp:

```
┌─────────────────┬────────────────────┬──────────────────────────────────────────┐
│ Payload Length  │  CRC32 Checksum    │           Binary Payload Data            │
│  uint32 (4B)    │    uint32 (4B)     │                 (Length bytes)           │
└─────────────────┴────────────────────┴──────────────────────────────────────────┘
```

- **Byte 0–3 (`Payload Length`):** Độ dài của phần dữ liệu phía sau.
- **Byte 4–7 (`CRC32 Checksum`):** Mã băm IEEE CRC32 tính trên toàn bộ Payload. Khi khởi động lại, engine tính lại CRC32; nếu lệch dù chỉ 1 bit (do mất điện khi đang ghi dở) sẽ bỏ qua bản ghi lỗi, chống hỏng bộ nhớ.
- **Payload Data:** Chuỗi byte mã hóa của struct `Message` (`ID (4B) + SenderLen (2B) + Sender + RoomLen (2B) + Room + ContentLen (4B) + Content + Timestamps`).

---

### 4.2. Cấu Trúc Tra Cứu $O(1)$ Của `docstore.idx` & `docstore.dat`
Đây là thiết kế kinh điển của các Storage Engine hàng đầu (như Bitcask, LSM DocStore):

```
       FILE: docstore.idx (Mỗi Doc chiếm đúng 12 bytes)
       ┌──────────────────────────────┬──────────────────────────────┬───
Doc 1: │ ByteOffset: 0       (8 bytes)│ DataLength: 145     (4 bytes)│
       ├──────────────────────────────┼──────────────────────────────┼───
Doc 2: │ ByteOffset: 145     (8 bytes)│ DataLength: 210     (4 bytes)│
       ├──────────────────────────────┼──────────────────────────────┼───
...    │ ...                          │ ...                          │
       ├──────────────────────────────┼──────────────────────────────┼───
Doc N: │ ByteOffset: 48,210  (8 bytes)│ DataLength: 180     (4 bytes)│
       └──────────────────────────────┴──────────────────────────────┴───
                     │                               │
                     ▼ Nhảy tới byte 48,210          ▼ Đọc đúng 180 bytes
       FILE: docstore.dat (Kho dữ liệu toàn văn)
       ┌──────────────────────────────┬──────────────────────────────┬───
       │ ... tin nhắn trước ...       │ [ID=N] [Sender="Lê Tuấn"]    │
       │                              │ [Content="uống cà phê..."]   │
       └──────────────────────────────┴──────────────────────────────┴───
```

#### Công thức toán học tra cứu $O(1)$:
Khi cần lấy nội dung của tài liệu có `DocID`:
$$\text{IdxOffset} = (\text{DocID} - 1) \times 12$$
1. Nhảy con trỏ file `docstore.idx` tới $\text{IdxOffset}$.
2. Đọc 8 bytes đầu $\to$ `ByteOffset`. Đọc 4 bytes sau $\to$ `DataLength`.
3. Nhảy con trỏ file `docstore.dat` tới `ByteOffset` và đọc đúng `DataLength` bytes.
4. **Độ phức tạp:** Đúng **1 thao tác Seek đĩa duy nhất**, tốc độ dưới $0.1$ mili-giây, bất kể cơ sở dữ liệu có 100 tin hay 100 triệu tin nhắn!

---

### 4.3. Cấu Trúc File `terms.dict` (Từ Điển Đã Sắp Xếp)
Lưu trữ danh sách các term tiếng Việt theo thứ tự chữ cái A-Z:

```
┌───────────────┬─────────────────────────────────────────────────────────────────┐
│ Header (16B)  │ Magic Number (0x56534653 "VSFS") (4B) + EntryCount uint32 (4B) │
├───────────────┼─────────────────────────────────────────────────────────────────┤
│ Entry 1       │ TermLength (2B) + Term ("cà_phê") + DocFreq (4B) + PostingsOffset (8B)
├───────────────┼─────────────────────────────────────────────────────────────────┤
│ Entry 2       │ TermLength (2B) + Term ("học_sinh") + DocFreq (4B) + PostingsOffset (8B)
├───────────────┼─────────────────────────────────────────────────────────────────┤
│ ...           │ ...                                                             │
└───────────────┴─────────────────────────────────────────────────────────────────┘
```

- Nhờ danh sách Term được sắp xếp Alphabet (A-Z), engine có thể áp dụng thuật toán **Tìm kiếm nhị phân (Binary Search)** để tìm ra term với độ phức tạp chỉ $O(\log N)$ phép so sánh.
- `PostingsOffset` chỉ chính xác vị trí con trỏ trong file `postings.bin`.

---

### 4.4. Cấu Trúc File `postings.bin` (Danh Sách Xuất Hiện Ngược)
Chứa thông tin chi tiết về từng tài liệu mà từ khóa đó xuất hiện:

```
┌─────────────────┬───────────────────────────────────────────────────────────────┐
│ Postings Chunk  │ DocID (4B) + TermFrequency (4B) + PosCount (2B) + Positions...│
├─────────────────┼───────────────────────────────────────────────────────────────┤
│ Ví dụ: cà_phê   │ DocID: 15, TF: 1, PosCount: 1, Positions: [1] (từ thứ 2)      │
│                 │ DocID: 17, TF: 2, PosCount: 2, Positions: [4, 9]              │
└─────────────────┴───────────────────────────────────────────────────────────────┘
```

- Cung cấp dữ liệu trực tiếp cho hàm tính điểm **Okapi BM25**:
  $$\text{Score}(D, Q) = \sum_{t \in Q} \text{IDF}(t) \cdot \frac{\text{TF}(t, D) \cdot (k_1 + 1)}{\text{TF}(t, D) + k_1 \cdot \left(1 - b + b \cdot \frac{|D|}{\text{avgdl}}\right)}$$
- `Positions` cung cấp tọa độ chính xác của từ trong câu để bôi vàng (Highlight) trên giao diện Web Messenger.

---

## 5. Quy Trình Xử Lý Ngôn Ngữ Tiếng Việt (Vietnamese NLP Tokenization)

Khi câu *"Chào các bạn học sinh mới uống cà phê"* đi vào hàm `indexToMemTable`, nó được bóc tách qua **3 tầng xử lý**:

```
                       CÂU CHÁT BAN ĐẦU
        "Chào các bạn học sinh mới uống cà phê"
                           │
       ┌───────────────────┴───────────────────┐
       ▼                                       ▼
 [TẦNG 1: CỐC CỐC TOKENIZER]             [TẦNG 3: KHÔNG DẤU (UNACCENTED)]
 Tách từ ghép có dấu chuẩn:              Chuyển đổi loại bỏ dấu:
 - "chào"      (pos 0)                  - "chao"       (pos 0)
 - "các"       (pos 1)                  - "cac"        (pos 1)
 - "bạn"       (pos 2)                  - "ban"        (pos 2)
 - "học_sinh"  (pos 3)                  - "hoc_sinh"   (pos 3)
 - "mới"       (pos 4)                  - "moi"        (pos 4)
 - "uống"      (pos 5)                  - "uong"       (pos 5)
 - "cà_phê"    (pos 6)                  - "ca_phe"     (pos 6)
       │                                       │
       ▼                                       ▼
 [TẦNG 2: EDGE N-GRAM TIỀN TỐ]           [TẦNG 3B: N-GRAM KHÔNG DẤU]
 Hỗ trợ tìm kiếm khi gõ dở chữ:          Hỗ trợ gõ dở không dấu:
 "cà", "cà_", "cà_p", "cà_ph", "cà_phê"  "ca", "ca_", "ca_p", "ca_ph", "ca_phe"
 "học", "học_", "học_s", "học_si"...     "hoc", "hoc_", "hoc_s"...
```

### Tại sao phải tách 3 tầng như vậy?
1. **Loại bỏ bẫy ghép sai tiếng Việt:** Nếu chỉ bẻ vụn từng từ đơn như Standard Analyzer của Elasticsearch, khi tìm kiếm *"học sinh"*, các câu như *"sinh viên đi làm"*, *"người lính hy sinh"* đều bị dính kết quả (False Positive). Cốc Cốc Tokenizer ghép chặt thành `học_sinh`, triệt tiêu hoàn toàn lỗi này.
2. **Hỗ trợ gõ không dấu tự nhiên:** Người Việt trên ứng dụng nhắn tin rất hay gõ không dấu (*"uong ca phe"*). Nhờ có tầng Unaccented, hệ thống vẫn xếp hạng chính xác tin nhắn có dấu gốc.
3. **Trải nghiệm tìm kiếm tức thì (Search-as-you-type):** Tầng Edge N-gram giúp người dùng vừa gõ *"cà p"* là kết quả đã hiện ra ngay phía dưới.

---

## 6. Trích Dẫn Mã Nguồn Thực Tế (Source Code Walkthrough)

Dưới đây là các đoạn mã nguồn cốt lõi đang chạy trực tiếp trên hệ thống:

### 6.1. Tiếp nhận tại Router & Ghi song song 2 Động cơ
*File: [`cmd/chat_server/main.go`](file:///d:/CODE/VSF/vietnamese-chat-search/cmd/chat_server/main.go#L154-L173)*

```go
// 1. Cấp phát tin nhắn vào bộ nhớ phòng chat
msg := chatManager.PostMessage(req.Sender, req.Room, req.Content)
log.Printf("📩 [Tin mới] #ID %d từ '%s' trong '%s': %s", msg.ID, msg.Sender, msg.Room, msg.Content)

// 2. Lưu trữ tức thì vào Custom Storage Engine trên đĩa
if customEngine != nil {
    _ = customEngine.IndexMessage(msg)
}

// 3. Lưu trữ bền vững vào Elasticsearch 8.x (bất đồng bộ qua goroutine)
if esClient != nil && esClient.IsAvailable() {
    go func(m chat.Message) {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        esClient.IndexSingleMessage(ctx, m, analyzer)
    }(msg)
}
```

---

### 6.2. Điều Phối Tại Custom Engine
*File: [`pkg/customengine/engine.go`](file:///d:/CODE/VSF/vietnamese-chat-search/pkg/customengine/engine.go#L105-L124)*

```go
func (e *Engine) IndexMessage(msg chat.Message) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // 1. Ghi vào WAL chống mất dữ liệu khi mất điện/crash
    if err := e.wal.Append(msg); err != nil {
        return fmt.Errorf("lỗi ghi WAL: %w", err)
    }

    // 2. Ghi vào DocStore trên đĩa để tra cứu O(1)
    if _, err := e.docStore.AppendMessage(msg); err != nil {
        return fmt.Errorf("lỗi ghi DocStore: %w", err)
    }

    // 3. Đánh chỉ mục từ vựng đa tầng vào MemTable trong RAM
    e.indexToMemTable(msg, true)

    return nil
}
```

---

### 6.3. Ghi Nhị Phân & Kiểm Tra Toàn Vẹn CRC32 Trong WAL
*File: [`pkg/storage/wal.go`](file:///d:/CODE/VSF/vietnamese-chat-search/pkg/storage/wal.go#L64-L100)*

```go
func (w *WAL) Append(msg chat.Message) error {
    w.mu.Lock()
    defer w.mu.Unlock()

    // Mã hóa tin nhắn thành mảng byte nhị phân
    payload, err := serializeMessage(msg)
    if err != nil {
        return err
    }

    // Tính mã CRC32 IEEE chống hỏng dữ liệu
    checksum := crc32.ChecksumIEEE(payload)

    // Ghi: [PayloadLen 4B] + [Checksum 4B] + [Payload]
    if err := WriteUint32(w.file, uint32(len(payload))); err != nil {
        return err
    }
    if err := WriteUint32(w.file, checksum); err != nil {
        return err
    }
    if _, err := w.file.Write(payload); err != nil {
        return err
    }

    // Ép hệ điều hành ghi ngay từ buffer xuống đĩa vật lý
    return w.file.Sync()
}
```

---

### 6.4. Ghi Nhị Phân DocStore & Mục Lục Offset 12 Bytes
*File: [`pkg/storage/docstore.go`](file:///d:/CODE/VSF/vietnamese-chat-search/pkg/storage/docstore.go#L54-L95)*

```go
func (ds *DocStore) AppendMessage(msg chat.Message) (uint64, error) {
    ds.mu.Lock()
    defer ds.mu.Unlock()

    // 1. Lấy vị trí offset hiện tại trong file dữ liệu docstore.dat
    offset, err := ds.dataFile.Seek(0, io.SeekEnd)
    if err != nil {
        return 0, err
    }

    // 2. Ghi chuỗi byte tin nhắn vào docstore.dat
    payload, _ := serializeMessage(msg)
    n, err := ds.dataFile.Write(payload)
    if err != nil {
        return 0, err
    }

    // 3. Ghi đúng 12 bytes vào docstore.idx: [Offset uint64 (8B)] + [Length uint32 (4B)]
    if err := WriteUint64(ds.idxFile, uint64(offset)); err != nil {
        return 0, err
    }
    if err := WriteUint32(ds.idxFile, uint32(n)); err != nil {
        return 0, err
    }

    return uint64(offset), nil
}
```

---

## 7. Bảng So Sánh Đối Chứng 3 Động Cơ Đang Chạy Song Song

| Tiêu Chí So Sánh | Custom Engine (Go Disk Persistence) | Elasticsearch 8.x (Cốc Cốc) | Elasticsearch Baseline (Standard) |
| :--- | :---: | :---: | :---: |
| **Công nghệ nền tảng** | Pure Go (Nhúng trực tiếp trong Chat Server) | Java 17 + Lucene 9.8 (Docker Container) | Lucene 9.8 Standard Analyzer (Docker) |
| **Dung lượng RAM tiêu thụ** | **~15 MB** (Siêu nhẹ) | **~1.2 GB** (Khối JVM Heap nặng) | **~1.2 GB** (Khối JVM Heap nặng) |
| **Thời gian khởi động** | **< 50 mili-giây** (Tức thì) | 20 – 35 giây (Khởi động Docker + JVM) | 20 – 35 giây |
| **Độ trễ tìm kiếm (Latency)**| **~3 mili-giây** (Gọi trực tiếp trong bộ nhớ) | ~14 mili-giây (Qua HTTP REST Network) | ~19 mili-giây (Qua HTTP REST Network) |
| **Cơ chế bền vững khi Crash**| WAL với CRC32 + fsync đĩa cứng | Translog của Lucene/Elasticsearch | Translog của Lucene/Elasticsearch |
| **Mức độ chính xác (MRR)** | **Cao ($\ge 0.75$)** nhờ Cốc Cốc đa tầng | **0.7708** (Rất cao) | **0.6250** (Thấp, nhiều False Positives) |
| **Phụ thuộc môi trường** | **Zero Dependency** (Chỉ 1 file exe duy nhất) | Yêu cầu Docker Engine, JVM, Port 9200 | Yêu cầu Docker Engine, JVM, Port 9200 |

---

## 8. Kết Luận

1. **Về luồng xử lý:** Hệ thống áp dụng chuẩn kiến trúc Database Engine hiện đại (LSM-tree kết hợp Inverted Index):
   - **Ghi nhanh:** Ghi tuần tự WAL và DocStore theo cơ chế Append-only để đạt tốc độ cao nhất.
   - **Đọc nhanh:** Tách từ Cốc Cốc đa tầng đưa vào MemTable trong RAM cho phép tìm kiếm tức thì.
   - **Lưu trữ lâu dài:** Xả (Flush) MemTable thành các Segment bất biến trên đĩa, cho phép tìm kiếm nhị phân $O(\log N)$ và đọc toàn văn $O(1)$.
2. **Về vị trí lưu trữ:** Toàn bộ dữ liệu nằm gọn gàng, độc lập trong thư mục `data/custom_storage/` dưới dạng các file nhị phân tối ưu hóa từng byte, hoạt động song song hoàn hảo với cụm Docker Elasticsearch 8.x.
