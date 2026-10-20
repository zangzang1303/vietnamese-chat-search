import json
import pathlib
import sys
from collections import Counter

sys.stdout.reconfigure(encoding="utf-8")

ROOT = pathlib.Path(__file__).resolve().parents[1]
MISMATCH_FILE = ROOT / "benchmarks" / "final_run" / "final_20261020_093100" / "parity" / "full_2000" / "token_parity_mismatches.json"

with open(MISMATCH_FILE, encoding="utf-8") as f:
    mismatches = json.load(f)

print(f"Total mismatches: {len(mismatches)} / 8000 (95.525% parity)")

field_counter = Counter(m["field"] for m in mismatches)
print("\nMismatches by field:")
for field, count in field_counter.items():
    print(f"  - {field}: {count} / 2000 ({count / 2000 * 100:.2f}%)")

# Categorize causes
causes = Counter()
examples = {}

for m in mismatches:
    vt = [t["token"] for t in m["vanilla"]]
    pt = [t["token"] for t in m["plugin"]]
    
    # Identify diffs
    diff_v = set(vt) - set(pt)
    diff_p = set(pt) - set(vt)
    
    cat = "other"
    if any("%" in p for p in diff_p):
        cat = "percent_sign_dropped_by_standard_tokenizer"
    elif any("'" in v or "'" in p for v in diff_v for p in diff_p):
        cat = "apostrophe_split_difference"
    elif any("," in v or "." in p or any(c.isdigit() for c in v) for v in diff_v for p in diff_p):
        cat = "numeric_comma_dot_formatting"
    elif any(":" in v or "/" in v or "@" in v for v in diff_v):
        cat = "punctuation_url_colon_hyphen"
    elif len(vt) != len(pt):
        cat = "compound_boundary_token_count_diff"
    else:
        cat = "word_token_diff"
        
    causes[cat] += 1
    if cat not in examples:
        examples[cat] = {
            "id": m["id"],
            "field": m["field"],
            "diff_vanilla": list(diff_v)[:5],
            "diff_plugin": list(diff_p)[:5]
        }

print("\nRoot cause breakdown:")
for cat, count in causes.most_common():
    print(f"  - {cat}: {count} ({count/len(mismatches)*100:.1f}%)")
    ex = examples[cat]
    print(f"    Example (doc {ex['id']}, {ex['field']}):")
    print(f"      Vanilla diff: {ex['diff_vanilla']}")
    print(f"      Plugin diff:  {ex['diff_plugin']}")

summary_obj = {
    "total_mismatches": len(mismatches),
    "comparisons_expected": 8000,
    "parity_rate": (8000 - len(mismatches)) / 8000,
    "by_field": dict(field_counter),
    "by_cause": dict(causes),
    "examples": examples
}

out_path = ROOT / "benchmarks" / "final_run" / "final_20261020_093100" / "parity" / "mismatch_cause_analysis.json"
out_path.write_text(json.dumps(summary_obj, indent=2, ensure_ascii=False), encoding="utf-8")
print(f"\nWrote cause analysis to {out_path}")
