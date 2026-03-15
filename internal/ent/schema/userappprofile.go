package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type UserAppProfile struct {
	ent.Schema
}

func (UserAppProfile) Fields() []ent.Field {
	return []ent.Field{
		// 记录授权时间
		field.Time("first_authorized_at").Default(time.Now).Immutable(),
		field.Time("last_active_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (UserAppProfile) Edges() []ent.Edge {
	return []ent.Edge{
		// 连结 User 和 App
		edge.From("user", User.Type).Ref("profiles").Unique().Required(),
		edge.From("app", App.Type).Ref("profiles").Unique().Required(),
	}
}
