package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"shadmin/domain"
	"shadmin/ent"

	_ "github.com/mattn/go-sqlite3"
)

// 增量同步不能在截断时把没交付的记录跳过：账本比账单新时，旧实现的 next_since
// 会直接越过尚未返回的账单，这些账单再也不会被拉到。
func TestQueryChangesDoesNotSkipTruncatedBills(t *testing.T) {
	client := newSyncTestClient(t)
	ctx := context.Background()
	owner := createSyncTestUser(t, client)
	bookID := createSyncTestBook(t, client, owner.ID)

	base := time.Now().Add(-time.Hour)
	for i := 0; i < 3; i++ {
		b := createSyncTestBill(t, client, bookID, owner.ID)
		setBillUpdatedAt(t, client, b.ID, base.Add(time.Duration(i)*time.Second))
	}
	// 账本比所有账单都新：旧实现的 next_since 会推进到这里
	touchBook(t, client, bookID, base.Add(time.Minute))

	repo := NewSyncRepository(client)
	seen := map[string]bool{}
	since := int64(0)
	for page := 0; page < 10; page++ {
		result, err := repo.QueryChanges(ctx, owner.ID, since, 2)
		if err != nil {
			t.Fatalf("QueryChanges: %v", err)
		}
		for _, b := range result.Bills {
			seen[b.ID] = true
		}
		if !result.HasMore {
			break
		}
		if page > 0 && result.NextSince <= since {
			t.Fatalf("next_since did not advance: %d", result.NextSince)
		}
		since = result.NextSince
	}

	if len(seen) != 3 {
		t.Fatalf("pulled %d bills, want 3", len(seen))
	}
}

// 分类与图片也属于增量拉取范围，否则换设备后分类缺失、他人账单的图片不可见。
func TestQueryChangesIncludesCategories(t *testing.T) {
	client := newSyncTestClient(t)
	ctx := context.Background()
	owner := createSyncTestUser(t, client)
	bookID := createSyncTestBook(t, client, owner.ID)

	if _, err := client.Category.Create().
		SetBookID(bookID).
		SetName("午餐").
		SetType(domain.BillTypeExpenditure).
		Save(ctx); err != nil {
		t.Fatalf("create category: %v", err)
	}

	result, err := NewSyncRepository(client).QueryChanges(ctx, owner.ID, 0, 100)
	if err != nil {
		t.Fatalf("QueryChanges: %v", err)
	}
	if len(result.Categories) != 1 || result.Categories[0].Name != "午餐" {
		t.Fatalf("categories = %+v, want one 午餐", result.Categories)
	}
}

// 软删除的分类也要交付，客户端据此删除本地行。
func TestQueryChangesDeliversDeletedCategory(t *testing.T) {
	client := newSyncTestClient(t)
	ctx := context.Background()
	owner := createSyncTestUser(t, client)
	bookID := createSyncTestBook(t, client, owner.ID)

	created, err := client.Category.Create().
		SetBookID(bookID).
		SetName("零食").
		SetType(domain.BillTypeExpenditure).
		Save(ctx)
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	repo := NewCategoryRepository(client)
	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	result, err := NewSyncRepository(client).QueryChanges(ctx, owner.ID, 0, 100)
	if err != nil {
		t.Fatalf("QueryChanges: %v", err)
	}
	if len(result.Categories) != 1 || result.Categories[0].DeletedAt == nil {
		t.Fatalf("categories = %+v, want one soft-deleted row", result.Categories)
	}
	// 常规列表不应包含已删除的分类
	list, err := repo.ListByBookID(ctx, bookID)
	if err != nil {
		t.Fatalf("ListByBookID: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("ListByBookID returned %d rows, want 0", len(list))
	}
}

func newSyncTestClient(t *testing.T) *ent.Client {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	client, err := ent.Open("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", name))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	return client
}

func createSyncTestUser(t *testing.T, client *ent.Client) *ent.User {
	t.Helper()
	u, err := client.User.Create().
		SetUsername("owner").
		SetPassword("hash").
		SetStatus("active").
		Save(context.Background())
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func createSyncTestBook(t *testing.T, client *ent.Client, ownerID string) string {
	t.Helper()
	b, err := client.Book.Create().
		SetName("测试账本").
		SetCrtUserID(ownerID).
		AddMemberIDs(ownerID).
		Save(context.Background())
	if err != nil {
		t.Fatalf("create book: %v", err)
	}
	return b.ID
}

func createSyncTestBill(t *testing.T, client *ent.Client, bookID, userID string) *domain.Bill {
	t.Helper()
	b := &domain.Bill{
		BookID:  bookID,
		Money:   100,
		Type:    domain.BillTypeExpenditure,
		CrtUser: userID,
		Time:    "2026-10-27",
	}
	if err := NewBillRepository(client).Create(context.Background(), b); err != nil {
		t.Fatalf("create bill: %v", err)
	}
	return b
}

// setBillUpdatedAt 把 updated_at 调到指定时刻：增量同步按这一列翻页，正常写入只会拿到当前时间。
func setBillUpdatedAt(t *testing.T, client *ent.Client, id string, at time.Time) {
	t.Helper()
	if err := client.Bill.UpdateOneID(id).SetUpdatedAt(at).Exec(context.Background()); err != nil {
		t.Fatalf("set bill updated_at: %v", err)
	}
}

func touchBook(t *testing.T, client *ent.Client, bookID string, at time.Time) {
	t.Helper()
	if err := client.Book.UpdateOneID(bookID).SetName("测试账本改").SetUpdatedAt(at).Exec(context.Background()); err != nil {
		t.Fatalf("update book: %v", err)
	}
}
