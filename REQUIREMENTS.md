# Đề bài & Yêu cầu ban đầu (Original Requirements)

## Tình trạng hiện tại:
- **Chat 1-1**: Đang mã hóa đầu cuối, server không thể search tin nhắn => mình không quan tâm case này.
- **Chat nhóm**: Không mã hóa, server có thể search, đã có search tin nhắn từ server nhưng kết quả search chưa tốt cho tiếng Việt.

## Bài toán:
Cải thiện kết quả search cho tiếng Việt sử dụng Cốc Cốc tokenizer:
- GitHub: [coccoc/coccoc-tokenizer](https://github.com/coccoc/coccoc-tokenizer) (high performance tokenizer for Vietnamese language).

## Scope:
- Search match (không phải semantic search).
- Hỗ trợ cả search có dấu và không dấu.
- Hỗ trợ partial matching (match 1 phần từ cần search).

## Yêu cầu tìm hiểu:
1. Làm việc cơ bản với ElasticSearch trong Go: tìm hiểu về ES, Lucene.
2. Hiểu cách hệ thống search hoạt động ở mức cơ bản:
   - **Luồng indexing**: tokenize -> filtering -> build inverted index -> storage.
   - **Luồng search**: tokenize -> search -> ranking results.

## Output mong muốn:
1. Trình bày kết quả tìm hiểu về hệ thống search.
2. **Build sample**:
   - Dựng ES ở local, seed data tiếng Việt tùy chọn.
   - Build tokenizer ở Go tích hợp Cốc Cốc lib (cần tìm hiểu cách gọi thư viện C trong Go qua CGO).
   - So sánh kết quả search khi có và không dùng bộ tokenizer phía trên.
3. **Nâng cao (optional)**:
   - Compare performance 2 luồng index và search khi có và không dùng tokenizer tiếng Việt.
