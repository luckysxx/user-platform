package repository

import (
	"context"
	"encoding/json"

	mqoutbox "github.com/luckysxx/common/mq/outbox"
	"github.com/luckysxx/user-platform/internal/ent"
)

type EventOutboxRepository interface{ mqoutbox.Writer }

type eventOutboxRepository struct {
	client *ent.Client
}

func NewEventOutboxRepository(client *ent.Client) EventOutboxRepository {
	return &eventOutboxRepository{client: client}
}

func (r *eventOutboxRepository) Append(ctx context.Context, record *mqoutbox.Record) error {
	// 统一 Record 会直接落到当前纯 CDC 风格的 Outbox 表结构里。
	// 现在出库已经交给 Debezium CDC，应用侧只负责可靠写入，不再维护 Relay 状态机。
	tx := ent.TxFromContext(ctx)
	if tx != nil {
		builder := tx.EventOutbox.Create()
		applyOutboxRecord(builder, record)
		_, err := builder.Save(ctx)
		return err
	}

	builder := r.client.EventOutbox.Create()
	applyOutboxRecord(builder, record)
	_, err := builder.Save(ctx)
	return err
}

func applyOutboxRecord(builder *ent.EventOutboxCreate, record *mqoutbox.Record) {
	if record.AggregateType != "" {
		builder.SetAggregatetype(record.AggregateType)
	}
	if record.AggregateID != "" {
		builder.SetAggregateid(record.AggregateID)
	}
	if record.EventType != "" {
		builder.SetType(record.EventType)
	}
	builder.SetPayload(json.RawMessage(record.Payload))
	if len(record.Headers) > 0 {
		builder.SetHeaders(json.RawMessage(record.Headers))
	}
	if !record.CreatedAt.IsZero() {
		builder.SetCreatedAt(record.CreatedAt)
	}
}
