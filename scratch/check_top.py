import urllib.request
import urllib.parse
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

queries = ['cà p', 'cà phe', 'ca phe', 'học s', 'sinh v', 'trà s']
for q in queries:
    print(f"\n=== TRUY VẤN: '{q}' ===")
    for eng in ['custom', 'es']:
        url = f'http://localhost:8080/api/search?q={urllib.parse.quote(q)}&engine={eng}'
        try:
            res = json.loads(urllib.request.urlopen(url).read().decode())
            results = res.get('results', [])
            edge_ngrams = res.get('edge_ngrams', [])
            print(f"[{eng.upper()}] Hits: {len(results)} | Edge N-grams: {edge_ngrams[:5]}")
            for i, r in enumerate(results[:2]):
                msg = r['message']
                print(f"   #{i+1} [ID {msg['id']}] Score {r['score']}: {msg['content']}")
        except Exception as ex:
            print(f"[{eng.upper()}] Error: {ex}")
