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
package backoff

import (
	"math"
	"time"
)

var DefaultBackoff Backoff = &PowerBackoff{
	MaxDelay:  30 * time.Second,
	InitDelay: 1 * time.Second,
	Factor:    1.6,
}

type Backoff interface {
	Delay(retries int) time.Duration
}

// delay = min(MaxDelay, InitDelay * power(Factor, retries))
type PowerBackoff struct {
	MaxDelay  time.Duration
	InitDelay time.Duration
	Factor    float64
}

func (pb *PowerBackoff) Delay(retries int) time.Duration {
	if retries <= 0 {
		return pb.InitDelay
	}

	return time.Duration(math.Min(float64(pb.MaxDelay), float64(pb.InitDelay)*math.Pow(pb.Factor, float64(retries))))
}

func GetBackoff() Backoff {
	return DefaultBackoff
}
