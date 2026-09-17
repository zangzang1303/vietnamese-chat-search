// W3Schools Go Tutorial: Go Constants (Hằng số)
// Link tham khảo: https://www.w3schools.com/go/go_constants.php

package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 04: GO CONSTANTS (HẰNG SỐ)")
	fmt.Println("==================================================")

	// --- 1. ĐẶC ĐIỂM CỦA HẰNG SỐ (CONST) ---
	// - Hằng số được khai báo với từ khóa "const".
	// - Giá trị của hằng số là BẤT BIẾN (không thể gán lại hay thay đổi sau khi tạo).
	// - Hằng số PHẢI được gán giá trị ngay khi khai báo.
	// - KHÔNG thể sử dụng toán tử ngắn ":=" để khai báo const.

	// 1.1 Typed Constants (Hằng số có định kiểu rõ ràng):
	const MAX_USERS int = 100
	const APP_VERSION string = "1.0.0"

	// 1.2 Untyped Constants (Hằng số không chỉ định kiểu cụ thể):
	// Go sẽ suy luận kiểu dựa trên ngữ cảnh sử dụng
	const PI = 3.14159265359

	fmt.Println("Max Users:", MAX_USERS)
	fmt.Println("App Version:", APP_VERSION)
	fmt.Println("Số PI:", PI)

	// Dòng dưới nếu bỏ comment sẽ báo lỗi biên dịch:
	// MAX_USERS = 200 // Cannot assign to MAX_USERS (declared const)

	// --- 2. KHAI BÁO NHIỀU HẰNG SỐ TRONG BLOCK `const (...)` ---
	fmt.Println("\n--- 2. Khai báo nhiều hằng số cùng lúc ---")
	const (
		STATUS_OK       = 200
		STATUS_NOTFOUND = 404
		STATUS_ERROR    = 500
	)
	fmt.Println("Mã HTTP:", STATUS_OK, STATUS_NOTFOUND, STATUS_ERROR)

	// --- 3. ĐẶC BIỆT TRONG GO: IOTA (ENUM AUTO-INCREMENT) ---
	fmt.Println("\n--- 3. Mẹo nâng cao: Hằng số tự tăng với 'iota' ---")
	const (
		MONDAY    = iota // 0
		TUESDAY          // 1
		WEDNESDAY        // 2
		THURSDAY         // 3
		FRIDAY           // 4
	)
	fmt.Printf("Thứ 2: %d, Thứ 3: %d, Thứ 4: %d, Thứ 5: %d, Thứ 6: %d\n",
		MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY)
}
