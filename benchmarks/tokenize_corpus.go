package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"vietnamese-chat-search/pkg/tokenizer/coccoc"
)

type inputDocument struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type outputDocument struct {
	ID        int64  `json:"id"`
	Text      string `json:"text"`
	Segmented string `json:"segmented"`
	Error     string `json:"error,omitempty"`
}

func main() {
	input := flag.String("input", "benchmarks/parity_corpus.jsonl", "input JSONL")
	output := flag.String("output", "benchmarks/parity_corpus_tokenized.jsonl", "output JSONL")
	dict := flag.String("dict", "data/dicts/coccoc", "Coc Coc dictionary directory")
	flag.Parse()

	tok, err := coccoc.New(*dict, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialize tokenizer: %v\n", err)
		os.Exit(1)
	}

	in, err := os.Open(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open input: %v\n", err)
		os.Exit(1)
	}
	defer in.Close()

	out, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create output: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	encoder := json.NewEncoder(out)
	count := 0
	errors := 0
	for scanner.Scan() {
		var doc inputDocument
		if err := json.Unmarshal(scanner.Bytes(), &doc); err != nil {
			errors++
			continue
		}
		segmented, segErr := tok.SegmentOriginal(doc.Text)
		result := outputDocument{ID: doc.ID, Text: doc.Text, Segmented: segmented}
		if segErr != nil {
			result.Error = segErr.Error()
			errors++
		}
		if err := encoder.Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "encode output: %v\n", err)
			os.Exit(1)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "scan input: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("tokenized=%d errors=%d\n", count, errors)
	if errors > 0 {
		os.Exit(2)
	}
}
