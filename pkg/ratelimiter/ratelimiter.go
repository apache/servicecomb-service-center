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

package ratelimiter

import (
	"sync"
	"time"
)

type LeakyBucket struct {
	startTime       time.Time
	capacity        int64
	quantum         int64
	interval        time.Duration
	mutex           sync.Mutex
	available       int64
	availableTicker int64
}

func NewLeakyBucket(fillInterval time.Duration, capacity, quantum int64) *LeakyBucket {
	if fillInterval <= 0 {
		panic("leaky bucket fill interval is not > 0")
	}
	if capacity <= 0 {
		panic("leaky bucket capacity is not > 0")
	}
	if quantum <= 0 {
		panic("leaky bucket quantum is not > 0")
	}
	return &LeakyBucket{
		startTime: time.Now(),
		capacity:  capacity,
		quantum:   quantum,
		available: capacity,
		interval:  fillInterval,
	}
}

func (leakyBucket *LeakyBucket) Wait(count int64) {
	if d := leakyBucket.Take(count); d > 0 {
		time.Sleep(d)
	}
}

func (leakyBucket *LeakyBucket) MaximumWaitDuration(count int64, maxWait time.Duration) bool {
	d, ok := leakyBucket.MaximumTakeDuration(count, maxWait)
	if d > 0 {
		time.Sleep(d)
	}
	return ok
}

const sleepForever time.Duration = 0x7fffffffffffffff

func (leakyBucket *LeakyBucket) Take(count int64) time.Duration {
	d, _ := leakyBucket.take(time.Now(), count, sleepForever)
	return d
}

func (leakyBucket *LeakyBucket) MaximumTakeDuration(count int64, maxWait time.Duration) (time.Duration, bool) {
	return leakyBucket.take(time.Now(), count, maxWait)
}

func (leakyBucket *LeakyBucket) Rate() float64 {
	return 1e9 * float64(leakyBucket.quantum) / float64(leakyBucket.interval)
}

func (leakyBucket *LeakyBucket) take(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool) {
	if count <= 0 {
		return 0, true
	}
	leakyBucket.mutex.Lock()
	defer leakyBucket.mutex.Unlock()

	currentTick := leakyBucket.adjust(now)
	avail := leakyBucket.available - count
	if avail >= 0 {
		leakyBucket.available = avail
		return 0, true
	}
	endTick := currentTick + (-avail+leakyBucket.quantum-1)/leakyBucket.quantum
	endTime := leakyBucket.startTime.Add(time.Duration(endTick) * leakyBucket.interval)
	waitTime := endTime.Sub(now)
	if waitTime > maxWait {
		return 0, false
	}
	leakyBucket.available = avail
	return waitTime, true
}

func (leakyBucket *LeakyBucket) adjust(now time.Time) (currentTick int64) {
	currentTick = int64(now.Sub(leakyBucket.startTime) / leakyBucket.interval)

	if leakyBucket.available >= leakyBucket.capacity {
		return
	}
	leakyBucket.available += (currentTick - leakyBucket.availableTicker) * leakyBucket.quantum
	if leakyBucket.available > leakyBucket.capacity {
		leakyBucket.available = leakyBucket.capacity
	}
	leakyBucket.availableTicker = currentTick
	return
}
