package callback

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/henomis/langfuse-go"
	"github.com/henomis/langfuse-go/model"
)

// LangfuseCallback 实现 Langfuse 可观测性集成
type LangfuseCallback struct {
	Client *langfuse.Langfuse
}

// NewLangfuseCallback 创建 Langfuse 回调处理器
func NewLangfuseCallback(publicKey, secretKey, baseURL string) *LangfuseCallback {
	if publicKey == "" || secretKey == "" {
		return nil
	}

	// henomis/langfuse-go SDK 会自动从环境变量读取配置
	os.Setenv("LANGFUSE_PUBLIC_KEY", publicKey)
	os.Setenv("LANGFUSE_SECRET_KEY", secretKey)
	os.Setenv("LANGFUSE_HOST", baseURL)

	return &LangfuseCallback{
		Client: langfuse.New(context.Background()),
	}
}

// OnStart 在节点开始执行时调用
func (l *LangfuseCallback) OnStart(ctx context.Context, info *RunInfo, input *CallbackInput) context.Context {
	if l.Client == nil {
		return ctx
	}

	// 记录开始时间
	startTime := time.Now()
	ctx = context.WithValue(ctx, "langfuse_start_time", startTime)

	return ctx
}

// OnEnd 在节点执行结束时调用
func (l *LangfuseCallback) OnEnd(ctx context.Context, info *RunInfo, output *CallbackOutput) context.Context {
	if l.Client == nil {
		return ctx
	}

	startTime, _ := ctx.Value("langfuse_start_time").(time.Time)
	duration := time.Since(startTime)

	// 识别是否是 LLM 调用
	isLLM := strings.Contains(info.Name, "LLM") || strings.Contains(info.Name, "Model") || strings.Contains(info.Name, "Agent")

	if isLLM {
		l.recordGeneration(ctx, info, output, startTime, duration)
	} else {
		ctx = l.recordSpan(ctx, info, output, startTime, duration)
	}

	return ctx
}

func (l *LangfuseCallback) recordGeneration(ctx context.Context, info *RunInfo, output *CallbackOutput, start time.Time, duration time.Duration) {
	traceID, _ := ctx.Value("langfuse_trace_id").(string)
	parentID, _ := ctx.Value("langfuse_observation_id").(*string)

	var promptTokens, completionTokens, totalTokens int
	var outputContent string
	if output != nil {
		usage := output.TokenUsage
		if usage != nil {
			promptTokens = usage.PromptTokens
			completionTokens = usage.CompletionTokens
			totalTokens = usage.TotalTokens
		}
		if output.Message != nil {
			outputContent = output.Message.Content
		}
	}

	endTime := start.Add(duration)
	modelName := info.Name
	if strings.Contains(modelName, "Aliyun-") {
		modelName = strings.TrimPrefix(modelName, "Aliyun-")
	}

	g := &model.Generation{
		TraceID:   traceID,
		Name:      info.Name,
		StartTime: &start,
		EndTime:   &endTime,
		Model:     modelName,
		Usage: model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
	}

	if outputContent != "" {
		g.Output = outputContent
	}

	gen, err := l.Client.Generation(g, parentID)
	if err == nil {
		l.Client.GenerationEnd(gen)
	}
}

func (l *LangfuseCallback) recordSpan(ctx context.Context, info *RunInfo, output *CallbackOutput, start time.Time, duration time.Duration) context.Context {
	traceID, _ := ctx.Value("langfuse_trace_id").(string)
	parentID, _ := ctx.Value("langfuse_observation_id").(*string)

	endTime := start.Add(duration)
	s := &model.Span{
		TraceID:   traceID,
		Name:      info.Name,
		StartTime: &start,
		EndTime:   &endTime,
	}
	if output != nil && output.Message != nil {
		s.Output = output.Message.Content
	}

	span, err := l.Client.Span(s, parentID)
	if err == nil {
		l.Client.SpanEnd(span)
		// 关键：将当前 Span 设置为子节点的 ParentID
		ctx = context.WithValue(ctx, "langfuse_observation_id", &span.ID)
	}
	return ctx
}

// SetTraceInfo 在上下文设置 Trace 信息
func SetTraceInfo(ctx context.Context, traceID string, userID string) context.Context {
	ctx = context.WithValue(ctx, "langfuse_trace_id", traceID)
	return ctx
}
