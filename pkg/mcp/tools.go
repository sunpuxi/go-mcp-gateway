package mcp

import "encoding/json"

// --- tools/list ---

// ToolListResult 是 tools/list 的响应结果
type ToolListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

// ToolDefinition 是 MCP 协议中单个工具的定义
type ToolDefinition struct {
	Name        string          `json:"name"`
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// --- tools/call ---

// ToolCallRequest 是 tools/call 的请求参数
type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResult 是 tools/call 的响应结果
type ToolCallResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem 是 MCP 响应内容条目
type ContentItem struct {
	Type string `json:"type"` // "text" 或其他
	Text string `json:"text"`
}
