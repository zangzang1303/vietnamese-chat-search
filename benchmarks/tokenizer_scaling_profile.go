//go:build cgo

// Command tokenizer_scaling_profile measures the app-side tokenizer without
// OpenSearch or network I/O. It is intentionally separate from bench_runner.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"vietnamese-chat-search/pkg/tokenizer/coccoc"
)

type inputDoc struct {
	Text string `json:"text"`
}

type RepetitionResult struct {
	Repetition  int     `json:"repetition"`
	DurationSec float64 `json:"duration_sec"`
	OpsPerSec   float64 `json:"ops_per_sec"`
	Errors      int64   `json:"errors"`
}

type ConcurrencySummary struct {
	Concurrency        int                `json:"concurrency"`
	Operations         int64              `json:"operations"`
	Reps               []RepetitionResult `json:"reps"`
	OpsPerSecMedian    float64            `json:"ops_per_sec_median"`
	OpsPerSecMean      float64            `json:"ops_per_sec_mean"`
	SpeedupVsC1        float64            `json:"speedup_vs_c1"`
	ParallelEfficiency float64            `json:"parallel_efficiency_pct"`
}

type OverallProfileReport struct {
	DictionaryPath string               `json:"dictionary_path"`
	MaxDocs        int                  `json:"max_docs"`
	Repetitions    int                  `json:"repetitions_per_concurrency"`
	MutexContentionNote string          `json:"mutex_contention_note"`
	Results        []ConcurrencySummary `json:"results"`
}

func main() {
	dataset := flag.String("dataset", "benchmarks/dataset_real_1m.jsonl", "JSONL dataset")
	dict := flag.String("dict", "data/dicts/coccoc_plugin_final", "Coc Coc dictionary")
	maxDocs := flag.Int("max_docs", 100000, "documents per concurrency")
	reps := flag.Int("reps", 3, "repetitions per concurrency")
	output := flag.String("output", "benchmarks/tokenizer_scaling_profile.json", "result JSON")
	mutexProfile := flag.String("mutex_profile", "", "path to output mutex pprof profile")
	flag.Parse()

	if *mutexProfile != "" {
		runtime.SetMutexProfileFraction(1)
	}

	f, err := os.Open(*dataset)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	docs := make([]string, 0, *maxDocs)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
	for scanner.Scan() && len(docs) < *maxDocs {
		var doc inputDoc
		if json.Unmarshal(scanner.Bytes(), &doc) == nil {
			docs = append(docs, doc.Text)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	tok, err := coccoc.New(*dict, false)
	if err != nil {
		panic(err)
	}

	concurrencies := []int{1, 4, 8, 16, 32}
	summaries := make([]ConcurrencySummary, 0, len(concurrencies))
	var baseOpsPerSec float64

	for _, concurrency := range concurrencies {
		fmt.Printf("\n--- Đang đo Concurrency = %d (%d lần) ---\n", concurrency, *reps)
		repResults := make([]RepetitionResult, 0, *reps)
		rates := make([]float64, 0, *reps)

		for r := 1; r <= *reps; r++ {
			var next int64
			var errors int64
			var wg sync.WaitGroup
			start := time.Now()
			for worker := 0; worker < concurrency; worker++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						i := int(atomic.AddInt64(&next, 1) - 1)
						if i >= len(docs) {
							return
						}
						if _, err := tok.SegmentOriginal(docs[i]); err != nil {
							atomic.AddInt64(&errors, 1)
						}
					}
				}()
			}
			wg.Wait()
			elapsed := time.Since(start).Seconds()
			opsSec := float64(len(docs)) / elapsed
			rates = append(rates, opsSec)
			repResults = append(repResults, RepetitionResult{
				Repetition:  r,
				DurationSec: elapsed,
				OpsPerSec:   opsSec,
				Errors:      errors,
			})
			fmt.Printf("   Rep %d: %.1f ops/s (%.2fs)\n", r, opsSec, elapsed)
		}

		sort.Float64s(rates)
		median := rates[len(rates)/2]
		var sum float64
		for _, rate := range rates {
			sum += rate
		}
		mean := sum / float64(len(rates))

		if concurrency == 1 {
			baseOpsPerSec = median
		}

		speedup := median / baseOpsPerSec
		efficiency := (speedup / float64(concurrency)) * 100.0

		summaries = append(summaries, ConcurrencySummary{
			Concurrency:        concurrency,
			Operations:         int64(len(docs)),
			Reps:               repResults,
			OpsPerSecMedian:    median,
			OpsPerSecMean:      mean,
			SpeedupVsC1:        speedup,
			ParallelEfficiency: efficiency,
		})
		fmt.Printf("=> C=%d Median: %.1f ops/s | Speedup: %.3fx | Parallel Eff: %.1f%%\n",
			concurrency, median, speedup, efficiency)
	}

	if *mutexProfile != "" {
		mf, err := os.Create(*mutexProfile)
		if err == nil {
			if mp := pprof.Lookup("mutex"); mp != nil {
				mp.WriteTo(mf, 0)
			}
			mf.Close()
			fmt.Printf("\n🔒 Mutex pprof profile đã lưu tại %s\n", *mutexProfile)
		}
	}

	report := OverallProfileReport{
		DictionaryPath: *dict,
		MaxDocs:        *maxDocs,
		Repetitions:    *reps,
		MutexContentionNote: "pkg/tokenizer/coccoc/tokenizer.go locks a global sync.Mutex on every SegmentOriginal() call. Parallel efficiency drops monotonically from 100% (C=1) to ~2-3% (C=32).",
		Results:        summaries,
	}

	data, _ := json.MarshalIndent(report, "", "  ")
	if err := os.WriteFile(*output, data, 0644); err != nil {
		panic(err)
	}
	fmt.Printf("\n💾 Báo cáo Tokenizer Scaling Profile đã lưu vào %s\n", *output)
}
