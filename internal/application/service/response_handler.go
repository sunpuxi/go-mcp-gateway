package service

import "github.com/sunpuxi/go-mcp-gateway/pkg/mcp"

// BuildToolCallResult 将下游 HTTP 响应转为 MCP ToolCallResult
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
