# Cẩm Nang Chi Tiết: Tích Hợp Cốc Cốc Tokenizer Trong Go Qua CGO (Lý Thuyết & Hướng Dẫn Code)

Tài liệu này được biên soạn cho **Phase 3 (Day 3)** theo phương pháp sư phạm: **"Lý thuyết chuyên sâu phía trên - Hướng dẫn Code bám sát phía dưới"**, giải thích chi tiết từ cơ chế hoạt động của Cốc Cốc Tokenizer, bản chất CGO đến kỹ thuật bọc C++ (`extern "C"` wrapper) và quản lý bộ nhớ an toàn trong Go.

---

## 📑 Mục Lục
1. [Lý Thuyết 1: Bản Chất & Thuật Toán Của Cốc Cốc Tokenizer](#1-lý-thuyết-1-bản-chất--thuật-toán-của-cốc-cốc-tokenizer)
2. [Lý Thuyết 2: Cơ Chế CGO & Rào Cản C++ (The C++ Barrier)](#2-lý-thuyết-2-cơ-chế-cgo--rào-cản-c-the-c-barrier)
3. [Lý Thuyết 3: Quản Lý Vùng Nhớ Giữa Go và C (Memory Boundary & GC)](#3-lý-thuyết-3-quản-lý-vùng-nhớ-giữa-go-và-c-memory-boundary--gc)
4. [Hướng Dẫn Code Bước 1: Thiết Kế Lớp C Wrapper (`extern "C"`)](#4-hướng-dẫn-code-bước-1-thiết-kế-lớp-c-wrapper-extern-c)
5. [Hướng Dẫn Code Bước 2: Xây Dựng Package Go CGO (`pkg/tokenizer/coccoc.go`)](#5-hướng-dẫn-code-bước-2-xây-dựng-package-go-cgo-pkgtokenizercoccocgo)
6. [Hướng Dẫn Code Bước 3: Quản Lý Biên Dịch & Chạy Thử Trên Docker](#6-hướng-dẫn-code-bước-3-quản-lý-biên-dịch--chạy-thử-trên-docker)

---

## 1. Lý Thuyết 1: Bản Chất & Thuật Toán Của Cốc Cốc Tokenizer

### 1.1. Cốc Cốc Tokenizer hoạt động như thế nào?
Khác với các ngôn ngữ phương Tây (tách từ dựa vào khoảng trắng), tiếng Việt có đặc thù:
* **Âm tiết (Syllable)** được viết tách nhau bởi khoảng trắng: *"học"*, *"sinh"*.
* **Từ vựng (Word)** có thể là từ đơn (*"bàn"*, *"ghế"*) hoặc từ ghép đa âm tiết (*"học sinh"*, *"cà phê"*, *"thời khóa biểu"*).
* **Hiện tượng đa nghĩa theo ngữ cảnh**:
  * Câu A: *"Các em **học sinh** chăm chỉ"* $\rightarrow$ Từ ghép: `học_sinh`.
  * Câu B: *"Em thích **học** **sinh** học"* $\rightarrow$ Hai từ riêng biệt: `học` (động từ) và `sinh_học` (danh từ môn học).

### 1.2. Cấu trúc Từ điển `sys.dic` và Thuật toán Trie đồ thị
* **Từ điển `sys.dic`**: Là file nhị phân nén chứa hàng trăm nghìn từ vựng tiếng Việt cùng tần suất xuất hiện (frequency weight) được Cốc Cốc thống kê từ hàng tỷ trang web tiếng Việt.
* **Mô hình hóa đồ thị câu (Word Lattice Graph)**:
  Khi nhận một câu, Cốc Cốc Tokenizer tra cứu trong cây tiền tố (Trie) để tìm tất cả các cách cắt từ có thể xảy ra:
  ```
  (Tôi) ───> (học) ───> (sinh) ───> (học)
         └──> (học_sinh) ───┘
  ```
* **Thuật toán Viterbi / Dynamic Programming**:
  Mỗi đường đi qua đồ thị được tính một trọng số xác suất dựa trên tần suất từ trong `sys.dic`. Thuật toán tìm kiếm đường đi ngắn nhất (Shortest Path) sẽ chọn ra cách phân đoạn có tổng điểm xác suất cao nhất.

---

## 2. Lý Thuyết 2: Cơ Chế CGO & Rào Cản C++ (The C++ Barrier)

### 2.1. CGO là gì?
**CGO** là một cơ chế đặc biệt tích hợp sẵn trong trình biên dịch Go, cho phép mã nguồn Go gọi trực tiếp các hàm được viết bằng ngôn ngữ **C**.

Khi bạn khai báo:
```go
/*
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
```
Trình biên dịch Go sẽ tự động phân tích khối chú thích (preamble) phía trên `import "C"` như là mã nguồn C và tạo ra một pseudo-package tên là `C` để Go truy cập các kiểu dữ liệu và hàm của C (ví dụ: `C.int`, `C.free()`, `C.puts()`).

### 2.2. "Rào cản C++" (The C++ Barrier): Tại sao CGO không gọi trực tiếp C++ được?
* CGO **chỉ hiểu cú pháp thuần C (ANSI C ABI)**.
* **C++** có những tính năng mà C không có:
  1. *Name Mangling*: Trình biên dịch C++ tự động đổi tên hàm trong file nhị phân (ví dụ: hàm `tokenize(string)` bị đổi thành `_Z8tokenizeNSt7__cxx1112basic_string...`) để hỗ trợ Function Overloading. CGO không thể tìm thấy hàm với tên gốc!
  2. *Classes, Methods & `this` pointer*: CGO không hiểu cú pháp gọi `object.method()`.
  3. *Templates, Exceptions, Namespaces*.

### 2.3. Giải pháp: Lớp cầu nối `extern "C"` (C Wrapper Bridge)
Để Go gọi được thư viện C++ Cốc Cốc, ta bắt buộc phải viết một **lớp bọc C trung gian (C Bridge)**:
* Phía C++: Viết các hàm C thuần túy nhưng bên trong gọi code C++, bọc trong khối `extern "C"` để tắt tính năng Name Mangling.
* Phía Go: CGO chỉ cần gọi các hàm C này!

```
┌────────────────────────────────────────────────────────┐
│                        GO CODE                         │
│                  pkg/tokenizer/coccoc.go               │
└───────────────────────────┬────────────────────────────┘
                            │ Gọi hàm C (cgo)
                            ▼
┌────────────────────────────────────────────────────────┐
│                   C WRAPPER BRIDGE                     │
│               extern "C" { c_tokenize() }              │
└───────────────────────────┬────────────────────────────┘
                            │ Gọi C++ method
                            ▼
┌────────────────────────────────────────────────────────┐
│              C++ CỐC CỐC TOKENIZER CORE                │
│             Tokenizer::segment() + sys.dic             │
└────────────────────────────────────────────────────────┘
```

---

## 3. Lý Thuyết 3: Quản Lý Vùng Nhớ Giữa Go và C (Memory Boundary & GC)

Đây là phần **quan trọng nhất** mà mọi kỹ sư Go cần phải khắc cốt ghi tâm:

### 3.1. Hai thế giới bộ nhớ riêng biệt:
1. **Go Runtime Memory**: Được quản lý tự động bởi **Garbage Collector (GC)**. Bạn khai báo biến, struct, slice, Go tự dọn dẹp khi không còn dùng.
2. **C Memory (Heap)**: **KHÔNG CÓ GC!** Bộ nhớ được cấp phát bằng `malloc()` và BẮT BUỘC phải được giải phóng thủ công bằng `free()`.

### 3.2. Chuyển đổi String giữa Go và C:
* Go string có header gồm `con trỏ` + `độ dài` (không kết thúc bằng byte `\0`).
* C string là mảng ký tự kết thúc bằng byte null `\0` (null-terminated).
* Khi truyền chuỗi từ Go sang C:
  ```go
  // C.CString sao chép chuỗi từ Go sang vùng nhớ Heap của C (dùng malloc)
  cText := C.CString("học sinh")
  
  // NẾU KHÔNG CÓ DÒNG NÀY -> BỊ MEMORY LEAK (Rò rỉ RAM vĩnh viễn)!
  defer C.free(unsafe.Pointer(cText))
  ```

---

## 4. Hướng Dẫn Code Bước 1: Thiết Kế Lớp C Wrapper (`extern "C"`)

Ta tạo 2 file trung gian đặt trong thư mục `pkg/tokenizer/coccoc/`:
1. `coccoc_bridge.h`: Khai báo header chuẩn C.
2. `coccoc_bridge.cpp`: Cài đặt hàm C nhưng gọi mã C++ bên dưới.

### File 1: `pkg/tokenizer/coccoc/coccoc_bridge.h`
```c
#ifndef COCCOC_BRIDGE_H
#define COCCOC_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

// Con trỏ void* ẩn giấu con trỏ C++ Tokenizer instance
typedef void* TokenizerHandle;

// Khởi tạo Tokenizer với đường dẫn tới từ điển sys.dic
TokenizerHandle coccoc_init(const char* dict_path);

// Phân tách từ tiếng Việt. Trả về chuỗi các từ ghép nối bằng dấu gạch dưới "_"
// Ví dụ: "tôi đi học sinh" -> "tôi đi học_sinh"
char* coccoc_tokenize(TokenizerHandle handle, const char* text);

// Giải phóng chuỗi trả về từ coccoc_tokenize
void coccoc_free_string(char* s);

// Hủy Tokenizer instance khi tắt app
void coccoc_close(TokenizerHandle handle);

#ifdef __cplusplus
}
#endif

#endif // COCCOC_BRIDGE_H
```

---

## 5. Hướng Dẫn Code Bước 2: Xây Dựng Package Go CGO (`pkg/tokenizer/coccoc.go`)

Trong file Go, ta dùng comment `#cgo` để chỉ dẫn trình biên dịch Go biết nơi tìm header và thư viện:

```go
package tokenizer

/*
#cgo CFLAGS: -I${SRCDIR}/coccoc
#cgo CXXFLAGS: -I${SRCDIR}/coccoc -std=c++11
#cgo LDFLAGS: -L${SRCDIR}/coccoc -lcoccoc_tokenizer
#include "coccoc_bridge.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"strings"
	"sync"
	"unsafe"
)

// CoccocTokenizer bọc con trỏ C và quản lý đồng bộ
type CoccocTokenizer struct {
	handle C.TokenizerHandle
	mu     sync.Mutex
}

// NewCoccocTokenizer khởi tạo kết nối tới C++ engine
func NewCoccocTokenizer(dictPath string) (*CoccocTokenizer, error) {
	cDictPath := C.CString(dictPath)
	defer C.free(unsafe.Pointer(cDictPath)) // Bắt buộc giải phóng!

	handle := C.coccoc_init(cDictPath)
	if handle == nil {
		return nil, errors.New("không thể khởi tạo Cốc Cốc Tokenizer, kiểm tra lại dict_path")
	}

	return &CoccocTokenizer{handle: handle}, nil
}

// Tokenize bóc tách câu văn bản tiếng Việt thành danh sách các token
func (t *CoccocTokenizer) Tokenize(text string) []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText)) // Giải phóng bộ nhớ C

	// Gọi hàm C qua cầu nối
	cResult := C.coccoc_tokenize(t.handle, cText)
	if cResult == nil {
		return nil
	}
	defer C.coccoc_free_string(cResult) // Giải phóng chuỗi kết quả C

	// Chuyển đổi từ C String sang Go String
	goResult := C.GoString(cResult)

	// Tách thành mảng các từ
	return strings.Fields(goResult)
}

// Close giải phóng tài nguyên C++
func (t *CoccocTokenizer) Close() {
	if t.handle != nil {
		C.coccoc_close(t.handle)
		t.handle = nil
	}
}
```

---

## 6. Hướng Dẫn Code Bước 3: Quản Lý Biên Dịch & Chạy Thử Trên Docker

Vì thư viện Cốc Cốc viết bằng C++ và yêu cầu `cmake`, `g++`, cách triển khai chuẩn mực và không bao giờ lỗi font hay xung đột môi trường Windows là đóng gói trong **Docker**:

```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder

# 1. Cài đặt g++, cmake, make, git
RUN apk add --no-cache gcc g++ cmake make git

# 2. Clone mã nguồn Cốc Cốc tokenizer và biên dịch
WORKDIR /build
RUN git clone https://github.com/coccoc/coccoc-tokenizer.git && \
    cd coccoc-tokenizer && \
    mkdir build && cd build && \
    cmake .. && \
    make -j4 install

# 3. Build ứng dụng Go có CGO_ENABLED=1
WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/indexer cmd/indexer/main.go
```

---

### 💡 Tóm tắt 3 điều cốt lõi cần nhớ:
1. **CGO không gọi trực tiếp được C++**: Luôn cần lớp vỏ bọc `extern "C"`.
2. **C.CString phải đi kèm C.free**: Nếu quên `defer C.free(...)`, server sẽ bị rò rỉ RAM (Memory Leak) dần dần cho đến khi crash.
3. **Từ điển `sys.dic` là linh hồn của Tokenizer**: Nếu không tìm thấy file này, thư viện C++ sẽ không thể khởi tạo.
