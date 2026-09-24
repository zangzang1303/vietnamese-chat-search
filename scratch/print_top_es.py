import urllib.request
import urllib.parse
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

q = 'học sin'
url = f'http://localhost:8080/api/search?q={urllib.parse.quote(q)}&engine=es'
res = json.loads(urllib.request.urlopen(url).read().decode())
print(f"=== TOP 5 KẾT QUẢ ELASTICSEARCH CHO '{q}' ===")
for i, r in enumerate(res['results'][:5]):
    msg = r['message']
    print(f"#{i+1} [ID {msg['id']}] Score: {r['score']:.3f} | Sender: {msg['sender']}")
    print(f"    Content: {msg['content']}")
