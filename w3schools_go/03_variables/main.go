// W3Schools Go Tutorial: Go Variables (Biến số), Multi-Variables & Naming Rules
// Link tham khảo: 
// - https://www.w3schools.com/go/go_variables.php
// - https://www.w3schools.com/go/go_variable_multi.php
// - https://www.w3schools.com/go/go_variable_naming_rules.php

package main

import "fmt"

// Biến phạm vi toàn cục (Package level): Bắt buộc phải dùng từ khóa "var", KHÔNG được dùng ":="
var globalMessage string = "Đây là biến toàn cục (Package level)"

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 03: GO VARIABLES (KHAI BÁO BIẾN)")
	fmt.Println("==================================================")

	// --- PHẦN 1: CÁC CÁCH KHAI BÁO BIẾN CƠ BẢN ---
	fmt.Println("\n--- 1. Các cú pháp khai báo biến ---")

	// Cách 1: Dùng từ khóa var có chỉ định rõ kiểu dữ liệu
	var studentName string = "Nguyen Van A"
	var studentAge int = 20

	// Cách 2: Dùng var không chỉ định kiểu (Go tự động suy luận kiểu - Type Inference)
	var score = 9.5 // Go tự hiểu score là float64

	// Cách 3: Dùng toán tử ngắn ":=" (Walrus operator)
	// LƯU Ý: ":=" CHỈ ĐƯỢC DÙNG BÊN TRONG HÀM, không dùng ngoài hàm!
	city := "Hà Nội"
	isGraduated := true

	fmt.Println("Tên:", studentName)
	fmt.Println("Tuổi:", studentAge)
	fmt.Println("Điểm:", score)
	fmt.Println("Thành phố:", city)
	fmt.Println("Đã tốt nghiệp?:", isGraduated)
	fmt.Println("Biến toàn cục:", globalMessage)

	// --- PHẦN 2: GIÁ TRỊ MẶC ĐỊNH (DEFAULT / ZERO VALUES) ---
	fmt.Println("\n--- 2. Giá trị mặc định (Zero Values) khi chưa gán ---")
	// Trong Go, nếu khai báo biến mà không gán giá trị, biến sẽ tự nhận "zero value":
	var defaultInt int       // 0
	var defaultFloat float64 // 0
	var defaultString string // "" (chuỗi rỗng)
	var defaultBool bool     // false

	fmt.Printf("Default int: %d\n", defaultInt)
	fmt.Printf("Default float: %f\n", defaultFloat)
	fmt.Printf("Default string: '%s'\n", defaultString)
	fmt.Printf("Default bool: %t\n", defaultBool)

	// --- PHẦN 3: KHAI BÁO NHIỀU BIẾN CÙNG LÚC (MULTI-VARIABLE) ---
	fmt.Println("\n--- 3. Khai báo nhiều biến (Multiple Variables) ---")

	// 3.1 Cùng kiểu dữ liệu trên một dòng:
	var x, y, z int = 10, 20, 30
	fmt.Printf("x = %d, y = %d, z = %d\n", x, y, z)

	// 3.2 Khác kiểu dữ liệu (tự suy luận kiểu):
	var productID, productName, inStock = 101, "Bàn phím cơ", true
	fmt.Println("Sản phẩm:", productID, productName, "Còn hàng:", inStock)

	// 3.3 Khai báo nhiều biến với :=
	width, height := 1920, 1080
	fmt.Printf("Độ phân giải: %d x %d\n", width, height)

	// 3.4 Khai báo theo khối block `var (...)`:
	var (
		serverHost string = "localhost"
		serverPort int    = 8080
		isRunning  bool   = true
	)
	fmt.Println("Server:", serverHost, "Port:", serverPort, "Running:", isRunning)

	// --- PHẦN 4: QUY TẮC ĐẶT TÊN BIẾN (NAMING RULES) ---
	fmt.Println("\n--- 4. Quy tắc đặt tên biến (Variable Naming Rules) ---")
	// - Tên biến phải bắt đầu bằng chữ cái hoặc dấu gạch dưới (_)
	// - Không được bắt đầu bằng chữ số (ví dụ 1name là sai)
	// - Chỉ chứa ký tự chữ, số và dấu gạch dưới (a-z, A-Z, 0-9, _)
	// - Phân biệt HOA/thường: `myvar` khác `myVar`
	// - Không được trùng với các từ khóa của Go: func, var, package, if, for, v.v.

	// 3 phong cách đặt tên phổ biến:
	// Camel Case: chữ đầu viết thường, các chữ sau viết hoa
	var myItemCount int = 5

	// Pascal Case: mọi chữ cái đầu từ đều viết hoa (ĐẶC BIỆT TRONG GO: Bắt đầu chữ hoa nghĩa là EXPORTED - public ra ngoài package)
	var TotalRevenue float64 = 1500000.50

	// Snake Case: dùng dấu gạch dưới phân tách
	var user_login_attempt int = 3

	fmt.Println("CamelCase:", myItemCount)
	fmt.Println("PascalCase (Exported):", TotalRevenue)
	fmt.Println("Snake_case:", user_login_attempt)
}
