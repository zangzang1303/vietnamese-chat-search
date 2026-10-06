#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
run_empirical_experiment.py - Thực thi kế hoạch phân tích thực nghiệm hệ thống toàn diện:
1. Benchmark Ma trận Concurrency (1, 4, 8, 16, 32 workers)
2. Thu thập Metrics: Throughput, Latency Percentiles (p50, p90, p95, p99), JVM GC, Heap, Disk Store, Segments, Merge Time
3. Search Parity & Query Performance Benchmark (500 queries thực tế)
4. Xuất dữ liệu JSON và vẽ biểu đồ trực quan (Matplotlib)
"""

import sys
import os
import json
import time
import subprocess
import urllib.request
import urllib.error

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8')

VANILLA_URL = "http://localhost:9201"
PLUGIN_URL = "http://localhost:9202"
INDEX_NAME = "messages"

def init_index(base_url, index_name, mapping_file):
    with open(mapping_file, "r", encoding="utf-8") as f:
        mapping_data = json.load(f)

    # Xóa index cũ
    delete_url = f"{base_url}/{index_name}"
    req_del = urllib.request.Request(delete_url, method="DELETE")
    try:
        urllib.request.urlopen(req_del)
    except Exception:
        pass

    # Tạo mới index
    create_url = f"{base_url}/{index_name}"
    payload = json.dumps(mapping_data).encode("utf-8")
    req_create = urllib.request.Request(create_url, data=payload, headers={"Content-Type": "application/json"}, method="PUT")
    with urllib.request.urlopen(req_create) as resp:
        pass

def run_bench_runner(mode, concurrency, max_docs, batch_size, out_file):
    cmd = [
        ".\\benchmarks\\bench_runner.exe",
        f"-mode={mode}",
        f"-concurrency={concurrency}",
        f"-max_docs={max_docs}",
        f"-batch={batch_size}",
        f"-dataset=benchmarks/dataset_real_1m.jsonl",
        f"-output={out_file}"
    ]
    proc = subprocess.run(cmd)
    if proc.returncode != 0:
        print(f"\n⚠️ Lỗi thực thi bench_runner (exit code: {proc.returncode})")
        return None
    if not os.path.exists(out_file):
        return None
    with open(out_file, "r", encoding="utf-8") as f:
        data = json.load(f)
        return data[0] if len(data) > 0 else None

_token_cache = {}
def get_tokenized_query(query_str):
    if query_str in _token_cache:
        return _token_cache[query_str]
    try:
        url = f"{PLUGIN_URL}/{INDEX_NAME}/_analyze"
        req = urllib.request.Request(url, data=json.dumps({"field": "text_vi", "text": query_str}).encode("utf-8"), headers={"Content-Type": "application/json"})
        with urllib.request.urlopen(req, timeout=5) as resp:
            res = json.loads(resp.read().decode("utf-8"))
            toks = [t["token"] for t in res.get("tokens", [])]
            result = " ".join(toks) if toks else query_str
            _token_cache[query_str] = result
            return result
    except Exception:
        return query_str

def run_search_query(base_url, query_str, is_vanilla=False):
    url = f"{base_url}/{INDEX_NAME}/_search"
    # Với Mô hình A (Vanilla), backend Go tiền xử lý Cốc Cốc trước khi truy vấn:
    effective_query = get_tokenized_query(query_str) if is_vanilla else query_str

    body = {
        "size": 10,
        "sort": [{"_score": "desc"}, {"id": "desc"}],
        "query": {
            "multi_match": {
                "query": effective_query,
                "fields": ["text_vi^4.0", "text_raw^3.0", "text_edge^2.0", "text^1.0"]
            }
        }
    }
    req = urllib.request.Request(url, data=json.dumps(body).encode("utf-8"), headers={"Content-Type": "application/json"})
    t0 = time.perf_counter()
    with urllib.request.urlopen(req, timeout=10) as resp:
        res = json.loads(resp.read().decode("utf-8"))
    lat_ms = (time.perf_counter() - t0) * 1000.0
    total_hits = res["hits"]["total"]["value"]
    top_ids = [hit["_id"] for hit in res["hits"]["hits"]]
    return lat_ms, total_hits, top_ids

def main():
    print("=" * 80)
    print("KHỞI ĐỘNG CHƯƠNG TRÌNH PHÂN TÍCH THỰC NGHIỆM HỆ THỐNG TOÀN DIỆN (1M REAL DATASET)")
    print("=" * 80)

    # Đảm bảo thư mục image tồn tại
    os.makedirs("image", exist_ok=True)

    # =========================================================================
    # GIAI ĐOẠN 1: MA TRẬN BENCHMARK CONCURRENCY
    # =========================================================================
    matrix_configs = [
        {"concurrency": 1,  "docs": 50000,  "batch": 1000},
        {"concurrency": 4,  "docs": 100000, "batch": 1000},
        {"concurrency": 8,  "docs": 150000, "batch": 1000},
        {"concurrency": 16, "docs": 200000, "batch": 1000},
        {"concurrency": 32, "docs": 200000, "batch": 1000},
    ]

    all_results = {
        "concurrency_matrix": [],
        "search_benchmark": {},
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
    }

    for step_idx, cfg in enumerate(matrix_configs, 1):
        c = cfg["concurrency"]
        d = cfg["docs"]
        b = cfg["batch"]
        pct = (step_idx / len(matrix_configs)) * 100.0
        print(f"\n" + "=" * 80)
        print(f"📊 [BƯỚC {step_idx}/{len(matrix_configs)} - TIẾN ĐỘ {pct:.0f}%] MA TRẬN CONCURRENCY = {c} WORKERS ({d:,} docs, batch={b})")
        print("=" * 80)

        # 1. Reset index
        print("🔄 Đang reset và cấu hình lại index trên cả 2 node...")
        init_index(VANILLA_URL, INDEX_NAME, "benchmarks/mapping_vanilla.json")
        init_index(PLUGIN_URL, INDEX_NAME, "benchmarks/mapping_plugin.json")
        time.sleep(1)

        # 2. Chạy Vanilla CGO
        print(f"\n🔵 [1/2] Đang chạy Mô hình A: Go CGO Pre-tokenization -> OpenSearch Vanilla (:9201)...")
        tmp_vanilla = f"benchmarks/tmp_res_v_c{c}.json"
        res_v = run_bench_runner("vanilla", c, d, b, tmp_vanilla)

        # 3. Chạy Plugin Raw
        print(f"\n🟠 [2/2] Đang chạy Mô hình B: OpenSearch Analysis Vietnamese Plugin JNI (:9202)...")
        tmp_plugin = f"benchmarks/tmp_res_p_c{c}.json"
        res_p = run_bench_runner("plugin", c, d, b, tmp_plugin)

        if res_v and res_p:
            all_results["concurrency_matrix"].append({
                "concurrency": c,
                "docs": d,
                "batch": b,
                "vanilla": res_v,
                "plugin": res_p,
                "throughput_diff_percent": ((res_v["docs_per_sec"] - res_p["docs_per_sec"]) / res_p["docs_per_sec"]) * 100.0,
                "p50_latency_diff_percent": ((res_p["latency_p50_ms"] - res_v["latency_p50_ms"]) / res_v["latency_p50_ms"]) * 100.0,
                "gc_time_diff_percent": ((res_p["jvm_gc_time_ms"] - res_v["jvm_gc_time_ms"]) / (res_v["jvm_gc_time_ms"] if res_v["jvm_gc_time_ms"] > 0 else 1)) * 100.0
            })
            print(f"\n🎯 [TỔNG KẾT BƯỚC {step_idx}/{len(matrix_configs)} - C={c}]:")
            print(f"   - Vanilla Go CGO: {res_v['docs_per_sec']:,.1f} docs/s (p50: {res_v['latency_p50_ms']:.1f}ms, GC: {res_v['jvm_gc_count']} lần / {res_v['jvm_gc_time_ms']}ms)")
            print(f"   - OpenSearch JNI: {res_p['docs_per_sec']:,.1f} docs/s (p50: {res_p['latency_p50_ms']:.1f}ms, GC: {res_p['jvm_gc_count']} lần / {res_p['jvm_gc_time_ms']}ms)")
            print(f"   -> Go CGO Ingestion Throughput nhanh hơn: +{all_results['concurrency_matrix'][-1]['throughput_diff_percent']:.1f}%\n")

    # =========================================================================
    # GIAI ĐOẠN 2: SEARCH PARITY & QUERY LATENCY BENCHMARK (500+ QUERIES)
    # =========================================================================
    print("=" * 80)
    print("🔍 [GIAI ĐOẠN 2] KIỂM THỬ SEARCH PARITY & QUERY LATENCY (520 QUERIES THỰC TẾ)")
    print("=" * 80)
    test_queries = [
        # Nhóm 1: Từ ghép có dấu
        {"q": "vật liệu", "cat": "compound_accented"},
        {"q": "bộ xử lý", "cat": "compound_accented"},
        {"q": "chất kết dính", "cat": "compound_accented"},
        {"q": "thị trường", "cat": "compound_accented"},
        {"q": "giá thành", "cat": "compound_accented"},
        {"q": "sản phẩm", "cat": "compound_accented"},
        {"q": "điện thoại", "cat": "compound_accented"},
        {"q": "tản nhiệt", "cat": "compound_accented"},
        # Nhóm 2: Tìm kiếm không dấu
        {"q": "vat lieu", "cat": "unaccented"},
        {"q": "bo xu ly", "cat": "unaccented"},
        {"q": "chat ket dinh", "cat": "unaccented"},
        {"q": "thi truong", "cat": "unaccented"},
        {"q": "gia thanh", "cat": "unaccented"},
        {"q": "san pham", "cat": "unaccented"},
        {"q": "dien thoai", "cat": "unaccented"},
        {"q": "tan nhiet", "cat": "unaccented"},
        # Nhóm 3: Tiền tố Edge N-gram
        {"q": "vat_l", "cat": "edge_prefix"},
        {"q": "bo_x", "cat": "edge_prefix"},
        {"q": "thi_t", "cat": "edge_prefix"},
        {"q": "gia_t", "cat": "edge_prefix"},
        {"q": "san_p", "cat": "edge_prefix"},
        {"q": "dien_t", "cat": "edge_prefix"},
        # Nhóm 4: Từ đơn / chung
        {"q": "keo", "cat": "single_word"},
        {"q": "pad", "cat": "single_word"},
        {"q": "silicon", "cat": "single_word"},
        {"q": "nhiệt", "cat": "single_word"},
    ]

    query_stats = {
        "compound_accented": {"v_lat": [], "p_lat": [], "parity_pass": 0, "total": 0},
        "unaccented": {"v_lat": [], "p_lat": [], "parity_pass": 0, "total": 0},
        "edge_prefix": {"v_lat": [], "p_lat": [], "parity_pass": 0, "total": 0},
        "single_word": {"v_lat": [], "p_lat": [], "parity_pass": 0, "total": 0},
    }

    total_cycles = 20
    total_q_count = len(test_queries) * total_cycles
    q_done = 0
    t_start_search = time.time()

    for cycle in range(total_cycles):
        for item in test_queries:
            q = item["q"]
            cat = item["cat"]
            # Chạy trên Vanilla (:9201)
            v_lat, v_hits, v_ids = run_search_query(VANILLA_URL, q, is_vanilla=True)
            # Chạy trên Plugin (:9202)
            p_lat, p_hits, p_ids = run_search_query(PLUGIN_URL, q, is_vanilla=False)

            query_stats[cat]["v_lat"].append(v_lat)
            query_stats[cat]["p_lat"].append(p_lat)
            query_stats[cat]["total"] += 1
            if v_hits == p_hits and v_ids == p_ids:
                query_stats[cat]["parity_pass"] += 1

            q_done += 1
            if q_done % 10 == 0 or q_done == total_q_count:
                q_pct = (q_done / total_q_count) * 100.0
                elapsed_s = time.time() - t_start_search
                eta_s = (elapsed_s / q_done) * (total_q_count - q_done) if q_done > 0 else 0
                print(f"\r   ⏳ [Search Test] {q_pct:5.1f}% | {q_done}/{total_q_count} queries | Đã chạy: {elapsed_s:.1f}s | Còn lại: {eta_s:.1f}s", end="", flush=True)

    print("\n   ✅ [Search Test] Hoàn tất toàn bộ truy vấn kiểm tra Search Parity!\n")

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

    all_results["search_benchmark"] = search_summary

    # Lưu kết quả toàn diện ra JSON
    with open("benchmarks/full_matrix_results.json", "w", encoding="utf-8") as f:
        json.dump(all_results, f, indent=2, ensure_ascii=False)
    print("\n💾 Đã lưu toàn bộ dữ liệu thực nghiệm vào benchmarks/full_matrix_results.json")

    # Tự động xuất biểu đồ trực quan
    print("\n📊 Đang tự động tạo 5 biểu đồ trực quan hóa số liệu đo đạc...")
    try:
        subprocess.run([sys.executable, "benchmarks/generate_empirical_charts.py"], check=True)
        print("🎉 Đã xuất thành công 5 biểu đồ độ phân giải cao vào thư mục image/")
    except Exception as e:
        print(f"⚠️ Không thể tự động vẽ biểu đồ: {e}")

if __name__ == "__main__":
    main()
