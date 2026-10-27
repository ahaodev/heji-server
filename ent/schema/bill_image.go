package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

// BillImage holds the schema definition for the 账单图片 entity.
type BillImage struct {
	ent.Schema
}

// Fields of the BillImage.
func (BillImage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string { return xid.New().String() }).
			Unique().
			Immutable().
			Comment("图片ID；客户端上传时可指定"),
		field.String("bill_id").
			NotEmpty().
			Comment("所属账单ID"),
		field.String("key").
			NotEmpty().
			Comment("文件存储 key（磁盘相对路径或对象名）"),
		field.String("ext").
			MaxLen(16).
			Default("").
			Comment("文件扩展名，含点"),
		field.Int64("size").
			Default(0).
			Comment("文件大小（字节）"),
		field.String("md5").
			MaxLen(64).
			Default("").
			Comment("内容校验和"),
	}
}

// Edges of the BillImage.
func (BillImage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("bill", Bill.Type).
			Ref("bill_images").
			Field("bill_id").
			Unique().
			Required().
			Comment("所属账单"),
	}
}

// Indexes of the BillImage.
func (BillImage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("bill_id"),
	}
}
