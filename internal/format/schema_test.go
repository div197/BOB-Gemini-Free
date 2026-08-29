package format

import (
	"strings"
	"testing"
)

func TestValidateToolSchemaBoundsStructure(t *testing.T) {
	properties := make(map[string]any, MaxToolSchemaProperties+1)
	for i := 0; i <= MaxToolSchemaProperties; i++ {
		properties["field"+string(rune('a'+i%26))+string(rune('0'+i/26))] = map[string]any{"type": "string"}
	}
	if err := ValidateToolSchema(map[string]any{"type": "object", "properties": properties}); err == nil || !strings.Contains(err.Error(), "properties") {
		t.Fatalf("expected property budget error, got %v", err)
	}

	enum := make([]any, MaxToolSchemaEnumValues+1)
	for i := range enum {
		enum[i] = i
	}
	if err := ValidateToolSchema(map[string]any{"type": "string", "enum": enum}); err == nil || !strings.Contains(err.Error(), "enum") {
		t.Fatalf("expected enum budget error, got %v", err)
	}

	deep := map[string]any{}
	cursor := deep
	for i := 0; i <= MaxToolSchemaDepth; i++ {
		next := map[string]any{}
		cursor["properties"] = map[string]any{"nested": next}
		cursor = next
	}
	if err := ValidateToolSchema(deep); err == nil || !strings.Contains(err.Error(), "depth") {
		t.Fatalf("expected depth budget error, got %v", err)
	}

	typedEnum := map[string]any{"enum": []string{"one", "two"}}
	if err := ValidateToolSchema(typedEnum); err != nil {
		t.Fatalf("typed enum was rejected unexpectedly: %v", err)
	}
	largeTypedEnum := map[string]any{"enum": make([]string, MaxToolSchemaEnumValues+1)}
	if err := ValidateToolSchema(largeTypedEnum); err == nil || !strings.Contains(err.Error(), "enum") {
		t.Fatal("typed enum budget was bypassed")
	}

	cyclic := map[string]any{}
	cyclic["self"] = cyclic
	if err := ValidateToolSchema(cyclic); err == nil || !strings.Contains(err.Error(), "depth") {
		t.Fatal("cyclic schema was accepted without a depth guard")
	}
}
