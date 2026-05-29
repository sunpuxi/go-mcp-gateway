package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain"
	"github.com/sunpuxi/go-mcp-gateway/pkg/jsonrpc"
	"github.com/sunpuxi/go-mcp-gateway/pkg/mcp"
)

const sseHeartbeatInterval = 15 * time.Second

type Handler struct {
	sessionManager *domain.SessionManager
	toolRepo       domain.ToolQuerier
	httpClient     *domain.HTTPClient
}

func NewHandler(sm *domain.SessionManager, tr domain.ToolQuerier, hc *domain.HTTPClient) *Handler {
	return &Handler{sessionManager: sm, toolRepo: tr, httpClient: hc}
}

// ============================================================================
//  SSE 传输 — GET /sse
//  建立 SSE 长连接，作为服务端推送通道
// ============================================================================

func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// 为本次 SSE 连接创建 Session
	session := h.sessionManager.Create()

	// 创建 SSE 响应通道 (缓冲 16 条，防止慢消费阻塞)
	ch := make(chan []byte, 16)
	session.SSECh = ch

	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 发送 endpoint 事件，告知客户端 POST 消息的地址（含 session_id）
	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: /messages?session_id=%s\n\n", session.ID)
	flusher.Flush()

	log.Printf("[SSE] 新连接 session=%s", session.ID)

	// 心跳 + 转发响应
	for {
		select {
		case data, ok := <-ch:
			if !ok {
				// 通道已关闭
				return
			}
			_, _ = fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			flusher.Flush()

		case <-r.Context().Done():
			// 客户端断开连接
			log.Printf("[SSE] 断开连接 session=%s", session.ID)
			session.SSECh = nil
			close(ch)
			h.sessionManager.Delete(session.ID)
			return

		case <-time.After(sseHeartbeatInterval):
			// 心跳保活 (SSE comment-only line)
			_, _ = fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

// ============================================================================
//  消息接收 — POST /messages?session_id=xxx
//  接收客户端发送的 JSON-RPC 消息，处理后通过 SSE 通道返回响应
// ============================================================================

func (h *Handler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	// 从查询参数获取 session_id
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "missing session_id", http.StatusBadRequest)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(nil, jsonrpc.CodeParseError, "无法读取请求体"))
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// 解析 JSON-RPC 请求
	req, err := jsonrpc.ParseRequest(body)
	if err != nil {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(nil, jsonrpc.CodeParseError, "JSON 解析失败: "+err.Error()))
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// 路由到对应 method handler
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

	// POST 统一快速返回 202 Accepted，响应走 SSE 通道
	w.WriteHeader(http.StatusAccepted)
}

// ============================================================================
//  消息处理 — initialize
// ============================================================================

func (h *Handler) handleInitialize(sessionID string, req *jsonrpc.Request) {
	var initReq mcp.InitializeRequest
	if err := json.Unmarshal(req.Params, &initReq); err != nil {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidParams, "initialize 参数解析失败: "+err.Error()))
		return
	}
	_ = initReq

	// 标记 Session 为已初始化（初始化通过 SSE 连接建立时自动创建，此处只做标记）
	h.sessionManager.MarkInitialized(sessionID, "2025-06-18")

	// 返回 InitializeResult
	result := mcp.NewInitializeResult()
	resp := jsonrpc.NewResponse(req.ID, result)
	h.sendSSE(sessionID, resp)
}

// ============================================================================
//  消息处理 — notifications/initialized
// ============================================================================

func (h *Handler) handleInitialized(sessionID string, req *jsonrpc.Request) {
	_ = req // notification, 无 ID 无响应
	h.sessionManager.MarkInitialized(sessionID, "2025-06-18")
	// Notification 不发送任何响应，仅返回 HTTP 202
}

// ============================================================================
//  消息处理 — tools/list
// ============================================================================

func (h *Handler) handleToolsList(sessionID string, req *jsonrpc.Request) {
	// 验证 Session
	session, ok := h.sessionManager.Get(sessionID)
	if !ok || !session.Initialized {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidRequest, "请先完成 initialize"))
		return
	}

	// 查询所有启用的工具
	tools, err := h.toolRepo.ListEnabled()
	if err != nil {
		log.Printf("查询工具列表失败: %v", err)
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInternalError, "查询工具列表失败"))
		return
	}

	// 组装 MCP 工具定义列表
	defs := make([]mcp.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		rules, _ := t.ParseParams()
		inputSchema := domain.BuildInputSchema(rules)

		defs = append(defs, mcp.ToolDefinition{
			Name:        t.Name,
			Title:       t.Title,
			Description: t.Description,
			InputSchema: inputSchema,
		})
	}

	result := mcp.ToolListResult{Tools: defs}
	resp := jsonrpc.NewResponse(req.ID, result)
	h.sendSSE(sessionID, resp)
}

// ============================================================================
//  消息处理 — tools/call
// ============================================================================

func (h *Handler) handleToolsCall(sessionID string, req *jsonrpc.Request) {
	// 验证 Session
	session, ok := h.sessionManager.Get(sessionID)
	if !ok || !session.Initialized {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidRequest, "请先完成 initialize"))
		return
	}
	_ = session

	// 解析 tools/call 请求参数
	var callReq mcp.ToolCallRequest
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidParams, "tools/call 参数解析失败: "+err.Error()))
		return
	}

	// 查询工具定义
	tool, err := h.toolRepo.FindByName(callReq.Name)
	if err != nil {
		log.Printf("查询工具失败 %s: %v", callReq.Name, err)
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeInvalidParams, "工具不存在: "+callReq.Name))
		return
	}

	// 参数映射
	mappedReq, err := domain.MapParams(tool, callReq.Arguments)
	if err != nil {
		log.Printf("参数映射失败: %v", err)
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeParamInvalid, "参数映射失败: "+err.Error()))
		return
	}

	// 拼装完整 URL
	url := mappedReq.BuildURL(tool.BaseURL)

	// 构建请求体
	body := mappedReq.BuildBody()

	// 发起 HTTP 请求
	statusCode, respBody, httpErr := h.httpClient.DoRequest(
		tool.HTTPMethod, url, mappedReq.Header, body, tool.TimeoutMs,
	)

	// 处理响应
	result, isJSONRPCError := domain.BuildToolCallResult(statusCode, respBody, httpErr)
	if isJSONRPCError {
		var msg string
		if httpErr != nil {
			msg = "下游服务异常: " + httpErr.Error()
		} else {
			msg = "下游服务异常，状态码: " + http.StatusText(statusCode)
		}
		h.sendSSE(sessionID, jsonrpc.NewErrorResponse(req.ID, jsonrpc.CodeDownstreamErr, msg))
		return
	}

	// 正常返回（含 isError 标记）
	resp := jsonrpc.NewResponse(req.ID, result)
	h.sendSSE(sessionID, resp)
}

// ============================================================================
//  辅助方法
// ============================================================================

// sendSSE 将 JSON-RPC 响应发送到指定 Session 的 SSE 通道
func (h *Handler) sendSSE(sessionID string, resp *jsonrpc.Response) {
	session, ok := h.sessionManager.Get(sessionID)
	if !ok || session.SSECh == nil {
		log.Printf("[SSE] 发送失败: session=%s 不存在或无 SSE 通道", sessionID)
		return
	}

	data, err := resp.Marshal()
	if err != nil {
		log.Printf("[SSE] 响应序列化失败: %v", err)
		return
	}

	select {
	case session.SSECh <- data:
	default:
		log.Printf("[SSE] 通道已满，丢弃响应 session=%s", sessionID)
	}
}
