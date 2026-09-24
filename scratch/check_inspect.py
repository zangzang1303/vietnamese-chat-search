import urllib.request
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

# Kiểm tra endpoint inspect trong chat_server
url = 'http://localhost:8080/api/inspect/2'
try:
    res = json.loads(urllib.request.urlopen(url).read().decode())
    print("=== NỘI DUNG DOC #2 TRONG INVERTED INDEX ===")
    print("Content:", res.get("content"))
    print("\n=== DANH SÁCH TERMS & EDGE N-GRAMS ĐÃ ĐƯỢC INDEX ===")
    for p in res.get("posting_list", []):
        t = p["term"]
        if any(k in t for k in ["học", "sin", "hoc"]):
            print(f"  • Term: '{t:12}' | TF: {p['term_frequency']} | Vị trí: {p['positions']}")
except Exception as e:
    print("Lỗi:", e)
