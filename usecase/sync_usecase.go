package usecase

import (
	"context"
	"fmt"
	"time"

	"shadmin/domain"
)

type syncUsecase struct {
	syncRepository domain.SyncRepository
	contextTimeout time.Duration
}

func NewSyncUsecase(syncRepository domain.SyncRepository, timeout time.Duration) domain.SyncUseCase {
	return &syncUsecase{
		syncRepository: syncRepository,
		contextTimeout: timeout,
	}
}

// GetChanges 返回 since（毫秒）之后的增量变更；limit 默认 100，上限 500。
func (u *syncUsecase) GetChanges(ctx context.Context, userID string, sinceMs int64, limit int) (*domain.SyncChangesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if limit <= 0 {
		limit = defaultSyncLimit
	}
	if limit > maxSyncLimit {
		limit = maxSyncLimit
	}

	result, err := u.syncRepository.QueryChanges(ctx, userID, sinceMs, limit)
	if err != nil {
		return nil, fmt.Errorf("query sync changes: %w", err)
	}
	return result, nil
}
