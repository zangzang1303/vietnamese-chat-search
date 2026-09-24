import urllib.request
import urllib.parse
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

q = 'học sin'
for eng in ['custom', 'es']:
    url = f'http://localhost:8080/api/search?q={urllib.parse.quote(q)}&engine={eng}'
    try:
        res = json.loads(urllib.request.urlopen(url).read().decode())
        tokens = res.get('tokens', [])
        edge_ngrams = res.get('edge_ngrams', [])
        results = res.get('results', [])
        print(f"[{eng.upper()}] Query: '{q}'")
        print(f"   Tokens cắt ra: {tokens}")
        print(f"   Edge N-grams: {edge_ngrams}")
        print(f"   Số kết quả: {len(results)}")
        if results:
            print(f"   Top 1 (Score {results[0]['score']}): {results[0]['message']['content']}")
    except Exception as e:
        print(f"[{eng.upper()}] Error: {e}")
