package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sunpuxi/go-mcp-gateway/internal/application/command"
	"github.com/sunpuxi/go-mcp-gateway/internal/application/query"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/circuitbreaker"
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
	cbRegistry  *circuitbreaker.Registry  // 熔断器注册表（可为 nil，表示不启用熔断）
	rateLimiter ratelimit.RateLimiter     // 限流器（nil 时使用 NoopLimiter 放行）
}

func NewMCPService(toolRepo repository.ToolQuerier, httpClient *infrahttp.HTTPClient, cbRegistry *circuitbreaker.Registry, rateLimiter ratelimit.RateLimiter) *MCPService {
	return &MCPService{toolRepo: toolRepo, httpClient: httpClient, cbRegistry: cbRegistry, rateLimiter: rateLimiter}
}

// ListTools 获取客户端有权限的 MCP 工具定义列表
func (s *MCPService) ListTools(permissions []string) (*query.ToolListOutput, error) {
	tools, err := s.toolRepo.ListEnabled()
	if err != nil {
		return nil, fmt.Errorf("查询工具列表失败: %w", err)
	}

	// 用 map 加速权限查找
	permSet := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		permSet[p] = struct{}{}
	}

	defs := make([]mcp.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		if _, ok := permSet[t.Name]; !ok {
			continue // 跳过无权限的工具
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

// CallTool 执行一个工具的完整调用链路：
//
//	鉴权 → 查工具 → 参数映射 → 熔断检查 → HTTP 请求（含重试） → 记录熔断 → 处理响应
func (s *MCPService) CallTool(input command.CallToolInput, permissions []string) (*command.CallToolOutput, error) {
	// 0. 鉴权：检查工具调用权限
	if !s.hasPermission(permissions, input.Name) {
		return &command.CallToolOutput{
			AuthError: fmt.Sprintf("无权调用工具: %s", input.Name),
		}, nil
	}

	// 1. 查询工具定义
	tool, err := s.toolRepo.FindByName(input.Name)
	if err != nil {
		return nil, fmt.Errorf("工具不存在: %s", input.Name)
	}

	// 1.5 限流检查（在参数映射前，避免浪费计算资源）
	if cfg := tool.RateLimitConfig; cfg != nil && cfg.IsEnabled() {
		key := "rate_limit:tool:" + input.Name
		allowed, remaining, rlErr := s.getRateLimiter().Allow(context.Background(), key, cfg.MaxRequests, cfg.WindowSeconds)
		if rlErr != nil {
			logger.Warn("限流检查异常，降级放行",
				"tool_name", input.Name,
				"error", rlErr,
			)
		} else if !allowed {
			logger.Warn("触发限流",
				"tool_name", input.Name,
				"max_requests", cfg.MaxRequests,
				"window_seconds", cfg.WindowSeconds,
			)
			return &command.CallToolOutput{
				RateLimited: fmt.Sprintf("工具 %s 请求过于频繁，请稍后重试", input.Name),
			}, nil
		} else {
			logger.Debug("限流检查通过",
				"tool_name", input.Name,
				"remaining", remaining,
			)
		}
	}

	// 2. 参数映射（MCP arguments → HTTP path/query/body/header）
	mappedReq, err := mapper.MapParams(tool, input.Arguments)
	if err != nil {
		return nil, fmt.Errorf("参数映射失败: %w", err)
	}

	// 3. 拼装完整 URL 和请求体
	url := mappedReq.BuildURL(tool.BaseURL)
	body := mappedReq.BuildBody()

	// 4. 熔断检查
	cb := s.getCircuitBreaker(tool.ProjectID)
	if cb != nil && !cb.Allow() {
		logger.Warn("熔断打开，拒绝请求",
			"project_id", tool.ProjectID,
			"tool_name", input.Name,
		)
		return &command.CallToolOutput{
			CircuitOpen: fmt.Sprintf("下游服务 %s 正在熔断中，请稍后重试", tool.ProjectID),
		}, nil
	}

	// 5. 发起 HTTP 请求到下游服务（含重试）
	statusCode, respBody, httpErr := doRequestWithRetry(
		s.httpClient, tool.HTTPMethod, url, mappedReq.Header, body, tool.TimeoutMs,
		tool.RetryConfig,
	)

	// 6. 记录熔断结果（网络错误 / 5xx 视为失败）
	if cb != nil {
		prevState := cb.State()
		if httpErr != nil || statusCode >= 500 {
			cb.RecordFailure()
			currState := cb.State()
			if prevState != currState {
				logger.Warn("熔断器状态变更",
					"project_id", tool.ProjectID,
					"tool_name", input.Name,
					"from", prevState.String(),
					"to", currState.String(),
				)
			}
		} else {
			cb.RecordSuccess()
			currState := cb.State()
			if prevState != currState {
				logger.Info("熔断器状态变更",
					"project_id", tool.ProjectID,
					"tool_name", input.Name,
					"from", prevState.String(),
					"to", currState.String(),
				)
			}
		}
	}

	// 7. 处理响应
	result, isDownstreamErr := BuildToolCallResult(statusCode, respBody, httpErr)
	if isDownstreamErr {
		var msg string
		if httpErr != nil {
			msg = "下游服务异常: " + httpErr.Error()
		} else {
			msg = "下游服务异常，状态码: " + http.StatusText(statusCode)
		}
		return &command.CallToolOutput{DownstreamError: msg}, nil
	}

	return &command.CallToolOutput{Result: result}, nil
}

// getCircuitBreaker 获取指定 Project 的熔断器，nil 表示未启用熔断
func (s *MCPService) getCircuitBreaker(projectID string) *circuitbreaker.CircuitBreaker {
	if s.cbRegistry == nil {
		return nil
	}
	return s.cbRegistry.Get(projectID)
}

// getRateLimiter 返回限流器实例，nil 时降级为 NoopLimiter
func (s *MCPService) getRateLimiter() ratelimit.RateLimiter {
	if s.rateLimiter == nil {
		return &ratelimit.NoopLimiter{}
	}
	return s.rateLimiter
}

// hasPermission 检查工具名是否在权限列表中
func (s *MCPService) hasPermission(permissions []string, toolName string) bool {
	for _, p := range permissions {
		if p == toolName {
			return true
		}
	}
	return false
}
