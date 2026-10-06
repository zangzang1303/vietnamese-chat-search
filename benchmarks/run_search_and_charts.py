#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import sys
import os
sys.path.insert(0, ".")
import json
import time
import subprocess
from benchmarks.run_empirical_experiment import run_search_query, VANILLA_URL, PLUGIN_URL

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8')

def main():
    print("=" * 80)
    print("CHẠY BENCHMARK SEARCH PARITY & QUERY LATENCY (500 QUERIES)")
    print("=" * 80)

    test_queries = [
        {"q": "học sinh", "cat": "compound_accented"},
        {"q": "cà phê", "cat": "compound_accented"},
        {"q": "việt nam", "cat": "compound_accented"},
        {"q": "văn phòng", "cat": "compound_accented"},
        {"q": "thời khóa biểu", "cat": "compound_accented"},
        {"q": "sinh viên", "cat": "compound_accented"},
        {"q": "máy lạnh", "cat": "compound_accented"},
        {"q": "công nghệ thông tin", "cat": "compound_accented"},
        {"q": "hoc sinh", "cat": "unaccented"},
        {"q": "ca phe", "cat": "unaccented"},
        {"q": "viet nam", "cat": "unaccented"},
        {"q": "van phong", "cat": "unaccented"},
        {"q": "thoi khoa bieu", "cat": "unaccented"},
        {"q": "sinh vien", "cat": "unaccented"},
        {"q": "vie", "cat": "edge_prefix"},
        {"q": "viet_n", "cat": "edge_prefix"},
        {"q": "hoc_s", "cat": "edge_prefix"},
        {"q": "van_p", "cat": "edge_prefix"},
        {"q": "ca_p", "cat": "edge_prefix"},
        {"q": "chào", "cat": "single_word"},
    ]

    query_stats = {
        "compound_accented": {"v_lat": [], "p_lat": [], "parity_pass": 0, "total": 0},
        "unaccented": {"v_lat": [], "p_lat": [], "parity_pass": 0, "total": 0},
        "edge_prefix": {"v_lat": [], "p_lat": [], "parity_pass": 0, "total": 0},
        "single_word": {"v_lat": [], "p_lat": [], "parity_pass": 0, "total": 0},
    }

    for cycle in range(25):
        for item in test_queries:
            q = item["q"]
            cat = item["cat"]
            v_lat, v_hits, v_ids = run_search_query(VANILLA_URL, q, is_vanilla=True)
            p_lat, p_hits, p_ids = run_search_query(PLUGIN_URL, q, is_vanilla=False)

            query_stats[cat]["v_lat"].append(v_lat)
            query_stats[cat]["p_lat"].append(p_lat)
            query_stats[cat]["total"] += 1
            if v_hits == p_hits and v_ids == p_ids:
                query_stats[cat]["parity_pass"] += 1

    search_summary = {}
    for cat, d in query_stats.items():
        v_lats = sorted(d["v_lat"])
        p_lats = sorted(d["p_lat"])
        search_summary[cat] = {
            "total_queries": d["total"],
            "parity_match_percent": (d["parity_pass"] / d["total"]) * 100.0,
            "vanilla_latency_p50_ms": v_lats[int(len(v_lats)*0.5)],
            "vanilla_latency_p95_ms": v_lats[int(len(v_lats)*0.95)],
            "plugin_latency_p50_ms": p_lats[int(len(p_lats)*0.5)],
            "plugin_latency_p95_ms": p_lats[int(len(p_lats)*0.95)],
        }
        print(f"   [SEARCH PARITY] Thể loại '{cat}': Khớp {search_summary[cat]['parity_match_percent']:.1f}% | Latency p50: Vanilla={search_summary[cat]['vanilla_latency_p50_ms']:.2f}ms vs Plugin={search_summary[cat]['plugin_latency_p50_ms']:.2f}ms")

    with open("benchmarks/full_matrix_results.json", "r", encoding="utf-8") as f:
        full_data = json.load(f)

    full_data["search_benchmark"] = search_summary

    with open("benchmarks/full_matrix_results.json", "w", encoding="utf-8") as f:
        json.dump(full_data, f, indent=2, ensure_ascii=False)

    print("\n📊 Đang sinh biểu đồ trực quan hóa dữ liệu bằng Matplotlib...")
    subprocess.run(["py", "-3.13", "benchmarks/generate_empirical_charts.py"], check=True)
    print("✨ Hoàn tất toàn bộ quy trình đo đạc và trực quan hóa!")

if __name__ == "__main__":
    main()
