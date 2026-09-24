import urllib.request
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

combos = [
    ('es', 'coccoc'),
    ('custom', 'coccoc'),
    ('es', 'standard'),
    ('custom', 'standard'),
]

print("=== KIỂM TRA TOÀN DIỆN LAB API (4 CẤU HÌNH TỰ CHỌN) ===")
for engine, tokenizer in combos:
    url = f"http://localhost:8080/api/search?q=h%E1%BB%8Dc+sinh&engine={engine}&tokenizer={tokenizer}"
    req = urllib.request.urlopen(url)
    data = json.loads(req.read().decode('utf-8'))
    
    hits = data.get('total_hits', 0)
    latency = data.get('latency_ms', 0)
    tokens = data.get('tokens', [])
    engine_name = data.get('engine_name', '')
    tok_name = data.get('tokenizer_name', '')
    scoring = data.get('scoring_formula', '')
    top_content = data['results'][0]['message']['content'] if data.get('results') else "Không có"
    
    print(f"\n⚙️ CẤU HÌNH: [{engine.upper()}] + [{tokenizer.upper()}]")
    print(f"   • Động cơ      : {engine_name}")
    print(f"   • Bộ tách từ   : {tok_name}")
    print(f"   • Chuỗi Tokens : {tokens}")
    print(f"   • Số kết quả   : {hits} tin nhắn")
    print(f"   • Độ trễ       : {latency} ms")
    print(f"   • Chấm điểm    : {scoring}")
    print(f"   • Top 1 nội dung: \"{top_content[:55]}...\"")
