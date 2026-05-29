package domain

import (
	"github.com/sunpuxi/go-mcp-gateway/pkg/mcp"
)

// BuildToolCallResult 将下游 HTTP 响应转为 MCP ToolCallResult
// statusCode: 下游 HTTP 状态码
// body: 下游响应体
// err: 网络/传输层错误（超时、连接失败等）
// 返回 (result, shouldReturnError)
//   - shouldReturnError=true 时，调用方应返回 JSON-RPC error（协议级错误）
//   - shouldReturnError=false 时，调用方应返回 result（业务级结果，含 isError 标记）
func BuildToolCallResult(statusCode int, body []byte, err error) (*mcp.ToolCallResult, bool) {
	// 网络/传输层错误 → JSON-RPC error
	if err != nil {
		return nil, true
	}

	bodyStr := string(body)

	// 2xx → 正常结果
	if statusCode >= 200 && statusCode < 300 {
		return &mcp.ToolCallResult{
			Content: []mcp.ContentItem{
				{Type: "text", Text: bodyStr},
			},
			IsError: false,
		}, false
	}

	// 4xx → 业务错误（isError: true）
	if statusCode >= 400 && statusCode < 500 {
		return &mcp.ToolCallResult{
			Content: []mcp.ContentItem{
				{Type: "text", Text: bodyStr},
			},
			IsError: true,
		}, false
	}

	// 5xx → Gateway 未能获得有效响应 → JSON-RPC error
	return nil, true
}
