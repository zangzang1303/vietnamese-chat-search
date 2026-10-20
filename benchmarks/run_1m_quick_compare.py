#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
run_1m_quick_compare.py - Thực thi đo kiểm 1.000.000 documents tại mức concurrency tối ưu (16 workers)
So sánh định lượng giữa:
- Mô hình A: Go CGO Pre-tokenization (OpenSearch Vanilla :9201)
- Mô hình B: OpenSearch Plugin JNI (OpenSearch Plugin :9202)
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
VANILLA_CONTAINER = "os_vanilla_bench"
PLUGIN_CONTAINER = "os_plugin_bench"
NUM_DOCS = 1000000
CONCURRENCY = 16
BATCH_SIZE = 1000
DATASET_PATH = "benchmarks/dataset_real_1m.jsonl"

def drop_system_caches():
    try:
        subprocess.run(
            ["wsl", "-u", "root", "-d", "Ubuntu", "bash", "-c", "sync && echo 3 > /proc/sys/vm/drop_caches"],
            timeout=10, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
    except Exception:
        pass

def pause_container(container_name):
    try:
        subprocess.run(
            ["wsl", "-u", "root", "-d", "Ubuntu", "docker", "pause", container_name],
            timeout=10, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
    except Exception:
        pass

def unpause_container(container_name):
    try:
        subprocess.run(
            ["wsl", "-u", "root", "-d", "Ubuntu", "docker", "unpause", container_name],
            timeout=10, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL
        )
    except Exception:
        pass

def check_node_alive(url, name, retries=5):
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
    with open(mapping_file, "r", encoding="utf-8") as f:
        mapping_data = json.load(f)

    delete_url = f"{base_url}/{index_name}"
    try:
        req_del = urllib.request.Request(delete_url, method="DELETE")
        urllib.request.urlopen(req_del, timeout=5)
    except Exception:
        pass

    create_url = f"{base_url}/{index_name}"
    payload = json.dumps(mapping_data).encode("utf-8")
    req_create = urllib.request.Request(create_url, data=payload, headers={"Content-Type": "application/json"}, method="PUT")
    with urllib.request.urlopen(req_create, timeout=10) as resp:
        pass

    try:
        health_url = f"{base_url}/_cluster/health/{index_name}?wait_for_status=yellow&timeout=10s"
        req_health = urllib.request.Request(health_url, method="GET")
        with urllib.request.urlopen(req_health, timeout=10) as resp:
            pass
    except Exception:
        pass

def flush_index(base_url, index_name):
    try:
        req_flush = urllib.request.Request(f"{base_url}/{index_name}/_flush", method="POST")
        urllib.request.urlopen(req_flush, timeout=15)
    except Exception:
        pass
    try:
        req_refresh = urllib.request.Request(f"{base_url}/{index_name}/_refresh", method="POST")
        urllib.request.urlopen(req_refresh, timeout=15)
    except Exception:
        pass

def run_bench(mode, concurrency, max_docs, batch_size, out_file):
    cmd = [
        ".\\benchmarks\\bench_runner.exe",
        f"-mode={mode}",
        f"-concurrency={concurrency}",
        f"-max_docs={max_docs}",
        f"-batch={batch_size}",
        f"-dataset={DATASET_PATH}",
        f"-output={out_file}"
    ]
    proc = subprocess.run(cmd)
    if proc.returncode != 0:
        print(f"❌ Lỗi thực thi bench_runner (exit code: {proc.returncode})")
        return None
    if not os.path.exists(out_file):
        return None
    with open(out_file, "r", encoding="utf-8") as f:
        data = json.load(f)
        return data[0] if len(data) > 0 else None

def main():
    print("=" * 80)
    print("🔥 KHỞI CHẠY BENCHMARK 1.000.000 DOCUMENTS TRÊN OPENSEARCH 2.19")
    print(f"Khối lượng nạp: {NUM_DOCS:,} docs | Concurrency: {CONCURRENCY} workers | Batch: {BATCH_SIZE}")
    print(f"Dataset: {DATASET_PATH}")
    print("=" * 80)

    # Đảm bảo unpause ban đầu
    unpause_container(VANILLA_CONTAINER)
    unpause_container(PLUGIN_CONTAINER)
    time.sleep(1)

    if not check_node_alive(VANILLA_URL, "Vanilla"):
        print("❌ Node Vanilla (:9201) không phản hồi!")
        return
    if not check_node_alive(PLUGIN_URL, "Plugin"):
        print("❌ Node Plugin (:9202) không phản hồi!")
        return

    # =========================================================================
    # 1. CHẠY MÔ HÌNH A: GO CGO PRE-TOKENIZATION
    # =========================================================================
    print("\n" + "-" * 80)
    print("🔵 [BƯỚC 1/2] THỰC NGHIỆM MÔ HÌNH A: GO CGO PRE-TOKENIZATION (:9201)")
    print("   -> Đóng băng container Mô hình B để cách ly 100% tài nguyên CPU/RAM/Disk IO!")
    print("-" * 80)

    pause_container(PLUGIN_CONTAINER)
    drop_system_caches()
    print("🧹 Khởi tạo lại index 'messages' với Vanilla Mapping...")
    init_index(VANILLA_URL, INDEX_NAME, "benchmarks/mapping_vanilla.json")
    time.sleep(3)

    out_vanilla = "benchmarks/res_1m_vanilla.json"
    res_v = run_bench("vanilla", CONCURRENCY, NUM_DOCS, BATCH_SIZE, out_vanilla)

    print("\n🧹 Xả dirty cache và giải phóng bộ đệm...")
    flush_index(VANILLA_URL, INDEX_NAME)
    drop_system_caches()
    unpause_container(PLUGIN_CONTAINER)

    print("\n⏳ Cooldown 10s trước khi chuyển sang Mô hình B...")
    time.sleep(10)

    # =========================================================================
    # 2. CHẠY MÔ HÌNH B: OPENSEARCH PLUGIN JNI
    # =========================================================================
    print("\n" + "-" * 80)
    print("🟠 [BƯỚC 2/2] THỰC NGHIỆM MÔ HÌNH B: OPENSEARCH PLUGIN JNI (:9202)")
    print("   -> Đóng băng container Mô hình A để cách ly 100% tài nguyên CPU/RAM/Disk IO!")
    print("-" * 80)

    pause_container(VANILLA_CONTAINER)
    drop_system_caches()
    print("🧹 Khởi tạo lại index 'messages' với Plugin Mapping...")
    init_index(PLUGIN_URL, INDEX_NAME, "benchmarks/mapping_plugin.json")
    time.sleep(3)

    out_plugin = "benchmarks/res_1m_plugin.json"
    res_p = run_bench("plugin", CONCURRENCY, NUM_DOCS, BATCH_SIZE, out_plugin)

    print("\n🧹 Xả dirty cache và giải phóng bộ đệm...")
    flush_index(PLUGIN_URL, INDEX_NAME)
    drop_system_caches()
    unpause_container(VANILLA_CONTAINER)

    # =========================================================================
    # 3. BẢNG SO SÁNH ĐỐI SOÁT TỔNG HỢP (SUMMARY REPORT)
    # =========================================================================
    if res_v and res_p:
        tps_diff = ((res_v["docs_per_sec"] - res_p["docs_per_sec"]) / res_p["docs_per_sec"]) * 100.0
        p50_diff = ((res_v["latency_p50_ms"] - res_p["latency_p50_ms"]) / res_p["latency_p50_ms"]) * 100.0
        p95_diff = ((res_v["latency_p95_ms"] - res_p["latency_p95_ms"]) / res_p["latency_p95_ms"]) * 100.0
        gc_pause_diff = ((res_v["jvm_gc_time_ms"] - res_p["jvm_gc_time_ms"]) / (res_p["jvm_gc_time_ms"] if res_p["jvm_gc_time_ms"] > 0 else 1)) * 100.0

        print("\n" + "=" * 85)
        print("🏆 BẢNG TỔNG HỢP SO SÁNH HIỆU NĂNG THỰC NGHIỆM 1.000.000 DOCUMENTS (C=16)")
        print("=" * 85)
        print(f"{'Chỉ số Đo lường':<30} | {'Mô hình A (Go CGO)':<20} | {'Mô hình B (Plugin)':<20} | {'Chênh lệch %':<12}")
        print("-" * 85)
        print(f"{'Số docs hoàn tất':<30} | {res_v['total_docs']:<20,d} | {res_p['total_docs']:<20,d} | {'--':<12}")
        print(f"{'Thời gian nạp (Duration)':<30} | {res_v['duration']/1e9:<19.2f}s | {res_p['duration']/1e9:<19.2f}s | {((res_v['duration']-res_p['duration'])/res_p['duration'])*100:+.1f}%")
        print(f"{'Thông lượng (Throughput)':<30} | {res_v['docs_per_sec']:<17,.1f} d/s | {res_p['docs_per_sec']:<17,.1f} d/s | {tps_diff:+.1f}%")
        print(f"{'Băng thông Ingestion':<30} | {res_v['mb_per_sec']:<17.2f} MB/s| {res_p['mb_per_sec']:<17.2f} MB/s| {((res_v['mb_per_sec']-res_p['mb_per_sec'])/res_p['mb_per_sec'])*100:+.1f}%")
        print(f"{'Độ trễ p50 (Median)':<30} | {res_v['latency_p50_ms']:<18.2f}ms | {res_p['latency_p50_ms']:<18.2f}ms | {p50_diff:+.1f}%")
        print(f"{'Độ trễ p90':<30} | {res_v['latency_p90_ms']:<18.2f}ms | {res_p['latency_p90_ms']:<18.2f}ms | {((res_v['latency_p90_ms']-res_p['latency_p90_ms'])/res_p['latency_p90_ms'])*100:+.1f}%")
        print(f"{'Độ trễ p95':<30} | {res_v['latency_p95_ms']:<18.2f}ms | {res_p['latency_p95_ms']:<18.2f}ms | {p95_diff:+.1f}%")
        print(f"{'Độ trễ p99':<30} | {res_v['latency_p99_ms']:<18.2f}ms | {res_p['latency_p99_ms']:<18.2f}ms | {((res_v['latency_p99_ms']-res_p['latency_p99_ms'])/res_p['latency_p99_ms'])*100:+.1f}%")
        print(f"{'JVM GC Collections':<30} | {res_v['jvm_gc_count']:<20d} | {res_p['jvm_gc_count']:<20d} | {((res_v['jvm_gc_count']-res_p['jvm_gc_count'])/(res_p['jvm_gc_count'] if res_p['jvm_gc_count']>0 else 1))*100:+.1f}%")
        print(f"{'JVM GC Pause Time':<30} | {res_v['jvm_gc_time_ms']:<18d}ms | {res_p['jvm_gc_time_ms']:<18d}ms | {gc_pause_diff:+.1f}%")
        print(f"{'Dung lượng Disk Store':<30} | {res_v['disk_store_mb']:<17.1f} MB | {res_p['disk_store_mb']:<17.1f} MB | {((res_v['disk_store_mb']-res_p['disk_store_mb'])/res_p['disk_store_mb'])*100:+.1f}%")
        print(f"{'Số Lucene Segments':<30} | {res_v['segments_count']:<20d} | {res_p['segments_count']:<20d} | {'--':<12}")
        print("=" * 85)

        # Lưu bản so sánh vào file JSON tổng hợp
        summary_result = {
            "num_docs": NUM_DOCS,
            "concurrency": CONCURRENCY,
            "batch_size": BATCH_SIZE,
            "vanilla": res_v,
            "plugin": res_p,
            "throughput_diff_percent": tps_diff,
            "p50_latency_diff_percent": p50_diff,
            "p95_latency_diff_percent": p95_diff,
            "gc_pause_diff_percent": gc_pause_diff,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        with open("benchmarks/res_1m_comparison.json", "w", encoding="utf-8") as f:
            json.dump(summary_result, f, indent=2, ensure_ascii=False)
        print("💾 Đã lưu kết quả so sánh đối soát vào: benchmarks/res_1m_comparison.json\n")

if __name__ == "__main__":
    main()
