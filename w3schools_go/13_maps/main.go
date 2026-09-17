// W3Schools Go Tutorial: Go Maps (Cặp Khóa - Giá Trị Key-Value)
// Link tham khảo: https://www.w3schools.com/go/go_maps.php

package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 13: GO MAPS (BẢNG BĂM KEY - VALUE)")
	fmt.Println("==================================================")

	// --- ĐẶC ĐIỂM CỦA MAP TRONG GO ---
	// 1. Map lưu trữ dữ liệu dưới dạng các cặp Khóa (Key) và Giá trị (Value).
	// 2. Các Key là DUY NHẤT (không thể trùng lặp).
	// 3. Thứ tự các phần tử trong Map KHÔNG cố định (Unordered).
	// 4. Map là KIỂU THAM CHIẾU (Reference Type) - gán map sẽ trỏ cùng vùng nhớ.

	// --- 1. CÁC CÁCH KHỞI TẠO MAP ---
	fmt.Println("\n--- 1. Các cách khởi tạo Map ---")

	// Cách 1: Khởi tạo với Map Literal
	capitals := map[string]string{
		"Vietnam": "Hà Nội",
		"Japan":   "Tokyo",
		"USA":     "Washington, D.C.",
		"France":  "Paris",
	}
	fmt.Println("Thủ đô các nước:", capitals)

	// Cách 2: Dùng hàm make(map[KeyType]ValueType)
	userScores := make(map[string]int)
	userScores["Alice"] = 95
	userScores["Bob"] = 80
	fmt.Println("Điểm người dùng:", userScores)

	// LƯU Ý QUAN TRỌNG VỀ NIL MAP:
	// var nilMap map[string]int // Khai báo thế này là nil map!
	// nilMap["key"] = 10       // SẼ BỊ PANIC RUNTIME vì chưa khởi tạo bộ nhớ bằng make hoặc literal!

	// --- 2. THÊM, SỬA VÀ XÓA PHẦN TỬ TRONG MAP ---
	fmt.Println("\n--- 2. Thao tác Thêm, Sửa, Xóa (delete) ---")

	// Thêm phần tử mới
	capitals["Germany"] = "Berlin"
	fmt.Println("Sau khi thêm Germany:", capitals)

	// Sửa giá trị của key đã có
	userScores["Bob"] = 88
	fmt.Println("Sau khi sửa điểm Bob thành 88:", userScores)

	// Xóa một phần tử bằng hàm built-in delete(map, key):
	delete(capitals, "France")
	fmt.Println("Sau khi xóa France:", capitals)

	// --- 3. KIỂM TRA KEY CÓ TỒN TẠI HAY KHÔNG (COMMA-OK IDIOM) ---
	fmt.Println("\n--- 3. Kiểm tra sự tồn tại của Key ('val, ok' idiom) ---")
	// Trong Go, nếu truy vấn một key không tồn tại, Go sẽ trả về zero-value chứ không báo lỗi.
	// Để phân biệt giữa "giá trị là 0" và "key không hề tồn tại", ta dùng:
	// val, ok := myMap[key] (ok = true nếu tồn tại, ok = false nếu không có)

	capital, exists := capitals["Japan"]
	if exists {
		fmt.Println("Thủ đô Nhật Bản là:", capital)
	} else {
		fmt.Println("Không tìm thấy thông tin thủ đô Nhật Bản.")
	}

	_, existsUK := capitals["UK"]
	fmt.Println("Key 'UK' có tồn tại trong map không?:", existsUK) // false

	// --- 4. DUYỆT QUA MAP VỚI RANGE ---
	fmt.Println("\n--- 4. Duyệt qua các phần tử của Map với for range ---")
	for country, capCity := range capitals {
		fmt.Printf("  Quốc gia: %-10s -> Thủ đô: %s\n", country, capCity)
	}

	// --- 5. TÍNH CHẤT THAM CHIẾU CỦA MAP ---
	fmt.Println("\n--- 5. Map là Reference Type ---")
	mapA := map[string]int{"itemA": 100}
	mapB := mapA // mapB trỏ cùng vùng nhớ với mapA

	mapB["itemA"] = 999
	fmt.Println("mapA sau khi sửa trên mapB:", mapA["itemA"]) // 999
}
