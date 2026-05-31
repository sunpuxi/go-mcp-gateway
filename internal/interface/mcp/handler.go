package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	appcommand "github.com/sunpuxi/go-mcp-gateway/internal/application/command"
	appservice "github.com/sunpuxi/go-mcp-gateway/internal/application/service"
	sessionpkg "github.com/sunpuxi/go-mcp-gateway/internal/application/session"
	domainservice "github.com/sunpuxi/go-mcp-gateway/internal/domain/service"
	"github.com/sunpuxi/go-mcp-gateway/pkg/jsonrpc"
	"github.com/sunpuxi/go-mcp-gateway/pkg/logger"
	"github.com/sunpuxi/go-mcp-gateway/pkg/mcp"
)

const sseHeartbeatInterval = 15 * time.Second

// Handler 是 MCP 传输层的 HTTP 处理器
type Handler struct {
	sessionManager *sessionpkg.Manager
	mcpService     *appservice.MCPService
	authService    *domainservice.AuthService
}

func NewHandler(sm *sessionpkg.Manager, ms *appservice.MCPService, as *domainservice.AuthService) *Handler {
	return &Handler{sessionManager: sm, mcpService: ms, authService: as}
}

// extractAPIKey 从 Authorization Header 中提取 Bearer Token
func extractAPIKey(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", false
	}
	// 支持 "Bearer <key>" 和 "<key>" 两种格式
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(auth[7:]), true
	}
	return strings.TrimSpace(auth), true
}

// ============================================================================
//  SSE 传输 — GET /sse
// ============================================================================

func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// 0. 鉴权
	apiKey, ok := extractAPIKey(r)
	if !ok || apiKey == "" {
		http.Error(w, "缺少 API Key，请在 Authorization Header 中提供", http.StatusUnauthorized)
		return
	}

	clientID, permissions, err := h.authService.Authenticate(r.Context(), apiKey)
	if err != nil {
		logger.Warn("SSE 鉴权失败", "error", err)
		http.Error(w, "API Key 认证失败: "+err.Error(), http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Error("SSE 连接失败：ResponseWriter 不支持 Flusher，请检查反向代理是否关闭缓冲",
			"client_id", clientID)
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	session := h.sessionManager.Create(clientID, permissions)

	ch := make(chan []byte, 16)
	session.SSECh = ch

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: /messages?session_id=%s\n\n", session.ID)
	flusher.Flush()

	logger.Info("SSE 新连接",
		"session_id", session.ID,
		"client_id", clientID,
		"tools", len(permissions))

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			flusher.Flush()

		case <-r.Context().Done():
			logger.Info("SSE 断开连接", "session_id", session.ID)
			session.SSECh = nil
			close(ch)
			h.sessionManager.Delete(session.ID)
			return

		case <-time.After(sseHeartbeatInterval):
			_, _ = fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

// ============================================================================
//  消息接收 — POST /messages?session_id=xxx
// ============================================================================

func (h *Handler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "missing session_id", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(nil, jsonrpc.CodeParseError, "无法读取请求体"))
		w.WriteHeader(http.StatusAccepted)
		return
	}

	req, err := jsonrpc.ParseRequest(body)
	if err != nil {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(nil, jsonrpc.CodeParseError, "JSON 解析失败: "+err.Error()))
		w.WriteHeader(http.StatusAccepted)
		return
	}

	switch req.Method {
	case "initialize":
		h.handleInitialize(sessionID, req)
	case "notifications/initialized":
		h.handleInitialized(sessionID, req)
	case "tools/list":
		h.handleToolsList(sessionID, req)
	case "tools/call":
		h.handleToolsCall(sessionID, req)
	default:
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeMethodNotFound, "不支持的 method: "+req.Method))
	}

	w.WriteHeader(http.StatusAccepted)
}

// ============================================================================
//  消息处理
// ============================================================================

func (h *Handler) handleInitialize(sessionID string, req *jsonrpc.Request) {
	var initReq mcp.InitializeRequest
	if err := json.Unmarshal(req.Params, &initReq); err != nil {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidParams, "initialize 参数解析失败: "+err.Error()))
		return
	}
	_ = initReq

	h.sessionManager.MarkInitialized(sessionID, "2025-06-18")

	result := mcp.NewInitializeResult()
	resp := jsonrpc.NewResponse(req.ID, result)
	h.sendSSE(sessionID, resp)
}

func (h *Handler) handleInitialized(sessionID string, req *jsonrpc.Request) {
	_ = req
	h.sessionManager.MarkInitialized(sessionID, "2025-06-18")
}

func (h *Handler) handleToolsList(sessionID string, req *jsonrpc.Request) {
	session, ok := h.sessionManager.Get(sessionID)
	if !ok || !session.Initialized {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidRequest, "请先完成 initialize"))
		return
	}

	output, err := h.mcpService.ListTools(session.Permissions)
	if err != nil {
		logger.Error("查询工具列表失败", "error", err)
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInternalError, "查询工具列表失败"))
		return
	}

	result := mcp.ToolListResult{Tools: output.Tools}
	h.sendSSE(sessionID, jsonrpc.NewResponse(req.ID, result))
}

func (h *Handler) handleToolsCall(sessionID string, req *jsonrpc.Request) {
	session, ok := h.sessionManager.Get(sessionID)
	if !ok || !session.Initialized {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidRequest, "请先完成 initialize"))
		return
	}

	var callReq mcp.ToolCallRequest
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidParams, "tools/call 参数解析失败: "+err.Error()))
		return
	}

	output, err := h.mcpService.CallTool(appcommand.CallToolInput{
		Name:      callReq.Name,
		Arguments: callReq.Arguments,
	}, session.Permissions)
	if err != nil {
		logger.Error("工具调用失败", "error", err)
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidParams, err.Error()))
		return
	}
	if output.AuthError != "" {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeAuthError, output.AuthError))
		return
	}
	if output.RateLimited != "" {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeRateLimit, output.RateLimited))
		return
	}
	if output.CircuitOpen != "" {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeCircuitOpen, output.CircuitOpen))
		return
	}
	if output.DownstreamError != "" {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeDownstreamErr, output.DownstreamError))
		return
	}

	h.sendSSE(sessionID, jsonrpc.NewResponse(req.ID, output.Result))
}

// ============================================================================
//  辅助方法
// ============================================================================

func (h *Handler) sendSSE(sessionID string, resp *jsonrpc.Response) {
	session, ok := h.sessionManager.Get(sessionID)
	if !ok || session.SSECh == nil {
		logger.Warn("SSE 发送失败，session 不存在或通道已关闭", "session_id", sessionID)
		return
	}

	data, err := resp.Marshal()
	if err != nil {
		logger.Error("SSE 响应序列化失败", "error", err)
		return
	}

	select {
	case session.SSECh <- data:
	default:
		logger.Warn("SSE 通道已满，丢弃响应", "session_id", sessionID)
	}
}
