package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

// Category holds the schema definition for the 分类 entity.
type Category struct {
	ent.Schema
}

// Fields of the Category.
func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string { return xid.New().String() }).
			Unique().
			Immutable().
			Comment("分类ID；客户端可指定以保持离线数据一致"),
		field.String("book_id").
			NotEmpty().
			Comment("所属账本ID"),
		field.String("name").
			MaxLen(50).
			Default("其他").
			Comment("分类名称"),
		field.Int("type").
			Comment("收入(1) / 支出(-1)"),
		field.Int("level").
			Default(0).
			Comment("层级"),
		field.Int("index").
			Default(0).
			Comment("排序序号"),
		field.String("parent_id").
			MaxLen(32).
			Default("").
			Comment("父分类ID"),
		field.String("create_user").
			MaxLen(32).
			Default("").
			Comment("创建人ID"),
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

// Edges of the Category.
func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("book", Book.Type).
			Ref("categories").
			Field("book_id").
			Unique().
			Required().
			Comment("所属账本"),
	}
}

// Indexes of the Category.
func (Category) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("book_id"),
		index.Fields("type"),
		// 增量同步按 updated_at 扫描
		index.Fields("updated_at"),
	}
}
