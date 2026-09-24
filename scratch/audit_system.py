import urllib.request
import urllib.parse
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

def audit_es():
    print("=" * 80)
    print("1. RÀ SOÁT DỮ LIỆU ĐANG LƯU TRÊN ELASTICSEARCH")
    print("=" * 80)
    
    # Check count
    try:
        res = urllib.request.urlopen("http://localhost:9200/chat_messages_vietnamese/_count")
        count_data = json.loads(res.read().decode('utf-8'))
        print(f"Tổng số documents trong chat_messages_vietnamese: {count_data.get('count')}")
    except Exception as e:
        print(f"Lỗi đọc ES count: {e}")
        return

    # Check 3 documents sample (ID 1, 15, 132/133)
    try:
        sample_ids = [1, 2, 14, 15]
        for sid in sample_ids:
            res = urllib.request.urlopen(f"http://localhost:9200/chat_messages_vietnamese/_doc/{sid}")
            doc = json.loads(res.read().decode('utf-8'))['_source']
            print(f"\n[Document ID {sid}]")
            print(f"  • content:            {doc.get('content')}")
            print(f"  • content_tokenized:  {doc.get('content_tokenized')}")
            print(f"  • content_unaccented: {doc.get('content_unaccented')}")
            print(f"  • content_partial:    {doc.get('content_partial')}")
    except Exception as e:
        print(f"Lỗi đọc sample docs: {e}")

def audit_queries():
    print("\n" + "=" * 80)
    print("2. RÀ SOÁT ĐẦU VÀO VÀ XỬ LÝ QUERY (TEST SUITE 12 KỊCH BẢN)")
    print("=" * 80)

    test_queries = [
        "học sinh",
        "hoc sinh",
        "sinh viên",
        "sinh vien",
        "cà phê",
        "ca phe",
        "phê bình",
        "phe binh",
        "trà sữa",
        "tra sua",
        "bàn ghế",
        "ban ghe",
    ]

    for q in test_queries:
        url = f"http://localhost:8080/api/search/compare?q={urllib.parse.quote(q)}"
        try:
            res = urllib.request.urlopen(url)
            data = json.loads(res.read().decode('utf-8'))
            tokens = data.get("tokens", [])
            v_res = data.get("vietnamese_results", [])
            c_res = data.get("custom_results", [])
            b_res = data.get("baseline_results", [])

            top_v = v_res[0]["message"]["content"] if v_res else "KHÔNG TÌM THẤY"
            top_v_score = v_res[0]["score"] if v_res else 0.0
            top_c = c_res[0]["message"]["content"] if c_res else "KHÔNG TÌM THẤY"
            top_c_score = c_res[0]["score"] if c_res else 0.0
            top_b = b_res[0]["message"]["content"] if b_res else "KHÔNG TÌM THẤY"
            top_b_score = b_res[0]["score"] if b_res else 0.0

            print(f"\n🔎 Query: '{q}' -> Tokens phân tích: {tokens}")
            print(f"   [Cốc Cốc ES]   (Top 1, BM25={top_v_score:.3f}): {top_v[:65]}...")
            print(f"   [Custom Go]    (Top 1, BM25={top_c_score:.3f}): {top_c[:65]}...")
            print(f"   [Baseline ES]  (Top 1, BM25={top_b_score:.3f}): {top_b[:65]}...")
        except Exception as e:
            print(f"Lỗi test query '{q}': {e}")

if __name__ == "__main__":
    audit_es()
    audit_queries()
