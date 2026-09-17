// W3Schools Go Tutorial: Go Conditions (Câu Lệnh Điều Kiện If, Else, Else If)
// Link tham khảo:
// - https://www.w3schools.com/go/go_conditions.php
// - https://www.w3schools.com/go/go_if_statement.php
// - https://www.w3schools.com/go/go_else_statement.php
// - https://www.w3schools.com/go/go_elseif_statement.php
// - https://www.w3schools.com/go/go_nested_if.php

package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 08: GO CONDITIONS (ĐIỀU KIỆN IF - ELSE)")
	fmt.Println("==================================================")

	// --- NGUYÊN TẮC CÚ PHÁP BẮT BUỘC TRONG GO ---
	// 1. Biểu thức điều kiện KHÔNG cần dấu ngoặc tròn ( ).
	// 2. Khối lệnh BẮT BUỘC phải nằm trong cặp ngoặc nhọn { }.
	// 3. Dấu mở ngoặc { PHẢI nằm cùng dòng với từ khóa "if" hoặc "else".
	// 4. Từ khóa "else" PHẢI nằm cùng dòng với dấu đóng ngoặc nhọn }: "} else {"
	//    Nếu viết xuống dòng:
	//    }
	//    else {  --> GO SẼ BÁO LỖI BIÊN DỊCH NGAY LẬP TỨC!

	// --- 1. LỆNH IF ĐƠN GIẢN ---
	fmt.Println("\n--- 1. Câu lệnh if đơn giản ---")
	age := 18
	if age >= 18 {
		fmt.Println("Bạn đã đủ tuổi công dân!")
	}

	// --- 2. LỆNH IF...ELSE ---
	fmt.Println("\n--- 2. Cấu trúc if...else ---")
	temperature := 15
	if temperature > 20 {
		fmt.Println("Thời tiết ấm áp.")
	} else {
		fmt.Println("Thời tiết se lạnh, hãy mặc thêm áo ấm!")
	}

	// --- 3. LỆNH ELSE IF (CHUỖI ĐIỀU KIỆN) ---
	fmt.Println("\n--- 3. Chuỗi điều kiện if - else if - else ---")
	score := 82

	if score >= 90 {
		fmt.Println("Xếp loại: Xuất sắc (A)")
	} else if score >= 80 {
		fmt.Println("Xếp loại: Giỏi (B)")
	} else if score >= 65 {
		fmt.Println("Xếp loại: Khá (C)")
	} else if score >= 50 {
		fmt.Println("Xếp loại: Trung bình (D)")
	} else {
		fmt.Println("Xếp loại: Cần cố gắng thêm (F)")
	}

	// --- 4. NESTED IF (ĐIỀU KIỆN LỒNG NHAU) ---
	fmt.Println("\n--- 4. Nested if (If lồng nhau) ---")
	num := 24
	if num > 0 {
		fmt.Println("Số này là số dương.")
		if num%2 == 0 {
			fmt.Println("Và đây là số chẵn.")
		} else {
			fmt.Println("Và đây là số lẻ.")
		}
	}

	// --- 5. ĐẶC TRƯNG IDIOMATIC TRONG GO: IF VỚI SHORT STATEMENT ---
	fmt.Println("\n--- 5. Khai báo ngắn gọn trong if (Short Statement) ---")
	// Go cho phép khai báo và khởi tạo biến ngay trước điều kiện:
	// Cú pháp: if <khởi tạo biến>; <điều kiện> { ... }
	// Biến 'length' chỉ tồn tại và có phạm vi sử dụng bên trong khối if...else này!
	message := "Xin chào Golang"
	if length := len(message); length > 10 {
		fmt.Printf("Chuỗi dài (%d ký tự)\n", length)
	} else {
		fmt.Printf("Chuỗi ngắn (%d ký tự)\n", length)
	}
	// Ngoài khối if, biến 'length' sẽ không còn tồn tại -> Giúp tránh ô nhiễm biến toàn cục.
}
