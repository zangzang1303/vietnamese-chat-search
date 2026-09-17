// W3Schools Go Tutorial: Go Arrays (Mảng Cố Định)
// Link tham khảo: https://www.w3schools.com/go/go_arrays.php

package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 11: GO ARRAYS (MẢNG CỐ ĐỊNH KÍCH THƯỚC)")
	fmt.Println("==================================================")

	// --- ĐẶC ĐIỂM CỦA ARRAY TRONG GO ---
	// 1. Array có độ dài CỐ ĐỊNH, không thể phóng to hay thu nhỏ sau khi khai báo.
	// 2. Độ dài là một phần của kiểu dữ liệu (ví dụ [3]int và [5]int là 2 kiểu hoàn toàn KHÁC NHAU).
	// 3. Trong Go, Array là KIỂU TRUYỀN GIÁ TRỊ (Value Type), khi gán sang biến mới sẽ copy toàn bộ mảng.

	// --- 1. CÁC CÁCH KHAI BÁO MẢNG ---
	fmt.Println("\n--- 1. Các cách khai báo mảng ---")

	// Cách 1: Chỉ định rõ độ dài
	var scores = [3]int{85, 92, 78}

	// Cách 2: Khai báo với toán tử ngắn :=
	cities := [4]string{"Hà Nội", "Đà Nẵng", "TP. Hồ Chí Minh", "Cần Thơ"}

	// Cách 3: Để trình biên dịch tự đếm số lượng phần tử với dấu ba chấm [...]
	primes := [...]int{2, 3, 5, 7, 11, 13}

	fmt.Println("Điểm số [3]int:", scores)
	fmt.Println("Thành phố [4]string:", cities)
	fmt.Println("Số nguyên tố [...]int:", primes)
	fmt.Println("Độ dài mảng primes (len):", len(primes))

	// --- 2. MẢNG CHƯA KHỞI TẠO HOẶC KHỞI TẠO MỘT PHẦN ---
	fmt.Println("\n--- 2. Khởi tạo một phần & Giá trị mặc định ---")
	// Phần tử nào không điền sẽ nhận zero-value (số 0):
	arrPart := [5]int{10, 20} // [10 20 0 0 0]
	fmt.Println("Khởi tạo 2 phần tử đầu:", arrPart)

	// Khởi tạo theo CHỈ SỐ CỤ THỂ (Specific Index Initialization):
	// Khởi tạo phần tử tại index 1 là 50, index 4 là 100
	customArr := [5]int{1: 50, 4: 100}
	fmt.Println("Khởi tạo theo index cụ thể:", customArr)

	// --- 3. TRUY CẬP VÀ SỬA ĐỔI PHẦN TỬ MẢNG ---
	fmt.Println("\n--- 3. Truy cập và thay đổi giá trị ---")
	cars := [3]string{"Volvo", "BMW", "Ford"}
	fmt.Println("Trước khi đổi:", cars)

	// Truy cập phần tử đầu tiên (index 0)
	fmt.Println("Phần tử đầu tiên:", cars[0])

	// Thay đổi phần tử thứ 2 (index 1)
	cars[1] = "VinFast"
	fmt.Println("Sau khi đổi cars[1] thành VinFast:", cars)

	// --- 4. TÍNH CHẤT SAO CHÉP GIÁ TRỊ (VALUE SEMANTICS) ---
	fmt.Println("\n--- 4. Mảng trong Go là Value Type (Sao chép toàn bộ) ---")
	original := [3]int{1, 2, 3}
	copied := original // Tạo ra một bản sao ĐỘC LẬP

	copied[0] = 999
	fmt.Println("Mảng gốc (original):", original) // Vẫn giữ nguyên [1 2 3]
	fmt.Println("Mảng copy (copied)  :", copied)   // [999 2 3]
}
