package models

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Model represents a Gemini backend routing configuration.
// Mode corresponds to the Gemini upstream mode ID, Think specifies the default reasoning depth,
// and Extra contains index-based overrides for Gemini's sparse RPC request array.
type Model struct {
	Mode  int         `json:"mode"`
	Think int         `json:"think"`
	Desc  string      `json:"desc"`
	Extra map[int]any `json:"extra,omitempty"`
}

var MODELS = map[string]Model{
	"gemini-3.7-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Latest flagship fast model (Gemini 3.7 Flash)",
	},
	"gemini-3.7-flash-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Latest flagship deep thinking mode (Gemini 3.7 Flash Thinking)",
	},
	"gemini-3.6-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "All-around model (Gemini 3.6 Flash)",
	},
	"gemini-3.5-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for gemini-3.6-flash",
	},
	"gemini-3.5-flash-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Deep thinking mode, longest output (~20k chars)",
	},
	"gemini-3.1-pro": {
		Mode:  3,
		Think: 4,
		Desc:  "Pro model (requires cookie for real routing)",
	},
	"gemini-3.1-pro-enhanced": {
		Mode:  3,
		Think: 4,
		Desc:  "Pro with enhanced output (experimental)",
		Extra: map[int]any{31: 2, 80: 3},
	},
	"gemini-auto": {
		Mode:  4,
		Think: 4,
		Desc:  "Auto model selection",
	},
	"gemini-3.5-flash-thinking-lite": {
		Mode:  5,
		Think: 0,
		Desc:  "Dynamic thinking with adaptive depth",
	},
	"gemini-flash-lite": {
		Mode:  6,
		Think: 4,
		Desc:  "Lightweight fast model",
	},
	// Developer Convenience Aliases
	"gemini-pro": {
		Mode:  3,
		Think: 4,
		Desc:  "Alias for gemini-3.1-pro",
	},
	"gemini-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for gemini-3.6-flash",
	},
	"gemini-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Alias for gemini-3.5-flash-thinking",
	},
	"gemini-lite": {
		Mode:  6,
		Think: 4,
		Desc:  "Alias for gemini-flash-lite",
	},
	"gemini-2.5-pro": {
		Mode:  3,
		Think: 4,
		Desc:  "Alias for gemini-3.1-pro",
	},
	"gemini-2.5-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for gemini-3.6-flash",
	},
	"gemini-2.5-flash-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Alias for gemini-3.7-flash-thinking",
	},
	"gemini-2.0-pro-exp-02-05": {
		Mode:  2,
		Think: 0,
		Desc:  "Gemini 2.0 Pro Experimental Reasoning alias (routed to gemini-3.7-flash-thinking)",
	},
	"gemini-exp-1206": {
		Mode:  2,
		Think: 0,
		Desc:  "Gemini Experimental Frontier alias (routed to gemini-3.7-flash-thinking)",
	},
	"gemini-2.0-flash-exp": {
		Mode:  1,
		Think: 4,
		Desc:  "Gemini 2.0 Flash Experimental alias (routed to gemini-3.7-flash)",
	},
	"gemini-2.0-flash-lite": {
		Mode:  6,
		Think: 4,
		Desc:  "Gemini 2.0 Flash Lite alias (routed to gemini-flash-lite)",
	},
	"gemini-2.0-flash-lite-preview-02-05": {
		Mode:  6,
		Think: 4,
		Desc:  "Gemini 2.0 Flash Lite Preview alias (routed to gemini-flash-lite)",
	},
	"gemini-1.5-flash-8b": {
		Mode:  6,
		Think: 4,
		Desc:  "Gemini 1.5 Flash 8B lightweight alias (routed to gemini-flash-lite)",
	},
	"gemini-2.0-flash-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Alias for gemini-3.5-flash-thinking",
	},
	"gemini-2.0-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for gemini-3.6-flash",
	},
	"gemini-1.5-pro": {
		Mode:  3,
		Think: 4,
		Desc:  "Alias for gemini-3.1-pro",
	},
	"gemini-1.5-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for gemini-3.6-flash",
	},
	// OpenAI Drop-in Aliases (Frontier, Reasoning & Codex)
	"gpt-5.6":           {Mode: 1, Think: 4, Desc: "OpenAI GPT-5.6 alias (routed to gemini-3.7-flash)"},
	"gpt-5.6-sol":       {Mode: 1, Think: 4, Desc: "OpenAI GPT-5.6 Sol alias (routed to gemini-3.7-flash)"},
	"gpt-5.6-terra":     {Mode: 1, Think: 4, Desc: "OpenAI GPT-5.6 Terra alias (routed to gemini-3.7-flash)"},
	"gpt-5.6-luna":      {Mode: 6, Think: 4, Desc: "OpenAI GPT-5.6 Luna alias (routed to gemini-flash-lite)"},
	"gpt-5.6-cyber":     {Mode: 3, Think: 4, Desc: "OpenAI Daybreak/Cyber alias (routed to gemini-3.1-pro)"},
	"gpt-5.5":           {Mode: 1, Think: 4, Desc: "OpenAI GPT-5.5 alias (routed to gemini-3.7-flash)"},
	"gpt-5.5-pro":       {Mode: 3, Think: 4, Desc: "OpenAI GPT-5.5 Pro alias (routed to gemini-3.1-pro)"},
	"gpt-5.4":           {Mode: 1, Think: 4, Desc: "OpenAI GPT-5.4 alias (routed to gemini-3.7-flash)"},
	"gpt-5.4-mini":      {Mode: 6, Think: 4, Desc: "OpenAI GPT-5.4 mini alias (routed to gemini-flash-lite)"},
	"gpt-5.4-pro":       {Mode: 3, Think: 4, Desc: "OpenAI GPT-5.4 Pro alias (routed to gemini-3.1-pro)"},
	"gpt-5":             {Mode: 1, Think: 4, Desc: "OpenAI GPT-5 alias (routed to gemini-3.7-flash)"},
	"gpt-5-pro":         {Mode: 3, Think: 4, Desc: "OpenAI GPT-5 Pro alias (routed to gemini-3.1-pro)"},
	"gpt-5-codex":       {Mode: 1, Think: 4, Desc: "OpenAI Codex alias (routed to gemini-3.7-flash)"},
	"gpt-5.1-codex":     {Mode: 1, Think: 4, Desc: "OpenAI Codex alias (routed to gemini-3.7-flash)"},
	"gpt-5.2-codex":     {Mode: 1, Think: 4, Desc: "OpenAI Codex alias (routed to gemini-3.7-flash)"},
	"gpt-5.3-codex":     {Mode: 1, Think: 4, Desc: "OpenAI Codex alias (routed to gemini-3.7-flash)"},
	"codex-mini-latest": {Mode: 6, Think: 4, Desc: "OpenAI Codex mini alias (routed to gemini-flash-lite)"},
	"gpt-4o":            {Mode: 1, Think: 4, Desc: "OpenAI GPT-4o alias (routed to gemini-3.7-flash)"},
	"gpt-4o-2024-11-20": {Mode: 1, Think: 4, Desc: "OpenAI GPT-4o date-versioned alias (routed to gemini-3.7-flash)"},
	"gpt-4o-2024-08-06": {Mode: 1, Think: 4, Desc: "OpenAI GPT-4o structured outputs alias (routed to gemini-3.7-flash)"},
	"gpt-4o-2024-05-13": {Mode: 1, Think: 4, Desc: "OpenAI GPT-4o original release alias (routed to gemini-3.7-flash)"},
	"gpt-4o-mini":       {Mode: 6, Think: 4, Desc: "OpenAI GPT-4o mini alias (routed to gemini-flash-lite)"},
	"gpt-4o-mini-2024-07-18": {Mode: 6, Think: 4, Desc: "OpenAI GPT-4o mini date-versioned alias (routed to gemini-flash-lite)"},
	"chatgpt-4o-latest": {Mode: 1, Think: 4, Desc: "OpenAI ChatGPT latest dynamic alias (routed to gemini-3.7-flash)"},
	"gpt-4.1":           {Mode: 1, Think: 4, Desc: "OpenAI GPT-4.1 alias (routed to gemini-3.7-flash)"},
	"gpt-4.1-mini":      {Mode: 6, Think: 4, Desc: "OpenAI GPT-4.1 mini alias (routed to gemini-flash-lite)"},
	"gpt-4-turbo":       {Mode: 1, Think: 4, Desc: "OpenAI GPT-4 Turbo alias (routed to gemini-3.7-flash)"},
	"gpt-4-turbo-preview": {Mode: 1, Think: 4, Desc: "OpenAI GPT-4 Turbo Preview alias (routed to gemini-3.7-flash)"},
	"gpt-4":             {Mode: 1, Think: 4, Desc: "OpenAI GPT-4 classic alias (routed to gemini-3.7-flash)"},
	"gpt-3.5-turbo":     {Mode: 6, Think: 4, Desc: "OpenAI GPT-3.5 Turbo alias (routed to gemini-flash-lite)"},
	"gpt-3.5-turbo-0125": {Mode: 6, Think: 4, Desc: "OpenAI GPT-3.5 Turbo date-versioned alias (routed to gemini-flash-lite)"},
	"chat-latest":       {Mode: 1, Think: 4, Desc: "OpenAI ChatGPT latest alias (routed to gemini-3.7-flash)"},
	"o3":                {Mode: 2, Think: 0, Desc: "OpenAI o3 reasoning alias (routed to gemini-3.7-flash-thinking)"},
	"o3-mini":           {Mode: 2, Think: 0, Desc: "OpenAI o3-mini reasoning alias (routed to gemini-3.7-flash-thinking)"},
	"o3-mini-2025-01-31": {Mode: 2, Think: 0, Desc: "OpenAI o3-mini date-versioned alias (routed to gemini-3.7-flash-thinking)"},
	"o3-pro":            {Mode: 3, Think: 4, Desc: "OpenAI o3-pro reasoning alias (routed to gemini-3.1-pro)"},
	"o4-mini":           {Mode: 2, Think: 0, Desc: "OpenAI o4-mini reasoning alias (routed to gemini-3.7-flash-thinking)"},
	"o1":                {Mode: 2, Think: 0, Desc: "OpenAI o1 reasoning alias (routed to gemini-3.7-flash-thinking)"},
	"o1-2024-12-17":     {Mode: 2, Think: 0, Desc: "OpenAI o1 date-versioned alias (routed to gemini-3.7-flash-thinking)"},
	"o1-preview":        {Mode: 2, Think: 0, Desc: "OpenAI o1-preview reasoning alias (routed to gemini-3.7-flash-thinking)"},
	"o1-preview-2024-09-12": {Mode: 2, Think: 0, Desc: "OpenAI o1-preview date-versioned alias (routed to gemini-3.7-flash-thinking)"},
	"o1-mini":           {Mode: 2, Think: 0, Desc: "OpenAI o1-mini reasoning alias (routed to gemini-3.7-flash-thinking)"},
	"o1-pro":            {Mode: 3, Think: 4, Desc: "OpenAI o1-pro reasoning alias (routed to gemini-3.1-pro)"},
	"gemini-3.5-flash-lite": {
		Mode:  6,
		Think: 4,
		Desc:  "Gemini 3.5 Flash Lite high-throughput model",
	},
	"gemini-3.1-pro-preview": {
		Mode:  3,
		Think: 4,
		Desc:  "Gemini 3.1 Pro Preview advanced reasoning model",
	},
	"gemini-3.1-flash-lite": {
		Mode:  6,
		Think: 4,
		Desc:  "Gemini 3.1 Flash Lite lightweight model",
	},
	"gemini-omni-1.1-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Gemini Omni 1.1 Flash conversational multimodal model",
	},
	"gpt-4.5":           {Mode: 1, Think: 4, Desc: "OpenAI GPT-4.5 research preview alias (routed to gemini-3.7-flash)"},
	"gpt-4.5-preview":   {Mode: 1, Think: 4, Desc: "OpenAI GPT-4.5 Preview alias (routed to gemini-3.7-flash)"},
	"gpt-4.5-preview-2025-02-27": {Mode: 1, Think: 4, Desc: "OpenAI GPT-4.5 date-versioned alias (routed to gemini-3.7-flash)"},
	// Anthropic Claude 5, 4.5 & 4 Family
	"claude-5-sonnet":            {Mode: 2, Think: 0, Desc: "Anthropic Claude Sonnet 5 flagship daily driver (routed to gemini-3.7-flash-thinking)"},
	"claude-5-opus":              {Mode: 3, Think: 4, Desc: "Anthropic Claude Opus 5 flagship agentic model (routed to gemini-3.1-pro)"},
	"claude-5-fable":             {Mode: 3, Think: 4, Desc: "Anthropic Claude Fable 5 frontier intelligence model (routed to gemini-3.1-pro)"},
	"claude-sonnet-5":            {Mode: 2, Think: 0, Desc: "Anthropic Claude Sonnet 5 alias (routed to gemini-3.7-flash-thinking)"},
	"claude-opus-5":              {Mode: 3, Think: 4, Desc: "Anthropic Claude Opus 5 alias (routed to gemini-3.1-pro)"},
	"claude-fable-5":             {Mode: 3, Think: 4, Desc: "Anthropic Claude Fable 5 alias (routed to gemini-3.1-pro)"},
	"claude-4-5-sonnet":          {Mode: 2, Think: 0, Desc: "Anthropic Claude 4.5 Sonnet alias (routed to gemini-3.7-flash-thinking)"},
	"claude-4-5-haiku":           {Mode: 6, Think: 4, Desc: "Anthropic Claude 4.5 Haiku alias (routed to gemini-flash-lite)"},
	"claude-4-5-opus":            {Mode: 3, Think: 4, Desc: "Anthropic Claude 4.5 Opus alias (routed to gemini-3.1-pro)"},
	"claude-haiku-4.5":           {Mode: 6, Think: 4, Desc: "Anthropic Claude Haiku 4.5 alias (routed to gemini-flash-lite)"},
	"claude-opus-4.8":            {Mode: 3, Think: 4, Desc: "Anthropic Claude Opus 4.8 alias (routed to gemini-3.1-pro)"},
	"claude-opus-4.7":            {Mode: 3, Think: 4, Desc: "Anthropic Claude Opus 4.7 alias (routed to gemini-3.1-pro)"},
	"claude-opus-4.6":            {Mode: 3, Think: 4, Desc: "Anthropic Claude Opus 4.6 alias (routed to gemini-3.1-pro)"},
	"claude-4-sonnet":            {Mode: 2, Think: 0, Desc: "Anthropic Claude 4 Sonnet alias (routed to gemini-3.7-flash-thinking)"},
	"claude-4-opus":              {Mode: 3, Think: 4, Desc: "Anthropic Claude 4 Opus alias (routed to gemini-3.1-pro)"},
	"claude-4-haiku":             {Mode: 6, Think: 4, Desc: "Anthropic Claude 4 Haiku alias (routed to gemini-flash-lite)"},
	"claude-3-7-sonnet":          {Mode: 2, Think: 0, Desc: "Anthropic Claude 3.7 Sonnet alias (routed to gemini-3.7-flash-thinking)"},
	"claude-3-7-sonnet-latest":   {Mode: 2, Think: 0, Desc: "Anthropic Claude 3.7 Sonnet alias (routed to gemini-3.7-flash-thinking)"},
	"claude-3-7-sonnet-20250219": {Mode: 2, Think: 0, Desc: "Anthropic Claude 3.7 Sonnet date-versioned alias (routed to gemini-3.7-flash-thinking)"},
	"claude-3-5-sonnet":          {Mode: 2, Think: 0, Desc: "Anthropic Claude 3.5 Sonnet alias (routed to gemini-3.7-flash-thinking)"},
	"claude-3-5-sonnet-latest":   {Mode: 2, Think: 0, Desc: "Anthropic Claude 3.5 Sonnet alias (routed to gemini-3.7-flash-thinking)"},
	"claude-3-5-sonnet-20241022": {Mode: 2, Think: 0, Desc: "Anthropic Claude 3.5 Sonnet v2 date-versioned alias (routed to gemini-3.7-flash-thinking)"},
	"claude-3-5-sonnet-20240620": {Mode: 2, Think: 0, Desc: "Anthropic Claude 3.5 Sonnet v1 date-versioned alias (routed to gemini-3.7-flash-thinking)"},
	"claude-3-5-haiku":           {Mode: 6, Think: 4, Desc: "Anthropic Claude 3.5 Haiku alias (routed to gemini-flash-lite)"},
	"claude-3-5-haiku-20241022":  {Mode: 6, Think: 4, Desc: "Anthropic Claude 3.5 Haiku date-versioned alias (routed to gemini-flash-lite)"},
	"claude-3-haiku-20240307":    {Mode: 6, Think: 4, Desc: "Anthropic Claude 3 Haiku date-versioned alias (routed to gemini-flash-lite)"},
	"claude-3-opus":              {Mode: 3, Think: 4, Desc: "Anthropic Claude 3 Opus alias (routed to gemini-3.1-pro)"},
	"claude-3-opus-20240229":     {Mode: 3, Think: 4, Desc: "Anthropic Claude 3 Opus date-versioned alias (routed to gemini-3.1-pro)"},
	"claude-sonnet-4":            {Mode: 2, Think: 0, Desc: "Anthropic Claude Sonnet 4 frontier alias (routed to gemini-3.7-flash-thinking)"},
	"claude-opus-4":              {Mode: 3, Think: 4, Desc: "Anthropic Claude Opus 4 frontier alias (routed to gemini-3.1-pro)"},
	"claude-code":                {Mode: 2, Think: 0, Desc: "Anthropic Claude Code alias (routed to gemini-3.7-flash-thinking)"},
	// DeepSeek & Open Weights Drop-in Aliases
	"deepseek-r1":                {Mode: 2, Think: 0, Desc: "DeepSeek R1 reasoning alias (routed to gemini-3.7-flash-thinking)"},
	"deepseek-reasoner":          {Mode: 2, Think: 0, Desc: "DeepSeek Reasoner API alias (routed to gemini-3.7-flash-thinking)"},
	"deepseek-v3":                {Mode: 1, Think: 4, Desc: "DeepSeek V3 chat alias (routed to gemini-3.7-flash)"},
	"deepseek-chat":              {Mode: 1, Think: 4, Desc: "DeepSeek Chat API alias (routed to gemini-3.7-flash)"},
	"llama-3.3-70b":              {Mode: 1, Think: 4, Desc: "Meta LLaMA 3.3 70B alias (routed to gemini-3.7-flash)"},
	"llama-3.1-405b":             {Mode: 3, Think: 4, Desc: "Meta LLaMA 3.1 405B flagship alias (routed to gemini-3.1-pro)"},
	"qwen-2.5-coder-32b":         {Mode: 1, Think: 4, Desc: "Qwen 2.5 Coder 32B alias (routed to gemini-3.7-flash)"},
	// Google Image Generation & Multimodal Nano Banana Models
	"imagen-3":                {Mode: 1, Think: 4, Desc: "Google Imagen 3 High-Fidelity Photorealistic Image Generation Model"},
	"imagen-3-fast":           {Mode: 1, Think: 4, Desc: "Google Imagen 3 Fast Generation Model"},
	"imagen-3.0-generate-002": {Mode: 1, Think: 4, Desc: "Google Imagen 3.0 Generate 002"},
	"gemini-nano-banana":      {Mode: 1, Think: 4, Desc: "Google Gemini Nano Banana Multimodal Image Generation Model"},
	"gemini-nano-banana-2":    {Mode: 1, Think: 4, Desc: "Google Gemini Nano Banana 2 Native Image Generation Model"},
	"gemini-nano-banana-pro":  {Mode: 3, Think: 4, Desc: "Google Gemini Nano Banana Pro High-Resolution Image Synthesis Model"},
	"dall-e-3":                {Mode: 1, Think: 4, Desc: "OpenAI DALL-E 3 alias (routed to Google Imagen 3 / Gemini Nano Banana)"},
	"dall-e-2":                {Mode: 1, Think: 4, Desc: "OpenAI DALL-E 2 alias (routed to Google Imagen 3 Fast)"},
}

type Resolved struct {
	Name  string
	Mode  int
	Think int
	Extra map[int]any
}

const DefaultModelName = "gemini-3.6-flash"

// Resolve maps a requested model string to its corresponding backend configuration.
// It supports model name suffixes like "@think=N" (e.g. "gemini-3.6-flash@think=0")
// to dynamically override the model's default thinking depth without modifying the model mapping.
func Resolve(modelName, defaultName string) (Resolved, error) {
	if defaultName == "" {
		defaultName = DefaultModelName
	}
	modelName = strings.TrimPrefix(modelName, "models/")

	var thinkOverride *int
	if idx := strings.LastIndex(modelName, "@think="); idx != -1 {
		thinkStr := modelName[idx+len("@think="):]
		modelName = modelName[:idx]
		val, err := strconv.Atoi(thinkStr)
		if err != nil {
			return Resolved{}, fmt.Errorf("Invalid think level: %s", thinkStr)
		}
		thinkOverride = &val
	}

	cfg, exists := MODELS[modelName]
	if !exists {
		log.Printf("Unknown model '%s', falling back to '%s'", modelName, defaultName)
		modelName = defaultName
		cfg = MODELS[defaultName]
	}

	thinkMode := cfg.Think
	if thinkOverride != nil {
		thinkMode = *thinkOverride
	}

	return Resolved{
		Name:  modelName,
		Mode:  cfg.Mode,
		Think: thinkMode,
		Extra: cfg.Extra,
	}, nil
}
