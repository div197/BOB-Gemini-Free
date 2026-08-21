package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/div197/bob-gemini-free/internal/diag"
)

func main() {
	requests := flag.Int("requests", 100, "requests per concurrency profile")
	profileText := flag.String("profiles", "1,10,50,100", "comma-separated concurrency profiles")
	flag.Parse()

	var reports []diag.LocalBenchmarkReport
	for _, text := range strings.Split(*profileText, ",") {
		concurrency, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil || concurrency <= 0 {
			fmt.Fprintf(os.Stderr, "invalid concurrency profile %q\n", text)
			os.Exit(2)
		}
		reports = append(reports, diag.RunLocalBenchmark(concurrency, *requests))
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(reports); err != nil {
		fmt.Fprintf(os.Stderr, "encode benchmark report: %v\n", err)
		os.Exit(1)
	}
}
