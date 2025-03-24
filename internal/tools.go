package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func toolStartPipeline() mcp.Tool {
	return mcp.NewTool("start_pipeline",
		mcp.WithDescription("Start a Pipeline"),
		mcp.WithString("name", mcp.Required(),
			mcp.Description("Name or Reference of the Pipeline to sart"),
		),
		// TODO add "parameters" objects
	)
}

func handlerStartPipeline(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return nil, errors.New("name must be a string")
	}

	return nil, fmt.Errorf("Not implemented yet (%s)", name)
}

func AddTools(s *server.MCPServer) {
	s.AddTool(toolStartPipeline(), handlerStartPipeline)
}
