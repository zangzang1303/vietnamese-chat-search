// W3Schools Go Tutorial: Go Exercises & Challenges (Bài Tập Thực Hành & Củng Cố)
// Link tham khảo: https://www.w3schools.com/go/go_exercises.php

package main

import (
	"fmt"
	"strings"
)

// Struct phục vụ Bài tập 4
type BankAccount struct {
	AccountNumber string
	OwnerName     string
	Balance       float64
}

// Method gửi tiền
func (acc *BankAccount) Deposit(amount float64) {
	if amount > 0 {
		acc.Balance += amount
		fmt.Printf(" [Giao dịch] Nạp thành công: +$%.2f. Số dư mới: $%.2f\n", amount, acc.Balance)
	}
}

// Method rút tiền (trả về lỗi nếu không đủ tiền)
func (acc *BankAccount) Withdraw(amount float64) bool {
	if amount > acc.Balance {
		fmt.Printf(" [Giao dịch] Rút $%.2f THẤT BẠI: Số dư không đủ ($%.2f)!\n", amount, acc.Balance)
		return false
	}
	acc.Balance -= amount
	fmt.Printf(" [Giao dịch] Rút thành công: -$%.2f. Số dư còn lại: $%.2f\n", amount, acc.Balance)
	return true
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("🎯 CHƯƠNG 16: BÀI TẬP TỔNG HỢP THEO W3SCHOOLS GO")
	fmt.Println("==================================================")

	// ==========================================
	// BÀI TẬP 1: BIẾN, HẰNG SỐ & ĐỊNH DẠNG VERBS
	// ==========================================
	fmt.Println("\n📌 BÀI TẬP 1: TÍNH HÓA ĐƠN BÁN HÀNG VỚI THUẾ VAT")
	const VAT_RATE float64 = 0.10 // Thuế 10%
	productName := "Bàn phím cơ Không Dây"
	unitPrice := 1250000.0
	quantity := 2

	subTotal := unitPrice * float64(quantity)
	taxAmount := subTotal * VAT_RATE
	grandTotal := subTotal + taxAmount

	fmt.Printf("Sản phẩm       : %s\n", productName)
	fmt.Printf("Đơn giá        : %12.0f VNĐ\n", unitPrice)
	fmt.Printf("Số lượng       : %12d\n", quantity)
	fmt.Printf("Tạm tính       : %12.0f VNĐ\n", subTotal)
	fmt.Printf("Thuế VAT (10%%) : %12.0f VNĐ\n", taxAmount)
	fmt.Println("---------------------------------------")
	fmt.Printf("TỔNG CỘNG      : %12.0f VNĐ\n", grandTotal)

	// ==========================================
	// BÀI TẬP 2: THAO TÁC SLICE (TÌM MAX, MIN, TRUNG BÌNH)
	// ==========================================
	fmt.Println("\n📌 BÀI TẬP 2: THỐNG KÊ ĐIỂM THI VỚI SLICE")
	examScores := []float64{7.5, 8.0, 9.5, 6.0, 8.5, 10.0, 5.5}
	fmt.Println("Danh sách điểm:", examScores)

	minScore := examScores[0]
	maxScore := examScores[0]
	totalSum := 0.0

	for _, score := range examScores {
		if score < minScore {
			minScore = score
		}
		if score > maxScore {
			maxScore = score
		}
		totalSum += score
	}
	avgScore := totalSum / float64(len(examScores))

	fmt.Printf("-> Điểm thấp nhất : %.1f\n", minScore)
	fmt.Printf("-> Điểm cao nhất  : %.1f\n", maxScore)
	fmt.Printf("-> Điểm trung bình: %.2f\n", avgScore)

	// ==========================================
	// BÀI TẬP 3: ĐẾM TẦN SUẤT XUẤT HIỆN TỪ VỚI MAP
	// ==========================================
	fmt.Println("\n📌 BÀI TẬP 3: ĐẾM TỪ (WORD FREQUENCY) BẰNG MAP")
	paragraph := "go la ngon ngu lap trinh hien dai go rat nhanh va go de hoc"
	words := strings.Fields(paragraph) // Tách chuỗi thành slice các từ

	wordCounts := make(map[string]int)
	for _, w := range words {
		wordCounts[w]++
	}

	fmt.Println("Câu văn mẫu:", paragraph)
	fmt.Println("Kết quả đếm tần suất:")
	for word, count := range wordCounts {
		fmt.Printf("  Từ '%-8s' xuất hiện: %d lần\n", word, count)
	}

	// ==========================================
	// BÀI TẬP 4: STRUCT & QUẢN LÝ TÀI KHOẢN NGÂN HÀNG
	// ==========================================
	fmt.Println("\n📌 BÀI TẬP 4: MÔ PHỎNG TÀI KHOẢN NGÂN HÀNG (STRUCT & METHOD)")
	myAccount := BankAccount{
		AccountNumber: "VCB-888999",
		OwnerName:     "Nguyen Van A",
		Balance:       1000.0,
	}

	fmt.Printf("Chủ tài khoản: %s (STK: %s) - Số dư ban đầu: $%.2f\n",
		myAccount.OwnerName, myAccount.AccountNumber, myAccount.Balance)

	myAccount.Deposit(500.0)
	myAccount.Withdraw(300.0)
	myAccount.Withdraw(2000.0) // Thử rút quá số dư

	fmt.Println("\n🎉 CHÚC MỪNG BẠN ĐÃ HOÀN THÀNH TOÀN BỘ GIÁO TRÌNH W3SCHOOLS GO!")
	fmt.Println("Hãy thử tự mình sửa đổi code trong các thư mục và chạy lại nhé!")
}
