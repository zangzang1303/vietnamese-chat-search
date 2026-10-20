"""Recompute statistically correct summaries for the final paired experiment."""

import csv
import json
import math
import pathlib
import random
import statistics


ROOT = pathlib.Path(__file__).resolve().parents[1]
RUN_DIR = ROOT / "benchmarks" / "final_run" / "final_20261020_093100"
RAW_DIR = RUN_DIR / "raw"
OUT_JSON = RUN_DIR / "summaries" / "statistical_analysis.json"
OUT_CSV = RUN_DIR / "summaries" / "paired_throughput.csv"
T95_DF4 = 2.7764451051977987


def percentile(values, p):
    values = sorted(values)
    if not values:
        return 0.0
    pos = (len(values) - 1) * p
    lo = int(math.floor(pos))
    hi = int(math.ceil(pos))
    if lo == hi:
        return values[lo]
    return values[lo] + (values[hi] - values[lo]) * (pos - lo)


def bootstrap_median_ci(values, seed, samples=50000):
    rng = random.Random(seed)
    n = len(values)
    medians = [statistics.median(rng.choices(values, k=n)) for _ in range(samples)]
    return [percentile(medians, 0.025), percentile(medians, 0.975)]


def mean_ci(values):
    mean = statistics.mean(values)
    if len(values) < 2:
        return [mean, mean]
    margin = T95_DF4 * statistics.stdev(values) / math.sqrt(len(values))
    return [mean - margin, mean + margin]


def describe(values, seed):
    mean = statistics.mean(values)
    std = statistics.stdev(values) if len(values) > 1 else 0.0
    return {
        "n": len(values),
        "median": round(statistics.median(values), 3),
        "median_bootstrap_ci95": [round(x, 3) for x in bootstrap_median_ci(values, seed)],
        "mean": round(mean, 3),
        "mean_t_ci95": [round(x, 3) for x in mean_ci(values)],
        "std": round(std, 3),
        "cv_pct": round(std / mean * 100, 3) if mean else 0.0,
        "min": round(min(values), 3),
        "max": round(max(values), 3),
    }


def main():
    runs = []
    for path in sorted(RAW_DIR.glob("rep*_*.json")):
        item = json.loads(path.read_text(encoding="utf-8"))
        item["source_file"] = path.name
        item["model"] = "vanilla" if "vanilla" in path.name else "plugin"
        runs.append(item)

    by_key = {(r["model"], int(r["concurrency"]), int(r["repetition"])): r for r in runs}
    results = []
    csv_rows = []
    for concurrency in [1, 4, 8, 16, 32]:
        vanilla = [by_key[("vanilla", concurrency, rep)]["docs_per_sec"] for rep in range(1, 6)]
        plugin = [by_key[("plugin", concurrency, rep)]["docs_per_sec"] for rep in range(1, 6)]
        differences = [a - b for a, b in zip(vanilla, plugin)]
        diff_mean = statistics.mean(differences)
        diff_std = statistics.stdev(differences)
        diff_ci = mean_ci(differences)
        dz = diff_mean / diff_std if diff_std else 0.0
        conclusive = diff_ci[0] > 0 or diff_ci[1] < 0
        row = {
            "concurrency": concurrency,
            "vanilla_throughput": describe(vanilla, 1000 + concurrency),
            "plugin_throughput": describe(plugin, 2000 + concurrency),
            "paired_difference_a_minus_b": {
                "values": [round(x, 3) for x in differences],
                "mean": round(diff_mean, 3),
                "mean_t_ci95": [round(x, 3) for x in diff_ci],
                "cohen_dz": round(dz, 3),
                "direction_consistency": f"{sum(x > 0 for x in differences)}/5 A>B",
                "statistically_conclusive_at_0_05": conclusive,
            },
        }
        results.append(row)
        csv_rows.append({
            "concurrency": concurrency,
            "a_median": row["vanilla_throughput"]["median"],
            "a_median_ci_low": row["vanilla_throughput"]["median_bootstrap_ci95"][0],
            "a_median_ci_high": row["vanilla_throughput"]["median_bootstrap_ci95"][1],
            "b_median": row["plugin_throughput"]["median"],
            "b_median_ci_low": row["plugin_throughput"]["median_bootstrap_ci95"][0],
            "b_median_ci_high": row["plugin_throughput"]["median_bootstrap_ci95"][1],
            "paired_mean_diff": round(diff_mean, 3),
            "paired_ci_low": round(diff_ci[0], 3),
            "paired_ci_high": round(diff_ci[1], 3),
            "cohen_dz": round(dz, 3),
            "conclusive": conclusive,
        })

    payload = {
        "design": "paired randomized AB/BA, five repetitions per model and concurrency",
        "primary_metric": "docs_per_sec",
        "median_interval_method": "percentile bootstrap, 50,000 resamples, deterministic seed",
        "paired_interval_method": "two-sided 95% Student-t interval on within-repetition A-B differences, df=4",
        "multiple_comparison_note": "No multiplicity-adjusted significance claim is made.",
        "interpretation": "A result is conclusive only when the paired 95% CI excludes zero.",
        "concurrency_results": results,
        "overall_conclusion": "No concurrency level has a paired 95% CI excluding zero; observed median differences are descriptive, not confirmed wins.",
    }
    OUT_JSON.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    with OUT_CSV.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=list(csv_rows[0]))
        writer.writeheader()
        writer.writerows(csv_rows)
    print(OUT_JSON)
    print(OUT_CSV)


if __name__ == "__main__":
    main()
