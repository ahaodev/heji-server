package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

// Book holds the schema definition for the 账本 entity.
type Book struct {
	ent.Schema
}

// Fields of the Book.
func (Book) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string { return xid.New().String() }).
			Unique().
			Immutable().
			Comment("账本ID；客户端离线创建时可直接指定"),
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Comment("账本名称"),
		field.String("type").
			MaxLen(50).
			Default("").
			Comment("账本类型"),
		field.String("banner").
			MaxLen(255).
			Default("").
			Comment("封面图片URL"),
		field.String("crt_user_id").
			NotEmpty().
			Comment("创建人ID"),
		field.Bool("is_initial").
			Default(false).
			Comment("是否为初始账本"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("软删除时间，nil 表示未删除"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

// Edges of the Book.
func (Book) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bills", Bill.Type).Comment("账本下的账单"),
		edge.To("categories", Category.Type).Comment("账本下的分类"),
		edge.From("members", User.Type).Ref("books").Comment("账本成员"),
	}
}

// Indexes of the Book.
func (Book) Indexes() []ent.Index {
	return []ent.Index{
		// 同一创建人下账本名称唯一，避免客户端重复建账本
		index.Fields("name", "crt_user_id").Unique(),
		index.Fields("crt_user_id"),
		// 增量同步按 updated_at 扫描
		index.Fields("updated_at"),
	}
}
