package coccoc

/*
#cgo CXXFLAGS: -std=c++11 -Wno-error -Wno-cast-user-defined -I${SRCDIR}/../../../temp_coccoc -I${SRCDIR}/../../../temp_coccoc/tokenizer -I${SRCDIR}/../../../temp_coccoc/build/auto
#cgo LDFLAGS: -lstdc++
#include "coccoc_bridge.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode"
	"unsafe"
)

// ==============================================================================
// KIẾN THỨC CƠ BẢN VỀ CGO CHO NGƯỜI MỚI HỌC GO:
//
// 1. CGO là gì?
//    - CGO cho phép mã Go gọi trực tiếp các hàm được viết bằng C (hoặc C++ qua C bridge).
//    - Đoạn comment ngay phía trên `import "C"` (gọi là preamble) chứa các cờ biên dịch
//      `#cgo CXXFLAGS` và các lệnh `#include` thư viện C.
//
// 2. Quản lý bộ nhớ qua biên giới Go <-> C (Memory Boundary):
//    - Vùng nhớ do Go quản lý có Garbage Collector (GC) tự động dọn dẹp.
//    - Vùng nhớ do C quản lý (`malloc`) KHÔNG được GC của Go quản lý!
//    - Do đó, mọi chuỗi cấp phát bằng `C.CString(str)` bắt buộc phải được giải phóng
//      bằng `C.free(unsafe.Pointer(cStr))` thông qua từ khóa `defer` ngay sau khi tạo.
//    - Tương tự, chuỗi C trả về từ hàm C++ (`coccoc_tokenize_original`) phải được giải phóng
//      bằng `C.coccoc_free_string(resPtr)` sau khi sao chép sang chuỗi Go (`C.GoString`).
//
// 3. Thread-safety (An toàn đa luồng):
//    - Trong hệ thống web/chat, nhiều request (goroutine) có thể gọi tách từ cùng lúc.
//    - Ta sử dụng `sync.Mutex` để đảm bảo an toàn truy cập vào instance Tokenizer.
// ==============================================================================

// Tokenizer đóng gói Cốc Cốc Tokenizer thành một struct Go tiện lợi
type Tokenizer struct {
	mu       sync.Mutex
	dictPath string
}

var (
	// singletonInstance đảm bảo Tokenizer chỉ khởi tạo từ điển một lần duy nhất
	singletonInstance *Tokenizer
	once              sync.Once
	initErr           error
)

// New khởi tạo Cốc Cốc Tokenizer từ thư mục từ điển.
// loadNontone: false (khuyên dùng cho chat - siêu nhanh ~150ms, tốn ~45MB RAM).
//              true (nạp thêm bảng 2-gram 199MB hỗ trợ tách từ dính liền URL).
func New(dictPath string, loadNontone bool) (*Tokenizer, error) {
	// Kiểm tra sự tồn tại của thư mục từ điển
	if _, err := os.Stat(dictPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("thư mục từ điển không tồn tại: %s", dictPath)
	}

	once.Do(func() {
		// Chuyển chuỗi Go sang chuỗi C (char*)
		cDictPath := C.CString(dictPath)
		// Bắt buộc giải phóng vùng nhớ C sau khi hàm kết thúc
		defer C.free(unsafe.Pointer(cDictPath))

		nontoneFlag := C.int(0)
		if loadNontone {
			nontoneFlag = C.int(1)
		}

		// Gọi hàm C bridge để nạp từ điển vào C++ Engine
		res := C.coccoc_init(cDictPath, nontoneFlag)
		if res < 0 {
			initErr = fmt.Errorf("khởi tạo Cốc Cốc Tokenizer thất bại, mã lỗi: %d", int(res))
			return
		}

		singletonInstance = &Tokenizer{
			dictPath: dictPath,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return singletonInstance, nil
}

// SegmentOriginal phân tách văn bản và nối các từ ghép bằng dấu gạch dưới "_"
// Ví dụ: "Học sinh đi học uống cà phê" -> "Học_sinh đi học uống cà_phê"
func (t *Tokenizer) SegmentOriginal(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 1. Chuyển đổi Go string sang C string
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	// 2. Gọi C Bridge
	cResult := C.coccoc_tokenize_original(cText)
	if cResult == nil {
		return "", errors.New("lỗi khi phân tách từ bằng Cốc Cốc Tokenizer")
	}
	// 3. Giải phóng con trỏ chuỗi C sau khi đã đọc xong dữ liệu sang Go
	defer C.coccoc_free_string(cResult)

	// 4. Sao chép C string sang Go string
	goResult := C.GoString(cResult)
	return goResult, nil
}

// Tokenize nhận vào một câu văn bản và trả về danh sách các tokens đã được:
// 1. Phân tách từ ghép bằng Cốc Cốc (ví dụ: "học_sinh", "cà_phê")
// 2. Chuyển thành chữ thường (lowercase)
// 3. Tách bỏ dấu câu, chỉ giữ lại từ vựng có ý nghĩa
func (t *Tokenizer) Tokenize(text string) ([]string, error) {
	segmented, err := t.SegmentOriginal(text)
	if err != nil {
		return nil, err
	}

	if segmented == "" {
		return []string{}, nil
	}

	// Chuyển toàn bộ chuỗi sang chữ thường
	lower := strings.ToLower(segmented)

	// Tách từ theo khoảng trắng và dấu câu
	// Giữ lại: chữ cái (letter), chữ số (number) và dấu gạch dưới "_" (từ ghép)
	tokens := strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})

	return tokens, nil
}
