package repository

import (
	"context"
	"time"

	"shadmin/domain"
	"shadmin/ent"
	"shadmin/ent/bill"
	"shadmin/ent/book"
	"shadmin/ent/predicate"
	"shadmin/ent/user"
)

type entSyncRepository struct {
	client *ent.Client
}

func NewSyncRepository(client *ent.Client) domain.SyncRepository {
	return &entSyncRepository{client: client}
}

// QueryChanges 返回 since 之后有变更的账本与账单（含已软删除记录，客户端据此删除本地行）。
// limit 同时约束两类实体，has_more 表示任一类达到上限，调用方用 next_since 续拉。
func (r *entSyncRepository) QueryChanges(ctx context.Context, userID string, sinceMs int64, limit int) (*domain.SyncChangesResponse, error) {
	client := clientFromContext(ctx, r.client)
	since := time.UnixMilli(sinceMs)

	bookMember := func() predicate.Book { return book.HasMembersWith(user.ID(userID)) }

	books, err := client.Book.Query().
		Where(bookMember(), book.UpdatedAtGT(since)).
		WithMembers().
		Order(ent.Asc(book.FieldUpdatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 用户全部账本（含已删除）下的账单变更
	bookIDs, err := client.Book.Query().Where(bookMember()).IDs(ctx)
	if err != nil {
		return nil, err
	}

	bills, err := client.Bill.Query().
		Where(bill.BookIDIn(bookIDs...), bill.UpdatedAtGT(since)).
		Order(ent.Asc(bill.FieldUpdatedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}

	domainBooks := mapSlice(books, entBookToDomain)
	domainBills := mapSlice(bills, entBillToDomain)

	nextSince := sinceMs
	for _, b := range domainBooks {
		if ms := b.UpdatedAt.UnixMilli(); ms > nextSince {
			nextSince = ms
		}
	}
	for _, b := range domainBills {
		if ms := b.UpdatedAt.UnixMilli(); ms > nextSince {
			nextSince = ms
		}
	}

	return &domain.SyncChangesResponse{
		Books:     domainBooks,
		Bills:     domainBills,
		HasMore:   len(books) >= limit || len(bills) >= limit,
		NextSince: nextSince,
	}, nil
}
