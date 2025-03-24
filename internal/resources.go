package internal

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func resourcePipeline() mcp.ResourceTemplate {
	return mcp.NewResourceTemplate(
		"tekton://{namespace}/{type}/{name}",
		"Pipeline",
		mcp.WithTemplateDescription("Returns Tekton Pipeline's information"),
		mcp.WithTemplateMIMEType("application/json"),
	)
}

func handlerPipeline(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return nil, errors.New("Not implemented yet")
}

func AddResources(s *server.MCPServer) {
	s.AddResourceTemplate(resourcePipeline(), handlerPipeline)
}
