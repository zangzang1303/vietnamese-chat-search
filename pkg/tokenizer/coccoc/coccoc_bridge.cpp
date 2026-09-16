#include "coccoc_bridge.h"
#include <tokenizer/tokenizer.hpp>
#include <cstdlib>
#include <cstring>
#include <string>
#include <vector>

// ==============================================================================
// COCCOC_BRIDGE_CPP: C++ Implementation cho C Bridge
// ==============================================================================

int coccoc_init(const char* dict_path, int load_nontone) {
    if (!dict_path) return -1;
    bool nontone = (load_nontone != 0);
    return Tokenizer::instance().initialize(std::string(dict_path), nontone);
}

char* coccoc_tokenize_original(const char* text) {
    if (!text || text[0] == '\0') {
        char* empty = (char*)malloc(1);
        if (empty) empty[0] = '\0';
        return empty;
    }

    std::string input(text);
    std::vector<FullToken> res = Tokenizer::instance().segment_original(input, Tokenizer::TOKENIZE_NORMAL);

    if (res.empty()) {
        char* result = (char*)malloc(input.size() + 1);
        if (!result) return NULL;
        std::memcpy(result, input.c_str(), input.size() + 1);
        return result;
    }

    std::string output;
    output.reserve(input.size() + 16);

    for (size_t i = 0; i < res.size(); ++i) {
        size_t punct_start = (i > 0) ? res[i - 1].original_end : 0;
        size_t punct_len = res[i].original_start - punct_start;

        if (punct_len > 0) {
            output += input.substr(punct_start, punct_len);
        } else if (i > 0) {
            output += ' ';
        }

        output += res[i].text;
    }

    size_t last_end = res.back().original_end;
    if (last_end < input.size()) {
        output += input.substr(last_end);
    }

    // Cấp phát bộ nhớ C để trả về cho tầng Go
    char* result = (char*)malloc(output.size() + 1);
    if (!result) return NULL;
    std::memcpy(result, output.c_str(), output.size() + 1);
    return result;
}

void coccoc_free_string(char* str) {
    if (str) {
        free(str);
    }
}
