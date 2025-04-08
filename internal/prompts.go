package internal

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func handlerExplainPipelineError(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return nil, errors.New("not Implemented yet")
}

func AddPrompts(s *server.MCPServer) {
	s.AddPrompt(mcp.NewPrompt("explain_pipeline_error",
		mcp.WithPromptDescription("Explain the error of a Pipeline"),
		mcp.WithArgument("pipeline", mcp.RequiredArgument(),
			mcp.ArgumentDescription("Name of the Pipeline"),
		),
	), handlerExplainPipelineError)
}
