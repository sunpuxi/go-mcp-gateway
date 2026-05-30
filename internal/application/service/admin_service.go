package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	domainservice "github.com/sunpuxi/go-mcp-gateway/internal/domain/service"
)

// ======================== DTOs ========================

// ToolDTO 管理后台工具响应/请求，params 为 ParamRule 数组而非 RawMessage
type ToolDTO struct {
	ToolID      int64             `json:"tool_id"`
	ProjectID   string            `json:"project_id"`
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	HTTPMethod  string            `json:"http_method"`
	URLTemplate string            `json:"url_template"`
	BaseURL     string            `json:"base_url"`
	TimeoutMs   int               `json:"timeout_ms"`
	Params      []entity.ParamRule `json:"params"`
	Status      int               `json:"status"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// ProjectDTO 项目管理请求/响应
type ProjectDTO struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	BaseURL     string `json:"base_url"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// ClientDTO 客户端管理响应
type ClientDTO struct {
	ClientID     string `json:"client_id"`
	Name         string `json:"name"`
	APIKeyPrefix string `json:"api_key_prefix"`
	Description  string `json:"description"`
	Status       int    `json:"status"`
	ToolCount    int    `json:"tool_count"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// StatsDTO 仪表盘统计
type StatsDTO struct {
	Projects int `json:"projects"`
	Tools    int `json:"tools"`
	Clients  int `json:"clients"`
	Sessions int `json:"sessions"`
}

// ======================== AdminService ========================

// AdminService 管理后台应用层服务
type AdminService struct {
	projectRepo    repository.ProjectRepository
	toolRepo       repository.ToolAdminRepository
	clientRepo     repository.ClientRepository
	sessionManager *domainservice.SessionManager
}

// NewAdminService 创建 AdminService
func NewAdminService(
	projectRepo repository.ProjectRepository,
	toolRepo repository.ToolAdminRepository,
	clientRepo repository.ClientRepository,
	sessionManager *domainservice.SessionManager,
) *AdminService {
	return &AdminService{
		projectRepo:    projectRepo,
		toolRepo:       toolRepo,
		clientRepo:     clientRepo,
		sessionManager: sessionManager,
	}
}

// ======================== 仪表盘 ========================

// GetStats 获取仪表盘统计数据
func (s *AdminService) GetStats() (*StatsDTO, error) {
	projects, _ := s.projectRepo.Count()
	tools, _ := s.toolRepo.Count()
	clients, _ := s.clientRepo.Count(context.Background())
	sessions := s.sessionManager.ActiveCount()

	return &StatsDTO{
		Projects: projects,
		Tools:    tools,
		Clients:  clients,
		Sessions: sessions,
	}, nil
}

// ======================== 项目 CRUD ========================

// ListProjects 项目列表
func (s *AdminService) ListProjects(page, size int) ([]ProjectDTO, int, error) {
	projects, total, err := s.projectRepo.List(page, size)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]ProjectDTO, len(projects))
	for i, p := range projects {
		dtos[i] = ProjectDTO{
			ProjectID:   p.ProjectID,
			Name:        p.Name,
			BaseURL:     p.BaseURL,
			Description: p.Description,
			Status:      p.Status,
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return dtos, total, nil
}

// CreateProject 新建项目
func (s *AdminService) CreateProject(dto ProjectDTO) (*ProjectDTO, error) {
	project := &entity.Project{
		ProjectID:   dto.ProjectID,
		Name:        dto.Name,
		BaseURL:     dto.BaseURL,
		Description: dto.Description,
		Status:      dto.Status,
	}
	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}
	return &ProjectDTO{
		ProjectID:   project.ProjectID,
		Name:        project.Name,
		BaseURL:     project.BaseURL,
		Description: project.Description,
		Status:      project.Status,
	}, nil
}

// UpdateProject 更新项目
func (s *AdminService) UpdateProject(projectID string, dto ProjectDTO) (*ProjectDTO, error) {
	existing, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("项目不存在: %s", projectID)
	}

	existing.Name = dto.Name
	existing.BaseURL = dto.BaseURL
	existing.Description = dto.Description
	existing.Status = dto.Status

	if err := s.projectRepo.Update(existing); err != nil {
		return nil, err
	}

	return &ProjectDTO{
		ProjectID:   existing.ProjectID,
		Name:        existing.Name,
		BaseURL:     existing.BaseURL,
		Description: existing.Description,
		Status:      existing.Status,
	}, nil
}

// DeleteProject 删除项目（级联删除关联的工具及权限）
func (s *AdminService) DeleteProject(projectID string) error {
	// 1. 查找项目下的所有工具ID
	toolIDs, err := s.toolRepo.FindToolIDsByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("查询项目工具失败: %w", err)
	}

	// 2. 删除这些工具的客户端授权信息
	if len(toolIDs) > 0 {
		if err := s.clientRepo.DeletePermissionsByToolIDs(context.Background(), toolIDs); err != nil {
			return fmt.Errorf("删除工具授权信息失败: %w", err)
		}
	}

	// 3. 删除项目下的所有工具
	if err := s.toolRepo.DeleteByProjectID(projectID); err != nil {
		return fmt.Errorf("删除项目工具失败: %w", err)
	}

	// 4. 删除项目自身
	if err := s.projectRepo.Delete(projectID); err != nil {
		return fmt.Errorf("删除项目失败: %w", err)
	}

	return nil
}

// ======================== 工具 CRUD ========================

// ListTools 工具列表
func (s *AdminService) ListTools(page, size int) ([]ToolDTO, int, error) {
	tools, total, err := s.toolRepo.ListAll(page, size)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]ToolDTO, len(tools))
	for i, t := range tools {
		rules, _ := t.ParseParams()
		if rules == nil {
			rules = []entity.ParamRule{}
		}
		dtos[i] = ToolDTO{
			ToolID:      t.ToolID,
			ProjectID:   t.ProjectID,
			Name:        t.Name,
			Title:       t.Title,
			Description: t.Description,
			HTTPMethod:  t.HTTPMethod,
			URLTemplate: t.URLTemplate,
			BaseURL:     t.BaseURL,
			TimeoutMs:   t.TimeoutMs,
			Params:      rules,
			Status:      t.Status,
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   t.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return dtos, total, nil
}

// CreateTool 新建工具
func (s *AdminService) CreateTool(dto ToolDTO) (*ToolDTO, error) {
	paramsJSON, _ := json.Marshal(dto.Params)
	raw := json.RawMessage(paramsJSON)

	tool := &entity.Tool{
		ProjectID:   dto.ProjectID,
		Name:        dto.Name,
		Title:       dto.Title,
		Description: dto.Description,
		HTTPMethod:  dto.HTTPMethod,
		URLTemplate: dto.URLTemplate,
		TimeoutMs:   dto.TimeoutMs,
		Params:      &raw,
		Status:      dto.Status,
	}

	if err := s.toolRepo.Create(tool); err != nil {
		return nil, err
	}

	dto.ToolID = tool.ToolID
	return &dto, nil
}

// UpdateTool 更新工具
func (s *AdminService) UpdateTool(toolID int64, dto ToolDTO) (*ToolDTO, error) {
	existing, err := s.toolRepo.FindByID(toolID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("工具不存在: %d", toolID)
	}

	paramsJSON, _ := json.Marshal(dto.Params)
	raw := json.RawMessage(paramsJSON)

	existing.ProjectID = dto.ProjectID
	existing.Name = dto.Name
	existing.Title = dto.Title
	existing.Description = dto.Description
	existing.HTTPMethod = dto.HTTPMethod
	existing.URLTemplate = dto.URLTemplate
	existing.TimeoutMs = dto.TimeoutMs
	existing.Params = &raw
	existing.Status = dto.Status

	if err := s.toolRepo.Update(existing); err != nil {
		return nil, err
	}

	dto.ToolID = toolID
	return &dto, nil
}

// DeleteTool 删除工具（级联删除关联权限）
func (s *AdminService) DeleteTool(toolID int64) error {
	// 1. 删除该工具的客户端授权信息
	if err := s.clientRepo.DeletePermissionsByToolIDs(context.Background(), []int64{toolID}); err != nil {
		return fmt.Errorf("删除工具授权信息失败: %w", err)
	}
	// 2. 删除工具自身
	return s.toolRepo.Delete(toolID)
}

// ======================== 客户端 CRUD ========================

// ListClients 客户端列表
func (s *AdminService) ListClients(page, size int) ([]ClientDTO, int, error) {
	clients, toolCounts, total, err := s.clientRepo.ListAll(page, size)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]ClientDTO, len(clients))
	for i, c := range clients {
		dtos[i] = ClientDTO{
			ClientID:     c.ClientID,
			Name:         c.Name,
			APIKeyPrefix: c.APIKeyPrefix,
			Description:  c.Description,
			Status:       c.Status,
			ToolCount:    toolCounts[i],
			CreatedAt:    c.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    c.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return dtos, total, nil
}

// CreateClient 新建客户端（初始无 API Key，Key 需要单独生成）
func (s *AdminService) CreateClient(dto ClientDTO) (*ClientDTO, error) {
	ctx := context.Background()
	client := &entity.Client{
		ClientID:     dto.ClientID,
		Name:         dto.Name,
		APIKeyPrefix: "",
		APIKeyHash:   "",
		Description:  dto.Description,
		Status:       dto.Status,
	}
	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, err
	}
	dto.APIKeyPrefix = ""
	dto.ToolCount = 0
	return &dto, nil
}

// UpdateClient 更新客户端
func (s *AdminService) UpdateClient(clientID string, dto ClientDTO) (*ClientDTO, error) {
	ctx := context.Background()
	existing, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("客户端不存在: %s", clientID)
	}

	existing.Name = dto.Name
	existing.Description = dto.Description
	existing.Status = dto.Status

	if err := s.clientRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	dto.ClientID = clientID
	dto.APIKeyPrefix = existing.APIKeyPrefix
	return &dto, nil
}

// DeleteClient 删除客户端
func (s *AdminService) DeleteClient(clientID string) error {
	return s.clientRepo.Delete(context.Background(), clientID)
}

// GenerateAPIKey 生成 API Key，返回明文 Key（仅展示一次）
func (s *AdminService) GenerateAPIKey(clientID string) (string, error) {
	ctx := context.Background()
	existing, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		return "", err
	}
	if existing == nil {
		return "", fmt.Errorf("客户端不存在: %s", clientID)
	}

	// 生成随机 20 字节 hex
	randomBytes := make([]byte, 20)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}
	randomPart := hex.EncodeToString(randomBytes)

	// 前缀：sk- + client_id 中 cli_ 后面的部分
	prefix := "sk-" + clientID
	if len(clientID) > 4 && clientID[:4] == "cli_" {
		prefix = "sk-" + clientID[4:]
	}
	apiKey := prefix + "-" + randomPart

	// SHA256 哈希
	hashBytes := sha256.Sum256([]byte(apiKey))
	hash := hex.EncodeToString(hashBytes[:])

	if err := s.clientRepo.UpdateAPIKey(ctx, clientID, hash, prefix); err != nil {
		return "", err
	}

	return apiKey, nil
}

// ======================== 权限管理 ========================

// GetClientPermissions 查询客户端权限（tool_id 列表）
func (s *AdminService) GetClientPermissions(clientID string) ([]int64, error) {
	return s.clientRepo.FindToolPermissions(context.Background(), clientID)
}

// UpdateClientPermissions 更新客户端权限
func (s *AdminService) UpdateClientPermissions(clientID string, toolIDs []int64) error {
	return s.clientRepo.SavePermissions(context.Background(), clientID, toolIDs)
}
