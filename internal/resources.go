package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	pipelineruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1/pipelinerun"
	"k8s.io/client-go/tools/cache"
)

type resources struct {
	prlock      sync.RWMutex
	pipelineRun map[string]*v1.PipelineRun
}

func (r *resources) AddPipelineRun(name string, pipelineRun *v1.PipelineRun) {
	r.prlock.Lock()
	defer r.prlock.Unlock()
	r.pipelineRun[name] = pipelineRun
}

func (r *resources) UpdatePipelineRun(name string, pipelineRun *v1.PipelineRun) {
	r.prlock.Lock()
	defer r.prlock.Unlock()
	r.pipelineRun[name] = pipelineRun
}

func (r *resources) RemovePipelineRun(name string) {
	r.prlock.Lock()
	defer r.prlock.Unlock()
	delete(r.pipelineRun, name)
}

func (r *resources) handlerPipelineRun(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	r.prlock.RLock()
	defer r.prlock.RUnlock()
	uri := request.Params.URI
	pr, ok := r.pipelineRun[uri]

	if !ok {
		return nil, fmt.Errorf("PipelineRun %s not found", request.Params.URI)
	}

	// Convert the result to JSON
	jsonData, err := json.Marshal(pr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource to JSON: %w", err)
	}

	// Create resource contents
	contentType := fmt.Sprintf("application/json;type=%s", "pipelinerun")
	contents := mcp.TextResourceContents{
		URI:      uri,
		MIMEType: contentType,
		Text:     string(jsonData),
	}

	return []mcp.ResourceContents{contents}, nil
}

func AddResources(ctx context.Context, s *server.MCPServer) {
	r := &resources{
		prlock:      sync.RWMutex{},
		pipelineRun: make(map[string]*v1.PipelineRun),
	}
	pipelineRunInformer := pipelineruninformer.Get(ctx)
	if _, err := pipelineRunInformer.Informer().AddEventHandler(updatePipelineRunResource(r, s)); err != nil {
		slog.Error(fmt.Sprintf("Couldn't register PipelineRun informer event handler: %v\n", err))
	}
}

func updatePipelineRunResource(r *resources, s *server.MCPServer) cache.ResourceEventHandler {
	h := func(obj interface{}) {
		pipelineRun := obj.(*v1.PipelineRun)
		uri := fmt.Sprintf("tekton://%s/pipelinerun/%s", pipelineRun.Namespace, pipelineRun.Name)
		resource := mcp.Resource{
			URI:  uri,
			Name: pipelineRun.Name,
		}
		slog.Info(fmt.Sprintf("Updating resource: %s\n", pipelineRun.Name))
		s.AddResource(resource, r.handlerPipelineRun)
		r.AddPipelineRun(uri, pipelineRun)
	}
	dh := func(obj interface{}) {
		pipelineRun := obj.(*v1.PipelineRun)
		slog.Info(fmt.Sprintf("Removing resource: %s\n", pipelineRun.Name))
		r.RemovePipelineRun(pipelineRun.Name)
	}
	return cache.ResourceEventHandlerFuncs{
		AddFunc: h,
		UpdateFunc: func(first, second interface{}) {
			h(second)
		},
		DeleteFunc: dh,
	}
}
