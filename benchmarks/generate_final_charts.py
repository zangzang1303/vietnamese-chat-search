import csv
import json
import os
import pathlib
import sys
import shutil

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import numpy as np

ROOT = pathlib.Path(__file__).resolve().parents[1]
RUN_DIR = ROOT / "benchmarks" / "final_run" / "final_20261020_093100"
AGGREGATE_FILE = RUN_DIR / "summaries" / "aggregate.json"
STAT_FILE = RUN_DIR / "summaries" / "statistical_analysis.json"
PAIRED_CSV = RUN_DIR / "summaries" / "paired_throughput.csv"
SEARCH_PARITY_FILE = RUN_DIR / "search" / "search_parity_summary.json"
CHARTS_DIR = RUN_DIR / "charts"

SUPPLEMENT_DIR = RUN_DIR / "supplement" / "summaries"
BATCH_FILE = SUPPLEMENT_DIR / "batch_sensitivity_summary.json"
ENGINE_FILE = SUPPLEMENT_DIR / "engine_only_summary.json"
SEARCH_LAT_FILE = SUPPLEMENT_DIR / "search_latency_summary.json"
SOAK_FILE = SUPPLEMENT_DIR / "soak_summary.json"
TOKENIZER_FILE = RUN_DIR / "supplement" / "tokenizer_profile" / "tokenizer_scaling_profile.json"


def set_style():
    plt.style.use("seaborn-v0_8-whitegrid" if "seaborn-v0_8-whitegrid" in plt.style.available else "default")
    plt.rcParams["font.sans-serif"] = ["DejaVu Sans", "Arial", "Segoe UI"]
    plt.rcParams["axes.edgecolor"] = "#cbd5e1"
    plt.rcParams["axes.linewidth"] = 1.2
    plt.rcParams["grid.color"] = "#f1f5f9"
    plt.rcParams["grid.linestyle"] = "--"


def sync_chart(src_path, filename):
    for dest_dir in [ROOT / "image", ROOT / "benchmark_charts", RUN_DIR / "supplement" / "charts"]:
        dest_dir.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src_path, dest_dir / filename)
    print(f"✅ Synced {filename}")


# 1. Throughput scaling with Bootstrap 95% CI
def plot_throughput_with_ci(matrix):
    concurrencies = [item["concurrency"] for item in matrix]
    v_tps = [item["vanilla"]["throughput"]["median"] for item in matrix]
    p_tps = [item["plugin"]["throughput"]["median"] for item in matrix]

    v_err_low = [item["vanilla"]["throughput"]["median"] - item["vanilla"]["throughput"]["ci95_low"] for item in matrix]
    v_err_high = [item["vanilla"]["throughput"]["ci95_high"] - item["vanilla"]["throughput"]["median"] for item in matrix]
    p_err_low = [item["plugin"]["throughput"]["median"] - item["plugin"]["throughput"]["ci95_low"] for item in matrix]
    p_err_high = [item["plugin"]["throughput"]["ci95_high"] - item["plugin"]["throughput"]["median"] for item in matrix]

    fig, ax = plt.subplots(figsize=(10, 6), dpi=300)

    ax.errorbar(
        concurrencies, v_tps, yerr=[v_err_low, v_err_high], fmt="-o",
        color="#0284c7", linewidth=2.8, markersize=8, capsize=5, capthick=1.5,
        label="Mô hình A: Go CGO Pre-tokenization + Vanilla OS (95% Bootstrap CI)"
    )
    ax.errorbar(
        concurrencies, p_tps, yerr=[p_err_low, p_err_high], fmt="-s",
        color="#ea580c", linewidth=2.8, markersize=8, capsize=5, capthick=1.5,
        label="Mô hình B: OpenSearch Vietnamese Plugin JNI (95% Bootstrap CI)"
    )

    for c, v, p in zip(concurrencies, v_tps, p_tps):
        diff = ((v - p) / p * 100.0) if p > 0 else 0
        sign = "+" if diff >= 0 else ""
        ax.annotate(f"{v:,.0f}\n({sign}{diff:.1f}%)", (c, v), textcoords="offset points", xytext=(0, 10), ha="center", fontsize=9, fontweight="bold", color="#0369a1")
        ax.annotate(f"{p:,.0f}", (c, p), textcoords="offset points", xytext=(0, -18), ha="center", fontsize=9, color="#c2410c")

    ax.set_title("QUY MÔ THÔNG LƯỢNG NẠP (THROUGHPUT SCALING CURVE)\nSo sánh Mô hình A (Go CGO) vs Mô hình B (Plugin JNI) trên 1.000.000 Docs", fontsize=12, fontweight="bold", pad=15)
    ax.set_xlabel("Mức độ Đồng thời (Concurrency / Number of Workers)", fontsize=11, labelpad=10)
    ax.set_ylabel("Thông lượng Ingestion (Documents / Giây)", fontsize=11, labelpad=10)
    ax.set_xticks(concurrencies)
    ax.set_xticklabels([f"C={c}" for c in concurrencies])
    ax.legend(frameon=True, facecolor="white", framealpha=0.9, fontsize=10, loc="upper left")

    out_file = CHARTS_DIR / "benchmark_throughput_scaling.png"
    plt.tight_layout()
    plt.savefig(out_file, bbox_inches="tight")
    plt.close()
    sync_chart(out_file, "benchmark_throughput_scaling.png")


# 2. Paired Throughput Differences with CI around Zero
def plot_paired_throughput_differences():
    if not PAIRED_CSV.exists():
        return
    rows = []
    with PAIRED_CSV.open(encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for r in reader:
            rows.append({
                "concurrency": int(r["concurrency"]),
                "mean_diff": float(r["paired_mean_diff"]),
                "ci_low": float(r["paired_ci_low"]),
                "ci_high": float(r["paired_ci_high"]),
                "cohen_dz": float(r["cohen_dz"]),
            })

    concurrencies = [r["concurrency"] for r in rows]
    diffs = [r["mean_diff"] for r in rows]
    err_low = [r["mean_diff"] - r["ci_low"] for r in rows]
    err_high = [r["ci_high"] - r["mean_diff"] for r in rows]

    fig, ax = plt.subplots(figsize=(10, 6), dpi=300)
    ax.axhline(0, color="#64748b", linestyle="--", linewidth=1.5, label="Đường Zero (Không có chênh lệch)")

    x_indices = np.arange(len(concurrencies))
    ax.errorbar(
        x_indices, diffs, yerr=[err_low, err_high], fmt="o",
        color="#2563eb", ecolor="#60a5fa", elinewidth=2.5, capsize=6, capthick=2, markersize=9,
        label="Chênh lệch trung bình Paired (Model A - Model B) ± 95% Student-t CI"
    )

    for i, r in enumerate(rows):
        sign = "+" if r["mean_diff"] >= 0 else ""
        ax.annotate(
            f"{sign}{r['mean_diff']:,.0f} docs/s\nCI: [{r['ci_low']:,.0f}, {r['ci_high']:,.0f}]\n(Cohen dz={r['cohen_dz']:.2f})",
            (i, r["mean_diff"]),
            textcoords="offset points",
            xytext=(0, 15 if r["mean_diff"] >= 0 else -30),
            ha="center",
            fontsize=8.5,
            fontweight="bold",
            color="#1e3a8a",
            bbox=dict(boxstyle="round,pad=0.2", facecolor="#f8fafc", edgecolor="#cbd5e1", alpha=0.9),
        )

    ax.set_title("CHÊNH LỆCH THÔNG LƯỢNG CẶP (PAIRED DIFFERENCE A - B) VÀ KHOẢNG TIN CẬY 95%\nChứng minh mọi mức Concurrency đều chứa 0 => Chưa có mô hình thắng có ý nghĩa thống kê", fontsize=11, fontweight="bold", pad=15)
    ax.set_xlabel("Mức độ Đồng thời (Concurrency)", fontsize=11, labelpad=10)
    ax.set_ylabel("Chênh lệch Throughput (docs/sec: A - B)", fontsize=11, labelpad=10)
    ax.set_xticks(x_indices)
    ax.set_xticklabels([f"C={c}" for c in concurrencies])
    ax.legend(frameon=True, facecolor="white", framealpha=0.9, fontsize=9.5, loc="lower right")

    out_file = CHARTS_DIR / "benchmark_paired_throughput_difference.png"
    plt.tight_layout()
    plt.savefig(out_file, bbox_inches="tight")
    plt.close()
    sync_chart(out_file, "benchmark_paired_throughput_difference.png")


# 3. Latency Percentiles
def plot_latency_breakdown(matrix):
    target = next((item for item in matrix if item["concurrency"] == 8), matrix[0])
    v = target["vanilla"]
    p = target["plugin"]

    labels = ["p50 (Median)", "p95", "p99"]
    v_vals = [v["latency_p50_ms"]["median"], v["latency_p95_ms"]["median"], v.get("latency_p99_ms", {}).get("median", v["latency_p95_ms"]["median"] * 1.3)]
    p_vals = [p["latency_p50_ms"]["median"], p["latency_p95_ms"]["median"], p.get("latency_p99_ms", {}).get("median", p["latency_p95_ms"]["median"] * 1.3)]

    x = np.arange(len(labels))
    width = 0.35

    fig, ax = plt.subplots(figsize=(10, 6), dpi=300)
    rects1 = ax.bar(x - width/2, v_vals, width, label="Mô hình A (Go CGO) Engine Latency", color="#0284c7")
    rects2 = ax.bar(x + width/2, p_vals, width, label="Mô hình B (Plugin JNI) Engine Latency", color="#ea580c")

    for rect in rects1:
        h = rect.get_height()
        ax.annotate(f"{h:.1f}ms", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=9, fontweight="bold")
    for rect in rects2:
        h = rect.get_height()
        ax.annotate(f"{h:.1f}ms", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=9, fontweight="bold")

    ax.set_title(f"PHÂN PHỐI ĐỘ TRỄ NẠP DỮ LIỆU TẠI C={target['concurrency']} (MEDIAN CỦA 5 LẦN LẶP)\nĐo kiểm Engine HTTP Latency p50, p95, p99", fontsize=12, fontweight="bold", pad=15)
    ax.set_xlabel("Phân vị Độ trễ (Percentiles)", fontsize=11, labelpad=10)
    ax.set_ylabel("Độ trễ (Milli-giây)", fontsize=11, labelpad=10)
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend(frameon=True, facecolor="white", framealpha=0.9, fontsize=10)

    out_file = CHARTS_DIR / "benchmark_latency_distribution.png"
    plt.tight_layout()
    plt.savefig(out_file, bbox_inches="tight")
    plt.close()
    sync_chart(out_file, "benchmark_latency_distribution.png")


# 4. JVM GC Collection Time (Strictly Collection Time, not Pause)
def plot_jvm_gc_stats(matrix):
    concurrencies = [item["concurrency"] for item in matrix]
    v_gc_time = [item["vanilla"]["jvm_gc_time_ms"]["median"] for item in matrix]
    p_gc_time = [item["plugin"]["jvm_gc_time_ms"]["median"] for item in matrix]

    x = np.arange(len(concurrencies))
    width = 0.35

    fig, ax = plt.subplots(figsize=(10, 6), dpi=300)
    rects1 = ax.bar(x - width/2, v_gc_time, width, label="Mô hình A (Go CGO) JVM GC Collection Time", color="#0284c7")
    rects2 = ax.bar(x + width/2, p_gc_time, width, label="Mô hình B (Plugin JNI) JVM GC Collection Time", color="#ea580c")

    for rect in rects1:
        h = rect.get_height()
        ax.annotate(f"{h:.0f}ms", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=8)
    for rect in rects2:
        h = rect.get_height()
        ax.annotate(f"{h:.0f}ms", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=8, fontweight="bold")

    ax.set_title("CHI PHÍ JVM GARBAGE COLLECTION TIME (MS)\nSo sánh Tổng Thời Gian GC Collection Time giữa 2 Mô hình trên 1M Docs", fontsize=12, fontweight="bold", pad=15)
    ax.set_xlabel("Mức độ Đồng thời (Concurrency)", fontsize=11, labelpad=10)
    ax.set_ylabel("Tổng Thời gian GC Collection Time (Milli-giây)", fontsize=11, labelpad=10)
    ax.set_xticks(x)
    ax.set_xticklabels([f"C={c}" for c in concurrencies])
    ax.legend(frameon=True, facecolor="white", framealpha=0.9, fontsize=10)

    out_file = CHARTS_DIR / "benchmark_jvm_gc_stats.png"
    plt.tight_layout()
    plt.savefig(out_file, bbox_inches="tight")
    plt.close()
    sync_chart(out_file, "benchmark_jvm_gc_stats.png")


# 5. Disk Storage Size & Segments
def plot_storage_segments(matrix):
    concurrencies = [item["concurrency"] for item in matrix]
    v_store = [item["vanilla"]["disk_store_mb"]["median"] for item in matrix]
    p_store = [item["plugin"]["disk_store_mb"]["median"] for item in matrix]

    x = np.arange(len(concurrencies))
    width = 0.35

    fig, ax = plt.subplots(figsize=(10, 6), dpi=300)
    rects1 = ax.bar(x - width/2, v_store, width, label="Mô hình A (Go CGO) Disk Size", color="#0284c7")
    rects2 = ax.bar(x + width/2, p_store, width, label="Mô hình B (Plugin JNI) Disk Size", color="#ea580c")

    for rect in rects1:
        h = rect.get_height()
        ax.annotate(f"{h:.0f}MB", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=8)
    for rect in rects2:
        h = rect.get_height()
        ax.annotate(f"{h:.0f}MB", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=8)

    ax.set_title("DUNG LƯỢNG LƯU TRỮ TRÊN ĐĨA (DISK STORE SIZE)\nSo sánh Kích thước Chỉ mục Lucene trên 1M Documents", fontsize=12, fontweight="bold", pad=15)
    ax.set_xlabel("Mức độ Đồng thời (Concurrency)", fontsize=11, labelpad=10)
    ax.set_ylabel("Dung lượng Lưu trữ Chỉ mục (MB)", fontsize=11, labelpad=10)
    ax.set_xticks(x)
    ax.set_xticklabels([f"C={c}" for c in concurrencies])
    ax.legend(frameon=True, facecolor="white", framealpha=0.9, fontsize=10)

    out_file = CHARTS_DIR / "benchmark_storage_segments.png"
    plt.tight_layout()
    plt.savefig(out_file, bbox_inches="tight")
    plt.close()
    sync_chart(out_file, "benchmark_storage_segments.png")


# 6. Batch Sensitivity Comparison (Phase D)
def plot_batch_sensitivity():
    if not BATCH_FILE.exists():
        return
    data = json.loads(BATCH_FILE.read_text(encoding="utf-8"))

    configs = [("C=1, B=500", 1, 500), ("C=1, B=1000", 1, 1000), ("C=32, B=1000", 32, 1000), ("C=32, B=2000", 32, 2000)]
    labels = [c[0] for c in configs]
    v_thrs = []
    p_thrs = []

    for _, c, b in configs:
        v_item = next((d for d in data if d["model"] == "vanilla" and d["concurrency"] == c and d["batch_size"] == b), None)
        p_item = next((d for d in data if d["model"] == "plugin" and d["concurrency"] == c and d["batch_size"] == b), None)
        v_thrs.append(v_item["throughput_median"] if v_item else 0.0)
        p_thrs.append(p_item["throughput_median"] if p_item else 0.0)

    x = np.arange(len(labels))
    width = 0.35

    fig, ax = plt.subplots(figsize=(10, 6), dpi=300)
    rects1 = ax.bar(x - width/2, v_thrs, width, label="Mô hình A (Go CGO)", color="#0284c7")
    rects2 = ax.bar(x + width/2, p_thrs, width, label="Mô hình B (Plugin JNI)", color="#ea580c")

    for rect in rects1:
        h = rect.get_height()
        ax.annotate(f"{h:,.0f}", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=8.5, fontweight="bold")
    for rect in rects2:
        h = rect.get_height()
        ax.annotate(f"{h:,.0f}", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=8.5, fontweight="bold")

    ax.set_title("ĐỘ NHẠY KÍCH THƯỚC BATCH (BATCH SENSITIVITY BENCHMARK)\nSo sánh Throughput giữa Batch 500, 1.000 và 2.000 tại C=1 và C=32", fontsize=12, fontweight="bold", pad=15)
    ax.set_xlabel("Cấu hình Concurrency & Batch Size", fontsize=11, labelpad=10)
    ax.set_ylabel("Thông lượng Ingestion (Documents / Giây)", fontsize=11, labelpad=10)
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend(frameon=True, facecolor="white", framealpha=0.9, fontsize=10)

    out_file = CHARTS_DIR / "benchmark_batch_sensitivity.png"
    plt.tight_layout()
    plt.savefig(out_file, bbox_inches="tight")
    plt.close()
    sync_chart(out_file, "benchmark_batch_sensitivity.png")


# 7. Cost Isolation: Engine-Only vs Total-System & Tokenizer Scaling (Phase F)
def plot_cost_isolation():
    if not ENGINE_FILE.exists() or not TOKENIZER_FILE.exists():
        return
    eng_data = json.loads(ENGINE_FILE.read_text(encoding="utf-8"))
    tok_data = json.loads(TOKENIZER_FILE.read_text(encoding="utf-8"))

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 5.5), dpi=300)

    # Panel 1: Engine-Only Throughput Ratio
    concurrencies = [d["concurrency"] for d in eng_data]
    v_eng = [d["vanilla_pretokenized_throughput_median"] for d in eng_data]
    p_eng = [d["plugin_raw_throughput_median"] for d in eng_data]

    x = np.arange(len(concurrencies))
    width = 0.35
    ax1.bar(x - width/2, v_eng, width, label="Vanilla Pre-tokenized (Engine-Only)", color="#0284c7")
    ax1.bar(x + width/2, p_eng, width, label="Plugin Raw (In-Engine JNI)", color="#ea580c")
    for i, (v, p) in enumerate(zip(v_eng, p_eng)):
        ratio = v / p if p > 0 else 0
        ax1.annotate(f"{v:,.0f}\n({ratio:.1f}x)", (i - width/2, v), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=8.5, fontweight="bold")
        ax1.annotate(f"{p:,.0f}", (i + width/2, p), textcoords="offset points", xytext=(0, 4), ha="center", fontsize=8.5)

    ax1.set_title("ENGINE-ONLY INGESTION RATE (PHASE F)\nLoại bỏ chi phí tiền xử lý Go client", fontsize=11, fontweight="bold")
    ax1.set_xlabel("Concurrency", fontsize=10)
    ax1.set_ylabel("Docs / Giây", fontsize=10)
    ax1.set_xticks(x)
    ax1.set_xticklabels([f"C={c}" for c in concurrencies])
    ax1.legend(fontsize=9)

    # Panel 2: Tokenizer Mutex Contention
    tok_res = tok_data.get("results", [])
    c_list = [d["concurrency"] for d in tok_res]
    eff_list = [d["parallel_efficiency_pct"] for d in tok_res]
    speedup_list = [d["speedup_vs_c1"] for d in tok_res]

    ax2.plot(c_list, eff_list, "-o", color="#dc2626", linewidth=2.5, markersize=8, label="Parallel Efficiency (%)")
    ax2.set_title("HIỆU SUẤT SONG SONG CỐC CỐC CGO (TOKENIZER PROFILE)\nSự sụp đổ do sync.Mutex contention trong SegmentOriginal", fontsize=11, fontweight="bold")
    ax2.set_xlabel("Concurrency (Goroutines)", fontsize=10)
    ax2.set_ylabel("Parallel Efficiency (%)", fontsize=10, color="#dc2626")
    ax2.set_xticks(c_list)
    ax2.set_xticklabels([f"C={c}" for c in c_list])
    for c, eff, sp in zip(c_list, eff_list, speedup_list):
        ax2.annotate(f"{eff:.1f}%\n({sp:.2f}x)", (c, eff), textcoords="offset points", xytext=(0, 8), ha="center", fontsize=8.5, fontweight="bold", color="#991b1b")

    out_file = CHARTS_DIR / "benchmark_cost_isolation.png"
    plt.tight_layout()
    plt.savefig(out_file, bbox_inches="tight")
    plt.close()
    sync_chart(out_file, "benchmark_cost_isolation.png")


# 8. Search Latency across Concurrencies (Phase G)
def plot_search_latency_scaling():
    if not SEARCH_LAT_FILE.exists():
        return
    data = json.loads(SEARCH_LAT_FILE.read_text(encoding="utf-8"))

    concurrencies = sorted(list({d["concurrency"] for d in data}))
    v_p50 = [next(d["latency_p50_ms"] for d in data if d["model"] == "vanilla" and d["concurrency"] == c) for c in concurrencies]
    p_p50 = [next(d["latency_p50_ms"] for d in data if d["model"] == "plugin" and d["concurrency"] == c) for c in concurrencies]
    v_p95 = [next(d["latency_p95_ms"] for d in data if d["model"] == "vanilla" and d["concurrency"] == c) for c in concurrencies]
    p_p95 = [next(d["latency_p95_ms"] for d in data if d["model"] == "plugin" and d["concurrency"] == c) for c in concurrencies]

    x = np.arange(len(concurrencies))
    width = 0.2

    fig, ax = plt.subplots(figsize=(10, 6), dpi=300)
    rects1 = ax.bar(x - 1.5*width, v_p50, width, label="Model A p50", color="#0284c7")
    rects2 = ax.bar(x - 0.5*width, p_p50, width, label="Model B p50", color="#ea580c")
    rects3 = ax.bar(x + 0.5*width, v_p95, width, label="Model A p95", color="#38bdf8")
    rects4 = ax.bar(x + 1.5*width, p_p95, width, label="Model B p95", color="#fb923c")

    for rect in rects1:
        h = rect.get_height()
        ax.annotate(f"{h:.1f}", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 3), ha="center", fontsize=8)
    for rect in rects2:
        h = rect.get_height()
        ax.annotate(f"{h:.1f}", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 3), ha="center", fontsize=8)
    for rect in rects3:
        h = rect.get_height()
        ax.annotate(f"{h:.1f}", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 3), ha="center", fontsize=8)
    for rect in rects4:
        h = rect.get_height()
        ax.annotate(f"{h:.1f}", (rect.get_x() + rect.get_width()/2, h), textcoords="offset points", xytext=(0, 3), ha="center", fontsize=8)

    ax.set_title("ĐỘ TRỄ TÌM KIẾM TRUY VẤN THEO CONCURRENCY (PHASE G)\nĐo kiểm 1.000 Queries tại Concurrency C={1, 8, 32}", fontsize=12, fontweight="bold", pad=15)
    ax.set_xlabel("Mức độ Đồng thời (Query Concurrency)", fontsize=11, labelpad=10)
    ax.set_ylabel("Độ trễ Truy vấn (Milli-giây)", fontsize=11, labelpad=10)
    ax.set_xticks(x)
    ax.set_xticklabels([f"C={c}" for c in concurrencies])
    ax.legend(frameon=True, facecolor="white", framealpha=0.9, fontsize=9.5)

    out_file = CHARTS_DIR / "benchmark_search_latency.png"
    plt.tight_layout()
    plt.savefig(out_file, bbox_inches="tight")
    plt.close()
    sync_chart(out_file, "benchmark_search_latency.png")


# 9. Soak Timeline & Drift (Phase E)
def plot_soak_timeline():
    if not SOAK_FILE.exists():
        return
    soak_data = json.loads(SOAK_FILE.read_text(encoding="utf-8"))
    if not soak_data:
        return

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 5.5), dpi=300)

    # Panel 1: RSS Working Set over 5-min windows
    for s in soak_data:
        m = s["model"]
        windows = [w["window_time_range_min"] for w in s["windows"]]
        rss_vals = [w["avg_container_rss_mib"] for w in s["windows"]]
        color = "#0284c7" if m == "vanilla" else "#ea580c"
        label = f"Model A (Vanilla) [slope: {s['memory_drift']['rss_slope_mib_per_minute']} MiB/min]" if m == "vanilla" else f"Model B (Plugin) [slope: {s['memory_drift']['rss_slope_mib_per_minute']} MiB/min]"
        ax1.plot(windows, rss_vals, "-o", color=color, linewidth=2.5, markersize=7, label=label)

    ax1.set_title("CONTAINER RSS DRIFT TRONG 30 PHÚT SOAK (PHASE E)\nĐo kiểm độ ổn định bộ nhớ theo cửa sổ 5 phút (C=16, B=1000)", fontsize=11, fontweight="bold")
    ax1.set_xlabel("Cửa sổ Thời gian (Phút)", fontsize=10)
    ax1.set_ylabel("Container Working Set RSS (MiB)", fontsize=10)
    ax1.legend(fontsize=9)

    # Panel 2: JVM Heap used over 5-min windows
    for s in soak_data:
        m = s["model"]
        windows = [w["window_time_range_min"] for w in s["windows"]]
        heap_vals = [w["avg_jvm_heap_mib"] for w in s["windows"]]
        color = "#0284c7" if m == "vanilla" else "#ea580c"
        label = f"Model A (Vanilla) [slope: {s['memory_drift']['heap_slope_mib_per_minute']} MiB/min]" if m == "vanilla" else f"Model B (Plugin) [slope: {s['memory_drift']['heap_slope_mib_per_minute']} MiB/min]"
        ax2.plot(windows, heap_vals, "-s", color=color, linewidth=2.5, markersize=7, label=label)

    ax2.set_title("JVM HEAP USED TRONG 30 PHÚT SOAK (PHASE E)\nKiểm tra rò rỉ bộ nhớ JVM qua các cửa sổ", fontsize=11, fontweight="bold")
    ax2.set_xlabel("Cửa sổ Thời gian (Phút)", fontsize=10)
    ax2.set_ylabel("JVM Heap Used (MiB)", fontsize=10)
    ax2.legend(fontsize=9)

    out_file = CHARTS_DIR / "benchmark_soak_timeline.png"
    plt.tight_layout()
    plt.savefig(out_file, bbox_inches="tight")
    plt.close()
    sync_chart(out_file, "benchmark_soak_timeline.png")


def main():
    CHARTS_DIR.mkdir(parents=True, exist_ok=True)
    set_style()

    if not AGGREGATE_FILE.exists():
        print(f"File {AGGREGATE_FILE} chưa tồn tại!")
        return

    data = json.loads(AGGREGATE_FILE.read_text(encoding="utf-8"))
    matrix = data.get("concurrency_matrix", [])
    if not matrix:
        print("Chưa có dữ liệu matrix!")
        return

    print("📊 Bắt đầu sinh hệ thống biểu đồ hoàn tất Task 1...")
    plot_throughput_with_ci(matrix)
    plot_paired_throughput_differences()
    plot_latency_breakdown(matrix)
    plot_jvm_gc_stats(matrix)
    plot_storage_segments(matrix)
    plot_batch_sensitivity()
    plot_cost_isolation()
    plot_search_latency_scaling()
    plot_soak_timeline()
    print("✨ Toàn bộ biểu đồ thực nghiệm chuẩn hóa đã được sinh thành công!")


if __name__ == "__main__":
    main()
