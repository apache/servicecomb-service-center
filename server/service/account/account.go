package account

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"

	"github.com/apache/servicecomb-service-center/datasource"
)

func IsBanned(ctx context.Context, key string) (bool, error) {
	lock, err := datasource.GetAccountLockManager().GetLock(ctx, key)
	if err != nil {
		if err == datasource.ErrAccountLockNotExist {
			return false, nil
		}
		return false, err
	}
	if lock.ReleaseAt < time.Now().Unix() {
		err := datasource.GetAccountLockManager().DeleteLock(ctx, key)
		if err != nil {
			log.Errorf(err, "remove lock failed")
			return false, datasource.ErrCannotReleaseLock
		}
		log.Info(fmt.Sprintf("release lock for %s", key))
		return false, nil
	}
	if lock.Status == datasource.StatusBanned {
		return true, nil
	}
	return false, nil
}
func Ban(ctx context.Context, key string) error {
	return datasource.GetAccountLockManager().Ban(ctx, key)
}

func UpsertLock(ctx context.Context, lock *datasource.AccountLock) error {
	return datasource.GetAccountLockManager().UpsertLock(ctx, lock)
}

func ListLock(ctx context.Context) ([]*datasource.AccountLock, int64, error) {
	return datasource.GetAccountLockManager().ListLock(ctx)
}

func DeleteLock(ctx context.Context, key string) error {
	return datasource.GetAccountLockManager().DeleteLock(ctx, key)
}

func DeleteLockList(ctx context.Context, keys []string) error {
	return datasource.GetAccountLockManager().DeleteLockList(ctx, keys)
}
