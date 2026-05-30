package mapper

import (
	"encoding/json"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
)

// BuildInputSchema 将 ParamRule 切片转为 MCP inputSchema（JSON Schema 格式）
func BuildInputSchema(rules []entity.ParamRule) json.RawMessage {
	type propertySchema struct {
		Type        string `json:"type"`
		Description string `json:"description,omitempty"`
		Default     any    `json:"default,omitempty"`
	}

	type inputSchema struct {
		Type       string                    `json:"type"`
		Properties map[string]propertySchema `json:"properties"`
		Required   []string                  `json:"required,omitempty"`
	}

	schema := inputSchema{
		Type:       "object",
		Properties: make(map[string]propertySchema),
	}

	for _, r := range rules {
		prop := propertySchema{
			Type:        r.Type,
			Description: r.Description,
		}
		if r.DefaultValue != "" && !r.Required {
			prop.Default = r.DefaultValue
		}
		schema.Properties[r.Name] = prop
		if r.Required {
			schema.Required = append(schema.Required, r.Name)
		}
	}

	data, _ := json.Marshal(schema)
	return data
}
