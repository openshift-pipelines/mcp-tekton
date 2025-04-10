package main

import (
	// "context"
	"fmt"
	"io"
	"log/slog"
	"os"

	// "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/openshift-pipelines/mcp-tekton/internal"
	"k8s.io/client-go/tools/clientcmd"
	filteredinformerfactory "knative.dev/pkg/client/injection/kube/informers/factory/filtered"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/signals"
)

// ManagedByLabelKey is the label key used to mark what is managing this resource
const ManagedByLabelKey = "app.kubernetes.io/managed-by"

func main() {
	// hooks := &server.Hooks{}
	//
	// hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
	// 	fmt.Printf("beforeAny: %s, %v, %v\n", method, id, message)
	// })
	// hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
	// 	fmt.Printf("onSuccess: %s, %v, %v, %v\n", method, id, message, result)
	// })
	// hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
	// 	fmt.Printf("onError: %s, %v, %v, %v\n", method, id, message, err)
	// })
	// hooks.AddBeforeInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest) {
	// 	fmt.Printf("beforeInitialize: %v, %v\n", id, message)
	// })
	// hooks.AddAfterInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest, result *mcp.InitializeResult) {
	// 	fmt.Printf("afterInitialize: %v, %v, %v\n", id, message, result)
	// })
	// hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
	// 	fmt.Printf("afterCallTool: %v, %v, %v\n", id, message, result)
	// })
	// hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
	// 	fmt.Printf("beforeCallTool: %v, %v\n", id, message)
	// })

	// Create MCP server
	s := server.NewMCPServer(
		"Tekton",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
		server.WithLogging(),
		// server.WithHooks(hooks),
	)

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	cfg, err := kubeConfig.ClientConfig()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to get Kubernetes config: %v", err))
		os.Exit(1)
	}

	ctx := signals.NewContext()
	ctx = filteredinformerfactory.WithSelectors(ctx, ManagedByLabelKey)
	// slog.Info("Registering %d informer factories", len(injection.Default.GetInformerFactories()))
	// slog.Info("Registering %d informers", len(injection.Default.GetInformers()))
	ctx, startInformers := injection.EnableInjectionOrDie(ctx, cfg)

	// Start the injection clients and informers.
	startInformers()

	slog.Info("Addingtools, prompts, and resources to the server.")
	internal.AddTools(s)
	internal.AddPrompts(s)
	internal.AddResources(ctx, s)

	slog.Info("Starting the server.")
	// Start the stdio server
	stdioServer := server.NewStdioServer(s)
	// Start listening for messages
	errC := make(chan error, 1)
	go func() {
		in, out := io.Reader(os.Stdin), io.Writer(os.Stdout)

		errC <- stdioServer.Listen(ctx, in, out)
	}()

	// Output tekton-mcp string
	_, _ = fmt.Fprintf(os.Stderr, "Tekton MCP Server running on stdio\n")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		slog.Info("shutting down server...")
	case err := <-errC:
		if err != nil {
			slog.Error(fmt.Sprintf("error running server: %v", err))
			os.Exit(1)
		}
	}
}
