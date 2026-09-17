// W3Schools Go Tutorial: Go Functions (Hàm, Tham Số, Đa Giá Trị Trả Về & Đệ Quy)
// Link tham khảo:
// - https://www.w3schools.com/go/go_functions.php
// - https://www.w3schools.com/go/go_function_parameters.php
// - https://www.w3schools.com/go/go_function_returns.php
// - https://www.w3schools.com/go/go_function_recursion.php

package main

import (
	"errors"
	"fmt"
)

// 1. Hàm đơn giản không có tham số và không trả về giá trị
func sayHello() {
	fmt.Println("Xin chào từ hàm sayHello()!")
}

// 2. Hàm có tham số (Parameters)
// Khi hai tham số liên tiếp cùng kiểu, có thể viết rút gọn: (a, b int) thay vì (a int, b int)
func addNumbers(a, b int) int {
	return a + b
}

// 3. Hàm trả về NHIỀU GIÁ TRỊ (Multiple Return Values - Đặc sản cực mạnh của Go)
// Trong Go, lập trình viên thường dùng tính năng này để trả về (kết quả, lỗi - result, error)
func safeDivide(dividend, divisor float64) (float64, error) {
	if divisor == 0 {
		return 0, errors.New("lỗi: không thể chia cho số 0")
	}
	return dividend / divisor, nil
}

// 4. Hàm có NAMED RETURN VALUES (Đặt tên trước cho giá trị trả về)
// Giúp code tự sinh tài liệu và có thể dùng "naked return" (gọi return không cần tham số kèm theo)
func calculateRectangle(width, height int) (area int, perimeter int) {
	area = width * height
	perimeter = 2 * (width + height)
	return // Tự động trả về area và perimeter
}

// 5. Hàm ĐỆ QUY (Recursion - Hàm tự gọi lại chính nó)
// Bắt buộc phải có ĐIỀU KIỆN DỪNG (Base case) để tránh tràn bộ nhớ (Stack Overflow)
func factorial(n int) int {
	// Base case: 0! = 1 hoặc 1! = 1
	if n <= 1 {
		return 1
	}
	// Recursive call: n * (n - 1)!
	return n * factorial(n-1)
}

func countdown(count int) {
	if count <= 0 {
		fmt.Println("🚀 Phóng tàu vũ trụ! (Hết đếm ngược)")
		return
	}
	fmt.Printf("Đếm ngược: %d...\n", count)
	countdown(count - 1)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 15: GO FUNCTIONS (HÀM & ĐỆ QUY)")
	fmt.Println("==================================================")

	// --- 1. GỌI HÀM CƠ BẢN ---
	fmt.Println("\n--- 1. Gọi hàm cơ bản ---")
	sayHello()

	// --- 2. HÀM CÓ THAM SỐ VÀ TRẢ VỀ GIÁ TRỊ ---
	fmt.Println("\n--- 2. Tham số và trả về 1 giá trị ---")
	sum := addNumbers(25, 17)
	fmt.Printf("Tổng của 25 + 17 = %d\n", sum)

	// --- 3. ĐA GIÁ TRỊ TRẢ VỀ (MULTIPLE RETURN VALUES) ---
	fmt.Println("\n--- 3. Hàm trả về nhiều giá trị (Result & Error) ---")

	// Trường hợp 1: Phép chia hợp lệ
	res1, err1 := safeDivide(100, 4)
	if err1 != nil {
		fmt.Println("Lỗi xảy ra:", err1)
	} else {
		fmt.Printf("100 / 4 = %.2f\n", res1)
	}

	// Trường hợp 2: Phép chia cho 0
	res2, err2 := safeDivide(50, 0)
	if err2 != nil {
		fmt.Println("Bắt lỗi thành công:", err2)
	} else {
		fmt.Println("Kết quả:", res2)
	}

	// Bỏ qua giá trị không cần thiết bằng dấu gạch dưới (_):
	onlyResult, _ := safeDivide(30, 2)
	fmt.Println("Chỉ lấy kết quả (bỏ qua err):", onlyResult)

	// --- 4. NAMED RETURN VALUES ---
	fmt.Println("\n--- 4. Named Return Values ---")
	dienTich, chuVi := calculateRectangle(10, 5)
	fmt.Printf("Hình chữ nhật (10x5): Diện tích = %d, Chu vi = %d\n", dienTich, chuVi)

	// --- 5. ĐỆ QUY (RECURSION) ---
	fmt.Println("\n--- 5. Đệ quy (Recursion) ---")
	num := 5
	fmt.Printf("Giai thừa của %d! = %d (5 * 4 * 3 * 2 * 1)\n", num, factorial(num))

	fmt.Println("\nĐếm ngược phóng tàu:")
	countdown(3)
}
