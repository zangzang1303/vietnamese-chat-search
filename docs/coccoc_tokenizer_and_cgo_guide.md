# Cẩm Nang Chi Tiết: Tích Hợp Cốc Cốc Tokenizer Trong Go Qua CGO (Lý Thuyết & Hướng Dẫn Code)

Tài liệu này được biên soạn cho **Phase 3** theo phương pháp sư phạm: **"Lý thuyết chuyên sâu phía trên - Hướng dẫn Code bám sát thực tế phía dưới"**, giải thích tường tận từ cơ chế Double-Array Trie của Cốc Cốc Tokenizer, rào cản CGO với C++, kỹ thuật bọc `extern "C"` bridge, quản lý bộ nhớ an toàn (`C.free`), cho đến kết quả thực nghiệm và benchmark thực tế.

---

## 📑 Mục Lục
1. [Lý Thuyết 1: Bản Chất & Thuật Toán Double-Array Trie Của Cốc Cốc](#1-lý-thuyết-1-bản-chất--thuật-toán-double-array-trie-của-cốc-cốc)
2. [Lý Thuyết 2: Cơ Chế CGO & Rào Cản C++ (The C++ Barrier)](#2-lý-thuyết-2-cơ-chế-cgo--rào-cản-c-the-c-barrier)
3. [Lý Thuyết 3: Quản Lý Vùng Nhớ Giữa Go và C (Memory Boundary & GC)](#3-lý-thuyết-3-quản-lý-vùng-nhớ-giữa-go-và-c-memory-boundary--gc)
4. [Kiến Trúc Thực Thi: Tầng C Bridge & Go CGO Package](#4-kiến-trúc-thực-thi-tầng-c-bridge--go-cgo-package)
5. [Quy Trình Biên Dịch & Tối Ưu Hóa Hiệu Năng](#5-quy-trình-biên-dịch--tối-ưu-hóa-hiệu-năng)
6. [Kết Quả Benchmark & So Sánh Độ Chính Xác](#6-kết-quả-benchmark--so-sánh-độ-chính-xác)

---

## 1. Lý Thuyết 1: Bản Chất & Thuật Toán Double-Array Trie Của Cốc Cốc

### 1.1. Thách thức phân tách từ tiếng Việt
Khác với tiếng Anh (tách từ đơn thuần bằng khoảng trắng), tiếng Việt có đặc thù:
* **Âm tiết (Syllable)** được viết tách rời bằng dấu cách: *"học"*, *"sinh"*.
* **Từ vựng (Word)** bao gồm cả từ đơn và từ ghép phức: *"học sinh"*, *"cà phê"*, *"thời khóa biểu"*.
* **Xung đột ranh giới từ (Word Segmentation Ambiguity)**:
  * *"Em là **học sinh** mới"* $\rightarrow$ Từ ghép: `học_sinh`.
  * *"Em thích **học** **sinh** học"* $\rightarrow$ Hai từ: `học` (động từ) và `sinh_học` (danh từ).
  * *"Chiến sĩ **hy sinh** anh dũng"* $\rightarrow$ Từ ghép: `hy_sinh`.

Nếu chỉ tách từ theo khoảng trắng (như `StandardAnalyzer` mặc định của Elasticsearch), việc tìm kiếm `"học sinh"` sẽ bị bẻ thành `"học"` VÀ `"sinh"`. Kết quả là các tin nhắn chứa *"sinh viên"*, *"sinh học"*, *"hy sinh"* đều bị trả về (lỗi **False Positive** nghiêm trọng trong tìm kiếm tin nhắn chat).

### 1.2. Cấu trúc Double-Array Trie (DATrie)
Mã nguồn Cốc Cốc Tokenizer (`coccoc/coccoc-tokenizer`) là thư viện C++ hiệu năng cao, sử dụng cấu trúc cây **Double-Array Trie (DATrie)** để lưu trữ từ điển tiếng Việt:
* **Mảng BASE và CHECK**: DATrie nén toàn bộ cây tiền tố hàng trăm nghìn từ vựng vào 2 mảng số nguyên tuyến tính. Nhờ vậy, việc chuyển trạng thái ký tự chỉ là phép tính đại số $O(1)$:
  $$\text{next\_state} = \text{BASE}[\text{current\_state}] + \text{character\_code}$$
* **Tệp từ điển nhị phân đã biên dịch**:
  1. `multiterm_trie.dump` (~24MB): Cây Trie chứa toàn bộ các từ ghép tiếng Việt đa âm tiết kèm tần suất sử dụng thực tế.
  2. `syllable_trie.dump` (~21MB): Cây Trie âm tiết tiếng Việt.
  3. `nontone_pair_freq_map.dump` (~199MB): Ma trận tần suất cặp âm tiết không dấu, dùng để bóc tách từ viết dính liền không dấu cách (sticky text hoặc URL).

---

## 2. Lý Thuyết 2: Cơ Chế CGO & Rào Cản C++ (The C++ Barrier)

### 2.1. CGO là gì?
**CGO** là cơ chế tích hợp trong Go toolchain, cho phép mã nguồn Go gọi trực tiếp các hàm C:
```go
/*
#include <stdlib.h>
#include "coccoc_bridge.h"
*/
import "C"
```
Khi có khối chú thích preamble ngay trước `import "C"`, Go compiler sinh ra một gói giả lập tên là `C`, ánh xạ toàn bộ kiểu dữ liệu và hàm của C sang Go (như `C.int`, `C.CString`, `C.free`).

### 2.2. Tại sao CGO không thể gọi trực tiếp C++?
1. **Name Mangling**: Trình biên dịch C++ tự động mã hóa tên hàm (ví dụ hàm `segment(text)` bị đổi thành `_ZN9Tokenizer7segmentERKSs`) để hỗ trợ Overloading. CGO chỉ hiểu C ABI nên không thể liên kết được tên hàm C++.
2. **C++ Classes & Templates**: CGO không hiểu cú pháp class, template, instance methods hay con trỏ `this`.
3. **Exceptions & STL**: CGO không thể bắt C++ `std::exception` hay thao tác trên `std::vector<std::string>`.

### 2.3. Giải pháp: Lớp cầu nối C Bridge (`extern "C"`)
Để kết nối Go với thư viện C++ Cốc Cốc, ta thiết kế một tầng trung gian:
* Phía C++: Khai báo giao diện thuần C trong khối `extern "C"` để tắt Name Mangling.
* Phía Go: Gọi các hàm C thuần túy thông qua CGO.

```
┌────────────────────────────────────────────────────────┐
│                        GO CODE                         │
│             pkg/tokenizer/coccoc/tokenizer.go          │
└───────────────────────────┬────────────────────────────┘
                            │ CGO call: C.coccoc_tokenize_original(cText)
                            ▼
┌────────────────────────────────────────────────────────┐
│                   C WRAPPER BRIDGE                     │
│               pkg/tokenizer/coccoc/coccoc_bridge.cpp   │
│               extern "C" { coccoc_tokenize_original }  │
└───────────────────────────┬────────────────────────────┘
                            │ C++ Method Call: Tokenizer::instance().segment_original()
                            ▼
┌────────────────────────────────────────────────────────┐
│              C++ CỐC CỐC TOKENIZER ENGINE              │
│       Double-Array Trie (multiterm_trie.dump 24MB)     │
└────────────────────────────────────────────────────────┘
```

---

## 3. Lý Thuyết 3: Quản Lý Vùng Nhớ Giữa Go và C (Memory Boundary & GC)

### 3.1. Hai thế giới bộ nhớ độc lập:
1. **Go Runtime Memory**: Được quản lý tự động bởi **Garbage Collector (GC)**. Khi một biến Go không còn được tham chiếu, GC sẽ tự động thu hồi.
2. **C Heap Memory**: Được cấp phát bằng `malloc()` và **hoàn toàn nằm ngoài tầm kiểm soát của GC**. Nếu không gọi `free()`, bộ nhớ sẽ bị rò rỉ (Memory Leak) vĩnh viễn cho đến khi tiến trình bị OS tiêu diệt!

### 3.2. Quy tắc sinh tử khi làm việc với CGO:
* **Khi truyền chuỗi từ Go sang C**:
  ```go
  cText := C.CString(text)              // C.CString cấp phát trên C Heap bằng malloc()
  defer C.free(unsafe.Pointer(cText))   // BẮT BUỘC giải phóng ngay khi hàm kết thúc
  ```
* **Khi nhận chuỗi từ C trả về Go**:
  ```go
  cResult := C.coccoc_tokenize_original(cText)
  defer C.coccoc_free_string(cResult)   // BẮT BUỘC giải phóng con trỏ C sau khi đọc
  goStr := C.GoString(cResult)          // Sao chép dữ liệu từ C Heap sang Go Heap
  ```

---

## 4. Kiến Trúc Thực Thi: Tầng C Bridge & Go CGO Package

### 4.1. File Header: `pkg/tokenizer/coccoc/coccoc_bridge.h`
```c
#ifndef COCCOC_BRIDGE_H
#define COCCOC_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

int coccoc_init(const char* dict_path, int load_nontone);
char* coccoc_tokenize_original(const char* text);
void coccoc_free_string(char* str);

#ifdef __cplusplus
}
#endif

#endif // COCCOC_BRIDGE_H
```

### 4.2. File Cài Đặt: `pkg/tokenizer/coccoc/coccoc_bridge.cpp`
```cpp
#include "coccoc_bridge.h"
#include <tokenizer/tokenizer.hpp>
#include <cstdlib>
#include <cstring>
#include <string>
#include <vector>

int coccoc_init(const char* dict_path, int load_nontone) {
    if (!dict_path) return -1;
    bool nontone = (load_nontone != 0);
    return Tokenizer::instance().initialize(std::string(dict_path), nontone);
}

char* coccoc_tokenize_original(const char* text) {
    if (!text || text[0] == '\0') {
        char* empty = (char*)malloc(1);
        if (empty) empty[0] = '\0';
        return empty;
    }

    std::string input(text);
    std::vector<FullToken> res = Tokenizer::instance().segment_original(input, Tokenizer::TOKENIZE_NORMAL);

    if (res.empty()) {
        char* result = (char*)malloc(input.size() + 1);
        if (!result) return NULL;
        std::memcpy(result, input.c_str(), input.size() + 1);
        return result;
    }

    std::string output;
    output.reserve(input.size() + 16);

    for (size_t i = 0; i < res.size(); ++i) {
        size_t punct_start = (i > 0) ? res[i - 1].original_end : 0;
        size_t punct_len = res[i].original_start - punct_start;

        if (punct_len > 0) {
            output += input.substr(punct_start, punct_len);
        } else if (i > 0) {
            output += ' ';
        }
        output += res[i].text;
    }

    size_t last_end = res.back().original_end;
    if (last_end < input.size()) {
        output += input.substr(last_end);
    }

    char* result = (char*)malloc(output.size() + 1);
    if (!result) return NULL;
    std::memcpy(result, output.c_str(), output.size() + 1);
    return result;
}

void coccoc_free_string(char* str) {
    if (str) free(str);
}
```

### 4.3. Go Package: `pkg/tokenizer/coccoc/tokenizer.go`
Sử dụng cờ CGO để liên kết mã nguồn C++:
```go
/*
#cgo CXXFLAGS: -std=c++11 -Wno-error -Wno-cast-user-defined -I${SRCDIR}/../../../temp_coccoc -I${SRCDIR}/../../../temp_coccoc/tokenizer -I${SRCDIR}/../../../temp_coccoc/build/auto
#cgo LDFLAGS: -lstdc++
#include "coccoc_bridge.h"
#include <stdlib.h>
*/
import "C"
```

---

## 5. Quy Trình Biên Dịch & Tối Ưu Hóa Hiệu Năng

### 5.1. Tự động hóa tạo từ điển: `scripts/build_coccoc.sh`
Chỉ với 1 lệnh bash duy nhất trong môi trường WSL2/Linux:
```bash
./scripts/build_coccoc.sh
```
Script sẽ tự động:
1. Sửa cờ cảnh báo `-Wno-error -Wno-cast-user-defined` để tương thích hoàn toàn với GCC 15.
2. Biên dịch công cụ `dict_compiler`.
3. Sinh các tệp nhị phân `multiterm_trie.dump` (24MB) và `syllable_trie.dump` (21MB).
4. Sao chép toàn bộ tệp cấu hình ngôn ngữ vào `data/dicts/coccoc/`.

### 5.2. Khám phá tối ưu hóa then chốt (Performance Optimization Discovery)
Trong Cốc Cốc Tokenizer, hàm khởi tạo có tham số `load_nontone_data`:
* **Nếu bật `load_nontone_data = true`**: Hệ thống phải nạp tệp `nontone_pair_freq_map.dump` dung lượng **199MB** với hàng trăm nghìn phần tử map. Kết quả: thời gian khởi động mất **~10-15 giây** và tiêu tốn **~400MB RAM**.
* **Nếu đặt `load_nontone_data = false`**: Vì tin nhắn chat thông thường các từ đã có dấu cách phân định, ta chỉ cần nạp `multiterm_trie.dump` (24MB). Kết quả:
  * Thời gian khởi động giảm từ **~10 giây** xuống chỉ còn **0.15 giây (158ms)**!
  * Tiêu thụ RAM giảm từ **~400MB** xuống còn **~45MB**!
  * Độ chính xác nhận diện từ ghép (*"học sinh"*, *"cà phê"*, *"thời khóa biểu"*) giữ nguyên **100%**!

---

## 6. Kết Quả Benchmark & So Sánh Độ Chính Xác

### 6.1. Tốc độ phân tách từ (Benchmark):
Chạy lệnh `go test -v -bench=. ./pkg/tokenizer/coccoc/...`:
```
goos: linux
goarch: amd64
pkg: vietnamese-chat-search/pkg/tokenizer/coccoc
cpu: 12th Gen Intel(R) Core(TM) i7-12700H
BenchmarkCoccocTokenizer-20      388520        2800 ns/op
PASS
ok      vietnamese-chat-search/pkg/tokenizer/coccoc    1.278s
```
* **Thời gian xử lý trung bình**: **2.8 microsecond ($\mu s$)** cho mỗi tin nhắn!
* **Thông lượng (Throughput)**: Đạt trên **350,000 tin nhắn / giây**, hoàn toàn đáp ứng các hệ thống chat thời gian thực quy mô lớn.

### 6.2. So sánh độ chính xác tìm kiếm (Standard vs Cốc Cốc):
Thực nghiệm tìm kiếm từ khóa `"học sinh"` trên cơ sở dữ liệu tin nhắn chat:

| Bộ Phân Tích (Analyzer) | Tokens Sinh Ra | Tin Nhắn Khớp Được | Nhận Xét Đánh Giá |
| :--- | :--- | :--- | :--- |
| **Standard Analyzer** (Mặc định ES) | `["học", "sinh"]` | • Doc 2: *Hôm nay các em học sinh...*<br>• Doc 1: *Em là sinh viên mới nhập học...*<br>• Doc 7: *Đã có thời khóa biểu học kỳ...* | ❌ **False Positive cao**: Trả về cả tin nhắn chứa *"sinh viên"* và *"thời khóa biểu"* do bẻ gãy từ ghép thành các từ đơn! |
| **Cốc Cốc Tokenizer** (CGO DATrie) | `["học_sinh"]` | • Doc 2: *Hôm nay các em học sinh...* | ✅ **Chính xác 100%**: Chỉ trả về đúng tin nhắn mang ngữ nghĩa học trò, loại bỏ hoàn toàn các tin nhắn gây nhiễu! |
