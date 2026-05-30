package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sessionpkg "github.com/sunpuxi/go-mcp-gateway/internal/application/session"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	"github.com/sunpuxi/go-mcp-gateway/internal/interface/dto"
)

// AdminService 管理后台应用层服务
type AdminService struct {
	projectRepo    repository.ProjectRepository
	toolRepo       repository.ToolAdminRepository
	clientRepo     repository.ClientRepository
	sessionManager *sessionpkg.Manager
}

// NewAdminService 创建 AdminService
func NewAdminService(
	projectRepo repository.ProjectRepository,
	toolRepo repository.ToolAdminRepository,
	clientRepo repository.ClientRepository,
	sessionManager *sessionpkg.Manager,
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
func (s *AdminService) GetStats() (*dto.StatsDTO, error) {
	projects, _ := s.projectRepo.Count()
	tools, _ := s.toolRepo.Count()
	clients, _ := s.clientRepo.Count(context.Background())
	sessions := s.sessionManager.ActiveCount()

	// 构建活跃 session 列表
	sessionEntities := s.sessionManager.List()
	sessionList := make([]dto.SessionInfoDTO, len(sessionEntities))
	for i, se := range sessionEntities {
		sessionList[i] = dto.SessionInfoDTO{
			ID:              se.ID,
			ClientID:        se.ClientID,
			ProtocolVersion: se.ProtocolVersion,
			Initialized:     se.Initialized,
			CreatedAt:       se.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &dto.StatsDTO{
		Projects:    projects,
		Tools:       tools,
		Clients:     clients,
		Sessions:    sessions,
		SessionList: sessionList,
	}, nil
}

// ======================== 项目 CRUD ========================

// ListProjects 项目列表
func (s *AdminService) ListProjects(page, size int) ([]dto.ProjectDTO, int, error) {
	projects, total, err := s.projectRepo.List(page, size)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]dto.ProjectDTO, len(projects))
	for i, p := range projects {
		dtos[i] = dto.ProjectDTO{
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
func (s *AdminService) CreateProject(req dto.ProjectDTO) (*dto.ProjectDTO, error) {
	project := &entity.Project{
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		BaseURL:     req.BaseURL,
		Description: req.Description,
		Status:      req.Status,
	}
	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}
	return &dto.ProjectDTO{
		ProjectID:   project.ProjectID,
		Name:        project.Name,
		BaseURL:     project.BaseURL,
		Description: project.Description,
		Status:      project.Status,
	}, nil
}

// UpdateProject 更新项目
func (s *AdminService) UpdateProject(projectID string, req dto.ProjectDTO) (*dto.ProjectDTO, error) {
	existing, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("项目不存在: %s", projectID)
	}

	existing.Name = req.Name
	existing.BaseURL = req.BaseURL
	existing.Description = req.Description
	existing.Status = req.Status

	if err := s.projectRepo.Update(existing); err != nil {
		return nil, err
	}

	return &dto.ProjectDTO{
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
func (s *AdminService) ListTools(page, size int) ([]dto.ToolDTO, int, error) {
	tools, total, err := s.toolRepo.ListAll(page, size)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]dto.ToolDTO, len(tools))
	for i, t := range tools {
		rules, _ := t.ParseParams()
		if rules == nil {
			rules = []entity.ParamRule{}
		}
		dtos[i] = dto.ToolDTO{
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
			RetryConfig: t.RetryConfig,
			Status:      t.Status,
			CreatedAt:   t.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   t.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return dtos, total, nil
}

// CreateTool 新建工具
func (s *AdminService) CreateTool(req dto.ToolDTO) (*dto.ToolDTO, error) {
	paramsJSON, _ := json.Marshal(req.Params)
	raw := json.RawMessage(paramsJSON)

	tool := &entity.Tool{
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Title:       req.Title,
		Description: req.Description,
		HTTPMethod:  req.HTTPMethod,
		URLTemplate: req.URLTemplate,
		TimeoutMs:   req.TimeoutMs,
		Params:      &raw,
		RetryConfig: req.RetryConfig,
		Status:      req.Status,
	}

	if err := s.toolRepo.Create(tool); err != nil {
		return nil, err
	}

	req.ToolID = tool.ToolID
	return &req, nil
}

// UpdateTool 更新工具
// 仅更新显式传入的非零值字段，支持 partial update
func (s *AdminService) UpdateTool(toolID int64, req dto.ToolDTO) (*dto.ToolDTO, error) {
	existing, err := s.toolRepo.FindByID(toolID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("工具不存在: %d", toolID)
	}

	// 仅覆盖非零值字段（Status 特殊处理：始终更新，因为 0=禁用 是有效值）
	if req.ProjectID != "" {
		existing.ProjectID = req.ProjectID
	}
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Title != "" {
		existing.Title = req.Title
	}
	existing.Description = req.Description
	if req.HTTPMethod != "" {
		existing.HTTPMethod = req.HTTPMethod
	}
	if req.URLTemplate != "" {
		existing.URLTemplate = req.URLTemplate
	}
	if req.TimeoutMs != 0 {
		existing.TimeoutMs = req.TimeoutMs
	}
	if len(req.Params) > 0 {
		paramsJSON, _ := json.Marshal(req.Params)
		raw := json.RawMessage(paramsJSON)
		existing.Params = &raw
	}
	// RetryConfig：nil 表示不修改，非 nil 直接覆盖（允许设为 nil 关闭重试）
	existing.RetryConfig = req.RetryConfig
	// Status 始终应用（0=禁用 是合法值）
	existing.Status = req.Status

	if err := s.toolRepo.Update(existing); err != nil {
		return nil, err
	}

	req.ToolID = toolID
	return &req, nil
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
func (s *AdminService) ListClients(page, size int) ([]dto.ClientDTO, int, error) {
	clients, toolCounts, total, err := s.clientRepo.ListAll(page, size)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]dto.ClientDTO, len(clients))
	for i, c := range clients {
		dtos[i] = dto.ClientDTO{
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
func (s *AdminService) CreateClient(req dto.ClientDTO) (*dto.ClientDTO, error) {
	ctx := context.Background()
	client := &entity.Client{
		ClientID:     req.ClientID,
		Name:         req.Name,
		APIKeyPrefix: "",
		APIKeyHash:   "",
		Description:  req.Description,
		Status:       req.Status,
	}
	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, err
	}
	req.APIKeyPrefix = ""
	req.ToolCount = 0
	return &req, nil
}

// UpdateClient 更新客户端
func (s *AdminService) UpdateClient(clientID string, req dto.ClientDTO) (*dto.ClientDTO, error) {
	ctx := context.Background()
	existing, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("客户端不存在: %s", clientID)
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Status = req.Status

	if err := s.clientRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	req.ClientID = clientID
	req.APIKeyPrefix = existing.APIKeyPrefix
	return &req, nil
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
