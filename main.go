package main

import (
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/openshift-pipelines/mcp-tekton/internal"
)

func main() {
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(id any, method mcp.MCPMethod, message any) {
		fmt.Printf("beforeAny: %s, %v, %v\n", method, id, message)
	})
	hooks.AddOnSuccess(func(id any, method mcp.MCPMethod, message any, result any) {
		fmt.Printf("onSuccess: %s, %v, %v, %v\n", method, id, message, result)
	})
	hooks.AddOnError(func(id any, method mcp.MCPMethod, message any, err error) {
		fmt.Printf("onError: %s, %v, %v, %v\n", method, id, message, err)
	})
	hooks.AddBeforeInitialize(func(id any, message *mcp.InitializeRequest) {
		fmt.Printf("beforeInitialize: %v, %v\n", id, message)
	})
	hooks.AddAfterInitialize(func(id any, message *mcp.InitializeRequest, result *mcp.InitializeResult) {
		fmt.Printf("afterInitialize: %v, %v, %v\n", id, message, result)
	})
	hooks.AddAfterCallTool(func(id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		fmt.Printf("afterCallTool: %v, %v, %v\n", id, message, result)
	})
	hooks.AddBeforeCallTool(func(id any, message *mcp.CallToolRequest) {
		fmt.Printf("beforeCallTool: %v, %v\n", id, message)
	})

	// Create MCP server
	s := server.NewMCPServer(
		"Tekton",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
		server.WithHooks(hooks),
	)

	slog.Info("Addingtools, prompts, and resources to the server.")
	internal.AddTools(s)
	internal.AddPrompts(s)
	internal.AddResources(s)

	slog.Info("Starting the server.")
	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
