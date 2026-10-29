package repository

import (
	"context"
	"math"
	"time"

	"shadmin/domain"
	"shadmin/ent"
	"shadmin/ent/bill"
	"shadmin/ent/billimage"
	"shadmin/ent/book"
	"shadmin/ent/category"
	"shadmin/ent/predicate"
	"shadmin/ent/user"
)

type entSyncRepository struct {
	client *ent.Client
}

func NewSyncRepository(client *ent.Client) domain.SyncRepository {
	return &entSyncRepository{client: client}
}

// entityPage 是一类实体的单页结果。cursor 是“已完整交付到（不含）”的毫秒边界；
// 调用方取各类 cursor 的最小值作为下一页 since，保证没有实体被跳过。MaxInt64 表示该类已拉完。
type entityPage[T any] struct {
	rows   []*T
	cursor int64
}

// QueryChanges 返回 since 之后有变更的账本、账单、分类与图片（含已软删除记录，客户端据此删除本地行）。
//
// 游标按毫秒推进，而 updated_at 带亚毫秒精度，直接按条截断会让游标卡在原地或跨过未交付的记录：
// 因此翻页以“整毫秒桶”为单位，桶内记录整批交付，游标推进到桶之后。
func (r *entSyncRepository) QueryChanges(ctx context.Context, userID string, sinceMs int64, limit int) (*domain.SyncChangesResponse, error) {
	client := clientFromContext(ctx, r.client)
	since := time.UnixMilli(sinceMs)
	bookMember := func() predicate.Book { return book.HasMembersWith(user.ID(userID)) }

	bookIDs, err := client.Book.Query().Where(bookMember()).IDs(ctx)
	if err != nil {
		return nil, err
	}
	billIDs, err := client.Bill.Query().Where(bill.BookIDIn(bookIDs...)).IDs(ctx)
	if err != nil {
		return nil, err
	}

	books, err := loadPage(limit,
		func(b *ent.Book) time.Time { return b.UpdatedAt },
		func(n int, before *time.Time) ([]*ent.Book, error) {
			q := client.Book.Query().Where(bookMember(), book.UpdatedAtGT(since)).WithMembers()
			if before != nil {
				q = q.Where(book.UpdatedAtLTE(*before))
			}
			return q.Order(ent.Asc(book.FieldUpdatedAt), ent.Asc(book.FieldID)).Limit(n).All(ctx)
		},
	)
	if err != nil {
		return nil, err
	}

	bills, err := loadPage(limit,
		func(b *ent.Bill) time.Time { return b.UpdatedAt },
		func(n int, before *time.Time) ([]*ent.Bill, error) {
			q := client.Bill.Query().Where(bill.BookIDIn(bookIDs...), bill.UpdatedAtGT(since))
			if before != nil {
				q = q.Where(bill.UpdatedAtLTE(*before))
			}
			return q.Order(ent.Asc(bill.FieldUpdatedAt), ent.Asc(bill.FieldID)).Limit(n).All(ctx)
		},
	)
	if err != nil {
		return nil, err
	}

	categories, err := loadPage(limit,
		func(c *ent.Category) time.Time { return c.UpdatedAt },
		func(n int, before *time.Time) ([]*ent.Category, error) {
			q := client.Category.Query().Where(category.BookIDIn(bookIDs...), category.UpdatedAtGT(since))
			if before != nil {
				q = q.Where(category.UpdatedAtLTE(*before))
			}
			return q.Order(ent.Asc(category.FieldUpdatedAt), ent.Asc(category.FieldID)).Limit(n).All(ctx)
		},
	)
	if err != nil {
		return nil, err
	}

	// 图片没有 book_id，经所属账单归属到用户可见的账本
	images, err := loadPage(limit,
		func(i *ent.BillImage) time.Time { return i.UpdatedAt },
		func(n int, before *time.Time) ([]*ent.BillImage, error) {
			q := client.BillImage.Query().Where(billimage.BillIDIn(billIDs...), billimage.UpdatedAtGT(since))
			if before != nil {
				q = q.Where(billimage.UpdatedAtLTE(*before))
			}
			return q.Order(ent.Asc(billimage.FieldUpdatedAt), ent.Asc(billimage.FieldID)).Limit(n).All(ctx)
		},
	)
	if err != nil {
		return nil, err
	}

	return &domain.SyncChangesResponse{
		Books:      mapSlice(books.rows, entBookToDomain),
		Bills:      mapSlice(bills.rows, entBillToDomain),
		Categories: mapSlice(categories.rows, entCategoryToDomain),
		Images:     mapSlice(images.rows, entBillImageToDomain),
		HasMore:    isPaginating(books) || isPaginating(bills) || isPaginating(categories) || isPaginating(images),
		NextSince: nextSince(sinceMs,
			min4(books.cursor, bills.cursor, categories.cursor, images.cursor),
			max(
				latestOf(books.rows, func(b *ent.Book) time.Time { return b.UpdatedAt }),
				latestOf(bills.rows, func(b *ent.Bill) time.Time { return b.UpdatedAt }),
				latestOf(categories.rows, func(c *ent.Category) time.Time { return c.UpdatedAt }),
				latestOf(images.rows, func(i *ent.BillImage) time.Time { return i.UpdatedAt }),
			),
		),
	}, nil
}

// nextSince 选择下一页游标：
//   - 有实体被截断：取它们游标的最小值，未拉完的实体不会被跳过；
//   - 全部拉完：推进到本页最新记录之后。updated_at 精度高于毫秒，停在最新毫秒会让客户端
//     每次都重新拉到同一条记录，因此必须越过该毫秒。
func nextSince(sinceMs, truncatedCursor, latestDelivered int64) int64 {
	if truncatedCursor != math.MaxInt64 {
		return max(truncatedCursor, sinceMs)
	}
	if latestDelivered >= sinceMs {
		return latestDelivered + 1
	}
	return sinceMs
}

// latestOf 返回一批记录里最新的 updated_at 毫秒值；空集合返回 0。
func latestOf[T any](rows []*T, at func(*T) time.Time) int64 {
	var latest int64
	for _, row := range rows {
		if ms := at(row).UnixMilli(); ms > latest {
			latest = ms
		}
	}
	return latest
}

// loadPage 先取 limit+1 条判断是否截断：
//   - 未截断：该类已全部交付，cursor 返回 MaxInt64，不再限制公共游标；
//   - 截断：游标停在第 limit 条所在毫秒的末尾，该毫秒内的记录整批补齐后一起交付。
func loadPage[T any](
	limit int,
	at func(*T) time.Time,
	query func(n int, before *time.Time) ([]*T, error),
) (entityPage[T], error) {
	rows, err := query(limit+1, nil)
	if err != nil {
		return entityPage[T]{}, err
	}
	if len(rows) == 0 {
		return entityPage[T]{cursor: math.MaxInt64}, nil
	}
	if len(rows) <= limit {
		return entityPage[T]{rows: rows, cursor: math.MaxInt64}, nil
	}

	bucketEndMs := at(rows[limit-1]).UnixMilli() + 1
	bucketEnd := time.UnixMilli(bucketEndMs)
	complete, err := query(math.MaxInt32, &bucketEnd)
	if err != nil {
		return entityPage[T]{}, err
	}
	return entityPage[T]{rows: complete, cursor: bucketEndMs}, nil
}

func isPaginating[T any](page entityPage[T]) bool {
	return page.cursor != math.MaxInt64
}

// min4 返回四类实体的公共游标；全部拉完时返回 MaxInt64（表示不限）。
func min4(a, b, c, d int64) int64 {
	return min(min(a, b), min(c, d))
}
