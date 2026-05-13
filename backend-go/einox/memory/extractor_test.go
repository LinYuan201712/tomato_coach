package memory

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/tomato/backend/einox/model"
)

func TestFactExtractor_Extract(t *testing.T) {
	// 1. 初始化 Mock，返回 JSON 格式的事实
	mockInner := &model.MockChatModel{
		Response: &schema.Message{
			Role:    schema.Assistant,
			Content: `[{"op": "update", "key": "learning_goal", "value": "学好 Golang"}, {"op": "add", "key": "preferred_style", "value": "硬核深入"}]`,
		},
	}
	chatModel := model.NewChatModelFromInner(mockInner)
	extractor := NewFactExtractor(chatModel)
	
	// 2. 执行提取
	facts, err := extractor.Extract(context.Background(), "我想在三个月内学好 Golang", "太棒了，我会陪你深入钻研。")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	
	// 3. 验证
	if len(facts) != 2 {
		t.Fatalf("Expected 2 facts, got %d", len(facts))
	}
	
	if facts[0].Op != "update" || facts[0].Key != "learning_goal" || facts[0].Value != "学好 Golang" {
		t.Errorf("Unexpected fact 0: %+v", facts[0])
	}
	
	if facts[1].Key != "preferred_style" || facts[1].Value != "硬核深入" {
		t.Errorf("Unexpected fact 1: %+v", facts[1])
	}
}

func TestFactExtractor_InvalidJSON(t *testing.T) {
	mockInner := &model.MockChatModel{
		Response: &schema.Message{
			Role:    schema.Assistant,
			Content: `不是 JSON 格式`,
		},
	}
	chatModel := model.NewChatModelFromInner(mockInner)
	extractor := NewFactExtractor(chatModel)
	
	facts, err := extractor.Extract(context.Background(), "hi", "hello")
	if err != nil {
		t.Fatalf("Expected no error for invalid JSON, got %v", err)
	}
	if len(facts) != 0 {
		t.Errorf("Expected 0 facts for invalid JSON, got %d", len(facts))
	}
}
