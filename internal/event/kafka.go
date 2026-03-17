// internal/event/kafka.go
package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Publisher 接口：定义了业务层能调用哪些发送动作
type Publisher interface {
	PublishUserRegistered(ctx context.Context, userID int64, email string) error
	Close() error
}

type publisher struct {
	logger *zap.Logger
	writer *kafka.Writer
}

// NewKafkaPublisher 初始化 Kafka 生产者
func NewKafkaPublisher(addr string, logger *zap.Logger) *publisher {
	w := &kafka.Writer{
		Addr: kafka.TCP(addr),
		// 这里不写死 Topic，让后续发送不同类型事件时更灵活
		Balancer: &kafka.LeastBytes{},
	}
	return &publisher{writer: w, logger: logger}
}

// PublishUserRegistered 具体发送用户注册事件的方法
func (k *publisher) PublishUserRegistered(ctx context.Context, userID int64, email string) error {
	// 1. 构造标准的消息体 JSON
	msg := map[string]interface{}{
		"event_type": "user_registered",
		"user_id":    userID,
		"email":      email,
		"timestamp":  time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 2. 发送到 user_events 这个频道
	err = k.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: "user_events",
			Key:   []byte(email), // 用邮箱做 Key，确保同一个用户的消息在同一个分区排队
			Value: msgBytes,
		},
	)

	if err != nil {
		k.logger.Error("[Kafka] 投递新用户注册消息失败", zap.Error(err))
		return err
	}

	k.logger.Info("[Kafka] 成功投递新用户注册消息", zap.String("email", email))
	return nil
}

func (k *publisher) Close() error {
	return k.writer.Close()
}
