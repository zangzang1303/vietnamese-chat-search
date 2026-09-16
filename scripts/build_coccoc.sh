#!/usr/bin/env bash
set -e

# ==============================================================================
# Script: scripts/build_coccoc.sh
# Mục đích:
#   1. Biên dịch công cụ dict_compiler từ mã nguồn Cốc Cốc
#   2. Sinh các file từ điển nhị phân Double-Array Trie (.dump)
#   3. Sao chép các tệp từ điển vào thư mục data/dicts/coccoc/
# ==============================================================================

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COCCOC_SRC="${PROJECT_ROOT}/temp_coccoc"
BUILD_DIR="${COCCOC_SRC}/build"
OUTPUT_DICT_DIR="${PROJECT_ROOT}/data/dicts/coccoc"

if [ ! -d "${COCCOC_SRC}" ]; then
    echo "[0/4] Đang clone mã nguồn Cốc Cốc Tokenizer..."
    git clone https://github.com/coccoc/coccoc-tokenizer.git "${COCCOC_SRC}"
fi

echo "[1/4] Tạo thư mục build và thư mục đích..."
mkdir -p "${BUILD_DIR}"
mkdir -p "${OUTPUT_DICT_DIR}"

echo "[2/4] Cấu hình CMake và biên dịch dict_compiler..."
cd "${BUILD_DIR}"
cmake -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_CXX_FLAGS="-Wno-error -Wno-cast-user-defined" \
      "${COCCOC_SRC}"

make -j$(nproc) dict_compiler

echo "[3/4] Biên dịch dữ liệu từ điển thành Double-Array Trie dumps..."
# dict_compiler nhận vào: {thư_mục_chứa_dicts} {thư_mục_xuất_kết_quả}
./dict_compiler "${COCCOC_SRC}/dicts" "${OUTPUT_DICT_DIR}"

echo "[4/4] Sao chép các tệp cấu hình ngôn ngữ vào thư mục từ điển..."
cp "${COCCOC_SRC}/dicts/vn_lang_tool/alphabetic" "${OUTPUT_DICT_DIR}/"
cp "${COCCOC_SRC}/dicts/vn_lang_tool/numeric" "${OUTPUT_DICT_DIR}/"
cp "${COCCOC_SRC}/dicts/vn_lang_tool/d_and_gi.txt" "${OUTPUT_DICT_DIR}/"
cp "${COCCOC_SRC}/dicts/vn_lang_tool/i_and_y.txt" "${OUTPUT_DICT_DIR}/"

echo "=================================================================="
echo "Hoàn tất! Các tệp từ điển đã sẵn sàng tại: ${OUTPUT_DICT_DIR}"
ls -lh "${OUTPUT_DICT_DIR}"
echo "=================================================================="
