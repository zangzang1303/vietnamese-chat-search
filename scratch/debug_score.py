import urllib.request
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

with open('scratch/test_explain.py', 'r', encoding='utf-8') as f:
    text = f.read()

# Gọi explain API cho ID 108 và ID 2
for doc_id in ['108', '2']:
    url = f"http://localhost:9200/chat_messages_vietnamese/_explain/{doc_id}"
    body = {
        "query": {
            "bool": {
                "must": [
                    {
                        "multi_match": {
                            "query": "học sin",
                            "type": "most_fields",
                            "fields": [
                                "content_tokenized^5.0",
                                "content_unaccented^3.0",
                                "content_partial^2.0"
                            ]
                        }
                    }
                ],
                "should": [
                    {
                        "match": {
                            "content_partial": {
                                "query": "học_sin",
                                "boost": 3.5
                            }
                        }
                    }
                ]
            }
        }
    }
    req = urllib.request.Request(url, data=json.dumps(body).encode('utf-8'), headers={'Content-Type': 'application/json'})
    res = json.loads(urllib.request.urlopen(req).read().decode())
    print(f"=== EXPLAIN DOC {doc_id} (Score {res['explanation']['value']}) ===")
    
    def print_expl(node, indent=0):
        val = node.get('value', 0)
        desc = node.get('description', '')
        print("  " * indent + f"- [{val:.2f}] {desc}")
        for child in node.get('details', [])[:3]:
            print_expl(child, indent + 1)
            
    print_expl(res['explanation'])
