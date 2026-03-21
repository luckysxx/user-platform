package repository

import (
	"context"

	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/ent/eventoutbox"
)

type EventOutboxRepository interface {
	// SaveEvent 保存事件
	SaveEvent(ctx context.Context, topic string, payload []byte) error
	// GetPendingEvents 获取待处理的事件
	GetPendingEvents(ctx context.Context, limit int) ([]*ent.EventOutbox, error)
	// MarkEventAsSuccess 标记事件为成功
	MarkEventAsSuccess(ctx context.Context, id int) error
	// MarkEventAsFailed 标记事件为失败
	MarkEventAsFailed(ctx context.Context, id int) error
}

type eventOutboxRepository struct {
	client *ent.Client
}

func NewEventOutboxRepository(client *ent.Client) EventOutboxRepository {
	return &eventOutboxRepository{client: client}
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
		Where(eventoutbox.StatusEQ("pending")).
		Limit(limit).
		All(ctx)
}

func (r *eventOutboxRepository) MarkEventAsSuccess(ctx context.Context, id int) error {
	_, err := r.client.EventOutbox.UpdateOneID(id).
		SetStatus("success").
		Save(ctx)
	return err
}

func (r *eventOutboxRepository) MarkEventAsFailed(ctx context.Context, id int) error {
	_, err := r.client.EventOutbox.UpdateOneID(id).
		SetStatus("failed").
		Save(ctx)
	return err
}
