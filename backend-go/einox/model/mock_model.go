package model

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MockChatModel 用于测试的 Mock 模型
type MockChatModel struct {
	Response         *schema.Message
	ResponseQueue    []*schema.Message
	StreamResponses  []*schema.Message
	BindToolsCalled  bool
	LastMessages     []*schema.Message
	LastToolInfos    []*schema.ToolInfo
	GenerateFunc     func(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error)
	Err              error
}

func (m *MockChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 归一化模拟
	sr, err := m.Stream(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}
	defer sr.Close()

	var fullMsg *schema.Message
	for {
		msg, err := sr.Recv()
		if err != nil {
			break
		}
		if fullMsg == nil {
			fullMsg = msg
		} else {
			fullMsg.Content += msg.Content
		}
	}
	return fullMsg, nil
}

func (m *MockChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	m.LastMessages = messages
	if m.Err != nil {
		return nil, m.Err
	}
	
	sr, sw := schema.Pipe[*schema.Message](10)
	go func() {
		defer sw.Close()
		if len(m.StreamResponses) > 0 {
			for _, resp := range m.StreamResponses {
				sw.Send(resp, nil)
			}
			return
		}

		if len(m.ResponseQueue) > 0 {
			resp := m.ResponseQueue[0]
			m.ResponseQueue = m.ResponseQueue[1:]
			sw.Send(resp, nil)
			return
		}

		if m.GenerateFunc != nil {
			resp, err := m.GenerateFunc(ctx, messages, opts...)
			sw.Send(resp, err)
			return
		}

		if m.Response != nil {
			sw.Send(m.Response, nil)
			return
		}

		sw.Send(&schema.Message{Role: schema.Assistant, Content: "Mock Response"}, nil)
	}()
	return sr, nil
}

func (m *MockChatModel) BindTools(tools []*schema.ToolInfo) error {
	m.BindToolsCalled = true
	m.LastToolInfos = tools
	return nil
}

// 确保 MockChatModel 实现了 ChatInner 接口
var _ ChatInner = (*MockChatModel)(nil)
