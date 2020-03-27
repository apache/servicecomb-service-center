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

package serf

import (
	"time"

	"github.com/hashicorp/serf/serf"
)

// QueryOption func
type QueryOption func(*serf.QueryParam)

// WithFilterNodes returns node filter query option
func WithFilterNodes(nodes ...string) QueryOption {
	return func(p *serf.QueryParam) { p.FilterNodes = nodes }
}

// WithFilterTags returns tags filter query option
func WithFilterTags(tags map[string]string) QueryOption {
	return func(p *serf.QueryParam) { p.FilterTags = tags }
}

// WithRequestAck returns request ack query option
func WithRequestAck(ack bool) QueryOption {
	return func(p *serf.QueryParam) { p.RequestAck = ack }
}

// WithRelayFactor returns relay factor query option
func WithRelayFactor(num uint8) QueryOption {
	return func(p *serf.QueryParam) { p.RelayFactor = num }
}

// WithTimeout returns timeout query option
func WithTimeout(timeout time.Duration) QueryOption {
	return func(p *serf.QueryParam) { p.Timeout = timeout }
}
