package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"shadmin/domain"
)

// notifyTopic 是账本变更的推送主题：heji/notify/{userId}/{entity}。
const notifyTopic = "heji/notify/%s/%s"

// syncMessage 与客户端 SyncMessage 对应；content 是实体 JSON。
type syncMessage struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	BookID    string `json:"book_id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// MQTTNotifier 把账本变更推送给账本其他成员（不含操作者）。
// MQTT 不可用时业务写入照常成功：推送是尽力而为的通知，不是数据一致性的一部分。
type MQTTNotifier struct {
	client         mqtt.Client
	bookRepo       domain.BookRepository
	publishTimeout time.Duration
}

// NewMQTTNotifier 连接 broker 并返回通知器；地址为空时返回 (nil, nil)，
// 调用方据此跳过推送。连接失败只记录日志，不阻止启动。
func NewMQTTNotifier(address, tcpPort string, bookRepo domain.BookRepository) (*MQTTNotifier, error) {
	if address == "" {
		return nil, nil
	}

	broker := fmt.Sprintf("tcp://%s:%s", address, tcpPort)
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID("shadmin-billing-notifier").
		SetAutoReconnect(true).
		SetMaxReconnectInterval(30 * time.Second).
		SetKeepAlive(60 * time.Second).
		SetConnectTimeout(5 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("connect MQTT broker %s: %w", broker, token.Error())
	}
	slog.Info("MQTT notifier connected", "broker", broker)

	return &MQTTNotifier{
		client:         client,
		bookRepo:       bookRepo,
		publishTimeout: 3 * time.Second,
	}, nil
}

// NotifyBookChange 查询账本成员并逐个推送；异步执行，不阻塞调用方。
func (n *MQTTNotifier) NotifyBookChange(parent context.Context, bookID, actorID, changeType, entity string, payload any) {
	if n == nil || n.client == nil || !n.client.IsConnected() {
		return
	}

	content, err := json.Marshal(payload)
	if err != nil {
		slog.Warn("failed to marshal billing notification", "entity", entity, "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), n.publishTimeout)
	message := syncMessage{
		Type:      changeType,
		BookID:    bookID,
		Content:   string(content),
		Timestamp: time.Now().UnixMilli(),
	}

	go func() {
		defer cancel()
		book, err := n.bookRepo.GetByID(ctx, bookID)
		if err != nil {
			// 账本已删除时取不到成员：按操作者已知信息无法推送，直接跳过
			slog.Debug("skip billing notification, book unavailable", "book_id", bookID, "error", err)
			return
		}
		for _, memberID := range book.Members {
			if memberID == actorID {
				continue
			}
			n.publish(fmt.Sprintf(notifyTopic, memberID, entity), message)
		}
	}()
}

func (n *MQTTNotifier) publish(topic string, message syncMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		slog.Warn("failed to marshal notify message", "topic", topic, "error", err)
		return
	}
	token := n.client.Publish(topic, 1, false, data)
	if !token.WaitTimeout(n.publishTimeout) {
		slog.Warn("MQTT publish timed out", "topic", topic)
		return
	}
	if token.Error() != nil {
		slog.Warn("MQTT publish failed", "topic", topic, "error", token.Error())
	}
}

// Close 断开连接；未连接时为空操作。
func (n *MQTTNotifier) Close() {
	if n == nil || n.client == nil {
		return
	}
	if n.client.IsConnected() {
		n.client.Disconnect(500)
	}
}
