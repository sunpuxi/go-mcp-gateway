package query

import "github.com/sunpuxi/go-mcp-gateway/pkg/mcp"

// ToolListOutput 封装 tools/list 的输出
type ToolListOutput struct {
	Tools []mcp.ToolDefinition
}
