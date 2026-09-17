// W3Schools Go Tutorial: Go Output Functions & Formatting Verbs
// Link tham khảo:
// - https://www.w3schools.com/go/go_output.php
// - https://www.w3schools.com/go/go_formatting_verbs.php

package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 05: GO OUTPUT & FORMATTING VERBS")
	fmt.Println("==================================================")

	// --- PHẦN 1: BA HÀM XUẤT DỮ LIỆU CHÍNH ---
	// 1. fmt.Print(): In liên tục, KHÔNG tự động xuống dòng và KHÔNG tự cách khoảng trắng giữa các chuỗi (trừ khi đối số không phải là chuỗi).
	// 2. fmt.Println(): Tự động thêm khoảng trắng giữa các đối số và TỰ ĐỘNG XUỐNG DÒNG ở cuối.
	// 3. fmt.Printf(): In có định dạng dựa theo các "động từ" (formatting verbs).

	fmt.Println("\n--- 1. So sánh Print() và Println() ---")
	fmt.Print("Chào", "bạn!") // Nối liền: Chàobạn!
	fmt.Print(" Cùng học Go\n")

	fmt.Println("Dòng 1 với Println")
	fmt.Println("Dòng 2:", "Tuổi", 25, "Quê quán:", "Đà Nẵng")

	// --- PHẦN 2: GENERAL FORMATTING VERBS (CÁC VERB TỔNG QUÁT) ---
	fmt.Println("\n--- 2. Các formatting verbs tổng quát (verbs: v, #v, T, %) ---")
	type Product struct {
		Name  string
		Price float64
	}
	p := Product{Name: "MacBook Pro", Price: 1999.99}

	// %v: In giá trị mặc định
	fmt.Printf("%%v  (Giá trị mặc định)     : %v\n", p)

	// %#v: In cú pháp cấu trúc mã nguồn Go của biến
	fmt.Printf("%%#v (Đầy đủ cú pháp Go)    : %#v\n", p)

	// %T: In kiểu dữ liệu (Data Type)
	fmt.Printf("%%T  (Kiểu dữ liệu)         : %T\n", p)

	// %%: In ra ký tự dấu phần trăm '%'
	fmt.Printf("Giảm giá 10%% hôm nay!\n")

	// --- PHẦN 3: INTEGER FORMATTING VERBS (ĐỊNH DẠNG SỐ NGUYÊN) ---
	fmt.Println("\n--- 3. Định dạng số nguyên (Integer Verbs) ---")
	num := 255
	fmt.Printf("Thập phân (%%d)           : %d\n", num)
	fmt.Printf("Nhị phân (%%b)            : %b\n", num)
	fmt.Printf("Bát phân (%%o)            : %o\n", num)
	fmt.Printf("Thập lục phân thường (%%x): %x\n", num)
	fmt.Printf("Thập lục phân HOA (%%X)   : %X\n", num)
	fmt.Printf("Luôn có dấu (%%+d)        : %+d\n", num)

	// --- PHẦN 4: STRING & CHAR FORMATTING VERBS (CHUỖI & KÝ TỰ) ---
	fmt.Println("\n--- 4. Định dạng chuỗi (String Verbs) ---")
	msg := "Golang"
	fmt.Printf("Chuỗi thông thường (%%s)  : %s\n", msg)
	fmt.Printf("Chuỗi có ngoặc kép (%%q)  : %q\n", msg)
	fmt.Printf("Căn phải độ rộng 10 (%%10s): [%10s]\n", msg)
	fmt.Printf("Căn trái độ rộng 10 (%%-10s): [%-10s]\n", msg)

	// --- PHẦN 5: BOOLEAN & FLOAT FORMATTING VERBS ---
	fmt.Println("\n--- 5. Định dạng số thực (Float) và Boolean ---")
	isReady := true
	piValue := 3.14159265

	fmt.Printf("Boolean (%%t)              : %t\n", isReady)
	fmt.Printf("Float mặc định (%%f)      : %f\n", piValue)
	fmt.Printf("Lấy 2 chữ số lẻ (%%.2f)   : %.2f\n", piValue)
	fmt.Printf("Độ rộng 8, 3 số lẻ (%%8.3f): [%8.3f]\n", piValue)
	fmt.Printf("Dạng khoa học (%%e)       : %e\n", piValue)
}
