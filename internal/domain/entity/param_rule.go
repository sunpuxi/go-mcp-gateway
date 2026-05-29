package entity

// ParamRule 是 params JSON 字段中单条参数映射规则
type ParamRule struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	Location     string `json:"location"` // path / query / body / header
	Required     bool   `json:"required,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
	Description  string `json:"description,omitempty"`
}
