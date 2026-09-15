# Danh Mục Tài Liệu Nghiên Cứu (Study & Reference Links)

Tài liệu được phân loại theo từng chủ đề cốt lõi của bài toán, sắp xếp theo thứ tự ưu tiên đọc để bạn nắm vững từ lý thuyết gốc đến thực hành code.

---

## 📚 1. Lý Thuyết Nền Tảng: Search Engine, Inverted Index & Lucene

Hiểu được cách dữ liệu được chia nhỏ, đánh chỉ mục và xếp hạng.

* **Cấu trúc Inverted Index (Chỉ mục đảo)**:
  * [Elasticsearch Guide: Making Text Searchable (Inverted Index)](https://www.elastic.co/guide/en/elasticsearch/guide/current/inverted-index.html)  
    *Tài liệu kinh điển của Elastic giải thích trực quan: Inverted index là gì, Tokenization hoạt động thế nào và tại sao tìm kiếm lại siêu nhanh.*
  * [Apache Lucene - How it works](https://lucenetutorial.com/lucene-in-5-minutes.html)  
    *Cơ chế cốt lõi của Lucene: Documents, Fields, Terms, Segments và Posting List.*

* **Luồng Xử Lý Văn Bản (Text Analysis Pipeline)**:
  * [Elasticsearch: Anatomy of an Analyzer](https://www.elastic.co/guide/en/elasticsearch/reference/current/analyzer-anatomy.html)  
    *Giải thích 3 thành phần: Character Filters -> Tokenizer -> Token Filters.*
  * [Edge N-gram Token Filter](https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-edgengram-tokenfilter.html)  
    *Cực kỳ quan trọng cho yêu cầu **Partial Matching** (gõ tiền tố như "cà ph" ra "cà phê").*

* **Thuật Toán Xếp Hạng & Điểm Số (Scoring - BM25)**:
  * [Practical BM25: The BM25 Algorithm and its Variables (Elastic Blog)](https://www.elastic.co/blog/practical-bm25-part-2-the-bm25-algorithm-and-its-variables)  
    *Hiểu cách ES tính điểm: TF (tần suất từ), IDF (độ hiếm của từ) và Document Length Normalization.*

* **📖 Sách Chuyên Sâu (PDF Nội Bộ Dự Án)**:
  * [Introduction to Information Retrieval (PDF)](irbookonlinereading.pdf) — *Christopher D. Manning, Prabhakar Raghavan, Hinrich Schütze (Cambridge University Press)*  
    *Giáo trình "kinh thánh" về Information Retrieval. Đọc trọng tâm: Chương 1-2 (Cấu trúc Inverted Index, Postings list, Dictionary), Chương 6-11 (Mô hình vector, TF-IDF, thuật toán BM25 và Ranking).*
  * [Elasticsearch: The Definitive Guide (PDF)](elasticsearch-the-definitive-guide.pdf) — *Clinton Gormley, Zachary Tong (O'Reilly)*  
    *Sách hướng dẫn toàn diện nhất về Elasticsearch. Đọc trọng tâm: Phần Inside a Shard (Inverted Index, Segments), Phần Getting Started with Languages & Analysis, và Phần Controlling Relevance.*

---

## 🇻🇳 2. Bài Toán Tách Từ Tiếng Việt & Cốc Cốc Tokenizer

Hiểu vì sao tiếng Việt cần tách từ riêng biệt và cách hoạt động của Cốc Cốc Tokenizer.

* **Bài toán Tách từ Tiếng Việt (Word Segmentation)**:
  * [Vietnamese Word Segmentation Overview (VNTK)](https://github.com/vntk/vntk#t%C3%A1ch-t%E1%BB%AB-word-segmentation)  
    *Giải thích vì sao tiếng Việt khác tiếng Anh: "học sinh" là 1 từ gồm 2 âm tiết, nếu tách bằng space sẽ gây sai lệch ngữ nghĩa.*

* **Thư viện Cốc Cốc Tokenizer**:
  * [coccoc/coccoc-tokenizer (GitHub chính thức)](https://github.com/coccoc/coccoc-tokenizer)  
    *Mã nguồn C++, kiến trúc từ điển `sys.dic`, hướng dẫn compile bằng CMake và benchmark hiệu năng.*
  * [minh-vib/coccoc-tokenizer-go (Go Binding tham khảo)](https://github.com/minh-vib/coccoc-tokenizer-go)  
    *Mã nguồn Go sử dụng CGO để gọi thư viện Cốc Cốc C++, rất hữu ích để học cách đóng gói CGO.*

---

## 🐹 3. Lập Trình CGO Trong Go (Gọi Thư Viện C/C++)

Cần nắm vững để thực hiện yêu cầu: *"Build tokenizer ở Go tích hợp Cốc Cốc lib qua CGO"*.

* [C? Go? Cgo! (Official Go Blog)](https://go.dev/blog/cgo)  
  *Bài viết nhập môn chính thức từ Go team, giải thích cách viết comment `#include` và gọi hàm C từ Go.*
* [Go Cgo Documentation (pkg.go.dev)](https://pkg.go.dev/cmd/cgo)  
  *Tài liệu kỹ thuật đầy đủ: CFLAGS, LDFLAGS, ép kiểu dữ liệu giữa C và Go (`C.CString`, `C.free`).*
* [Cgo Best Practices & Gotchas (Go Wiki)](https://github.com/golang/go/wiki/cgo)  
  *Các lưu ý về memory leak, quản lý con trỏ và chi phí chuyển ngữ cảnh (context switch overhead).*

---

## 🔍 4. Làm Việc Với Elasticsearch Trong Go

Cách kết nối, định nghĩa mapping, bulk index và query từ Go.

* [Official Elasticsearch Go Client Repository](https://github.com/elastic/go-elasticsearch)  
  *Thư viện chính thức của Elastic cho Go (hỗ trợ ES 7.x, 8.x).*
* [Elasticsearch Go Client User Guide](https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/index.html)  
  *Hướng dẫn chính thức: Khởi tạo client, tạo index, Index Document, Bulk API và Search API.*
* [Elasticsearch Query DSL (Bool & Multi-match Query)](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-bool-query.html)  
  *Cú pháp viết truy vấn `bool`: kết hợp `must`, `should`, `filter` và gắn điểm trọng số `boost` cho từng trường.*

---

## 🗺 Lộ Trình Đọc Khuyến Nghị (Reading Order)

1. **Bước 1 (1 buổi)**: Đọc [Inverted Index](https://www.elastic.co/guide/en/elasticsearch/guide/current/inverted-index.html) và [Anatomy of an Analyzer](https://www.elastic.co/guide/en/elasticsearch/reference/current/analyzer-anatomy.html) để hiểu bản chất search engine.
2. **Bước 2 (1 buổi)**: Xem mã nguồn [coccoc-tokenizer](https://github.com/coccoc/coccoc-tokenizer) và đọc bài viết [C? Go? Cgo!](https://go.dev/blog/cgo) để hiểu cơ chế CGO.
3. **Bước 3 (1 buổi)**: Đọc [Elasticsearch Go Client Guide](https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/index.html) và [Edge N-gram Filter](https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-edgengram-tokenfilter.html) để chuẩn bị cho phần mapping và query.
