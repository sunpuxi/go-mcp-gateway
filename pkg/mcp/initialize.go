package mcp

// InitializeRequest 是 Agent 发来的 initialize 请求参数
type InitializeRequest struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo      `json:"clientInfo"`
}

type ClientCapabilities struct {
	Tools     *struct{} `json:"tools,omitempty"`
	Resources *struct{} `json:"resources,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult 是 Gateway 返回的 initialize 响应
type InitializeResult struct {
	ProtocolVersion string              `json:"protocolVersion"`
	Capabilities    ServerCapabilities  `json:"capabilities"`
	ServerInfo      ServerInfo          `json:"serverInfo"`
}

type ServerCapabilities struct {
	Tools     *ServerToolsCapabilities `json:"tools,omitempty"`
	Resources *struct{}                `json:"resources,omitempty"`
}

type ServerToolsCapabilities struct {
	ListChanged bool `json:"listChanged"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NewInitializeResult 创建默认的 initialize 响应
func NewInitializeResult() *InitializeResult {
	return &InitializeResult{
		ProtocolVersion: "2025-06-18",
		Capabilities: ServerCapabilities{
			Tools: &ServerToolsCapabilities{ListChanged: false},
		},
		ServerInfo: ServerInfo{
			Name:    "mcp-gateway",
			Version: "1.0.0",
		},
	}
}
