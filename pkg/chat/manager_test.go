package chat_test

import (
	"testing"

	"vietnamese-chat-search/pkg/chat"
	"vietnamese-chat-search/pkg/invertedindex"
)

func TestChatManager_RealTimeIndexingAndUpdating(t *testing.T) {
	// Khởi tạo ChatManager với VietnameseAnalyzer
	analyzer := invertedindex.NewVietnameseAnalyzer()
	cm := chat.NewChatManager(analyzer)

	// BƯỚC 1: Gửi tin nhắn đầu tiên
	msg1 := cm.PostMessage("Nguyễn Văn A", "Lớp Học", "Em là học sinh cấp 3 đang ôn thi")
	if msg1.ID != 1 {
		t.Fatalf("Mong đợi ID = 1, nhận được %d", msg1.ID)
	}

	// BƯỚC 2: Tìm kiếm từ khóa "học sinh"
	results1 := cm.Search("học sinh", "")
	if len(results1) == 0 {
		t.Fatalf("Lỗi: Không tìm thấy tin nhắn vừa gửi với từ khóa 'học sinh'!")
	}
	if results1[0].Message.ID != msg1.ID {
		t.Fatalf("Mong đợi tìm thấy tin nhắn ID 1, nhận được ID %d", results1[0].Message.ID)
	}
	t.Logf("✅ Tìm thấy thành công tin nhắn 1 với điểm BM25: %.4f", results1[0].Score)

	// BƯỚC 3: Cập nhật tin nhắn 1 (đổi thành sinh viên, bỏ chữ học sinh)
	updatedContent := "Em đã tốt nghiệp và hiện là sinh viên bách khoa"
	_, err := cm.UpdateMessage(msg1.ID, updatedContent)
	if err != nil {
		t.Fatalf("Lỗi khi cập nhật tin nhắn: %v", err)
	}

	// BƯỚC 4: Tìm lại từ khóa cũ "học sinh" -> Bắt buộc KHÔNG được tìm thấy nữa!
	resultsOld := cm.Search("học sinh", "")
	if len(resultsOld) > 0 {
		t.Fatalf("LỖI: Tin nhắn đã sửa bỏ chữ 'học sinh' nhưng tìm kiếm vẫn trả về kết quả! Re-index bị lỗi!")
	}
	t.Logf("✅ Đã kiểm tra: Tìm kiếm 'học sinh' trả về 0 kết quả như mong đợi sau khi sửa tin nhắn.")

	// BƯỚC 5: Tìm từ khóa mới "sinh viên" -> Phải tìm thấy ngay lập tức!
	resultsNew := cm.Search("sinh viên", "")
	if len(resultsNew) == 0 {
		t.Fatalf("LỖI: Không tìm thấy tin nhắn sau khi cập nhật với từ khóa mới 'sinh viên'!")
	}
	if resultsNew[0].Message.Content != updatedContent {
		t.Fatalf("Nội dung tin nhắn không khớp với nội dung sau khi sửa!")
	}
	t.Logf("✅ Đã kiểm tra: Tìm kiếm 'sinh viên' tìm thấy ngay lập tức với điểm BM25: %.4f", resultsNew[0].Score)

	// BƯỚC 6: Kiểm tra chi tiết chỉ mục (Inspect)
	details := cm.InspectMessageIndex(msg1.ID)
	if details == nil {
		t.Fatalf("Không lấy được chi tiết chỉ mục của tin nhắn 1")
	}
	foundSinhVien := false
	for _, p := range details.PostingList {
		if p.Term == "sinh_viên" {
			foundSinhVien = true
			break
		}
	}
	if !foundSinhVien {
		t.Fatalf("Không tìm thấy term 'sinh_viên' trong Posting List của tin nhắn 1!")
	}
	t.Logf("✅ Inspect kiểm tra thành công: Term 'sinh_viên' đã nằm gọn trong Posting List.")

	// BƯỚC 7: Xóa tin nhắn
	deleted := cm.DeleteMessage(msg1.ID)
	if !deleted {
		t.Fatalf("Không xóa được tin nhắn 1")
	}
	resultsAfterDelete := cm.Search("sinh viên", "")
	if len(resultsAfterDelete) > 0 {
		t.Fatalf("LỖI: Tin nhắn đã bị xóa nhưng vẫn xuất hiện trong kết quả tìm kiếm!")
	}
	t.Logf("✅ Đã kiểm tra: Sau khi xóa tin nhắn, tìm kiếm trả về rỗng.")

	// BƯỚC 8: Kiểm tra Stats
	stats := cm.GetStats()
	if stats.TotalMessages != 0 || stats.TotalTokens != 0 {
		t.Fatalf("LỖI: Stats sau khi xóa hết không về 0: %+v", stats)
	}
	t.Logf("✅ Stats dọn dẹp sạch sẽ: %+v", stats)
}
