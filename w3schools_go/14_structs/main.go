// W3Schools Go Tutorial: Go Struct (Cấu Trúc Dữ Liệu Tự Định Nghĩa)
// Link tham khảo: https://www.w3schools.com/go/go_struct.php

package main

import "fmt"

// 1. Định nghĩa kiểu dữ liệu Struct bằng từ khóa "type" và "struct"
type Person struct {
	Name   string
	Age    int
	Job    string
	Salary int
}

// Struct lồng nhau (Nested Struct)
type Company struct {
	CompanyName string
	Location    string
	CEO         Person // Trường CEO là một struct Person
}

// Hàm nhận struct theo dạng GIÁ TRỊ (Pass by value - sao chép một bản copy)
func printPersonDetails(p Person) {
	fmt.Printf("Thông tin: %s, %d tuổi, Nghề nghiệp: %s, Lương: $%d\n",
		p.Name, p.Age, p.Job, p.Salary)
}

// Hàm nhận struct theo dạng CON TRỎ (Pass by pointer - cho phép sửa trực tiếp đối tượng gốc)
func increaseSalary(p *Person, amount int) {
	p.Salary += amount // Trong Go, p.Salary tự động giải con trỏ (*p).Salary
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 CHƯƠNG 14: GO STRUCTS (KIỂU DỮ LIỆU CẤU TRÚC)")
	fmt.Println("==================================================")

	// --- ĐẶC ĐIỂM CỦA STRUCT TRONG GO ---
	// - Go KHÔNG có class (lớp) hay kế thừa (inheritance) như OOP truyền thống.
	// - Struct là công cụ chính để tạo ra các đối tượng dữ liệu phức tạp.

	// --- 1. CÁC CÁCH KHỞI TẠO ĐỐI TƯỢNG STRUCT ---
	fmt.Println("\n--- 1. Các cách khởi tạo Struct ---")

	// Cách 1: Khởi tạo kèm tên trường cụ thể (Khuyên dùng - Rõ ràng, không sợ nhầm vị trí)
	person1 := Person{
		Name:   "Nguyen Van A",
		Age:    28,
		Job:    "Kỹ sư Phần mềm",
		Salary: 3000,
	}

	// Cách 2: Khởi tạo theo đúng thứ tự các trường (không điền tên trường)
	person2 := Person{"Tran Thi B", 24, "Chuyên viên Marketing", 1800}

	// Cách 3: Khởi tạo rỗng (nhận các zero-values)
	var person3 Person // Name: "", Age: 0, Job: "", Salary: 0

	printPersonDetails(person1)
	printPersonDetails(person2)
	printPersonDetails(person3)

	// --- 2. TRUY CẬP VÀ THAY ĐỔI CÁC TRƯỜNG (DOT NOTATION) ---
	fmt.Println("\n--- 2. Truy cập và sửa đổi giá trị bằng dấu chấm '.' ---")
	person3.Name = "Le Van C"
	person3.Age = 32
	person3.Job = "Quản lý Dự án"
	person3.Salary = 4000
	fmt.Println("Sau khi cập nhật person3:")
	printPersonDetails(person3)

	// --- 3. TRUYỀN STRUCT VÀO HÀM (VALUE VS POINTER) ---
	fmt.Println("\n--- 3. Truyền Struct vào hàm: Con trỏ (Pointer) vs Giá trị (Value) ---")
	fmt.Println("Lương trước khi tăng của person1:", person1.Salary)

	// Truyền địa chỉ bộ nhớ &person1 để hàm có thể thay đổi dữ liệu gốc:
	increaseSalary(&person1, 500)
	fmt.Println("Lương sau khi tăng (qua con trỏ *Person):", person1.Salary)

	// --- 4. STRUCT LỒNG NHAU (NESTED STRUCTS) ---
	fmt.Println("\n--- 4. Struct lồng nhau (Nested Struct) ---")
	techCorp := Company{
		CompanyName: "VinAI Tech",
		Location:    "Hà Nội",
		CEO: Person{
			Name:   "Dr. Hung",
			Age:    45,
			Job:    "CEO / AI Scientist",
			Salary: 15000,
		},
	}

	fmt.Printf("Công ty: %s tại %s | Giám đốc: %s (%d tuổi)\n",
		techCorp.CompanyName, techCorp.Location, techCorp.CEO.Name, techCorp.CEO.Age)
}
