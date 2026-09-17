// W3Schools Go Tutorial: Go Comments
// Link tham khảo: https://www.w3schools.com/go/go_comments.php

package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 02: GO COMMENTS (GHI CHÚ TRONG CODE)")
	fmt.Println("==================================================")

	// 1. Single-line comment (Ghi chú một dòng)
	// Bắt đầu bằng 2 dấu gạch chéo //
	// Bất kỳ văn bản nào giữa // và cuối dòng đều được trình biên dịch Go bỏ qua.

	fmt.Println("Ví dụ 1: Ghi chú một dòng (Single-line comments)") // Đây là ghi chú ở cuối dòng lệnh

	// 2. Multi-line comment (Ghi chú nhiều dòng)
	/* 
	   Bắt đầu bằng /* và kết thúc bằng dấu sao gạch chéo.
	   Mọi nội dung nằm giữa cặp dấu này sẽ bị bỏ qua khi biên dịch.
	   Thích hợp để viết tài liệu giải thích phức tạp hoặc tạm thời vô hiệu hóa khối code.
	*/
	fmt.Println("Ví dụ 2: Ghi chú nhiều dòng (Multi-line comments)")

	// 3. Vô hiệu hóa code bằng ghi chú:
	// fmt.Println("Dòng này đã bị comment lại nên sẽ KHÔNG chạy!")
	fmt.Println("Ví dụ 3: Đã vô hiệu hóa thành công một câu lệnh bằng comment.")
}
