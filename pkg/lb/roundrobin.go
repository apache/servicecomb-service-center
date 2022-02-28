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

package lb

import "sync/atomic"

type RoundRobinLB struct {
	Endpoints []string
	index     int32
}

func (lb *RoundRobinLB) Next() string {
	l := len(lb.Endpoints)
	if l == 0 {
		return ""
	}
	c := atomic.LoadInt32(&lb.index)
	if c >= int32(l)-1 {
		atomic.StoreInt32(&lb.index, 0)
		return lb.Endpoints[0]
	} else if atomic.CompareAndSwapInt32(&lb.index, c, c+1) {
		return lb.Endpoints[c+1]
	}
	return lb.Endpoints[atomic.LoadInt32(&lb.index)]
}

func NewRoundRobinLB(endpoints []string) *RoundRobinLB {
	lb := &RoundRobinLB{
		Endpoints: make([]string, len(endpoints)),
		index:     -1,
	}
	copy(lb.Endpoints, endpoints)
	return lb
}
