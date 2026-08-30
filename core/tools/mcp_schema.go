package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

// validJSONSchemaTypes defines the set of valid JSON Schema type strings.
var validJSONSchemaTypes = map[string]bool{
	"null":    true,
	"boolean": true,
	"object":  true,
	"array":   true,
	"number":  true,
	"string":  true,
	"integer": true,
}

// typeNormalizationMap maps common invalid type strings to their valid equivalents.
var typeNormalizationMap = map[string]string{
	"bool":    "boolean",
	"int":     "integer",
	"int32":   "integer",
	"int64":   "integer",
	"uint":    "integer",
	"uint32":  "integer",
	"uint64":  "integer",
	"float":   "number",
	"float32": "number",
	"float64": "number",
	"double":  "number",
	"str":     "string",
}

// hasInvalidTypes recursively checks whether a schema contains any non-standard
// type names (e.g. "bool", "int", "float") that need normalization.
func hasInvalidTypes(s *jsonschema.Schema) bool {
	if s == nil {
		return false
	}

	if s.Type != "" && !validJSONSchemaTypes[s.Type] {
		if _, ok := typeNormalizationMap[s.Type]; ok {
			return true
		}
	}
	for _, t := range s.TypeEnhanced {
		if !validJSONSchemaTypes[t] {
			if _, ok := typeNormalizationMap[t]; ok {
				return true
			}
		}
	}

	// Check properties
	if s.Properties != nil {
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			if hasInvalidTypes(pair.Value) {
				return true
			}
		}
	}
	for _, prop := range s.PatternProperties {
		if hasInvalidTypes(prop) {
			return true
		}
	}

	if hasInvalidTypes(s.Items) {
		return true
	}
	for _, item := range s.PrefixItems {
		if hasInvalidTypes(item) {
			return true
		}
	}
	if hasInvalidTypes(s.AdditionalProperties) {
		return true
	}

	for _, sub := range s.AllOf {
		if hasInvalidTypes(sub) {
			return true
		}
	}
	for _, sub := range s.AnyOf {
		if hasInvalidTypes(sub) {
			return true
		}
	}
	for _, sub := range s.OneOf {
		if hasInvalidTypes(sub) {
			return true
		}
	}
	if hasInvalidTypes(s.Not) {
		return true
	}
	if hasInvalidTypes(s.If) || hasInvalidTypes(s.Then) || hasInvalidTypes(s.Else) {
		return true
	}
	for _, dep := range s.DependentSchemas {
		if hasInvalidTypes(dep) {
			return true
		}
	}
	for _, def := range s.Definitions {
		if hasInvalidTypes(def) {
			return true
		}
	}
	if hasInvalidTypes(s.Contains) || hasInvalidTypes(s.ContentSchema) || hasInvalidTypes(s.PropertyNames) {
		return true
	}

	return false
}

// normalizeJSONSchema recursively walks a JSON Schema and fixes invalid type names
// (e.g., "bool" -> "boolean", "int" -> "integer", "float" -> "number").
// Some MCP servers use non-standard type names that cause LLM providers to reject the schema.
func normalizeJSONSchema(s *jsonschema.Schema) {
	if s == nil {
		return
	}

	// Fix single type
	if s.Type != "" && !validJSONSchemaTypes[s.Type] {
		if fixed, ok := typeNormalizationMap[s.Type]; ok {
			s.Type = fixed
		}
	}

	// Fix multi-type (TypeEnhanced)
	for i, t := range s.TypeEnhanced {
		if !validJSONSchemaTypes[t] {
			if fixed, ok := typeNormalizationMap[t]; ok {
				s.TypeEnhanced[i] = fixed
			}
		}
	}

	// Recurse into properties (OrderedMap)
	if s.Properties != nil {
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			normalizeJSONSchema(pair.Value)
		}
	}

	// Recurse into pattern properties
	for _, prop := range s.PatternProperties {
		normalizeJSONSchema(prop)
	}

	// Recurse into items
	normalizeJSONSchema(s.Items)
	for _, item := range s.PrefixItems {
		normalizeJSONSchema(item)
	}

	// Recurse into additional properties
	normalizeJSONSchema(s.AdditionalProperties)

	// Recurse into combinators
	for _, sub := range s.AllOf {
		normalizeJSONSchema(sub)
	}
	for _, sub := range s.AnyOf {
		normalizeJSONSchema(sub)
	}
	for _, sub := range s.OneOf {
		normalizeJSONSchema(sub)
	}

	// Recurse into not
	normalizeJSONSchema(s.Not)

	// Recurse into if/then/else
	normalizeJSONSchema(s.If)
	normalizeJSONSchema(s.Then)
	normalizeJSONSchema(s.Else)

	// Recurse into dependent schemas
	for _, dep := range s.DependentSchemas {
		normalizeJSONSchema(dep)
	}

	// Recurse into definitions
	for _, def := range s.Definitions {
		normalizeJSONSchema(def)
	}

	// Recurse into contains
	normalizeJSONSchema(s.Contains)

	// Recurse into content schema
	normalizeJSONSchema(s.ContentSchema)

	// Recurse into property names
	normalizeJSONSchema(s.PropertyNames)
}

// NeedsSchemaNormalization checks whether a tool's parameter schema has
// non-standard type names that need fixing. Returns false for nil ParamsOneOf.
func NeedsSchemaNormalization(p *schema.ParamsOneOf) bool {
	if p == nil {
		return false
	}
	js, err := p.ToJSONSchema()
	if err != nil || js == nil {
		return false
	}
	return hasInvalidTypes(js)
}

// SchemaNormalizerTool wraps a BaseTool and normalizes the JSON Schema
// returned by Info() to fix non-standard type names from MCP servers.
type SchemaNormalizerTool struct {
	inner tool.BaseTool
}

// NewSchemaNormalizerTool creates a wrapper that normalizes the tool's parameter schema.
func NewSchemaNormalizerTool(inner tool.BaseTool) *SchemaNormalizerTool {
	return &SchemaNormalizerTool{inner: inner}
}

// Info returns the tool's metadata with normalized JSON Schema.
func (t *SchemaNormalizerTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	info, err := t.inner.Info(ctx)
	if err != nil {
		return nil, err
	}

	if info.ParamsOneOf != nil {
		js, jsonErr := info.ParamsOneOf.ToJSONSchema()
		if jsonErr == nil && js != nil {
			normalizeJSONSchema(js)
			info.ParamsOneOf = schema.NewParamsOneOfByJSONSchema(js)
		}
	}

	return info, nil
}

// InvokableRun delegates to the inner tool if it implements InvokableTool.
func (t *SchemaNormalizerTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if inv, ok := t.inner.(tool.InvokableTool); ok {
		return inv.InvokableRun(ctx, argumentsInJSON, opts...)
	}
	return "", fmt.Errorf("tool %s does not implement InvokableTool", t.inner)
}

// StreamableRun delegates to the inner tool if it implements StreamableTool.
func (t *SchemaNormalizerTool) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (*schema.StreamReader[string], error) {
	if str, ok := t.inner.(tool.StreamableTool); ok {
		return str.StreamableRun(ctx, argumentsInJSON, opts...)
	}
	return nil, fmt.Errorf("tool %s does not implement StreamableTool", t.inner)
}

// Ensure SchemaNormalizerTool satisfies the interfaces.
var _ tool.BaseTool = (*SchemaNormalizerTool)(nil)
var _ tool.InvokableTool = (*SchemaNormalizerTool)(nil)
var _ tool.StreamableTool = (*SchemaNormalizerTool)(nil)
