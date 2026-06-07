package service

import (
	"context"
	"testing"
	"time"

	sessionpkg "github.com/sunpuxi/go-mcp-gateway/internal/application/session"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/entity"
	"github.com/sunpuxi/go-mcp-gateway/internal/domain/repository"
	"github.com/sunpuxi/go-mcp-gateway/internal/interface/dto"
	"github.com/sunpuxi/go-mcp-gateway/pkg/logger"
)

func init() {
	_ = logger.Init(logger.Config{Level: "error", Format: "console"})
}

// ============================================================================
//  Mock Repositories（函数式 mock，遵循 retry_test.go 风格）
// ============================================================================

// ---- ToolAdminRepository mock ----

type mockToolAdminRepo struct {
	listAllFn                     func(page, size int) ([]entity.Tool, int, error)
	findByIDFn                    func(toolID int64) (*entity.Tool, error)
	createFn                      func(tool *entity.Tool) error
	updateFn                      func(tool *entity.Tool) error
	deleteFn                      func(toolID int64) error
	deleteByProjectIDFn           func(projectID string) error
	findToolIDsByProjectIDFn      func(projectID string) ([]int64, error)
	countFn                       func() (int, error)
	listEnabledFn                 func() ([]entity.Tool, error)
	findByNameFn                  func(name string) (*entity.Tool, error)
}

func (m *mockToolAdminRepo) ListAll(page, size int) ([]entity.Tool, int, error) {
	if m.listAllFn != nil {
		return m.listAllFn(page, size)
	}
	return nil, 0, nil
}
func (m *mockToolAdminRepo) FindByID(toolID int64) (*entity.Tool, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(toolID)
	}
	return nil, nil
}
func (m *mockToolAdminRepo) Create(tool *entity.Tool) error {
	if m.createFn != nil {
		return m.createFn(tool)
	}
	return nil
}
func (m *mockToolAdminRepo) Update(tool *entity.Tool) error {
	if m.updateFn != nil {
		return m.updateFn(tool)
	}
	return nil
}
func (m *mockToolAdminRepo) Delete(toolID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(toolID)
	}
	return nil
}
func (m *mockToolAdminRepo) DeleteByProjectID(projectID string) error {
	if m.deleteByProjectIDFn != nil {
		return m.deleteByProjectIDFn(projectID)
	}
	return nil
}
func (m *mockToolAdminRepo) FindToolIDsByProjectID(projectID string) ([]int64, error) {
	if m.findToolIDsByProjectIDFn != nil {
		return m.findToolIDsByProjectIDFn(projectID)
	}
	return nil, nil
}
func (m *mockToolAdminRepo) Count() (int, error) {
	if m.countFn != nil {
		return m.countFn()
	}
	return 0, nil
}
func (m *mockToolAdminRepo) ListEnabled() ([]entity.Tool, error) {
	if m.listEnabledFn != nil {
		return m.listEnabledFn()
	}
	return nil, nil
}
func (m *mockToolAdminRepo) FindByName(name string) (*entity.Tool, error) {
	if m.findByNameFn != nil {
		return m.findByNameFn(name)
	}
	return nil, nil
}

var _ repository.ToolAdminRepository = (*mockToolAdminRepo)(nil)

// ---- ProjectRepository mock ----

type mockProjectRepo struct{}

func (m *mockProjectRepo) List(page, size int) ([]entity.Project, int, error) { return nil, 0, nil }
func (m *mockProjectRepo) FindByID(projectID string) (*entity.Project, error)  { return nil, nil }
func (m *mockProjectRepo) Create(project *entity.Project) error                { return nil }
func (m *mockProjectRepo) Update(project *entity.Project) error                { return nil }
func (m *mockProjectRepo) Delete(projectID string) error                       { return nil }
func (m *mockProjectRepo) Count() (int, error)                                 { return 0, nil }

var _ repository.ProjectRepository = (*mockProjectRepo)(nil)

// ---- ClientRepository mock ----

type mockClientRepo struct{}

func (m *mockClientRepo) FindByAPIKeyHash(ctx context.Context, hash string) (*entity.Client, error) {
	return nil, nil
}
func (m *mockClientRepo) FindPermissions(ctx context.Context, clientID string) ([]string, error) {
	return nil, nil
}
func (m *mockClientRepo) ListAll(page, size int) ([]entity.Client, []int, int, error) {
	return nil, nil, 0, nil
}
func (m *mockClientRepo) FindByID(ctx context.Context, clientID string) (*entity.Client, error) {
	return nil, nil
}
func (m *mockClientRepo) Create(ctx context.Context, client *entity.Client) error { return nil }
func (m *mockClientRepo) Update(ctx context.Context, client *entity.Client) error { return nil }
func (m *mockClientRepo) UpdateAPIKey(ctx context.Context, clientID, hash, prefix string) error {
	return nil
}
func (m *mockClientRepo) Delete(ctx context.Context, clientID string) error { return nil }
func (m *mockClientRepo) Count(ctx context.Context) (int, error)            { return 0, nil }
func (m *mockClientRepo) FindToolPermissions(ctx context.Context, clientID string) ([]int64, error) {
	return nil, nil
}
func (m *mockClientRepo) SavePermissions(ctx context.Context, clientID string, toolIDs []int64) error {
	return nil
}
func (m *mockClientRepo) DeletePermissionsByToolIDs(ctx context.Context, toolIDs []int64) error {
	return nil
}

var _ repository.ClientRepository = (*mockClientRepo)(nil)

// ============================================================================
//  辅助函数
// ============================================================================

// newTestSessionManager 创建一个带 SSE 通道的会话管理器，返回 manager + 通道
func newTestSessionManager() (*sessionpkg.Manager, chan []byte) {
	sm := sessionpkg.NewManager(0)
	ch := make(chan []byte, 16)
	s := sm.Create("test-client", []string{"tool1"})
	s.SSECh = ch
	return sm, ch
}

// expectNotification 从 SSE 通道中读取并验证是否为 tools/list_changed 通知
func expectNotification(t *testing.T, ch chan []byte) {
	t.Helper()
	select {
	case msg := <-ch:
		want := `{"jsonrpc":"2.0","method":"notifications/tools/list_changed"}`
		if string(msg) != want {
			t.Errorf("notification = %s, want %s", msg, want)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected notification on SSE channel, got none")
	}
}

// ============================================================================
//  测试：CreateTool → 发送通知
// ============================================================================

func TestAdminService_CreateTool_NotifiesClients(t *testing.T) {
	toolRepo := &mockToolAdminRepo{
		createFn: func(tool *entity.Tool) error {
			tool.ToolID = 1 // 模拟数据库自增 ID
			return nil
		},
	}
	projectRepo := &mockProjectRepo{}
	clientRepo := &mockClientRepo{}
	sm, ch := newTestSessionManager()

	svc := NewAdminService(projectRepo, toolRepo, clientRepo, sm)

	req := dto.ToolDTO{
		ProjectID:   "proj-1",
		Name:        "test_tool",
		Title:       "Test Tool",
		Description: "A test tool",
		HTTPMethod:  "GET",
		URLTemplate: "/api/test",
		TimeoutMs:   5000,
		Params:      []entity.ParamRule{},
		Status:      1,
	}

	result, err := svc.CreateTool(req)
	if err != nil {
		t.Fatalf("CreateTool error: %v", err)
	}
	if result.ToolID != 1 {
		t.Errorf("ToolID = %d, want 1", result.ToolID)
	}

	expectNotification(t, ch)
}

// ============================================================================
//  测试：UpdateTool → 发送通知
// ============================================================================

func TestAdminService_UpdateTool_NotifiesClients(t *testing.T) {
	toolRepo := &mockToolAdminRepo{
		findByIDFn: func(toolID int64) (*entity.Tool, error) {
			return &entity.Tool{
				ToolID:      toolID,
				ProjectID:   "proj-1",
				Name:        "old_name",
				Title:       "Old Title",
				Description: "Old desc",
				HTTPMethod:  "GET",
				URLTemplate: "/api/old",
				TimeoutMs:   3000,
				Status:      1,
			}, nil
		},
		updateFn: func(tool *entity.Tool) error {
			return nil
		},
	}
	projectRepo := &mockProjectRepo{}
	clientRepo := &mockClientRepo{}
	sm, ch := newTestSessionManager()

	svc := NewAdminService(projectRepo, toolRepo, clientRepo, sm)

	req := dto.ToolDTO{
		Name: "new_name",
	}

	result, err := svc.UpdateTool(1, req)
	if err != nil {
		t.Fatalf("UpdateTool error: %v", err)
	}
	if result.Name != "new_name" {
		t.Errorf("Name = %s, want new_name", result.Name)
	}

	expectNotification(t, ch)
}

// ============================================================================
//  测试：DeleteTool → 发送通知
// ============================================================================

func TestAdminService_DeleteTool_NotifiesClients(t *testing.T) {
	toolRepo := &mockToolAdminRepo{
		deleteFn: func(toolID int64) error {
			return nil
		},
	}
	projectRepo := &mockProjectRepo{}
	clientRepo := &mockClientRepo{}
	sm, ch := newTestSessionManager()

	svc := NewAdminService(projectRepo, toolRepo, clientRepo, sm)

	err := svc.DeleteTool(1)
	if err != nil {
		t.Fatalf("DeleteTool error: %v", err)
	}

	expectNotification(t, ch)
}

// ============================================================================
//  测试：通知发送给多个会话
// ============================================================================

func TestAdminService_ToolChangeNotifiesMultipleSessions(t *testing.T) {
	toolRepo := &mockToolAdminRepo{
		createFn: func(tool *entity.Tool) error {
			tool.ToolID = 1
			return nil
		},
	}
	projectRepo := &mockProjectRepo{}
	clientRepo := &mockClientRepo{}
	sm := sessionpkg.NewManager(0)

	// 创建多个会话，各有独立 SSE 通道
	ch1 := make(chan []byte, 16)
	s1 := sm.Create("client-1", []string{"tool1"})
	s1.SSECh = ch1

	ch2 := make(chan []byte, 16)
	s2 := sm.Create("client-2", []string{"tool1", "tool2"})
	s2.SSECh = ch2

	// s3 无 SSECh（模拟未完成 SSE 握手的会话）
	sm.Create("client-3", []string{"tool3"})

	svc := NewAdminService(projectRepo, toolRepo, clientRepo, sm)

	_, err := svc.CreateTool(dto.ToolDTO{
		ProjectID:   "proj-1",
		Name:        "broadcast_test",
		HTTPMethod:  "GET",
		URLTemplate: "/api/test",
		TimeoutMs:   5000,
		Params:      []entity.ParamRule{},
		Status:      1,
	})
	if err != nil {
		t.Fatalf("CreateTool error: %v", err)
	}

	expectNotification(t, ch1)
	expectNotification(t, ch2)
	// s3 的 SSECh 为 nil，不会收到消息（不 panic 即通过）
}

// ============================================================================
//  测试：无活跃会话时不 panic
// ============================================================================

func TestAdminService_ToolChangeNoSessions(t *testing.T) {
	toolRepo := &mockToolAdminRepo{
		createFn: func(tool *entity.Tool) error {
			tool.ToolID = 1
			return nil
		},
	}
	projectRepo := &mockProjectRepo{}
	clientRepo := &mockClientRepo{}
	sm := sessionpkg.NewManager(0) // 无任何会话

	svc := NewAdminService(projectRepo, toolRepo, clientRepo, sm)

	_, err := svc.CreateTool(dto.ToolDTO{
		ProjectID:   "proj-1",
		Name:        "no_session_test",
		HTTPMethod:  "GET",
		URLTemplate: "/api/test",
		TimeoutMs:   5000,
		Params:      []entity.ParamRule{},
		Status:      1,
	})
	if err != nil {
		t.Fatalf("CreateTool error: %v", err)
	}
	// 不 panic 即通过
}
