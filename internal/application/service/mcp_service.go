package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sunpuxi/go-mcp-gateway/internal/application/command"
	"github.com/sunpuxi/go-mcp-gateway/internal/application/query"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/circuitbreaker"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/mapper"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	infrahttp "github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/http"
	"github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/ratelimit"
	"github.com/sunpuxi/go-mcp-gateway/pkg/logger"
	"github.com/sunpuxi/go-mcp-gateway/pkg/mcp"
)

// MCPService 应用层服务，组装领域层的工具相关操作
type MCPService struct {
	toolRepo    repository.ToolQuerier
	httpClient  *infrahttp.HTTPClient
	cbRegistry  *circuitbreaker.Registry
	rateLimiter ratelimit.RateLimiter
}

func NewMCPService(toolRepo repository.ToolQuerier, httpClient *infrahttp.HTTPClient, cbRegistry *circuitbreaker.Registry, rateLimiter ratelimit.RateLimiter) *MCPService {
	return &MCPService{toolRepo: toolRepo, httpClient: httpClient, cbRegistry: cbRegistry, rateLimiter: rateLimiter}
}

// httpRequest 封装一次完整的下游 HTTP 请求信息
type httpRequest struct {
	method  string
	url     string
	header  map[string][]string
	body    []byte
	timeout int
}

// ListTools 获取客户端有权限的 MCP 工具定义列表
func (s *MCPService) ListTools(permissions []string) (*query.ToolListOutput, error) {
	tools, err := s.toolRepo.ListEnabled()
	if err != nil {
		return nil, fmt.Errorf("查询工具列表失败: %w", err)
	}

	permSet := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		permSet[p] = struct{}{}
	}

	defs := make([]mcp.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		if _, ok := permSet[t.Name]; !ok {
			continue
		}
		rules, _ := t.ParseParams()
		inputSchema := mapper.BuildInputSchema(rules)

		defs = append(defs, mcp.ToolDefinition{
			Name:        t.Name,
			Title:       t.Title,
			Description: t.Description,
			InputSchema: inputSchema,
		})
	}

	return &query.ToolListOutput{Tools: defs}, nil
}

// CallTool 执行工具的完整调用链路
func (s *MCPService) CallTool(input command.CallToolInput, permissions []string) (*command.CallToolOutput, error) {
	if out := s.checkAuth(input.Name, permissions); out != nil {
		return out, nil
	}

	// 加载工具并检查限流
	tool, out, err := s.loadAndRateLimit(input)
	if out != nil || err != nil {
		return out, err
	}

	// 构建 HTTP 请求
	req, out, err := s.buildHTTPRequest(tool, input.Arguments)
	if out != nil || err != nil {
		return out, err
	}

	// 检查熔断
	if out := s.checkCircuitBreaker(tool.ProjectID); out != nil {
		return out, nil
	}

	// 转发请求（重试+超时时间限制）
	code, body, httpErr := doRequestWithRetry(
		s.httpClient, req.method, req.url, req.header, req.body, req.timeout,
		tool.RetryConfig,
	)

	// 熔断器信息更新（连续失败次数等信息的更新）
	s.recordCircuitResult(tool.ProjectID, tool.Name, httpErr, code)

	return s.buildResponse(code, body, httpErr), nil
}

// checkAuth 鉴权：检查工具调用权限
func (s *MCPService) checkAuth(toolName string, permissions []string) *command.CallToolOutput {
	if s.hasPermission(permissions, toolName) {
		return nil
	}
	return &command.CallToolOutput{
		Reject: &command.RejectReason{
			Type:    command.RejectAuth,
			Message: fmt.Sprintf("无权调用工具: %s", toolName),
		},
	}
}

// loadAndRateLimit 加载工具定义并检查限流
func (s *MCPService) loadAndRateLimit(input command.CallToolInput) (*entity.Tool, *command.CallToolOutput, error) {
	tool, err := s.toolRepo.FindByName(input.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("工具不存在: %s", input.Name)
	}

	if out := s.checkRateLimit(tool); out != nil {
		return nil, out, nil
	}

	return tool, nil, nil
}

// checkRateLimit 限流检查
func (s *MCPService) checkRateLimit(tool *entity.Tool) *command.CallToolOutput {
	cfg := tool.RateLimitConfig
	if cfg == nil || !cfg.IsEnabled() {
		return nil
	}

	key := "rate_limit:tool:" + tool.Name
	allowed, remaining, err := s.getRateLimiter().Allow(context.Background(), key, cfg.MaxRequests, cfg.WindowSeconds)
	if err != nil {
		logger.Warn("限流检查异常，降级放行", "tool_name", tool.Name, "error", err)
		return nil
	}
	if !allowed {
		logger.Warn("触发限流", "tool_name", tool.Name,
			"max_requests", cfg.MaxRequests, "window_seconds", cfg.WindowSeconds)
		return &command.CallToolOutput{
			Reject: &command.RejectReason{
				Type:    command.RejectRateLimit,
				Message: fmt.Sprintf("工具 %s 请求过于频繁，请稍后重试", tool.Name),
			},
		}
	}

	logger.Debug("限流检查通过", "tool_name", tool.Name, "remaining", remaining)
	return nil
}

// buildHTTPRequest 参数映射并构建 HTTP 请求
func (s *MCPService) buildHTTPRequest(tool *entity.Tool, arguments map[string]any) (*httpRequest, *command.CallToolOutput, error) {
	mappedReq, err := mapper.MapParams(tool, arguments)
	if err != nil {
		return nil, nil, fmt.Errorf("参数映射失败: %w", err)
	}

	return &httpRequest{
		method:  tool.HTTPMethod,
		url:     mappedReq.BuildURL(tool.BaseURL),
		header:  mappedReq.Header,
		body:    mappedReq.BuildBody(),
		timeout: tool.TimeoutMs,
	}, nil, nil
}

// checkCircuitBreaker 熔断检查
func (s *MCPService) checkCircuitBreaker(projectID string) *command.CallToolOutput {
	cb := s.getCircuitBreaker(projectID)
	if cb == nil || cb.Allow() {
		return nil
	}

	logger.Warn("熔断打开，拒绝请求", "project_id", projectID)
	return &command.CallToolOutput{
		Reject: &command.RejectReason{
			Type:    command.RejectCircuitOpen,
			Message: fmt.Sprintf("下游服务 %s 正在熔断中，请稍后重试", projectID),
		},
	}
}

// recordCircuitResult 记录熔断结果（网络错误 / 5xx 视为失败）
func (s *MCPService) recordCircuitResult(projectID, toolName string, httpErr error, statusCode int) {
	cb := s.getCircuitBreaker(projectID)
	if cb == nil {
		return
	}

	prevState := cb.State()
	if httpErr != nil || statusCode >= 500 {
		cb.RecordFailure()
	} else {
		cb.RecordSuccess()
	}

	if currState := cb.State(); prevState != currState {
		logger.Warn("熔断器状态变更",
			"project_id", projectID,
			"tool_name", toolName,
			"from", prevState.String(),
			"to", currState.String(),
		)
	}
}

// buildResponse 根据 HTTP 结果构建最终输出
func (s *MCPService) buildResponse(statusCode int, body []byte, httpErr error) *command.CallToolOutput {
	result, isDownstreamErr := BuildToolCallResult(statusCode, body, httpErr)
	if !isDownstreamErr {
		return &command.CallToolOutput{Result: result}
	}

	var msg string
	if httpErr != nil {
		msg = "下游服务异常: " + httpErr.Error()
	} else {
		msg = "下游服务异常，状态码: " + http.StatusText(statusCode)
	}
	return &command.CallToolOutput{
		Reject: &command.RejectReason{
			Type:    command.RejectDownstreamErr,
			Message: msg,
		},
	}
}

// ============================================================================
//  辅助方法
// ============================================================================

func (s *MCPService) getCircuitBreaker(projectID string) *circuitbreaker.CircuitBreaker {
	if s.cbRegistry == nil {
		return nil
	}
	return s.cbRegistry.Get(projectID)
}

func (s *MCPService) getRateLimiter() ratelimit.RateLimiter {
	if s.rateLimiter == nil {
		return &ratelimit.NoopLimiter{}
	}
	return s.rateLimiter
}

func (s *MCPService) hasPermission(permissions []string, toolName string) bool {
	for _, p := range permissions {
		if p == toolName {
			return true
		}
	}
	return false
}
