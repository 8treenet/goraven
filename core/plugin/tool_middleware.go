package plugin

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

type toolInterceptorAdapter struct {
	plugins *Plugins
}

func (a *toolInterceptorAdapter) wrapInvokable(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
		modifiedArgs, skip := a.plugins.FireBeforeTool(ctx, input.Name, input.Arguments)
		if skip {
			return &compose.ToolOutput{Result: modifiedArgs}, nil
		}
		originalArgs := input.Arguments
		if modifiedArgs != originalArgs {
			input.Arguments = modifiedArgs
		}
		output, err := next(ctx, input)
		if output != nil {
			output.Result = a.plugins.FireAfterTool(ctx, input.Name, originalArgs, output.Result, err)
		}
		return output, err
	}
}

func (a *toolInterceptorAdapter) wrapStreamable(next compose.StreamableToolEndpoint) compose.StreamableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (*compose.StreamToolOutput, error) {
		modifiedArgs, skip := a.plugins.FireBeforeTool(ctx, input.Name, input.Arguments)
		if skip {
			return &compose.StreamToolOutput{Result: nil}, nil
		}
		originalArgs := input.Arguments
		if modifiedArgs != originalArgs {
			input.Arguments = modifiedArgs
		}
		output, err := next(ctx, input)
		return output, err
	}
}

func (a *toolInterceptorAdapter) wrapEnhancedInvokable(next compose.EnhancedInvokableToolEndpoint) compose.EnhancedInvokableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (*compose.EnhancedInvokableToolOutput, error) {
		modifiedArgs, skip := a.plugins.FireBeforeTool(ctx, input.Name, input.Arguments)
		if skip {
			return &compose.EnhancedInvokableToolOutput{Result: nil}, nil
		}
		originalArgs := input.Arguments
		if modifiedArgs != originalArgs {
			input.Arguments = modifiedArgs
		}
		output, err := next(ctx, input)
		return output, err
	}
}

func (a *toolInterceptorAdapter) wrapEnhancedStreamable(next compose.EnhancedStreamableToolEndpoint) compose.EnhancedStreamableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (*compose.EnhancedStreamableToolOutput, error) {
		modifiedArgs, skip := a.plugins.FireBeforeTool(ctx, input.Name, input.Arguments)
		if skip {
			return &compose.EnhancedStreamableToolOutput{Result: nil}, nil
		}
		originalArgs := input.Arguments
		if modifiedArgs != originalArgs {
			input.Arguments = modifiedArgs
		}
		output, err := next(ctx, input)
		return output, err
	}
}
