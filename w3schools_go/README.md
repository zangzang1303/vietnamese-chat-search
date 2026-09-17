# Khóa học Thực hành Go (Golang) theo W3Schools Tutorial

Thư mục này được thiết kế dựa trên toàn bộ nội dung của [W3Schools Go Tutorial](https://www.w3schools.com/go/index.php). Mỗi thư mục con đại diện cho một chủ đề lớn, có code mẫu hoàn chỉnh, chú thích chi tiết bằng tiếng Việt và có thể chạy trực tiếp từ dòng lệnh.

---

## 🧭 Mục lục Lộ trình Học tập

| STT | Thư mục | Nội dung W3Schools | Lệnh chạy thử |
|:---:|:---|:---|:---|
| 01 | [`01_intro_and_syntax`](./01_intro_and_syntax/main.go) | Go Intro, Get Started & Cấu trúc file Go | `go run ./w3schools_go/01_intro_and_syntax` |
| 02 | [`02_comments`](./02_comments/main.go) | Cách dùng ghi chú đơn dòng và nhiều dòng | `go run ./w3schools_go/02_comments` |
| 03 | [`03_variables`](./03_variables/main.go) | Khai báo biến (`var`, `:=`), đa biến, quy tắc đặt tên | `go run ./w3schools_go/03_variables` |
| 04 | [`04_constants`](./04_constants/main.go) | Khai báo hằng số (`const`), typed và untyped | `go run ./w3schools_go/04_constants` |
| 05 | [`05_output_and_formatting`](./05_output_and_formatting/main.go) | Các hàm Output (`Print`, `Println`, `Printf`) & Định dạng verbs | `go run ./w3schools_go/05_output_and_formatting` |
| 06 | [`06_data_types`](./06_data_types/main.go) | Các kiểu dữ liệu cơ bản (int, uint, float, bool, string) | `go run ./w3schools_go/06_data_types` |
| 07 | [`07_operators`](./07_operators/main.go) | Toán tử số học, gán, so sánh, logic, bitwise | `go run ./w3schools_go/07_operators` |
| 08 | [`08_conditions`](./08_conditions/main.go) | Điều kiện `if`, `else`, `else if`, `nested if`, `if-with-short-statement` | `go run ./w3schools_go/08_conditions` |
| 09 | [`09_switch`](./09_switch/main.go) | Cấu trúc rẽ nhánh `switch` đơn, đa trường hợp và switch không biểu thức | `go run ./w3schools_go/09_switch` |
| 10 | [`10_loops`](./10_loops/main.go) | Vòng lặp `for`, `break`, `continue`, `range` | `go run ./w3schools_go/10_loops` |
| 11 | [`11_arrays`](./11_arrays/main.go) | Mảng cố định (Array), khai báo, kích thước, gán giá trị | `go run ./w3schools_go/11_arrays` |
| 12 | [`12_slices`](./12_slices/main.go) | Mảng động (Slice), `make`, `len`, `cap`, `append`, `copy` | `go run ./w3schools_go/12_slices` |
| 13 | [`13_maps`](./13_maps/main.go) | Cặp Khóa-Giá trị (Key-Value), thêm, sửa, `delete`, kiểm tra key | `go run ./w3schools_go/13_maps` |
| 14 | [`14_structs`](./14_structs/main.go) | Kiểu dữ liệu cấu trúc (Struct), khởi tạo, lồng nhau, truyền vào hàm | `go run ./w3schools_go/14_structs` |
| 15 | [`15_functions`](./15_functions/main.go) | Hàm (Function), tham số, đa giá trị trả về, named return, đệ quy | `go run ./w3schools_go/15_functions` |
| 16 | [`16_exercises`](./16_exercises/main.go) | Tổng hợp các bài tập thực hành theo câu đố & bài tập W3Schools | `go run ./w3schools_go/16_exercises` |

---

## 🚀 Cách thực hành hiệu quả

1. **Cách 1 - Chạy từ thư mục gốc dự án:**
   ```bash
   go run ./w3schools_go/01_intro_and_syntax
   go run ./w3schools_go/02_comments
   # ... và tương tự cho các bài khác
   ```

2. **Cách 2 - Di chuyển vào từng thư mục bài học:**
   ```bash
   cd w3schools_go/05_output_and_formatting
   go run .
   ```

3. **Gợi ý học tập:**
   - Đọc phần chú thích code trong từng file `main.go`.
   - Thử thay đổi các giá trị biến, thử tạo lỗi để xem compiler Go báo lỗi gì.
   - Làm phần bài tập tự kiểm tra trong mục `16_exercises`.
