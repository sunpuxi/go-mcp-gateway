package mapper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
)

// MappedRequest 是参数映射后的 HTTP 请求信息
type MappedRequest struct {
	Path   string
	Query  url.Values
	Header http.Header
	Body   map[string]any
}

// MapParams 将 MCP arguments 按工具的参数规则映射为 HTTP 请求信息
func MapParams(tool *entity.Tool, arguments map[string]any) (*MappedRequest, error) {
	rules, err := tool.ParseParams()
	if err != nil {
		return nil, fmt.Errorf("解析参数规则失败: %w", err)
	}

	mapped := &MappedRequest{
		Path:   tool.URLTemplate,
		Query:  url.Values{},
		Header: http.Header{},
		Body:   make(map[string]any),
	}

	for _, rule := range rules {
		val, exists := arguments[rule.Name]

		if rule.Required && !exists {
			return nil, fmt.Errorf("缺少必填参数: %s", rule.Name)
		}

		if !exists {
			if rule.DefaultValue != "" {
				val = rule.DefaultValue
			} else {
				continue
			}
		}

		strVal := fmt.Sprintf("%v", val)

		switch rule.Location {
		case "path":
			mapped.Path = strings.ReplaceAll(mapped.Path, "{"+rule.Name+"}", url.PathEscape(strVal))
		case "query":
			mapped.Query.Add(rule.Name, strVal)
		case "header":
			mapped.Header.Set(rule.Name, strVal)
		case "body":
			mapped.Body[rule.Name] = val
		}
	}

	return mapped, nil
}

// BuildURL 将 MappedRequest 拼装为完整 URL 字符串
func (m *MappedRequest) BuildURL(baseURL string) string {
	u := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(m.Path, "/")
	if len(m.Query) > 0 {
		u += "?" + m.Query.Encode()
	}
	return u
}

// BuildBody 将 Body map 序列化为 JSON bytes，若无数据返回 nil
func (m *MappedRequest) BuildBody() []byte {
	if len(m.Body) == 0 {
		return nil
	}
	data, _ := json.Marshal(m.Body)
	return data
}
