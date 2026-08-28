package refiner

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRefinerPipeline(t *testing.T) {
	engine := NewEngine()

	mockInfer := func(ctx context.Context, prompt string) (string, error) {
		if strings.Contains(prompt, "DEEP REQUIREMENT DECOMPOSITION") {
			return "1. Objective: Build a 2D Cyberpunk Snake game.\n2. Invariants: 60 FPS, Web Audio API, Canvas rendering.", nil
		}
		if strings.Contains(prompt, "SELF-CRITICAL INVARIANT AUDIT") {
			return "Verified: Invariants hold. Add touch event listener fallback for mobile.", nil
		}
		if strings.Contains(prompt, "FINAL VERIFIED SYNTHESIS") {
			return "<!DOCTYPE html><html><body><canvas id='c'></canvas><script>/* Snake Game */</script></body></html>", nil
		}
		return "default output", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userPrompt := "Build a 2D Cyberpunk Snake game in HTML5 Canvas with Web Audio."
	res, err := engine.Refine(ctx, userPrompt, mockInfer)
	if err != nil {
		t.Fatalf("Refine failed: %v", err)
	}

	if res.OriginalPrompt != userPrompt {
		t.Errorf("expected original prompt %q, got %q", userPrompt, res.OriginalPrompt)
	}
	if len(res.Traces) != 3 {
		t.Fatalf("expected 3 traces, got %d", len(res.Traces))
	}
	if res.Traces[0].Stage != StageDecompose {
		t.Errorf("expected stage %q, got %q", StageDecompose, res.Traces[0].Stage)
	}
	if res.Traces[1].Stage != StageAudit {
		t.Errorf("expected stage %q, got %q", StageAudit, res.Traces[1].Stage)
	}
	if res.Traces[2].Stage != StageSynthesize {
		t.Errorf("expected stage %q, got %q", StageSynthesize, res.Traces[2].Stage)
	}
	if !strings.Contains(res.FinalOutput, "<canvas id='c'>") {
		t.Errorf("expected canvas in final output, got %s", res.FinalOutput)
	}
	if res.TotalTokens <= 0 {
		t.Errorf("expected positive total tokens, got %d", res.TotalTokens)
	}
}

func TestRefinerErrorHandling(t *testing.T) {
	engine := NewEngine()

	mockErrInfer := func(ctx context.Context, prompt string) (string, error) {
		return "", errors.New("upstream model failure")
	}

	ctx := context.Background()
	_, err := engine.Refine(ctx, "test prompt", mockErrInfer)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "decomposition failed") {
		t.Errorf("expected decomposition error, got %v", err)
	}
}
