package tests

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	"github.com/tomato/backend/einox/tools"
	"github.com/tomato/backend/internal/domain/entity"
)

// MockUserRepo 模拟用户 Repository
type MockUserRepo struct {
	users map[int64]*entity.User
}

func (m *MockUserRepo) FindByUserID(ctx context.Context, userID int64) (*entity.User, error) {
	return m.users[userID], nil
}

func (m *MockUserRepo) UpdateProfile(ctx context.Context, userID int64, goals string, style string) error {
	user, ok := m.users[userID]
	if !ok {
		user = &entity.User{UserID: userID}
		m.users[userID] = user
	}
	if goals != "" {
		user.Goals = goals
	}
	if style != "" {
		user.PreferredStyle = style
	}
	return nil
}

func TestUserProfileTools(t *testing.T) {
	mockRepo := &MockUserRepo{
		users: map[int64]*entity.User{
			123: {
				UserID:         123,
				Goals:          "Go,React",
				PreferredStyle: "Academic",
				Tomato:         10,
			},
		},
	}

	getTool := tools.NewUserProfilingTool(mockRepo)
	updateTool := tools.NewUpdateProfileTool(mockRepo)

	ctx := context.WithValue(context.Background(), "user_id", int64(123))

	// 1. 测试获取画像
	_, err := getTool.Info(ctx)
	if err != nil {
		t.Fatalf("Failed to get tool info: %v", err)
	}

	// 这里我们需要运行工具，但 getTool 是 tool.BaseTool，我们需要断言它是 tool.InvokableTool
	invokableGet, ok := getTool.(tool.InvokableTool)
	if !ok {
		t.Fatalf("Tool is not invokable")
	}

	resp, err := invokableGet.InvokableRun(ctx, "{}")
	if err != nil {
		t.Fatalf("Tool run failed: %v", err)
	}
	t.Logf("Get Profile Response: %s", resp)

	// 2. 测试更新画像
	invokableUpdate, ok := updateTool.(tool.InvokableTool)
	if !ok {
		t.Fatalf("Update tool is not invokable")
	}

	updateArgs := `{"goal": "Rust,Wasm", "style": "Funny"}`
	_, err = invokableUpdate.InvokableRun(ctx, updateArgs)
	if err != nil {
		t.Fatalf("Update tool run failed: %v", err)
	}

	// 3. 再次获取，验证是否更新
	resp, err = invokableGet.InvokableRun(ctx, "{}")
	if err != nil {
		t.Fatalf("Subsequent get tool run failed: %v", err)
	}
	t.Logf("Updated Profile Response: %s", resp)

	// 验证 mockRepo 状态
	user := mockRepo.users[123]
	if user.Goals != "Rust,Wasm" {
		t.Errorf("Goals not updated correctly: expected 'Rust,Wasm', got '%s'", user.Goals)
	}
	if user.PreferredStyle != "Funny" {
		t.Errorf("Style not updated correctly: expected 'Funny', got '%s'", user.PreferredStyle)
	}
}
