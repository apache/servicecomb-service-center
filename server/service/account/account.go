package account

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/rbac"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
)

const (
	defaultReleaseLockAfter     = 15 * time.Minute
	defaultRetainLockHistoryFor = 20 * time.Minute
)

func IsBanned(ctx context.Context, key string) (bool, error) {
	lock, err := GetLock(ctx, key)
	if err != nil {
		if err == rbac.ErrAccountLockNotExist {
			return false, nil
		}
		return false, err
	}
	if lock.ReleaseAt < time.Now().Unix() {
		err = DeleteLock(ctx, key)
		if err != nil {
			log.Error("remove lock failed", err)
			return false, rbac.ErrCannotReleaseLock
		}
		log.Info(fmt.Sprintf("release lock for %s", key))
		return false, nil
	}
	if lock.Status == rbac.StatusBanned {
		return true, nil
	}
	return false, nil
}

func Ban(ctx context.Context, key string) error {
	return Lock(ctx, key, rbac.StatusBanned)
}

func Lock(ctx context.Context, key, status string) error {
	duration := config.GetDuration("rbac.retainLockHistoryFor", defaultRetainLockHistoryFor)
	if status == rbac.StatusBanned {
		duration = config.GetDuration("rbac.releaseLockAfter", defaultReleaseLockAfter)
	}
	lock := &rbac.Lock{
		Key:       key,
		Status:    status,
		ReleaseAt: time.Now().Add(duration).Unix(),
	}
	return rbac.Instance().UpsertLock(ctx, lock)
}

func GetLock(ctx context.Context, key string) (*rbac.Lock, error) {
	return rbac.Instance().GetLock(ctx, key)
}

func ListLock(ctx context.Context) ([]*rbac.Lock, int64, error) {
	return rbac.Instance().ListLock(ctx)
}

func DeleteLock(ctx context.Context, key string) error {
	return rbac.Instance().DeleteLock(ctx, key)
}

func DeleteLockList(ctx context.Context, keys []string) error {
	return rbac.Instance().DeleteLockList(ctx, keys)
}
