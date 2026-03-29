package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		// 1. 主键：在中台架构里，通常会预留 int64 以便未来接入雪花算法 (Snowflake)
		field.Int64("id").
			Positive(),

		// 2. 邮箱：设置唯一索引，限制最大长度
		field.String("email").
			NotEmpty().
			Unique().
			MaxLen(255).
			Comment("用户邮箱，唯一索引"),

		// 3. 账号：设置唯一索引，限制最大长度
		field.String("username").
			NotEmpty().
			Unique().
			MaxLen(32).
			Comment("用户登录名"),

		// 4. 密码：核心安全字段
		field.String("password").
			NotEmpty().
			Sensitive(). // 加上这个标签后，打印日志或者 JSON 序列化时，这个字段会被自动隐藏，防止密码泄露！
			Comment("加密后的密码哈希"),

		// 5. 状态：用枚举代替
		field.Enum("status").
			Values("active", "banned", "deleted").
			Default("active").
			Comment("用户账号状态"),

		// 6. 创建时间：数据写入时自动生成，且不可篡改
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),

		// 7. 更新时间：每次 Update 操作时，Ent 会自动帮你把这个字段更新为当前时间
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now). // 不需要在业务代码里手动写更新时间逻辑
			Comment("最后更新时间"),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// 1对多：一个 User 可以授权多个 App
		edge.To("profiles", UserAppProfile.Type),
		// 1对1：一个 User 拥有一个 Profile
		edge.To("profile", Profile.Type).Unique(),
	}
}
