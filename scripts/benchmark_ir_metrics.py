import urllib.parse
import urllib.request
import json
import math
import os

# 1. Định nghĩa Test Suite với Ground Truth Categories
TEST_SUITE = [
    {
        "query": "học sinh",
        "target_categories": ["trap_hoc_sinh_true"],
        "trap_categories": ["trap_hoc_sinh_false_sinh_vien", "trap_hoc_sinh_false_hy_sinh", "trap_hoc_sinh_false_sinh_nhat"],
        "type": "Từ ghép có dấu"
    },
    {
        "query": "sinh viên",
        "target_categories": ["trap_hoc_sinh_false_sinh_vien"],
        "trap_categories": ["trap_hoc_sinh_true", "trap_hoc_sinh_false_hy_sinh", "trap_hoc_sinh_false_sinh_nhat"],
        "type": "Từ ghép có dấu"
    },
    {
        "query": "cà phê",
        "target_categories": ["trap_ca_phe_true"],
        "trap_categories": ["trap_ca_phe_false_phe_binh", "trap_ca_phe_false_phe_duyet"],
        "type": "Từ ghép có dấu"
    },
    {
        "query": "phê bình",
        "target_categories": ["trap_ca_phe_false_phe_binh"],
        "trap_categories": ["trap_ca_phe_true", "trap_ca_phe_false_phe_duyet"],
        "type": "Từ ghép có dấu"
    },
    {
        "query": "trà sữa",
        "target_categories": ["trap_tra_sua_true"],
        "trap_categories": ["trap_tra_sua_false_sua_chua", "trap_tra_sua_false_tra_da"],
        "type": "Từ ghép có dấu"
    },
    {
        "query": "bàn ghế",
        "target_categories": ["trap_ban_ghe_true"],
        "trap_categories": ["trap_ban_ghe_false_ban_bac", "trap_ban_ghe_false_ban_luan"],
        "type": "Từ ghép có dấu"
    },
    {
        "query": "bàn bạc",
        "target_categories": ["trap_ban_ghe_false_ban_bac"],
        "trap_categories": ["trap_ban_ghe_true", "trap_ban_ghe_false_ban_luan"],
        "type": "Từ ghép có dấu"
    },
    {
        "query": "sửa chữa",
        "target_categories": ["trap_tra_sua_false_sua_chua"],
        "trap_categories": ["trap_tra_sua_true"],
        "type": "Từ ghép có dấu"
    },
    {
        "query": "hoc sinh",
        "target_categories": ["trap_hoc_sinh_true"],
        "trap_categories": ["trap_hoc_sinh_false_sinh_vien", "trap_hoc_sinh_false_hy_sinh", "trap_hoc_sinh_false_sinh_nhat"],
        "type": "Không dấu (Unaccented)"
    },
    {
        "query": "ca phe",
        "target_categories": ["trap_ca_phe_true"],
        "trap_categories": ["trap_ca_phe_false_phe_binh", "trap_ca_phe_false_phe_duyet"],
        "type": "Không dấu (Unaccented)"
    },
    {
        "query": "tra sua",
        "target_categories": ["trap_tra_sua_true"],
        "trap_categories": ["trap_tra_sua_false_sua_chua", "trap_tra_sua_false_tra_da"],
        "type": "Không dấu (Unaccented)"
    },
    {
        "query": "ban ghe",
        "target_categories": ["trap_ban_ghe_true"],
        "trap_categories": ["trap_ban_ghe_false_ban_bac", "trap_ban_ghe_false_ban_luan"],
        "type": "Không dấu (Unaccented)"
    }
]

# Nạp metadata category của toàn bộ tin nhắn
def load_message_categories():
    candidates = ["data/sample_messages.json", "../data/sample_messages.json"]
    for path in candidates:
        if os.path.exists(path):
            with open(path, "r", encoding="utf-8") as f:
                data = json.load(f)
                return {item["id"]: item.get("category", "") for item in data}
    return {}

MSG_CATEGORIES = load_message_categories()

def get_relevance_label(msg_id, content, test_case):
    category = MSG_CATEGORIES.get(msg_id, "")
    target_cats = test_case["target_categories"]
    trap_cats = test_case["trap_categories"]

    # 3: Khớp đúng category mục tiêu (Relevant - True Match)
    if category in target_cats:
        return 3
    # 0: Khớp vào bẫy từ ghép (False Positive - Trap Match)
    if category in trap_cats:
        return 0

    # Nếu không thuộc trap định sẵn, kiểm tra xem nội dung có chứa từ khóa nguyên vẹn không
    q = test_case["query"].lower()
    c = content.lower()
    if q in c:
        return 2
    return 0

def compute_dcg_at_k(relevance_scores, k=10):
    dcg = 0.0
    for i, rel in enumerate(relevance_scores[:k], 1):
        if rel > 0:
            # Standard IR formula: (2^rel - 1) / log2(i + 1)
            gain = (2.0 ** rel) - 1.0
            discount = math.log2(i + 1.0)
            dcg += gain / discount
    return dcg

def compute_ndcg_at_k(relevance_scores, k=10):
    actual_dcg = compute_dcg_at_k(relevance_scores, k)
    # Sắp xếp giảm dần lý tưởng
    ideal_scores = sorted(relevance_scores, reverse=True)
    ideal_dcg = compute_dcg_at_k(ideal_scores, k)
    if ideal_dcg == 0.0:
        return 1.0 if actual_dcg == 0.0 else 0.0
    return actual_dcg / ideal_dcg

def compute_reciprocal_rank(relevance_scores, threshold=2):
    for i, rel in enumerate(relevance_scores, 1):
        if rel >= threshold:
            return 1.0 / i
    return 0.0

def compute_precision_at_k(relevance_scores, k=5, threshold=2):
    sub = relevance_scores[:k]
    if not sub:
        return 0.0
    relevant_count = sum(1 for rel in sub if rel >= threshold)
    return relevant_count / len(sub)

def run_benchmark():
    results = []

    for tc in TEST_SUITE:
        q = tc["query"]
        encoded_q = urllib.parse.quote(q)
        url = f"http://localhost:8080/api/search/compare?q={encoded_q}"

        try:
            req = urllib.request.Request(url)
            with urllib.request.urlopen(req) as resp:
                data = json.loads(resp.read().decode("utf-8"))

            vn_items = data.get("vietnamese_results", [])
            base_items = data.get("baseline_results", [])
            lat_vn = data.get("latency_vietnamese_ms", 0)
            lat_base = data.get("latency_baseline_ms", 0)

            # Tính relevance labels cho Cốc Cốc
            vn_rels = [
                get_relevance_label(item["message"]["id"], item["message"]["content"], tc)
                for item in vn_items
            ]

            # Tính relevance labels cho Baseline
            base_rels = [
                get_relevance_label(item["message"]["id"], item["message"]["content"], tc)
                for item in base_items
            ]

            # Metrics
            rr_vn = compute_reciprocal_rank(vn_rels)
            rr_base = compute_reciprocal_rank(base_rels)

            ndcg_vn = compute_ndcg_at_k(vn_rels, 10)
            ndcg_base = compute_ndcg_at_k(base_rels, 10)

            p1_vn = 1.0 if (len(vn_rels) > 0 and vn_rels[0] >= 2) else 0.0
            p1_base = 1.0 if (len(base_rels) > 0 and base_rels[0] >= 2) else 0.0

            p5_vn = compute_precision_at_k(vn_rels, 5)
            p5_base = compute_precision_at_k(base_rels, 5)

            results.append({
                "query": q,
                "type": tc["type"],
                "vn": {
                    "rr": rr_vn,
                    "ndcg10": ndcg_vn,
                    "p1": p1_vn,
                    "p5": p5_vn,
                    "latency": lat_vn,
                    "hits": len(vn_items),
                    "rels": vn_rels
                },
                "base": {
                    "rr": rr_base,
                    "ndcg10": ndcg_base,
                    "p1": p1_base,
                    "p5": p5_base,
                    "latency": lat_base,
                    "hits": len(base_items),
                    "rels": base_rels
                }
            })

        except Exception as e:
            print(f"Lỗi khi query '{q}': {e}")

    return results

if __name__ == "__main__":
    benchmark_data = run_benchmark()
    
    # In kết quả tổng hợp ra JSON stdout để các script/Go/Markdown dễ phân tích
    print(json.dumps(benchmark_data, ensure_ascii=False, indent=2))
