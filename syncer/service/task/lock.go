package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	datasourcedlock "github.com/apache/servicecomb-service-center/datasource/dlock"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/service/dlock"
	"github.com/go-chassis/foundation/gopool"
)

type DistributedLock struct {
	heartbeatDuration time.Duration
	ttl               int64
	do                func(ctx context.Context)
	key               string
	isLock            bool
}

func (dl *DistributedLock) LockDo() {
	gopool.Go(func(goctx context.Context) {
		log.Info("start lock key " + dl.key)
		ticker := time.NewTicker(dl.heartbeatDuration)
		var ctx context.Context
		var cancel context.CancelFunc
		failCount := 0
		for {
			select {
			case <-ticker.C:
				if !dl.isLock {
					err := dlock.TryLock(dl.key, dl.ttl)
					if err != nil {
						continue
					}

					ctx, cancel = context.WithCancel(context.Background())
					dl.do(ctx)
					dl.isLock = true
					continue
				}

				err := dlock.Renew(dl.key)
				if err == nil {
					log.Info(fmt.Sprintf("renew lock %s success", dl.key))
					continue
				}

				if !errors.Is(err, datasourcedlock.ErrDLockNotExists) {
					failCount++
					log.Error("renew lock failed", err)
					if failCount == 5 {
						log.Warn("renew lock failed 5 times, release lock")
						cancel()
						dl.isLock = false
						failCount = 0
					}
					continue
				}
			case <-goctx.Done():
				ticker.Stop()
				cancel()
				log.Info(fmt.Sprintf("release lock %s", dl.key))
				return
			}

		}
	})
}

func (dl *DistributedLock) tryLock() {
	err := dlock.TryLock(dl.key, dl.ttl)
	if err != nil {
		log.Warn(fmt.Sprintf("try lock failed, %s", err.Error()))
	} else {
		dl.isLock = true
	}
}
