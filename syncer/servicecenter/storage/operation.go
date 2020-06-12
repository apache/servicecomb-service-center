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

package storage

import (
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/coreos/etcd/clientv3"
)

var (
	// mappingsKey the key of instances mapping in etcd
	mappingsKey = "/syncer/v1/mappings"
	// servicesKey the key of service in etcd
	servicesKey = "/syncer/v1/services"
	// instancesKey the key of instance in etcd
	instancesKey = "/syncer/v1/instances"
)

func putServiceOp(serviceId string, data []byte) clientv3.Op {
	return clientv3.OpPut(servicesKey+"/"+serviceId, util.BytesToStringWithNoCopy(data))
}

func getServicesOp() clientv3.Op {
	return clientv3.OpGet(servicesKey, clientv3.WithPrefix())
}

func deleteServiceOp(serviceId string) clientv3.Op {
	return clientv3.OpDelete(servicesKey + "/" + serviceId)
}

func putInstanceOp(instanceID string, data []byte) clientv3.Op {
	return clientv3.OpPut(instancesKey+"/"+instanceID, util.BytesToStringWithNoCopy(data))
}

func getInstancesOp() clientv3.Op {
	return clientv3.OpGet(instancesKey, clientv3.WithPrefix())
}

func deleteInstanceOp(instanceID string) clientv3.Op {
	return clientv3.OpDelete(instancesKey + "/" + instanceID)
}

func putMappingOp(cluster, mappingID string, data []byte) clientv3.Op {
	return clientv3.OpPut(mappingsKey+"/"+cluster+"/"+mappingID, util.BytesToStringWithNoCopy(data))
}

func getClusterMappingsOp(cluster string) clientv3.Op {
	return clientv3.OpGet(mappingsKey+"/"+cluster, clientv3.WithPrefix())
}

func getAllMappingsOp() clientv3.Op {
	return clientv3.OpGet(mappingsKey, clientv3.WithPrefix())
}

func delMappingOp(cluster, mappingID string) clientv3.Op {
	return clientv3.OpDelete(mappingsKey + "/" + cluster + "/" + mappingID)
}
