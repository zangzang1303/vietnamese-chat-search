import urllib.request
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

queryMap = {
    "size": 5,
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
                    "match_phrase": {
                        "content_tokenized": {
                            "query": "học sin",
                            "boost": 4.0
                        }
                    }
                },
                {
                    "match": {
                        "content_unaccented": {
                            "query": "hoc sin",
                            "boost": 2.0
                        }
                    }
                },
                {
                    "match": {
                        "content_partial": {
                            "query": "học_sin",
                            "boost": 3.5
                        }
                    }
                },
                {
                    "match": {
                        "content_partial": {
                            "query": "hoc_sin",
                            "boost": 2.5
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
for hit in res['hits']['hits']:
    print(f"ID {hit['_id']} | Score: {hit['_score']}")
    print(f"   {hit['_source']['content']}")
