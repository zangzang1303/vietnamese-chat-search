// W3Schools Go Tutorial: Go Operators (Toán Tử)
// Link tham khảo:
// - https://www.w3schools.com/go/go_operators.php
// - https://www.w3schools.com/go/go_arithmetic_operators.php
// - https://www.w3schools.com/go/go_assignment_operators.php
// - https://www.w3schools.com/go/go_comparison_operators.php
// - https://www.w3schools.com/go/go_logical_operators.php
// - https://www.w3schools.com/go/go_bitwise_operators.php

package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 07: GO OPERATORS (TOÁN TỬ)")
	fmt.Println("==================================================")

	// --- 1. TOÁN TỬ SỐ HỌC (ARITHMETIC OPERATORS) ---
	fmt.Println("\n--- 1. Toán tử số học (+, -, *, /, %, ++, --) ---")
	a, b := 10, 3
	fmt.Printf("%d + %d = %d\n", a, b, a+b)
	fmt.Printf("%d - %d = %d\n", a, b, a-b)
	fmt.Printf("%d * %d = %d\n", a, b, a*b)
	fmt.Printf("%d / %d = %d (Chia lấy phần nguyên)\n", a, b, a/b)
	fmt.Printf("%d %% %d = %d (Chia lấy phần dư)\n", a, b, a%b)

	// ĐẶC BIỆT TRONG GO: ++ và -- là CÂU LỆNH (Statement), KHÔNG phải biểu thức (Expression)!
	// Nghĩa là: chỉ được viết a++ hoặc a--, KHÔNG được viết: b = a++ hoặc ++a (bị báo lỗi cú pháp).
	a++
	fmt.Println("Sau khi a++:", a)
	b--
	fmt.Println("Sau khi b--:", b)

	// --- 2. TOÁN TỬ GÁN (ASSIGNMENT OPERATORS) ---
	fmt.Println("\n--- 2. Toán tử gán (=, +=, -=, *=, /=, %=) ---")
	x := 20
	x += 5 // x = x + 5 (25)
	fmt.Println("x += 5 ->", x)
	x *= 2 // x = x * 2 (50)
	fmt.Println("x *= 2 ->", x)

	// --- 3. TOÁN TỬ SO SÁNH (COMPARISON OPERATORS) ---
	fmt.Println("\n--- 3. Toán tử so sánh (==, !=, >, <, >=, <=) ---")
	val1, val2 := 15, 20
	fmt.Printf("%d == %d : %t\n", val1, val2, val1 == val2)
	fmt.Printf("%d != %d : %t\n", val1, val2, val1 != val2)
	fmt.Printf("%d > %d  : %t\n", val1, val2, val1 > val2)
	fmt.Printf("%d <= %d : %t\n", val1, val2, val1 <= val2)

	// --- 4. TOÁN TỬ LOGIC (LOGICAL OPERATORS: &&, ||, !) ---
	fmt.Println("\n--- 4. Toán tử Logic (&& - AND, || - OR, ! - NOT) ---")
	isAdult := true
	hasLicense := false

	canDrive := isAdult && hasLicense
	needSupervisor := isAdult || hasLicense
	cannotDrive := !canDrive

	fmt.Println("Được phép lái xe (isAdult && hasLicense):", canDrive)
	fmt.Println("Cần giám sát (isAdult || hasLicense):", needSupervisor)
	fmt.Println("Không được lái xe (!canDrive):", cannotDrive)

	// --- 5. TOÁN TỬ BITWISE (THAO TÁC TRÊN BIT NHỊ PHÂN) ---
	fmt.Println("\n--- 5. Toán tử Bitwise (&, |, ^, <<, >>) ---")
	n1 := 12 // Nhị phân: 1100
	n2 := 25 // Nhị phân: 0001 1001

	fmt.Printf("%d & %d  (AND bitwise) : %d (Nhị phân: %b)\n", n1, n2, n1&n2, n1&n2)
	fmt.Printf("%d | %d  (OR bitwise)  : %d (Nhị phân: %b)\n", n1, n2, n1|n2, n1|n2)
	fmt.Printf("%d ^ %d  (XOR bitwise) : %d (Nhị phân: %b)\n", n1, n2, n1^n2, n1^n2)
	fmt.Printf("%d << 2 (Dịch trái 2 bit): %d\n", n1, n1<<2)
	fmt.Printf("%d >> 2 (Dịch phải 2 bit): %d\n", n1, n1>>2)
}
