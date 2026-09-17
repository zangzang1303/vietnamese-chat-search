// W3Schools Go Tutorial: Go Loops (Vòng Lặp For)
// Link tham khảo: https://www.w3schools.com/go/go_loops.php

package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 10: GO LOOPS (VÒNG LẶP FOR)")
	fmt.Println("==================================================")

	// --- ĐẶC ĐIỂM ĐỘC ĐÁO CỦA GO: CHỈ CÓ DUY NHẤT VÒNG LẶP "FOR" ---
	// Go không có từ khóa 'while' hay 'do-while'. Mọi dạng lặp đều được thực hiện qua 'for'.

	// --- 1. VÒNG LẶP FOR TIÊU CHUẨN (3 THÀNH PHẦN: INIT; CONDITION; POST) ---
	fmt.Println("\n--- 1. Vòng lặp for tiêu chuẩn ---")
	for i := 1; i <= 5; i++ {
		fmt.Printf("Lần lặp thứ: %d\n", i)
	}

	// --- 2. VÒNG LẶP FOR ĐÓNG VAI TRÒ NHƯ WHILE ---
	fmt.Println("\n--- 2. Dùng for thay thế cho 'while' ---")
	counter := 1
	for counter <= 3 {
		fmt.Printf("Giá trị counter: %d\n", counter)
		counter++
	}

	// --- 3. LỆNH CONTINUE VÀ BREAK ---
	fmt.Println("\n--- 3. Lệnh continue (bỏ qua lần lặp) và break (thoát vòng lặp) ---")
	for n := 1; n <= 10; n++ {
		if n%2 == 0 {
			// Bỏ qua các số chẵn, nhảy sang lần lặp tiếp theo
			continue
		}
		if n > 7 {
			// Dừng hẳn vòng lặp khi n vượt quá 7
			fmt.Println("Gặp số > 7, gọi break để thoát khỏi vòng lặp!")
			break
		}
		fmt.Println("Số lẻ:", n)
	}

	// --- 4. VÒNG LẶP LỒNG NHAU (NESTED LOOPS) ---
	fmt.Println("\n--- 4. Vòng lặp lồng nhau (Nested Loops) ---")
	adj := [2]string{"Nhanh", "Mạnh"}
	fruits := [2]string{"Táo", "Cam"}

	for i := 0; i < len(adj); i++ {
		for j := 0; j < len(fruits); j++ {
			fmt.Printf("%s %s\n", adj[i], fruits[j])
		}
	}

	// --- 5. VÒNG LẶP RANGE (DUYỆT QUA TẬP HỢP DỮ LIỆU) ---
	fmt.Println("\n--- 5. Vòng lặp 'range' (Rất phổ biến trong Go) ---")
	languages := []string{"Go", "Python", "Rust", "TypeScript"}

	// 5.1 Lấy cả chỉ số (index) và giá trị (value):
	fmt.Println("Lấy cả index và value:")
	for index, value := range languages {
		fmt.Printf("  Index: %d, Ngôn ngữ: %s\n", index, value)
	}

	// 5.2 Bỏ qua index bằng dấu gạch dưới (_ - Blank Identifier):
	fmt.Println("Chỉ lấy value (bỏ qua index bằng '_'):")
	for _, value := range languages {
		fmt.Printf("  -> %s\n", value)
	}
}
