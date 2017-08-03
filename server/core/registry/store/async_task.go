//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package store

import (
	"context"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"time"
)

type AsyncTask interface {
	Do(ctx context.Context) error
}

type AsyncTasker interface {
	AddTask(ctx context.Context, task AsyncTask) error
	Run()
	Stop()
}

type LeaseAsyncTask struct {
	LeaseID int64
	TTL     int64
}

func (lat *LeaseAsyncTask) Do(ctx context.Context) (err error) {
	lat.TTL, err = registry.GetRegisterCenter().LeaseRenew(ctx, lat.LeaseID)
	if err != nil {
		util.LOGGER.Errorf(err, "renew lease %d failed", lat.LeaseID)
	}
	return
}

type LeaseAsyncTasker struct {
	Key        string
	TTL        int64
	firstReady chan struct{}
	err        chan error
	uniq       *util.UniQueue
	goroutine  *util.GoRoutine
}

func (lat *LeaseAsyncTasker) AddTask(ctx context.Context, task AsyncTask) error {
	err := lat.uniq.Put(ctx, task)
	if err != nil {
		return err
	}
	select {
	case err = <-lat.err:
		return err
	case <-lat.firstReady:
	}
	return nil
}

func (lat *LeaseAsyncTasker) Run() {
	lat.goroutine.Do(func(stopCh <-chan struct{}) {
		for {
			select {
			case <-stopCh:
				util.LOGGER.Debugf("LeaseAsyncTasker is stopped, Key %s TTL %d", lat.Key, lat.TTL)
				return
			default:
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				task := lat.uniq.Get(ctx)
				if task == nil {
					continue
				}
				at := task.(*LeaseAsyncTask)
				err := at.Do(context.Background())
				if err != nil {
					select {
					case <-lat.err:
						lat.err <- err
					default:
						lat.err <- err
					}
					continue
				}
				lat.TTL = at.TTL

				lat.closeCh(lat.firstReady)
			}
		}
	})
}

func (lat *LeaseAsyncTasker) closeCh(c chan struct{}) {
	select {
	case <-c:
	default:
		close(c)
	}
}

func (lat *LeaseAsyncTasker) Stop() {
	lat.uniq.Close()
	lat.goroutine.Close(true)
	lat.closeCh(lat.firstReady)
	close(lat.err)
}

func NewLeaseAsyncTask(leaseID int64) *LeaseAsyncTask {
	return &LeaseAsyncTask{
		LeaseID: leaseID,
	}
}

func NewLeaseAsyncTasker(key string) *LeaseAsyncTasker {
	return &LeaseAsyncTasker{
		Key:        key,
		firstReady: make(chan struct{}),
		err:        make(chan error, 1),
		uniq:       util.NewUniQueue(),
		goroutine:  util.NewGo(make(chan struct{})),
	}
}
