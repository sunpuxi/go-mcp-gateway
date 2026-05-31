package command

import "github.com/sunpuxi/go-mcp-gateway/pkg/mcp"

// CallToolInput 封装 tools/call 的输入参数
type CallToolInput struct {
	Name      string
	Arguments map[string]any
}

// RejectType 表示请求被拒绝的原因类型
type RejectType string

const (
	RejectAuth          RejectType = "auth"
	RejectRateLimit     RejectType = "rate_limit"
	RejectCircuitOpen   RejectType = "circuit_open"
	RejectDownstreamErr RejectType = "downstream_error"
)

// RejectReason 统一表示请求被拦截的原因
type RejectReason struct {
	Type    RejectType
	Message string
}

// CallToolOutput 封装 tools/call 的输出结果
// Result 和 Reject 互斥：成功时 Result 非 nil，被拦截时 Reject 非 nil
type CallToolOutput struct {
	Result *mcp.ToolCallResult
	Reject *RejectReason
}
