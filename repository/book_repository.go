package repository

import (
	"context"
	"time"

	"shadmin/domain"
	"shadmin/ent"
	"shadmin/ent/book"
	"shadmin/ent/predicate"
	"shadmin/ent/user"
)

type entBookRepository struct {
	client *ent.Client
}

func NewBookRepository(client *ent.Client) domain.BookRepository {
	return &entBookRepository{client: client}
}

// bookMemberPredicate 匹配“账本成员包含该用户”。
func bookMemberPredicate(userID string) predicate.Book {
	return book.HasMembersWith(user.ID(userID))
}

func (r *entBookRepository) Create(ctx context.Context, b *domain.Book) error {
	create := clientFromContext(ctx, r.client).Book.Create().
		SetName(b.Name).
		SetType(b.Type).
		SetBanner(b.Banner).
		SetCrtUserID(b.CrtUserID).
		SetIsInitial(b.IsInitial)
	if b.ID != "" {
		create = create.SetID(b.ID)
	}

	created, err := create.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.ErrBookNameExists
		}
		return err
	}

	b.ID = created.ID
	b.CreatedAt = created.CreatedAt
	b.UpdatedAt = created.UpdatedAt

	// 创建者即账本成员
	if _, err := clientFromContext(ctx, r.client).Book.UpdateOneID(created.ID).
		AddMemberIDs(b.CrtUserID).
		Save(ctx); err != nil {
		return err
	}
	b.Members = []string{b.CrtUserID}
	return nil
}

func (r *entBookRepository) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	b, err := clientFromContext(ctx, r.client).Book.Query().
		Where(book.ID(id), book.DeletedAtIsNil()).
		WithMembers().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrBookNotFound
		}
		return nil, err
	}
	return entBookToDomain(b), nil
}

func (r *entBookRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Book, error) {
	books, err := clientFromContext(ctx, r.client).Book.Query().
		Where(bookMemberPredicate(userID), book.DeletedAtIsNil()).
		WithMembers().
		Order(ent.Desc(book.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return mapSlice(books, entBookToDomain), nil
}

func (r *entBookRepository) Update(ctx context.Context, b *domain.Book) error {
	updated, err := clientFromContext(ctx, r.client).Book.UpdateOneID(b.ID).
		SetName(b.Name).
		SetType(b.Type).
		SetBanner(b.Banner).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrBookNotFound
		}
		if ent.IsConstraintError(err) {
			return domain.ErrBookNameExists
		}
		return err
	}
	b.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *entBookRepository) Delete(ctx context.Context, id string) error {
	err := clientFromContext(ctx, r.client).Book.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrBookNotFound
		}
		return err
	}
	return nil
}

func (r *entBookRepository) AddMember(ctx context.Context, bookID, userID string) error {
	err := clientFromContext(ctx, r.client).Book.UpdateOneID(bookID).
		AddMemberIDs(userID).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrBookNotFound
		}
		return err
	}
	return nil
}

func (r *entBookRepository) IsMember(ctx context.Context, bookID, userID string) (bool, error) {
	return clientFromContext(ctx, r.client).Book.Query().
		Where(book.ID(bookID), book.DeletedAtIsNil(), bookMemberPredicate(userID)).
		Exist(ctx)
}

func (r *entBookRepository) ListMemberIDs(ctx context.Context, bookID string) ([]string, error) {
	b, err := clientFromContext(ctx, r.client).Book.Query().
		Where(book.ID(bookID)).
		WithMembers().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrBookNotFound
		}
		return nil, err
	}
	members := make([]string, 0, len(b.Edges.Members))
	for _, m := range b.Edges.Members {
		members = append(members, m.ID)
	}
	return members, nil
}

func entBookToDomain(b *ent.Book) *domain.Book {
	members := make([]string, 0, len(b.Edges.Members))
	for _, m := range b.Edges.Members {
		members = append(members, m.ID)
	}
	return &domain.Book{
		ID:        b.ID,
		Name:      b.Name,
		Type:      b.Type,
		Banner:    b.Banner,
		CrtUserID: b.CrtUserID,
		IsInitial: b.IsInitial,
		Members:   members,
		DeletedAt: b.DeletedAt,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}
