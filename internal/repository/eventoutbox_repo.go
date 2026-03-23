package repository

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/ent/eventoutbox"
	"github.com/redis/go-redis/v9"
)

type EventOutboxRepository interface {
	// SaveEvent 保存事件
	SaveEvent(ctx context.Context, topic string, payload []byte) error
	// GetPendingEvents 获取待处理的事件
	GetPendingEvents(ctx context.Context, limit int) ([]*ent.EventOutbox, error)
	// MarkEventAsSuccess 标记事件为成功
	MarkEventAsSuccess(ctx context.Context, ids []int) error
	// MarkEventAsFailed 标记事件为失败
	MarkEventAsFailed(ctx context.Context, ids []int) error
	// MarkEventRetry 标记事件重试
	MarkEventRetry(ctx context.Context, ids []int) error
	// Notify 发送唤醒信号
	Notify(ctx context.Context) error
}

type eventOutboxRepository struct {
	client      *ent.Client
	redisClient *redis.Client
}

func NewEventOutboxRepository(client *ent.Client, redisClient *redis.Client) EventOutboxRepository {
	return &eventOutboxRepository{
		client:      client,
		redisClient: redisClient,
	}
}

func (r *eventOutboxRepository) SaveEvent(ctx context.Context, topic string, payload []byte) error {
	// 检查是否在事务中
	tx := ent.TxFromContext(ctx)
	if tx != nil {
		_, err := tx.EventOutbox.Create().
			SetTopic(topic).
			SetPayload(payload).
			Save(ctx)
		return err
	}
	// 没有事务, 使用普通单表落库
	_, err := r.client.EventOutbox.Create().
		SetTopic(topic).
		SetPayload(payload).
		Save(ctx)
	return err
}

func (r *eventOutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]*ent.EventOutbox, error) {
	return r.client.EventOutbox.Query().
		Where(
			eventoutbox.StatusEQ("pending"),
			eventoutbox.RetryCountLT(5),
		).
		Limit(limit).
		Order(ent.Asc(eventoutbox.FieldUpdatedAt)).
		ForUpdate(sql.WithLockAction(sql.SkipLocked)).
		All(ctx)
}

func (r *eventOutboxRepository) MarkEventAsSuccess(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.client.EventOutbox.Update().
		Where(eventoutbox.IDIn(ids...)).
		SetStatus("success").
		Save(ctx)
	return err
}

func (r *eventOutboxRepository) MarkEventAsFailed(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.client.EventOutbox.Update().
		Where(eventoutbox.IDIn(ids...)).
		SetStatus("failed").
		Save(ctx)
	return err
}

func (r *eventOutboxRepository) MarkEventRetry(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.client.EventOutbox.Update().
		Where(eventoutbox.IDIn(ids...)).
		AddRetryCount(1).
		Save(ctx)
	return err
}

func (r *eventOutboxRepository) Notify(ctx context.Context) error {
	// LPUSH outbox_wakeup_list 并在 OutboxWorker 中阻塞接收
	return r.redisClient.LPush(ctx, "outbox_wakeup_list", "1").Err()
}
