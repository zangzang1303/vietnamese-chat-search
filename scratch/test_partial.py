import urllib.request
import urllib.parse
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

test_queries = [
    'cà phê',
    'cà phe',
    'ca phe',
    'cà p',
    'học s',
    'sinh v',
    'trà s',
]

print("=== KIỂM TRA TÌM KIẾM GÕ DỞ / KHUYẾT DẤU / PREFIX ===")
for q in test_queries:
    for e in ['custom', 'es']:
        url = f'http://localhost:8080/api/search?q={urllib.parse.quote(q)}&engine={e}&tokenizer=coccoc'
        try:
            res = json.loads(urllib.request.urlopen(url).read().decode())
            hits = res.get('total_hits', 0)
            tokens = res.get('tokens', [])
            print(f'Query: "{q:8}" | Engine: {e:6} | Hits: {hits:2} | Tokens: {tokens}')
        except Exception as ex:
            print(f'Query: "{q:8}" | Engine: {e:6} | Error: {ex}')
