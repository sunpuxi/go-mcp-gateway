package service

import (
	"github.com/sunpuxi/go-mcp-gateway/pkg/mcp"
)

// BuildToolCallResult 将下游 HTTP 响应转为 MCP ToolCallResult
// statusCode: 下游 HTTP 状态码
// body: 下游响应体
// err: 网络/传输层错误（超时、连接失败等）
// 返回 (result, shouldReturnError)
//   - shouldReturnError=true  调用方应返回 JSON-RPC error（协议级错误）
//   - shouldReturnError=false 调用方应返回 result（业务级结果，含 isError 标记）
func BuildToolCallResult(statusCode int, body []byte, err error) (*mcp.ToolCallResult, bool) {
	if err != nil {
		return nil, true
	}

	bodyStr := string(body)

	if statusCode >= 200 && statusCode < 300 {
		return &mcp.ToolCallResult{
			Content: []mcp.ContentItem{
				{Type: "text", Text: bodyStr},
			},
			IsError: false,
		}, false
	}

	if statusCode >= 400 && statusCode < 500 {
		return &mcp.ToolCallResult{
			Content: []mcp.ContentItem{
				{Type: "text", Text: bodyStr},
			},
			IsError: true,
		}, false
	}

	return nil, true
}
