import argparse
import csv
import datetime
import json
import math
import os
import pathlib
import random
import subprocess
import sys
import threading
import time
import urllib.error
import urllib.request

sys.stdout.reconfigure(encoding="utf-8")

ROOT = pathlib.Path(__file__).resolve().parents[1]
DATASET = ROOT / "benchmarks" / "dataset_real_1m.jsonl"
RUN_ID = "final_20261020_093100"
RUN_DIR = ROOT / "benchmarks" / "final_run" / RUN_ID

MODELS = {
    "vanilla": {
        "name": "Model A (Vanilla CGO)",
        "url": "http://localhost:9201",
        "container": "os_vanilla_bench",
        "other": "os_plugin_bench",
        "mapping": ROOT / "benchmarks" / "mapping_vanilla.json",
        "runner_mode": "vanilla",
    },
    "plugin": {
        "name": "Model B (Plugin JNI)",
        "url": "http://localhost:9202",
        "container": "os_plugin_bench",
        "other": "os_vanilla_bench",
        "mapping": ROOT / "benchmarks" / "mapping_plugin.json",
        "runner_mode": "plugin",
    },
}

def docker(*args):
    proc = subprocess.run(
        ["wsl", "-u", "root", "-d", "Ubuntu", "docker", *args],
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace"
    )
    if proc.returncode != 0:
        raise RuntimeError(f"docker {' '.join(args)} failed: {proc.stderr}")
    return proc.stdout.strip()

def wsl_path(path):
    p = pathlib.Path(path).resolve()
    drive = p.drive.rstrip(":").lower()
    tail = p.as_posix().split(":", 1)[1]
    return f"/mnt/{drive}{tail}"

def request_json(url, method="GET", payload=None, timeout=15):
    data = None if payload is None else json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=data,
        method=method,
        headers={"Content-Type": "application/json"}
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read().decode("utf-8"))

def wait_ready(url, seconds=180):
    deadline = time.time() + seconds
    while time.time() < deadline:
        try:
            res = request_json(url + "/_cluster/health", timeout=3)
            if res.get("status") in {"green", "yellow"}:
                time.sleep(3)
                return
        except Exception:
            time.sleep(2)
    raise TimeoutError(f"OpenSearch not ready at {url}")

def recreate_index(url, mapping_path):
    try:
        request_json(url + "/messages", method="DELETE")
    except urllib.error.HTTPError as exc:
        if exc.code != 404:
            raise
    mapping = json.loads(pathlib.Path(mapping_path).read_text(encoding="utf-8"))
    request_json(url + "/messages", method="PUT", payload=mapping)
    wait_ready(url)

class TelemetryCollector:
    def __init__(self, container_name, target_url):
        self.container_name = container_name
        self.target_url = target_url
        self.samples = []
        self._stop_event = threading.Event()
        self._thread = None

    def start(self):
        self.samples = []
        self._stop_event.clear()
        self._thread = threading.Thread(target=self._run, daemon=True)
        self._thread.start()

    def stop(self):
        self._stop_event.set()
        if self._thread:
            self._thread.join(timeout=3)

    def _run(self):
        while not self._stop_event.is_set():
            t_now = time.time()
            sample = {"timestamp": t_now}
            
            # Docker stats
            try:
                proc = subprocess.run(
                    ["wsl", "-u", "root", "-d", "Ubuntu", "docker", "stats", "--no-stream", "--format", "{{.CPUPerc}}|{{.MemUsage}}", self.container_name],
                    capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=2
                )
                if proc.returncode == 0 and proc.stdout.strip():
                    parts = proc.stdout.strip().split("|")
                    sample["container_cpu_str"] = parts[0]
                    sample["container_mem_str"] = parts[1] if len(parts) > 1 else ""
                    sample["container_cpu_pct"] = float(parts[0].rstrip("%")) if "%" in parts[0] else 0.0
            except Exception:
                pass

            # OpenSearch JVM & ThreadPool stats
            try:
                ns = request_json(self.target_url + "/_nodes/stats/jvm,thread_pool,indices", timeout=2)
                for node_id, n in ns.get("nodes", {}).items():
                    sample["jvm_heap_used_mb"] = n.get("jvm", {}).get("mem", {}).get("heap_used_in_bytes", 0) / 1024.0 / 1024.0
                    young = n.get("jvm", {}).get("gc", {}).get("collectors", {}).get("young", {})
                    old = n.get("jvm", {}).get("gc", {}).get("collectors", {}).get("old", {})
                    sample["young_gc_count"] = young.get("collection_count", 0)
                    sample["young_gc_time_ms"] = young.get("collection_time_in_millis", 0)
                    sample["old_gc_count"] = old.get("collection_count", 0)
                    sample["old_gc_time_ms"] = old.get("collection_time_in_millis", 0)
                    sample["write_active"] = n.get("thread_pool", {}).get("write", {}).get("active", 0)
                    sample["write_queue"] = n.get("thread_pool", {}).get("write", {}).get("queue", 0)
                    sample["write_rejected"] = n.get("thread_pool", {}).get("write", {}).get("rejected", 0)
                    sample["index_total"] = n.get("indices", {}).get("indexing", {}).get("index_total", 0)
                    break
            except Exception:
                pass

            self.samples.append(sample)
            self._stop_event.wait(1.0)

def compute_stats(values):
    if not values:
        return {"median": 0, "mean": 0, "std": 0, "min": 0, "max": 0, "cv_pct": 0, "ci95_low": 0, "ci95_high": 0}
    vals = sorted(values)
    n = len(vals)
    median = vals[n // 2] if n % 2 != 0 else (vals[n // 2 - 1] + vals[n // 2]) / 2.0
    mean = sum(vals) / n
    variance = sum((x - mean) ** 2 for x in vals) / (n - 1) if n > 1 else 0.0
    std = math.sqrt(variance)
    cv_pct = (std / mean * 100.0) if mean > 0 else 0.0
    # 95% CI (using t-score for n=5: t_0.025,4 = 2.776, or z=1.96)
    t_val = 2.776 if n == 5 else (1.96 if n > 5 else 4.303)
    margin = t_val * (std / math.sqrt(n)) if n > 1 else 0.0
    return {
        "median": round(median, 2),
        "mean": round(mean, 2),
        "std": round(std, 2),
        "min": round(vals[0], 2),
        "max": round(vals[-1], 2),
        "cv_pct": round(cv_pct, 2),
        "ci95_low": round(mean - margin, 2),
        "ci95_high": round(mean + margin, 2)
    }

def run_single_benchmark(model_key, concurrency, batch_size, docs_count, is_warmup=False, rep=0):
    cfg = MODELS[model_key]
    run_label = f"warmup_{model_key}_c{concurrency}" if is_warmup else f"rep{rep}_{model_key}_c{concurrency}"
    
    # 1. Switch container
    docker("stop", cfg["other"])
    docker("start", cfg["container"])
    wait_ready(cfg["url"])

    # 2. Recreate index
    recreate_index(cfg["url"], cfg["mapping"])

    # 3. Idle Cooldown
    time.sleep(4)

    # 4. Telemetry Collector
    collector = TelemetryCollector(cfg["container"], cfg["url"])
    if not is_warmup:
        collector.start()

    # 5. Run Go bench_runner_final_linux
    runner_bin = wsl_path(ROOT / "benchmarks" / "bench_runner_final_linux")
    dataset_wsl = wsl_path(DATASET)
    dict_wsl = wsl_path(ROOT / "data" / "dicts" / "coccoc_plugin_final")
    raw_out = RUN_DIR / "raw" / f"{run_label}.json"

    url_flag = f"-vanilla_url={cfg['url']}" if model_key == "vanilla" else f"-plugin_url={cfg['url']}"
    cmd = [
        "wsl", "-u", "root", "-d", "Ubuntu", "bash", "-c",
        f"{runner_bin} -mode={cfg['runner_mode']} -concurrency={concurrency} -batch={batch_size} "
        f"-max_docs={docs_count} -dataset={dataset_wsl} -dict={dict_wsl} {url_flag} "
        f"-output={wsl_path(raw_out)}"
    ]

    t_start = time.time()
    proc = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace")
    duration_real = time.time() - t_start

    if not is_warmup:
        collector.stop()

    if proc.returncode != 0 or not raw_out.exists():
        print(f"❌ Runner failed for {run_label}! Stderr:\n{proc.stderr}\nStdout:\n{proc.stdout}")
        raise RuntimeError(f"Runner execution failed for {run_label}")

    raw_data = json.loads(raw_out.read_text(encoding="utf-8"))[0]

    # Verify Bulk Correctness Gate
    count_resp = request_json(cfg["url"] + "/messages/_count")
    indexed_in_es = count_resp.get("count", 0)

    if raw_data.get("bulk_item_errors", 0) > 0 or raw_data.get("http_error_batches", 0) > 0 or indexed_in_es != docs_count:
        raise RuntimeError(
            f"Bulk Correctness Gate VIOLATED on {run_label}: ES count={indexed_in_es}, expected={docs_count}, "
            f"bulk_item_errors={raw_data.get('bulk_item_errors')}, http_errors={raw_data.get('http_error_batches')}"
        )

    # Save Telemetry
    if not is_warmup:
        telem_file = RUN_DIR / "telemetry" / f"{run_label}_telemetry.json"
        telem_file.write_text(json.dumps(collector.samples, indent=2, ensure_ascii=False), encoding="utf-8")

        # Telemetry summary
        cpu_samples = [s.get("container_cpu_pct", 0) for s in collector.samples if "container_cpu_pct" in s]
        heap_samples = [s.get("jvm_heap_used_mb", 0) for s in collector.samples if "jvm_heap_used_mb" in s]
        raw_data["telemetry_summary"] = {
            "avg_container_cpu_pct": round(sum(cpu_samples) / len(cpu_samples), 2) if cpu_samples else 0,
            "peak_container_cpu_pct": round(max(cpu_samples), 2) if cpu_samples else 0,
            "peak_jvm_heap_used_mb": round(max(heap_samples), 2) if heap_samples else 0,
            "samples_count": len(collector.samples)
        }

    raw_data["run_label"] = run_label
    raw_data["repetition"] = rep
    raw_data["wall_time_sec"] = round(duration_real, 2)
    raw_out.write_text(json.dumps(raw_data, indent=2, ensure_ascii=False), encoding="utf-8")

    return raw_data

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--reps", type=int, default=5, help="Số lần lặp mỗi config")
    parser.add_argument("--docs", type=int, default=1000000, help="Số documents mỗi lần chạy")
    parser.add_argument("--warmup", type=int, default=100000, help="Số documents warmup")
    parser.add_argument("--batch", type=int, default=1000, help="Kích thước batch")
    parser.add_argument("--seed", type=int, default=42, help="Random seed")
    parser.add_argument("--concurrencies", nargs="+", type=int, default=[1, 4, 8, 16, 32])
    args = parser.parse_args()

    random.seed(args.seed)
    concurrencies = args.concurrencies
    print(f"=== KHỞI ĐỘNG FULL MATRIX BENCHMARK (1M DOCS) ===")
    print(f"Concurrencies: {concurrencies} | Repetitions: {args.reps} | Docs/run: {args.docs:,} | Warmup: {args.warmup:,}")
    print(f"Seed: {args.seed} | Output Dir: {RUN_DIR}")

    per_run_results = []
    per_run_file = RUN_DIR / "summaries" / "per_run.json"
    aggregate_file = RUN_DIR / "summaries" / "aggregate.json"
    aggregate_csv = RUN_DIR / "summaries" / "aggregate.csv"

    # Pre-generate randomized plan
    plan = []
    for rep in range(1, args.reps + 1):
        shuffled_c = list(concurrencies)
        random.shuffle(shuffled_c)
        for c in shuffled_c:
            # Alternating AB vs BA
            if (rep + c) % 2 == 0:
                pair = [("vanilla", c, rep), ("plugin", c, rep)]
            else:
                pair = [("plugin", c, rep), ("vanilla", c, rep)]
            plan.extend(pair)

    plan_file = RUN_DIR / "environment" / "experimental_plan.json"
    plan_file.write_text(json.dumps([{"step": i+1, "model": m, "concurrency": c, "rep": r} for i, (m, c, r) in enumerate(plan)], indent=2), encoding="utf-8")
    print(f"Plan saved ({len(plan)} total experimental runs).")

    # 1. Warm-up runs (one for Vanilla, one for Plugin)
    print("\n--- [WARM-UP PHASE] Thực hiện warmup 100k docs cho mỗi mô hình ---")
    print("Warmup Model A (Vanilla)...")
    run_single_benchmark("vanilla", 8, args.batch, args.warmup, is_warmup=True)
    print("Warmup Model B (Plugin)...")
    run_single_benchmark("plugin", 8, args.batch, args.warmup, is_warmup=True)
    print("✅ Warm-up hoàn tất!\n")

    # 2. Execute Plan
    total_steps = len(plan)
    for step_idx, (model_key, c, rep) in enumerate(plan):
        step_num = step_idx + 1
        model_name = MODELS[model_key]["name"]
        print(f"\n[{step_num}/{total_steps}] ĐANG CHẠY: {model_name} | C={c} | Repetition {rep}/{args.reps} | Docs={args.docs:,}")
        
        t0 = time.time()
        res = run_single_benchmark(model_key, c, args.batch, args.docs, is_warmup=False, rep=rep)
        elapsed = time.time() - t0

        tps = res.get("docs_per_sec", 0)
        p50 = res.get("latency_p50_ms", 0)
        e2e_p50 = res.get("end_to_end_p50_ms", 0)
        gc_time = res.get("jvm_gc_time_ms", 0)
        print(f"   -> Hoàn thành trong {elapsed:.1f}s | Throughput: {tps:,.1f} docs/s | p50: {p50:.1f}ms | EndToEnd p50: {e2e_p50:.1f}ms | GC pause: {gc_time}ms")

        per_run_results.append(res)
        per_run_file.write_text(json.dumps(per_run_results, indent=2, ensure_ascii=False), encoding="utf-8")

        # Incremental Aggregate computation
        aggregate_data = compute_full_aggregate(per_run_results, concurrencies)
        aggregate_file.write_text(json.dumps(aggregate_data, indent=2, ensure_ascii=False), encoding="utf-8")
        write_aggregate_csv(aggregate_data, aggregate_csv)

    print("\n" + "=" * 80)
    print(f"🎉 TẤT CẢ {total_steps} EXPERIMENTAL RUNS ĐÃ HOÀN TẤT THÀNH CÔNG VỚI 100% QUALITY GATE!")
    print(f"Kết quả lưu tại: {aggregate_file}")
    print("=" * 80)

def compute_full_aggregate(runs, concurrencies):
    grouped = {}
    for r in runs:
        m = "vanilla" if "vanilla" in r["mode"] else "plugin"
        c = r["concurrency"]
        key = (m, c)
        grouped.setdefault(key, []).append(r)

    matrix = []
    for c in sorted(concurrencies):
        v_runs = grouped.get(("vanilla", c), [])
        p_runs = grouped.get(("plugin", c), [])

        def extract(run_list, field):
            return [x[field] for x in run_list if field in x]

        v_tps = compute_stats(extract(v_runs, "docs_per_sec"))
        p_tps = compute_stats(extract(p_runs, "docs_per_sec"))
        
        v_lat_p50 = compute_stats(extract(v_runs, "latency_p50_ms"))
        p_lat_p50 = compute_stats(extract(p_runs, "latency_p50_ms"))

        v_lat_p95 = compute_stats(extract(v_runs, "latency_p95_ms"))
        p_lat_p95 = compute_stats(extract(p_runs, "latency_p95_ms"))

        v_e2e_p50 = compute_stats(extract(v_runs, "end_to_end_p50_ms"))
        p_e2e_p50 = compute_stats(extract(p_runs, "end_to_end_p50_ms"))

        v_gc_time = compute_stats(extract(v_runs, "jvm_gc_time_ms"))
        p_gc_time = compute_stats(extract(p_runs, "jvm_gc_time_ms"))

        v_gc_count = compute_stats(extract(v_runs, "jvm_gc_count"))
        p_gc_count = compute_stats(extract(p_runs, "jvm_gc_count"))

        v_store = compute_stats(extract(v_runs, "disk_store_mb"))
        p_store = compute_stats(extract(p_runs, "disk_store_mb"))

        # Effect sizes: (Model A - Model B) / Model B * 100%
        tps_effect = ((v_tps["median"] - p_tps["median"]) / p_tps["median"] * 100.0) if p_tps["median"] > 0 else 0
        p50_effect = ((v_lat_p50["median"] - p_lat_p50["median"]) / p_lat_p50["median"] * 100.0) if p_lat_p50["median"] > 0 else 0
        gc_effect = ((p_gc_time["median"] - v_gc_time["median"]) / v_gc_time["median"] * 100.0) if v_gc_time["median"] > 0 else 0

        matrix.append({
            "concurrency": c,
            "repetitions_vanilla": len(v_runs),
            "repetitions_plugin": len(p_runs),
            "vanilla": {
                "throughput": v_tps,
                "latency_p50_ms": v_lat_p50,
                "latency_p95_ms": v_lat_p95,
                "end_to_end_p50_ms": v_e2e_p50,
                "jvm_gc_time_ms": v_gc_time,
                "jvm_gc_count": v_gc_count,
                "disk_store_mb": v_store
            },
            "plugin": {
                "throughput": p_tps,
                "latency_p50_ms": p_lat_p50,
                "latency_p95_ms": p_lat_p95,
                "end_to_end_p50_ms": p_e2e_p50,
                "jvm_gc_time_ms": p_gc_time,
                "jvm_gc_count": p_gc_count,
                "disk_store_mb": p_store
            },
            "effect_size": {
                "throughput_diff_percent": round(tps_effect, 2),
                "latency_p50_diff_percent": round(p50_effect, 2),
                "gc_overhead_plugin_percent": round(gc_effect, 2)
            }
        })

    return {
        "run_id": RUN_ID,
        "concurrency_matrix": matrix,
        "updated_at": datetime.datetime.now(datetime.timezone.utc).isoformat()
    }

def write_aggregate_csv(aggregate_data, csv_path):
    rows = []
    for item in aggregate_data.get("concurrency_matrix", []):
        c = item["concurrency"]
        v = item["vanilla"]
        p = item["plugin"]
        eff = item["effect_size"]
        rows.append({
            "concurrency": c,
            "v_tps_median": v["throughput"]["median"],
            "v_tps_ci95_low": v["throughput"]["ci95_low"],
            "v_tps_ci95_high": v["throughput"]["ci95_high"],
            "v_tps_cv_pct": v["throughput"]["cv_pct"],
            "p_tps_median": p["throughput"]["median"],
            "p_tps_ci95_low": p["throughput"]["ci95_low"],
            "p_tps_ci95_high": p["throughput"]["ci95_high"],
            "p_tps_cv_pct": p["throughput"]["cv_pct"],
            "tps_diff_pct": eff["throughput_diff_percent"],
            "v_p50_median": v["latency_p50_ms"]["median"],
            "p_p50_median": p["latency_p50_ms"]["median"],
            "v_e2e_p50_median": v["end_to_end_p50_ms"]["median"],
            "p_e2e_p50_median": p["end_to_end_p50_ms"]["median"],
            "v_gc_time_ms_median": v["jvm_gc_time_ms"]["median"],
            "p_gc_time_ms_median": p["jvm_gc_time_ms"]["median"],
            "v_store_mb_median": v["disk_store_mb"]["median"],
            "p_store_mb_median": p["disk_store_mb"]["median"]
        })

    if rows:
        with open(csv_path, "w", newline="", encoding="utf-8") as f:
            writer = csv.DictWriter(f, fieldnames=list(rows[0].keys()))
            writer.writeheader()
            writer.writerows(rows)

if __name__ == "__main__":
    main()
