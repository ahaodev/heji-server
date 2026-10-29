package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"shadmin/domain"
)

type categoryUsecase struct {
	categoryRepository domain.CategoryRepository
	bookRepository     domain.BookRepository
	contextTimeout     time.Duration
}

func NewCategoryUsecase(categoryRepository domain.CategoryRepository, bookRepository domain.BookRepository, timeout time.Duration) domain.CategoryUseCase {
	return &categoryUsecase{
		categoryRepository: categoryRepository,
		bookRepository:     bookRepository,
		contextTimeout:     timeout,
	}
}

func (u *categoryUsecase) CreateCategory(ctx context.Context, userID string, req *domain.CreateCategoryRequest) (*domain.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.requireMember(ctx, req.BookID, userID); err != nil {
		return nil, err
	}
	return u.create(ctx, userID, req)
}

func (u *categoryUsecase) CreateCategories(ctx context.Context, userID string, reqs []*domain.CreateCategoryRequest) ([]*domain.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	result := make([]*domain.Category, 0, len(reqs))
	for _, req := range reqs {
		if err := u.requireMember(ctx, req.BookID, userID); err != nil {
			return nil, err
		}
		category, err := u.create(ctx, userID, req)
		if err != nil {
			return nil, fmt.Errorf("create category %q: %w", req.Name, err)
		}
		result = append(result, category)
	}
	return result, nil
}

func (u *categoryUsecase) GetCategory(ctx context.Context, userID, id string) (*domain.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	category, err := u.categoryRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.requireMember(ctx, category.BookID, userID); err != nil {
		return nil, err
	}
	return category, nil
}

func (u *categoryUsecase) ListCategories(ctx context.Context, userID, bookID string) ([]*domain.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if bookID == "" {
		return nil, fmt.Errorf("book_id is required: %w", domain.ErrNotBookMember)
	}
	if err := u.requireMember(ctx, bookID, userID); err != nil {
		return nil, err
	}

	categories, err := u.categoryRepository.ListByBookID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
}

func (u *categoryUsecase) UpdateCategory(ctx context.Context, userID, id string, req *domain.UpdateCategoryRequest) (*domain.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	category, err := u.categoryRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.requireMember(ctx, category.BookID, userID); err != nil {
		return nil, err
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Type != nil {
		category.Type = *req.Type
	}
	if req.Level != nil {
		category.Level = *req.Level
	}
	if req.Index != nil {
		category.Index = *req.Index
	}
	if req.ParentID != nil {
		category.ParentID = *req.ParentID
	}

	if err := u.categoryRepository.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return category, nil
}

func (u *categoryUsecase) DeleteCategory(ctx context.Context, userID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	category, err := u.categoryRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.requireMember(ctx, category.BookID, userID); err != nil {
		return err
	}

	if err := u.categoryRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

func (u *categoryUsecase) create(ctx context.Context, userID string, req *domain.CreateCategoryRequest) (*domain.Category, error) {
	// 客户端离线创建分类时自带 ID，重复上传按更新处理，保证离线重试幂等
	if req.ID != "" {
		existing, err := u.categoryRepository.GetByID(ctx, req.ID)
		switch {
		case err == nil:
			if existing.BookID != req.BookID {
				return nil, domain.ErrCategoryMismatch
			}
			existing.Name = req.Name
			existing.Type = req.Type
			existing.Level = req.Level
			existing.Index = req.Index
			existing.ParentID = req.ParentID
			if err := u.categoryRepository.Update(ctx, existing); err != nil {
				return nil, fmt.Errorf("update category: %w", err)
			}
			return existing, nil
		case !errors.Is(err, domain.ErrCategoryNotFound):
			return nil, err
		}
	}

	category := &domain.Category{
		ID:         req.ID,
		BookID:     req.BookID,
		Name:       req.Name,
		Type:       req.Type,
		Level:      req.Level,
		Index:      req.Index,
		ParentID:   req.ParentID,
		CreateUser: userID,
	}
	if err := u.categoryRepository.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return category, nil
}

func (u *categoryUsecase) requireMember(ctx context.Context, bookID, userID string) error {
	ok, err := u.bookRepository.IsMember(ctx, bookID, userID)
	if err != nil {
		return fmt.Errorf("check book membership: %w", err)
	}
	if !ok {
		return domain.ErrNotBookMember
	}
	return nil
}
