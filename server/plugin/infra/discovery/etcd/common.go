// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

const (
	// force re-list
	DEFAULT_FORCE_LIST_INTERVAL = 4
	DEFAULT_METRICS_INTERVAL    = 30 * time.Second
	DEFAULT_COMPACT_TIMES       = 2
	DEFAULT_COMPACT_TIMEOUT     = 5 * time.Minute

	minWaitInterval = 1 * time.Second
	eventBlockSize  = 1000
	eventBusSize    = 1000
)

var closedCh = make(chan struct{})

func init() {
	close(closedCh)
}

func FromEtcdKeyValue(dist *discovery.KeyValue, src *mvccpb.KeyValue, parser pb.Parser) (err error) {
	dist.Key = src.Key
	dist.Version = src.Version
	dist.CreateRevision = src.CreateRevision
	dist.ModRevision = src.ModRevision
	if parser == nil {
		return
	}
	dist.Value, err = parser.Unmarshal(src.Value)
	return
}
