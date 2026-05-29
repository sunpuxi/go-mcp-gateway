package mcp

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain"
	"github.com/sunpuxi/go-mcp-gateway/pkg/jsonrpc"
	"github.com/sunpuxi/go-mcp-gateway/pkg/mcp"
)

type Handler struct {
	sessionManager *domain.SessionManager
	toolRepo       domain.ToolQuerier
	httpClient     *domain.HTTPClient
}

func NewHandler(sm *domain.SessionManager, tr domain.ToolQuerier, hc *domain.HTTPClient) *Handler {
	return &Handler{sessionManager: sm, toolRepo: tr, httpClient: hc}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, nil, jsonrpc.CodeParseError, "无法读取请求体")
		return
	}

	// 解析 JSON-RPC 请求
	req, err := jsonrpc.ParseRequest(body)
	if err != nil {
		writeError(w, nil, jsonrpc.CodeParseError, "JSON 解析失败: "+err.Error())
		return
	}

	// 路由到对应 method handler
	switch req.Method {
	case "initialize":
		h.handleInitialize(w, r, req)
	case "notifications/initialized":
		h.handleInitialized(w, r, req)
	case "tools/list":
		h.handleToolsList(w, r, req)
	case "tools/call":
		h.handleToolsCall(w, r, req)
	default:
		writeError(w, req.ID, jsonrpc.CodeMethodNotFound, "不支持的 method: "+req.Method)
	}
}

// --- initialize ---

func (h *Handler) handleInitialize(w http.ResponseWriter, r *http.Request, req *jsonrpc.Request) {
	var initReq mcp.InitializeRequest
	if err := json.Unmarshal(req.Params, &initReq); err != nil {
		writeError(w, req.ID, jsonrpc.CodeInvalidParams, "initialize 参数解析失败: "+err.Error())
		return
	}

	// 创建 Session
	session := h.sessionManager.Create()

	// 设置 Session Header
	w.Header().Set("Mcp-Session-Id", session.ID)

	// 返回 InitializeResult
	result := mcp.NewInitializeResult()
	resp := jsonrpc.NewResponse(req.ID, result)
	writeJSON(w, http.StatusOK, resp)
}

// --- notifications/initialized ---

func (h *Handler) handleInitialized(w http.ResponseWriter, r *http.Request, req *jsonrpc.Request) {
	sessionID := getSessionID(r)
	if sessionID == "" || !h.sessionManager.MarkInitialized(sessionID, "2025-06-18") {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// --- tools/list ---

func (h *Handler) handleToolsList(w http.ResponseWriter, r *http.Request, req *jsonrpc.Request) {
	// 验证 Session
	sessionID := getSessionID(r)
	session, ok := h.sessionManager.Get(sessionID)
	if !ok || !session.Initialized {
		writeError(w, req.ID, jsonrpc.CodeInvalidRequest, "请先完成 initialize")
		return
	}

	// 查询所有启用的工具
	tools, err := h.toolRepo.ListEnabled()
	if err != nil {
		log.Printf("查询工具列表失败: %v", err)
		writeError(w, req.ID, jsonrpc.CodeInternalError, "查询工具列表失败")
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
	writeJSON(w, http.StatusOK, resp)
}

// --- tools/call ---

func (h *Handler) handleToolsCall(w http.ResponseWriter, r *http.Request, req *jsonrpc.Request) {
	// 验证 Session
	sessionID := getSessionID(r)
	session, ok := h.sessionManager.Get(sessionID)
	if !ok || !session.Initialized {
		writeError(w, req.ID, jsonrpc.CodeInvalidRequest, "请先完成 initialize")
		return
	}
	_ = session

	// 解析 tools/call 请求参数
	var callReq mcp.ToolCallRequest
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		writeError(w, req.ID, jsonrpc.CodeInvalidParams, "tools/call 参数解析失败: "+err.Error())
		return
	}

	// 查询工具定义
	tool, err := h.toolRepo.FindByName(callReq.Name)
	if err != nil {
		log.Printf("查询工具失败 %s: %v", callReq.Name, err)
		writeError(w, req.ID, jsonrpc.CodeInvalidParams, "工具不存在: "+callReq.Name)
		return
	}

	// 参数映射
	mappedReq, err := domain.MapParams(tool, callReq.Arguments)
	if err != nil {
		log.Printf("参数映射失败: %v", err)
		writeError(w, req.ID, jsonrpc.CodeParamInvalid, "参数映射失败: "+err.Error())
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
		// 网络错误或 5xx → JSON-RPC error
		var msg string
		if httpErr != nil {
			msg = "下游服务异常: " + httpErr.Error()
		} else {
			msg = "下游服务异常，状态码: " + http.StatusText(statusCode)
		}
		writeError(w, req.ID, jsonrpc.CodeDownstreamErr, msg)
		return
	}

	// 正常返回（含 isError 标记）
	resp := jsonrpc.NewResponse(req.ID, result)
	writeJSON(w, http.StatusOK, resp)
}

// --- 辅助方法 ---

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("响应序列化失败: %v", err)
	}
}

func writeError(w http.ResponseWriter, id *int64, code int, message string) {
	resp := jsonrpc.NewErrorResponse(id, code, message)
	writeJSON(w, http.StatusOK, resp)
}

// getSessionID 从请求 Header 中获取 Mcp-Session-Id
func getSessionID(r *http.Request) string {
	return r.Header.Get("Mcp-Session-Id")
}
