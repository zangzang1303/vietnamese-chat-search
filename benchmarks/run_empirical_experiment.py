#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
run_empirical_experiment.py - Chương trình đo kiểm thực nghiệm hệ thống độc lập:
1. Chạy hoàn toàn độc lập từng Mô hình (Model A: Go CGO vs Model B: Plugin JNI)
2. Đảm bảo điều kiện môi trường lý tưởng:
   - Khi Model A chạy: Mô hình B bị đóng băng (docker pause) để A nhận 100% tài nguyên CPU/RAM/Disk IO.
   - Khi Model B chạy: Mô hình A bị đóng băng (docker pause) để B nhận 100% tài nguyên CPU/RAM/Disk IO.
   - Trước mỗi lần test: Flush disk, drop OS page cache, reset index, cooldown ổn định JVM GC.
3. Hỗ trợ chạy riêng từng mô hình qua cờ lệnh:
   - python benchmarks/run_empirical_experiment.py --model vanilla
   - python benchmarks/run_empirical_experiment.py --model plugin
   - python benchmarks/run_empirical_experiment.py --model search
   - python benchmarks/run_empirical_experiment.py --model aggregate
   - python benchmarks/run_empirical_experiment.py --model all
"""

import sys
import os
import json
import time
import argparse
import subprocess
import urllib.request
import urllib.error

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8')

VANILLA_URL = "http://localhost:9201"
PLUGIN_URL = "http://localhost:9202"
INDEX_NAME = "messages"
VANILLA_CONTAINER = "os_test_foreground"
PLUGIN_CONTAINER = "os_plugin_run"

# =============================================================================
# HỆ THỐNG CÔ LẬP MÔI TRƯỜNG & DỌN DẸP CACHE
# =============================================================================

def drop_system_caches():
    """Xả sạch dirty page cache và sync disk trên WSL Linux để môi trường hoàn toàn sạch."""
    try:
        subprocess.run(
            ["wsl", "-u", "root", "-d", "Ubuntu", "bash", "-c", "sync && echo 3 > /proc/sys/vm/drop_caches"],
            timeout=10, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
    except Exception:
        pass

def pause_container(container_name):
    """Đóng băng container đối thủ để loại bỏ 100% tranh chấp CPU, RAM, IO."""
    try:
        subprocess.run(
            ["wsl", "-u", "root", "-d", "Ubuntu", "docker", "pause", container_name],
            timeout=10, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
    except Exception:
        pass

def unpause_container(container_name):
    """Mở khóa lại container sau khi bài test độc lập kết thúc."""
    try:
        subprocess.run(
            ["wsl", "-u", "root", "-d", "Ubuntu", "docker", "unpause", container_name],
            timeout=10, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
    except Exception:
        pass

def check_node_alive(url, name, retries=5):
    """Kiểm tra xem node OpenSearch đã phản hồi chưa."""
    for _ in range(retries):
        try:
            req = urllib.request.Request(url, method="GET")
            with urllib.request.urlopen(req, timeout=3) as resp:
                if resp.status == 200:
                    return True
        except Exception:
            time.sleep(1)
    return False

def init_index(base_url, index_name, mapping_file):
    """Xóa index cũ và tạo mới mapping."""
    with open(mapping_file, "r", encoding="utf-8") as f:
        mapping_data = json.load(f)

    # 1. Xóa index cũ nếu có
    delete_url = f"{base_url}/{index_name}"
    try:
        req_del = urllib.request.Request(delete_url, method="DELETE")
        urllib.request.urlopen(req_del, timeout=5)
    except Exception:
        pass

    # 2. Tạo index mới
    create_url = f"{base_url}/{index_name}"
    payload = json.dumps(mapping_data).encode("utf-8")
    req_create = urllib.request.Request(create_url, data=payload, headers={"Content-Type": "application/json"}, method="PUT")
    with urllib.request.urlopen(req_create, timeout=10) as resp:
        pass

    # 3. Đợi cluster cập nhật trạng thái
    try:
        health_url = f"{base_url}/_cluster/health/{index_name}?wait_for_status=yellow&timeout=10s"
        req_health = urllib.request.Request(health_url, method="GET")
        with urllib.request.urlopen(req_health, timeout=10) as resp:
            pass
    except Exception:
        pass

def flush_index(base_url, index_name):
    """Flush và commit Lucene segments xuống đĩa."""
    try:
        req_flush = urllib.request.Request(f"{base_url}/{index_name}/_flush", method="POST")
        urllib.request.urlopen(req_flush, timeout=10)
    except Exception:
        pass
    try:
        req_refresh = urllib.request.Request(f"{base_url}/{index_name}/_refresh", method="POST")
        urllib.request.urlopen(req_refresh, timeout=10)
    except Exception:
        pass

# =============================================================================
# THỰC THI BENCHMARK RUNNER (GO)
# =============================================================================

def run_bench_runner(mode, concurrency, max_docs, batch_size, out_file):
    """Gọi binary Go bench_runner.exe để thực hiện nạp dữ liệu chuẩn."""
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

# =============================================================================
# CHẠY ĐỘC LẬP TỪNG MÔ HÌNH (INDEPENDENT SUITES)
# =============================================================================

def run_vanilla_suite(matrix_configs, cooldown_sec=8):
    """Chạy độc lập trọn vẹn toàn bộ ma trận Concurrency cho Mô hình A (Vanilla CGO)."""
    print("\n" + "=" * 80)
    print("🔵 [PHASE A] BẮT ĐẦU ĐO KIỂM ĐỘC LẬP MÔ HÌNH A: GO CGO PRE-TOKENIZATION (:9201)")
    print("   -> Đóng băng container Mô hình B (os_plugin_run) để cách ly 100% tài nguyên!")
    print("=" * 80)

    # Đảm bảo Vanilla unpaused, Plugin paused
    unpause_container(VANILLA_CONTAINER)
    pause_container(PLUGIN_CONTAINER)
    time.sleep(1)

    if not check_node_alive(VANILLA_URL, "Vanilla"):
        print("❌ LỖI: Node Vanilla (:9201) không phản hồi!")
        unpause_container(PLUGIN_CONTAINER)
        return False

    for idx, cfg in enumerate(matrix_configs, 1):
        c = cfg["concurrency"]
        d = cfg["docs"]
        b = cfg["batch"]
        print(f"\n--------------------------------------------------------------------------------")
        print(f"📊 [MÔ HÌNH A - ĐỘC LẬP] BƯỚC {idx}/{len(matrix_configs)}: C={c} Workers | Docs={d:,} | Batch={b}")
        print(f"--------------------------------------------------------------------------------")

        # 1. Dọn dẹp cache OS & reset index
        print("🧹 Làm sạch page cache hệ điều hành và reset index messages...")
        drop_system_caches()
        init_index(VANILLA_URL, INDEX_NAME, "benchmarks/mapping_vanilla.json")
        time.sleep(3)

        # 2. Chạy benchmark
        tmp_vanilla = f"benchmarks/tmp_res_v_c{c}.json"
        res_v = run_bench_runner("vanilla", c, d, b, tmp_vanilla)

        if res_v:
            print(f"\n🎯 [KẾT QUẢ MÔ HÌNH A - C={c}]:")
            print(f"   - Thông lượng Throughput: {res_v['docs_per_sec']:,.1f} docs/s ({res_v['mb_per_sec']:.2f} MB/s)")
            print(f"   - Latency p50: {res_v['latency_p50_ms']:.1f}ms | p95: {res_v['latency_p95_ms']:.1f}ms | p99: {res_v['latency_p99_ms']:.1f}ms")
            print(f"   - JVM GC: {res_v['jvm_gc_count']} lần ({res_v['jvm_gc_time_ms']}ms pause) | Storage: {res_v['disk_store_mb']:.1f} MB")

        # 3. Flush & sync
        flush_index(VANILLA_URL, INDEX_NAME)
        drop_system_caches()

        if idx < len(matrix_configs):
            print(f"⏳ Cooldown {cooldown_sec}s để CPU & Disk I/O trở về trạng thái tĩnh hoàn hảo...")
            time.sleep(cooldown_sec)

    # Giải phóng đóng băng container Plugin
    unpause_container(PLUGIN_CONTAINER)
    print("\n✅ [PHASE A] ĐÃ HOÀN TẤT ĐO KIỂM ĐỘC LẬP MÔ HÌNH A TRÊN TOÀN BỘ CÁC MỨC CONCURRENCY!\n")
    return True

def run_plugin_suite(matrix_configs, cooldown_sec=8):
    """Chạy độc lập trọn vẹn toàn bộ ma trận Concurrency cho Mô hình B (Plugin JNI)."""
    print("\n" + "=" * 80)
    print("🟠 [PHASE B] BẮT ĐẦU ĐO KIỂM ĐỘC LẬP MÔ HÌNH B: OPENSEARCH VIETNAMESE PLUGIN JNI (:9202)")
    print("   -> Đóng băng container Mô hình A (os_test_foreground) để cách ly 100% tài nguyên!")
    print("=" * 80)

    # Đảm bảo Plugin unpaused, Vanilla paused
    unpause_container(PLUGIN_CONTAINER)
    pause_container(VANILLA_CONTAINER)
    time.sleep(1)

    if not check_node_alive(PLUGIN_URL, "Plugin"):
        print("❌ LỖI: Node Plugin (:9202) không phản hồi!")
        unpause_container(VANILLA_CONTAINER)
        return False

    for idx, cfg in enumerate(matrix_configs, 1):
        c = cfg["concurrency"]
        d = cfg["docs"]
        b = cfg["batch"]
        print(f"\n--------------------------------------------------------------------------------")
        print(f"📊 [MÔ HÌNH B - ĐỘC LẬP] BƯỚC {idx}/{len(matrix_configs)}: C={c} Workers | Docs={d:,} | Batch={b}")
        print(f"--------------------------------------------------------------------------------")

        # 1. Dọn dẹp cache OS & reset index
        print("🧹 Làm sạch page cache hệ điều hành và reset index messages...")
        drop_system_caches()
        init_index(PLUGIN_URL, INDEX_NAME, "benchmarks/mapping_plugin.json")
        time.sleep(3)

        # 2. Chạy benchmark
        tmp_plugin = f"benchmarks/tmp_res_p_c{c}.json"
        res_p = run_bench_runner("plugin", c, d, b, tmp_plugin)

        if res_p:
            print(f"\n🎯 [KẾT QUẢ MÔ HÌNH B - C={c}]:")
            print(f"   - Thông lượng Throughput: {res_p['docs_per_sec']:,.1f} docs/s ({res_p['mb_per_sec']:.2f} MB/s)")
            print(f"   - Latency p50: {res_p['latency_p50_ms']:.1f}ms | p95: {res_p['latency_p95_ms']:.1f}ms | p99: {res_p['latency_p99_ms']:.1f}ms")
            print(f"   - JVM GC: {res_p['jvm_gc_count']} lần ({res_p['jvm_gc_time_ms']}ms pause) | Storage: {res_p['disk_store_mb']:.1f} MB")

        # 3. Flush & sync
        flush_index(PLUGIN_URL, INDEX_NAME)
        drop_system_caches()

        if idx < len(matrix_configs):
            print(f"⏳ Cooldown {cooldown_sec}s để CPU & Disk I/O trở về trạng thái tĩnh hoàn hảo...")
            time.sleep(cooldown_sec)

    # Giải phóng đóng băng container Vanilla
    unpause_container(VANILLA_CONTAINER)
    print("\n✅ [PHASE B] ĐÃ HOÀN TẤT ĐO KIỂM ĐỘC LẬP MÔ HÌNH B TRÊN TOÀN BỘ CÁC MỨC CONCURRENCY!\n")
    return True

# =============================================================================
# KIỂM THỬ SEARCH PARITY & QUERY LATENCY
# =============================================================================

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

def run_search_parity_suite():
    print("\n" + "=" * 80)
    print("🔍 [PHASE C] KIỂM THỬ SEARCH PARITY & QUERY LATENCY (520 QUERIES THỰC TẾ)")
    print("=" * 80)

    unpause_container(VANILLA_CONTAINER)
    unpause_container(PLUGIN_CONTAINER)
    time.sleep(2)

    if not check_node_alive(VANILLA_URL, "Vanilla") or not check_node_alive(PLUGIN_URL, "Plugin"):
        print("❌ LỖI: Một trong 2 node chưa sẵn sàng cho bài test Search!")
        return None

    test_queries = [
        {"q": "vật liệu", "cat": "compound_accented"},
        {"q": "bộ xử lý", "cat": "compound_accented"},
        {"q": "chất kết dính", "cat": "compound_accented"},
        {"q": "thị trường", "cat": "compound_accented"},
        {"q": "giá thành", "cat": "compound_accented"},
        {"q": "sản phẩm", "cat": "compound_accented"},
        {"q": "điện thoại", "cat": "compound_accented"},
        {"q": "tản nhiệt", "cat": "compound_accented"},
        {"q": "vat lieu", "cat": "unaccented"},
        {"q": "bo xu ly", "cat": "unaccented"},
        {"q": "chat ket dinh", "cat": "unaccented"},
        {"q": "thi truong", "cat": "unaccented"},
        {"q": "gia thanh", "cat": "unaccented"},
        {"q": "san pham", "cat": "unaccented"},
        {"q": "dien thoai", "cat": "unaccented"},
        {"q": "tan nhiet", "cat": "unaccented"},
        {"q": "vat_l", "cat": "edge_prefix"},
        {"q": "bo_x", "cat": "edge_prefix"},
        {"q": "thi_t", "cat": "edge_prefix"},
        {"q": "gia_t", "cat": "edge_prefix"},
        {"q": "san_p", "cat": "edge_prefix"},
        {"q": "dien_t", "cat": "edge_prefix"},
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
            v_lat, v_hits, v_ids = run_search_query(VANILLA_URL, q, is_vanilla=True)
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
                print(f"\r   ⏳ [Search Test] {q_pct:5.1f}% | {q_done}/{total_q_count} queries | Đã chạy: {elapsed_s:.1f}s | Còn: {eta_s:.1f}s", end="", flush=True)

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

    return search_summary

# =============================================================================
# TỔNG HỢP KẾT QUẢ & XUẤT BIỂU ĐỒ, BÁO CÁO
# =============================================================================

def aggregate_and_render(matrix_configs, search_summary=None):
    print("\n" + "=" * 80)
    print("📊 [PHASE D] TỔNG HỢP TOÀN BỘ KẾT QUẢ THỰC NGHIỆM ĐỘC LẬP")
    print("=" * 80)

    # Đảm bảo thư mục image tồn tại
    os.makedirs("image", exist_ok=True)

    # Đọc kết quả search cũ nếu không truyền search_summary mới
    if search_summary is None and os.path.exists("benchmarks/full_matrix_results.json"):
        try:
            with open("benchmarks/full_matrix_results.json", "r", encoding="utf-8") as f:
                old = json.load(f)
                search_summary = old.get("search_benchmark", {})
        except Exception:
            search_summary = {}

    matrix_list = []
    for cfg in matrix_configs:
        c = cfg["concurrency"]
        d = cfg["docs"]
        b = cfg["batch"]
        file_v = f"benchmarks/tmp_res_v_c{c}.json"
        file_p = f"benchmarks/tmp_res_p_c{c}.json"

        if not os.path.exists(file_v) or not os.path.exists(file_p):
            print(f"⚠️ Chưa đủ dữ liệu đối soát cho C={c} (tìm {file_v} và {file_p})")
            continue

        with open(file_v, "r", encoding="utf-8") as f:
            v_arr = json.load(f)
            res_v = v_arr[0] if v_arr else None
        with open(file_p, "r", encoding="utf-8") as f:
            p_arr = json.load(f)
            res_p = p_arr[0] if p_arr else None

        if res_v and res_p:
            tps_diff = ((res_v["docs_per_sec"] - res_p["docs_per_sec"]) / res_p["docs_per_sec"]) * 100.0
            p50_diff = ((res_p["latency_p50_ms"] - res_v["latency_p50_ms"]) / res_v["latency_p50_ms"]) * 100.0
            gc_diff = ((res_p["jvm_gc_time_ms"] - res_v["jvm_gc_time_ms"]) / (res_v["jvm_gc_time_ms"] if res_v["jvm_gc_time_ms"] > 0 else 1)) * 100.0
            matrix_list.append({
                "concurrency": c,
                "docs": d,
                "batch": b,
                "vanilla": res_v,
                "plugin": res_p,
                "throughput_diff_percent": tps_diff,
                "p50_latency_diff_percent": p50_diff,
                "gc_time_diff_percent": gc_diff
            })
            print(f"Workers {c:2d} | Docs: {d:,} | Vanilla: {res_v['docs_per_sec']:9,.1f} d/s (p50: {res_v['latency_p50_ms']:5.1f}ms) | Plugin: {res_p['docs_per_sec']:9,.1f} d/s (p50: {res_p['latency_p50_ms']:5.1f}ms) | Chênh lệch: {tps_diff:+6.1f}%")

    all_results = {
        "concurrency_matrix": matrix_list,
        "search_benchmark": search_summary if search_summary else {},
        "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
    }

    with open("benchmarks/full_matrix_results.json", "w", encoding="utf-8") as f:
        json.dump(all_results, f, indent=2, ensure_ascii=False)
    print("\n💾 Đã lưu ma trận số liệu vào benchmarks/full_matrix_results.json")

    # Vẽ biểu đồ bằng matplotlib (thử py -3.13 trước vì có cài sẵn matplotlib)
    print("\n📈 Đang tự động vẽ 5 biểu đồ trực quan hóa...")
    chart_rendered = False
    for py_cmd in [["py", "-3.13", "benchmarks/generate_empirical_charts.py"], [sys.executable, "benchmarks/generate_empirical_charts.py"]]:
        try:
            p = subprocess.run(py_cmd, capture_output=True, text=True)
            if p.returncode == 0:
                print("🎉 Đã xuất thành công 5 biểu đồ độ phân giải cao vào thư mục image/")
                chart_rendered = True
                break
        except Exception:
            pass

    if not chart_rendered:
        print("⚠️ Không thể tự động chạy generate_empirical_charts.py, vui lòng chạy thủ công bằng 'py -3.13 benchmarks/generate_empirical_charts.py'")

    # Cập nhật báo cáo markdown
    print("\n📝 Đang cập nhật báo cáo docs/benchmark_opensearch_plugin_report.md...")
    try:
        subprocess.run([sys.executable, "benchmarks/update_report_from_matrix.py"])
    except Exception as e:
        print(f"⚠️ Cập nhật báo cáo: {e}")

# =============================================================================
# MAIN ENTRY POINT
# =============================================================================

def main():
    parser = argparse.ArgumentParser(description="Chương trình đo kiểm thực nghiệm hệ thống độc lập OpenSearch")
    parser.add_argument("--model", choices=["all", "vanilla", "plugin", "search", "aggregate"], default="all",
                        help="Chọn chế độ chạy: 'vanilla' (chỉ Model A độc lập), 'plugin' (chỉ Model B độc lập), 'search' (chỉ Search test), 'aggregate' (tổng hợp & vẽ biểu đồ), 'all' (chạy tuần tự toàn bộ độc lập)")
    parser.add_argument("--docs", type=int, default=100000, help="Số lượng docs nạp cho mỗi mức concurrency (mặc định 100,000)")
    parser.add_argument("--batch", type=int, default=1000, help="Batch size mỗi request bulk (mặc định 1,000)")
    parser.add_argument("--cooldown", type=int, default=8, help="Thời gian nghỉ (giây) giữa các mức concurrency (mặc định 8s)")
    args = parser.parse_args()

    matrix_configs = [
        {"concurrency": 1,  "docs": args.docs, "batch": args.batch},
        {"concurrency": 4,  "docs": args.docs, "batch": args.batch},
        {"concurrency": 8,  "docs": args.docs, "batch": args.batch},
        {"concurrency": 16, "docs": args.docs, "batch": args.batch},
        {"concurrency": 32, "docs": args.docs, "batch": args.batch},
    ]

    print("=" * 80)
    print("KHỞI ĐỘNG HỆ THỐNG ĐO KIỂM THỰC NGHIỆM ĐỘC LẬP & TỐI ƯU HÓA MÔI TRƯỜNG")
    print(f"Chế độ chỉ định: --model={args.model} | Docs/mức: {args.docs:,} | Batch: {args.batch} | Cooldown: {args.cooldown}s")
    print("=" * 80)

    # Đảm bảo ban đầu cả 2 container đều unpaused để kiểm tra
    unpause_container(VANILLA_CONTAINER)
    unpause_container(PLUGIN_CONTAINER)

    if args.model == "vanilla":
        run_vanilla_suite(matrix_configs, args.cooldown)
    elif args.model == "plugin":
        run_plugin_suite(matrix_configs, args.cooldown)
    elif args.model == "search":
        run_search_parity_suite()
    elif args.model == "aggregate":
        aggregate_and_render(matrix_configs)
    elif args.model == "all":
        # 1. Chạy hoàn toàn độc lập Mô hình A
        success_a = run_vanilla_suite(matrix_configs, args.cooldown)
        if not success_a:
            print("❌ Dừng bài test do Model A gặp lỗi!")
            return

        print("\n" + "#" * 80)
        print("⏸️ NGHỈ GIỮA 2 MÔ HÌNH (15 GIÂY) ĐỂ HỆ THỐNG GIẢI PHÓNG BỘ NHỚ VÀ COOLDOWN HOÀN TOÀN")
        print("#" * 80)
        drop_system_caches()
        time.sleep(15)

        # 2. Chạy hoàn toàn độc lập Mô hình B
        success_b = run_plugin_suite(matrix_configs, args.cooldown)
        if not success_b:
            print("❌ Dừng bài test do Model B gặp lỗi!")
            return

        print("\n" + "#" * 80)
        print("⏸️ NGHỈ 5 GIÂY TRƯỚC KHI BƯỚC VÀO ĐO KIỂM SEARCH PARITY")
        print("#" * 80)
        drop_system_caches()
        time.sleep(5)

        # 3. Chạy Search Parity & Query Latency
        search_res = run_search_parity_suite()

        # 4. Tổng hợp & Xuất biểu đồ + Báo cáo
        aggregate_and_render(matrix_configs, search_res)

if __name__ == "__main__":
    main()
