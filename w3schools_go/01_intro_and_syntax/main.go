// W3Schools Go Tutorial: Go Introduction, Getting Started & Syntax
// Link tham khảo: https://www.w3schools.com/go/go_syntax.php

// 1. "package main": Khai báo rằng file này thuộc package "main".
// Trong Go, mọi chương trình muốn tạo ra file thực thi (.exe hoặc binary)
// bắt buộc phải có package main và hàm func main().
package main

// 2. "import": Dùng để nạp các package thư viện cần thiết.
// "fmt" (Format package) là thư viện chuẩn dùng cho việc in ấn, định dạng dữ liệu ra console.
import (
	"fmt"
)

// 3. "func main()": Điểm bắt đầu (entry point) khi chương trình khởi chạy.
// Mọi code bên trong cặp ngoặc { } của func main() sẽ được thực thi tuần tự từ trên xuống dưới.
func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 01: GO INTRODUCTION & SYNTAX (W3SCHOOLS)")
	fmt.Println("==================================================")

	// 4. In thông điệp kinh điển
	fmt.Println("Hello, World! Chào mừng bạn đến với ngôn ngữ Go (Golang)!")

	// 5. Các đặc điểm cú pháp quan trọng trong Go:
	// - Go KHÔNG cần dấu chấm phẩy (;) ở cuối mỗi câu lệnh (Go tự động thêm vào khi biên dịch).
	// - Go phân biệt chữ hoa và chữ thường (Case-sensitive): ví dụ "myVar" khác "myvar".
	// - Cặp ngoặc nhọn mở '{' của hàm hay khối lệnh PHẢI nằm cùng dòng với tên hàm/lệnh (không được xuống dòng mới).

	fmt.Println("\nCác nguyên tắc cú pháp cốt lõi:")
	fmt.Println("1. package main: Định nghĩa gói thực thi chính.")
	fmt.Println("2. import \"fmt\": Thư viện in ấn chuẩn.")
	fmt.Println("3. func main(): Điểm khởi đầu của chương trình.")
	fmt.Println("4. Không cần chấm phẩy (;) kết thúc dòng.")
	fmt.Println("5. Dấu ngoặc mở '{' phải nằm cùng dòng với câu lệnh khai báo.")
}
