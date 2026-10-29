package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"shadmin/domain"
)

type bookUsecase struct {
	bookRepository domain.BookRepository
	notifier       domain.BillingNotifier
	contextTimeout time.Duration
}

func NewBookUsecase(bookRepository domain.BookRepository, notifier domain.BillingNotifier, timeout time.Duration) domain.BookUseCase {
	return &bookUsecase{
		bookRepository: bookRepository,
		notifier:       notifier,
		contextTimeout: timeout,
	}
}

func (u *bookUsecase) CreateBook(ctx context.Context, userID string, req *domain.CreateBookRequest) (*domain.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 客户端离线创建账本时自带 ID，且把“改账本”也发到这里：
	// 同 ID 再提交按更新处理，保证离线重试幂等。
	if req.ID != "" {
		existing, err := u.bookRepository.GetByID(ctx, req.ID)
		switch {
		case err == nil:
			return u.applyUpdate(ctx, userID, existing, req)
		case !errors.Is(err, domain.ErrBookNotFound):
			return nil, err
		}
	}

	book := &domain.Book{
		ID:        req.ID,
		Name:      req.Name,
		Type:      req.Type,
		Banner:    req.Banner,
		CrtUserID: userID,
		IsInitial: req.IsInitial,
	}
	if err := u.bookRepository.Create(ctx, book); err != nil {
		return nil, fmt.Errorf("create book: %w", err)
	}

	notifyBookChange(ctx, u.notifier, book.ID, userID, "ADD_BOOK", "book", book)
	return book, nil
}

// applyUpdate 把 Create 请求当作更新写入已存在的账本；ID 归属其他人时拒绝，避免覆盖他人数据。
func (u *bookUsecase) applyUpdate(ctx context.Context, userID string, book *domain.Book, req *domain.CreateBookRequest) (*domain.Book, error) {
	if book.CrtUserID != userID {
		return nil, domain.ErrBookIDTaken
	}
	book.Name = req.Name
	book.Type = req.Type
	book.Banner = req.Banner
	if err := u.bookRepository.Update(ctx, book); err != nil {
		return nil, fmt.Errorf("update book: %w", err)
	}
	notifyBookChange(ctx, u.notifier, book.ID, userID, "UPDATE_BOOK", "book", book)
	return book, nil
}

func (u *bookUsecase) GetBook(ctx context.Context, userID, id string) (*domain.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	book, err := u.bookRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 账本详情只对成员开放，避免通过 ID 枚举他人账本
	if err := u.requireMember(ctx, id, userID); err != nil {
		return nil, err
	}
	return book, nil
}

func (u *bookUsecase) ListBooks(ctx context.Context, userID string) ([]*domain.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	books, err := u.bookRepository.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	return books, nil
}

func (u *bookUsecase) UpdateBook(ctx context.Context, userID, id string, req *domain.UpdateBookRequest) (*domain.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.requireMember(ctx, id, userID); err != nil {
		return nil, err
	}

	book, err := u.bookRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		book.Name = *req.Name
	}
	if req.Type != nil {
		book.Type = *req.Type
	}
	if req.Banner != nil {
		book.Banner = *req.Banner
	}

	if err := u.bookRepository.Update(ctx, book); err != nil {
		return nil, fmt.Errorf("update book: %w", err)
	}

	notifyBookChange(ctx, u.notifier, book.ID, userID, "UPDATE_BOOK", "book", book)
	return book, nil
}

func (u *bookUsecase) DeleteBook(ctx context.Context, userID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	book, err := u.bookRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// 只有创建人可以删除账本，成员只能退出（当前版本不提供退出）
	if book.CrtUserID != userID {
		return domain.ErrNotBookMember
	}

	if err := u.bookRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete book: %w", err)
	}

	notifyBookChange(ctx, u.notifier, id, userID, "DELETE_BOOK", "book", id)
	return nil
}

func (u *bookUsecase) JoinBook(ctx context.Context, userID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 分享码即账本 ID；加入前确认账本仍然存在
	if _, err := u.bookRepository.GetByID(ctx, id); err != nil {
		return err
	}
	if err := u.bookRepository.AddMember(ctx, id, userID); err != nil {
		return fmt.Errorf("join book: %w", err)
	}
	return nil
}

func (u *bookUsecase) ShareBook(ctx context.Context, userID, id string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.requireMember(ctx, id, userID); err != nil {
		return "", err
	}
	return id, nil
}

func (u *bookUsecase) requireMember(ctx context.Context, bookID, userID string) error {
	ok, err := u.bookRepository.IsMember(ctx, bookID, userID)
	if err != nil {
		return fmt.Errorf("check book membership: %w", err)
	}
	if !ok {
		return domain.ErrNotBookMember
	}
	return nil
}

// notifyBookChange 推送账本变更；通知失败不影响写入结果。
func notifyBookChange(ctx context.Context, notifier domain.BillingNotifier, bookID, actorID, changeType, entity string, payload any) {
	if notifier == nil {
		return
	}
	notifier.NotifyBookChange(ctx, bookID, actorID, changeType, entity, payload)
}
