package format

import (
	"fmt"
	"reflect"
)

const (
	MaxToolSchemaDepth      = 16
	MaxToolSchemaNodes      = 2048
	MaxToolSchemaProperties = 256
	MaxToolSchemaEnumValues = 256
)

// ValidateToolSchema bounds the structural complexity of a user-supplied
// JSON schema before it is copied into a prompt or sent to a provider. The
// byte budget is enforced by the caller after JSON encoding; this validator
// protects against many small nested values that would otherwise fit below
// that byte limit while still causing disproportionate traversal work.
func ValidateToolSchema(schema any) error {
	if schema == nil {
		return nil
	}

	nodes := 0
	properties := 0
	enumValues := 0
	var walk func(value reflect.Value, depth int) error
	walk = func(value reflect.Value, depth int) error {
		if !value.IsValid() {
			return nil
		}
		for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return nil
			}
			value = value.Elem()
		}
		if depth > MaxToolSchemaDepth {
			return fmt.Errorf("tool schema depth exceeds %d", MaxToolSchemaDepth)
		}

		switch value.Kind() {
		case reflect.Map:
			nodes++
			if nodes > MaxToolSchemaNodes {
				return fmt.Errorf("tool schema exceeds %d nodes", MaxToolSchemaNodes)
			}
			iter := value.MapRange()
			for iter.Next() {
				keyName := ""
				if iter.Key().Kind() == reflect.String {
					keyName = iter.Key().String()
				}
				child := iter.Value()
				childValue := child
				for childValue.IsValid() && (childValue.Kind() == reflect.Interface || childValue.Kind() == reflect.Pointer) && !childValue.IsNil() {
					childValue = childValue.Elem()
				}
				if keyName == "properties" && childValue.IsValid() && (childValue.Kind() == reflect.Map || childValue.Kind() == reflect.Slice || childValue.Kind() == reflect.Array) {
					properties += childValue.Len()
					if properties > MaxToolSchemaProperties {
						return fmt.Errorf("tool schema exceeds %d properties", MaxToolSchemaProperties)
					}
				}
				if keyName == "enum" && childValue.IsValid() && (childValue.Kind() == reflect.Slice || childValue.Kind() == reflect.Array) {
					enumValues += childValue.Len()
					if enumValues > MaxToolSchemaEnumValues {
						return fmt.Errorf("tool schema exceeds %d enum values", MaxToolSchemaEnumValues)
					}
				}
				if err := walk(child, depth+1); err != nil {
					return err
				}
			}
		case reflect.Slice, reflect.Array:
			nodes++
			if nodes > MaxToolSchemaNodes {
				return fmt.Errorf("tool schema exceeds %d nodes", MaxToolSchemaNodes)
			}
			for index := 0; index < value.Len(); index++ {
				if err := walk(value.Index(index), depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return walk(reflect.ValueOf(schema), 0)
}
