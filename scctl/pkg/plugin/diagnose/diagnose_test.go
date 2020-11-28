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
	"fmt"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/go-chassis/cari/discovery"
)

func TestNewDiagnoseCommand(t *testing.T) {
	services := dump.MicroserviceSlice{
		dump.NewMicroservice(&dump.KV{Key: "1", Rev: 1,
			Value: &discovery.MicroService{
				ServiceId: "667570b6842411e89c66286ed488de36", AppId: "app", ServiceName: "name1", Version: "0.0.1",
			}}), // greater
		dump.NewMicroservice(&dump.KV{Key: "6", Rev: 1,
			Value: &discovery.MicroService{
				ServiceId: "667570b6842411e89c66286ed488de36", AppId: "app", ServiceName: "name2", Version: "0.0.1",
			}}), // greater
		dump.NewMicroservice(&dump.KV{Key: "2", Rev: 1, Value: &discovery.MicroService{ServiceId: "2"}}), // mismatch
		dump.NewMicroservice(&dump.KV{Key: "4", Rev: 2, Value: &discovery.MicroService{ServiceId: "4"}}), // pass
	}
	instances := dump.InstanceSlice{
		dump.NewInstance(&dump.KV{Key: "1", Rev: 1,
			Value: &discovery.MicroServiceInstance{
				ServiceId: "667570b6842411e89c66286ed488de36", InstanceId: "667570b6842411e89c66286ed488de36", Version: "0.0.1",
				Endpoints: []string{"rest://127.0.0.1:8080"},
			}}), // greater
		dump.NewInstance(&dump.KV{Key: "2", Rev: 1,
			Value: &discovery.MicroServiceInstance{
				ServiceId: "667570b6842411e89c66286ed488de36", InstanceId: "667570b6842411e89c66286ed488de36", Version: "0.0.1",
				Endpoints: []string{"rest://127.0.0.2:8080"},
			}}), // greater
	}
	kvs := []*mvccpb.KeyValue{
		{Key: []byte("2"), ModRevision: 2, Value: []byte(`{"ServiceID":"22"}`)},
		{Key: []byte("3"), ModRevision: 3, Value: []byte(`{"ServiceID":"3"}`)}, // less
		{Key: []byte("4"), ModRevision: 2, Value: []byte(`{"ServiceID":"4"}`)},
		{Key: []byte("5"), ModRevision: 4, Value: []byte(`xxxx`)},
	}

	//for {
	err, details := diagnose(&dump.Cache{Microservices: services, Instances: instances}, etcdResponse{service: kvs})
	if err == nil || len(details) == 0 {
		t.Fatalf("TestNewDiagnoseCommand failed")
	}
	fmt.Println(err)
	fmt.Println(details)
	//}
}
