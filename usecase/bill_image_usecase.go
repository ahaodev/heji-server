package usecase

import (
	"context"
	"fmt"
	"time"

	"shadmin/domain"
)

type billImageUsecase struct {
	imageRepository domain.BillImageRepository
	billRepository  domain.BillRepository
	bookRepository  domain.BookRepository
	fileRepository  domain.FileRepository
	contextTimeout  time.Duration
}

func NewBillImageUsecase(
	imageRepository domain.BillImageRepository,
	billRepository domain.BillRepository,
	bookRepository domain.BookRepository,
	fileRepository domain.FileRepository,
	timeout time.Duration,
) domain.BillImageUseCase {
	return &billImageUsecase{
		imageRepository: imageRepository,
		billRepository:  billRepository,
		bookRepository:  bookRepository,
		fileRepository:  fileRepository,
		contextTimeout:  timeout,
	}
}

func (u *billImageUsecase) UploadImage(ctx context.Context, userID, billID string, img *domain.BillImage, upload *domain.UploadRequest) (*domain.BillImage, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.requireBillMember(ctx, billID, userID); err != nil {
		return nil, err
	}

	result, err := u.fileRepository.Upload(ctx, upload)
	if err != nil {
		return nil, fmt.Errorf("upload image file: %w", err)
	}

	img.BillID = billID
	img.Key = result.Key
	img.Size = result.Size
	img.MD5 = result.ETag

	if err := u.imageRepository.Create(ctx, img); err != nil {
		// 元数据写入失败时清理已上传的文件，避免孤儿对象
		_ = u.fileRepository.Delete(ctx, upload.Bucket, result.Key)
		return nil, fmt.Errorf("save image record: %w", err)
	}
	return img, nil
}

func (u *billImageUsecase) GetImage(ctx context.Context, userID, id string) (*domain.BillImage, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	img, err := u.imageRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.requireBillMember(ctx, img.BillID, userID); err != nil {
		return nil, err
	}
	return img, nil
}

func (u *billImageUsecase) ListByBillID(ctx context.Context, userID, billID string) ([]*domain.BillImage, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.requireBillMember(ctx, billID, userID); err != nil {
		return nil, err
	}
	images, err := u.imageRepository.ListByBillID(ctx, billID)
	if err != nil {
		return nil, fmt.Errorf("list bill images: %w", err)
	}
	return images, nil
}

func (u *billImageUsecase) DeleteImage(ctx context.Context, userID, billID, imageID string) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.requireBillMember(ctx, billID, userID); err != nil {
		return err
	}

	img, err := u.imageRepository.GetByID(ctx, imageID)
	if err != nil {
		return err
	}
	if img.BillID != billID {
		return domain.ErrImageBillMismatch
	}

	// 先删元数据再删文件：文件残留可清理，元数据残留会导致坏链
	if err := u.imageRepository.Delete(ctx, imageID); err != nil {
		return fmt.Errorf("delete image record: %w", err)
	}
	_ = u.fileRepository.Delete(ctx, domain.BillImageBucket, img.Key)
	return nil
}

// requireBillMember 确认账单存在且当前用户是该账单所属账本的成员。
func (u *billImageUsecase) requireBillMember(ctx context.Context, billID, userID string) error {
	bill, err := u.billRepository.GetByID(ctx, billID)
	if err != nil {
		return err
	}
	ok, err := u.bookRepository.IsMember(ctx, bill.BookID, userID)
	if err != nil {
		return fmt.Errorf("check book membership: %w", err)
	}
	if !ok {
		return domain.ErrNotBookMember
	}
	return nil
}
