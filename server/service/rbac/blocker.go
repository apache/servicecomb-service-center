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
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	MaxAttempts = 2

	BlockInterval = 1 * time.Hour
)

var BanTime = 1 * time.Hour

type Client struct {
	limiter   *rate.Limiter
	Key       string
	Banned    bool
	ReleaseAt time.Time //at this time client can be allow to attempt to do something
}

var clients sync.Map

func BannedList() []*Client {
	cs := make([]*Client, 0)
	clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)
		if client.Banned && time.Now().After(client.ReleaseAt) {
			client.Banned = false
			client.ReleaseAt = time.Time{}
			return true
		}
		cs = append(cs, client)
		return true
	})
	return cs
}

//CountFailure can cause a client banned
// it use time/rate to allow certainty failure,
//but will ban client if rate limiter can not accept failures
func CountFailure(key string) {
	var c interface{}
	var client *Client
	var ok bool
	now := time.Now()
	if c, ok = clients.Load(key); !ok {
		client = &Client{
			Key:       key,
			limiter:   rate.NewLimiter(rate.Every(BlockInterval), MaxAttempts),
			ReleaseAt: time.Time{},
		}
		clients.Store(key, client)
	} else {
		client = c.(*Client)
	}

	allow := client.limiter.AllowN(time.Now(), 1)
	if !allow {
		client.Banned = true
		client.ReleaseAt = now.Add(BanTime)
	}
}

//IsBanned check if a client is banned, and if client ban time expire,
//it will release the client from banned status
func IsBanned(key string) bool {
	var c interface{}
	var client *Client
	var ok bool
	if c, ok = clients.Load(key); !ok {
		return false
	}
	client = c.(*Client)
	if client.Banned && time.Now().After(client.ReleaseAt) {
		client.Banned = false
		client.ReleaseAt = time.Time{}
	}
	return client.Banned
}
