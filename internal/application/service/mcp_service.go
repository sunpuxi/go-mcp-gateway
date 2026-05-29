package service

import (
	"fmt"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/mapper"
	"net/http"

	"github.com/sunpuxi/go-mcp-gateway/internal/application/command"
	"github.com/sunpuxi/go-mcp-gateway/internal/application/query"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	domainservice "github.com/sunpuxi/go-mcp-gateway/internal/domain/service"
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

// ListTools 获取所有启用的 MCP 工具定义列表
func (s *MCPService) ListTools() (*query.ToolListOutput, error) {
	tools, err := s.toolRepo.ListEnabled()
	if err != nil {
		return nil, fmt.Errorf("查询工具列表失败: %w", err)
	}

	defs := make([]mcp.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		rules, _ := t.ParseParams()
		inputSchema := domainservice.BuildInputSchema(rules)

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
func (s *MCPService) CallTool(input command.CallToolInput) (*command.CallToolOutput, error) {
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

	// 4. 发起 HTTP 请求到下游服务
	statusCode, respBody, httpErr := s.httpClient.DoRequest(
		tool.HTTPMethod, url, mappedReq.Header, body, tool.TimeoutMs,
	)

	// 5. 处理响应
	result, isDownstreamErr := domainservice.BuildToolCallResult(statusCode, respBody, httpErr)
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
