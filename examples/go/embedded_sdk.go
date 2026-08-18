package main

import (
	"context"
	"fmt"
	"log"

	"github.com/div197/bob-gemini-free/pkg/gateway"
)

func main() {
	// 1. Initialize embedded in-process Gemini engine (zero HTTP port needed)
	engine := gateway.NewEngine(
		gateway.WithDefaultModel("gemini-3.7-flash"),
		gateway.WithLogRequests(false),
	)

	ctx := context.Background()

	// 2. Synchronous in-process inference
	fmt.Println("--- 1. Synchronous In-Process Go Inference ---")
	resp, err := engine.Generate(ctx, "Explain goroutines in Go in 2 sentences.", "gemini-3.7-flash")
	if err != nil {
		log.Fatalf("Generate failed: %v", err)
	}
	fmt.Printf("Response:\n%s\n\n", resp)

	// 3. Real-time streaming in-process inference
	fmt.Println("--- 2. Real-Time Streaming In-Process Go Inference ---")
	err = engine.GenerateStream(ctx, "Write a bubble sort function in Go.", "gemini-3.7-flash", func(delta string) error {
		fmt.Print(delta)
		return nil
	})
	if err != nil {
		log.Fatalf("Stream failed: %v", err)
	}
	fmt.Println("\n\nDone!")
}
