#ifndef COCCOC_BRIDGE_H
#define COCCOC_BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

// ==============================================================================
// COCCOC_BRIDGE_H: C Interface cho Cốc Cốc Tokenizer (C++ -> C -> Go CGO)
// ==============================================================================

/**
 * Khởi tạo Cốc Cốc Tokenizer từ thư mục chứa các file từ điển đã biên dịch.
 * 
 * @param dict_path: Đường dẫn đến thư mục chứa các tệp .dump và các tệp ngôn ngữ.
 * @param load_nontone: 
 *   - 0: Không nạp dữ liệu nontone (dành cho tách từ tin nhắn chat thông thường, 
 *        khởi động siêu nhanh ~150ms, RAM ~45MB).
 *   - 1: Nạp đầy đủ dữ liệu nontone (hỗ trợ tách từ dính liền sticky-text như URL/tên miền).
 * 
 * @return 0 nếu thành công, số âm nếu xảy ra lỗi (ví dụ không tìm thấy file từ điển).
 */
int coccoc_init(const char* dict_path, int load_nontone);

/**
 * Phân tách văn bản tiếng Việt thành chuỗi có gắn dấu gạch dưới "_" cho các từ ghép.
 * Ví dụ: "Học sinh uống cà phê" -> "Học_sinh uống cà_phê"
 * 
 * Chuỗi trả về được cấp phát trên C heap (malloc/strdup).
 * Caller (phía Go) PHẢI gọi hàm coccoc_free_string() để giải phóng bộ nhớ sau khi dùng xong!
 * 
 * @param text: Chuỗi văn bản tiếng Việt UTF-8 đầu vào.
 * @return Chuỗi kết quả UTF-8 (C string), hoặc NULL nếu text rỗng/lỗi.
 */
char* coccoc_tokenize_original(const char* text);

/**
 * Giải phóng bộ nhớ chuỗi C được cấp phát bởi coccoc_tokenize_original().
 * 
 * @param str: Con trỏ chuỗi cần giải phóng.
 */
void coccoc_free_string(char* str);

#ifdef __cplusplus
}
#endif

#endif // COCCOC_BRIDGE_H
