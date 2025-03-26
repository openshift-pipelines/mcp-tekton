package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektoncs "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func toolStartPipeline() mcp.Tool {
	return mcp.NewTool("start_pipeline",
		mcp.WithDescription("Start a Pipeline"),
		mcp.WithString("name", mcp.Required(),
			mcp.Description("Name or Reference of the Pipeline to sart"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace where the Pipeline is located"),
			mcp.DefaultString("default"),
		),
		// TODO add "parameters" objects
	)
}

func handlerStartPipeline(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return nil, errors.New("name must be a string")
	}
	namespace, _ := request.Params.Arguments["namespace"].(string)

	// Get Kubernetes config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return mcpError(fmt.Errorf("failed to get Kubernetes config: %w", err).Error()), nil
	}

	// Create Tekton clientset
	tektonClient, err := tektoncs.NewForConfig(config)
	if err != nil {
		return mcpError(fmt.Errorf("failed to create Tekton client: %w", err).Error()), nil
	}

	if _, err := tektonClient.TektonV1().Pipelines(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		return mcpError(fmt.Sprintf("Failed to get Pipeline %s/%s: %v", namespace, name, err)), nil
	}

	pr := &v1.PipelineRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1",
			Kind:       "PipelineRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: fmt.Sprintf("%s-", name),
		},
		Spec: v1.PipelineRunSpec{
			PipelineRef: &v1.PipelineRef{
				Name: name,
			},
		},
	}

	if _, err := tektonClient.TektonV1().PipelineRuns(namespace).Create(ctx, pr, metav1.CreateOptions{}); err != nil {
		return mcpError(fmt.Sprintf("Failed to create PipelineRun %s/%s: %v", namespace, name, err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Starting pipeline %s in namespace %s", name, namespace)),
		},
	}, nil
}

func AddTools(s *server.MCPServer) {
	s.AddTool(toolStartPipeline(), handlerStartPipeline)
}
