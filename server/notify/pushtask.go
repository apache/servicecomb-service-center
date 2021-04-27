package notify

import (
	"context"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/patrickmn/go-cache"
	"runtime"
	"sync"
	"time"
)

var (
	once    sync.Once
	PushCfg PushTaskEngine
)

type PushTaskEngine struct {
	cacheChan      chan string
	tasks          *cache.Cache
	pushC          *PushCache
	workerNum      int
	pushInterval   int
	newTaskTimeout int
}

func NewPushTaskEngine() *PushTaskEngine {
	once.Do(func() {
		PushCfg.workerNum = runtime.NumCPU()
		PushCfg.cacheChan = make(chan string, 100)
		PushCfg.newTaskTimeout = 3
		PushCfg.tasks = cache.New(0, 1*time.Second)
		PushCfg.pushC = NewPushCache()
		PushCfg.pushInterval=6
		PushCfg.tasks.OnEvicted(func(k string, v interface{}) {
			instances, rev, tenant := PushCfg.pushC.GetInstances(k)
			evt := NewInstanceEvents(k, tenant, rev, instances)
			err := Center().Publish(evt)
			if err != nil {
				log.Error(fmt.Sprintf("failed to publish instances"),err)
			}
		})
	})
	for i := 1; i <= PushCfg.workerNum; i++ {
		gopool.Go(func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case taskInfo, ok := <-PushCfg.cacheChan:
					if ok {
						_, exist := PushCfg.tasks.Get(taskInfo)
						if !exist {
							PushCfg.tasks.Set(taskInfo, struct{}{}, time.Duration(PushCfg.pushInterval)*time.Second)
						}
					}
				}
			}
		})
	}
	return &PushCfg
}

func (t *PushTaskEngine) AddTask(domainProject string, service string, rev int64, instance *pb.WatchInstanceResponse) error {
	t.pushC.AddInstance(domainProject, service, rev, instance)
	select {
	case t.cacheChan <- service:
		return nil
	case <-time.After(time.Duration(t.newTaskTimeout) * time.Second):
		return errors.New("pushTask's channel in full.")
	}
	return nil
}
