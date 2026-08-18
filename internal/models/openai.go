package models

type OpenAIChatRequest struct {
	Model         string               `json:"model"`
	Messages      []OpenAIMessage      `json:"messages"`
	Tools         []OpenAITool         `json:"tools,omitempty"`
	ToolChoice    any                  `json:"tool_choice,omitempty"`
	Stream        bool                 `json:"stream,omitempty"`
	StreamOptions *OpenAIStreamOptions `json:"stream_options,omitempty"`
	Temperature   *float64             `json:"temperature,omitempty"`
	TopP          *float64             `json:"top_p,omitempty"`
	MaxTokens     *int                 `json:"max_tokens,omitempty"`
	User          string               `json:"user,omitempty"`
}

type OpenAIStreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type OpenAIMessage struct {
	Role             string           `json:"role,omitempty"`
	Content          any              `json:"content"`
	ReasoningContent string           `json:"reasoning_content,omitempty"`
	Name             string           `json:"name,omitempty"`
	ToolCallID       string           `json:"tool_call_id,omitempty"`
	ToolCalls        []OpenAIToolCall `json:"tool_calls,omitempty"`
}

type OpenAITool struct {
	Type     string         `json:"type"`
	Function OpenAIFunction `json:"function"`

	// Flat tool declarations fallback
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type OpenAIFunction struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type OpenAIToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function OpenAIToolCallFunction `json:"function"`
}

type OpenAIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAIChatResponse struct {
	ID                string         `json:"id"`
	Object            string         `json:"object"`
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	SystemFingerprint string         `json:"system_fingerprint,omitempty"`
	Choices           []OpenAIChoice `json:"choices"`
	Usage             *OpenAIUsage   `json:"usage,omitempty"`
}

type OpenAIChoice struct {
	Index        int            `json:"index"`
	Message      *OpenAIMessage `json:"message,omitempty"`
	Delta        *OpenAIMessage `json:"delta,omitempty"`
	FinishReason *string        `json:"finish_reason"`
}

type OpenAIUsage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	PromptTokensDetails     *PromptTokensDetails     `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
}

type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}
