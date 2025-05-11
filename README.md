# ARCHIVED: mcp-tekton: a Tekton Model Context Protocol Server

> [!IMPORTANT]
> This repository is archived and work on a Tekton MCP server is happening [upstream](https://tekton.dev), in [`tektoncd/mcp-server`](https://github.com/tektoncd/mcp-server)

This is a proof of concept of a [Model Context Protocol](https://modelcontextprotocol.io/introduction) for Tetkon.

## Notes

- inspector: npx @modelcontextprotocol/inspector ./mcp-tekton
- https://github.com/Flux159/mcp-server-kubernetes/blob/main/src/resources/handlers.ts
- https://github.com/ckreiling/mcp-server-docker/blob/main/src/mcp_server_docker/server.py
- https://github.com/ckreiling/mcp-server-docker/blob/main/src/mcp_server_docker/server.py

## To-dos

- Go from listing some `PipelineRun` (or other objects) and then inspect it.
  - a tool to list / filter
  - a resource to inspect one
