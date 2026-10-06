# TÀI LIỆU KẾ HOẠCH CHI TIẾT: TASK 2
## ĐÓNG GÓI CỐC CỐC TOKENIZER THÀNH GO MODULE ĐỘC LẬP (STANDALONE GITHUB REPOSITORY)

---

## 1. TỔNG QUAN & TẦM NHÌN DỰ ÁN

### 1.1. Bối cảnh kỹ thuật
Hiện tại, bộ tách từ tiếng Việt Cốc Cốc Tokenizer đang nằm tích hợp bên trong dự án monolithic (`pkg/tokenizer/coccoc`), phụ thuộc vào đường dẫn tương đối tới mã nguồn C++ và tệp từ điển (`../../../temp_coccoc`, `data/dicts/coccoc`). Điều này gây ra các hạn chế lớn:
* Không thể tái sử dụng cho các microservices hoặc các dự án Go khác trong tổ chức.
* Quy trình build phức tạp, đòi hỏi cấu hình CGO thủ công và biên dịch C++ rời rạc.
* Khó quản lý phiên bản (versioning) và cập nhật độc lập khi có cải tiến về từ điển hoặc thuật toán tách từ.

### 1.2. Mục tiêu trọng tâm của Task 2
Tách toàn bộ phần tích hợp Cốc Cốc Tokenizer thành một **Go Module độc lập chất lượng cao (Production-grade Open-source Package)** trên GitHub (ví dụ: `github.com/<org>/coccoc-tokenizer-go`):
1. **Dễ dàng cài đặt & Tích hợp**: Hỗ trợ cài đặt trực tiếp qua lệnh chuẩn:
   ```bash
   go get github.com/<org>/coccoc-tokenizer-go
   ```
2. **Kiến trúc Dual-Mode (CGO Engine & Pure-Go Fallback)**:
   * **Môi trường Production (Linux Docker / Kubernetes)**: Kích hoạt CGO và liên kết với C++ Engine để đạt tốc độ xử lý hàng triệu token/giây.
   * **Môi trường Local Dev / CI đơn giản (Windows, macOS không có GCC)**: Tự động fallback về thuật toán Pure-Go (Max-Matching / Trie) mà không cần cài đặt C++ compiler.
3. **An toàn bộ nhớ tuyệt đối (Zero Memory Leak)**: Quản lý con trỏ C/Go chặt chẽ, đảm bảo không rò rỉ bộ nhớ dù chạy hàng tỷ requests.
4. **Thread-Safe & High Concurrency**: Hỗ trợ gọi đồng thời an toàn từ hàng nghìn goroutines trong các dịch vụ web/chat backend.
5. **CI/CD Tự động hóa hoàn chỉnh**: Quy trình GitHub Actions kiểm thử đa nền tảng (Ubuntu, Alpine, macOS, Windows), tự động release version theo SemVer.

---

## 2. THIẾT KẾ CẤU TRÚC REPOSITORY & PACKAGE LAYOUT

Một cấu trúc repository chuẩn cho thư viện Go tích hợp C/C++:

```
coccoc-tokenizer-go/
├── .github/
│   └── workflows/
│       ├── test.yml            # CI: Chạy test trên Linux (CGO + Non-CGO), macOS, Windows
│       ├── benchmark.yml       # CI: Đo lường regression về hiệu năng và allocations
│       └── release.yml         # CI: Tự động gắn git tag, publish release assets
├── csrc/                       # Toàn bộ mã nguồn C/C++ phục vụ biên dịch CGO
│   ├── include/
│   │   ├── coccoc_bridge.h     # Interface extern "C" chuẩn, không bị mangling
│   │   └── tokenizer/          # Cốc Cốc C++ headers gốc
│   ├── src/
│   │   ├── coccoc_bridge.cpp   # Implementation của C bridge
│   │   └── tokenizer/          # Mã nguồn C++ Cốc Cốc (được nhúng để go build tự biên dịch)
│   └── CMakeLists.txt          # Dành cho trường hợp người dùng muốn build dynamic/static lib rời
├── dicts/                      # Dữ liệu từ điển nhị phân đã biên dịch sẵn
│   ├── multiterm_trie.dump     # Từ điển âm tiết ghép (~24.5 MB)
│   ├── syllable_trie.dump      # Từ điển âm tiết đơn (~21.4 MB)
│   ├── nontone_pair_freq_map.dump # Bảng tần suất không dấu (tùy chọn, ~208 MB)
│   └── words.txt               # Danh sách từ ghép rút gọn cho Pure-Go mode
├── examples/                   # Ví dụ mẫu hướng dẫn sử dụng
│   ├── basic/main.go           # Ví dụ cơ bản: Tách từ một câu đơn
│   ├── http_server/main.go     # Ví dụ tích hợp vào Gin / Fiber web server
│   └── opensearch_indexer/     # Ví dụ tiền xử lý dữ liệu trước khi đẩy vào OpenSearch
├── tokenizer.go                # Public API & Core struct
├── tokenizer_cgo.go            # CGO implementation (//go:build cgo)
├── tokenizer_pure.go           # Pure Go implementation (//go:build !cgo)
├── options.go                  # Functional Options Pattern cho cấu hình khởi tạo
├── types.go                    # Định nghĩa Token, Segment, Tokenizer Interface
├── tokenizer_test.go           # Unit tests kiểm tra tính chính xác tách từ
├── benchmark_test.go           # Benchmark đo ns/op, B/op, allocs/op
├── leak_test.go                # Stress-test kiểm tra rò rỉ bộ nhớ CGO
├── go.mod
├── go.sum
├── LICENSE                     # Giấy phép bản quyền (LGPL-2.1 hoặc MIT tùy biến)
└── README.md                   # Tài liệu hướng dẫn sử dụng toàn diện
```

---

## 3. THIẾT KẾ PUBLIC API CHUẨN IDIOMATIC GO

### 3.1. Interface & Factory Pattern

Thư viện định nghĩa giao diện trừu tượng để người dùng có thể dễ dàng mock trong unit test của ứng dụng:

```go
// Tokenizer đại diện cho bộ tách từ tiếng Việt
type Tokenizer interface {
    // Segment phân tách câu và ghép các từ ghép bằng dấu gạch dưới ("_")
    // Ví dụ: "Tôi yêu Việt Nam" -> "Tôi yêu Việt_Nam"
    Segment(text string) (string, error)

    // Tokenize phân đoạn câu và trả về danh sách các tokens đã chuẩn hóa (lowercase, bỏ dấu câu)
    // Ví dụ: "Tôi yêu Việt Nam!" -> ["tôi", "yêu", "việt_nam"]
    Tokenize(text string) ([]string, error)

    // TokenizeFull trả về danh sách token chi tiết kèm offset và loại từ
    TokenizeFull(text string) ([]TokenInfo, error)

    // Close giải phóng tài nguyên native C++ (nếu có)
    Close() error
}

type TokenInfo struct {
    Text      string // Chuỗi token
    StartByte int    // Vị trí byte bắt đầu trong văn bản gốc
    EndByte   int    // Vị trí byte kết thúc trong văn bản gốc
    Type      int    // Phân loại token (từ đơn, từ ghép, số, URL...)
}
```

### 3.2. Cấu hình bằng Functional Options Pattern

Cho phép cấu hình linh hoạt mà vẫn giữ API đơn giản, dễ mở rộng:

```go
// Khởi tạo thông thường (sử dụng đường dẫn thư mục từ điển mặc định)
tok, err := coccoc.New()

// Khởi tạo tùy biến chuyên sâu
tok, err := coccoc.New(
    coccoc.WithDictPath("/usr/local/share/tokenizer/dicts"),
    coccoc.WithLoadNontone(false), // false: Tối ưu bộ nhớ ~45MB; true: Nạp bảng tần suất 200MB
    coccoc.WithNormalizeLowercase(true),
    coccoc.WithFallbackToPureGo(true), // Tự động dùng Pure-Go nếu môi trường không có CGO
)
defer tok.Close()
```

---

## 4. CHI TIẾT KỸ THUẬT: CGO BRIDGE, MEMORY SAFETY & PURE-GO FALLBACK

### 4.1. Cơ chế CGO Bridge & Zero-Leak Memory Management
Trong giao tiếp giữa Go và C++, các vấn đề rò rỉ bộ nhớ (memory leak) là rủi ro hàng đầu. Quy tắc quản lý bộ nhớ nghiêm ngặt được thiết lập như sau:

```go
// tokenizer_cgo.go
//go:build cgo

package coccoc

/*
#cgo CXXFLAGS: -std=c++11 -O3 -I${SRCDIR}/csrc/include
#cgo LDFLAGS: -lstdc++
#include "csrc/include/coccoc_bridge.h"
#include <stdlib.h>
*/
import "C"
import (
    "errors"
    "unsafe"
)

func (t *cgoTokenizer) Segment(text string) (string, error) {
    if text == "" {
        return "", nil
    }

    // 1. Chuyển chuỗi Go sang chuỗi C trên heap
    cText := C.CString(text)
    // BẮT BUỘC: Đảm bảo chuỗi C được giải phóng ngay sau khi kết thúc hàm
    defer C.free(unsafe.Pointer(cText))

    // 2. Gọi C Bridge
    cResult := C.coccoc_tokenize_original(cText)
    if cResult == nil {
        return "", errors.New("coccoc: lỗi phân tích chuỗi từ C++ engine")
    }
    // BẮT BUỘC: Giải phóng chuỗi trả về từ C++ heap
    defer C.coccoc_free_string(cResult)

    // 3. Sao chép kết quả an toàn sang chuỗi do Go runtime quản lý
    return C.GoString(cResult), nil
}
```

### 4.2. Tối ưu Đa luồng (Thread-Safety & Lock Elimination)
* Mã nguồn gốc C++ của Cốc Cốc Tokenizer sử dụng mẫu thiết kế Singleton `Tokenizer::instance()`.
* Cần kiểm tra kỹ: Phương thức `segment_original()` của C++ engine có phải là hàm *re-entrant* (tái nhập, thread-safe khi đọc đồng thời từ từ điển) hay không?
  * Nếu C++ engine chỉ đọc cấu trúc Trie (`read-only`) trên bộ nhớ mà không ghi vào biến static/toàn cục: **Loại bỏ hoàn toàn `sync.Mutex`** phía Go, cho phép hàng nghìn goroutines gọi đồng thời song song mà không bị tranh chấp lock (Zero Contention).
  * Nếu C++ engine có sử dụng trạng thái dùng chung (internal mutable buffer): Thiết kế cơ chế **Worker Pool** hoặc `sync.Pool` chứa các đối tượng tokenizer độc lập để tránh nghẽn cổ chai đa luồng.

### 4.3. Kiến trúc Pure-Go Fallback (Chế độ Không cần CGO)
Xây dựng một engine tách từ hoàn toàn bằng Go trong `tokenizer_pure.go` với build tag `//go:build !cgo`:
* Áp dụng thuật toán **Forward Maximum Matching (FMM)** hoặc **Double-Array Trie (DAT)** trên danh sách từ điển từ ghép tiếng Việt chuẩn (~45.000 từ).
* Loại bỏ sự phụ thuộc vào GCC/G++, cho phép biên dịch tĩnh hoàn toàn (`CGO_ENABLED=0`) ra file nhị phân độc lập (scratch container) chỉ vài MB.

---

## 5. CHIẾN LƯỢC QUẢN LÝ TỪ ĐIỂN (DICTIONARY DISTRIBUTION)

Bộ từ điển của Cốc Cốc Tokenizer bao gồm các tệp nhị phân có dung lượng đáng kể:
* `multiterm_trie.dump` (~24.5 MB)
* `syllable_trie.dump` (~21.4 MB)
* `nontone_pair_freq_map.dump` (~208 MB)

Để module thuận tiện và không làm phình dung lượng git repo, kế hoạch triển khai 3 cấp độ phân phối từ điển:
1. **Local System Path (Ưu tiên sản xuất)**:
   * Cho phép trỏ đến thư mục cài đặt từ điển trên hệ thống (ví dụ: `/usr/local/share/tokenizer/dicts` hoặc `/etc/coccoc/dicts`).
   * Tự động kiểm tra biến môi trường `COCCOC_DICT_PATH`.
2. **Auto-Downloader (Tải tự động)**:
   * Cung cấp một package con `github.com/.../coccoc-tokenizer-go/dictutil` có tính năng tự động tải tệp từ điển đã nén từ **GitHub Release Assets** về thư mục cache cục bộ (`~/.cache/coccoc/dicts`) nếu máy tính chưa có.
3. **Embedded Core Dictionary (Dành cho Pure-Go Mode)**:
   * Sử dụng tính năng Go `//go:embed` nhúng tệp danh sách từ ghép rút gọn (~500KB) trực tiếp vào binary Go, giúp thư viện luôn sẵn sàng hoạt động ngay lập tức mà không cần bất kỳ file ngoài nào.

---

## 6. CHIẾN LƯỢC KIỂM THỬ TOÀN DIỆN (TESTING & QUALITY ASSURANCE)

### 6.1. Ma trận Kiểm thử Unit Test (Accuracy & Correctness)
Xây dựng bộ test cases phong phú trong `tokenizer_test.go`:
* **Từ ghép cơ bản**: "học sinh", "cà phê", "sinh viên", "Việt Nam", "thời khóa biểu".
* **Hiện tượng đa nghĩa & chồng lấn**: "học sinh học sinh học", "bàn ghế bàn bạc".
* **Ký tự đặc biệt, Dấu câu & Cấu trúc phức tạp**: "Hà Nội, thủ đô của Việt Nam!", "Giá: 50.000đ/ly cà phê.", "Email: info@coccoc.com".
* **Xen lẫn tiếng Anh và tiếng Việt**: "Hôm nay tôi deploy code lên OpenSearch cluster".
* **Kiểm tra tính nhất quán (Parity Test)**: Đảm bảo kết quả tách từ giữa chế độ CGO và chế độ Pure-Go đối với các từ ghép phổ thông đạt độ tương đồng > 95%.

### 6.2. Kiểm thử Rò rỉ Bộ nhớ (Memory Leak Audit)
* **Go Loop Test**: Viết test chạy liên tục 1.000.000 lần tách từ trong một tiến trình Go, sử dụng `runtime.ReadMemStats` và `OS Process RSS` để khẳng định bộ nhớ RAM giữ nguyên ở mức phẳng (flat line), không tăng tịnh tiến theo thời gian.
* **Valgrind / ASan (AddressSanitizer)**: Thiết lập pipeline chạy trên Linux với cờ `-fsanitize=address` để phát hiện bất kỳ con trỏ C/C++ nào bị leak hoặc use-after-free.

### 6.3. Đo lường Hiệu năng (Micro-benchmarking)
Thực thi lệnh chuẩn:
```bash
go test -bench=. -benchmem -benchtime=5s
```
Đo lường các chỉ số quan trọng:
* Thời gian xử lý: `ns/op` cho một câu ngắn (10 từ), câu trung bình (50 từ) và văn bản dài (200 từ).
* Mức cấp phát bộ nhớ Go: `B/op` và `allocs/op` (Mục tiêu: Cực tiểu hóa heap allocation ở tầng Go wrapper).
* Throughput: Megabytes xử lý mỗi giây ($\text{MB/s}$).

---

## 7. QUY TRÌNH CI/CD, TÀI LIỆU HÓA & PHÁT HÀNH GITHUB REPO

### 7.1. Thiết lập GitHub Actions Workflows

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [ master, main ]
  pull_request:
    branches: [ master, main ]

jobs:
  test-linux-cgo:
    name: Linux CGO (Ubuntu + GCC)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Install C++ Build Tools
        run: sudo apt-get update && sudo apt-get install -y build-essential cmake
      - name: Run Tests with CGO
        env:
          CGO_ENABLED: 1
        run: go test -v -race ./...

  test-pure-go:
    name: Pure Go (No CGO, Multi-platform)
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Run Tests without CGO
        env:
          CGO_ENABLED: 0
        run: go test -v ./...
```

### 7.2. Tài liệu Hướng dẫn (README.md) & Example Guides
Soạn thảo tài liệu chuẩn quốc tế với đầy đủ các mục:
1. **Badges**: Go Report Card, CI Status, License, GoDoc reference.
2. **Quickstart Guide**: Hướng dẫn cài đặt và 3 dòng code chạy được ngay.
3. **Architecture Overview**: Sơ đồ giải thích cơ chế CGO bridge và Pure-Go fallback.
4. **Performance Benchmarks**: Bảng số liệu chi tiết đo lường thông lượng so sánh giữa CGO mode và Pure-Go mode.
5. **Docker Integration**: Hướng dẫn viết Dockerfile multi-stage để build ứng dụng Go chứa module Cốc Cốc với dung lượng image tối ưu.

---

## 8. KẾ HOẠCH TRIỂN KHAI THEO TỪNG GIAI ĐOẠN (ROADMAP)

| Giai đoạn | Nội dung công việc chi tiết | Thời gian | Sản phẩm đầu ra (Deliverable) |
| :--- | :--- | :--- | :--- |
| **Phase 1** | - Khởi tạo cấu trúc repository Go module mới.<br>- Tách lọc mã nguồn C++ Cốc Cốc cần thiết và thiết kế `coccoc_bridge.h/.cpp` sạch sẽ.<br>- Thiết lập `go.mod` và khai báo `#cgo` build flags chuẩn xác. | Ngày 1 | Skeleton Repo hoàn chỉnh, build CGO thành công độc lập |
| **Phase 2** | - Cài đặt Public API Surface (`New`, `Segment`, `Tokenize`, `Options`).<br>- Hoàn thiện quản lý con trỏ C/Go đảm bảo Zero Memory Leak.<br>- Cải tiến và tối ưu Pure-Go Max-Matching Tokenizer (`//go:build !cgo`). | Ngày 2 | Core engine hoạt động mượt mà ở cả 2 chế độ CGO và Non-CGO |
| **Phase 3** | - Viết toàn diện bộ Unit Test, Concurrency Test và Memory Leak Test.<br>- Đo micro-benchmarks và tối ưu allocations.<br>- Xây dựng package hỗ trợ tải và nạp từ điển tự động. | Ngày 3 | Test suite phủ > 85% coverage, Benchmark numbers chuẩn |
| **Phase 4** | - Viết các ví dụ mẫu trong thư mục `examples/`.<br>- Hoàn thiện `README.md` chuyên nghiệp.<br>- Cấu hình GitHub Actions CI/CD và sẵn sàng tạo Repo riêng trên GitHub. | Ngày 4 | Ready-to-publish GitHub Repository (`v1.0.0`) |
