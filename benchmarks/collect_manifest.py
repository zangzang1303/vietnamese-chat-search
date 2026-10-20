import datetime
import hashlib
import json
import os
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[1]
RUN_ID = "final_20261020_093100"
RUN_DIR = ROOT / "benchmarks" / "final_run" / RUN_ID

def run_cmd(cmd, cwd=ROOT, wsl=False):
    if wsl:
        full_cmd = ["wsl", "-u", "root", "-d", "Ubuntu"] + cmd
    else:
        full_cmd = cmd
    res = subprocess.run(
        full_cmd,
        cwd=cwd,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace"
    )
    return res.stdout.strip(), res.stderr.strip(), res.returncode

def sha256_file(path):
    p = pathlib.Path(path)
    if not p.is_file():
        return None
    h = hashlib.sha256()
    with open(p, "rb") as f:
        while chunk := f.read(1024 * 1024):
            h.update(chunk)
    return h.hexdigest()

def sha256_dir(dir_path):
    p = pathlib.Path(dir_path)
    if not p.is_dir():
        return {}
    res = {}
    for item in sorted(p.rglob("*")):
        if item.is_file():
            rel = str(item.relative_to(p)).replace("\\", "/")
            res[rel] = sha256_file(item)
    return res

def main():
    print(f"Creating directories for {RUN_ID}...")
    for sub in [
        "environment",
        "raw",
        "telemetry",
        "parity",
        "search",
        "summaries",
        "charts",
        "logs"
    ]:
        (RUN_DIR / sub).mkdir(parents=True, exist_ok=True)

    env_dir = RUN_DIR / "environment"

    # Git
    git_sha, _, _ = run_cmd(["git", "rev-parse", "HEAD"])
    git_status, _, _ = run_cmd(["git", "status", "--short"])
    git_diff, _, _ = run_cmd(["git", "diff"])
    (env_dir / "git_status.txt").write_text(git_status, encoding="utf-8")
    (env_dir / "git_diff.txt").write_text(git_diff, encoding="utf-8")

    # System & Versions
    uname, _, _ = run_cmd(["uname", "-a"], wsl=True)
    lscpu, _, _ = run_cmd(["lscpu"], wsl=True)
    free_m, _, _ = run_cmd(["free", "-m"], wsl=True)
    docker_ver, _, _ = run_cmd(["docker", "version"], wsl=True)
    wsl_go_ver, _, _ = run_cmd(["go", "version"], wsl=True)

    (env_dir / "uname.txt").write_text(uname, encoding="utf-8")
    (env_dir / "lscpu.txt").write_text(lscpu, encoding="utf-8")
    (env_dir / "free_m.txt").write_text(free_m, encoding="utf-8")
    (env_dir / "docker_version.txt").write_text(docker_ver, encoding="utf-8")
    (env_dir / "go_version.txt").write_text(wsl_go_ver, encoding="utf-8")

    # Docker Image IDs
    docker_images, _, _ = run_cmd(["docker", "images", "--no-trunc"], wsl=True)
    (env_dir / "docker_images.txt").write_text(docker_images, encoding="utf-8")

    # Docker inspect
    vanilla_inspect, _, _ = run_cmd(["docker", "inspect", "os_vanilla_bench"], wsl=True)
    plugin_inspect, _, _ = run_cmd(["docker", "inspect", "os_plugin_bench"], wsl=True)
    (env_dir / "docker_inspect_vanilla.json").write_text(vanilla_inspect, encoding="utf-8")
    (env_dir / "docker_inspect_plugin.json").write_text(plugin_inspect, encoding="utf-8")

    vanilla_cfg = json.loads(vanilla_inspect)[0] if vanilla_inspect else {}
    plugin_cfg = json.loads(plugin_inspect)[0] if plugin_inspect else {}

    # Extract native library from plugin container / inspect plugins
    run_cmd(["docker", "start", "os_plugin_bench"], wsl=True)
    # Check plugin list
    plugin_list, _, _ = run_cmd(["docker", "exec", "os_plugin_bench", "/usr/share/opensearch/bin/opensearch-plugin", "list"], wsl=True)
    # Find libcoccoc in plugin container
    find_so, _, _ = run_cmd(["docker", "exec", "os_plugin_bench", "find", "/usr/share/opensearch", "-name", "*coccoc*"], wsl=True)
    (env_dir / "plugin_files.txt").write_text(f"Plugins:\n{plugin_list}\n\nFiles:\n{find_so}", encoding="utf-8")

    # Check sha256 inside container for native libs
    container_so_sha, _, _ = run_cmd(["docker", "exec", "os_plugin_bench", "sh", "-c", "sha256sum /usr/share/opensearch/plugins/*/* 2>/dev/null || true"], wsl=True)
    (env_dir / "container_plugin_sha256.txt").write_text(container_so_sha, encoding="utf-8")

    # Java versions inside containers
    java_ver, _, _ = run_cmd(["docker", "exec", "os_plugin_bench", "/usr/share/opensearch/jdk/bin/java", "-version"], wsl=True)
    (env_dir / "java_version.txt").write_text(java_ver, encoding="utf-8")

    # SHA256 of artifacts
    sha_map = {
        "dataset_real_1m": {
            "path": "benchmarks/dataset_real_1m.jsonl",
            "sha256": sha256_file(ROOT / "benchmarks" / "dataset_real_1m.jsonl"),
            "size_bytes": (ROOT / "benchmarks" / "dataset_real_1m.jsonl").stat().st_size if (ROOT / "benchmarks" / "dataset_real_1m.jsonl").exists() else 0
        },
        "mapping_vanilla": {
            "path": "benchmarks/mapping_vanilla.json",
            "sha256": sha256_file(ROOT / "benchmarks" / "mapping_vanilla.json")
        },
        "mapping_plugin": {
            "path": "benchmarks/mapping_plugin.json",
            "sha256": sha256_file(ROOT / "benchmarks" / "mapping_plugin.json")
        },
        "bench_runner_final_linux": {
            "path": "benchmarks/bench_runner_final_linux",
            "sha256": sha256_file(ROOT / "benchmarks" / "bench_runner_final_linux")
        },
        "tokenize_corpus_final_linux": {
            "path": "benchmarks/tokenize_corpus_final_linux",
            "sha256": sha256_file(ROOT / "benchmarks" / "tokenize_corpus_final_linux")
        },
        "dictionary_coccoc_plugin_final": sha256_dir(ROOT / "data" / "dicts" / "coccoc_plugin_final"),
        "dictionary_coccoc": sha256_dir(ROOT / "data" / "dicts" / "coccoc")
    }

    manifest = {
        "run_id": RUN_ID,
        "created_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "git": {
            "sha": git_sha,
            "status_summary": git_status.splitlines()[:20]
        },
        "system": {
            "uname": uname,
            "go_version": wsl_go_ver,
            "docker_version": docker_ver.splitlines()[0] if docker_ver else "",
            "java_version": java_ver.splitlines()[:3] if java_ver else []
        },
        "resource_limits": {
            "vanilla": {
                "NanoCPUs": vanilla_cfg.get("HostConfig", {}).get("NanoCPUs"),
                "Memory": vanilla_cfg.get("HostConfig", {}).get("Memory"),
                "CpuQuota": vanilla_cfg.get("HostConfig", {}).get("CpuQuota")
            },
            "plugin": {
                "NanoCPUs": plugin_cfg.get("HostConfig", {}).get("NanoCPUs"),
                "Memory": plugin_cfg.get("HostConfig", {}).get("Memory"),
                "CpuQuota": plugin_cfg.get("HostConfig", {}).get("CpuQuota")
            }
        },
        "provenance": {
            "opensearch_plugin_vietnamese": "unknown (pre-built from Dockerfile without git commit provenance)",
            "coccoc_tokenizer_library": "unknown (pre-compiled binary in Docker image)",
            "dictionary_source": "extracted from os-plugin:2.19.0 to data/dicts/coccoc_plugin_final for byte-for-byte parity"
        },
        "checksums": sha_map
    }

    manifest_file = RUN_DIR / "manifest.json"
    manifest_file.write_text(json.dumps(manifest, indent=2, ensure_ascii=False), encoding="utf-8")
    print(f"Manifest written successfully to {manifest_file}")

if __name__ == "__main__":
    main()
