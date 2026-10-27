package usecase

import (
	"context"
	"fmt"
	"time"

	"shadmin/domain"
)

type billUsecase struct {
	billRepository domain.BillRepository
	bookRepository domain.BookRepository
	userRepository domain.UserRepository
	notifier       domain.BillingNotifier
	contextTimeout time.Duration
}

func NewBillUsecase(
	billRepository domain.BillRepository,
	bookRepository domain.BookRepository,
	userRepository domain.UserRepository,
	notifier domain.BillingNotifier,
	timeout time.Duration,
) domain.BillUseCase {
	return &billUsecase{
		billRepository: billRepository,
		bookRepository: bookRepository,
		userRepository: userRepository,
		notifier:       notifier,
		contextTimeout: timeout,
	}
}

func (u *billUsecase) CreateBill(ctx context.Context, userID string, req *domain.CreateBillRequest) (*domain.Bill, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := u.requireMember(ctx, req.BookID, userID); err != nil {
		return nil, err
	}
	if err := validateBillTime(req.Time); err != nil {
		return nil, err
	}

	bill := &domain.Bill{
		ID:       req.ID,
		BookID:   req.BookID,
		Money:    req.Money,
		Type:     req.Type,
		Category: req.Category,
		CrtUser:  userID,
		Time:     req.Time,
		Remark:   req.Remark,
		Images:   req.Images,
	}
	if err := u.billRepository.Create(ctx, bill); err != nil {
		return nil, fmt.Errorf("create bill: %w", err)
	}

	notifyBookChange(ctx, u.notifier, bill.BookID, userID, "ADD_BILL", "bill", bill)
	return bill, nil
}

func (u *billUsecase) GetBill(ctx context.Context, userID, id string) (*domain.Bill, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	bill, err := u.billRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.requireMember(ctx, bill.BookID, userID); err != nil {
		return nil, err
	}
	u.fillCreatorNames(ctx, []*domain.Bill{bill})
	return bill, nil
}

func (u *billUsecase) ListBills(ctx context.Context, userID string, filter domain.BillQueryFilter) (*domain.BillPagedResult, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if filter.BookID == "" {
		return nil, fmt.Errorf("book_id is required: %w", domain.ErrNotBookMember)
	}
	if err := u.requireMember(ctx, filter.BookID, userID); err != nil {
		return nil, err
	}

	result, err := u.billRepository.Query(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list bills: %w", err)
	}
	u.fillCreatorNames(ctx, result.List)
	return result, nil
}

func (u *billUsecase) UpdateBill(ctx context.Context, userID, id string, req *domain.UpdateBillRequest) (*domain.Bill, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	bill, err := u.billRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.requireMember(ctx, bill.BookID, userID); err != nil {
		return nil, err
	}
	if req.Time != nil {
		if err := validateBillTime(*req.Time); err != nil {
			return nil, err
		}
	}

	if req.Money != nil {
		bill.Money = *req.Money
	}
	if req.Type != nil {
		bill.Type = *req.Type
	}
	if req.Category != nil {
		bill.Category = *req.Category
	}
	if req.Time != nil {
		bill.Time = *req.Time
	}
	if req.Remark != nil {
		bill.Remark = *req.Remark
	}
	if req.Images != nil {
		bill.Images = req.Images
	}

	if err := u.billRepository.Update(ctx, bill); err != nil {
		return nil, fmt.Errorf("update bill: %w", err)
	}

	notifyBookChange(ctx, u.notifier, bill.BookID, userID, "UPDATE_BILL", "bill", bill)
	return bill, nil
}

func (u *billUsecase) DeleteBill(ctx context.Context, userID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	bill, err := u.billRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.requireMember(ctx, bill.BookID, userID); err != nil {
		return err
	}

	if err := u.billRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete bill: %w", err)
	}

	notifyBookChange(ctx, u.notifier, bill.BookID, userID, "DELETE_BILL", "bill", id)
	return nil
}

func (u *billUsecase) requireMember(ctx context.Context, bookID, userID string) error {
	ok, err := u.bookRepository.IsMember(ctx, bookID, userID)
	if err != nil {
		return fmt.Errorf("check book membership: %w", err)
	}
	if !ok {
		return domain.ErrNotBookMember
	}
	return nil
}

// validateBillTime 校验账单日期：接受 YYYY-MM-DD，也兼容客户端同步时带来的
// YYYY-MM-DD HH:MM:SS。空字符串由请求体 binding 拦截，不在此处理。
func validateBillTime(value string) error {
	layouts := []string{"2006-01-02", "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, value); err == nil {
			return nil
		}
	}
	return domain.ErrInvalidBillTime
}

// fillCreatorNames 为列表补充创建人用户名；查询失败时保留 ID，不阻断列表返回。
func (u *billUsecase) fillCreatorNames(ctx context.Context, bills []*domain.Bill) {
	if u.userRepository == nil || len(bills) == 0 {
		return
	}

	names := make(map[string]string)
	for _, bill := range bills {
		if bill.CrtUser == "" {
			continue
		}
		if _, cached := names[bill.CrtUser]; cached {
			continue
		}
		name := ""
		if user, err := u.userRepository.GetByID(ctx, bill.CrtUser); err == nil && user != nil {
			name = user.Username
		}
		names[bill.CrtUser] = name
	}

	for _, bill := range bills {
		bill.CrtUserName = names[bill.CrtUser]
	}
}
