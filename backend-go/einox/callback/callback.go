// Package callback 提供回调处理组件
package callback

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/schema"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CallbackInput 鍥炶皟杈撳叆
type CallbackInput struct {
	Messages []*schema.Message
}

// CallbackOutput 鍥炶皟杈撳嚭
type CallbackOutput struct {
	Message    *schema.Message
	TokenUsage *schema.TokenUsage
}

// RunInfo 杩愯淇℃伅
type RunInfo struct {
	Name string
}

// LoggingCallback 日志与追踪回调
type LoggingCallback struct {
	Verbose  bool
	Tracer   trace.Tracer
	Langfuse *LangfuseCallback
}

// OnStart 开始时回调
func (c *LoggingCallback) OnStart(ctx context.Context, info *RunInfo, input *CallbackInput) context.Context {
	if c.Verbose {
		log.Printf("[Callback] 开始调用: %s", info.Name)
		if input != nil {
			log.Printf("  消息数量: %d", len(input.Messages))
		}
	}

	// 开启 OpenTelemetry Span
	if c.Tracer != nil {
		var span trace.Span
		ctx, span = c.Tracer.Start(ctx, info.Name)
		if input != nil {
			span.SetAttributes(
				attribute.String("llm.name", info.Name),
				attribute.Int("llm.input_messages_count", len(input.Messages)),
			)
			// 记录最近一条 User 消息
			if len(input.Messages) > 0 {
				lastMsg := input.Messages[len(input.Messages)-1]
				span.SetAttributes(attribute.String("llm.last_user_query", truncate(lastMsg.Content, 200)))
			}
		}
		ctx = context.WithValue(ctx, "active_span", span)
	}

	// Langfuse 开始记录
	if c.Langfuse != nil {
		ctx = c.Langfuse.OnStart(ctx, info, input)
	}

	return context.WithValue(ctx, "start_time", time.Now())
}

// OnEnd 结束时回调
func (c *LoggingCallback) OnEnd(ctx context.Context, info *RunInfo, output *CallbackOutput) context.Context {
	startTime, _ := ctx.Value("start_time").(time.Time)
	duration := time.Since(startTime)

	span, _ := ctx.Value("active_span").(trace.Span)

	if c.Verbose {
		log.Printf("[Callback] 模型调用完成")
		log.Printf("  耗时: %v", duration)
		if output != nil && output.Message != nil {
			log.Printf("  响应: %s", truncate(output.Message.Content, 100))
		}
	}

	if span != nil {
		if output != nil {
			if output.TokenUsage != nil {
				span.SetAttributes(
					attribute.Int("llm.usage.prompt_tokens", output.TokenUsage.PromptTokens),
					attribute.Int("llm.usage.completion_tokens", output.TokenUsage.CompletionTokens),
					attribute.Int("llm.usage.total_tokens", output.TokenUsage.TotalTokens),
				)
			}
			if output.Message != nil {
				span.SetAttributes(attribute.String("llm.output", truncate(output.Message.Content, 500)))
			}
		}
		span.End()
	}

	// Langfuse 结束记录
	if c.Langfuse != nil {
		ctx = c.Langfuse.OnEnd(ctx, info, output)
	}

	return ctx
}

// OnError 閿欒鏃跺洖璋?
func (c *LoggingCallback) OnError(ctx context.Context, info *RunInfo, err error) context.Context {
	log.Printf("[Callback] 妯″瀷璋冪敤閿欒: %v", err)
	return ctx
}

// MetricsCallback 鎸囨爣鏀堕泦鍥炶皟
type MetricsCallback struct {
	TotalCalls     int
	TotalTokens    int
	TotalLatencyMs int64
	SuccessCount   int
	ErrorCount     int
}

func (c *MetricsCallback) OnStart(ctx context.Context, info *RunInfo, input *CallbackInput) context.Context {
	c.TotalCalls++
	return context.WithValue(ctx, "metrics_start", time.Now())
}

func (c *MetricsCallback) OnEnd(ctx context.Context, info *RunInfo, output *CallbackOutput) context.Context {
	c.SuccessCount++
	if startTime, ok := ctx.Value("metrics_start").(time.Time); ok {
		c.TotalLatencyMs += time.Since(startTime).Milliseconds()
	}
	if output.TokenUsage != nil {
		c.TotalTokens += output.TokenUsage.TotalTokens
	}
	return ctx
}

func (c *MetricsCallback) OnError(ctx context.Context, info *RunInfo, err error) context.Context {
	c.ErrorCount++
	return ctx
}

func (c *MetricsCallback) Report() string {
	avgLatency := float64(0)
	if c.TotalCalls > 0 {
		avgLatency = float64(c.TotalLatencyMs) / float64(c.TotalCalls)
	}
	return fmt.Sprintf(
		"璋冪敤缁熻:\n  鎬昏皟鐢? %d\n  鎴愬姛: %d\n  澶辫触: %d\n  鎬籘oken: %d\n  骞冲潎寤惰繜: %.2fms",
		c.TotalCalls, c.SuccessCount, c.ErrorCount, c.TotalTokens, avgLatency,
	)
}

// NewModelCallbackHandler 创建模型回调处理器
func NewModelCallbackHandler(verbose bool, langfuse *LangfuseCallback) *LoggingCallback {
	return &LoggingCallback{
		Verbose:  verbose,
		Tracer:   otel.Tracer("einox-callback"),
		Langfuse: langfuse,
	}
}

// StreamCallback 娴佸紡鍥炶皟
type StreamCallback struct {
	OnChunk func(content string)
	OnDone  func(fullContent string)
}

// CollectStream 鏀堕泦娴佸紡鍝嶅簲
func CollectStream(stream *schema.StreamReader[*schema.Message], cb *StreamCallback) (string, error) {
	var fullContent string

	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		if chunk != nil && chunk.Content != "" {
			fullContent += chunk.Content
			if cb != nil && cb.OnChunk != nil {
				cb.OnChunk(chunk.Content)
			}
		}
	}

	if cb != nil && cb.OnDone != nil {
		cb.OnDone(fullContent)
	}

	return fullContent, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
