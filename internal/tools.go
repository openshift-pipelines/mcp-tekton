package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	v1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelineclient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	pipelineinformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1/pipeline"
	pipelineruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1/pipelinerun"
	taskinformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1/task"
	taskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1/taskrun"
	stepactioninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1beta1/stepaction"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

func toolStartTask() mcp.Tool {
	return mcp.NewTool("start_task",
		mcp.WithDescription("Start a Task"),
		mcp.WithString("name", mcp.Required(),
			mcp.Description("Name or Reference of the Task to sart"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace where the Task is located"),
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

	pipelineInformer := pipelineinformer.Get(ctx)
	pipelineclientset := pipelineclient.Get(ctx)

	if _, err := pipelineInformer.Lister().Pipelines(namespace).Get(name); err != nil {
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

	if _, err := pipelineclientset.TektonV1().PipelineRuns(namespace).Create(ctx, pr, metav1.CreateOptions{}); err != nil {
		return mcpError(fmt.Sprintf("Failed to create PipelineRun %s/%s: %v", namespace, name, err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Starting pipeline %s in namespace %s", name, namespace)),
		},
	}, nil
}

func handlerStartTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return nil, errors.New("name must be a string")
	}
	namespace, _ := request.Params.Arguments["namespace"].(string)

	taskInformer := taskinformer.Get(ctx)
	pipelineclientset := pipelineclient.Get(ctx)

	if _, err := taskInformer.Lister().Tasks(namespace).Get(name); err != nil {
		return mcpError(fmt.Sprintf("Failed to get Task %s/%s: %v", namespace, name, err)), nil
	}

	pr := &v1.TaskRun{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1",
			Kind:       "TaskRun",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: fmt.Sprintf("%s-", name),
		},
		Spec: v1.TaskRunSpec{
			TaskRef: &v1.TaskRef{
				Name: name,
			},
		},
	}

	if _, err := pipelineclientset.TektonV1().TaskRuns(namespace).Create(ctx, pr, metav1.CreateOptions{}); err != nil {
		return mcpError(fmt.Sprintf("Failed to create TaskRun %s/%s: %v", namespace, name, err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Starting task %s in namespace %s", name, namespace)),
		},
	}, nil
}

func handlerListTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskInformer := taskinformer.Get(ctx)
	namespace, err := OptionalParam[string](request, "namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lselector, err := OptionalParam[string](request, "label-selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prefix, err := OptionalParam[string](request, "prefix")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var selector labels.Selector
	if lselector != "" {
		selector, err = labels.Parse(lselector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		selector = labels.NewSelector()
	}

	var trs []*v1.Task

	if namespace == "" {
		// No namespace, searching all PipelineRuns
		trs, err = taskInformer.Lister().List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		trs, err = taskInformer.Lister().Tasks(namespace).List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	// Filter after the fact
	if prefix != "" {
		filteredTRs := []*v1.Task{}
		for _, pr := range trs {
			if strings.HasPrefix(pr.Name, prefix) {
				filteredTRs = append(filteredTRs, pr)
			}
		}
		trs = filteredTRs
	}

	jsonData, err := json.Marshal(trs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource to JSON: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func handlerListTaskRun(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskRunInformer := taskruninformer.Get(ctx)
	namespace, err := OptionalParam[string](request, "namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lselector, err := OptionalParam[string](request, "label-selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prefix, err := OptionalParam[string](request, "prefix")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var selector labels.Selector
	if lselector != "" {
		selector, err = labels.Parse(lselector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		selector = labels.NewSelector()
	}

	var trs []*v1.TaskRun

	if namespace == "" {
		// No namespace, searching all PipelineRuns
		trs, err = taskRunInformer.Lister().List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		trs, err = taskRunInformer.Lister().TaskRuns(namespace).List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	// Filter after the fact
	if prefix != "" {
		filteredTRs := []*v1.TaskRun{}
		for _, pr := range trs {
			if strings.HasPrefix(pr.Name, prefix) {
				filteredTRs = append(filteredTRs, pr)
			}
		}
		trs = filteredTRs
	}

	jsonData, err := json.Marshal(trs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource to JSON: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func handlerListStepaction(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stepactionInformer := stepactioninformer.Get(ctx)
	namespace, err := OptionalParam[string](request, "namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lselector, err := OptionalParam[string](request, "label-selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prefix, err := OptionalParam[string](request, "prefix")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var selector labels.Selector
	if lselector != "" {
		selector, err = labels.Parse(lselector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		selector = labels.NewSelector()
	}

	var trs []*v1beta1.StepAction

	if namespace == "" {
		// No namespace, searching all PipelineRuns
		trs, err = stepactionInformer.Lister().List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		trs, err = stepactionInformer.Lister().StepActions(namespace).List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	// Filter after the fact
	if prefix != "" {
		filteredTRs := []*v1beta1.StepAction{}
		for _, pr := range trs {
			if strings.HasPrefix(pr.Name, prefix) {
				filteredTRs = append(filteredTRs, pr)
			}
		}
		trs = filteredTRs
	}

	jsonData, err := json.Marshal(trs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource to JSON: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func handlerListPipeline(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pipelineInformer := pipelineinformer.Get(ctx)
	namespace, err := OptionalParam[string](request, "namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lselector, err := OptionalParam[string](request, "label-selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prefix, err := OptionalParam[string](request, "prefix")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var selector labels.Selector
	if lselector != "" {
		selector, err = labels.Parse(lselector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		selector = labels.NewSelector()
	}

	var prs []*v1.Pipeline

	if namespace == "" {
		// No namespace, searching all Pipelines
		prs, err = pipelineInformer.Lister().List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		prs, err = pipelineInformer.Lister().Pipelines(namespace).List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	// Filter after the fact
	if prefix != "" {
		filteredPRs := []*v1.Pipeline{}
		for _, pr := range prs {
			if strings.HasPrefix(pr.Name, prefix) {
				filteredPRs = append(filteredPRs, pr)
			}
		}
		prs = filteredPRs
	}

	jsonData, err := json.Marshal(prs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource to JSON: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func handlerListPipelineRun(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pipelineRunInformer := pipelineruninformer.Get(ctx)
	namespace, err := OptionalParam[string](request, "namespace")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	lselector, err := OptionalParam[string](request, "label-selector")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	prefix, err := OptionalParam[string](request, "prefix")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var selector labels.Selector
	if lselector != "" {
		selector, err = labels.Parse(lselector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		selector = labels.NewSelector()
	}

	var prs []*v1.PipelineRun

	if namespace == "" {
		// No namespace, searching all PipelineRuns
		prs, err = pipelineRunInformer.Lister().List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	} else {
		prs, err = pipelineRunInformer.Lister().PipelineRuns(namespace).List(selector)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	// Filter after the fact
	if prefix != "" {
		filteredPRs := []*v1.PipelineRun{}
		for _, pr := range prs {
			if strings.HasPrefix(pr.Name, prefix) {
				filteredPRs = append(filteredPRs, pr)
			}
		}
		prs = filteredPRs
	}

	jsonData, err := json.Marshal(prs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource to JSON: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func AddTools(s *server.MCPServer) {
	s.AddTool(toolStartPipeline(), handlerStartPipeline)
	s.AddTool(toolStartTask(), handlerStartTask)
	s.AddTool(mcp.NewTool("list_pipelineruns",
		mcp.WithDescription("List pipelineruns in the cluster with filtering options"),
		mcp.WithString("namespace", mcp.Description("Which namespace to use to look for PipelineRuns")),
		mcp.WithString("prefix", mcp.Description("Name prefix to filter PipelineRuns")),
		mcp.WithString("label-selector", mcp.Description("Label selector to filter PipelineRuns")),
	), handlerListPipelineRun)
	s.AddTool(mcp.NewTool("list_taskruns",
		mcp.WithDescription("List taskruns in the cluster with filtering options"),
		mcp.WithString("namespace", mcp.Description("Which namespace to use to look for Taskruns")),
		mcp.WithString("prefix", mcp.Description("Name prefix to filter Taskruns")),
		mcp.WithString("label-selector", mcp.Description("Label selector to filter Taskruns")),
	), handlerListTaskRun)
	s.AddTool(mcp.NewTool("list_pipelines",
		mcp.WithDescription("List pipelines in the cluster with filtering options"),
		mcp.WithString("namespace", mcp.Description("Which namespace to use to look for Pipeline")),
		mcp.WithString("prefix", mcp.Description("Name prefix to filter Pipeline")),
		mcp.WithString("label-selector", mcp.Description("Label selector to filter Pipeline")),
	), handlerListPipeline)
	s.AddTool(mcp.NewTool("list_tasks",
		mcp.WithDescription("List tasks in the cluster with filtering options"),
		mcp.WithString("namespace", mcp.Description("Which namespace to use to look for Task")),
		mcp.WithString("prefix", mcp.Description("Name prefix to filter Task")),
		mcp.WithString("label-selector", mcp.Description("Label selector to filter Task")),
	), handlerListTask)
	s.AddTool(mcp.NewTool("list_stepactions",
		mcp.WithDescription("List stepactions in the cluster with filtering options"),
		mcp.WithString("namespace", mcp.Description("Which namespace to use to look for Stepactions")),
		mcp.WithString("prefix", mcp.Description("Name prefix to filter Stepactions")),
		mcp.WithString("label-selector", mcp.Description("Label selector to filter Stepactions")),
	), handlerListStepaction)
}
