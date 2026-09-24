package customengine

import (
	"os"
	"testing"
	"time"
	"vietnamese-chat-search/pkg/chat"
	"vietnamese-chat-search/pkg/invertedindex"
)

func TestCustomEngine_EndToEndLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_customengine_*")
	if err != nil {
		t.Fatalf("Lỗi tạo temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	analyzer := invertedindex.NewVietnameseAnalyzer()

	// 1. Khởi tạo Engine mới
	engine, err := Open(tempDir, analyzer)
	if err != nil {
		t.Fatalf("Open engine thất bại: %v", err)
	}

	messages := []chat.Message{
		{ID: 1, Sender: "An", Room: "general", Content: "Học sinh uống cà phê tại quán", CreatedAt: time.Now()},
		{ID: 2, Sender: "Binh", Room: "general", Content: "Chào mừng các bạn tân sinh viên", CreatedAt: time.Now()},
		{ID: 3, Sender: "Chi", Room: "tech", Content: "Cần mua bàn ghế và bàn học mới", CreatedAt: time.Now()},
	}

	// 2. Index tin nhắn
	for _, m := range messages {
		if err := engine.IndexMessage(m); err != nil {
			t.Fatalf("IndexMessage %d thất bại: %v", m.ID, err)
		}
	}

	// 3. Test Search trên MemTable (trước khi Flush)
	results, _, err := engine.Search("cà phê", "")
	if err != nil {
		t.Fatalf("Search 'cà phê' lỗi: %v", err)
	}
	if len(results) == 0 || results[0].Message.ID != 1 {
		t.Errorf("Kỳ vọng ID 1 đứng đầu khi tìm 'cà phê', got %+v", results)
	}

	// Test tìm không dấu
	unaccRes, _, err := engine.Search("ca phe", "")
	if err != nil {
		t.Fatalf("Search 'ca phe' lỗi: %v", err)
	}
	if len(unaccRes) == 0 || unaccRes[0].Message.ID != 1 {
		t.Errorf("Kỳ vọng ID 1 đứng đầu khi tìm không dấu 'ca phe', got %+v", unaccRes)
	}

	// 4. Test Flush xuống đĩa cứng (Binary Segments)
	if err := engine.Flush(); err != nil {
		t.Fatalf("Flush xuống đĩa thất bại: %v", err)
	}

	// 5. Đóng Engine và Mở lại từ đĩa (Persistence Verification)
	engine.Close()

	reopenedEngine, err := Open(tempDir, analyzer)
	if err != nil {
		t.Fatalf("Open lại engine sau khi đóng thất bại: %v", err)
	}
	defer reopenedEngine.Close()

	// Kiểm tra tìm kiếm sau khi khôi phục từ đĩa
	resultsAfterReopen, _, err := reopenedEngine.Search("cà phê", "")
	if err != nil {
		t.Fatalf("Search sau khi reopen lỗi: %v", err)
	}
	if len(resultsAfterReopen) == 0 || resultsAfterReopen[0].Message.ID != 1 {
		t.Errorf("Dữ liệu trên đĩa không trả về kết quả kỳ vọng sau khi reopen: %+v", resultsAfterReopen)
	}

	// 6. Test Xóa tin nhắn qua Tombstone
	if err := reopenedEngine.DeleteMessage(1); err != nil {
		t.Fatalf("DeleteMessage lỗi: %v", err)
	}

	// Tìm lại "cà phê" -> không còn thấy ID 1 nữa
	resultsAfterDelete, _, _ := reopenedEngine.Search("cà phê", "")
	if len(resultsAfterDelete) > 0 {
		t.Errorf("Tin nhắn ID 1 đáng lẽ đã bị xóa khỏi kết quả tìm kiếm, nhưng vẫn còn: %+v", resultsAfterDelete)
	}
}
