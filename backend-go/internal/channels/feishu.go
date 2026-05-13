package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/tomato/backend/config"
	"github.com/tomato/backend/internal/pkg/bus"
	"github.com/tomato/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// FeishuChannel 飞书通道 - WebSocket 模式
type FeishuChannel struct {
	*BaseChannelImpl
	appID             string
	appSecret         string
	domain            string
	encryptKey        string
	verificationToken string
	dmPolicy          string // DM policy: open, allowlist, closed
	wsClient          *larkws.Client
	eventDispatcher   *dispatcher.EventDispatcher
	httpClient        *lark.Client
	// typing indicator state: messageID -> reactionID mapping
	typingReactions   map[string]string
	typingReactionsMu sync.RWMutex
	// bot open_id for mention checking
	botOpenId string
	cronOutputChatID string // cron output target chat ID
}

// NewFeishuChannel 创建飞书通道
func NewFeishuChannel(accountID string, cfg config.FeishuChannelConfig, bus *bus.MessageBus, logger *logger.Logger) (*FeishuChannel, error) {
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return nil, fmt.Errorf("feishu app_id and app_secret are required")
	}

	// 创建 HTTP client for sending messages
	client := lark.NewClient(
		cfg.AppID,
		cfg.AppSecret,
		lark.WithAppType(larkcore.AppTypeSelfBuilt),
		lark.WithOpenBaseUrl(resolveDomain(cfg.Domain)),
	)

	baseCfg := BaseChannelConfig{
		Enabled:    cfg.Enabled,
		AllowedIDs: cfg.AllowedIDs,
	}

	dmPolicy := cfg.DMPolicy
	if dmPolicy == "" {
		dmPolicy = "allowlist"
	}

	return &FeishuChannel{
		BaseChannelImpl:   NewBaseChannelImpl("feishu", accountID, baseCfg, bus, logger),
		appID:             cfg.AppID,
		appSecret:         cfg.AppSecret,
		domain:            cfg.Domain,
		encryptKey:        cfg.EncryptKey,
		verificationToken: cfg.VerificationToken,
		dmPolicy:          dmPolicy,
		httpClient:        client,
		typingReactions:   make(map[string]string),
		cronOutputChatID:   cfg.CronOutputChatID,
	}, nil
}

// Start 启动飞书通道
func (c *FeishuChannel) Start(ctx context.Context) error {
	if err := c.BaseChannelImpl.Start(ctx); err != nil {
		return err
	}

	c.logger.Info("Starting Feishu channel (WebSocket mode)",
		zap.String("app_id", c.appID),
		zap.String("domain", c.domain))

	// 获取机器人的 open_id
	if err := c.fetchBotOpenId(); err != nil {
		c.logger.Warn("Failed to fetch bot open_id, mention checking will be disabled", zap.Error(err))
	}

	// 创建事件分发器
	c.eventDispatcher = dispatcher.NewEventDispatcher(
		c.verificationToken,
		c.encryptKey,
	)

	// 注册事件处理器
	c.registerEventHandlers(ctx)

	// 创建 WebSocket 客户端
	c.wsClient = larkws.NewClient(
		c.appID,
		c.appSecret,
		larkws.WithEventHandler(c.eventDispatcher),
		larkws.WithDomain(resolveDomain(c.domain)),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)

	// 启动 WebSocket 连接
	go c.startWebSocket(ctx)

	return nil
}

func resolveDomain(domain string) string {
	if domain == "lark" {
		return lark.LarkBaseUrl
	}
	return lark.FeishuBaseUrl
}

func (c *FeishuChannel) fetchBotOpenId() error {
	ctx := context.Background()
	tokenReq := &larkcore.SelfBuiltAppAccessTokenReq{
		AppID:     c.appID,
		AppSecret: c.appSecret,
	}

	tokenResp, err := c.httpClient.GetAppAccessTokenBySelfBuiltApp(ctx, tokenReq)
	if err != nil {
		return fmt.Errorf("failed to get app access token: %w", err)
	}
	if !tokenResp.Success() {
		return fmt.Errorf("app access token error: %s", tokenResp.Msg)
	}

	apiResp, err := c.httpClient.Get(ctx, "/open-apis/bot/v3/info", nil, larkcore.AccessTokenTypeApp)
	if err != nil {
		return fmt.Errorf("failed to fetch bot info: %w", err)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Bot  struct {
			OpenId string `json:"open_id"`
		} `json:"bot"`
	}

	if err := json.Unmarshal(apiResp.RawBody, &result); err != nil {
		return err
	}

	if result.Code == 0 {
		c.botOpenId = result.Bot.OpenId
	}
	return nil
}

func (c *FeishuChannel) registerEventHandlers(ctx context.Context) {
	c.eventDispatcher.OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
		c.handleMessageReceived(ctx, event)
		return nil
	}).OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
		// 忽略已读事件，避免 SDK 报错
		return nil
	})
}

func (c *FeishuChannel) startWebSocket(ctx context.Context) {
	if err := c.wsClient.Start(ctx); err != nil {
		c.logger.Error("Feishu WebSocket error", zap.Error(err))
	}
}

func (c *FeishuChannel) handleMessageReceived(ctx context.Context, event *larkim.P2MessageReceiveV1) {
	if event.Event == nil || event.Event.Message == nil {
		return
	}

	senderID := ""
	if event.Event.Sender.SenderId != nil {
		senderID = getStringPtr(event.Event.Sender.SenderId.OpenId)
	}

	chatID := getStringPtr(event.Event.Message.ChatId)
	messageID := getStringPtr(event.Event.Message.MessageId)
	chatType := getStringPtr(event.Event.Message.ChatType)

	c.logger.Info("[Feishu] Received message", zap.String("chat_id", chatID), zap.String("sender_id", senderID))

	// 检查白名单或策略 (简化)
	if chatType == "p2p" {
		if c.dmPolicy == "closed" {
			return
		}
	} else if chatType == "group" {
		if !c.checkBotMentioned(event.Event.Message) {
			return
		}
	}

	content, media := c.extractMessageContentAndMedia(event.Event.Message)
	if content == "" && len(media) == 0 {
		return
	}

	// 发送 typing
	_ = c.addTypingIndicator(messageID)

	// 路由策略优化：在单聊中，优先使用 open_id 作为标识，确保回复能准确送达个人
	targetChatID := chatID
	if chatType == "p2p" && senderID != "" {
		targetChatID = senderID
	}

	inbound := &bus.InboundMessage{
		ID:        messageID,
		Content:   content,
		SenderID:  senderID,
		ChatID:    targetChatID,
		Channel:   c.Name(),
		AccountID: c.accountID,
		Timestamp: time.Now(),
		Media:     media,
	}

	if err := c.PublishInbound(ctx, inbound); err != nil {
		c.logger.Error("Failed to publish inbound message", zap.Error(err))
		_ = c.removeTypingIndicator(messageID)
	}
}

func (c *FeishuChannel) checkBotMentioned(msg *larkim.EventMessage) bool {
	if c.botOpenId == "" {
		return true // 如果不知道机器人 ID，默认允许
	}
	for _, mention := range msg.Mentions {
		if getStringPtr(mention.Id.OpenId) == c.botOpenId {
			return true
		}
	}
	return false
}

func (c *FeishuChannel) extractMessageContentAndMedia(msg *larkim.EventMessage) (string, []bus.Media) {
	if msg.Content == nil {
		return "", nil
	}

	var content map[string]string
	if err := json.Unmarshal([]byte(*msg.Content), &content); err != nil {
		return "", nil
	}

	msgType := getStringPtr(msg.MessageType)
	switch msgType {
	case "text":
		return content["text"], nil
	case "image":
		return "[图片]", []bus.Media{{Type: "image", URL: "feishu:" + content["image_key"]}}
	}
	return "", nil
}

func (c *FeishuChannel) Send(msg *bus.OutboundMessage) error {
	c.logger.Debug("Feishu sending message", zap.String("chat_id", msg.ChatID))

	receiveIDType := larkim.ReceiveIdTypeChatId
	if strings.HasPrefix(msg.ChatID, "ou_") {
		receiveIDType = larkim.ReceiveIdTypeOpenId
	}

	var err error
	if msg.Content != "" {
		err = c.sendCardMessage(msg, receiveIDType)
	}

	if msg.ReplyTo != "" {
		_ = c.removeTypingIndicator(msg.ReplyTo)
	}

	return err
}

func (c *FeishuChannel) sendTextMessage(msg *bus.OutboundMessage, receiveIDType string) error {
	content := fmt.Sprintf(`{"text":%s}`, jsonEscape(msg.Content))
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIDType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(msg.ChatID).
			MsgType(larkim.MsgTypeText).
			Content(content).
			Build()).
		Build()

	resp, err := c.httpClient.Im.Message.Create(context.Background(), req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("feishu api error: %d %s", resp.Code, resp.Msg)
	}
	return nil
}

func (c *FeishuChannel) sendCardMessage(msg *bus.OutboundMessage, receiveIDType string) error {
	// 选一个好看的页眉颜色
	headerColor := "orange" // 番茄色
	
	cardContent := fmt.Sprintf(`{
		"config": {
			"wide_screen_mode": true
		},
		"header": {
			"template": "%s",
			"title": {
				"content": "🍅 智能学习助手",
				"tag": "plain_text"
			}
		},
		"elements": [
			{
				"tag": "markdown",
				"content": %s
			},
			{
				"tag": "hr"
			},
			{
				"tag": "note",
				"elements": [
					{
						"tag": "plain_text",
						"content": "来自 TomatoStudy 智能桌宠"
					}
				]
			}
		]
	}`, headerColor, jsonEscape(msg.Content))
	
	c.logger.Debug("Sending Feishu card", zap.String("chat_id", msg.ChatID), zap.String("content", cardContent))

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIDType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(msg.ChatID).
			MsgType(larkim.MsgTypeInteractive).
			Content(cardContent).
			Build()).
		Build()

	resp, err := c.httpClient.Im.Message.Create(context.Background(), req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("feishu api error: %d %s", resp.Code, resp.Msg)
	}
	return nil
}

func (c *FeishuChannel) addTypingIndicator(messageID string) error {
	emojiType := "Typing"
	req := larkim.NewCreateMessageReactionReqBuilder().
		MessageId(messageID).
		Body(larkim.NewCreateMessageReactionReqBodyBuilder().
			ReactionType(&larkim.Emoji{EmojiType: &emojiType}).
			Build()).
		Build()

	resp, err := c.httpClient.Im.MessageReaction.Create(context.Background(), req)
	if err != nil {
		return err
	}
	if resp.Success() && resp.Data.ReactionId != nil {
		c.typingReactionsMu.Lock()
		c.typingReactions[messageID] = *resp.Data.ReactionId
		c.typingReactionsMu.Unlock()
	}
	return nil
}

func (c *FeishuChannel) removeTypingIndicator(messageID string) error {
	c.typingReactionsMu.Lock()
	reactionID, ok := c.typingReactions[messageID]
	if !ok {
		c.typingReactionsMu.Unlock()
		return nil
	}
	delete(c.typingReactions, messageID)
	c.typingReactionsMu.Unlock()

	req := larkim.NewDeleteMessageReactionReqBuilder().
		MessageId(messageID).
		ReactionId(reactionID).
		Build()

	_, err := c.httpClient.Im.MessageReaction.Delete(context.Background(), req)
	return err
}

func (c *FeishuChannel) Stop() error {
	return c.BaseChannelImpl.Stop()
}

func getStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
