package model

import (
	"context"
	"fmt"
	"io"

	"github.com/avast/retry-go/v4"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/sony/gobreaker"
	"github.com/tomato/backend/einox/callback"
)

// RetryAndCircuitBreakerDecorator 聊天模型装饰器，提供重试与熔断功能
type RetryAndCircuitBreakerDecorator struct {
	inner ChatInner
	cb    *gobreaker.CircuitBreaker
}

func NewRetryAndCircuitBreakerDecorator(inner ChatInner, name string) *RetryAndCircuitBreakerDecorator {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: name,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("[CIRCUIT] Circuit breaker %s changed from %s to %s\n", name, from, to)
		},
	})

	return &RetryAndCircuitBreakerDecorator{
		inner: inner,
		cb:    cb,
	}
}

func (d *RetryAndCircuitBreakerDecorator) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	result, err := d.cb.Execute(func() (interface{}, error) {
		var msg *schema.Message
		var innerErr error

		innerErr = retry.Do(
			func() error {
				msg, innerErr = d.inner.Generate(ctx, messages, opts...)
				return innerErr
			},
			retry.Context(ctx),
			retry.Attempts(3),
			retry.DelayType(retry.BackOffDelay),
			// 只有部分错误需要重试，但这依赖于 inner 的实现
			// 装饰器层默认尝试 3 次，除非 inner 返回不可重试错误
		)

		return msg, innerErr
	})

	if err != nil {
		return nil, err
	}

	return result.(*schema.Message), nil
}

func (d *RetryAndCircuitBreakerDecorator) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 对于流式输出，熔断器包裹初始连接过程
	// 在建立流的初始阶段进行重试，提升网络抖动下的成功率
	result, err := d.cb.Execute(func() (interface{}, error) {
		var sr *schema.StreamReader[*schema.Message]
		var innerErr error

		innerErr = retry.Do(
			func() error {
				sr, innerErr = d.inner.Stream(ctx, messages, opts...)
				return innerErr
			},
			retry.Context(ctx),
			retry.Attempts(3),
			retry.DelayType(retry.BackOffDelay),
		)

		return sr, innerErr
	})

	if err != nil {
		return nil, err
	}

	return result.(*schema.StreamReader[*schema.Message]), nil
}

func (d *RetryAndCircuitBreakerDecorator) BindTools(tools []*schema.ToolInfo) error {
	return d.inner.BindTools(tools)
}

// CallbackDecorator 聊天模型装饰器，提供回调追踪功能
type CallbackDecorator struct {
	inner   ChatInner
	name    string
	handler *callback.LoggingCallback
}

func NewCallbackDecorator(inner ChatInner, name string, handler *callback.LoggingCallback) *CallbackDecorator {
	return &CallbackDecorator{
		inner:   inner,
		name:    name,
		handler: handler,
	}
}

func (d *CallbackDecorator) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if d.handler == nil {
		return d.inner.Generate(ctx, messages, opts...)
	}

	info := &callback.RunInfo{Name: d.name}
	input := &callback.CallbackInput{Messages: messages}
	ctx = d.handler.OnStart(ctx, info, input)

	msg, err := d.inner.Generate(ctx, messages, opts...)
	if err != nil {
		d.handler.OnError(ctx, info, err)
		return nil, err
	}

	output := &callback.CallbackOutput{
		Message:    msg,
		TokenUsage: msg.ResponseMeta.Usage,
	}
	d.handler.OnEnd(ctx, info, output)

	return msg, nil
}

func (d *CallbackDecorator) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	if d.handler == nil {
		return d.inner.Stream(ctx, messages, opts...)
	}

	info := &callback.RunInfo{Name: d.name}
	input := &callback.CallbackInput{Messages: messages}
	ctx = d.handler.OnStart(ctx, info, input)

	sr, err := d.inner.Stream(ctx, messages, opts...)
	if err != nil {
		d.handler.OnError(ctx, info, err)
		return nil, err
	}

	// 对于流式输出，我们需要包装 StreamReader 来捕获结束时的 TokenUsage
	resSr, sw := schema.Pipe[*schema.Message](10)
	go func() {
		defer sw.Close()
		defer sr.Close()
		var fullMsg *schema.Message
		for {
			msg, err := sr.Recv()
			if err != nil {
				if err == io.EOF {
					// 结束时上报
					output := &callback.CallbackOutput{
						Message: fullMsg,
					}
					if fullMsg != nil && fullMsg.ResponseMeta != nil {
						output.TokenUsage = fullMsg.ResponseMeta.Usage
					}
					d.handler.OnEnd(ctx, info, output)
					break
				}
				d.handler.OnError(ctx, info, err)
				sw.Send(nil, err)
				return
			}
			sw.Send(msg, nil)

			if fullMsg == nil {
				fullMsg = &schema.Message{Role: msg.Role, Content: msg.Content}
			} else {
				fullMsg.Content += msg.Content
			}
			if msg.ResponseMeta != nil {
				fullMsg.ResponseMeta = msg.ResponseMeta
			}
		}
	}()

	return resSr, nil
}

func (d *CallbackDecorator) BindTools(tools []*schema.ToolInfo) error {
	return d.inner.BindTools(tools)
}
