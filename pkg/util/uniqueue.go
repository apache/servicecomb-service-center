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
package util

import (
	"errors"
	"golang.org/x/net/context"
	"math"
	"sync"
)

const DEFAULT_MAX_BUFFER_SIZE = 1000

type UniQueue struct {
	size   int
	buffer chan interface{}
	queue  chan interface{}
	close  chan struct{}
	once   sync.Once
}

func (uq *UniQueue) Get(ctx context.Context) interface{} {
	select {
	case <-uq.close:
		return nil
	case <-ctx.Done():
		return nil
	case item := <-uq.queue:
		return item
	}
}

func (uq *UniQueue) Chan() <-chan interface{} {
	return uq.queue
}

func (uq *UniQueue) Put(ctx context.Context, value interface{}) error {
	uq.once.Do(func() {
		go uq.do()
	})
	select {
	case <-uq.close:
		return errors.New("channel is closed")
	default:
		defer RecoverAndReport()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case uq.buffer <- value:
		}
	}
	return nil
}

func (uq *UniQueue) do() {
	for {
		select {
		case item, ok := <-uq.buffer:
			if !ok {
				return
			}
			select {
			case _, ok := <-uq.queue:
				if !ok {
					return
				}
			default:
			}
			// 1. drop the old item
			// 2. if queue is empty
			uq.sendToQueue(item)
		}
	}
}

func (uq *UniQueue) sendToQueue(item interface{}) {
	defer RecoverAndReport()
	select {
	case <-uq.close:
		return
	default:
		select {
		case uq.queue <- item:
		default:
			// drop item
		}
	}
}

func (uq *UniQueue) Close() {
	select {
	case <-uq.close:
	default:
		close(uq.close)
		close(uq.queue)
		close(uq.buffer)
	}
}

func newUniQueue(size int) (*UniQueue, error) {
	if size <= 0 || size >= math.MaxInt32 {
		return nil, errors.New("invalid buffer size")
	}
	return &UniQueue{
		size:   size,
		queue:  make(chan interface{}, 1),
		buffer: make(chan interface{}, size),
		close:  make(chan struct{}),
	}, nil
}

func NewUniQueue() (uq *UniQueue) {
	uq, _ = newUniQueue(DEFAULT_MAX_BUFFER_SIZE)
	return
}
