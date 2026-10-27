package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

// Bill holds the schema definition for the 账单 entity.
type Bill struct {
	ent.Schema
}

// Fields of the Bill.
func (Bill) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string { return xid.New().String() }).
			Unique().
			Immutable().
			Comment("账单ID；客户端离线创建时可直接指定"),
		field.String("book_id").
			NotEmpty().
			Comment("所属账本ID"),
		field.Int64("money").
			NonNegative().
			Comment("金额（分）"),
		field.Int("type").
			Comment("收入(1) / 支出(-1)"),
		field.String("category").
			MaxLen(100).
			Default("").
			Comment("分类名称"),
		field.String("crt_user").
			NotEmpty().
			Comment("创建人ID"),
		field.String("time").
			MaxLen(32).
			Comment("账单日期：YYYY-MM-DD，或客户端同步时的 YYYY-MM-DD HH:MM:SS"),
		field.String("remark").
			MaxLen(500).
			Default("").
			Comment("备注"),
		field.JSON("images", []string{}).
			Optional().
			Comment("票据图片ID列表"),
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

// Edges of the Bill.
func (Bill) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("book", Book.Type).
			Ref("bills").
			Field("book_id").
			Unique().
			Required().
			Comment("所属账本"),
		edge.To("bill_images", BillImage.Type).Comment("账单关联的图片"),
	}
}

// Indexes of the Bill.
func (Bill) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("book_id"),
		index.Fields("time"),
		index.Fields("type"),
		index.Fields("crt_user"),
		// 增量同步按账本 + updated_at 扫描
		index.Fields("book_id", "updated_at"),
	}
}
