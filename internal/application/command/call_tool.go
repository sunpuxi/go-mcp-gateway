package command

import "github.com/sunpuxi/go-mcp-gateway/pkg/mcp"

// CallToolInput 封装 tools/call 的输入参数
type CallToolInput struct {
	Name      string
	Arguments map[string]any
}

// CallToolOutput 封装 tools/call 的输出结果
type CallToolOutput struct {
	Result          *mcp.ToolCallResult
	AuthError       string // 非空时表示鉴权失败
	DownstreamError string // 非空时表示下游服务异常（网络错误 / 5xx）
}
