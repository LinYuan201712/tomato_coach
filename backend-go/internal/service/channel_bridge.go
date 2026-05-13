package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tomato/backend/internal/pkg/bus"
	"github.com/tomato/backend/internal/pkg/logger"
)

// ChannelBridge 连接消息渠道与 AI 教练服务的桥梁
type ChannelBridge struct {
	coachService CoachService
	messageBus   *bus.MessageBus
	logger       *logger.Logger
}

func NewChannelBridge(coachService CoachService, messageBus *bus.MessageBus, logger *logger.Logger) *ChannelBridge {
	return &ChannelBridge{
		coachService: coachService,
		messageBus:   messageBus,
		logger:       logger,
	}
}

// Start 启动桥接服务，监听入站消息
func (b *ChannelBridge) Start(ctx context.Context) error {
	b.logger.Info("ChannelBridge 启动中...")

	// 订阅所有入站消息
	sub := b.messageBus.SubscribeInbound()
	defer sub.Unsubscribe()

	b.logger.Info("ChannelBridge 已订阅入站消息总线")

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("ChannelBridge 收到停止信号")
			return ctx.Err()
		case msg, ok := <-sub.Channel:
			if !ok {
				b.logger.Warn("入站消息通道已关闭")
				return nil
			}
			// 异步处理每条消息
			go b.handleInbound(ctx, msg)
		}
	}
}

func (b *ChannelBridge) handleInbound(ctx context.Context, msg *bus.InboundMessage) {
	// 1. 用户映射逻辑
	// 从 AccountID 中解析出本地 UserID
	userID, _ := strconv.ParseInt(msg.AccountID, 10, 64)
	if userID == 0 {
		userID = 1 // 回退到默认用户
	}

	b.logger.Infof("[Bridge] 收到来自渠道 %s 的消息: sender=%s, content=%s", msg.Channel, msg.SenderID, msg.Content)

	// 2. 调用 CoachService 获取回复
	// 我们使用 ChatStream，但由于渠道通常不支持实时流式显示，我们内部累积后发送
	sessionID := fmt.Sprintf("%s_%s", msg.Channel, msg.ChatID)
	stream, err := b.coachService.ChatStream(ctx, userID, sessionID, msg.Content, true, "fast")
	if err != nil {
		b.logger.Errorf("[Bridge] CoachService 调用失败: %v", err)
		b.sendReply(ctx, msg, "抱歉，我遇到了一点技术问题，请稍后再试。")
		return
	}
	defer stream.Close()

	var fullContent strings.Builder
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		if chunk != nil && chunk.Content != "" {
			fullContent.WriteString(chunk.Content)
		}
	}

	replyContent := fullContent.String()
	if replyContent == "" {
		replyContent = "我听到了你的消息，但我不知道该如何回复。"
	}

	// 3. 发布出站消息
	b.sendReply(ctx, msg, replyContent)
}

func (b *ChannelBridge) sendReply(ctx context.Context, originalMsg *bus.InboundMessage, content string) {
	outbound := &bus.OutboundMessage{
		Channel:   originalMsg.Channel,
		AccountID: originalMsg.AccountID,
		ChatID:    originalMsg.ChatID,
		Content:   content,
		ReplyTo:   originalMsg.ID, // 引用回复
	}

	if err := b.messageBus.PublishOutbound(ctx, outbound); err != nil {
		b.logger.Errorf("[Bridge] 发布出站消息失败: %v", err)
	} else {
		b.logger.Infof("[Bridge] 已发送回复到渠道 %s, 目标 %s", originalMsg.Channel, originalMsg.ChatID)
	}
}
