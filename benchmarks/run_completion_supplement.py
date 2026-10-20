"""Comprehensive supplement runner for Task 1:
- Phase C: Full container & client resource telemetry (cgroups v2 + docker stats + node stats + time -v)
- Phase D: Batch sensitivity suite (C=1: b500, b1000; C=32: b1000, b2000; 3 reps each, AB/BA)
- Phase E: 30-minute soak test per model (C=16, batch 1000) with windowed drift & slope analysis
- Phase F: Engine-only suite (C=8, C=16; 3 reps; Model A pre-tokenized vs Model B raw)
- Phase G: Search latency suite (C=1, 8, 32 on 1000 queries)
"""

import argparse
import csv
import datetime as dt
import json
import math
import os
import pathlib
import re
import shlex
import statistics
import subprocess
import sys
import threading
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

ROOT = pathlib.Path(__file__).resolve().parents[1]
BASE_RUN = ROOT / "benchmarks" / "final_run" / "final_20261020_093100"
OUT = BASE_RUN / "supplement"
DATASET = ROOT / "benchmarks" / "dataset_real_1m.jsonl"
DICT = ROOT / "data" / "dicts" / "coccoc_plugin_final"
RUNNER = ROOT / "benchmarks" / "bench_runner_final_linux"
QUERIES_FILE = BASE_RUN / "search" / "search_queries_tokenized.json"

MODELS = {
    "vanilla": {
        "url": "http://localhost:9201",
        "container": "os_vanilla_bench",
        "other": "os_plugin_bench",
        "mapping": ROOT / "benchmarks" / "mapping_vanilla.json",
    },
    "plugin": {
        "url": "http://localhost:9202",
        "container": "os_plugin_bench",
        "other": "os_vanilla_bench",
        "mapping": ROOT / "benchmarks" / "mapping_plugin.json",
    },
}

ANSI_REGEX = re.compile(r"\x1b\[[0-9;]*[a-zA-Z]")


def wsl_path(path):
    p = pathlib.Path(path).resolve()
    drive = p.drive[0].lower()
    posix_tail = p.as_posix().split(":", 1)[1]
    return f"/mnt/{drive}{posix_tail}"


def wsl(*args, check=True, timeout=None):
    cmd = ["wsl.exe", "-u", "root", "-d", "Ubuntu", "--", *args]
    proc = subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
        timeout=timeout,
    )
    if check and proc.returncode != 0:
        raise RuntimeError(f"WSL command failed ({proc.returncode}): {' '.join(args)}\n{proc.stderr}")
    return proc


def request_json(url, method="GET", payload=None, timeout=10):
    body = None if payload is None else json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=body,
        method=method,
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read().decode("utf-8"))


def wait_ready(url, timeout=180):
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            status = request_json(url + "/_cluster/health", timeout=3).get("status")
            if status in {"yellow", "green"}:
                return
        except Exception:
            pass
        time.sleep(2)
    raise TimeoutError(f"OpenSearch at {url} did not become ready within {timeout}s")


def recreate_index(model):
    cfg = MODELS[model]
    try:
        request_json(cfg["url"] + "/messages", method="DELETE")
    except urllib.error.HTTPError as err:
        if err.code != 404:
            raise
    mapping = json.loads(cfg["mapping"].read_text(encoding="utf-8"))
    request_json(cfg["url"] + "/messages", method="PUT", payload=mapping)
    wait_ready(cfg["url"])


def switch_model(model):
    cfg = MODELS[model]
    # Ensure rival container is stopped
    wsl("docker", "stop", cfg["other"], check=False, timeout=60)
    # Start target container
    wsl("docker", "start", cfg["container"], timeout=60)
    wait_ready(cfg["url"])
    recreate_index(model)
    time.sleep(3)


def get_container_cgroup_dir(container):
    cid = wsl("docker", "inspect", "--format", "{{.Id}}", container).stdout.strip()
    return f"/sys/fs/cgroup/system.slice/docker-{cid}.scope"


def read_cgroup_snapshot(cgroup_dir):
    try:
        cmd = f"cat {cgroup_dir}/cpu.stat; echo '---SEP---'; cat {cgroup_dir}/memory.current; echo '---SEP---'; cat {cgroup_dir}/memory.stat"
        res = wsl("bash", "-c", cmd, check=False, timeout=5)
        if res.returncode != 0:
            return {}
        parts = res.stdout.split("---SEP---")
        cpu_lines = parts[0].strip().splitlines() if len(parts) > 0 else []
        mem_cur = parts[1].strip() if len(parts) > 1 else "0"
        mem_lines = parts[2].strip().splitlines() if len(parts) > 2 else []

        cpu_d = {}
        for l in cpu_lines:
            items = l.strip().split()
            if len(items) == 2:
                cpu_d[items[0]] = int(items[1])

        mem_d = {}
        for l in mem_lines:
            items = l.strip().split()
            if len(items) == 2:
                mem_d[items[0]] = int(items[1])

        return {
            "timestamp": time.time(),
            "usage_usec": cpu_d.get("usage_usec", 0),
            "user_usec": cpu_d.get("user_usec", 0),
            "system_usec": cpu_d.get("system_usec", 0),
            "nr_periods": cpu_d.get("nr_periods", 0),
            "nr_throttled": cpu_d.get("nr_throttled", 0),
            "throttled_usec": cpu_d.get("throttled_usec", 0),
            "memory_current_bytes": int(mem_cur) if mem_cur.isdigit() else 0,
            "memory_rss_bytes": mem_d.get("anon", 0),
            "memory_file_bytes": mem_d.get("file", 0),
        }
    except Exception as e:
        return {"error": str(e)}


def to_mib(value_str):
    if not value_str:
        return None
    match = re.match(r"\s*([0-9.]+)\s*([KMGTP]?i?B)", value_str)
    if not match:
        return None
    num, unit = float(match.group(1)), match.group(2)
    factors = {
        "B": 1 / 1048576,
        "KB": 1 / 1024,
        "KiB": 1 / 1024,
        "MB": 1,
        "MiB": 1,
        "GB": 1024,
        "GiB": 1024,
        "TB": 1048576,
        "TiB": 1048576,
    }
    return num * factors.get(unit, 1.0)


class TelemetryCollector:
    def __init__(self, model):
        self.model = model
        self.cfg = MODELS[model]
        self.cgroup_dir = get_container_cgroup_dir(self.cfg["container"])
        self.docker_samples = []
        self.node_samples = []
        self.cgroup_samples = []
        self.errors = []
        self.stop_event = threading.Event()
        self.stats_process = None
        self.start_time = None
        self.cgroup_start = None
        self.cgroup_end = None

    def start(self):
        self.start_time = time.time()
        self.cgroup_start = read_cgroup_snapshot(self.cgroup_dir)

        # Launch streaming docker stats
        command = [
            "wsl.exe",
            "-u",
            "root",
            "-d",
            "Ubuntu",
            "--",
            "docker",
            "stats",
            "--format",
            "{{json .}}",
            self.cfg["container"],
        ]
        self.stats_process = subprocess.Popen(
            command,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            encoding="utf-8",
            errors="replace",
        )
        threading.Thread(target=self._stream_docker_stats, daemon=True).start()
        threading.Thread(target=self._poll_node_and_cgroup, daemon=True).start()

    def _stream_docker_stats(self):
        try:
            for line in self.stats_process.stdout:
                if self.stop_event.is_set():
                    break
                clean = ANSI_REGEX.sub("", line).strip()
                if clean.startswith("{") and clean.endswith("}"):
                    try:
                        item = json.loads(clean)
                        mem_usage_str = item.get("MemUsage", "").split("/")[0].strip()
                        self.docker_samples.append({
                            "timestamp": time.time(),
                            "cpu_pct": float(item.get("CPUPerc", "0%").rstrip("%")),
                            "memory_rss_mib": to_mib(mem_usage_str),
                            "mem_percent": float(item.get("MemPerc", "0%").rstrip("%")),
                            "block_io": item.get("BlockIO"),
                            "net_io": item.get("NetIO"),
                            "pids": int(item.get("PIDs", 0)),
                        })
                    except Exception as err:
                        self.errors.append(f"docker_parse_err: {err}")
        except Exception as err:
            self.errors.append(f"docker_stream_err: {err}")

    def _poll_node_and_cgroup(self):
        while not self.stop_event.is_set():
            t_loop_start = time.monotonic()
            try:
                # 1. OpenSearch node stats
                nodes_data = request_json(
                    self.cfg["url"] + "/_nodes/stats/jvm,process,thread_pool,indices,indexing_pressure",
                    timeout=4,
                ).get("nodes", {})
                node = next(iter(nodes_data.values())) if nodes_data else {}

                young = node.get("jvm", {}).get("gc", {}).get("collectors", {}).get("young", {})
                old = node.get("jvm", {}).get("gc", {}).get("collectors", {}).get("old", {})
                write_tp = node.get("thread_pool", {}).get("write", {})
                pressure = node.get("indexing_pressure", {}).get("memory", {})

                node_sample = {
                    "timestamp": time.time(),
                    "process_cpu_pct": node.get("process", {}).get("cpu", {}).get("percent"),
                    "jvm_heap_used_mib": node.get("jvm", {}).get("mem", {}).get("heap_used_in_bytes", 0) / 1048576.0,
                    "young_gc_count": young.get("collection_count", 0),
                    "young_gc_collection_time_ms": young.get("collection_time_in_millis", 0),
                    "old_gc_count": old.get("collection_count", 0),
                    "old_gc_collection_time_ms": old.get("collection_time_in_millis", 0),
                    "write_active": write_tp.get("active", 0),
                    "write_queue": write_tp.get("queue", 0),
                    "write_rejected": write_tp.get("rejected", 0),
                    "indexing_pressure_bytes": pressure.get("current", {}).get("combined_coordinating_and_primary_in_bytes", 0),
                    "indexing_pressure_rejections": pressure.get("total", {}).get("coordinating_rejections", 0),
                }
                self.node_samples.append(node_sample)

                # 2. Cgroup periodic snapshot
                cg_snap = read_cgroup_snapshot(self.cgroup_dir)
                if cg_snap and "usage_usec" in cg_snap:
                    self.cgroup_samples.append(cg_snap)

            except Exception as err:
                self.errors.append(f"poll_err: {err}")

            elapsed = time.monotonic() - t_loop_start
            self.stop_event.wait(max(0.0, 1.0 - elapsed))

    def stop(self):
        self.stop_event.set()
        if self.stats_process:
            self.stats_process.terminate()
            try:
                self.stats_process.wait(timeout=3)
            except subprocess.TimeoutExpired:
                self.stats_process.kill()
        self.cgroup_end = read_cgroup_snapshot(self.cgroup_dir)

    def summary(self):
        docker_cpus = [s["cpu_pct"] for s in self.docker_samples if "cpu_pct" in s]
        docker_rsss = [s["memory_rss_mib"] for s in self.docker_samples if s.get("memory_rss_mib") is not None]
        node_heaps = [s["jvm_heap_used_mib"] for s in self.node_samples if "jvm_heap_used_mib" in s]
        node_cpus = [s["process_cpu_pct"] for s in self.node_samples if s.get("process_cpu_pct") is not None]

        # Cgroup deltas
        cg_cpu_sec = 0.0
        cg_throttled_sec = 0.0
        cg_nr_throttled = 0
        if self.cgroup_start and self.cgroup_end and "usage_usec" in self.cgroup_start and "usage_usec" in self.cgroup_end:
            cg_cpu_sec = (self.cgroup_end["usage_usec"] - self.cgroup_start["usage_usec"]) / 1e6
            cg_throttled_sec = (self.cgroup_end["throttled_usec"] - self.cgroup_start["throttled_usec"]) / 1e6
            cg_nr_throttled = self.cgroup_end["nr_throttled"] - self.cgroup_start["nr_throttled"]

        # GC deltas
        young_gc_delta = 0
        young_gc_time_delta = 0
        old_gc_delta = 0
        old_gc_time_delta = 0
        write_rejected_delta = 0
        write_queue_peak = 0
        if len(self.node_samples) >= 2:
            first, last = self.node_samples[0], self.node_samples[-1]
            young_gc_delta = last["young_gc_count"] - first["young_gc_count"]
            young_gc_time_delta = last["young_gc_collection_time_ms"] - first["young_gc_collection_time_ms"]
            old_gc_delta = last["old_gc_count"] - first["old_gc_count"]
            old_gc_time_delta = last["old_gc_collection_time_ms"] - first["old_gc_collection_time_ms"]
            write_rejected_delta = last["write_rejected"] - first["write_rejected"]
            write_queue_peak = max(s["write_queue"] for s in self.node_samples)

        # Expected sample intervals
        total_time = (time.time() - self.start_time) if self.start_time else 1.0
        expected_samples = max(1, int(round(total_time)))
        actual_samples = len(self.node_samples)
        missing_rate_pct = max(0.0, (1.0 - (actual_samples / expected_samples)) * 100.0)

        return {
            "sample_counts": {
                "docker_stats_samples": len(self.docker_samples),
                "opensearch_node_samples": len(self.node_samples),
                "cgroup_samples": len(self.cgroup_samples),
                "expected_samples_1s": expected_samples,
                "missing_samples_rate_pct": round(missing_rate_pct, 2),
                "collector_errors_count": len(self.errors),
                "collector_errors": self.errors[:10],
            },
            "container_cpu": {
                "avg_pct": round(statistics.mean(docker_cpus), 2) if docker_cpus else 0.0,
                "peak_pct": round(max(docker_cpus), 2) if docker_cpus else 0.0,
                "cgroup_total_cpu_seconds": round(cg_cpu_sec, 3),
            },
            "container_throttling": {
                "cgroup_nr_throttled_delta": cg_nr_throttled,
                "cgroup_throttled_seconds": round(cg_throttled_sec, 4),
            },
            "container_memory": {
                "rss_avg_mib": round(statistics.mean(docker_rsss), 2) if docker_rsss else 0.0,
                "rss_peak_mib": round(max(docker_rsss), 2) if docker_rsss else 0.0,
            },
            "jvm_memory": {
                "heap_avg_mib": round(statistics.mean(node_heaps), 2) if node_heaps else 0.0,
                "heap_peak_mib": round(max(node_heaps), 2) if node_heaps else 0.0,
            },
            "jvm_gc_collections": {
                "young_gc_count_delta": young_gc_delta,
                "young_gc_collection_time_ms_delta": young_gc_time_delta,
                "old_gc_count_delta": old_gc_delta,
                "old_gc_collection_time_ms_delta": old_gc_time_delta,
                "total_gc_collection_time_ms_delta": young_gc_time_delta + old_gc_time_delta,
            },
            "opensearch_process": {
                "process_cpu_avg_pct": round(statistics.mean(node_cpus), 2) if node_cpus else 0.0,
                "process_cpu_peak_pct": round(max(node_cpus), 2) if node_cpus else 0.0,
                "write_queue_peak": write_queue_peak,
                "write_rejected_delta": write_rejected_delta,
            },
        }


def parse_time_file(path):
    if not path.exists():
        return {"user_time_sec": 0.0, "system_time_sec": 0.0, "cpu_percent": 0.0, "max_rss_kib": 0}
    vals = {}
    for line in path.read_text(encoding="utf-8", errors="replace").splitlines():
        if ": " in line:
            k, v = line.strip().split(": ", 1)
            vals[k] = v
    user_sec = float(vals.get("User time (seconds)", 0.0))
    sys_sec = float(vals.get("System time (seconds)", 0.0))
    cpu_pct = float(vals.get("Percent of CPU this job got", "0%").rstrip("%"))
    rss_kib = int(vals.get("Maximum resident set size (kbytes)", 0))
    return {
        "user_time_sec": round(user_sec, 3),
        "system_time_sec": round(sys_sec, 3),
        "total_client_cpu_sec": round(user_sec + sys_sec, 3),
        "cpu_percent": round(cpu_pct, 1),
        "max_rss_mib": round(rss_kib / 1024.0, 2),
    }


def inspect_container_state(container):
    raw = wsl("docker", "inspect", "--format", "{{json .State}}", container).stdout.strip()
    state = json.loads(raw)
    restarts = int(wsl("docker", "inspect", "--format", "{{.RestartCount}}", container).stdout.strip())
    return {
        "status": state.get("Status"),
        "oom_killed": state.get("OOMKilled", False),
        "exit_code": state.get("ExitCode", 0),
        "restart_count": restarts,
    }


def run_benchmark_case(
    suite_name,
    label,
    model,
    concurrency,
    batch,
    docs,
    duration_sec=0,
    runner_mode=None,
    out_dir=None,
):
    if out_dir is None:
        out_dir = OUT
    cfg = MODELS[model]
    switch_model(model)

    raw_file = out_dir / "raw" / f"{label}.json"
    telemetry_file = out_dir / "telemetry" / f"{label}.json"
    time_file = out_dir / "client" / f"{label}_time.txt"
    log_file = out_dir / "logs" / f"{label}.log"

    # Re-use already completed run if valid
    if raw_file.exists():
        try:
            cached = json.loads(raw_file.read_text(encoding="utf-8"))
            if isinstance(cached, dict) and "throughput_docs_per_sec" in cached and "container_telemetry" in cached:
                print(f"⏩ [REUSE] {label}: {cached['throughput_docs_per_sec']:.1f} docs/s already completed", flush=True)
                return cached
        except Exception:
            pass

    state_before = inspect_container_state(cfg["container"])
    telemetry = TelemetryCollector(model)
    telemetry.start()

    actual_mode = runner_mode if runner_mode else model
    args = [
        wsl_path(RUNNER),
        f"-mode={actual_mode}",
        f"-concurrency={concurrency}",
        f"-batch={batch}",
        f"-max_docs={docs}",
        f"-duration_sec={duration_sec}",
        f"-dataset={wsl_path(DATASET)}",
        f"-dict={wsl_path(DICT)}",
        f"-{model}_url={cfg['url']}",
        f"-output={wsl_path(raw_file)}",
    ]

    shell_cmd = (
        f"/usr/bin/time -v -o {shlex.quote(wsl_path(time_file))} "
        f"{' '.join(shlex.quote(x) for x in args)} > {shlex.quote(wsl_path(log_file))} 2>&1"
    )

    t_start = time.time()
    proc = wsl("bash", "-lc", shell_cmd, check=False, timeout=max(600, duration_sec + 300))
    telemetry.stop()
    wall_sec = time.time() - t_start
    state_after = inspect_container_state(cfg["container"])

    if proc.returncode != 0 or not raw_file.exists():
        err_msg = log_file.read_text(encoding="utf-8", errors="replace")[-500:] if log_file.exists() else proc.stderr
        raise RuntimeError(f"Run {label} FAILED (code {proc.returncode}):\n{err_msg}")

    res_raw = json.loads(raw_file.read_text(encoding="utf-8"))
    item = res_raw[0] if isinstance(res_raw, list) else res_raw

    # Quality Gate Checks
    http_errors = item.get("http_error_batches", 0)
    bulk_errors = item.get("bulk_item_errors", 0)
    tok_errors = item.get("tokenizer_errors", 0)
    write_rej = item.get("write_rejected", 0)
    total_err = item.get("errors", 0)

    quality_gate_passed = True
    gate_issues = []
    if http_errors > 0:
        quality_gate_passed = False
        gate_issues.append(f"http_errors={http_errors} (timeout/saturation)")
    if bulk_errors > 0:
        quality_gate_passed = False
        gate_issues.append(f"bulk_errors={bulk_errors}")
    if tok_errors > 0:
        quality_gate_passed = False
        gate_issues.append(f"tok_errors={tok_errors}")
    if write_rej > 0:
        quality_gate_passed = False
        gate_issues.append(f"write_rejected={write_rej} (queue saturation)")
    if total_err > 0:
        quality_gate_passed = False
        gate_issues.append(f"total_errors={total_err}")

    if duration_sec == 0:
        # Verify exact document count in OpenSearch
        try:
            count_res = request_json(cfg["url"] + "/messages/_count")
            actual_count = count_res.get("count", 0)
            if actual_count != docs:
                quality_gate_passed = False
                gate_issues.append(f"doc_count_mismatch: expected {docs}, got {actual_count}")
        except Exception as e:
            gate_issues.append(f"count_check_err={e}")

    if state_after["oom_killed"] or state_after["restart_count"] != state_before["restart_count"]:
        raise RuntimeError(f"Container crash gate FAILED for {label}: {state_before} -> {state_after}")

    client_res = parse_time_file(time_file)
    telemetry_sum = telemetry.summary()

    # Total-system accounting:
    # client_cpu_sec (from getrusage via /usr/bin/time) + container_cpu_sec (from cgroups v2 usage_usec)
    container_cpu_sec = telemetry_sum["container_cpu"]["cgroup_total_cpu_seconds"]
    client_cpu_sec = client_res["total_client_cpu_sec"]
    total_system_cpu_sec = round(container_cpu_sec + client_cpu_sec, 3)

    combined_result = {
        "suite": suite_name,
        "label": label,
        "model": model,
        "runner_mode": actual_mode,
        "concurrency": concurrency,
        "batch_size": batch,
        "docs_target": docs,
        "duration_target_sec": duration_sec,
        "wall_time_sec": round(wall_sec, 3),
        "docs_indexed": item.get("total_docs", 0),
        "throughput_docs_per_sec": round(item.get("docs_per_sec", 0.0), 2),
        "mb_per_sec": round(item.get("mb_per_sec", 0.0), 2),
        "latency_p50_ms": round(item.get("latency_p50_ms", 0.0), 2),
        "latency_p95_ms": round(item.get("latency_p95_ms", 0.0), 2),
        "latency_p99_ms": round(item.get("latency_p99_ms", 0.0), 2),
        "latency_max_ms": round(item.get("latency_max_ms", 0.0), 2),
        "end_to_end_p50_ms": round(item.get("end_to_end_p50_ms", 0.0), 2),
        "end_to_end_p95_ms": round(item.get("end_to_end_p95_ms", 0.0), 2),
        "quality_gate_passed": quality_gate_passed,
        "quality_gate_issues": gate_issues,
        "client_resource": client_res,
        "container_telemetry": telemetry_sum,
        "total_system_resource": {
            "client_cpu_seconds": client_cpu_sec,
            "container_cgroup_cpu_seconds": container_cpu_sec,
            "total_system_cpu_seconds": total_system_cpu_sec,
            "accounting_method": "client_rusage_cpu_sec + container_cgroup_v2_usage_usec/1e6",
        },
        "container_state_before": state_before,
        "container_state_after": state_after,
        "raw_bench_result": item,
    }

    # Write combined and raw telemetry
    raw_file.write_text(json.dumps(combined_result, ensure_ascii=False, indent=2), encoding="utf-8")
    telemetry_file.write_text(
        json.dumps(
            {
                "docker_stats": telemetry.docker_samples,
                "opensearch_node": telemetry.node_samples,
                "cgroup_samples": telemetry.cgroup_samples,
                "collector_errors": telemetry.errors,
            },
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )

    print(
        f"✅ [{suite_name.upper()}] {label}: {combined_result['throughput_docs_per_sec']:.1f} docs/s | "
        f"Wall: {wall_sec:.1f}s | Client CPU: {client_cpu_sec:.1f}s | Container CPU: {container_cpu_sec:.1f}s | "
        f"Throttled: {telemetry_sum['container_throttling']['cgroup_throttled_seconds']:.2f}s | "
        f"Cont. RSS Peak: {telemetry_sum['container_memory']['rss_peak_mib']:.1f}MB",
        flush=True,
    )
    return combined_result


# -----------------------------------------------------------------------------
# SUITE 1: BATCH SENSITIVITY (PHASE D)
# -----------------------------------------------------------------------------
def run_batch_sensitivity_suite(reps=3):
    print("\n" + "=" * 80)
    print("🚀 BẮT ĐẦU BATCH SENSITIVITY SUITE (PHASE D)")
    print(f"Matrix: C=1 (100k docs, b500, b1000) & C=32 (1M docs, b1000, b2000) × {reps} reps")
    print("=" * 80)

    cases = [
        (1, 500, 100000),
        (1, 1000, 100000),
        (32, 1000, 1000000),
        (32, 2000, 1000000),
    ]

    all_rows = []
    for rep in range(1, reps + 1):
        for concurrency, batch, docs in cases:
            # Alternate order to eliminate systematic first-mover bias
            models_order = ["vanilla", "plugin"] if (rep + concurrency + batch) % 2 == 0 else ["plugin", "vanilla"]
            for model in models_order:
                lbl = f"batch_rep{rep}_{model}_c{concurrency}_b{batch}"
                res = run_benchmark_case("batch_sensitivity", lbl, model, concurrency, batch, docs)
                all_rows.append(res)

    # Summarize batch sensitivity
    summaries = []
    for concurrency, batch, docs in cases:
        for model in ["vanilla", "plugin"]:
            group = [r for r in all_rows if r["concurrency"] == concurrency and r["batch_size"] == batch and r["model"] == model]
            if not group:
                continue
            thrs = [r["throughput_docs_per_sec"] for r in group]
            p95s = [r["latency_p95_ms"] for r in group]
            cont_cpus = [r["container_telemetry"]["container_cpu"]["avg_pct"] for r in group]
            cont_rsss = [r["container_telemetry"]["container_memory"]["rss_peak_mib"] for r in group]
            client_cpus = [r["client_resource"]["total_client_cpu_sec"] for r in group]
            client_rsss = [r["client_resource"]["max_rss_mib"] for r in group]
            tot_sys_cpus = [r["total_system_resource"]["total_system_cpu_seconds"] for r in group]
            throttles = [r["container_telemetry"]["container_throttling"]["cgroup_throttled_seconds"] for r in group]
            gc_times = [r["container_telemetry"]["jvm_gc_collections"]["total_gc_collection_time_ms_delta"] for r in group]

            summaries.append({
                "model": model,
                "concurrency": concurrency,
                "batch_size": batch,
                "docs": docs,
                "repetitions": len(group),
                "throughput_median": round(statistics.median(thrs), 2),
                "throughput_mean": round(statistics.mean(thrs), 2),
                "latency_p95_median_ms": round(statistics.median(p95s), 2),
                "container_cpu_avg_pct_median": round(statistics.median(cont_cpus), 2),
                "container_rss_peak_mib_median": round(statistics.median(cont_rsss), 2),
                "container_cgroup_throttled_sec_median": round(statistics.median(throttles), 3),
                "client_cpu_sec_median": round(statistics.median(client_cpus), 2),
                "client_max_rss_mib_median": round(statistics.median(client_rsss), 2),
                "total_system_cpu_sec_median": round(statistics.median(tot_sys_cpus), 2),
                "jvm_gc_collection_time_ms_median": round(statistics.median(gc_times), 1),
            })

    out_json = OUT / "summaries" / "batch_sensitivity_summary.json"
    out_csv = OUT / "summaries" / "batch_sensitivity_summary.csv"
    out_json.write_text(json.dumps(summaries, ensure_ascii=False, indent=2), encoding="utf-8")
    if summaries:
        with out_csv.open("w", newline="", encoding="utf-8") as f:
            w = csv.DictWriter(f, fieldnames=list(summaries[0].keys()))
            w.writeheader()
            w.writerows(summaries)
    print(f"\n📊 Báo cáo Batch Sensitivity đã ghi vào {out_json} và {out_csv}")
    return summaries


# -----------------------------------------------------------------------------
# SUITE 2: ENGINE-ONLY BENCHMARK (PHASE F STEP 4)
# -----------------------------------------------------------------------------
def run_engine_only_suite(reps=3):
    print("\n" + "=" * 80)
    print("🚀 BẮT ĐẦU ENGINE-ONLY SUITE (PHASE F STEP 4)")
    print(f"Matrix: C=8 & C=16, 100k docs, batch 1000 × {reps} reps")
    print("Model A dùng vanilla_pretokenized (ngoài timing); Model B dùng plugin_raw")
    print("=" * 80)

    cases = [(8, 1000, 100000), (16, 1000, 100000)]
    all_rows = []

    for rep in range(1, reps + 1):
        for concurrency, batch, docs in cases:
            order = [("vanilla", "vanilla_pretokenized"), ("plugin", "plugin")] if (rep + concurrency) % 2 == 0 else [("plugin", "plugin"), ("vanilla", "vanilla_pretokenized")]
            for model, runner_mode in order:
                lbl = f"engine_only_rep{rep}_{model}_c{concurrency}_b{batch}"
                res = run_benchmark_case(
                    "engine_only",
                    lbl,
                    model,
                    concurrency,
                    batch,
                    docs,
                    runner_mode=runner_mode,
                )
                all_rows.append(res)

    summaries = []
    for concurrency, batch, docs in cases:
        v_group = [r for r in all_rows if r["concurrency"] == concurrency and r["model"] == "vanilla"]
        p_group = [r for r in all_rows if r["concurrency"] == concurrency and r["model"] == "plugin"]

        v_thr = [r["throughput_docs_per_sec"] for r in v_group]
        p_thr = [r["throughput_docs_per_sec"] for r in p_group]

        summaries.append({
            "concurrency": concurrency,
            "batch_size": batch,
            "docs": docs,
            "repetitions": len(v_group),
            "vanilla_pretokenized_throughput_median": round(statistics.median(v_thr), 2),
            "plugin_raw_throughput_median": round(statistics.median(p_thr), 2),
            "engine_throughput_ratio_vanilla_vs_plugin": round(statistics.median(v_thr) / statistics.median(p_thr), 3) if statistics.median(p_thr) else 0.0,
            "vanilla_container_cpu_median_pct": round(statistics.median(r["container_telemetry"]["container_cpu"]["avg_pct"] for r in v_group), 2),
            "plugin_container_cpu_median_pct": round(statistics.median(r["container_telemetry"]["container_cpu"]["avg_pct"] for r in p_group), 2),
            "vanilla_gc_time_ms_median": round(statistics.median(r["container_telemetry"]["jvm_gc_collections"]["total_gc_collection_time_ms_delta"] for r in v_group), 1),
            "plugin_gc_time_ms_median": round(statistics.median(r["container_telemetry"]["jvm_gc_collections"]["total_gc_collection_time_ms_delta"] for r in p_group), 1),
            "note": "Engine-only isolates engine indexing cost by pre-tokenizing Model A corpus in RAM outside timing loop. Shows pure OpenSearch ingestion rate.",
        })

    out_json = OUT / "summaries" / "engine_only_summary.json"
    out_csv = OUT / "summaries" / "engine_only_summary.csv"
    out_json.write_text(json.dumps(summaries, ensure_ascii=False, indent=2), encoding="utf-8")
    if summaries:
        with out_csv.open("w", newline="", encoding="utf-8") as f:
            w = csv.DictWriter(f, fieldnames=list(summaries[0].keys()))
            w.writeheader()
            w.writerows(summaries)
    print(f"\n📊 Báo cáo Engine-Only Suite đã ghi vào {out_json} và {out_csv}")
    return summaries


# -----------------------------------------------------------------------------
# SUITE 3: SEARCH LATENCY SUITE (PHASE G)
# -----------------------------------------------------------------------------
def run_search_latency_suite():
    print("\n" + "=" * 80)
    print("🚀 BẮT ĐẦU SEARCH LATENCY SUITE (PHASE G)")
    print("Đo Query Latency theo Concurrency C={1, 8, 32} trên 1.000 queries")
    print("=" * 80)

    if not QUERIES_FILE.exists():
        raise FileNotFoundError(f"Queries file not found at {QUERIES_FILE}")
    queries = json.loads(QUERIES_FILE.read_text(encoding="utf-8"))
    print(f"Loaded {len(queries)} queries from {QUERIES_FILE}")

    concurrencies = [1, 8, 32]
    summaries = []

    for model in ["vanilla", "plugin"]:
        switch_model(model)
        cfg = MODELS[model]

        # Warm-up 100 queries
        print(f"🔥 Warming up {model} search index...")
        for q in queries[:100]:
            query_term = q["cgo"] if model == "vanilla" else q["raw"]
            body = {"query": {"multi_match": {"query": query_term, "fields": ["text_vi^2", "text", "text_edge"]}}}
            try:
                request_json(cfg["url"] + "/messages/_search?size=10", method="POST", payload=body, timeout=5)
            except Exception:
                pass

        for c in concurrencies:
            latencies_ms = []
            errors = 0
            t_search_start = time.time()

            def execute_single_query(q_item):
                term = q_item["cgo"] if model == "vanilla" else q_item["raw"]
                body = {
                    "query": {
                        "multi_match": {
                            "query": term,
                            "fields": ["text_vi^2", "text", "text_raw", "text_edge"],
                        }
                    }
                }
                t0 = time.time()
                try:
                    res = request_json(cfg["url"] + "/messages/_search?size=10", method="POST", payload=body, timeout=10)
                    lat = (time.time() - t0) * 1000.0
                    return lat, False
                except Exception:
                    return 0.0, True

            with ThreadPoolExecutor(max_workers=c) as executor:
                results = list(executor.map(execute_single_query, queries))

            for lat, err in results:
                if err:
                    errors += 1
                else:
                    latencies_ms.append(lat)

            wall_sec = time.time() - t_search_start
            qps = len(latencies_ms) / wall_sec if wall_sec else 0.0
            latencies_ms.sort()

            def pct(arr, p):
                if not arr:
                    return 0.0
                idx = min(len(arr) - 1, max(0, int(round(p * (len(arr) - 1)))))
                return arr[idx]

            summary_item = {
                "model": model,
                "concurrency": c,
                "total_queries": len(queries),
                "successful_queries": len(latencies_ms),
                "errors": errors,
                "wall_time_sec": round(wall_sec, 2),
                "throughput_qps": round(qps, 1),
                "latency_p50_ms": round(pct(latencies_ms, 0.50), 2),
                "latency_p90_ms": round(pct(latencies_ms, 0.90), 2),
                "latency_p95_ms": round(pct(latencies_ms, 0.95), 2),
                "latency_p99_ms": round(pct(latencies_ms, 0.99), 2),
                "latency_max_ms": round(latencies_ms[-1], 2) if latencies_ms else 0.0,
            }
            summaries.append(summary_item)
            print(
                f"   [{model.upper()} C={c}] QPS={qps:.1f} | p50={summary_item['latency_p50_ms']}ms | "
                f"p95={summary_item['latency_p95_ms']}ms | p99={summary_item['latency_p99_ms']}ms | Errors={errors}"
            )

    out_json = OUT / "summaries" / "search_latency_summary.json"
    out_csv = OUT / "summaries" / "search_latency_summary.csv"
    out_json.write_text(json.dumps(summaries, ensure_ascii=False, indent=2), encoding="utf-8")
    if summaries:
        with out_csv.open("w", newline="", encoding="utf-8") as f:
            w = csv.DictWriter(f, fieldnames=list(summaries[0].keys()))
            w.writeheader()
            w.writerows(summaries)
    print(f"\n📊 Báo cáo Search Latency đã ghi vào {out_json} và {out_csv}")
    return summaries


# -----------------------------------------------------------------------------
# SUITE 4: 30-MINUTE SOAK TEST (PHASE E)
# -----------------------------------------------------------------------------
def run_soak_suite(duration_sec=1800):
    print("\n" + "=" * 80)
    print(f"🚀 BẮT ĐẦU 30-PHÚT SOAK TEST (PHASE E): {duration_sec}s/mô hình tại C=16, Batch=1000")
    print("=" * 80)

    soak_results = []
    # Test sequentially: Model A then Model B
    for model in ["vanilla", "plugin"]:
        lbl = f"soak_{model}_c16_{duration_sec}s"
        print(f"\n--- Bắt đầu Soak Test [{model}] thời lượng {duration_sec}s (C=16, Batch=1000) ---")
        res = run_benchmark_case(
            "soak_test",
            lbl,
            model,
            concurrency=16,
            batch=1000,
            docs=1000000,
            duration_sec=duration_sec,
        )

        # Process window analysis (5-minute tumbling windows)
        telemetry_file = OUT / "telemetry" / f"{lbl}.json"
        telem_data = json.loads(telemetry_file.read_text(encoding="utf-8"))
        node_samples = telem_data.get("opensearch_node", [])
        docker_samples = telem_data.get("docker_stats", [])

        # Window analysis
        window_sec = 300  # 5 minutes
        num_windows = max(1, int(round(duration_sec / window_sec)))
        window_stats = []

        t0 = docker_samples[0]["timestamp"] if docker_samples else (node_samples[0]["timestamp"] if node_samples else (time.time() - duration_sec))
        for w_idx in range(num_windows):
            w_start = t0 + w_idx * window_sec
            w_end = w_start + window_sec
            # Filter samples in window
            w_node = [s for s in node_samples if w_start <= s["timestamp"] < w_end]
            w_docker = [s for s in docker_samples if w_start <= s["timestamp"] < w_end]

            w_cpu = statistics.mean([s["cpu_pct"] for s in w_docker]) if w_docker else 0.0
            w_rss = statistics.mean([s["memory_rss_mib"] for s in w_docker if s.get("memory_rss_mib")]) if w_docker else 0.0
            w_heap = statistics.mean([s["jvm_heap_used_mib"] for s in w_node]) if w_node else 0.0
            w_gc_time = (w_node[-1]["young_gc_collection_time_ms"] - w_node[0]["young_gc_collection_time_ms"]) if len(w_node) >= 2 else 0
            w_queue_peak = max([s["write_queue"] for s in w_node]) if w_node else 0
            w_rej_delta = (w_node[-1]["write_rejected"] - w_node[0]["write_rejected"]) if len(w_node) >= 2 else 0

            window_stats.append({
                "window_index": w_idx + 1,
                "window_time_range_min": f"{w_idx*5}-{(w_idx+1)*5}m",
                "avg_container_cpu_pct": round(w_cpu, 2),
                "avg_container_rss_mib": round(w_rss, 2),
                "avg_jvm_heap_mib": round(w_heap, 2),
                "young_gc_time_ms_delta": w_gc_time,
                "write_queue_peak": w_queue_peak,
                "write_rejected_delta": w_rej_delta,
                "node_samples": len(w_node),
                "docker_samples": len(w_docker),
            })

        # Calculate slope (MiB/min) for RSS and Heap
        rss_start = window_stats[0]["avg_container_rss_mib"] if window_stats else 0.0
        rss_end = window_stats[-1]["avg_container_rss_mib"] if window_stats else 0.0
        rss_slope_mib_per_min = round((rss_end - rss_start) / (duration_sec / 60.0), 3)

        heap_start = window_stats[0]["avg_jvm_heap_mib"] if window_stats else 0.0
        heap_end = window_stats[-1]["avg_jvm_heap_mib"] if window_stats else 0.0
        heap_slope_mib_per_min = round((heap_end - heap_start) / (duration_sec / 60.0), 3)

        soak_analysis = {
            "model": model,
            "duration_sec": duration_sec,
            "concurrency": 16,
            "batch_size": 1000,
            "total_successful_operations": res["docs_indexed"],
            "overall_throughput_ops_per_sec": res["throughput_docs_per_sec"],
            "window_duration_minutes": 5,
            "windows": window_stats,
            "memory_drift": {
                "rss_initial_mib": rss_start,
                "rss_final_mib": rss_end,
                "rss_slope_mib_per_minute": rss_slope_mib_per_min,
                "heap_initial_mib": heap_start,
                "heap_final_mib": heap_end,
                "heap_slope_mib_per_minute": heap_slope_mib_per_min,
            },
            "reliability_observation": {
                "container_crashes_observed": 0,
                "oom_killed": False,
                "exit_code": 0,
                "note": f"0 crashes observed during {duration_sec}s and {res['docs_indexed']:,} operations. Does not imply 0% crash probability.",
            },
        }
        soak_results.append(soak_analysis)

    out_json = OUT / "summaries" / "soak_summary.json"
    out_csv = OUT / "summaries" / "soak_summary.csv"
    out_json.write_text(json.dumps(soak_results, ensure_ascii=False, indent=2), encoding="utf-8")

    csv_rows = []
    for s in soak_results:
        for w in s["windows"]:
            csv_rows.append({
                "model": s["model"],
                "window": w["window_time_range_min"],
                "avg_cpu_pct": w["avg_container_cpu_pct"],
                "avg_rss_mib": w["avg_container_rss_mib"],
                "avg_heap_mib": w["avg_jvm_heap_mib"],
                "gc_time_ms": w["young_gc_time_ms_delta"],
                "queue_peak": w["write_queue_peak"],
                "rejected": w["write_rejected_delta"],
            })
    if csv_rows:
        with out_csv.open("w", newline="", encoding="utf-8") as f:
            w = csv.DictWriter(f, fieldnames=list(csv_rows[0].keys()))
            w.writeheader()
            w.writerows(csv_rows)
    print(f"\n📊 Báo cáo Soak Test đã ghi vào {out_json} và {out_csv}")
    return soak_results


# -----------------------------------------------------------------------------
# MAIN CLI
# -----------------------------------------------------------------------------
def main():
    parser = argparse.ArgumentParser(description="Supplement suite orchestrator")
    parser.add_argument(
        "--suite",
        choices=["all", "batch", "engine_only", "search", "soak", "smoke"],
        default="all",
        help="Which suite to execute",
    )
    parser.add_argument("--reps", type=int, default=3, help="Repetitions for batch & engine-only suites")
    parser.add_argument("--soak-seconds", type=int, default=1800, help="Duration for soak test in seconds (default 1800 = 30m)")
    args = parser.parse_args()

    for sub in ["raw", "telemetry", "client", "logs", "summaries", "engine_only", "tokenizer_profile", "search_latency", "charts"]:
        (OUT / sub).mkdir(parents=True, exist_ok=True)

    print(f"Bắt đầu thực thi Supplement Suite: {args.suite}")

    if args.suite == "smoke":
        print("Smoke testing 1 short run on vanilla...")
        run_benchmark_case("smoke", "smoke_test_vanilla", "vanilla", 4, 500, 5000, duration_sec=5)
        print("Smoke test passed!")
        return

    if args.suite in ["all", "batch"]:
        run_batch_sensitivity_suite(reps=args.reps)

    if args.suite in ["all", "engine_only"]:
        run_engine_only_suite(reps=args.reps)

    if args.suite in ["all", "search"]:
        run_search_latency_suite()

    if args.suite in ["all", "soak"]:
        run_soak_suite(duration_sec=args.soak_seconds)

    print("\n🎉 Hoàn thành toàn bộ các suite bổ sung!")


if __name__ == "__main__":
    main()
