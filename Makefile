
.PHONY: build
build:
	go build -v .

.PHONY: inspect
inspect: build
	npx @modelcontextprotocol/inspector@0.7.0 ./mcp-tekton
