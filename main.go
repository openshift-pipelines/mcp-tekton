package main

import (
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/server"
	"github.com/openshift-pipelines/mcp-tekton/internal"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"Tekton",
		"0.0.1",
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
