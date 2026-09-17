// W3Schools Go Tutorial: Go Slices (Mảng Động) & Modify Slices
// Link tham khảo:
// - https://www.w3schools.com/go/go_slices.php
// - https://www.w3schools.com/go/go_slices_modify.php

package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 12: GO SLICES (MẢNG ĐỘNG - DYNAMIC ARRAYS)")
	fmt.Println("==================================================")

	// --- ĐẶC ĐIỂM CỦA SLICE TRONG GO ---
	// 1. Slice linh hoạt, có thể co giãn kích thước tự động (Dynamic size).
	// 2. Không chứa số phần tử trong ngoặc vuông [] khi khai báo.
	// 3. Slice thực chất là một "cửa sổ" (view) trỏ tới một mảng ngầm bên dưới (underlying array).
	// 4. Có 2 thông số: len() (chiều dài thực tế) và cap() (sức chứa tối đa trước khi cấp phát lại bộ nhớ).

	// --- 1. CÁC CÁCH TẠO SLICE ---
	fmt.Println("\n--- 1. Các cách tạo Slice ---")

	// Cách 1: Khai báo trực tiếp (Literal)
	s1 := []int{10, 20, 30}
	fmt.Printf("Slice s1: %v, len = %d, cap = %d\n", s1, len(s1), cap(s1))

	// Cách 2: Cắt từ một mảng có sẵn (Slicing an array: arr[start:end])
	// Lấy từ index 'start' đến trước 'end' (không lấy index end)
	arr := [6]int{100, 200, 300, 400, 500, 600}
	s2 := arr[1:4] // Lấy index 1, 2, 3 -> [200 300 400]
	fmt.Printf("Slice s2 cắt từ mảng: %v, len = %d, cap = %d\n", s2, len(s2), cap(s2))

	// Cách 3: Dùng hàm make(): make([]Type, len, cap)
	s3 := make([]string, 3, 5) // len = 3, cap = 5
	s3[0] = "A"
	s3[1] = "B"
	s3[2] = "C"
	fmt.Printf("Slice s3 tạo bằng make: %v, len = %d, cap = %d\n", s3, len(s3), cap(s3))

	// --- 2. THÊM PHẦN TỬ VỚI APPEND() ---
	fmt.Println("\n--- 2. Thêm phần tử với append() ---")
	numbers := []int{1, 2, 3}
	fmt.Println("Ban đầu:", numbers)

	// Thêm 1 hoặc nhiều phần tử:
	numbers = append(numbers, 4, 5)
	fmt.Println("Sau khi append(numbers, 4, 5):", numbers)

	// Thêm một slice khác vào slice hiện tại (dùng toán tử '...' giải nén phần tử):
	moreNumbers := []int{6, 7, 8}
	numbers = append(numbers, moreNumbers...)
	fmt.Println("Sau khi nối thêm slice khác:", numbers)

	// --- 3. ĐẶC ĐIỂM THAM CHIẾU & HÀM COPY() ---
	fmt.Println("\n--- 3. Phân biệt Slice Reference vs Copy độc lập ---")

	// 3.1 Vì Slice trỏ vào mảng gốc, sửa slice sẽ làm thay đổi mảng gốc:
	baseArray := [4]int{11, 22, 33, 44}
	subSlice := baseArray[0:2] // [11 22]
	subSlice[0] = 999          // Sửa slice
	fmt.Println("Mảng gốc sau khi sửa subSlice[0]:", baseArray) // [999 22 33 44] bị thay đổi!

	// 3.2 Dùng hàm copy() để tạo vùng nhớ ĐỘC LẬP (tránh rò rỉ bộ nhớ):
	// Cú pháp: copy(đích, nguồn) - slice đích phải có len đủ lớn!
	source := []int{1, 2, 3, 4, 5}
	destination := make([]int, len(source)) // Chuẩn bị slice đích có cùng độ dài

	copiedCount := copy(destination, source)
	destination[0] = 888 // Sửa slice đích

	fmt.Printf("Số phần tử đã copy: %d\n", copiedCount)
	fmt.Println("Slice nguồn (source)     :", source)      // Vẫn là [1 2 3 4 5]
	fmt.Println("Slice đích (destination) :", destination) // [888 2 3 4 5] hoàn toàn độc lập
}
