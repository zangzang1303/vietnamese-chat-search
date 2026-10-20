import json
import pathlib
import sys

sys.stdout.reconfigure(encoding="utf-8")

ROOT = pathlib.Path(__file__).resolve().parents[1]
MISMATCH_FILE = ROOT / "benchmarks" / "final_run" / "final_20261020_093100" / "parity" / "smoke" / "token_parity_mismatches.json"

with open(MISMATCH_FILE, encoding="utf-8") as f:
    mismatches = json.load(f)

print(f"Total mismatches: {len(mismatches)}")
field_counts = {}
for m in mismatches:
    f = m["field"]
    field_counts[f] = field_counts.get(f, 0) + 1
print("Mismatches by field:", field_counts)

for i, m in enumerate(mismatches[:10]):
    print(f"\n--- Mismatch {i+1} (id={m['id']}, field={m['field']}) ---")
    vt = [t["token"] for t in m["vanilla"]]
    pt = [t["token"] for t in m["plugin"]]
    print(f"Vanilla len={len(vt)}: {vt[:8]}")
    print(f"Plugin  len={len(pt)}: {pt[:8]}")
    # find first diff
    for j in range(max(len(vt), len(pt))):
        v = vt[j] if j < len(vt) else "<END>"
        p = pt[j] if j < len(pt) else "<END>"
        if v != p:
            print(f"  First diff at pos {j}: vanilla='{v}' vs plugin='{p}'")
            break
