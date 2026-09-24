package main

import (
	"fmt"
	"path/filepath"
	"vietnamese-chat-search/pkg/tokenizer/coccoc"
)

func main() {
	dictPath, err := filepath.Abs("data/dicts/coccoc")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Using dictPath: %s\n", dictPath)

	tok, err := coccoc.New(dictPath, false)
	if err != nil {
		panic(err)
	}

	samples := []string{
		"học sinh",
		"hoc sinh",
		"trường học y sinh",
		"Chào các bạn học sinh mới vào trường",
		"uống cà phê",
	}

	for _, s := range samples {
		seg, err := tok.SegmentOriginal(s)
		tokens, _ := tok.Tokenize(s)
		fmt.Printf("\nInput: %q\n", s)
		fmt.Printf("  SegmentOriginal: %q (err: %v)\n", seg, err)
		fmt.Printf("  Tokenize:        %v\n", tokens)
	}
}
