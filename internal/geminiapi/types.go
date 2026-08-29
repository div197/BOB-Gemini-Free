// Package geminiapi contains the explicit, opt-in Google Gemini Developer API
// transport. It is deliberately separate from internal/gemini, which speaks
// Google's undocumented web-RPC protocol using browser session cookies.
package geminiapi

// GenerateContentRequest is the stable subset of the public Gemini API request
// schema that BOB can translate from its adapter protocols.
type GenerateContentRequest struct {
	Contents          []Content         `json:"contents,omitempty"`
	SystemInstruction *Content          `json:"systemInstruction,omitempty"`
	Tools             []Tool            `json:"tools,omitempty"`
	ToolConfig        *ToolConfig       `json:"toolConfig,omitempty"`
	GenerationConfig  *GenerationConfig `json:"generationConfig,omitempty"`
}

type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text             string            `json:"text,omitempty"`
	InlineData       *InlineData       `json:"inlineData,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	Thought          bool              `json:"thought,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type FunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

type FunctionResponse struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations,omitempty"`
}

type FunctionDeclaration struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type ToolConfig struct {
	FunctionCallingConfig FunctionCallingConfig `json:"functionCallingConfig"`
}

type FunctionCallingConfig struct {
	Mode                 string   `json:"mode"`
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}

type GenerationConfig struct {
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"topP,omitempty"`
	MaxOutputTokens  *int     `json:"maxOutputTokens,omitempty"`
	ResponseMimeType string   `json:"responseMimeType,omitempty"`
}

type GenerateContentResponse struct {
	Candidates     []Candidate     `json:"candidates,omitempty"`
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *UsageMetadata  `json:"usageMetadata,omitempty"`
	ModelVersion   string          `json:"modelVersion,omitempty"`
	ResponseID     string          `json:"responseId,omitempty"`
}

type Candidate struct {
	Index         int      `json:"index,omitempty"`
	Content       *Content `json:"content,omitempty"`
	FinishReason  string   `json:"finishReason,omitempty"`
	FinishMessage string   `json:"finishMessage,omitempty"`
}

type PromptFeedback struct {
	BlockReason        string `json:"blockReason,omitempty"`
	BlockReasonMessage string `json:"blockReasonMessage,omitempty"`
}

type UsageMetadata struct {
	PromptTokenCount        int `json:"promptTokenCount,omitempty"`
	CachedContentTokenCount int `json:"cachedContentTokenCount,omitempty"`
	CandidatesTokenCount    int `json:"candidatesTokenCount,omitempty"`
	ThoughtsTokenCount      int `json:"thoughtsTokenCount,omitempty"`
	TotalTokenCount         int `json:"totalTokenCount,omitempty"`
}
