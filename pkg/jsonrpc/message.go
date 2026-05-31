package jsonrpc

import "encoding/json"

const Version = "2.0"

// --- Request ---

// Request 表示 JSON-RPC 2.0 请求
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`  // nil 表示通知（notification），无需回复
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func ParseRequest(body []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// --- Response ---

// Response 表示 JSON-RPC 2.0 响应（含 result 或 error）
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int64      `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ErrorObj   `json:"error,omitempty"`
}

func NewResponse(id *int64, result interface{}) *Response {
	return &Response{
		JSONRPC: Version,
		ID:      id,
		Result:  result,
	}
}

func NewErrorResponse(id *int64, code int, message string) *Response {
	return &Response{
		JSONRPC: Version,
		ID:      id,
		Error:   &ErrorObj{Code: code, Message: message},
	}
}

func (r *Response) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// --- Error ---

// ErrorObj 表示 JSON-RPC 2.0 错误对象
type ErrorObj struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// --- 标准错误码 ---

const (
	CodeParseError     = -32700 // JSON 解析错误
	CodeInvalidRequest = -32600 // 无效请求
	CodeMethodNotFound = -32601 // 方法不存在
	CodeInvalidParams  = -32602 // 无效参数
	CodeInternalError  = -32603 // 内部错误
)

// MCP Gateway 自定义错误码
const (
	CodeAuthError     = -32001 // 鉴权失败
	CodeRateLimit     = -32002 // 请求过于频繁
	CodeParamInvalid  = -32003 // 参数校验失败
	CodeDownstreamErr = -32005 // 下游服务异常
	CodeCircuitOpen   = -32006 // 熔断打开，下游服务暂时不可用
)

// --- 解析通知 ---

// IsNotification 判断请求是否为通知（无 ID）
func (r *Request) IsNotification() bool {
	return r.ID == nil
}
