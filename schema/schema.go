package schema

import (
	"encoding/json"
	"fmt"
	"fsagen/spec"
	"reflect"
	"strings"
)

func BuildManifestSchema() ([]byte, error) {
	defs := map[string]any{
		"Rfc3339Time": map[string]any{
			"type":        "string",
			"description": "Timestamp in RFC3339 format",
			"format":      "date-time",
		},
		"OperationAction": map[string]any{
			"type": "string",
			"enum": []string{"create", "update", "append", "delete", "mace", "rename", "truncate", "rotate", "ads", "motw"},
		},
	}

	operation := objectSchemaFromStruct(spec.Operation{}, nil)
	opProps := operation["properties"].(map[string]any)
	opProps["action"] = map[string]any{"$ref": "#/$defs/OperationAction"}
	opProps["type"] = map[string]any{"type": "string", "enum": []string{"file", "dir"}}
	opProps["atime"] = map[string]any{"$ref": "#/$defs/Rfc3339Time"}
	opProps["mtime"] = map[string]any{"$ref": "#/$defs/Rfc3339Time"}
	opProps["zone_id"] = map[string]any{"type": "integer", "minimum": 0, "maximum": 4}
	operation["required"] = []string{"action", "path"}
	operation["allOf"] = []any{
		conditionalRequire("action", "rename", "new_path"),
		conditionalRequire("action", "rotate", "new_path"),
		conditionalRequire("action", "ads", "stream"),
		map[string]any{
			"if": map[string]any{
				"properties": map[string]any{"action": map[string]any{"const": "motw"}},
				"required":   []string{"action"},
			},
			"then": map[string]any{
				"properties": map[string]any{
					"zone_id": map[string]any{"type": "integer", "minimum": 0, "maximum": 4},
				},
			},
		},
	}
	defs["Operation"] = operation

	root := map[string]any{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"$id":         "https://generator.local/manifest-schema.json",
		"title":       "FSAGen Manifest Schema",
		"description": "Validates fsagen manifest input files.",
		"type":        "object",
		"properties": map[string]any{
			"operations": map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": "#/$defs/Operation"},
			},
		},
		"required":             []string{"operations"},
		"additionalProperties": false,
		"$defs":                defs,
	}

	return marshalSchema(root)
}

func BuildPlaybookSchema() ([]byte, error) {
	defs := map[string]any{
		"Rfc3339Time": map[string]any{
			"type":        "string",
			"description": "Timestamp in RFC3339 format",
			"format":      "date-time",
		},
		"Condition": map[string]any{
			"type": "string",
			"enum": []string{"", "odd", "even", "first", "last"},
		},
		"OperationAction": map[string]any{
			"type": "string",
			"enum": []string{"create", "update", "append", "delete", "mace", "rename", "truncate", "rotate", "ads", "motw"},
		},
		"Template": map[string]any{
			"type": "string",
			"enum": []string{"email", "log", "script", "doc"},
		},
	}

	actor := objectSchemaFromStruct(spec.Actor{}, []string{"name"})
	defs["Actor"] = actor

	action := objectSchemaFromStruct(spec.Action{}, []string{"action", "path"})
	actionProps := action["properties"].(map[string]any)
	actionProps["action"] = map[string]any{"$ref": "#/$defs/OperationAction"}
	actionProps["type"] = map[string]any{"type": "string", "enum": []string{"file", "dir"}}
	actionProps["template"] = map[string]any{"$ref": "#/$defs/Template"}
	actionProps["condition"] = map[string]any{"$ref": "#/$defs/Condition"}
	actionProps["offset"] = map[string]any{"type": "string", "description": "Go duration string (for example: 15m, 2h, 30s)"}
	actionProps["atime"] = map[string]any{"$ref": "#/$defs/Rfc3339Time"}
	actionProps["mtime"] = map[string]any{"$ref": "#/$defs/Rfc3339Time"}
	actionProps["zone_id"] = map[string]any{"type": "integer", "minimum": 0, "maximum": 4}
	action["allOf"] = []any{
		conditionalRequire("action", "rename", "new_path"),
		conditionalRequire("action", "rotate", "new_path"),
		conditionalRequire("action", "ads", "stream"),
	}
	defs["Action"] = action

	step := objectSchemaFromStruct(spec.Step{}, []string{"actor", "actions"})
	stepProps := step["properties"].(map[string]any)
	stepProps["offset"] = map[string]any{"type": "string", "description": "Go duration string"}
	stepProps["every"] = map[string]any{"type": "string", "description": "Go duration string"}
	stepProps["condition"] = map[string]any{"$ref": "#/$defs/Condition"}
	stepProps["repeat"] = map[string]any{"type": "integer", "minimum": 1}
	stepProps["batch_count"] = map[string]any{"type": "integer", "minimum": 1}
	defs["Step"] = step

	root := map[string]any{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"$id":         "https://generator.local/playbook-schema.json",
		"title":       "FSAGen Playbook Schema",
		"description": "Validates fsagen playbook input files.",
		"type":        "object",
		"properties": map[string]any{
			"start": map[string]any{
				"oneOf": []any{
					map[string]any{"const": "now"},
					map[string]any{"$ref": "#/$defs/Rfc3339Time"},
				},
			},
			"variables": map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"type": "string"},
			},
			"actors": map[string]any{
				"type":     "array",
				"items":    map[string]any{"$ref": "#/$defs/Actor"},
				"minItems": 1,
			},
			"steps": map[string]any{
				"type":     "array",
				"items":    map[string]any{"$ref": "#/$defs/Step"},
				"minItems": 1,
			},
		},
		"required":             []string{"actors", "steps"},
		"additionalProperties": false,
		"$defs":                defs,
	}

	return marshalSchema(root)
}

func conditionalRequire(fieldName string, fieldValue string, requiredField string) map[string]any {
	return map[string]any{
		"if": map[string]any{
			"properties": map[string]any{fieldName: map[string]any{"const": fieldValue}},
			"required":   []string{fieldName},
		},
		"then": map[string]any{
			"required": []string{requiredField},
		},
	}
}

func objectSchemaFromStruct(v any, required []string) map[string]any {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	properties := map[string]any{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		jsonName, ok := jsonFieldName(f)
		if !ok {
			continue
		}
		properties[jsonName] = typeToSchema(f.Type)
	}
	out := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		out["required"] = required
	}
	return out
}

func typeToSchema(t reflect.Type) map[string]any {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice, reflect.Array:
		elem := t.Elem()
		for elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Struct {
			return map[string]any{
				"type":  "array",
				"items": map[string]any{"$ref": fmt.Sprintf("#/$defs/%s", elem.Name())},
			}
		}
		return map[string]any{
			"type":  "array",
			"items": typeToSchema(elem),
		}
	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			return map[string]any{
				"type":                 "object",
				"additionalProperties": typeToSchema(t.Elem()),
			}
		}
		return map[string]any{"type": "object"}
	case reflect.Struct:
		return map[string]any{"$ref": fmt.Sprintf("#/$defs/%s", t.Name())}
	default:
		return map[string]any{}
	}
}

func jsonFieldName(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("json")
	if tag == "" {
		return "", false
	}
	parts := strings.Split(tag, ",")
	name := strings.TrimSpace(parts[0])
	if name == "" || name == "-" {
		return "", false
	}
	return name, true
}

func marshalSchema(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
