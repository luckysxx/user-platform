// internal/event/kafka.go
package event

import (
	"context"
	"encoding/json"
	"time"

	mqevents "github.com/luckysxx/common/mq/events"
	mqtopics "github.com/luckysxx/common/mq/topics"
	"github.com/luckysxx/common/trace"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Publisher 接口：定义了业务层能调用哪些发送动作
type Publisher interface {
	PublishUserRegistered(ctx context.Context, userID int64, email string, username string) error
	Close() error
}

type publisher struct {
	logger *zap.Logger
	writer *kafka.Writer
	topic  string
}

// NewKafkaWriter 初始化底层 Kafka Writer（供 OutboxWorker 等组件共享）
func NewKafkaWriter(addr string) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(addr),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: false,
	}
}

// NewKafkaPublisher 初始化 Kafka 生产者
func NewKafkaPublisher(addr string, topic string, logger *zap.Logger) *publisher {
	return &publisher{writer: NewKafkaWriter(addr), logger: logger, topic: topic}
}

// PublishUserRegistered 具体发送用户注册事件的方法
func (k *publisher) PublishUserRegistered(ctx context.Context, userID int64, email string, username string) error {
	// 1. 构造标准的消息体 JSON
	msg := mqevents.UserRegistered{
		Version:   mqevents.UserRegisteredVersion,
		EventType: mqtopics.UserRegistered,
		UserID:    userID,
		Email:     email,
		Username:  username,
		Timestamp: time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	traceID := trace.FromContext(ctx)

	// 2. 发送到相应的频道
	err = k.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: k.topic,
			Key:   []byte(email),
			Value: msgBytes,
			Headers: []kafka.Header{
				{Key: trace.HeaderTraceID, Value: []byte(traceID)},
			},
		},
	)

	if err != nil {
		k.logger.Error("[Kafka] 投递新用户注册消息失败", zap.Error(err), zap.String("trace_id", traceID), zap.String("email", email), zap.String("username", username))
		return err
	}

	k.logger.Info("[Kafka] 成功投递新用户注册消息", zap.String("trace_id", traceID), zap.String("email", email), zap.String("username", username))
	return nil
}

func (k *publisher) Close() error {
	return k.writer.Close()
}
