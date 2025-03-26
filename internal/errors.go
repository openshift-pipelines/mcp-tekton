package internal

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func mcpError(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(message),
		},
		IsError: true,
	}
}
