package repository

import (
	"context"
	"time"

	"shadmin/domain"
	"shadmin/ent"
	"shadmin/ent/bill"
)

type entBillRepository struct {
	client *ent.Client
}

func NewBillRepository(client *ent.Client) domain.BillRepository {
	return &entBillRepository{client: client}
}

func (r *entBillRepository) Create(ctx context.Context, b *domain.Bill) error {
	create := clientFromContext(ctx, r.client).Bill.Create().
		SetBookID(b.BookID).
		SetMoney(b.Money).
		SetType(b.Type).
		SetCategory(b.Category).
		SetCrtUser(b.CrtUser).
		SetTime(b.Time).
		SetRemark(b.Remark)
	if b.ID != "" {
		create = create.SetID(b.ID)
	}
	if len(b.Images) > 0 {
		create = create.SetImages(b.Images)
	}

	created, err := create.Save(ctx)
	if err != nil {
		return err
	}

	b.ID = created.ID
	b.CreatedAt = created.CreatedAt
	b.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *entBillRepository) GetByID(ctx context.Context, id string) (*domain.Bill, error) {
	b, err := clientFromContext(ctx, r.client).Bill.Query().
		Where(bill.ID(id), bill.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrBillNotFound
		}
		return nil, err
	}
	return entBillToDomain(b), nil
}

func (r *entBillRepository) Query(ctx context.Context, filter domain.BillQueryFilter) (*domain.PagedResult[*domain.Bill], error) {
	query := clientFromContext(ctx, r.client).Bill.Query().
		Where(bill.BookID(filter.BookID), bill.DeletedAtIsNil())

	if filter.Type != nil {
		query = query.Where(bill.TypeEQ(*filter.Type))
	}
	if filter.Category != "" {
		query = query.Where(bill.Category(filter.Category))
	}
	if filter.TimeFrom != "" {
		query = query.Where(bill.TimeGTE(filter.TimeFrom))
	}
	if filter.TimeTo != "" {
		query = query.Where(bill.TimeLTE(filter.TimeTo))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	offset, limit := filter.Paginate()
	bills, err := query.
		Order(ent.Desc(bill.FieldTime), ent.Desc(bill.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return domain.NewPagedResult(mapSlice(bills, entBillToDomain), total, filter.Page, filter.PageSize), nil
}

func (r *entBillRepository) Update(ctx context.Context, b *domain.Bill) error {
	updated, err := clientFromContext(ctx, r.client).Bill.UpdateOneID(b.ID).
		SetMoney(b.Money).
		SetType(b.Type).
		SetCategory(b.Category).
		SetTime(b.Time).
		SetRemark(b.Remark).
		SetImages(b.Images).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrBillNotFound
		}
		return err
	}
	b.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *entBillRepository) Delete(ctx context.Context, id string) error {
	err := clientFromContext(ctx, r.client).Bill.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrBillNotFound
		}
		return err
	}
	return nil
}

func entBillToDomain(b *ent.Bill) *domain.Bill {
	images := b.Images
	if images == nil {
		images = []string{}
	}
	return &domain.Bill{
		ID:        b.ID,
		BookID:    b.BookID,
		Money:     b.Money,
		Type:      b.Type,
		Category:  b.Category,
		CrtUser:   b.CrtUser,
		Time:      b.Time,
		Remark:    b.Remark,
		Images:    images,
		DeletedAt: b.DeletedAt,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}
