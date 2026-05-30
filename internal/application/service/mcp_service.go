package service

import (
	"fmt"
	"net/http"

	"github.com/sunpuxi/go-mcp-gateway/internal/application/command"
	"github.com/sunpuxi/go-mcp-gateway/internal/application/query"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/mapper"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	infrahttp "github.com/sunpuxi/go-mcp-gateway/internal/infrastructure/http"
	"github.com/sunpuxi/go-mcp-gateway/pkg/mcp"
)

// MCPService 应用层服务，组装领域层的工具相关操作
type MCPService struct {
	toolRepo   repository.ToolQuerier
	httpClient *infrahttp.HTTPClient
}

func NewMCPService(toolRepo repository.ToolQuerier, httpClient *infrahttp.HTTPClient) *MCPService {
	return &MCPService{toolRepo: toolRepo, httpClient: httpClient}
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

	// 2. 参数映射（MCP arguments → HTTP path/query/body/header）
	mappedReq, err := mapper.MapParams(tool, input.Arguments)
	if err != nil {
		return nil, fmt.Errorf("参数映射失败: %w", err)
	}

	// 3. 拼装完整 URL 和请求体
	url := mappedReq.BuildURL(tool.BaseURL)
	body := mappedReq.BuildBody()

	// 4. 发起 HTTP 请求到下游服务（含重试）
	statusCode, respBody, httpErr := doRequestWithRetry(
		s.httpClient, tool.HTTPMethod, url, mappedReq.Header, body, tool.TimeoutMs,
		tool.RetryConfig,
	)

	// 5. 处理响应
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

// hasPermission 检查工具名是否在权限列表中
func (s *MCPService) hasPermission(permissions []string, toolName string) bool {
	for _, p := range permissions {
		if p == toolName {
			return true
		}
	}
	return false
}

