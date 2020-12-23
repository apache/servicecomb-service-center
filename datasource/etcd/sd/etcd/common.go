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
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/value"
	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

const (
	revMethodName = "Revision"
)

var closedCh = make(chan struct{})

func init() {
	close(closedCh)
}

func FromEtcdKeyValue(dist *sd.KeyValue, src *mvccpb.KeyValue, parser value.Parser) (err error) {
	dist.ClusterName = etcd.Configuration().ClusterName
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

func ParseResourceToEtcdKeyValue(dist *sd.KeyValue, src *sdcommon.Resource, parser value.Parser) (err error) {
	dist.ClusterName = etcd.Configuration().ClusterName
	dist.Key = util.StringToBytesWithNoCopy(src.Key)
	dist.Version = src.Version
	dist.CreateRevision = src.CreateRevision
	dist.ModRevision = src.ModRevision
	if parser == nil {
		return
	}
	dist.Value, err = parser.Unmarshal(src.Value.([]byte))
	return
}
