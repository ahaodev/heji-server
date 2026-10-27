package repository

import (
	"context"
	"shadmin/domain"
	"shadmin/ent"
	"shadmin/ent/category"
)

type entCategoryRepository struct {
	client *ent.Client
}

func NewCategoryRepository(client *ent.Client) domain.CategoryRepository {
	return &entCategoryRepository{client: client}
}

func (r *entCategoryRepository) Create(ctx context.Context, c *domain.Category) error {
	create := clientFromContext(ctx, r.client).Category.Create().
		SetBookID(c.BookID).
		SetName(c.Name).
		SetType(c.Type).
		SetLevel(c.Level).
		SetIndex(c.Index).
		SetParentID(c.ParentID).
		SetCreateUser(c.CreateUser)
	if c.ID != "" {
		create = create.SetID(c.ID)
	}

	created, err := create.Save(ctx)
	if err != nil {
		return err
	}
	c.ID = created.ID
	return nil
}

func (r *entCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	c, err := clientFromContext(ctx, r.client).Category.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return entCategoryToDomain(c), nil
}

func (r *entCategoryRepository) ListByBookID(ctx context.Context, bookID string) ([]*domain.Category, error) {
	categories, err := clientFromContext(ctx, r.client).Category.Query().
		Where(category.BookID(bookID)).
		Order(ent.Asc(category.FieldIndex)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return mapSlice(categories, entCategoryToDomain), nil
}

func (r *entCategoryRepository) Update(ctx context.Context, c *domain.Category) error {
	err := clientFromContext(ctx, r.client).Category.UpdateOneID(c.ID).
		SetName(c.Name).
		SetType(c.Type).
		SetLevel(c.Level).
		SetIndex(c.Index).
		SetParentID(c.ParentID).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrCategoryNotFound
		}
		return err
	}
	return nil
}

func (r *entCategoryRepository) Delete(ctx context.Context, id string) error {
	err := clientFromContext(ctx, r.client).Category.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrCategoryNotFound
		}
		return err
	}
	return nil
}

func entCategoryToDomain(c *ent.Category) *domain.Category {
	return &domain.Category{
		ID:         c.ID,
		BookID:     c.BookID,
		Name:       c.Name,
		Type:       c.Type,
		Level:      c.Level,
		Index:      c.Index,
		ParentID:   c.ParentID,
		CreateUser: c.CreateUser,
	}
}
