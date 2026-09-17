// W3Schools Go Tutorial: Go Data Types (Kiểu Dữ Liệu)
// Link tham khảo:
// - https://www.w3schools.com/go/go_data_types.php
// - https://www.w3schools.com/go/go_boolean_data_type.php
// - https://www.w3schools.com/go/go_integer_data_type.php
// - https://www.w3schools.com/go/go_float_data_type.php
// - https://www.w3schools.com/go/go_string_data_type.php

package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 06: GO DATA TYPES (KIỂU DỮ LIỆU)")
	fmt.Println("==================================================")

	// --- 1. KIỂU BOOLEAN (ĐÚNG / SAI) ---
	fmt.Println("\n--- 1. Boolean Data Type (bool) ---")
	var isGoFun bool = true
	var isFishTasty bool = false
	fmt.Println("Học Go có thú vị không?:", isGoFun)
	fmt.Println("Cá có ngon không?:", isFishTasty)

	// --- 2. KIỂU SỐ NGUYÊN CÓ DẤU (SIGNED INTEGERS) ---
	fmt.Println("\n--- 2. Signed Integers (Số nguyên có dấu: âm & dương) ---")
	// int: Phụ thuộc vào kiến trúc CPU (32-bit trên hệ thống 32-bit, 64-bit trên 64-bit)
	var normalInt int = -42
	var a8 int8 = 127             // Từ -128 đến 127 (1 byte)
	var a16 int16 = -32768        // Từ -32768 đến 32767 (2 bytes)
	var a32 int32 = 2147483647    // 4 bytes (rune là bí danh của int32)
	var a64 int64 = math.MaxInt64 // 8 bytes

	fmt.Printf("int: %d, int8: %d, int16: %d, int32: %d, int64: %d\n",
		normalInt, a8, a16, a32, a64)

	// --- 3. KIỂU SỐ NGUYÊN KHÔNG DẤU (UNSIGNED INTEGERS) ---
	fmt.Println("\n--- 3. Unsigned Integers (Số nguyên không dấu >= 0) ---")
	// Không thể chứa số âm, chỉ từ 0 trở lên
	var u8 uint8 = 255          // Từ 0 đến 255 (byte là bí danh của uint8)
	var u16 uint16 = 65535      // Từ 0 đến 65535
	var u32 uint32 = 4294967295 // Từ 0 đến 4,294,967,295
	var u64 uint64 = math.MaxUint64

	fmt.Printf("uint8 (byte): %d, uint16: %d, uint32: %d, uint64: %d\n",
		u8, u16, u32, u64)

	// --- 4. KIỂU SỐ THỰC (FLOATING POINT NUMBERS) ---
	fmt.Println("\n--- 4. Float Data Types (Số thực float32 & float64) ---")
	// float32: Độ chính xác đơn (~6-7 chữ số có nghĩa)
	// float64: Độ chính xác kép (~15-17 chữ số có nghĩa) - Khuyên dùng làm chuẩn
	var f1 float32 = 3.14159265
	var f2 float64 = 3.141592653589793
	fmt.Printf("float32: %f (Chính xác tới: %.7f)\n", f1, f1)
	fmt.Printf("float64: %f (Chính xác tới: %.15f)\n", f2, f2)

	// --- 5. KIỂU CHUỖI KÝ TỰ (STRING) ---
	fmt.Println("\n--- 5. String Data Type ---")
	// String trong Go là bất biến (immutable) và mã hóa UTF-8 theo mặc định.
	var greeting string = "Xin chào Việt Nam! 🇻🇳"
	fmt.Println("Lời chào:", greeting)
	fmt.Println("Độ dài byte của chuỗi (len):", len(greeting))

	// --- 6. ÉP KIỂU DỮ LIỆU RÕ RÀNG (EXPLICIT TYPE CASTING) ---
	fmt.Println("\n--- 6. Ép kiểu dữ liệu (Type Conversion) ---")
	// QUAN TRỌNG: Go KHÔNG BAO GIỜ tự động ép kiểu ngầm định (No implicit casting).
	// Muốn cộng số int với float, bắt buộc phải ép kiểu rõ ràng:
	var count int = 10
	var price float64 = 19.5

	// var total float64 = count * price // LỖI: mismatched types int and float64
	var total float64 = float64(count) * price
	fmt.Printf("Tổng tiền sau ép kiểu: %.2f\n", total)
}
