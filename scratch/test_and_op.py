import urllib.request
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

# Thử query với operator AND và match exact trên content_partial
queryMap = {
    "size": 5,
    "query": {
        "bool": {
            "must": [
                {
                    "multi_match": {
                        "query": "học sin",
                        "type": "most_fields",
                        "operator": "and",
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
                            "boost": 10.0
                        }
                    }
                },
                {
                    "match": {
                        "content_partial": {
                            "query": "hoc_sin",
                            "boost": 8.0
                        }
                    }
                }
            ]
        }
    }
}

req = urllib.request.Request(
    'http://localhost:9200/chat_messages_vietnamese/_search',
    data=json.dumps(queryMap).encode('utf-8'),
    headers={'Content-Type': 'application/json'}
)
res = json.loads(urllib.request.urlopen(req).read().decode())
print("=== KẾT QUẢ KHI CÓ OPERATOR: AND ===")
for hit in res['hits']['hits']:
    print(f"ID {hit['_id']} | Score: {hit['_score']}")
    print(f"   {hit['_source']['content']}")
