// Package refiner provides a three-stage reasoning and prompt-refinement
// orchestration. It calls a supplied inference function for decomposition,
// audit, and synthesis; it does not provide a local model or guarantee zero
// cloud use.
package refiner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Stage represents an execution phase in the deep reasoning refinement pipeline.
type Stage string

const (
	StageDecompose   Stage = "decompose"   // Stage 1: Deconstruct constraints, invariants & requirements
	StageHypothesize Stage = "hypothesize" // Stage 2: Formulate solution hypotheses & algorithms
	StageAudit       Stage = "audit"       // Stage 3: Self-critical verification & boundary checking
	StageSynthesize  Stage = "synthesize"  // Stage 4: Clean final implementation without preamble
)

const (
	// MaxUserPromptBytes bounds the prompt before it is repeated into the
	// decomposition, audit, and synthesis instructions.
	MaxUserPromptBytes = 256 << 10
	// MaxStageOutputBytes bounds intermediate model output before it is copied
	// into the next stage's prompt and retained in the result traces.
	MaxStageOutputBytes = 4 << 20
)

// ThoughtTrace holds the internal reasoning steps generated during refinement.
type ThoughtTrace struct {
	Stage     Stage         `json:"stage"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	Duration  time.Duration `json:"duration_ms"`
	TokensEst int           `json:"tokens_est"`
}

// RefinementResult contains the final refined output alongside all reasoning traces.
type RefinementResult struct {
	OriginalPrompt string         `json:"original_prompt"`
	Traces         []ThoughtTrace `json:"traces"`
	FinalOutput    string         `json:"final_output"`
	TotalTokens    int            `json:"total_tokens"`
	TotalDuration  time.Duration  `json:"total_duration"`
}

// InferenceFunc defines the inference operation supplied by the caller.
type InferenceFunc func(ctx context.Context, prompt string) (string, error)

// Engine manages multi-stage reasoning refinement.
type Engine struct {
	mu sync.RWMutex
}

// NewEngine creates a new instance of the local reasoning refiner.
func NewEngine() *Engine {
	return &Engine{}
}

// BuildDecompositionPrompt constructs a prompt to break down user requests into explicit invariants.
func BuildDecompositionPrompt(userPrompt string) string {
	return fmt.Sprintf(`### INSTRUCTION: DEEP REQUIREMENT DECOMPOSITION
You are an expert system architect and algorithm designer. Analyze the following request.
Do NOT write the complete final code yet. Instead, break it down into:
1. Core Objective & Functional Scope
2. Critical Constraints & Boundary Conditions (edge cases, performance limits)
3. Architectural Blueprint & Key Data Structures
4. Verification Criteria (how to mathematically or programmatically prove success)

---
USER REQUEST:
%s
---
Output your analysis cleanly with numbered sections.`, userPrompt)
}

// BuildAuditPrompt constructs a self-critical verification prompt.
func BuildAuditPrompt(userPrompt, plan string) string {
	return fmt.Sprintf(`### INSTRUCTION: SELF-CRITICAL INVARIANT AUDIT
Review the proposed implementation plan against the original user request.
Identify any missing requirements, subtle bugs, performance bottlenecks, or boundary edge cases.

ORIGINAL REQUEST:
%s

PROPOSED PLAN:
%s

List only actionable corrections and improvements to ensure 100%% precision.`, userPrompt, plan)
}

// BuildSynthesisPrompt constructs the final execution prompt.
func BuildSynthesisPrompt(userPrompt, plan, audit string) string {
	return fmt.Sprintf(`### INSTRUCTION: FINAL VERIFIED SYNTHESIS
Based on the verified architecture and audit notes below, generate the pristine, complete, production-ready implementation.
Do NOT include fluff, boilerplate apologies, or conversational filler. Output the exact self-contained result.

VERIFIED ARCHITECTURE:
%s

AUDIT NOTES:
%s

USER SPECIFICATION:
%s`, plan, audit, userPrompt)
}

// Refine executes the three-stage reasoning pipeline through the supplied
// inference function.
func (e *Engine) Refine(ctx context.Context, userPrompt string, infer InferenceFunc) (*RefinementResult, error) {
	if e == nil {
		return nil, errors.New("refiner engine is nil")
	}
	if infer == nil {
		return nil, errors.New("refiner inference function is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(userPrompt) == "" {
		return nil, errors.New("refiner prompt is empty")
	}
	if len(userPrompt) > MaxUserPromptBytes {
		return nil, fmt.Errorf("refiner prompt exceeds %d bytes", MaxUserPromptBytes)
	}

	startTime := time.Now()
	res := &RefinementResult{
		OriginalPrompt: userPrompt,
		Traces:         make([]ThoughtTrace, 0, 3),
	}

	// 1. Stage 1: Deconstruct & Plan
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	t0 := time.Now()
	decompPrompt := BuildDecompositionPrompt(userPrompt)
	planOutput, err := infer(ctx, decompPrompt)
	if err != nil {
		return nil, fmt.Errorf("decomposition failed: %w", err)
	}
	if err := validateStageOutput("decomposition", planOutput); err != nil {
		return nil, err
	}
	res.Traces = append(res.Traces, ThoughtTrace{
		Stage:     StageDecompose,
		Title:     "🧠 Requirement Decomposition & Blueprint",
		Content:   planOutput,
		Duration:  time.Since(t0),
		TokensEst: len(strings.Fields(planOutput)),
	})

	// 2. Stage 2: Audit & Self-Correction
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	t1 := time.Now()
	auditPrompt := BuildAuditPrompt(userPrompt, planOutput)
	auditOutput, err := infer(ctx, auditPrompt)
	if err != nil {
		return nil, fmt.Errorf("audit failed: %w", err)
	}
	if err := validateStageOutput("audit", auditOutput); err != nil {
		return nil, err
	}
	res.Traces = append(res.Traces, ThoughtTrace{
		Stage:     StageAudit,
		Title:     "🔍 Invariant & Boundary Verification",
		Content:   auditOutput,
		Duration:  time.Since(t1),
		TokensEst: len(strings.Fields(auditOutput)),
	})

	// 3. Stage 3: Final Synthesis
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	t2 := time.Now()
	synthPrompt := BuildSynthesisPrompt(userPrompt, planOutput, auditOutput)
	finalCode, err := infer(ctx, synthPrompt)
	if err != nil {
		return nil, fmt.Errorf("synthesis failed: %w", err)
	}
	if err := validateStageOutput("synthesis", finalCode); err != nil {
		return nil, err
	}
	res.Traces = append(res.Traces, ThoughtTrace{
		Stage:     StageSynthesize,
		Title:     "⚡ Verified Output Synthesis",
		Content:   "Synthesis completed successfully.",
		Duration:  time.Since(t2),
		TokensEst: len(strings.Fields(finalCode)),
	})

	res.FinalOutput = finalCode
	res.TotalDuration = time.Since(startTime)

	for _, tr := range res.Traces {
		res.TotalTokens += tr.TokensEst
	}

	return res, nil
}

func validateStageOutput(stage, output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("%s returned no usable output", stage)
	}
	if len(output) > MaxStageOutputBytes {
		return fmt.Errorf("%s output exceeds %d bytes", stage, MaxStageOutputBytes)
	}
	return nil
}
