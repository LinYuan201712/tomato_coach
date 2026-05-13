package bus

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InboundMessage 表示从渠道接收的消息
type InboundMessage struct {
	ID        string
	Content   string
	SenderID  string
	ChatID    string
	Channel   string
	AccountID string
	Timestamp time.Time
	Metadata  map[string]interface{}
	Media     []Media
}

// OutboundMessage 表示发送到渠道的消息
type OutboundMessage struct {
	ID        string
	Channel   string
	AccountID string
	ChatID    string
	Content   string
	ReplyTo   string
	Metadata  map[string]interface{}
	Media     []Media
}

// StreamMessage 表示流式消息片段
type StreamMessage struct {
	Content    string
	IsThinking bool
	IsFinal    bool
	IsComplete bool
	Error      string
}

// Media 表示消息中的媒体文件
type Media struct {
	Type   string
	URL    string
	Base64 string
}

type InboundSubscription struct {
	ID      string
	Channel chan *InboundMessage
	bus     *MessageBus
}

type OutboundSubscription struct {
	ID      string
	Channel chan *OutboundMessage
	bus     *MessageBus
}

// MessageBus 消息总线
type MessageBus struct {
	inboundSubscribers  map[string]*InboundSubscription
	outboundSubscribers map[string]*OutboundSubscription
	mu                  sync.RWMutex
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		inboundSubscribers:  make(map[string]*InboundSubscription),
		outboundSubscribers: make(map[string]*OutboundSubscription),
	}
}

func (b *MessageBus) SubscribeInbound() *InboundSubscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := uuid.New().String()
	sub := &InboundSubscription{
		ID:      id,
		Channel: make(chan *InboundMessage, 100),
		bus:     b,
	}
	b.inboundSubscribers[id] = sub
	return sub
}

func (b *MessageBus) SubscribeOutbound() *OutboundSubscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := uuid.New().String()
	sub := &OutboundSubscription{
		ID:      id,
		Channel: make(chan *OutboundMessage, 100),
		bus:     b,
	}
	b.outboundSubscribers[id] = sub
	return sub
}

func (b *MessageBus) PublishInbound(ctx context.Context, msg *InboundMessage) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.inboundSubscribers {
		select {
		case sub.Channel <- msg:
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

func (b *MessageBus) PublishOutbound(ctx context.Context, msg *OutboundMessage) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.outboundSubscribers {
		select {
		case sub.Channel <- msg:
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

func (b *MessageBus) OutboundCount() int {
	return 0
}

func (s *InboundSubscription) Unsubscribe() {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()
	delete(s.bus.inboundSubscribers, s.ID)
	close(s.Channel)
}

func (s *OutboundSubscription) Unsubscribe() {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()
	delete(s.bus.outboundSubscribers, s.ID)
	close(s.Channel)
}
