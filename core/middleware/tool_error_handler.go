package middleware

import (
	"context"
	"fmt"

	"goraven/util"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type ToolErrorHandlerConfig struct {
	// FormatError converts error to string message for LLM.
	FormatError func(ctx context.Context, toolName string, err error) string
}

type ToolErrorHandler struct {
	config *ToolErrorHandlerConfig
}

func NewToolErrorHandler(config *ToolErrorHandlerConfig) *ToolErrorHandler {
	if config == nil {
		config = &ToolErrorHandlerConfig{}
	}
	if config.FormatError == nil {
		config.FormatError = defaultFormatError
	}
	return &ToolErrorHandler{
		config: config,
	}
}

func defaultFormatError(_ context.Context, toolName string, err error) string {
	return fmt.Sprintf("Tool '%s' execution failed: %s", toolName, err.Error())
}

func (h *ToolErrorHandler) WrapInvokableToolCall(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (output *compose.ToolOutput, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v\n%s", r, util.TrimStack(3))
			}
			if err == nil {
				return
			}
			errorMsg := h.config.FormatError(ctx, input.Name, err)
			output = &compose.ToolOutput{Result: errorMsg}
			err = nil
		}()
		output, err = next(ctx, input)
		return output, err
	}
}

func (h *ToolErrorHandler) WrapStreamableToolCall(next compose.StreamableToolEndpoint) compose.StreamableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (output *compose.StreamToolOutput, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v\n%s", r, util.TrimStack(3))
			}
			if err == nil {
				return
			}
			errorMsg := h.config.FormatError(ctx, input.Name, err)
			output = &compose.StreamToolOutput{
				Result: schema.StreamReaderFromArray([]string{errorMsg}),
			}
			err = nil
		}()
		output, err = next(ctx, input)
		return output, err
	}
}

func (h *ToolErrorHandler) WrapEnhancedInvokableToolCall(next compose.EnhancedInvokableToolEndpoint) compose.EnhancedInvokableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (output *compose.EnhancedInvokableToolOutput, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v\n%s", r, util.TrimStack(3))
			}
			if err == nil {
				return
			}
			errorMsg := h.config.FormatError(ctx, input.Name, err)
			output = &compose.EnhancedInvokableToolOutput{
				Result: &schema.ToolResult{
					Parts: []schema.ToolOutputPart{
						{Type: schema.ToolPartTypeText, Text: errorMsg},
					},
				},
			}
			err = nil
		}()
		output, err = next(ctx, input)
		return output, err
	}
}

func (h *ToolErrorHandler) WrapEnhancedStreamableToolCall(next compose.EnhancedStreamableToolEndpoint) compose.EnhancedStreamableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (output *compose.EnhancedStreamableToolOutput, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v\n%s", r, util.TrimStack(3))
			}
			if err == nil {
				return
			}
			errorMsg := h.config.FormatError(ctx, input.Name, err)
			output = &compose.EnhancedStreamableToolOutput{
				Result: schema.StreamReaderFromArray([]*schema.ToolResult{
					{
						Parts: []schema.ToolOutputPart{
							{Type: schema.ToolPartTypeText, Text: errorMsg},
						},
					},
				}),
			}
			err = nil
		}()
		output, err = next(ctx, input)
		return output, err
	}
}

func (h *ToolErrorHandler) ToToolMiddleware() compose.ToolMiddleware {
	return compose.ToolMiddleware{
		Invokable:          h.WrapInvokableToolCall,
		Streamable:         h.WrapStreamableToolCall,
		EnhancedInvokable:  h.WrapEnhancedInvokableToolCall,
		EnhancedStreamable: h.WrapEnhancedStreamableToolCall,
	}
}