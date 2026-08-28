package models

import "sync"

// PricingInfo contains commercial cloud benchmark pricing per 1M tokens as of 28 August 2026.
type PricingInfo struct {
	InputPer1M   float64 `json:"input_per_1m"`
	OutputPer1M  float64 `json:"output_per_1m"`
	BlendedPer1M float64 `json:"blended_per_1m"`
	Category     string  `json:"category"`
	Provider     string  `json:"provider"`
}

// MODEL_PRICING stores real-world cloud benchmark rates for all models (August 2026).
var MODEL_PRICING = map[string]PricingInfo{
	// Anthropic Claude 5, 4.5 & 3.7 Tier
	"claude-5-opus":              {InputPer1M: 15.00, OutputPer1M: 75.00, BlendedPer1M: 45.00, Category: "Frontier Reasoning", Provider: "Anthropic"},
	"claude-5-fable":             {InputPer1M: 15.00, OutputPer1M: 75.00, BlendedPer1M: 45.00, Category: "Frontier Intelligence", Provider: "Anthropic"},
	"claude-5-sonnet":            {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Balanced Flagship", Provider: "Anthropic"},
	"claude-sonnet-5":            {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Balanced Flagship", Provider: "Anthropic"},
	"claude-opus-5":              {InputPer1M: 15.00, OutputPer1M: 75.00, BlendedPer1M: 45.00, Category: "Frontier Reasoning", Provider: "Anthropic"},
	"claude-fable-5":             {InputPer1M: 15.00, OutputPer1M: 75.00, BlendedPer1M: 45.00, Category: "Frontier Intelligence", Provider: "Anthropic"},
	"claude-4-5-opus":            {InputPer1M: 15.00, OutputPer1M: 75.00, BlendedPer1M: 45.00, Category: "Enterprise Agentic", Provider: "Anthropic"},
	"claude-4-5-sonnet":          {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Balanced Flagship", Provider: "Anthropic"},
	"claude-4-5-haiku":           {InputPer1M: 0.80,  OutputPer1M: 4.00,  BlendedPer1M: 2.40,  Category: "High Throughput", Provider: "Anthropic"},
	"claude-haiku-4.5":           {InputPer1M: 0.80,  OutputPer1M: 4.00,  BlendedPer1M: 2.40,  Category: "High Throughput", Provider: "Anthropic"},
	"claude-3-7-sonnet":          {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Hybrid Reasoning", Provider: "Anthropic"},
	"claude-3-7-sonnet-latest":   {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Hybrid Reasoning", Provider: "Anthropic"},
	"claude-3-7-sonnet-20250219": {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Hybrid Reasoning", Provider: "Anthropic"},
	"claude-3-5-sonnet":          {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Flagship Coding", Provider: "Anthropic"},
	"claude-3-5-sonnet-latest":   {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Flagship Coding", Provider: "Anthropic"},
	"claude-3-5-sonnet-20241022": {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Flagship Coding v2", Provider: "Anthropic"},
	"claude-3-5-sonnet-20240620": {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Flagship Coding v1", Provider: "Anthropic"},
	"claude-3-5-haiku":           {InputPer1M: 0.80,  OutputPer1M: 4.00,  BlendedPer1M: 2.40,  Category: "High Throughput", Provider: "Anthropic"},
	"claude-3-5-haiku-20241022":  {InputPer1M: 0.80,  OutputPer1M: 4.00,  BlendedPer1M: 2.40,  Category: "High Throughput", Provider: "Anthropic"},
	"claude-3-haiku-20240307":    {InputPer1M: 0.25,  OutputPer1M: 1.25,  BlendedPer1M: 0.75,  Category: "Legacy Fast", Provider: "Anthropic"},
	"claude-3-opus":              {InputPer1M: 15.00, OutputPer1M: 75.00, BlendedPer1M: 45.00, Category: "Complex Analysis", Provider: "Anthropic"},
	"claude-3-opus-20240229":     {InputPer1M: 15.00, OutputPer1M: 75.00, BlendedPer1M: 45.00, Category: "Complex Analysis", Provider: "Anthropic"},
	"claude-code":                {InputPer1M: 3.00,  OutputPer1M: 15.00, BlendedPer1M: 9.00,  Category: "Autonomous Coding", Provider: "Anthropic"},

	// OpenAI GPT-5.6, GPT-5.5, GPT-4.5 & Reasoning Tier
	"gpt-5.6":           {InputPer1M: 5.00,  OutputPer1M: 20.00, BlendedPer1M: 12.50, Category: "Frontier Professional", Provider: "OpenAI"},
	"gpt-5.6-sol":       {InputPer1M: 5.00,  OutputPer1M: 20.00, BlendedPer1M: 12.50, Category: "Frontier Professional", Provider: "OpenAI"},
	"gpt-5.6-terra":     {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "Balanced General", Provider: "OpenAI"},
	"gpt-5.6-luna":      {InputPer1M: 0.50,  OutputPer1M: 2.00,  BlendedPer1M: 1.25,  Category: "High Throughput", Provider: "OpenAI"},
	"gpt-5.6-cyber":     {InputPer1M: 5.00,  OutputPer1M: 20.00, BlendedPer1M: 12.50, Category: "Security & Deep Code", Provider: "OpenAI"},
	"gpt-5.5":           {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "Enterprise Workhorse", Provider: "OpenAI"},
	"gpt-5.5-pro":       {InputPer1M: 5.00,  OutputPer1M: 20.00, BlendedPer1M: 12.50, Category: "Enterprise Flagship", Provider: "OpenAI"},
	"gpt-5.4":           {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "General Purpose", Provider: "OpenAI"},
	"gpt-5.4-mini":      {InputPer1M: 0.15,  OutputPer1M: 0.60,  BlendedPer1M: 0.375, Category: "Lightweight Mini", Provider: "OpenAI"},
	"gpt-5.4-pro":       {InputPer1M: 5.00,  OutputPer1M: 20.00, BlendedPer1M: 12.50, Category: "Enterprise Pro", Provider: "OpenAI"},
	"gpt-5-codex":       {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "Codex Agentic", Provider: "OpenAI"},
	"gpt-5.1-codex":     {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "Codex Agentic", Provider: "OpenAI"},
	"gpt-5.2-codex":     {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "Codex Agentic", Provider: "OpenAI"},
	"gpt-5.3-codex":     {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "Codex Agentic", Provider: "OpenAI"},
	"gpt-4.5":           {InputPer1M: 75.00, OutputPer1M: 150.00, BlendedPer1M: 112.50, Category: "Frontier Knowledge", Provider: "OpenAI"},
	"gpt-4.5-preview":   {InputPer1M: 75.00, OutputPer1M: 150.00, BlendedPer1M: 112.50, Category: "Frontier Knowledge", Provider: "OpenAI"},
	"gpt-4o":            {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "Multimodal Fast", Provider: "OpenAI"},
	"gpt-4o-2024-11-20": {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "Multimodal Fast", Provider: "OpenAI"},
	"gpt-4o-2024-08-06": {InputPer1M: 2.50,  OutputPer1M: 10.00, BlendedPer1M: 6.25,  Category: "Structured Outputs", Provider: "OpenAI"},
	"gpt-4o-mini":       {InputPer1M: 0.15,  OutputPer1M: 0.60,  BlendedPer1M: 0.375, Category: "Cost Efficient", Provider: "OpenAI"},
	"gpt-4o-mini-2024-07-18": {InputPer1M: 0.15,  OutputPer1M: 0.60,  BlendedPer1M: 0.375, Category: "Cost Efficient", Provider: "OpenAI"},
	"chatgpt-4o-latest": {InputPer1M: 5.00,  OutputPer1M: 15.00, BlendedPer1M: 10.00, Category: "Dynamic Flagship", Provider: "OpenAI"},
	"gpt-4-turbo":       {InputPer1M: 10.00, OutputPer1M: 30.00, BlendedPer1M: 20.00, Category: "Legacy Turbo", Provider: "OpenAI"},
	"gpt-4":             {InputPer1M: 30.00, OutputPer1M: 60.00, BlendedPer1M: 45.00, Category: "Legacy Frontier", Provider: "OpenAI"},
	"gpt-3.5-turbo":     {InputPer1M: 0.50,  OutputPer1M: 1.50,  BlendedPer1M: 1.00,  Category: "Legacy Fast", Provider: "OpenAI"},
	"o3":                {InputPer1M: 15.00, OutputPer1M: 60.00, BlendedPer1M: 37.50, Category: "Deep Reasoning", Provider: "OpenAI"},
	"o3-mini":           {InputPer1M: 1.10,  OutputPer1M: 4.40,  BlendedPer1M: 2.75,  Category: "STEM Reasoning", Provider: "OpenAI"},
	"o3-mini-2025-01-31": {InputPer1M: 1.10,  OutputPer1M: 4.40,  BlendedPer1M: 2.75,  Category: "STEM Reasoning", Provider: "OpenAI"},
	"o3-pro":            {InputPer1M: 20.00, OutputPer1M: 80.00, BlendedPer1M: 50.00, Category: "Frontier Reasoning Pro", Provider: "OpenAI"},
	"o4-mini":           {InputPer1M: 1.10,  OutputPer1M: 4.40,  BlendedPer1M: 2.75,  Category: "Next-Gen Reasoning", Provider: "OpenAI"},
	"o1":                {InputPer1M: 15.00, OutputPer1M: 60.00, BlendedPer1M: 37.50, Category: "Deep Reasoning", Provider: "OpenAI"},
	"o1-2024-12-17":     {InputPer1M: 15.00, OutputPer1M: 60.00, BlendedPer1M: 37.50, Category: "Deep Reasoning", Provider: "OpenAI"},
	"o1-preview":        {InputPer1M: 15.00, OutputPer1M: 60.00, BlendedPer1M: 37.50, Category: "Reasoning Preview", Provider: "OpenAI"},
	"o1-mini":           {InputPer1M: 1.10,  OutputPer1M: 4.40,  BlendedPer1M: 2.75,  Category: "Fast Reasoning", Provider: "OpenAI"},
	"o1-pro":            {InputPer1M: 20.00, OutputPer1M: 80.00, BlendedPer1M: 50.00, Category: "Frontier Reasoning Pro", Provider: "OpenAI"},

	// Google Gemini 3.7, 3.5 & 3.1 Tier
	"gemini-3.7-flash":          {InputPer1M: 0.15,  OutputPer1M: 0.60,  BlendedPer1M: 0.375, Category: "Flagship Fast", Provider: "Google"},
	"gemini-3.7-flash-thinking": {InputPer1M: 0.15,  OutputPer1M: 0.60,  BlendedPer1M: 0.375, Category: "Deep Thinking", Provider: "Google"},
	"gemini-3.6-flash":          {InputPer1M: 0.15,  OutputPer1M: 0.60,  BlendedPer1M: 0.375, Category: "All-Around Workhorse", Provider: "Google"},
	"gemini-3.5-flash":          {InputPer1M: 0.15,  OutputPer1M: 0.60,  BlendedPer1M: 0.375, Category: "All-Around Workhorse", Provider: "Google"},
	"gemini-3.5-flash-lite":     {InputPer1M: 0.075, OutputPer1M: 0.30,  BlendedPer1M: 0.1875, Category: "Ultra-Lightweight", Provider: "Google"},
	"gemini-3.1-pro":            {InputPer1M: 1.25,  OutputPer1M: 5.00,  BlendedPer1M: 3.125, Category: "Pro Reasoning", Provider: "Google"},
	"gemini-3.1-pro-preview":    {InputPer1M: 1.25,  OutputPer1M: 5.00,  BlendedPer1M: 3.125, Category: "Pro Reasoning Preview", Provider: "Google"},
	"gemini-flash-lite":         {InputPer1M: 0.075, OutputPer1M: 0.30,  BlendedPer1M: 0.1875, Category: "Sub-Second Lite", Provider: "Google"},
	"gemini-omni-1.1-flash":     {InputPer1M: 0.20,  OutputPer1M: 0.80,  BlendedPer1M: 0.50,  Category: "Conversational Multimodal", Provider: "Google"},
	"gemini-2.0-flash":          {InputPer1M: 0.10,  OutputPer1M: 0.40,  BlendedPer1M: 0.25,  Category: "Legacy Fast", Provider: "Google"},
	"gemini-1.5-pro":            {InputPer1M: 1.25,  OutputPer1M: 5.00,  BlendedPer1M: 3.125, Category: "Legacy Pro", Provider: "Google"},
	"gemini-1.5-flash":          {InputPer1M: 0.075, OutputPer1M: 0.30,  BlendedPer1M: 0.1875, Category: "Legacy Lite", Provider: "Google"},

	// DeepSeek & Open Weights
	"deepseek-r1":       {InputPer1M: 0.55,  OutputPer1M: 2.19,  BlendedPer1M: 1.37,  Category: "Open Reasoning", Provider: "DeepSeek"},
	"deepseek-reasoner": {InputPer1M: 0.55,  OutputPer1M: 2.19,  BlendedPer1M: 1.37,  Category: "Open Reasoning", Provider: "DeepSeek"},
	"deepseek-v3":       {InputPer1M: 0.14,  OutputPer1M: 0.28,  BlendedPer1M: 0.21,  Category: "Open General", Provider: "DeepSeek"},
	"deepseek-chat":     {InputPer1M: 0.14,  OutputPer1M: 0.28,  BlendedPer1M: 0.21,  Category: "Open General", Provider: "DeepSeek"},
	"llama-3.1-405b":    {InputPer1M: 2.00,  OutputPer1M: 5.00,  BlendedPer1M: 3.50,  Category: "Open Weights Giant", Provider: "Meta"},
	"llama-3.3-70b":     {InputPer1M: 0.35,  OutputPer1M: 0.90,  BlendedPer1M: 0.625, Category: "Open Weights Fast", Provider: "Meta"},
	"qwen-2.5-coder-32b": {InputPer1M: 0.30,  OutputPer1M: 0.80,  BlendedPer1M: 0.55,  Category: "Coding Specialist", Provider: "Alibaba"},
}

var pricingMu sync.RWMutex

// RegisterPricing dynamically registers or overrides a model pricing definition at runtime.
func RegisterPricing(name string, p PricingInfo) {
	pricingMu.Lock()
	defer pricingMu.Unlock()
	MODEL_PRICING[name] = p
}

// GetAllPricing returns a thread-safe copy of all registered pricing information.
func GetAllPricing() map[string]PricingInfo {
	pricingMu.RLock()
	defer pricingMu.RUnlock()
	res := make(map[string]PricingInfo, len(MODEL_PRICING))
	for k, v := range MODEL_PRICING {
		res[k] = v
	}
	return res
}

// GetModelPricing returns the pricing info for a given model or a reasonable frontier fallback.
func GetModelPricing(modelName string) PricingInfo {
	pricingMu.RLock()
	defer pricingMu.RUnlock()
	if p, ok := MODEL_PRICING[modelName]; ok {
		return p
	}
	return PricingInfo{
		InputPer1M:   1.50,
		OutputPer1M:  6.00,
		BlendedPer1M: 3.75,
		Category:     "Commercial Cloud Frontier",
		Provider:     "Cloud Benchmark",
	}
}

// CalculateSavingsUSD calculates the exact commercial dollar savings for a specific model and token count.
func CalculateSavingsUSD(modelName string, tokens int64) float64 {
	p := GetModelPricing(modelName)
	return float64(tokens) / 1_000_000.0 * p.BlendedPer1M
}
