/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package etcd

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"sync"
	"time"
)

func (ds *DataSource) NewDLock(key string, ttl int64, wait bool) (datasource.DLock, error) {
	if len(key) == 0 {
		return nil, nil
	}
	key = fmt.Sprintf("%s%s", RootPath, key)
	if ttl < 1 {
		ttl = DefaultLockTTL
	}

	now := time.Now()
	l := &DLock{
		key:      key,
		ctx:      context.Background(),
		ttl:      ttl,
		id:       fmt.Sprintf("%v-%v-%v", hostname, pid, now.Format("20060102-15:04:05.999999999")),
		createAt: now,
		mutex:    &sync.Mutex{},
	}
	var err error
	for try := 1; try <= DefaultRetryTimes; try++ {
		err = l.Lock(wait)
		if err == nil {
			return l, nil
		}
		if !wait {
			break
		}
	}
	// failed
	log.Errorf(err, "Lock key %s failed, id=%s", l.key, l.id)
	return nil, err
}
