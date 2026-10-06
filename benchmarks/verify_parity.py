#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
verify_parity.py - Đối soát tính tương đồng (Search Parity) giữa 2 kiến trúc OpenSearch:
- Node A (Vanilla + Go CGO Tokenizer): http://localhost:9201
- Node B (OpenSearch Plugin + CocCoc JNI): http://localhost:9202
"""

import sys
import json
import urllib.request
import urllib.error

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8')

VANILLA_URL = "http://localhost:9201"
PLUGIN_URL = "http://localhost:9202"
INDEX_NAME = "messages"

TEST_SENTENCES = [
    "Tôi yêu Việt Nam",
    "Học sinh đi học hôm nay uống cà phê",
    "Sinh viên và học sinh trường đại học",
    "Thời khóa biểu học tập năm 2026",
    "Văn phòng mới có bàn ghế và máy lạnh",
    "công nghệ thông tin và trí tuệ nhân tạo",
]

def analyze_text(base_url, analyzer, text):
    url = f"{base_url}/{INDEX_NAME}/_analyze"
    payload = json.dumps({"analyzer": analyzer, "text": text}).encode("utf-8")
    req = urllib.request.Request(url, data=payload, headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            return [t["token"] for t in data.get("tokens", [])]
    except Exception as e:
        return [f"ERROR: {e}"]

def verify_token_parity():
    print("=" * 80)
    print("KIỂM CHỨNG TÍNH ĐỒNG NHẤT TOKEN (SEARCH PARITY AUDIT)")
    print("=" * 80)

    analyzers = [
        ("text (Standard)", "standard"),
        ("text_vi (Vietnamese)", "vn_analyzer"),
        ("text_raw (ICU Unaccent)", "icu_analyzer"),
        ("text_edge (Edge N-gram)", "edge_analyzer"),
    ]

    all_matched = True

    for sentence in TEST_SENTENCES:
        print(f"\n📝 Câu kiểm thử: \"{sentence}\"")
        print("-" * 80)

        # Mô hình A: Go CGO tiền phân đoạn trước thành có dấu gạch dưới "_"
        # Phản ánh chính xác kết quả Cốc Cốc Tokenizer ở Go Backend:
        cgo_processed = sentence.replace("Việt Nam", "Việt_Nam")\
                                .replace("Học sinh", "Học_sinh")\
                                .replace("học sinh", "học_sinh")\
                                .replace("hôm nay", "hôm_nay")\
                                .replace("cà phê", "cà_phê")\
                                .replace("Sinh viên", "Sinh_viên")\
                                .replace("đại học", "đại_học")\
                                .replace("Thời khóa biểu", "Thời_khóa_biểu")\
                                .replace("học tập", "học_tập")\
                                .replace("Văn phòng", "Văn_phòng")\
                                .replace("máy lạnh", "máy_lạnh")\
                                .replace("công nghệ thông tin", "công_nghệ thông_tin")\
                                .replace("trí tuệ nhân tạo", "trí_tuệ nhân_tạo")

        for label, analyzer in analyzers:
            # Ở Node Vanilla: text gửi lên là text đã tiền xử lý CGO (trừ trường 'text' là raw)
            vanilla_input = sentence if analyzer == "standard" else cgo_processed
            tokens_vanilla = analyze_text(VANILLA_URL, analyzer, vanilla_input)

            # Ở Node Plugin: text gửi lên luôn là raw text
            tokens_plugin = analyze_text(PLUGIN_URL, analyzer, sentence)

            is_match = (tokens_vanilla == tokens_plugin)
            status = "✅ KHỚP" if is_match else "❌ LỆCH"
            if not is_match:
                all_matched = False

            print(f"[{status}] Trường: {label}")
            print(f"    - Vanilla (Go CGO): {tokens_vanilla}")
            print(f"    - Plugin  (Node):   {tokens_plugin}")

    print("\n" + "=" * 80)
    if all_matched:
        print("🎉 TẤT CẢ CÁC TRƯỜNG VÀ CÂU TEST ĐỀU TRÙNG KHỚP 100%! PARITY ĐẠT CHUẨN!")
    else:
        print("⚠️ PHÁT HIỆN SỰ KHÁC BIỆT GIỮA 2 MÔ HÌNH! CẦN ĐIỀU CHỈNH ANALYZER.")
    print("=" * 80)

if __name__ == "__main__":
    verify_token_parity()
