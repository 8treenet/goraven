package tools

import (
	"encoding/json"
	"testing"

	"github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestNormalizeJSONSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    *jsonschema.Schema
		expected func(*jsonschema.Schema) bool
	}{
		{
			name: "fix bool type",
			input: &jsonschema.Schema{
				Type: "bool",
				Properties: orderedMapFrom(map[string]*jsonschema.Schema{
					"enabled": {Type: "bool"},
				}),
			},
			expected: func(s *jsonschema.Schema) bool {
				if s.Type != "boolean" {
					return false
				}
				v, _ := s.Properties.Get("enabled")
				return v.Type == "boolean"
			},
		},
		{
			name: "fix int type",
			input: &jsonschema.Schema{
				Type: "int",
				Properties: orderedMapFrom(map[string]*jsonschema.Schema{
					"count": {Type: "int32"},
				}),
			},
			expected: func(s *jsonschema.Schema) bool {
				if s.Type != "integer" {
					return false
				}
				v, _ := s.Properties.Get("count")
				return v.Type == "integer"
			},
		},
		{
			name: "fix float type",
			input: &jsonschema.Schema{
				Type: "float",
			},
			expected: func(s *jsonschema.Schema) bool {
				return s.Type == "number"
			},
		},
		{
			name: "fix types in anyOf",
			input: &jsonschema.Schema{
				AnyOf: []*jsonschema.Schema{
					{Type: "bool"},
					{Type: "string"},
				},
			},
			expected: func(s *jsonschema.Schema) bool {
				return s.AnyOf[0].Type == "boolean" && s.AnyOf[1].Type == "string"
			},
		},
		{
			name: "fix TypeEnhanced",
			input: &jsonschema.Schema{
				TypeEnhanced: []string{"bool", "string", "int"},
			},
			expected: func(s *jsonschema.Schema) bool {
				return s.TypeEnhanced[0] == "boolean" &&
					s.TypeEnhanced[1] == "string" &&
					s.TypeEnhanced[2] == "integer"
			},
		},
		{
			name: "valid types unchanged",
			input: &jsonschema.Schema{
				Type: "boolean",
				Properties: orderedMapFrom(map[string]*jsonschema.Schema{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
				}),
			},
			expected: func(s *jsonschema.Schema) bool {
				if s.Type != "boolean" {
					return false
				}
				v, _ := s.Properties.Get("name")
				if v.Type != "string" {
					return false
				}
				v2, _ := s.Properties.Get("age")
				return v2.Type == "integer"
			},
		},
		{
			name:  "nil schema",
			input: nil,
			expected: func(s *jsonschema.Schema) bool {
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeJSONSchema(tt.input)
			if tt.input != nil && !tt.expected(tt.input) {
				data, _ := json.MarshalIndent(tt.input, "", "  ")
				t.Errorf("normalization failed\nschema: %s", string(data))
			}
		})
	}
}

func orderedMapFrom(m map[string]*jsonschema.Schema) *orderedmap.OrderedMap[string, *jsonschema.Schema] {
	om := jsonschema.NewProperties()
	for k, v := range m {
		om.Set(k, v)
	}
	return om
}
