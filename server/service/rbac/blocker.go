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

package rbac

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	accountsvc "github.com/apache/servicecomb-service-center/server/service/account"

	"golang.org/x/time/rate"
)

const (
	MaxAttempts = 2

	BlockInterval = 1 * time.Hour
)

var BanTime = 1 * time.Hour

type LoginFailureLimiter struct {
	limiter *rate.Limiter
	Key     string
}

var clients sync.Map

//TryLockAccount try to lock the account login attempt
// it use time/rate to allow certainty failure,
//it will ban client if rate limiter can not accept failures
func TryLockAccount(key string) {
	var c interface{}
	var l *LoginFailureLimiter
	var ok bool
	if c, ok = clients.Load(key); !ok {
		l = &LoginFailureLimiter{
			Key:     key,
			limiter: rate.NewLimiter(rate.Every(BlockInterval), MaxAttempts),
		}
		clients.Store(key, l)
	} else {
		l = c.(*LoginFailureLimiter)
	}

	allow := l.limiter.AllowN(time.Now(), 1)
	if !allow {
		err := accountsvc.Ban(context.TODO(), key)
		if err != nil {
			log.Error(fmt.Sprintf("can not ban account %s", key), err)
		}
	}
}

//IsBanned check if a client is banned, and if client ban time expire,
//it will release the client from banned status
//use account name plus ip as key will maximum reduce the client conflicts
func IsBanned(key string) bool {
	IsBanned, err := accountsvc.IsBanned(context.TODO(), key)
	if err != nil {
		log.Error("can not check lock list, so return banned for security concern", err)
		return true
	}
	return IsBanned
}
