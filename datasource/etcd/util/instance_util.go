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
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetLeaseID(ctx context.Context, domainProject string, serviceID string, instanceID string) (int64, error) {
	opts := append(FromContext(ctx),
		client.WithStrKey(path.GenerateInstanceLeaseKey(domainProject, serviceID, instanceID)))
	resp, err := kv.Store().Lease().Search(ctx, opts...)
	if err != nil {
		return -1, err
	}
	if len(resp.Kvs) <= 0 {
		return -1, nil
	}
	leaseID, _ := strconv.ParseInt(resp.Kvs[0].Value.(string), 10, 64)
	return leaseID, nil
}

func GetInstance(ctx context.Context, domainProject string, serviceID string, instanceID string) (*pb.MicroServiceInstance, error) {
	key := path.GenerateInstanceKey(domainProject, serviceID, instanceID)
	opts := append(FromContext(ctx), client.WithStrKey(key))

	resp, err := kv.Store().Instance().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs[0].Value.(*pb.MicroServiceInstance), nil
}

func InstanceExist(ctx context.Context, domainProject string, serviceID string, instanceID string) (bool, error) {
	key := path.GenerateInstanceKey(domainProject, serviceID, instanceID)
	opts := append(FromContext(ctx), client.WithStrKey(key))
	resp, err := kv.Store().Instance().Search(ctx, opts...)
	if err != nil {
		return false, err
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}

func FormatRevision(revs, counts []int64) (s string) {
	for i, rev := range revs {
		s += fmt.Sprintf("%d.%d,", rev, counts[i])
	}
	return fmt.Sprintf("%x", sha1.Sum(util.StringToBytesWithNoCopy(s)))
}

func GetAllInstancesOfOneService(ctx context.Context, domainProject string, serviceID string) ([]*pb.MicroServiceInstance, error) {
	key := path.GenerateInstanceKey(domainProject, serviceID, "")
	opts := append(FromContext(ctx), client.WithStrKey(key), client.WithPrefix())
	resp, err := kv.Store().Instance().Search(ctx, opts...)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s instances failed", serviceID), err)
		return nil, err
	}

	instances := make([]*pb.MicroServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		instances = append(instances, kv.Value.(*pb.MicroServiceInstance))
	}
	return instances, nil
}

func GetInstanceCountOfOneService(ctx context.Context, domainProject string, serviceID string) (int64, error) {
	key := path.GenerateInstanceKey(domainProject, serviceID, "")
	opts := append(FromContext(ctx),
		client.WithStrKey(key),
		client.WithPrefix(),
		client.WithCountOnly())
	resp, err := kv.Store().Instance().Search(ctx, opts...)
	if err != nil {
		log.Error(fmt.Sprintf("get number of service[%s]'s instances failed", serviceID), err)
		return 0, err
	}
	return resp.Count, nil
}

type EndpointIndexValue struct {
	ServiceID  string
	InstanceID string
}

func ParseEndpointIndexValue(value []byte) EndpointIndexValue {
	endpointValue := EndpointIndexValue{}
	tmp := util.BytesToStringWithNoCopy(value)
	splitedTmp := strings.Split(tmp, "/")
	endpointValue.ServiceID = splitedTmp[0]
	endpointValue.InstanceID = splitedTmp[1]
	return endpointValue
}

func DeleteServiceAllInstances(ctx context.Context, serviceID string) error {
	domainProject := util.ParseDomainProject(ctx)

	instanceLeaseKey := path.GenerateInstanceLeaseKey(domainProject, serviceID, "")
	resp, err := kv.Store().Lease().Search(ctx,
		client.WithStrKey(instanceLeaseKey),
		client.WithPrefix(),
		client.WithNoCache())
	if err != nil {
		log.Error(fmt.Sprintf("delete all of service[%s]'s instances failed: get instance lease failed", serviceID), err)
		return err
	}
	if resp.Count <= 0 {
		log.Warn(fmt.Sprintf("service[%s] has no deployment of instance.", serviceID))
		return nil
	}
	for _, v := range resp.Kvs {
		leaseID, _ := strconv.ParseInt(v.Value.(string), 10, 64)
		err := client.Instance().LeaseRevoke(ctx, leaseID)
		if err != nil {
			log.Error("", err)
		}
	}
	return nil
}

func QueryServiceInstancesKvs(ctx context.Context, serviceID string, rev int64) ([]*sd.KeyValue, error) {
	domainProject := util.ParseDomainProject(ctx)
	key := path.GenerateInstanceKey(domainProject, serviceID, "")
	resp, err := kv.Store().Instance().Search(ctx,
		client.WithStrKey(key),
		client.WithPrefix(),
		client.WithRev(rev))
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s instances with revision %d failed",
			serviceID, rev), err)
		return nil, err
	}
	return resp.Kvs, nil
}

func UpdateInstance(ctx context.Context, domainProject string, instance *pb.MicroServiceInstance) *errsvc.Error {
	leaseID, err := GetLeaseID(ctx, domainProject, instance.ServiceId, instance.InstanceId)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	if leaseID == -1 {
		return pb.NewError(pb.ErrInstanceNotExists, "Instance's leaseId not exist.")
	}

	instance.ModTimestamp = strconv.FormatInt(time.Now().Unix(), 10)
	data, err := json.Marshal(instance)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	key := path.GenerateInstanceKey(domainProject, instance.ServiceId, instance.InstanceId)

	resp, err := client.Instance().TxnWithCmp(ctx,
		[]client.PluginOp{client.OpPut(
			client.WithStrKey(key),
			client.WithValue(data),
			client.WithLease(leaseID))},
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(path.GenerateServiceKey(domainProject, instance.ServiceId))),
			client.CmpNotEqual, 0)},
		nil)
	if err != nil {
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		return pb.NewError(pb.ErrInstanceNotExists, "Instance does not exist.")
	}
	return nil
}

func AppendFindResponse(ctx context.Context, index int64, resp *pb.Response, instances []*pb.MicroServiceInstance,
	updatedResult *[]*pb.FindResult, notModifiedResult *[]int64, failedResult **pb.FindFailedResult) {
	if code := resp.GetCode(); code != pb.ResponseSuccess {
		if *failedResult == nil {
			*failedResult = &pb.FindFailedResult{
				Error: pb.NewError(code, resp.GetMessage()),
			}
		}
		(*failedResult).Indexes = append((*failedResult).Indexes, index)
		return
	}
	iv, _ := ctx.Value(util.CtxRequestRevision).(string)
	ov, _ := ctx.Value(util.CtxResponseRevision).(string)
	if len(iv) > 0 && iv == ov {
		*notModifiedResult = append(*notModifiedResult, index)
		return
	}
	*updatedResult = append(*updatedResult, &pb.FindResult{
		Index:     index,
		Instances: instances,
		Rev:       ov,
	})
}
