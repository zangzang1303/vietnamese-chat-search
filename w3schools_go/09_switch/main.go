// W3Schools Go Tutorial: Go Switch Statements (Cấu Trúc Switch)
// Link tham khảo:
// - https://www.w3schools.com/go/go_switch.php
// - https://www.w3schools.com/go/go_switch_multi.php

package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 09: GO SWITCH STATEMENTS (RẼ NHÁNH SWITCH)")
	fmt.Println("==================================================")

	// --- ĐẶC ĐIỂM NỔI BẬT CỦA SWITCH TRONG GO ---
	// 1. Không cần từ khóa "break" ở mỗi case (Go tự động break khi hết case).
	// 2. Không sợ bị "fall-through" nhầm lẫn như trong C/C++/Java.
	// 3. Cho phép gộp nhiều giá trị vào một case bằng dấu phẩy (,).
	// 4. Có thể dùng switch không cần biểu thức (switch true) thay cho chuỗi if-else dài.

	// --- 1. SINGLE-CASE SWITCH (SWITCH ĐƠN GIẢN) ---
	fmt.Println("\n--- 1. Single-case switch cơ bản ---")
	dayIndex := 3

	switch dayIndex {
	case 1:
		fmt.Println("Thứ Hai: Đầu tuần tràn đầy năng lượng!")
	case 2:
		fmt.Println("Thứ Ba: Đang tập trung làm việc.")
	case 3:
		fmt.Println("Thứ Tư: Giữa tuần rồi.")
	case 4:
		fmt.Println("Thứ Năm: Sắp đến cuối tuần.")
	case 5:
		fmt.Println("Thứ Sáu: Ngày làm việc cuối tuần!")
	default:
		fmt.Println("Cuối tuần hoặc ngày không hợp lệ.")
	}

	// --- 2. MULTI-CASE SWITCH (GỘP NHIỀU TRƯỜNG HỢP) ---
	fmt.Println("\n--- 2. Multi-case switch (Gộp nhiều trường hợp bằng dấu phẩy) ---")
	dayName := "Saturday"

	switch dayName {
	case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday":
		fmt.Println(dayName, "là Ngày trong tuần (Weekday - Đi làm/Đi học).")
	case "Saturday", "Sunday":
		fmt.Println(dayName, "là Ngày cuối tuần (Weekend - Nghỉ ngơi/Đi chơi).")
	default:
		fmt.Println("Tên ngày không hợp lệ!")
	}

	// --- 3. SWITCH KHÔNG CẦN BIỂU THỨC (SWITCH TRUE) ---
	fmt.Println("\n--- 3. Switch không điều kiện (Thay thế chuỗi if-else if sạch sẽ hơn) ---")
	currentHour := time.Now().Hour()

	switch {
	case currentHour < 12:
		fmt.Printf("Bây giờ là %dh: Chào buổi sáng!\n", currentHour)
	case currentHour < 18:
		fmt.Printf("Bây giờ là %dh: Chào buổi chiều!\n", currentHour)
	default:
		fmt.Printf("Bây giờ là %dh: Chào buổi tối!\n", currentHour)
	}

	// --- 4. TỪ KHÓA NÂNG CAO: FALLTHROUGH ---
	fmt.Println("\n--- 4. Từ khóa 'fallthrough' (Chủ động đi tiếp xuống case tiếp theo) ---")
	level := 1
	switch level {
	case 1:
		fmt.Println("Cấp 1: Bạn nhận được Quyền Đọc (Read).")
		fallthrough // Cố ý thực thi tiếp case 2 mà không cần xét điều kiện
	case 2:
		fmt.Println("Cấp 2: Bạn nhận được Quyền Ghi (Write).")
	case 3:
		fmt.Println("Cấp 3: Quyền Quản Trị (Admin).")
	}
}
