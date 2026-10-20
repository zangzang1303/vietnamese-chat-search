import argparse
import json
import math
import os
import pathlib
import re
import subprocess
import sys
import time
import urllib.request
import urllib.error

sys.stdout.reconfigure(encoding="utf-8")

ROOT = pathlib.Path(__file__).resolve().parents[1]
DATASET = ROOT / "benchmarks" / "dataset_real_1m.jsonl"
OUTPUT_DIR = ROOT / "benchmarks" / "final_run" / "final_20261020_093100" / "search"

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

def request_json(url, method="GET", payload=None, timeout=30):
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

def build_queries(target_count=1000):
    """Trích xuất 1000 câu query thực tế đa dạng từ dataset."""
    queries_file = OUTPUT_DIR / "search_queries_raw.json"
    tokenized_queries_file = OUTPUT_DIR / "search_queries_tokenized.json"
    
    if tokenized_queries_file.exists():
        with open(tokenized_queries_file, "r", encoding="utf-8") as f:
            return json.load(f)

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    
    # 1. Standard curated seed queries
    curated = [
        # Compound accented
        ("học sinh", "compound_accented"),
        ("cà phê", "compound_accented"),
        ("việt nam", "compound_accented"),
        ("văn phòng", "compound_accented"),
        ("thời khóa biểu", "compound_accented"),
        ("sinh viên", "compound_accented"),
        ("máy lạnh", "compound_accented"),
        ("công nghệ thông tin", "compound_accented"),
        ("sản phẩm", "compound_accented"),
        ("điện thoại", "compound_accented"),
        ("tản nhiệt", "compound_accented"),
        ("bộ xử lý", "compound_accented"),
        ("thị trường", "compound_accented"),
        ("chất kết dính", "compound_accented"),
        ("vật liệu", "compound_accented"),
        ("chính sách", "compound_accented"),
        ("phát triển", "compound_accented"),
        ("quốc gia", "compound_accented"),
        ("kinh tế", "compound_accented"),
        ("thủ đô", "compound_accented"),
        # Unaccented
        ("hoc sinh", "unaccented"),
        ("ca phe", "unaccented"),
        ("viet nam", "unaccented"),
        ("van phong", "unaccented"),
        ("thoi khoa bieu", "unaccented"),
        ("sinh vien", "unaccented"),
        ("may lanh", "unaccented"),
        ("cong nghe thong tin", "unaccented"),
        ("san pham", "unaccented"),
        ("dien thoai", "unaccented"),
        ("tan nhiet", "unaccented"),
        ("bo xu ly", "unaccented"),
        ("thi truong", "unaccented"),
        ("vat lieu", "unaccented"),
        ("chinh sach", "unaccented"),
        ("phat trien", "unaccented"),
        ("quoc gia", "unaccented"),
        ("kinh te", "unaccented"),
        ("thu do", "unaccented"),
        # Edge prefixes
        ("vie", "edge_prefix"),
        ("viet_n", "edge_prefix"),
        ("hoc_s", "edge_prefix"),
        ("van_p", "edge_prefix"),
        ("ca_p", "edge_prefix"),
        ("sin_v", "edge_prefix"),
        ("may_l", "edge_prefix"),
        ("con_n", "edge_prefix"),
        ("san_p", "edge_prefix"),
        ("die_t", "edge_prefix"),
        ("tan_n", "edge_prefix"),
        ("bo_x", "edge_prefix"),
        ("thi_t", "edge_prefix"),
        ("vat_l", "edge_prefix"),
        # Single words
        ("chào", "single_word"),
        ("người", "single_word"),
        ("ngày", "single_word"),
        ("năm", "single_word"),
        ("tháng", "single_word"),
        ("làm", "single_word"),
        ("được", "single_word"),
        ("nhiều", "single_word"),
        ("rất", "single_word"),
        ("khi", "single_word"),
        ("chao", "single_word"),
        ("nguoi", "single_word"),
        ("ngay", "single_word"),
        ("nam", "single_word"),
        ("thang", "single_word"),
        ("lam", "single_word"),
        ("duoc", "single_word"),
        ("nhieu", "single_word"),
        ("rat", "single_word"),
        ("khi", "single_word"),
    ]

    seen = set()
    queries = []
    for q, cat in curated:
        if q not in seen:
            seen.add(q)
            queries.append({"raw": q, "category": cat})

    # Sample phrases from dataset
    with open(DATASET, "r", encoding="utf-8") as f:
        for line in f:
            if len(queries) >= target_count:
                break
            try:
                doc = json.loads(line)
                text = doc.get("text", "")
                words = [w for w in re.split(r'\s+', text) if len(w) > 1 and re.match(r'^[\w]+$', w)]
                if len(words) >= 4:
                    # 2-word phrase
                    p2 = f"{words[0]} {words[1]}".lower()
                    if p2 not in seen and len(p2) >= 4:
                        seen.add(p2)
                        queries.append({"raw": p2, "category": "dataset_phrase_2w"})
                    # 3-word phrase
                    p3 = f"{words[1]} {words[2]} {words[3]}".lower()
                    if len(queries) < target_count and p3 not in seen and len(p3) >= 6:
                        seen.add(p3)
                        queries.append({"raw": p3, "category": "dataset_phrase_3w"})
                    # Single word
                    w1 = words[2].lower()
                    if len(queries) < target_count and w1 not in seen and len(w1) >= 3:
                        seen.add(w1)
                        queries.append({"raw": w1, "category": "dataset_word"})
            except Exception:
                continue

    queries = queries[:target_count]
    print(f"Generated {len(queries)} unique search queries.")

    # Tokenize with CGO for Model A
    # Write temp JSONL for tokenize_corpus_final_linux
    tmp_in = OUTPUT_DIR / "tmp_queries_in.jsonl"
    tmp_out = OUTPUT_DIR / "tmp_queries_out.jsonl"
    with open(tmp_in, "w", encoding="utf-8") as f:
        for idx, item in enumerate(queries):
            f.write(json.dumps({"id": idx, "text": item["raw"]}, ensure_ascii=False) + "\n")

    dict_path = ROOT / "data" / "dicts" / "coccoc_plugin_final"
    tok_bin = ROOT / "benchmarks" / "tokenize_corpus_final_linux"

    cmd = [
        "wsl", "-u", "root", "-d", "Ubuntu",
        wsl_path(tok_bin),
        f"-input={wsl_path(tmp_in)}",
        f"-output={wsl_path(tmp_out)}",
        f"-dict={wsl_path(dict_path)}"
    ]
    subprocess.run(cmd, check=True)

    tokenized_map = {}
    with open(tmp_out, "r", encoding="utf-8") as f:
        for line in f:
            obj = json.loads(line)
            tokenized_map[obj["id"]] = obj["segmented"]

    final_queries = []
    for idx, item in enumerate(queries):
        seg = tokenized_map.get(idx, item["raw"])
        final_queries.append({
            "id": idx,
            "raw": item["raw"],
            "cgo": seg,
            "category": item["category"]
        })

    with open(tokenized_queries_file, "w", encoding="utf-8") as f:
        json.dump(final_queries, f, indent=2, ensure_ascii=False)

    return final_queries

def run_query(url, q_obj, field_name, is_vanilla):
    """Thực thi một query tương ứng theo mô hình và trường dữ liệu."""
    raw = q_obj["raw"]
    cgo = q_obj["cgo"]
    
    if field_name == "text":
        q_text = raw
        query_clause = {"match": {"text": q_text}}
    elif field_name == "text_vi":
        q_text = cgo if is_vanilla else raw
        query_clause = {"match": {"text_vi": q_text}}
    elif field_name == "text_raw":
        q_text = cgo if is_vanilla else raw
        query_clause = {"match": {"text_raw": q_text}}
    elif field_name == "text_edge":
        q_text = cgo if is_vanilla else raw
        query_clause = {"match": {"text_edge": q_text}}
    elif field_name == "multi_field":
        q_vi = cgo if is_vanilla else raw
        query_clause = {
            "bool": {
                "should": [
                    {"match": {"text": {"query": raw, "boost": 1.0}}},
                    {"match": {"text_vi": {"query": q_vi, "boost": 4.0}}},
                    {"match": {"text_raw": {"query": q_vi, "boost": 3.0}}},
                    {"match": {"text_edge": {"query": q_vi, "boost": 2.0}}}
                ]
            }
        }
    else:
        raise ValueError(f"Unknown field {field_name}")

    body = {
        "size": 100,
        "track_total_hits": True,
        "sort": [{"_score": "desc"}, {"id": "desc"}],
        "query": query_clause
    }

    t0 = time.perf_counter()
    res = request_json(url + "/messages/_search", method="POST", payload=body, timeout=15)
    lat_ms = (time.perf_counter() - t0) * 1000.0

    total_hits = res["hits"]["total"]["value"]
    hits = res["hits"]["hits"]
    top10_ids = [str(h["_id"]) for h in hits[:10]]
    top100_ids = [str(h["_id"]) for h in hits[:100]]
    top10_scores = [round(h["_score"], 4) if h["_score"] is not None else 0.0 for h in hits[:10]]

    return {
        "latency_ms": lat_ms,
        "total_hits": total_hits,
        "top10_ids": top10_ids,
        "top100_ids": top100_ids,
        "top10_scores": top10_scores
    }

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--docs", type=int, default=50000, help="Số docs index để test search")
    parser.add_argument("--queries", type=int, default=1000, help="Số queries kiểm thử")
    args = parser.parse_args()

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    queries = build_queries(args.queries)
    print(f"Loaded {len(queries)} search queries.")

    runner_bin = wsl_path(ROOT / "benchmarks" / "bench_runner_final_linux")
    dataset_wsl = wsl_path(DATASET)
    dict_wsl = wsl_path(ROOT / "data" / "dicts" / "coccoc_plugin_final")

    # =========================================================================
    # STEP 1: VANILLA (MODEL A)
    # =========================================================================
    print("\n" + "=" * 80)
    print("🔵 [SEARCH PARITY - STEP 1] MÔ HÌNH A: OPENSEARCH VANILLA (CGO)")
    print("=" * 80)
    docker("stop", "os_plugin_bench")
    docker("start", "os_vanilla_bench")
    wait_ready("http://localhost:9201")
    recreate_index("http://localhost:9201", ROOT / "benchmarks" / "mapping_vanilla.json")

    print(f"Indexing {args.docs} docs into Vanilla...")
    v_out = OUTPUT_DIR / "vanilla_index_stats.json"
    cmd_v = [
        "wsl", "-u", "root", "-d", "Ubuntu", "bash", "-c",
        f"{runner_bin} -mode=vanilla -concurrency=8 -batch=1000 -max_docs={args.docs} "
        f"-dataset={dataset_wsl} -dict={dict_wsl} -vanilla_url=http://localhost:9201 "
        f"-output={wsl_path(v_out)}"
    ]
    subprocess.run(cmd_v, check=True)

    # Bulk Correctness Gate for Vanilla
    v_stats = json.loads(v_out.read_text(encoding="utf-8"))[0]
    count_resp = request_json("http://localhost:9201/messages/_count")
    v_indexed_count = count_resp.get("count", 0)
    print(f"Vanilla indexed: {v_indexed_count} docs. Errors: bulk={v_stats.get('bulk_item_errors')}, http={v_stats.get('http_error_batches')}, tok={v_stats.get('tokenizer_errors')}")
    if v_stats.get("bulk_item_errors", 0) > 0 or v_stats.get("http_error_batches", 0) > 0 or v_indexed_count != args.docs:
        raise RuntimeError(f"Vanilla Bulk Correctness Gate FAILED! count={v_indexed_count}, expected={args.docs}")
    print("✅ Vanilla Bulk Correctness Gate PASSED (100% clean).")

    # Run queries on Vanilla
    print("Executing queries on Vanilla...")
    fields = ["text", "text_vi", "text_raw", "text_edge", "multi_field"]
    vanilla_search_results = {}
    for f in fields:
        vanilla_search_results[f] = {}
        for q in queries:
            qid = str(q["id"])
            res = run_query("http://localhost:9201", q, f, is_vanilla=True)
            vanilla_search_results[f][qid] = res

    (OUTPUT_DIR / "vanilla_search_results.json").write_text(
        json.dumps(vanilla_search_results, indent=2, ensure_ascii=False), encoding="utf-8"
    )

    # =========================================================================
    # STEP 2: PLUGIN (MODEL B)
    # =========================================================================
    print("\n" + "=" * 80)
    print("🟢 [SEARCH PARITY - STEP 2] MÔ HÌNH B: OPENSEARCH VIETNAMESE PLUGIN (JNI)")
    print("=" * 80)
    docker("stop", "os_vanilla_bench")
    docker("start", "os_plugin_bench")
    wait_ready("http://localhost:9202")
    recreate_index("http://localhost:9202", ROOT / "benchmarks" / "mapping_plugin.json")

    print(f"Indexing {args.docs} docs into Plugin...")
    p_out = OUTPUT_DIR / "plugin_index_stats.json"
    cmd_p = [
        "wsl", "-u", "root", "-d", "Ubuntu", "bash", "-c",
        f"{runner_bin} -mode=plugin -concurrency=8 -batch=1000 -max_docs={args.docs} "
        f"-dataset={dataset_wsl} -dict={dict_wsl} -plugin_url=http://localhost:9202 "
        f"-output={wsl_path(p_out)}"
    ]
    subprocess.run(cmd_p, check=True)

    # Bulk Correctness Gate for Plugin
    p_stats = json.loads(p_out.read_text(encoding="utf-8"))[0]
    count_resp_p = request_json("http://localhost:9202/messages/_count")
    p_indexed_count = count_resp_p.get("count", 0)
    print(f"Plugin indexed: {p_indexed_count} docs. Errors: bulk={p_stats.get('bulk_item_errors')}, http={p_stats.get('http_error_batches')}, tok={p_stats.get('tokenizer_errors')}")
    if p_stats.get("bulk_item_errors", 0) > 0 or p_stats.get("http_error_batches", 0) > 0 or p_indexed_count != args.docs:
        raise RuntimeError(f"Plugin Bulk Correctness Gate FAILED! count={p_indexed_count}, expected={args.docs}")
    print("✅ Plugin Bulk Correctness Gate PASSED (100% clean).")

    # Run queries on Plugin
    print("Executing queries on Plugin...")
    plugin_search_results = {}
    for f in fields:
        plugin_search_results[f] = {}
        for q in queries:
            qid = str(q["id"])
            res = run_query("http://localhost:9202", q, f, is_vanilla=False)
            plugin_search_results[f][qid] = res

    (OUTPUT_DIR / "plugin_search_results.json").write_text(
        json.dumps(plugin_search_results, indent=2, ensure_ascii=False), encoding="utf-8"
    )

    # =========================================================================
    # STEP 3: SEARCH PARITY COMPARISON & METRICS
    # =========================================================================
    print("\n" + "=" * 80)
    print("⚖️ [SEARCH PARITY - STEP 3] SO SÁNH KẾT QUẢ TÌM KIẾM CHI TIẾT")
    print("=" * 80)

    comparison_summary = {}
    detailed_mismatches = []

    for f in fields:
        total_q = len(queries)
        hits_matched = 0
        top10_matched = 0
        top100_matched = 0
        ranking_matched = 0
        score_close_count = 0
        vanilla_lats = []
        plugin_lats = []

        for q in queries:
            qid = str(q["id"])
            v_res = vanilla_search_results[f][qid]
            p_res = plugin_search_results[f][qid]

            vanilla_lats.append(v_res["latency_ms"])
            plugin_lats.append(p_res["latency_ms"])

            # 1. Total hits match
            if v_res["total_hits"] == p_res["total_hits"]:
                hits_matched += 1

            # 2. Top-10 IDs match (set equality)
            if set(v_res["top10_ids"]) == set(p_res["top10_ids"]):
                top10_matched += 1

            # 3. Top-100 IDs match (set equality)
            if set(v_res["top100_ids"]) == set(p_res["top100_ids"]):
                top100_matched += 1

            # 4. Exact ranking match (order equality)
            if v_res["top10_ids"] == p_res["top10_ids"]:
                ranking_matched += 1

            # 5. Score tolerance (within 0.001)
            score_diffs = []
            if len(v_res["top10_scores"]) == len(p_res["top10_scores"]):
                score_diffs = [abs(vs - ps) for vs, ps in zip(v_res["top10_scores"], p_res["top10_scores"])]
                if all(d <= 0.001 for d in score_diffs):
                    score_close_count += 1

            # Record mismatch if not 100% matched
            if v_res["top10_ids"] != p_res["top10_ids"] or v_res["total_hits"] != p_res["total_hits"]:
                detailed_mismatches.append({
                    "id": qid,
                    "field": f,
                    "query_raw": q["raw"],
                    "query_cgo": q["cgo"],
                    "category": q["category"],
                    "vanilla_hits": v_res["total_hits"],
                    "plugin_hits": p_res["total_hits"],
                    "vanilla_top10": v_res["top10_ids"],
                    "plugin_top10": p_res["top10_ids"]
                })

        vanilla_lats.sort()
        plugin_lats.sort()

        comparison_summary[f] = {
            "total_queries": total_q,
            "total_hits_match_rate": hits_matched / total_q,
            "top10_ids_match_rate": top10_matched / total_q,
            "top100_ids_match_rate": top100_matched / total_q,
            "top10_exact_rank_match_rate": ranking_matched / total_q,
            "top10_score_tolerance_match_rate": score_close_count / total_q,
            "vanilla_latency_p50_ms": vanilla_lats[int(total_q * 0.50)],
            "vanilla_latency_p95_ms": vanilla_lats[int(total_q * 0.95)],
            "plugin_latency_p50_ms": plugin_lats[int(total_q * 0.50)],
            "plugin_latency_p95_ms": plugin_lats[int(total_q * 0.95)]
        }

        print(f"Trường [{f:11s}]: Total Hits Match={comparison_summary[f]['total_hits_match_rate']*100:5.1f}% | "
              f"Top-10 IDs Match={comparison_summary[f]['top10_ids_match_rate']*100:5.1f}% | "
              f"Top-10 Exact Rank={comparison_summary[f]['top10_exact_rank_match_rate']*100:5.1f}% | "
              f"Latency p50: V={comparison_summary[f]['vanilla_latency_p50_ms']:.2f}ms vs P={comparison_summary[f]['plugin_latency_p50_ms']:.2f}ms")

    # By category breakdown for multi_field
    cat_summary = {}
    for q in queries:
        cat = q["category"]
        if cat not in cat_summary:
            cat_summary[cat] = {"total": 0, "hits_match": 0, "top10_match": 0, "v_lats": [], "p_lats": []}
        qid = str(q["id"])
        v = vanilla_search_results["multi_field"][qid]
        p = plugin_search_results["multi_field"][qid]
        cat_summary[cat]["total"] += 1
        if v["total_hits"] == p["total_hits"]:
            cat_summary[cat]["hits_match"] += 1
        if v["top10_ids"] == p["top10_ids"]:
            cat_summary[cat]["top10_match"] += 1
        cat_summary[cat]["v_lats"].append(v["latency_ms"])
        cat_summary[cat]["p_lats"].append(p["latency_ms"])

    for cat, d in cat_summary.items():
        v_lats = sorted(d["v_lats"])
        p_lats = sorted(d["p_lats"])
        n = d["total"]
        d["hits_match_rate"] = d["hits_match"] / n
        d["top10_match_rate"] = d["top10_match"] / n
        d["v_p50"] = v_lats[int(n * 0.5)]
        d["p_p50"] = p_lats[int(n * 0.5)]
        del d["v_lats"]
        del d["p_lats"]

    final_report = {
        "dataset_docs": args.docs,
        "queries_count": len(queries),
        "fields_summary": comparison_summary,
        "multi_field_category_summary": cat_summary,
        "mismatches_sample": detailed_mismatches[:50],
        "total_mismatches_count": len(detailed_mismatches)
    }

    summary_file = OUTPUT_DIR / "search_parity_summary.json"
    summary_file.write_text(json.dumps(final_report, indent=2, ensure_ascii=False), encoding="utf-8")
    (OUTPUT_DIR / "search_parity_mismatches.json").write_text(
        json.dumps(detailed_mismatches, indent=2, ensure_ascii=False), encoding="utf-8"
    )
    print(f"\n💾 Search Parity report written to {summary_file}")

if __name__ == "__main__":
    main()
