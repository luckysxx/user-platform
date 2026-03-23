package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type EventOutbox struct {
	ent.Schema
}

func (EventOutbox) Fields() []ent.Field {
	return []ent.Field{
		// 主键自增
		field.String("topic").
			NotEmpty().
			Comment("消息主题"),
		field.Bytes("payload").
			NotEmpty().
			Comment("消息体"),
		field.Enum("status").
			Values("pending", "success", "failed").
			Default("pending").
			Comment("消息状态"),
		field.Int("retry_count").
			Default(0).
			Comment("重试次数"),
		field.Time("created_at").
			Default(time.Now).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

func (EventOutbox) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (EventOutbox) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "retry_count"),
		index.Fields("status", "updated_at"),
	}
}
