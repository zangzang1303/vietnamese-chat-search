#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import sys
import json
import urllib.request
import urllib.error

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8')

def init_index(base_url, index_name, mapping_file):
    with open(mapping_file, "r", encoding="utf-8") as f:
        mapping_data = json.load(f)

    # 1. Xóa index nếu đã tồn tại
    delete_url = f"{base_url}/{index_name}"
    req_del = urllib.request.Request(delete_url, method="DELETE")
    try:
        urllib.request.urlopen(req_del)
        print(f"🗑️ Đã xóa index cũ tại {base_url}/{index_name}")
    except urllib.error.HTTPError as e:
        if e.code != 404:
            print(f"⚠️ Không thể xóa index: {e}")

    # 2. Tạo mới index với mapping và settings
    create_url = f"{base_url}/{index_name}"
    payload = json.dumps(mapping_data).encode("utf-8")
    req_create = urllib.request.Request(create_url, data=payload, headers={"Content-Type": "application/json"}, method="PUT")
    try:
        with urllib.request.urlopen(req_create) as resp:
            print(f"✅ Tạo mới thành công index '{index_name}' tại {base_url} (HTTP {resp.status})")
    except Exception as e:
        print(f"❌ Lỗi tạo index tại {base_url}: {e}")

if __name__ == "__main__":
    print("🚀 Bắt đầu khởi tạo các index cho Benchmark...")
    init_index("http://localhost:9201", "messages", "benchmarks/mapping_vanilla.json")
    init_index("http://localhost:9202", "messages", "benchmarks/mapping_plugin.json")
    print("✨ Hoàn tất khởi tạo môi trường!")
