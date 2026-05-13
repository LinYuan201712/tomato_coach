package channels

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tomato/backend/internal/pkg/bus"
	"github.com/tomato/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

// Manager 通道管理器
type Manager struct {
	channels             map[string]BaseChannel
	bus                  *bus.MessageBus
	logger               *logger.Logger
	mu                   sync.RWMutex
}

// NewManager 创建通道管理器
func NewManager(bus *bus.MessageBus, logger *logger.Logger) *Manager {
	return &Manager{
		channels: make(map[string]BaseChannel),
		bus:      bus,
		logger:   logger,
	}
}

// Register 注册通道
func (m *Manager) Register(channel BaseChannel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", channel.Name(), channel.AccountID())
	if _, ok := m.channels[key]; ok {
		return fmt.Errorf("channel %s already registered", key)
	}

	m.channels[key] = channel
	m.logger.Info("Channel registered", zap.String("key", key))
	return nil
}

// Unregister 注销并停止通道
func (m *Manager) Unregister(name, accountID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", name, accountID)
	channel, ok := m.channels[key]
	if !ok {
		return nil
	}

	if err := channel.Stop(); err != nil {
		m.logger.Error("Failed to stop channel during unregister", zap.String("key", key), zap.Error(err))
	}

	delete(m.channels, key)
	m.logger.Info("Channel unregistered", zap.String("key", key))
	return nil
}

// Start 启动所有通道
func (m *Manager) Start(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, channel := range m.channels {
		m.logger.Info("Starting channel", zap.String("channel", name))
		if err := channel.Start(ctx); err != nil {
			m.logger.Error("Failed to start channel",
				zap.String("channel", name),
				zap.Error(err),
			)
			continue
		}
	}

	return nil
}

// Stop 停止所有通道
func (m *Manager) Stop() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errors []error
	for name, channel := range m.channels {
		if err := channel.Stop(); err != nil {
			m.logger.Error("Failed to stop channel",
				zap.String("channel", name),
				zap.Error(err),
			)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop some channels: %d errors", len(errors))
	}

	return nil
}

// Get 获取通道
func (m *Manager) Get(name, accountID string) (BaseChannel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", name, accountID)
	channel, ok := m.channels[key]
	return channel, ok
}

// List 列出所有通道名称
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.channels))
	for name := range m.channels {
		names = append(names, name)
	}
	return names
}

// DispatchOutbound 分发出站消息
func (m *Manager) DispatchOutbound(ctx context.Context) error {
	m.logger.Debug(">>> Starting outbound message dispatcher <<<")
	defer m.logger.Debug(">>> Outbound dispatcher exited <<<")

	// 订阅出站消息
	subscription := m.bus.SubscribeOutbound()
	defer subscription.Unsubscribe()

	m.logger.Debug("Subscribed to outbound messages",
		zap.String("subscription_id", subscription.ID))

	busChan := subscription.Channel

	// 定期心跳日志
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Debug("Outbound dispatcher stopped by context")
			return ctx.Err()
		case <-heartbeat.C:
			m.logger.Debug("Outbound dispatcher heartbeat - waiting for messages...")
		case msg, ok := <-busChan:
			if !ok {
				m.logger.Warn("Outbound channel closed, exiting dispatcher")
				return nil
			}
			if msg == nil {
				m.logger.Warn("Received nil message, continuing")
				continue
			}

			m.logger.Debug("Outbound message received",
				zap.String("channel", msg.Channel),
				zap.String("chat_id", msg.ChatID),
				zap.Int("content_length", len(msg.Content)))

			// 如果没有 chat_id，跳过此消息
			if msg.ChatID == "" {
				m.logger.Warn("Outbound message has no chat_id, skipping",
					zap.String("channel", msg.Channel))
				continue
			}

			// 查找对应的通道 (使用 Channel + AccountID 组合键)
			channel, ok := m.Get(msg.Channel, msg.AccountID)
			if !ok {
				m.logger.Warn("Channel not found for outbound message",
					zap.String("channel", msg.Channel),
					zap.String("account_id", msg.AccountID),
				)
				continue
			}

			// 发送消息
			if err := channel.Send(msg); err != nil {
				m.logger.Error("Failed to send message via channel",
					zap.String("channel", msg.Channel),
					zap.Error(err),
				)
			} else {
				m.logger.Debug("Message sent successfully via channel",
					zap.String("channel", msg.Channel),
					zap.String("chat_id", msg.ChatID))
			}
		}
	}
}
