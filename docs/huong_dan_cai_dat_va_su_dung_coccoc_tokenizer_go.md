# HƯỚNG DẪN CÀI ĐẶT & SỬ DỤNG COCCOC-TOKENIZER-GO
## THƯ VIỆN TÁCH TỪ TIẾNG VIỆT ĐỘC LẬP CHO NGÔN NGỮ GO

Tài liệu này hướng dẫn chi tiết từ A-Z cách cài đặt và tích hợp thư viện tách từ tiếng Việt Cốc Cốc (`coccoc-tokenizer-go`) vào bất kỳ dự án Go nào của bạn.

---

## 1. CÀI ĐẶT THƯ VIỆN

Mở terminal trong thư mục dự án Go của bạn và chạy lệnh sau:

```bash
go get -u github.com/zangzang1303/coccoc-tokenizer-go
```

File `go.mod` của bạn sẽ tự động được bổ sung:
```go
require github.com/zangzang1303/coccoc-tokenizer-go v1.0.0
```

---

## 2. LỰA CHỌN CHẾ ĐỘ SỬ DỤNG

Thư viện hỗ trợ **2 chế độ chạy** cực kỳ linh hoạt:

| Tiêu Chí | Chế Độ 1: Pure-Go (Khuyên dùng khi lập trình) | Chế Độ 2: CGO Native (Khuyên dùng cho Server) |
| :--- | :--- | :--- |
| **Hệ điều hành** | Chạy tốt trên cả Windows, macOS và Linux | Tối ưu trên máy chủ Linux, container Docker |
| **Yêu cầu cài đặt** | **Không cần cài gì thêm** (Chỉ cần có Go) | Cần cài trình biên dịch C++ (`gcc`, `g++`) |
| **Từ điển** | Đã nhúng sẵn danh sách 40k từ ghép trong file chạy | Cần tải bộ từ điển C++ nhị phân (`.dump` ~45MB) |
| **Tốc độ** | ~22.000 câu / giây (~45 µs/câu) | ~830.000 câu / giây (~1,2 µs/câu) |

---

## 3. CÁCH SỬ DỤNG NHANH (QUICK START)

### 3.1. Chế độ Mặc định (Tự Động Chạy Mượt Mà Mọi Nơi)
Bạn không cần cấu hình phức tạp, chỉ cần khởi tạo `coccoc.New()`:

```go
package main

import (
	"fmt"
	"log"

	"github.com/zangzang1303/coccoc-tokenizer-go"
)

func main() {
	// Khởi tạo bộ tách từ (tự động nhận diện môi trường)
	tok, err := coccoc.New()
	if err != nil {
		log.Fatalf("Khởi tạo thất bại: %v", err)
	}
	defer tok.Close()

	text := "Sinh viên trường đại học Bách Khoa đi uống cà phê sữa đá."

	// 1. Tách từ ghép và nối bằng dấu gạch dưới "_"
	segmented, _ := tok.Segment(text)
	fmt.Println("Kết quả Segment:")
	fmt.Println(segmented)
	// Xuất ra: Sinh_viên trường đại_học Bách_Khoa đi uống cà_phê sữa đá.

	// 2. Băm thành danh sách các từ để tìm kiếm
	tokens, _ := tok.Tokenize(text)
	fmt.Println("\nDanh sách Tokens:")
	fmt.Printf("%v\n", tokens)
	// Xuất ra: [sinh_viên trường đại_học bách_khoa đi uống cà_phê sữa đá]
}
```

---

## 4. CÁC TÍNH NĂNG NÂNG CAO

### 4.1. Thêm từ vựng tùy chỉnh hoặc tên thương hiệu riêng
Nếu dự án của bạn có các từ ngữ chuyên ngành riêng (ví dụ: tên công ty, tên sản phẩm mới), bạn có thể bổ sung dễ dàng bằng tùy chọn `WithCustomWords`:

```go
tok, err := coccoc.New(
	coccoc.WithCustomWords(
		"trí tuệ nhân tạo",
		"thương mại điện tử",
		"máy bay không người lái",
	),
)
```

### 4.2. Lấy vị trí từ vựng (Byte Offset) để làm tính năng Highlight
Khi làm tính năng bôi vàng từ khóa tìm kiếm (Search Highlighting), bạn cần biết chính xác từ đó bắt đầu và kết thúc ở ký tự thứ mấy trong văn bản:

```go
details, _ := tok.TokenizeFull("Học sinh uống trà sữa.")
for _, t := range details {
	fmt.Printf("Từ: %-10s | Vị trí: %d -> %d | Loại: %s\n", 
		t.Text, t.StartByte, t.EndByte, t.Type)
}
```

### 4.3. Các tùy chọn cấu hình linh hoạt khác (Options)
```go
tok, err := coccoc.New(
	coccoc.WithNormalizeLowercase(true), // Tự động chuyển chữ thường trong Tokenize (mặc định: true)
	coccoc.WithKeepPunctuation(false),   // Loại bỏ các dấu chấm câu thừa (mặc định: false)
	coccoc.WithForcePureGo(true),        // Bắt buộc dùng Go thuần kể cả khi môi trường có CGO
)
```

---

## 5. CẤU HÌNH CHẾ ĐỘ C++ CGO CHO MÁY CHỦ SẢN XUẤT (LINUX SERVER)

Khi đưa ứng dụng lên máy chủ Production hoặc triển khai Docker để xử lý dữ liệu lớn, bạn nên bật chế độ CGO Native để đạt tốc độ tối đa (~800k câu/s):

### Bước 1: Cài đặt công cụ C++ trên Linux
```bash
# Trên Ubuntu / Debian:
sudo apt-get update && sudo apt-get install -y build-essential
```

### Bước 2: Tải bộ từ điển Cốc Cốc C++ (.dump)
Bạn có thể tự động tải bộ từ điển nhị phân bằng công cụ `dictutil` đi kèm:

```go
import "github.com/zangzang1303/coccoc-tokenizer-go/dictutil"

// Tự động tải từ điển về thư mục cache chuẩn của hệ thống (~/.cache/coccoc-tokenizer/dicts)
dictPath, err := dictutil.EnsureDicts("")
if err != nil {
	log.Fatalf("Không thể tải từ điển: %v", err)
}

// Khởi tạo tokenizer với đường dẫn từ điển vừa tải
tok, err := coccoc.New(coccoc.WithDictPath(dictPath))
```

### Bước 3: File Dockerfile mẫu chuẩn Production
```dockerfile
FROM golang:1.22-bookworm AS builder

# Cài đặt trình biên dịch g++
RUN apt-get update && apt-get install -y g++ wget

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Biên dịch với cờ CGO_ENABLED=1
RUN CGO_ENABLED=1 GOOS=linux go build -o server main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y libstdc++6 ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/server .
# Copy bộ từ điển vào container nếu có
COPY dicts/ /app/dicts/

CMD ["./server"]
```

---

## 6. VÍ DỤ THỰC TẾ: TIỀN XỬ LÝ DỮ LIỆU ĐẨY VÀO OPENSEARCH / ELASTICSEARCH

Đây là mô hình giúp máy chủ tìm kiếm tăng tốc ghi dữ liệu gấp gần 3 lần:

```go
type ChatMessage struct {
	ID       string `json:"id"`
	RawText  string `json:"text"`      // Văn bản gốc
	TextVi   string `json:"text_vi"`   // Văn bản có dấu đã ghép từ "_"
	TextRaw  string `json:"text_raw"`  // Văn bản không dấu đã ghép từ "_"
}

func PreprocessMessage(tok coccoc.Tokenizer, id string, text string) ChatMessage {
	// 1. Tách từ ghép có dấu
	textVi, _ := tok.Segment(text)

	// 2. Tách từ ghép không dấu (dùng để tìm kiếm không dấu)
	textRaw := removeVietnameseAccents(textVi)

	return ChatMessage{
		ID:      id,
		RawText: text,
		TextVi:  textVi,
		TextRaw: textRaw,
	}
}
```

---

## 7. CÂU HỎI THƯỜNG GẶP (FAQ)

**Q: Máy tính của tôi là Windows hoặc MacBook chip M1/M2/M3 có dùng được không?**  
- **Trả lời:** Có, hoàn toàn được. Thư viện sẽ tự động chuyển sang chế độ Go thuần (Pure-Go) và chạy ngay lập tức mà không đòi hỏi cài đặt thêm bất kỳ phần mềm C++ nào.

**Q: Bộ từ điển tiếng Việt được cập nhật từ đâu?**  
- **Trả lời:** Dữ liệu từ điển được trích xuất và tối ưu chuẩn hóa từ bộ từ điển chính thức của Cốc Cốc NLP, đảm bảo nhận diện chính xác hầu hết các từ vựng tiếng Việt hiện đại.

**Q: Có bị rò rỉ RAM khi chạy lâu dài trên server không?**  
- **Trả lời:** Không. Mã nguồn đã trải qua bài kiểm tra khắt khe nạp 50.000 câu văn liên tục với mức tiêu thụ RAM ổn định 100%, không bị tràn bộ nhớ.
