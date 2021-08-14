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

package diagnose

import (
	"testing"

	model2 "github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

func TestAbstractCompareHolder_Compare(t *testing.T) {
	services := model2.MicroserviceSlice{
		model2.NewMicroservice(&model2.KV{Key: "1", Rev: 1, Value: &registry.MicroService{ServiceId: "1"}}), // greater
		model2.NewMicroservice(&model2.KV{Key: "2", Rev: 1, Value: &registry.MicroService{ServiceId: "2"}}), // mismatch
		model2.NewMicroservice(&model2.KV{Key: "4", Rev: 2, Value: &registry.MicroService{ServiceId: "4"}}), // pass
	}
	kvs := []*mvccpb.KeyValue{
		{Key: []byte("2"), ModRevision: 2, Value: []byte(`{"ServiceID":"22"}`)},
		{Key: []byte("3"), ModRevision: 3, Value: []byte(`{"ServiceID":"3"}`)}, // less
		{Key: []byte("4"), ModRevision: 2, Value: []byte(`{"ServiceID":"4"}`)},
		{Key: []byte("5"), ModRevision: 4, Value: []byte(`xxxx`)},
	}
	service := ServiceCompareHolder{Cache: services, Kvs: kvs}
	rs := service.Compare()
	if rs == nil {
		t.Fatalf("TestAbstractCompareHolder_Compare failed")
	}
	if len(rs.Results) != 3 {
		t.Fatalf("TestAbstractCompareHolder_Compare failed")
	}
	if rs.Name != "service" ||
		rs.Results[greater][0] != "//(1)" ||
		rs.Results[mismatch][0] != "//(2)" ||
		!(rs.Results[less][0] == "unknown" || rs.Results[less][1] == "unknown") {
		t.Fatalf("TestAbstractCompareHolder_Compare failed, %v", rs)
	}
}
