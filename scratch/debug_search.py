import urllib.request
import urllib.parse
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

for q in ["hoc sinh", "học sinh"]:
    url = f"http://localhost:8080/api/search/compare?q={urllib.parse.quote(q)}"
    req = urllib.request.urlopen(url)
    data = json.loads(req.read().decode('utf-8'))
    print(f"\n==================== QUERY: '{q}' ====================")
    print("Tokens:", data.get("tokens"))
    print("\n--- TOP 3 VIETNAMESE (ES CỐC CỐC) ---")
    for i, item in enumerate(data.get("vietnamese_results", [])[:3]):
        msg = item["message"]
        print(f"#{i+1}: ID={msg['id']} | Score={item['score']:.4f} | Content: {msg['content']}")

    print("\n--- TOP 3 CUSTOM ENGINE ---")
    for i, item in enumerate(data.get("custom_results", [])[:3]):
        msg = item["message"]
        print(f"#{i+1}: ID={msg['id']} | Score={item['score']:.4f} | Content: {msg['content']}")
