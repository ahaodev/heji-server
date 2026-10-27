package usecase

import (
	"context"
	"testing"
	"time"

	"shadmin/domain"
)

// ---- 测试替身 ----

type stubBookRepo struct {
	books   map[string]*domain.Book
	members map[string]map[string]struct{}
	created *domain.Book
	deleted []string
}

func newStubBookRepo() *stubBookRepo {
	return &stubBookRepo{
		books:   map[string]*domain.Book{},
		members: map[string]map[string]struct{}{},
	}
}

func (r *stubBookRepo) Create(ctx context.Context, b *domain.Book) error {
	if b.ID == "" {
		b.ID = "book-" + b.Name
	}
	copied := *b
	r.books[b.ID] = &copied
	r.members[b.ID] = map[string]struct{}{b.CrtUserID: {}}
	r.created = b
	return nil
}

func (r *stubBookRepo) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	b, ok := r.books[id]
	if !ok {
		return nil, domain.ErrBookNotFound
	}
	return b, nil
}

func (r *stubBookRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Book, error) {
	var out []*domain.Book
	for id, b := range r.books {
		if _, ok := r.members[id][userID]; ok {
			out = append(out, b)
		}
	}
	return out, nil
}

func (r *stubBookRepo) Update(ctx context.Context, b *domain.Book) error {
	if _, ok := r.books[b.ID]; !ok {
		return domain.ErrBookNotFound
	}
	copied := *b
	r.books[b.ID] = &copied
	return nil
}

func (r *stubBookRepo) Delete(ctx context.Context, id string) error {
	if _, ok := r.books[id]; !ok {
		return domain.ErrBookNotFound
	}
	delete(r.books, id)
	r.deleted = append(r.deleted, id)
	return nil
}

func (r *stubBookRepo) AddMember(ctx context.Context, bookID, userID string) error {
	if _, ok := r.books[bookID]; !ok {
		return domain.ErrBookNotFound
	}
	r.members[bookID][userID] = struct{}{}
	return nil
}

func (r *stubBookRepo) IsMember(ctx context.Context, bookID, userID string) (bool, error) {
	_, ok := r.members[bookID][userID]
	return ok, nil
}

type stubNotifier struct {
	calls []string
}

func (n *stubNotifier) NotifyBookChange(ctx context.Context, bookID, actorID, changeType, entity string, payload any) {
	n.calls = append(n.calls, changeType)
}

type stubBillRepo struct {
	bills   []*domain.Bill
	nextID  int
	deleted []string
}

func (r *stubBillRepo) Create(ctx context.Context, b *domain.Bill) error {
	if b.ID == "" {
		r.nextID++
		b.ID = "bill-" + string(rune('0'+r.nextID))
	}
	r.bills = append(r.bills, b)
	return nil
}

func (r *stubBillRepo) GetByID(ctx context.Context, id string) (*domain.Bill, error) {
	for _, b := range r.bills {
		if b.ID == id {
			return b, nil
		}
	}
	return nil, domain.ErrBillNotFound
}

func (r *stubBillRepo) Query(ctx context.Context, filter domain.BillQueryFilter) (*domain.PagedResult[*domain.Bill], error) {
	var out []*domain.Bill
	for _, b := range r.bills {
		if b.BookID == filter.BookID {
			out = append(out, b)
		}
	}
	return domain.NewPagedResult(out, len(out), filter.Page, filter.PageSize), nil
}

func (r *stubBillRepo) Update(ctx context.Context, b *domain.Bill) error {
	for i, existing := range r.bills {
		if existing.ID == b.ID {
			r.bills[i] = b
			return nil
		}
	}
	return domain.ErrBillNotFound
}

func (r *stubBillRepo) Delete(ctx context.Context, id string) error {
	for i, b := range r.bills {
		if b.ID == id {
			r.bills = append(r.bills[:i], r.bills[i+1:]...)
			r.deleted = append(r.deleted, id)
			return nil
		}
	}
	return domain.ErrBillNotFound
}

func newTestBook(t *testing.T, repo *stubBookRepo, ownerID string) *domain.Book {
	t.Helper()
	book, err := NewBookUsecase(repo, nil, time.Second).CreateBook(context.Background(), ownerID, &domain.CreateBookRequest{Name: "测试账本"})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}
	return book
}

// ---- 账本用例 ----

func TestCreateBookAddsCreatorAsMember(t *testing.T) {
	repo := newStubBookRepo()
	book := newTestBook(t, repo, "user-1")

	if book.CrtUserID != "user-1" {
		t.Fatalf("creator = %q, want user-1", book.CrtUserID)
	}
	if ok, _ := repo.IsMember(context.Background(), book.ID, "user-1"); !ok {
		t.Fatal("creator should be a member of the book")
	}
}

func TestGetBookRejectsNonMember(t *testing.T) {
	repo := newStubBookRepo()
	book := newTestBook(t, repo, "user-1")
	usecase := NewBookUsecase(repo, nil, time.Second)

	if _, err := usecase.GetBook(context.Background(), "user-2", book.ID); err != domain.ErrNotBookMember {
		t.Fatalf("err = %v, want ErrNotBookMember", err)
	}

	usecase.JoinBook(context.Background(), "user-2", book.ID)
	if _, err := usecase.GetBook(context.Background(), "user-2", book.ID); err != nil {
		t.Fatalf("member should read book, got %v", err)
	}
}

func TestDeleteBookOnlyByCreator(t *testing.T) {
	repo := newStubBookRepo()
	book := newTestBook(t, repo, "user-1")
	usecase := NewBookUsecase(repo, nil, time.Second)

	if err := usecase.JoinBook(context.Background(), "user-2", book.ID); err != nil {
		t.Fatalf("join book: %v", err)
	}
	if err := usecase.DeleteBook(context.Background(), "user-2", book.ID); err != domain.ErrNotBookMember {
		t.Fatalf("non-creator delete err = %v, want ErrNotBookMember", err)
	}
	if err := usecase.DeleteBook(context.Background(), "user-1", book.ID); err != nil {
		t.Fatalf("creator delete: %v", err)
	}
}

func TestDeleteBookNotifiesMembers(t *testing.T) {
	repo := newStubBookRepo()
	notifier := &stubNotifier{}
	usecase := NewBookUsecase(repo, notifier, time.Second)
	book, err := usecase.CreateBook(context.Background(), "user-1", &domain.CreateBookRequest{Name: "通知账本"})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}

	if err := usecase.DeleteBook(context.Background(), "user-1", book.ID); err != nil {
		t.Fatalf("delete book: %v", err)
	}
	if len(notifier.calls) != 2 || notifier.calls[0] != "ADD_BOOK" || notifier.calls[1] != "DELETE_BOOK" {
		t.Fatalf("notifications = %v, want [ADD_BOOK DELETE_BOOK]", notifier.calls)
	}
}

// ---- 账单用例 ----

func TestCreateBillRejectsNonMember(t *testing.T) {
	books := newStubBookRepo()
	book := newTestBook(t, books, "user-1")
	usecase := NewBillUsecase(&stubBillRepo{}, books, nil, nil, time.Second)

	_, err := usecase.CreateBill(context.Background(), "user-2", &domain.CreateBillRequest{
		BookID: book.ID,
		Money:  100,
		Type:   domain.BillTypeExpenditure,
		Time:   "2026-10-27",
	})
	if err != domain.ErrNotBookMember {
		t.Fatalf("err = %v, want ErrNotBookMember", err)
	}
}

func TestCreateBillAcceptsTimestampForm(t *testing.T) {
	books := newStubBookRepo()
	book := newTestBook(t, books, "user-1")
	bills := &stubBillRepo{}
	usecase := NewBillUsecase(bills, books, nil, nil, time.Second)

	bill, err := usecase.CreateBill(context.Background(), "user-1", &domain.CreateBillRequest{
		BookID: book.ID,
		Money:  2580,
		Type:   domain.BillTypeExpenditure,
		Time:   "2026-10-27 12:30:00",
	})
	if err != nil {
		t.Fatalf("create bill with timestamp: %v", err)
	}
	if bill.CrtUser != "user-1" {
		t.Fatalf("crt_user = %q, want user-1", bill.CrtUser)
	}
}

func TestCreateBillRejectsInvalidTime(t *testing.T) {
	books := newStubBookRepo()
	book := newTestBook(t, books, "user-1")
	usecase := NewBillUsecase(&stubBillRepo{}, books, nil, nil, time.Second)

	_, err := usecase.CreateBill(context.Background(), "user-1", &domain.CreateBillRequest{
		BookID: book.ID,
		Money:  100,
		Type:   domain.BillTypeIncome,
		Time:   "27/10/2026",
	})
	if err != domain.ErrInvalidBillTime {
		t.Fatalf("err = %v, want ErrInvalidBillTime", err)
	}
}

func TestListBillsRequiresBookID(t *testing.T) {
	usecase := NewBillUsecase(&stubBillRepo{}, newStubBookRepo(), nil, nil, time.Second)
	if _, err := usecase.ListBills(context.Background(), "user-1", domain.BillQueryFilter{}); err == nil {
		t.Fatal("expected error when book_id is missing")
	}
}

func TestDeleteBillHidesFromListForNonMember(t *testing.T) {
	books := newStubBookRepo()
	book := newTestBook(t, books, "user-1")
	bills := &stubBillRepo{}
	usecase := NewBillUsecase(bills, books, nil, nil, time.Second)

	bill, err := usecase.CreateBill(context.Background(), "user-1", &domain.CreateBillRequest{
		BookID: book.ID,
		Money:  100,
		Type:   domain.BillTypeIncome,
		Time:   "2026-10-27",
	})
	if err != nil {
		t.Fatalf("create bill: %v", err)
	}
	if err := usecase.DeleteBill(context.Background(), "user-2", bill.ID); err != domain.ErrNotBookMember {
		t.Fatalf("non-member delete err = %v, want ErrNotBookMember", err)
	}
	if err := usecase.DeleteBill(context.Background(), "user-1", bill.ID); err != nil {
		t.Fatalf("owner delete: %v", err)
	}
}

// ---- 金额换算（前端契约） ----

func TestBillTypesAreOneAndMinusOne(t *testing.T) {
	if domain.BillTypeIncome != 1 || domain.BillTypeExpenditure != -1 {
		t.Fatalf("bill types = %d/%d, want 1/-1", domain.BillTypeIncome, domain.BillTypeExpenditure)
	}
}
