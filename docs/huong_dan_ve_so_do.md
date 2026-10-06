# HƯỚNG DẪN & MÃ NGUỒN VẼ TẤT CẢ SƠ ĐỒ, HÌNH VẼ TRONG BÁO CÁO THỰC TẬP

Tài liệu này cung cấp toàn bộ mã nguồn chuẩn **Mermaid.js** và **Chart.js** tương ứng với từng hình vẽ trong báo cáo [bao_cao_thuc_tap_hoan_thien.md](file:///d:/CODE/VSF/vietnamese-chat-search/docs/bao_cao_thuc_tap_hoan_thien.md).

> **Cách xuất hình ảnh nhanh nhất:**
> 1. **Cách 1 (Khuyên dùng):** Mở trực tiếp tệp [docs/render_diagrams.html](file:///d:/CODE/VSF/vietnamese-chat-search/docs/render_diagrams.html) bằng bất kỳ trình duyệt web nào (Chrome, Edge, Firefox). Toàn bộ sơ đồ và biểu đồ sẽ được dựng tự động với chất lượng cao. Bạn chỉ cần chụp ảnh màn hình hoặc lưu ảnh để dán vào tài liệu Word/Google Docs.
> 2. **Cách 2:** Sao chép các đoạn mã Mermaid dưới đây và dán vào trang web trực tuyến: [mermaid.live](https://mermaid.live) hoặc công cụ [draw.io](https://app.diagrams.net) (chọn `Arrange -> Insert -> Advanced -> Mermaid`) để tải về file ảnh PNG/SVG chuẩn 4K.
> 3. **Cách 3:** Tận dụng các ảnh đã được tạo sẵn trong thư mục `image/` của dự án.

---

## 1. Hình 2.1: Sơ đồ vị trí của Phân hệ Thu hồi Tin nhắn trong kiến trúc Nền tảng Nhắn tin quy mô lớn

```mermaid
flowchart TD
    subgraph Clients["Tầng Thiết Bị Đầu Cuối (Clients)"]
        Web["Web Messenger (Browser)"]
        Mobile["Mobile App (iOS / Android)"]
    end

    subgraph GatewayLayer["Tầng Cổng Kết Nối (Gateway)"]
        Gateway["Connection Gateway (WebSocket / gRPC Broker)"]
    end

    subgraph CoreLayer["Tầng Nghiệp Vụ Cốt Lõi (Core Services)"]
        CoreSvc["Message Core Service"]
        PrimaryDB[("Primary Database (PostgreSQL / Cassandra)")]
    end

    subgraph EventBus["Hạ Tầng Truyền Dẫn Thông Điệp"]
        Kafka{{"Message Broker / Kafka Event Stream"}}
    end

    subgraph SearchSubsystem["Phân Hệ Tìm Kiếm Tin Nhắn (Đề Tài Nghiên Cứu)"]
        NLP["Tầng Tiền Xử Lý NLP (CGO Cốc Cốc Tokenizer)"]
        Dispatcher["Bộ Điều Phối Lưu Trữ (Dual Dispatcher)"]
        ES["Cụm Elasticsearch 8.11 (Doanh Nghiệp)"]
        CustomEng["Động Cơ Nhị Phân Thuần Go (Custom Binary Engine)"]
    end

    Web <--> Gateway
    Mobile <--> Gateway
    Gateway <--> CoreSvc
    CoreSvc --> PrimaryDB
    CoreSvc -- "1. Publish Message Event" --> Kafka
    Kafka -- "2. Consume Real-time Stream" --> NLP
    NLP --> Dispatcher
    Dispatcher --> ES
    Dispatcher --> CustomEng
    Gateway -. "3. Query Search API" .-> Dispatcher
```

---

## 2. Hình 3.1: Minh họa cấu trúc Chỉ mục đảo (Inverted Index) và Danh sách Postings

### Lựa chọn 1: Chuẩn Giáo Trình IR Manning (Khuyên dùng - Chuẩn khoa học máy tính, không bao giờ bị lệch thứ tự)

```mermaid
flowchart TD
    subgraph Row1["Thuật ngữ: học_sinh (DF = 2)"]
        direction LR
        T1["Từ điển:<br><b>học_sinh</b><br>(DF = 2)"] --> P1_1["DocID: 1<br>TF: 1 | Pos: 0"] --> P1_2["DocID: 5<br>TF: 1 | Pos: 0"]
    end

    subgraph Row2["Thuật ngữ: cà_phê (DF = 2)"]
        direction LR
        T2["Từ điển:<br><b>cà_phê</b><br>(DF = 2)"] --> P2_1["DocID: 2<br>TF: 1 | Pos: 1"] --> P2_2["DocID: 5<br>TF: 1 | Pos: 4"]
    end

    subgraph Row3["Thuật ngữ: uống (DF = 2)"]
        direction LR
        T3["Từ điển:<br><b>uống</b><br>(DF = 2)"] --> P3_1["DocID: 2<br>TF: 1 | Pos: 0"] --> P3_2["DocID: 5<br>TF: 1 | Pos: 3"]
    end

    Row1 ~~~ Row2 ~~~ Row3

    classDef term fill:#e0e7ff,stroke:#4338ca,stroke-width:1.5px,color:#1e1b4b;
    classDef post fill:#f0fdf4,stroke:#16a34a,stroke-width:1.5px,color:#14532d;
    class T1,T2,T3 term;
    class P1_1,P1_2,P2_1,P2_2,P3_1,P3_2 post;
```

### Lựa chọn 2: Dạng 2 Khung Lớn (Lexicon bên trái, Postings bên phải)

```mermaid
flowchart LR
    subgraph Lexicon["Từ Điển Thuật Ngữ (Dictionary / Lexicon)"]
        T1["học_sinh (DF = 2)"]
        T2["cà_phê (DF = 2)"]
        T3["uống (DF = 2)"]
    end

    subgraph Postings["Danh Sách Postings (Postings Lists)"]
        P1["[DocID: 1 | TF: 1 | Pos: 0] ────► [DocID: 5 | TF: 1 | Pos: 0]"]
        P2["[DocID: 2 | TF: 1 | Pos: 1] ────► [DocID: 5 | TF: 1 | Pos: 4]"]
        P3["[DocID: 2 | TF: 1 | Pos: 0] ────► [DocID: 5 | TF: 1 | Pos: 3]"]
    end

    T1 --> P1
    T2 --> P2
    T3 --> P3

    classDef term fill:#e0e7ff,stroke:#4338ca,stroke-width:1.5px,color:#1e1b4b;
    classDef post fill:#f0fdf4,stroke:#16a34a,stroke-width:1.5px,color:#14532d;
    class T1,T2,T3 term;
    class P1,P2,P3 post;
```

---

## 3. Hình 3.2: Đồ thị trạng thái và Ma trận chuyển tiếp trong cấu trúc Double-Array Trie (DAT)

### Lựa chọn 1: Dàn 1 hàng ngang với Chữ trong ô tròn siêu to (36px Bold)

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'fontSize': '22px', 'fontFamily': 'arial' }}}%%
graph LR
    Root(("Root<br>s=1")):::nodeState
    H(("s=2<br>'h'")):::nodeState
    O(("s=3<br>'o'")):::nodeState
    C(("s=4<br>'c'")):::nodeState
    Under(("s=5<br>'_'")):::nodeState
    S(("s=6<br>'s'")):::nodeState
    I(("s=7<br>'i'")):::nodeState
    N(("s=8<br>'n'")):::nodeState
    H2((("s=9<br>'h'<br>Valid Word"))):::nodeValid

    Root -- "BASE[1] + 'h'" --> H
    H -- "BASE[2] + 'o'" --> O
    O -- "BASE[3] + 'c'" --> C
    C -- "BASE[4] + '_'" --> Under
    Under -- "BASE[5] + 's'" --> S
    S -- "BASE[6] + 'i'" --> I
    I -- "BASE[7] + 'n'" --> N
    N -- "BASE[8] + 'h'" --> H2

    classDef nodeState font-size:36px,font-weight:bold,fill:#eef2ff,stroke:#4338ca,stroke-width:3px;
    classDef nodeValid font-size:26px,font-weight:bold,fill:#dcfce7,stroke:#15803d,stroke-width:4px;
```

### Lựa chọn 2: Bố cục uốn 2 tầng (Tỷ lệ vuông vắn cho Báo cáo A4 / Word, chữ to rõ ràng gấp 3 lần)

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'fontSize': '20px', 'fontFamily': 'arial' }}}%%
flowchart TD
    subgraph Part1["1. Tiền tố từ đơn 'học'"]
        direction LR
        Root(("Root<br>s=1")):::nodeState -- "BASE[1] + 'h'" --> H(("s=2<br>'h'")):::nodeState
        H -- "BASE[2] + 'o'" --> O(("s=3<br>'o'")):::nodeState
        O -- "BASE[3] + 'c'" --> C(("s=4<br>'c'")):::nodeState
    end

    subgraph Part2["2. Ký tự nối và Hậu tố 'sinh'"]
        direction LR
        Under(("s=5<br>'_'")):::nodeState -- "BASE[5] + 's'" --> S(("s=6<br>'s'")):::nodeState
        S -- "BASE[6] + 'i'" --> I(("s=7<br>'i'")):::nodeState
        I -- "BASE[7] + 'n'" --> N(("s=8<br>'n'")):::nodeState
        N -- "BASE[8] + 'h'" --> H2((("s=9<br>'h'<br>Valid Word"))):::nodeValid
    end

    C -- "BASE[4] + '_'" --> Under

    classDef nodeState font-size:30px,font-weight:bold,fill:#eef2ff,stroke:#4338ca,stroke-width:3px;
    classDef nodeValid font-size:24px,font-weight:bold,fill:#dcfce7,stroke:#15803d,stroke-width:4px;
```

---

## 4. Hình 3.3: Đồ thị mạng lưới từ vựng (Word Lattice) và Đường giải mã tối ưu Viterbi HMM cho câu "học sinh học sinh học"

> **Lưu ý chuẩn hóa khi dùng trong draw.io:** Sơ đồ này phân tích ví dụ kinh điển về **Hiện tượng nhập nhằng giao thoa kép (Double Overlapping Ambiguity)** trong tiếng Việt với câu: *"học sinh học sinh học"* $\implies$ Kết quả giải mã chuẩn xác là: `["học_sinh", "học", "sinh_học"]`. Sơ đồ dùng mô hình chuẩn **Word Lattice Graph (Đồ thị mạng lưới từ vựng)** dạng DAG đơn tầng (từ trái sang phải), không sử dụng khối `subgraph` lồng ghép. Khi nhập vào draw.io (`Arrange -> Insert -> Advanced -> Mermaid`), 6 mốc Node 0 $\to$ Node 5 sẽ được dàn ngang thẳng hàng tuyệt đối, các cung từ đơn/từ ghép uốn lượn rõ ràng, đường đi tối ưu Viterbi được tô đậm màu xanh lá (`==>`).

```mermaid
flowchart LR
    N0((Node 0<br>Bắt đầu))
    N1((Node 1<br>sau học))
    N2((Node 2<br>sau sinh))
    N3((Node 3<br>sau học))
    N4((Node 4<br>sau sinh))
    N5((Node 5<br>sau học))

    %% Các cạnh từ đơn
    N0 -->|Từ đơn: học<br>Cost = 4.2| N1
    N1 -->|Từ đơn: sinh<br>Cost = 5.8| N2
    N2 -->|Từ đơn: học<br>Cost = 2.1| N3
    N3 -->|Từ đơn: sinh<br>Cost = 5.8| N4
    N4 -->|Từ đơn: học<br>Cost = 4.2| N5

    %% Các cạnh từ ghép cạnh tranh
    N1 -->|Từ ghép: sinh_học<br>Cost = 6.4 - Sai| N3
    N2 -->|Từ ghép: học_sinh<br>Cost = 7.1 - Sai| N4

    %% Đường giải mã Viterbi tối ưu (đậm nét xanh lá)
    N0 ==>|Từ ghép: học_sinh<br>Cost = 1.5 - Tối ưu| N2
    N2 ==>|Từ đơn: học<br>Cost = 2.1 - Tối ưu| N3
    N3 ==>|Từ ghép: sinh_học<br>Cost = 1.6 - Tối ưu| N5

    linkStyle 7 stroke:#16a34a,stroke-width:4px;
    linkStyle 8 stroke:#16a34a,stroke-width:4px;
    linkStyle 9 stroke:#16a34a,stroke-width:4px;
```

---

## 5. Hình 3.4: Kiến trúc lưu trữ Log-Structured Merge-Tree (LSM-Tree) và chu trình chuyển tiếp dữ liệu

```mermaid
flowchart TD
    subgraph WritePath["Luồng Ghi Dữ Liệu"]
        Msg["Bản ghi tin nhắn mới"]
        WAL[("Write-Ahead Log (wal.log)<br>Ghi tuần tự + CRC32")]
        MemTable["Bộ nhớ đệm RAM (MemTable)<br>Tra cứu tức thời"]
    end

    subgraph DiskSegments["Tầng Lưu Trữ Bất Biến Trên Đĩa (Immutable Segments)"]
        Seg1["Segment 1 (Read-Only)<br>terms.dict + postings.bin"]
        Seg2["Segment 2 (Read-Only)<br>terms.dict + postings.bin"]
        Tombstone[("Tệp Tombstone (tombstone.del)<br>Đánh dấu xóa mềm DocID")]
    end

    subgraph Background["Tiến Trình Chạy Ngầm (Compaction)"]
        Merge["Compaction Worker<br>Hợp nhất đa Segment và dọn Tombstone"]
    end

    Msg --> WAL
    Msg --> MemTable
    MemTable -- "1. Flush khi đầy RAM" --> Seg1
    MemTable -- "2. Flush định kỳ" --> Seg2
    DiskSegments --> Merge
    Tombstone -.-> Merge
```

---

## 6. Hình 4.1: Bản đồ kiến trúc hệ thống tổng thể 3 tầng

*(Có sẵn ảnh chất lượng cao tại: `image/Cốc Cốc Search Engine-2026-09-22-065029.png`)*

```mermaid
flowchart TD
    subgraph Layer1["TẦNG 1: CLIENT WEB MESSENGER UI"]
        UI_Chat["Khung Chat Thời Gian Thực (Messenger Dark Mode)"]
        UI_Inspect["Thanh Flow Inspector (Soi Token CGO)"]
        UI_Compare["Bảng Đối Soát 3 Động Cơ Song Song"]
    end

    subgraph Layer2["TẦNG 2: GOLANG REALTIME SERVER CORE (:8080)"]
        API["REST API Router / Gateway"]
        CGO["CGO Wrapper (coccoc_bridge.cpp)"]
        Norm["Normalizer (Asciifolding & Edge N-gram)"]
        Dispatcher["Dual Storage Write & Search Dispatcher"]
    end

    subgraph Layer3A["TẦNG 3A: ELASTICSEARCH 8.11"]
        ES_Mapping["Multi-field Mapping (tokenized, unaccented, partial)"]
        ES_Engine["Lucene Inverted Index & Okapi BM25 4-Tier Boosting"]
    end

    subgraph Layer3B["TẦNG 3B: CUSTOM GO STORAGE ENGINE"]
        Go_WAL[("wal.log (Append-only CRC32)")]
        Go_DocStore[("docstore.idx (12B) & docstore.dat O(1)")]
        Go_Index[("terms.dict & postings.bin Segments")]
        Go_Ranker["Pure Go BM25 Ranker (~3ms)"]
    end

    Layer1 <--> API
    API --> CGO
    CGO --> Norm
    Norm --> Dispatcher
    Dispatcher --> Layer3A
    Dispatcher --> Layer3B
```

---

## 7. Hình 4.2: Sơ đồ tuần tự Luồng Đánh chỉ mục (Indexing Pipeline) và cơ chế Dual-Write Dispatcher

*(Có sẵn ảnh chất lượng cao tại: `image/The Life of a Write Sequence Diagram.png`)*

```mermaid
sequenceDiagram
    autonumber
    actor User as Người dùng gửi tin
    participant API as Go Ingestion Server
    participant CGO as CGO C++ Tokenizer
    participant Norm as Normalizer Pipeline
    participant ES as Elasticsearch 8.11
    participant GoStore as Custom Go Storage

    User->>API: POST /api/messages {"content": "Học sinh uống cà phê"}
    activate API
    API->>CGO: Tokenize(text) qua C++ Viterbi HMM
    activate CGO
    CGO-->>API: ["học_sinh", "uống", "cà_phê"]
    deactivate CGO

    API->>Norm: Gọt dấu asciifolding & sinh Edge N-grams
    activate Norm
    Norm-->>API: Document 3 trường đa tầng hoàn chỉnh
    deactivate Norm

    par Ghi đồng thời sang Elasticsearch
        API->>ES: Gửi Bulk NDJSON Index Request
        activate ES
        ES->>ES: Ghi Translog & nạp Memory Buffer
        ES-->>API: 201 Created
        deactivate ES
    and Ghi đồng thời sang Custom Go Engine
        API->>GoStore: AppendWrite(doc)
        activate GoStore
        GoStore->>GoStore: Ghi wal.log (CRC32) + cập nhật docstore.idx (12B)
        GoStore->>GoStore: Đưa vào MemTable RAM
        GoStore-->>API: Ghi đĩa thành công
        deactivate GoStore
    end

    API-->>User: 201 Created (Thời gian xử lý: 1.78ms)
    deactivate API
```

---

## 8. Hình 4.3: Sơ đồ tuần tự Luồng Truy vấn và cơ chế Boosting 4 tầng điểm số

*(Có sẵn ảnh chất lượng cao tại: `image/Cốc Cốc Search Engine-2026-09-18-073414.png`)*

```mermaid
sequenceDiagram
    autonumber
    actor User as Người dùng tìm kiếm
    participant API as Go Search Service
    participant CGO as CGO Query Analyzer
    participant ES as Cụm Elasticsearch 8.11
    participant GoEng as Custom Go Engine

    User->>API: GET /api/search?q=hoc sinh
    activate API
    API->>CGO: Chuẩn hóa & bóc tách từ ghép truy vấn
    CGO-->>API: Term: "học_sinh", Unaccented: "hoc_sinh"

    par Tìm kiếm song song trên Elasticsearch
        API->>ES: Bool Query Boosting 4 Tầng
        activate ES
        Note over ES: Boost 5.0x: content_tokenized<br>Boost 4.0x: match_phrase<br>Boost 3.0x: content_unaccented<br>Boost 10.0x: content_partial AND
        ES->>ES: Lucene BM25 Ranking + Highlight
        ES-->>API: Top 10 kết quả chuẩn (~14.2ms)
        deactivate ES
    and Tìm kiếm song song trên Custom Go Engine
        API->>GoEng: Pure Go Search (hoc sinh)
        activate GoEng
        GoEng->>GoEng: Binary Search terms.dict O(log V)
        GoEng->>GoEng: Đọc postings.bin + Chấm điểm BM25
        GoEng->>GoEng: Truy xuất DocStore O(1) qua con trỏ 12B
        GoEng-->>API: Top 10 kết quả chuẩn (~3.0ms)
        deactivate GoEng
    end

    API->>API: sync.WaitGroup tổng hợp kết quả 3 bên
    API-->>User: Bảng đối soát 3 cột song song
    deactivate API
```

---

## 9. Hình 4.4: Sơ đồ cấu trúc nhị phân và quan hệ con trỏ giữa docstore.idx và docstore.dat

```mermaid
flowchart LR
    subgraph DocStoreIdx["Tệp docstore.idx (Mỗi bản ghi cố định 12 bytes)"]
        I0["DocID 0: [Offset: 0B, Len: 120B]"]
        I1["DocID 1: [Offset: 120B, Len: 185B]"]
        IK["DocID k: [Offset: 48,210B, Len: 140B]"]
    end

    subgraph DocStoreDat["Tệp docstore.dat (Kho dữ liệu toàn văn)"]
        D0["Vị trí byte 0: Nội dung tin nhắn Doc 0"]
        D1["Vị trí byte 120: Nội dung tin nhắn Doc 1"]
        DK["Vị trí byte 48,210: [ID=k, Sender, Room, Content, Timestamp]"]
    end

    I0 -- "ByteOffset = 0" --> D0
    I1 -- "ByteOffset = 120" --> D1
    IK -- "ByteOffset = 48,210 (Nhảy Seek O(1))" --> DK
```

---

## 10. Hình 5.1 & Hình 5.2: Biểu đồ Benchmark IR và Tài nguyên phần cứng

*(Có sẵn ảnh chất lượng cao tại: `image/benchmark_ir_comparison.jpg`)*

Bạn có thể mở tệp [docs/render_diagrams.html](file:///d:/CODE/VSF/vietnamese-chat-search/docs/render_diagrams.html) để xem và chụp ảnh trực tiếp 2 biểu đồ cột tương tác này được dựng bằng Chart.js.
