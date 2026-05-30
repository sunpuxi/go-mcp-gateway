package service

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
)

// AuthService 鉴权领域服务，负责 API Key 校验和权限判定
type AuthService struct {
	clientRepo repository.ClientRepository
}

// NewAuthService 创建鉴权服务
func NewAuthService(clientRepo repository.ClientRepository) *AuthService {
	return &AuthService{clientRepo: clientRepo}
}

// Authenticate 校验 API Key，认证通过后返回 clientID 和该客户端有权限的 tool 名称列表
func (s *AuthService) Authenticate(ctx context.Context, apiKey string) (clientID string, permissions []string, err error) {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(apiKey)))

	client, err := s.clientRepo.FindByAPIKeyHash(ctx, hash)
	if err != nil {
		return "", nil, fmt.Errorf("API Key 认证失败: %w", err)
	}
	if client.Status != 1 {
		return "", nil, fmt.Errorf("客户端已被禁用")
	}

	permissions, err = s.clientRepo.FindPermissions(ctx, client.ClientID)
	if err != nil {
		return "", nil, fmt.Errorf("查询客户端权限失败: %w", err)
	}

	return client.ClientID, permissions, nil
}

// CheckToolPermission 校验某个工具是否在权限列表中
func (s *AuthService) CheckToolPermission(permissions []string, toolName string) error {
	for _, p := range permissions {
		if p == toolName {
			return nil
		}
	}
	return fmt.Errorf("无权调用工具: %s", toolName)
}
