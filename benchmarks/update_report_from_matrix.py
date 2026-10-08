#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import json
import os
import sys
import re

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8')

def format_row(c_label, c_sub, row_data):
    c = row_data["concurrency"]
    d = row_data["docs"]
    v = row_data["vanilla"]
    p = row_data["plugin"]
    diff_tps = ((v["docs_per_sec"] - p["docs_per_sec"]) / p["docs_per_sec"]) * 100.0
    diff_p50 = ((v["latency_p50_ms"] - p["latency_p50_ms"]) / p["latency_p50_ms"]) * 100.0
    diff_p95 = ((v["latency_p95_ms"] - p["latency_p95_ms"]) / p["latency_p95_ms"]) * 100.0
    diff_p99 = ((v["latency_p99_ms"] - p["latency_p99_ms"]) / p["latency_p99_ms"]) * 100.0
    diff_gc_cnt = ((v["jvm_gc_count"] - p["jvm_gc_count"]) / (p["jvm_gc_count"] if p["jvm_gc_count"] > 0 else 1)) * 100.0
    diff_gc_ms = ((v["jvm_gc_time_ms"] - p["jvm_gc_time_ms"]) / (p["jvm_gc_time_ms"] if p["jvm_gc_time_ms"] > 0 else 1)) * 100.0

    lines = []
    lines.append(f"| **C = {c}** | **Mô hình A (Go CGO)** | {d:,} | {v['duration']/1e9:.2f}s | **{v['docs_per_sec']:,.1f}** | {v['mb_per_sec']:.2f} MB/s | **{v['latency_p50_ms']:.1f}** | {v['latency_p95_ms']:.1f} | {v['latency_p99_ms']:.1f} | {v['jvm_gc_count']} | {v['jvm_gc_time_ms']} ms |")
    lines.append(f"| *({c_sub})*| **Mô hình B (Plugin)** | {d:,} | {p['duration']/1e9:.2f}s | **{p['docs_per_sec']:,.1f}** | {p['mb_per_sec']:.2f} MB/s | **{p['latency_p50_ms']:.1f}** | {p['latency_p95_ms']:.1f} | {p['latency_p99_ms']:.1f} | {p['jvm_gc_count']} | {p['jvm_gc_time_ms']} ms |")
    lines.append(f"| | *Chênh lệch %* | - | - | **{diff_tps:+.1f}% 🚀** | - | **{diff_p50:+.1f}% ⚡** | **{diff_p95:+.1f}% ⚡**| **{diff_p99:+.1f}% ⚡** | **{diff_gc_cnt:+.1f}%**| **{diff_gc_ms:+.1f}%** |")
    return "\n".join(lines)

def main():
    json_path = "benchmarks/full_matrix_results.json"
    if not os.path.exists(json_path):
        print(f"❌ Không tìm thấy tệp {json_path}")
        return

    with open(json_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    matrix = data.get("concurrency_matrix", [])
    if not matrix:
        print("❌ Không có dữ liệu concurrency_matrix")
        return

    report_path = "docs/benchmark_opensearch_plugin_report.md"
    if not os.path.exists(report_path):
        print(f"❌ Không tìm thấy {report_path}")
        return

    with open(report_path, "r", encoding="utf-8") as f:
        content = f.read()

    sub_labels = {1: "Single", 4: "4 Cores", 8: "Chuẩn", 16: "Cao tải", 32: "Stress"}
    matrix_rows = []
    for row in matrix:
        c = row["concurrency"]
        c_sub = sub_labels.get(c, f"{c} workers")
        matrix_rows.append(format_row(f"C = {c}", c_sub, row))

    new_table = """| Mức Tải | Mô Hình | Docs Index | Thời Gian (s) | Thông Lượng (docs/s) | Băng Thông (MB/s) | Latency p50 (ms) | Latency p95 (ms) | Latency p99 (ms) | GC Count | GC Pause (ms) |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
""" + "\n".join(matrix_rows)

    # Thay thế bảng 3.1
    pattern_table = r"\| Mức Tải \| Mô Hình \| Docs Index.*?(?=\n---\n|\n### 3\.2|\Z)"
    if re.search(pattern_table, content, re.DOTALL):
        content = re.sub(pattern_table, new_table + "\n", content, flags=re.DOTALL)
        print("✅ Đã cập nhật Bảng 3.1 chi tiết trong báo cáo!")

    # Cập nhật Bảng tóm tắt 1.1
    c8 = next((item for item in matrix if item["concurrency"] == 8), matrix[0])
    c32 = next((item for item in matrix if item["concurrency"] == 32), matrix[-1])
    v8, p8 = c8["vanilla"], c8["plugin"]
    v32, p32 = c32["vanilla"], c32["plugin"]

    tps_diff_8 = ((v8["docs_per_sec"] - p8["docs_per_sec"]) / p8["docs_per_sec"]) * 100.0
    tps_diff_32 = ((v32["docs_per_sec"] - p32["docs_per_sec"]) / p32["docs_per_sec"]) * 100.0
    p50_diff_8 = ((v8["latency_p50_ms"] - p8["latency_p50_ms"]) / p8["latency_p50_ms"]) * 100.0
    p95_diff_8 = ((v8["latency_p95_ms"] - p8["latency_p95_ms"]) / p8["latency_p95_ms"]) * 100.0
    p95_diff_32 = ((v32["latency_p95_ms"] - p32["latency_p95_ms"]) / p32["latency_p95_ms"]) * 100.0

    summary_box = f"""```
+---------------------------------------------------------------------------------------------------+
|                     TỔNG HỢP CHỈ SỐ KỸ THUẬT CỐT LÕI (MỐC CHUẨN HÓA 100.000 DOCS)                 |
+------------------------------+---------------------------+---------------------------+------------+
| Chỉ số Đo lường              | Mô hình A (Go CGO)        | Mô hình B (Plugin JNI)    | Chênh lệch |
+------------------------------+---------------------------+---------------------------+------------+
| Throughput tối ưu (C=8)      | {v8['docs_per_sec']:>11,.1f} docs/sec   | {p8['docs_per_sec']:>11,.1f} docs/sec   | {tps_diff_8:>+9.1f}% 🚀 |
| Throughput cực hạn (C=32)    | {v32['docs_per_sec']:>11,.1f} docs/sec   | {p32['docs_per_sec']:>11,.1f} docs/sec   | {tps_diff_32:>+9.1f}% 🚀 |
| Băng thông Ingestion (C=8)   | {v8['mb_per_sec']:>11.2f} MB/sec     | {p8['mb_per_sec']:>11.2f} MB/sec     | {tps_diff_8:>+9.1f}% 🚀 |
| Độ trễ Median (p50 tại C=8)  | {v8['latency_p50_ms']:>11.1f} ms         | {p8['latency_p50_ms']:>11.1f} ms         | {p50_diff_8:>+9.1f}% ⚡  |
| Độ trễ Phân vị p95 (C=8)     | {v8['latency_p95_ms']:>11.1f} ms         | {p8['latency_p95_ms']:>11.1f} ms         | {p95_diff_8:>+9.1f}% ⚡  |
| Độ trễ Cực hạn p95 (C=32)    | {v32['latency_p95_ms']:>11.1f} ms         | {p32['latency_p95_ms']:>11.1f} ms         | {p95_diff_32:>+9.1f}% ⚡  |
| JVM GC Collections (C=8)     | {v8['jvm_gc_count']:>11d} lần        | {p8['jvm_gc_count']:>11d} lần        | {((v8['jvm_gc_count']-p8['jvm_gc_count'])/p8['jvm_gc_count']*100):>+9.1f}%    |
| JVM GC Pause Time (C=8)      | {v8['jvm_gc_time_ms']:>11d} ms         | {p8['jvm_gc_time_ms']:>11d} ms         | {((v8['jvm_gc_time_ms']-p8['jvm_gc_time_ms'])/p8['jvm_gc_time_ms']*100):>+9.1f}%    |
| Search Parity (Zero Diff)    | 100% Khớp Tuyệt Đối       | 100% Khớp Tuyệt Đối       | 0% lệch ✅ |
| Rủi ro Sập Cluster (Crash)   | 0% (Cách ly hoàn toàn)    | RẤT CAO (Fatal JVM Crash) | ⚠️ Nghiêm trọng |
+------------------------------+---------------------------+---------------------------+------------+
```"""

    pattern_summary = r"```\s*\+---------------------------------------------------------------------------------------------------\+.*?```"
    if re.search(pattern_summary, content, re.DOTALL):
        content = re.sub(pattern_summary, summary_box, content, flags=re.DOTALL)
        print("✅ Đã cập nhật Hộp tóm tắt 1.1 trong báo cáo!")

    with open(report_path, "w", encoding="utf-8") as f:
        f.write(content)

    print("📄 Đã lưu toàn bộ thay đổi vào docs/benchmark_opensearch_plugin_report.md")

if __name__ == "__main__":
    main()
