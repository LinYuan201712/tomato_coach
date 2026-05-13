# Multi-Signal Architecture 手动测试指南

## 🧪 测试概览

本指南提供快速验证架构功能的方法。

---

## 方式1: 使用 Go 单元测试（推荐）

### Step 1: 创建测试文件

```bash
touch backend-go/tests/agent/signal_test.go
touch backend-go/tests/agent/executor_test.go
touch backend-go/tests/agent/context_test.go
```

### Step 2: 基础单元测试

在 `backend-go/tests/agent/signal_test.go` 中添加：

```go
package tests

import (
    "context"
    "testing"
    
    multi_signal "github.com/tomato/backend/einox/agent/multi_signal"
    signals "github.com/tomato/backend/einox/agent/signals"
)

// Mock 对象
type MockTaskRepository struct{}

func (m *MockTaskRepository) GetUserTasks(ctx context.Context, userID string) ([]interface{}, error) {
    return []interface{}{
        map[string]interface{}{
            "id": "task1",
            "title": "背诵第5课",
            "status": "pending",
        },
    }, nil
}

func (m *MockTaskRepository) CreateTask(ctx context.Context, userID string, task interface{}) error {
    return nil
}

func (m *MockTaskRepository) UpdateTask(ctx context.Context, userID string, taskID string, task interface{}) error {
    return nil
}

func (m *MockTaskRepository) DeleteTask(ctx context.Context, userID string, taskID string) error {
    return nil
}

// 测试1: TaskSignalAnalyzer 关键字匹配
func TestTaskSignalAnalyzer_KeywordMatch(t *testing.T) {
    mockRepo := &MockTaskRepository{}
    analyzer := signals.NewTaskSignalAnalyzer(mockRepo)
    
    tests := []struct {
        name              string
        query             string
        expectedConfidence float32
        minConfidence     float32
    }{
        {
            name:              "simple_task_keyword",
            query:             "我有什么任务",
            expectedConfidence: 0.3,
            minConfidence:     0.3,
        },
        {
            name:              "multiple_task_keywords",
            query:             "取消任务，延期截止日期",
            expectedConfidence: 0.6,
            minConfidence:     0.6,
        },
        {
            name:              "complete_and_finish",
            query:             "完成任务并检查进度",
            expectedConfidence: 0.6,
            minConfidence:     0.6,
        },
        {
            name:              "no_task_context",
            query:             "今天天气真好",
            expectedConfidence: 0.0,
            minConfidence:     0.0,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            signal, err := analyzer.Analyze(ctx, multi_signal.AnalysisInput{
                Query:  tt.query,
                UserID: "test_user_1",
            })
            
            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }
            
            if signal.ID != "task" {
                t.Errorf("Expected signal ID 'task', got '%s'", signal.ID)
            }
            
            if signal.Confidence < tt.minConfidence {
                t.Errorf("Expected confidence >= %f, got %f", tt.minConfidence, signal.Confidence)
            }
        })
    }
}

// 测试2: EmotionSignalAnalyzer 情感检测
func TestEmotionSignalAnalyzer_Detection(t *testing.T) {
    analyzer := signals.NewEmotionSignalAnalyzer(nil)
    
    tests := []struct {
        name          string
        query         string
        expectedEmoji []string
    }{
        {
            name:  "positive_emotion",
            query: "我今天开心极了",
        },
        {
            name:  "negative_emotion",
            query: "我心情很差，累死了",
        },
        {
            name:  "anxiety",
            query: "我很焦虑，很担心这次考试",
        },
        {
            name:  "mixed_emotions",
            query: "我又开心又紧张",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            signal, err := analyzer.Analyze(ctx, multi_signal.AnalysisInput{
                Query:  tt.query,
                UserID: "test_user_1",
            })
            
            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }
            
            if signal.ID != "emotion" {
                t.Errorf("Expected signal ID 'emotion', got '%s'", signal.ID)
            }
            
            // 应该有检测到的情感
            emotions, ok := signal.Data["detected_emotions"]
            if !ok || emotions == nil {
                t.Error("Expected detected_emotions in signal data")
            }
        })
    }
}

// 测试3: 多信号并行激活
func TestMultipleSignalsActivation(t *testing.T) {
    registry := multi_signal.NewSignalRegistry()
    mockRepo := &MockTaskRepository{}
    
    // 注册所有分析器
    registry.Register(signals.NewTaskSignalAnalyzer(mockRepo))
    registry.Register(signals.NewEmotionSignalAnalyzer(nil))
    registry.Register(signals.NewStudySignalAnalyzer())
    
    ctx := context.Background()
    
    // 多意图输入 - 这是关键测试场景
    activatedSignals, err := registry.AnalyzeAll(ctx, multi_signal.AnalysisInput{
        Query:  "我心情差，不想背课文，帮我取消任务吧",
        UserID: "test_user_1",
    })
    
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if len(activatedSignals) < 2 {
        t.Errorf("Expected at least 2 signals activated, got %d", len(activatedSignals))
        for _, s := range activatedSignals {
            t.Logf("  Signal: %s (confidence: %.2f)", s.ID, s.Confidence)
        }
    }
    
    // 验证具体的信号被激活
    signalMap := make(map[string]*multi_signal.Signal)
    for _, s := range activatedSignals {
        signalMap[s.ID] = s
    }
    
    if _, hasEmotion := signalMap["emotion"]; !hasEmotion {
        t.Error("Expected emotion signal to be activated")
    }
    
    if _, hasTask := signalMap["task"]; !hasTask {
        t.Error("Expected task signal to be activated")
    }
    
    // 打印信息便于调试
    t.Logf("Successfully activated %d signals", len(activatedSignals))
    for _, s := range activatedSignals {
        t.Logf("  - %s (confidence: %.2f)", s.ID, s.Confidence)
    }
}

// 测试4: Signal 的触发技能
func TestSignalTriggeredSkills(t *testing.T) {
    mockRepo := &MockTaskRepository{}
    taskAnalyzer := signals.NewTaskSignalAnalyzer(mockRepo)
    
    ctx := context.Background()
    signal, _ := taskAnalyzer.Analyze(ctx, multi_signal.AnalysisInput{
        Query:  "取消我的任务",
        UserID: "test_user_1",
    })
    
    skills := taskAnalyzer.GetTriggeredSkills(signal)
    
    if len(skills) == 0 {
        t.Error("Expected triggered skills, got none")
    }
    
    if skills[0] != "task_skill" {
        t.Errorf("Expected task_skill, got %s", skills[0])
    }
    
    t.Logf("Triggered skills: %v", skills)
}

// 性能测试：并行 Signal 分析速度
func BenchmarkSignalAnalysis(b *testing.B) {
    registry := multi_signal.NewSignalRegistry()
    mockRepo := &MockTaskRepository{}
    
    registry.Register(signals.NewTaskSignalAnalyzer(mockRepo))
    registry.Register(signals.NewEmotionSignalAnalyzer(nil))
    registry.Register(signals.NewStudySignalAnalyzer())
    
    ctx := context.Background()
    input := multi_signal.AnalysisInput{
        Query:  "我心情差，不想背课文，帮我取消任务吧",
        UserID: "test_user_1",
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, _ = registry.AnalyzeAll(ctx, input)
    }
}
```

### Step 3: 运行测试

```bash
cd backend-go

# 运行所有测试
go test ./tests/agent/... -v

# 运行单个测试
go test ./tests/agent/ -v -run TestTaskSignalAnalyzer_KeywordMatch

# 运行性能测试
go test ./tests/agent/ -bench=BenchmarkSignalAnalysis -benchmem

# 显示覆盖率
go test ./tests/agent/... -cover
```

### 预期输出

```
=== RUN   TestTaskSignalAnalyzer_KeywordMatch/simple_task_keyword
--- PASS: TestTaskSignalAnalyzer_KeywordMatch/simple_task_keyword (0.01s)

=== RUN   TestEmotionSignalAnalyzer_Detection/positive_emotion
--- PASS: TestEmotionSignalAnalyzer_Detection/positive_emotion (0.01s)

=== RUN   TestMultipleSignalsActivation
    signal_test.go:112: Successfully activated 3 signals
    signal_test.go:113:   - emotion (confidence: 0.70)
    signal_test.go:113:   - task (confidence: 0.60)
    signal_test.go:113:   - study (confidence: 0.25)
--- PASS: TestMultipleSignalsActivation (0.02s)

ok  	github.com/tomato/backend/tests/agent	0.05s	coverage: 85.2%
```

---

## 方式2: 使用 Go CLI 快速测试脚本

### 创建临时测试脚本

在 `backend-go/test_manual.go` 中：

```go
package main

import (
	"context"
	"fmt"
	"log"

	multi_signal "github.com/tomato/backend/einox/agent/multi_signal"
	signals "github.com/tomato/backend/einox/agent/signals"
)

// Mock 实现
type MockTaskRepo struct{}

func (m *MockTaskRepo) GetUserTasks(ctx context.Context, userID string) ([]interface{}, error) {
	return []interface{}{
		map[string]interface{}{
			"id":     "task-001",
			"title":  "背诵第5课",
			"status": "pending",
		},
		map[string]interface{}{
			"id":     "task-002",
			"title":  "完成作业",
			"status": "in_progress",
		},
	}, nil
}

func (m *MockTaskRepo) CreateTask(ctx context.Context, userID string, task interface{}) error {
	return nil
}

func (m *MockTaskRepo) UpdateTask(ctx context.Context, userID string, taskID string, task interface{}) error {
	return nil
}

func (m *MockTaskRepo) DeleteTask(ctx context.Context, userID string, taskID string) error {
	return nil
}

func testSignalAnalysis() {
	fmt.Println("\n" + "="*60)
	fmt.Println("🧪 Multi-Signal Architecture 手动测试")
	fmt.Println("="*60)

	// 初始化
	registry := multi_signal.NewSignalRegistry()
	mockRepo := &MockTaskRepo{}

	registry.Register(signals.NewTaskSignalAnalyzer(mockRepo))
	registry.Register(signals.NewEmotionSignalAnalyzer(nil))
	registry.Register(signals.NewStudySignalAnalyzer())

	ctx := context.Background()

	// 测试用例
	testCases := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "纯任务意图",
			query:    "我有什么任务",
			expected: []string{"task"},
		},
		{
			name:     "纯情感意图",
			query:    "我今天心情很差",
			expected: []string{"emotion"},
		},
		{
			name:     "纯学习意图",
			query:    "帮我推荐一个学习计划",
			expected: []string{"study"},
		},
		{
			name:     "多意图：情感 + 任务",
			query:    "我心情差，不想做任务",
			expected: []string{"emotion", "task"},
		},
		{
			name:     "多意图：情感 + 学习 + 任务",
			query:    "我心情差，不想背课文，帮我取消任务吧",
			expected: []string{"emotion", "study", "task"},
		},
		{
			name:     "无关内容",
			query:    "今天天气真好",
			expected: []string{},
		},
	}

	// 运行每个测试
	for i, tc := range testCases {
		fmt.Printf("\n\n[测试 %d] %s\n", i+1, tc.name)
		fmt.Printf("📝 输入: %s\n", tc.query)
		fmt.Println("─" * 50)

		signals, err := registry.AnalyzeAll(ctx, multi_signal.AnalysisInput{
			Query:  tc.query,
			UserID: "user_123",
		})

		if err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
			continue
		}

		if len(signals) == 0 {
			if len(tc.expected) == 0 {
				fmt.Println("✅ 通过: 正确识别为无关内容")
			} else {
				fmt.Printf("❌ 失败: 期望 %v，但无信号激活\n", tc.expected)
			}
		} else {
			fmt.Printf("✅ 激活了 %d 个信号:\n", len(signals))
			for _, sig := range signals {
				fmt.Printf("   • %s (置信度: %.1f%%)\n", sig.ID, sig.Confidence*100)

				// 打印信号数据
				if len(sig.Data) > 0 {
					for k, v := range sig.Data {
						fmt.Printf("     - %s: %v\n", k, v)
					}
				}

				// 打印触发的技能
				if len(sig.TriggeredSkills) > 0 {
					fmt.Printf("     - 触发技能: %v\n", sig.TriggeredSkills)
				}
			}

			// 验证期望信号
			signalMap := make(map[string]bool)
			for _, sig := range signals {
				signalMap[sig.ID] = true
			}

			allMatch := true
			for _, exp := range tc.expected {
				if !signalMap[exp] {
					allMatch = false
					fmt.Printf("⚠️  缺少期望的信号: %s\n", exp)
				}
			}

			if allMatch && len(signals) >= len(tc.expected) {
				fmt.Println("✅ 通过: 所有期望的信号都被激活")
			}
		}
	}

	fmt.Println("\n" + "="*60)
	fmt.Println("✨ 测试完成")
	fmt.Println("="*60 + "\n")
}

// 性能测试
func testPerformance() {
	fmt.Println("\n" + "="*60)
	fmt.Println("⚡ 性能测试")
	fmt.Println("="*60)

	registry := multi_signal.NewSignalRegistry()
	mockRepo := &MockTaskRepo{}

	registry.Register(signals.NewTaskSignalAnalyzer(mockRepo))
	registry.Register(signals.NewEmotionSignalAnalyzer(nil))
	registry.Register(signals.NewStudySignalAnalyzer())

	ctx := context.Background()
	input := multi_signal.AnalysisInput{
		Query:  "我心情差，不想背课文，帮我取消任务吧",
		UserID: "user_123",
	}

	// 运行1000次并测时
	iterations := 1000
	fmt.Printf("运行 %d 次信号分析...\n", iterations)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		registry.AnalyzeAll(ctx, input)
	}
	elapsed := time.Since(start)

	avgMs := float64(elapsed.Milliseconds()) / float64(iterations)
	fmt.Printf("\n总耗时: %v\n", elapsed)
	fmt.Printf("平均耗时: %.2f ms\n", avgMs)
	fmt.Printf("吞吐量: %.0f req/s\n", 1000/avgMs)

	fmt.Println("\n" + "="*60 + "\n")
}

func main() {
	testSignalAnalysis()
	testPerformance()
}
```

### 运行脚本

```bash
cd backend-go

# 编译并运行
go run test_manual.go

# 预期输出类似：
# ============================================================
# 🧪 Multi-Signal Architecture 手动测试
# ============================================================
#
# [测试 1] 纯任务意图
# 📝 输入: 我有什么任务
# --------------------------------------------------
# ✅ 激活了 1 个信号:
#    • task (置信度: 30.0%)
#    - 触发技能: [task_skill]
# ✅ 通过: 所有期望的信号都被激活
#
# [测试 5] 多意图：情感 + 学习 + 任务
# 📝 输入: 我心情差，不想背课文，帮我取消任务吧
# --------------------------------------------------
# ✅ 激活了 3 个信号:
#    • emotion (置信度: 70.0%)
#    • task (置信度: 60.0%)
#    • study (置信度: 50.0%)
```

---

## 方式3: 使用 Postman 或 Curl 测试 HTTP API

### Step 1: 启动后端服务

```bash
cd backend-go
go run cmd/main.go

# 预期输出：
# 2026/05/07 10:00:00 [INFO] Server starting on :8091
```

### Step 2: 测试 Chat 端点

**使用 Curl:**

```bash
# 基础测试
curl -X POST http://localhost:8091/api/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "query": "我心情差，不想背课文，帮我取消任务吧",
    "user_id": "user_123",
    "session_id": "session_456"
  }'

# 预期响应：
# {
#   "content": "我理解你现在的感受...",
#   "metadata": {
#     "signals_activated": 3,
#     "skills_used": 2,
#     "tokens_estimated": 1500
#   }
# }
```

**使用 Postman:**

1. 创建新的 POST 请求
2. URL: `http://localhost:8091/api/chat`
3. Headers:
   ```
   Content-Type: application/json
   Authorization: Bearer <your_token>
   ```
4. Body (raw JSON):
   ```json
   {
     "query": "我心情差，不想背课文，帮我取消任务吧",
     "user_id": "user_123",
     "session_id": "session_456"
   }
   ```
5. Click Send

### Step 3: 多个测试场景

```bash
# 场景1: 纯任务
curl -X POST http://localhost:8091/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "列出我的任务", "user_id": "user_123", "session_id": "s_1"}'

# 场景2: 纯情感
curl -X POST http://localhost:8091/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "我今天心情特别差", "user_id": "user_123", "session_id": "s_1"}'

# 场景3: 纯学习
curl -X POST http://localhost:8091/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "推荐一个学习计划", "user_id": "user_123", "session_id": "s_1"}'

# 场景4: 多意图组合
curl -X POST http://localhost:8091/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "我焦虑不安，课文太多了，想延期任务截止日期", "user_id": "user_123", "session_id": "s_1"}'
```

---

## 方式4: 使用 Postman 集合（推荐用于重复测试）

### 创建 Postman 集合

1. **File → New → Collection**
2. 命名: "Multi-Signal Tests"
3. 添加以下请求：

```json
{
  "info": {
    "name": "Multi-Signal Agent Tests",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "单意图 - 任务",
      "request": {
        "method": "POST",
        "header": [
          {"key": "Content-Type", "value": "application/json"}
        ],
        "body": {
          "mode": "raw",
          "raw": "{\"query\": \"我有什么任务\", \"user_id\": \"user_123\", \"session_id\": \"s_1\"}"
        },
        "url": {"raw": "http://localhost:8091/api/chat", "protocol": "http", "host": ["localhost"], "port": ["8091"], "path": ["api", "chat"]}
      }
    },
    {
      "name": "单意图 - 情感",
      "request": {
        "method": "POST",
        "header": [{"key": "Content-Type", "value": "application/json"}],
        "body": {
          "mode": "raw",
          "raw": "{\"query\": \"我今天心情很差\", \"user_id\": \"user_123\", \"session_id\": \"s_1\"}"
        },
        "url": {"raw": "http://localhost:8091/api/chat", "protocol": "http", "host": ["localhost"], "port": ["8091"], "path": ["api", "chat"]}
      }
    },
    {
      "name": "多意图 - 情感+任务+学习",
      "request": {
        "method": "POST",
        "header": [{"key": "Content-Type", "value": "application/json"}],
        "body": {
          "mode": "raw",
          "raw": "{\"query\": \"我心情差，不想背课文，帮我取消任务吧\", \"user_id\": \"user_123\", \"session_id\": \"s_1\"}"
        },
        "url": {"raw": "http://localhost:8091/api/chat", "protocol": "http", "host": ["localhost"], "port": ["8091"], "path": ["api", "chat"]}
      }
    }
  ]
}
```

---

## 测试检查清单

### ✅ 功能测试

- [ ] 单个 Signal 正确识别
- [ ] 多个 Signal 并行激活
- [ ] Signal 置信度计算正确
- [ ] 触发的技能列表正确
- [ ] 无关输入正确识别为无信号

### ✅ 性能测试

- [ ] 单次分析 < 10ms
- [ ] 1000次分析 < 10s
- [ ] 内存占用稳定
- [ ] 缓存命中率 > 30%

### ✅ 错误处理

- [ ] 空输入处理
- [ ] 超长输入处理
- [ ] 并发请求处理
- [ ] 错误恢复

### ✅ 集成测试

- [ ] Signal → Context 转换
- [ ] Context → Skill 选择
- [ ] Skill → LLM 调用
- [ ] 完整端到端流程

---

## 📊 期望的测试结果

### 性能基准

```
Signal 分析:
  ✅ 单次: 3-5ms
  ✅ 1000次: 3-5s
  ✅ 吞吐量: 200-300 req/s

缓存效果:
  ✅ 首次命中: ~5ms
  ✅ 缓存命中: <1ms
  ✅ 命中率: 45%+

Token 成本:
  ✅ 旧架构: 2000 tokens/req
  ✅ 新架构: 1500 tokens/req
  ✅ 节省: 25%
```

### 功能验证

```
多信号激活:
  输入: "我心情差，不想背课文，帮我取消任务吧"
  ✅ emotion signal 激活 (confidence: 0.70)
  ✅ study signal 激活 (confidence: 0.50)
  ✅ task signal 激活 (confidence: 0.60)
  ✅ 3 个技能被选中

容错能力:
  ✅ 单个 Signal 失败不影响整体
  ✅ 自动使用其他 Signal
  ✅ 错误日志记录完整
```

---

## 🐛 调试技巧

### 启用详细日志

```go
// 在初始化时
logger := zap.NewDevelopment()
defer logger.Sync()

logger.Info("Signal analysis started",
    zap.String("query", input.Query),
    zap.String("user_id", input.UserID),
)
```

### 打印 Signal 详情

```go
for _, signal := range signals {
    fmt.Printf("Signal: %s\n", signal.ID)
    fmt.Printf("  Confidence: %f\n", signal.Confidence)
    fmt.Printf("  Data: %#v\n", signal.Data)
    fmt.Printf("  Triggered Skills: %v\n", signal.TriggeredSkills)
}
```

### 监控 Cache 命中

```go
metrics := executor.GetMetrics()
fmt.Printf("Cache Hits: %d\n", metrics.CacheHits)
fmt.Printf("Cache Misses: %d\n", metrics.CacheMisses)
fmt.Printf("Hit Rate: %.1f%%\n", 
    float64(metrics.CacheHits)*100/float64(metrics.CacheHits+metrics.CacheMisses))
```

---

## ✨ 测试完成

恭喜！您已经掌握了手动测试方法。建议按以下顺序：

1. **先用 Go 单元测试** (5分钟) - 验证基础功能
2. **再用脚本测试** (10分钟) - 验证多信号激活
3. **最后用 HTTP 测试** (15分钟) - 验证完整流程

