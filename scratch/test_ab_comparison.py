import urllib.parse
import urllib.request
import json

queries = ["cà phê", "bàn học", "học sinh", "trà sữa"]

print("=" * 80)
print("             KẾT QUẢ ĐỐI SOÁT TÌM KIẾM A/B: ELASTICSEARCH 8.x")
print("  (CỐC CỐC TOKENIZER + RELEVANCE BOOSTING VS STANDARD ANALYZER BASELINE)")
print("=" * 80)

for q in queries:
    url = f"http://localhost:8080/api/search/compare?q={urllib.parse.quote(q)}"
    try:
        req = urllib.request.Request(url)
        with urllib.request.urlopen(req) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            
        print(f"\n🔍 TỪ KHÓA TÌM KIẾM: '{q}'")
        print(f"⚡ Độ trễ thực thi: Cốc Cốc = {data.get('latency_vietnamese_ms', 0)}ms | Baseline = {data.get('latency_baseline_ms', 0)}ms")
        print(f"📦 Chuỗi Tokens Cốc Cốc: {data.get('tokens', [])}")
        
        v_res = data.get("vietnamese_results", [])
        b_res = data.get("baseline_results", [])
        print(f"📊 Số lượng kết quả: Cốc Cốc = {len(v_res)} tin | Baseline = {len(b_res)} tin")
        
        print("\n  [✨ CỐC CỐC TOKENIZER + BOOSTING TOP 3]:")
        for i, item in enumerate(v_res[:3], 1):
            msg = item.get("message", {})
            print(f"   {i}. [Score: {item.get('score', 0):.2f}] ({msg.get('sender')} - {msg.get('room')}): \"{msg.get('content')}\"")
            
        print("\n  [⚠️ BASELINE STANDARD ANALYZER TOP 3 (DỄ LẪN TỪ ĐƠN)]: ")
        for i, item in enumerate(b_res[:3], 1):
            msg = item.get("message", {})
            print(f"   {i}. [Score: {item.get('score', 0):.2f}] ({msg.get('sender')} - {msg.get('room')}): \"{msg.get('content')}\"")
        print("-" * 80)
    except Exception as e:
        print(f"Error querying '{q}': {e}")
