#!/usr/bin/env python3
"""Deterministic token parity gate using the real CGO output and index field analyzers."""

import argparse
import concurrent.futures
import json
import pathlib
import subprocess
import time
import urllib.error
import urllib.request

ROOT = pathlib.Path(__file__).resolve().parents[1]
DATASET = ROOT / "benchmarks" / "dataset_real_1m.jsonl"
CORPUS = ROOT / "benchmarks" / "parity_corpus_final.jsonl"
TOKENIZED = ROOT / "benchmarks" / "parity_corpus_final_tokenized.jsonl"

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
FIELDS = ("text", "text_vi", "text_raw", "text_edge")


def docker(*args):
    proc = subprocess.run(
        ["wsl", "-u", "root", "-d", "Ubuntu", "docker", *args],
        check=False,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    if proc.returncode:
        raise RuntimeError(f"docker {' '.join(args)} failed: {proc.stderr}")
    return proc.stdout.strip()


def wsl_path(path):
    path = pathlib.Path(path).resolve()
    drive = path.drive.rstrip(":").lower()
    tail = path.as_posix().split(":", 1)[1]
    return f"/mnt/{drive}{tail}"


def request_json(url, method="GET", payload=None, timeout=30):
    data = None if payload is None else json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=data,
        method=method,
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8"))


def wait_ready(url, seconds=180):
    deadline = time.time() + seconds
    while time.time() < deadline:
        try:
            if request_json(url + "/_cluster/health", timeout=3).get("status") in {"green", "yellow"}:
                time.sleep(5)
                return
        except Exception:
            time.sleep(2)
    raise TimeoutError(f"OpenSearch not ready: {url}")


def init_model(name):
    cfg = MODELS[name]
    docker("stop", cfg["other"])
    docker("start", cfg["container"])
    wait_ready(cfg["url"])
    try:
        request_json(cfg["url"] + "/messages", method="DELETE")
    except urllib.error.HTTPError as exc:
        if exc.code != 404:
            raise
    mapping = json.loads(cfg["mapping"].read_text(encoding="utf-8"))
    request_json(cfg["url"] + "/messages", method="PUT", payload=mapping)
    wait_ready(cfg["url"])


def build_corpus(count, dict_path):
    total = 1_000_000
    targets = {round(i * (total - 1) / (count - 1)) for i in range(count)}
    with DATASET.open(encoding="utf-8") as source, CORPUS.open("w", encoding="utf-8") as out:
        for idx, line in enumerate(source):
            if idx in targets:
                doc = json.loads(line)
                out.write(json.dumps({"id": doc["id"], "text": doc["text"]}, ensure_ascii=False) + "\n")
    subprocess.run(
        [
            "wsl",
            "-u",
            "root",
            "-d",
            "Ubuntu",
            wsl_path(ROOT / "benchmarks" / "tokenize_corpus_final_linux"),
            f"-input={wsl_path(CORPUS)}",
            f"-output={wsl_path(TOKENIZED)}",
            f"-dict={wsl_path(dict_path)}",
        ],
        cwd=ROOT,
        check=True,
    )


def analyze_one(url, field, text):
    last_error = None
    for attempt in range(3):
        try:
            response = request_json(
                url + "/messages/_analyze",
                method="POST",
                payload={"field": field, "text": text},
                timeout=20,
            )
            break
        except Exception as exc:
            last_error = exc
            time.sleep(0.25 * (attempt + 1))
    else:
        raise last_error
    return [
        {
            "token": token.get("token"),
            "position": token.get("position"),
            "positionLength": token.get("positionLength", 1),
            "start_offset": token.get("start_offset"),
            "end_offset": token.get("end_offset"),
            "type": token.get("type"),
        }
        for token in response.get("tokens", [])
    ]


def collect(name, docs, workers):
    init_model(name)
    cfg = MODELS[name]
    jobs = []
    for doc in docs:
        for field in FIELDS:
            if name == "vanilla" and field != "text":
                text = doc["segmented"]
            else:
                text = doc["text"]
            jobs.append((str(doc["id"]), field, text))

    results = {}
    errors = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
        future_map = {
            pool.submit(analyze_one, cfg["url"], field, text): (doc_id, field)
            for doc_id, field, text in jobs
        }
        for future in concurrent.futures.as_completed(future_map):
            key = future_map[future]
            try:
                results[key] = future.result()
            except Exception as exc:
                errors.append({"id": key[0], "field": key[1], "error": repr(exc)})
    return results, errors


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--count", type=int, default=2000)
    parser.add_argument("--workers", type=int, default=16)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument(
        "--dict",
        default=str(ROOT / "data" / "dicts" / "coccoc_plugin_final"),
        help="dictionary used by the CGO side",
    )
    args = parser.parse_args()

    output_dir = pathlib.Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    build_corpus(args.count, pathlib.Path(args.dict))
    docs = [json.loads(line) for line in TOKENIZED.read_text(encoding="utf-8").splitlines()]

    vanilla, vanilla_errors = collect("vanilla", docs, args.workers)
    plugin, plugin_errors = collect("plugin", docs, args.workers)

    mismatches = []
    token_type_divergences = 0
    matched = 0
    for key, vanilla_tokens in vanilla.items():
        plugin_tokens = plugin.get(key)
        def semantic(tokens):
            if tokens is None:
                return None
            return [{k: v for k, v in token.items() if k != "type"} for token in tokens]

        if semantic(vanilla_tokens) == semantic(plugin_tokens):
            matched += 1
            if vanilla_tokens != plugin_tokens:
                token_type_divergences += 1
        else:
            mismatches.append(
                {
                    "id": key[0],
                    "field": key[1],
                    "vanilla": vanilla_tokens,
                    "plugin": plugin_tokens,
                }
            )

    expected = len(docs) * len(FIELDS)
    summary = {
        "documents": len(docs),
        "fields": list(FIELDS),
        "comparisons_expected": expected,
        "comparisons_matched": matched,
        "parity_percent": matched / expected * 100 if expected else 0,
        "vanilla_request_errors": len(vanilla_errors),
        "plugin_request_errors": len(plugin_errors),
        "mismatch_count": len(mismatches),
        "token_type_metadata_divergences": token_type_divergences,
        "passed": matched == expected and not vanilla_errors and not plugin_errors,
    }
    (output_dir / "token_parity_summary.json").write_text(
        json.dumps(summary, indent=2, ensure_ascii=False), encoding="utf-8"
    )
    (output_dir / "token_parity_mismatches.json").write_text(
        json.dumps(mismatches, indent=2, ensure_ascii=False), encoding="utf-8"
    )
    (output_dir / "token_parity_request_errors.json").write_text(
        json.dumps(vanilla_errors + plugin_errors, indent=2, ensure_ascii=False), encoding="utf-8"
    )
    print(json.dumps(summary, indent=2, ensure_ascii=False))
    raise SystemExit(0 if summary["passed"] else 2)


if __name__ == "__main__":
    main()
