package repository

import (
	"context"
	"time"

	"shadmin/domain"
	"shadmin/ent"
	"shadmin/ent/billimage"
)

type entBillImageRepository struct {
	client *ent.Client
}

func NewBillImageRepository(client *ent.Client) domain.BillImageRepository {
	return &entBillImageRepository{client: client}
}

func (r *entBillImageRepository) Create(ctx context.Context, img *domain.BillImage) error {
	create := clientFromContext(ctx, r.client).BillImage.Create().
		SetBillID(img.BillID).
		SetKey(img.Key).
		SetExt(img.Ext).
		SetSize(img.Size).
		SetMd5(img.MD5)
	if img.ID != "" {
		create = create.SetID(img.ID)
	}

	created, err := create.Save(ctx)
	if err != nil {
		return err
	}
	img.ID = created.ID
	img.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *entBillImageRepository) GetByID(ctx context.Context, id string) (*domain.BillImage, error) {
	img, err := clientFromContext(ctx, r.client).BillImage.Query().
		Where(billimage.ID(id), billimage.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.ErrBillImageNotFound
		}
		return nil, err
	}
	return entBillImageToDomain(img), nil
}

func (r *entBillImageRepository) ListByBillID(ctx context.Context, billID string) ([]*domain.BillImage, error) {
	images, err := clientFromContext(ctx, r.client).BillImage.Query().
		Where(billimage.BillID(billID), billimage.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return mapSlice(images, entBillImageToDomain), nil
}

// Delete 软删除图片记录：客户端按 deleted_at 判断本地行是否需要移除。
// 文件本身由用例层删除，保留下来的记录只是同步事件。
func (r *entBillImageRepository) Delete(ctx context.Context, id string) error {
	err := clientFromContext(ctx, r.client).BillImage.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrBillImageNotFound
		}
		return err
	}
	return nil
}

func entBillImageToDomain(img *ent.BillImage) *domain.BillImage {
	return &domain.BillImage{
		ID:        img.ID,
		BillID:    img.BillID,
		Key:       img.Key,
		Ext:       img.Ext,
		Size:      img.Size,
		MD5:       img.Md5,
		DeletedAt: img.DeletedAt,
		UpdatedAt: img.UpdatedAt,
	}
}
