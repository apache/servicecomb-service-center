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
	"sort"
	"strconv"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
)

type InstanceSlice []*pb.MicroServiceInstance

func (s InstanceSlice) Len() int {
	return len(s)
}

func (s InstanceSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s InstanceSlice) Less(i, j int) bool {
	return s[i].InstanceId < s[j].InstanceId
}

func Statistics(ctx context.Context, withShared bool) (*pb.Statistics, error) {
	result := &pb.Statistics{
		Services:  &pb.StService{},
		Instances: &pb.StInstance{},
		Apps:      &pb.StApp{},
	}
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	filter := bson.M{db.ColumnDomain: domain, db.ColumnProject: project}

	services, err := GetMicroServices(ctx, filter)
	if err != nil {
		return nil, err
	}

	var svcIDs []string
	var svcKeys []*pb.MicroServiceKey
	for _, svc := range services {
		svcIDs = append(svcIDs, svc.ServiceId)
		svcKeys = append(svcKeys, datasource.TransServiceToKey(util.ParseDomainProject(ctx), svc))
	}
	svcIDToNonVerKey := datasource.SetStaticServices(result, svcKeys, svcIDs, withShared)

	respGetInstanceCountByDomain := make(chan datasource.GetInstanceCountByDomainResponse, 1)
	gopool.Go(func(_ context.Context) {
		getInstanceCountByDomain(ctx, svcIDToNonVerKey, respGetInstanceCountByDomain)
	})

	instances, err := GetInstances(ctx, filter)
	if err != nil {
		return nil, err
	}
	var instIDs []string
	for _, inst := range instances {
		instIDs = append(instIDs, inst.Instance.ServiceId)
	}
	datasource.SetStaticInstances(result, svcIDToNonVerKey, instIDs)
	data := <-respGetInstanceCountByDomain
	close(respGetInstanceCountByDomain)
	if data.Err != nil {
		return nil, data.Err
	}
	result.Instances.CountByDomain = data.CountByDomain
	return result, nil
}

func GetInstance(ctx context.Context, serviceID string, instanceID string) (*db.Instance, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnInstance, db.ColumnServiceID}):  serviceID,
		StringBuilder([]string{db.ColumnInstance, db.ColumnInstanceID}): instanceID}
	findRes, err := client.GetMongoClient().FindOne(ctx, db.CollectionInstance, filter)
	if err != nil {
		return nil, err
	}
	var instance *db.Instance
	if findRes.Err() != nil {
		//not get any service,not db err
		return nil, nil
	}
	err = findRes.Decode(&instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func GetInstances(ctx context.Context, filter bson.M) ([]*db.Instance, error) {
	res, err := client.GetMongoClient().Find(ctx, db.CollectionInstance, filter)
	if err != nil {
		return nil, err
	}
	var instances []*db.Instance
	for res.Next(ctx) {
		var tmp *db.Instance
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		instances = append(instances, tmp)
	}
	return instances, nil
}

func GetInstanceCountOfOneService(ctx context.Context, serviceID string) (int64, error) {
	filter := GeneratorServiceInstanceFilter(ctx, serviceID)
	count, err := client.GetMongoClient().Count(ctx, db.CollectionInstance, filter)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func GeneratorServiceInstanceFilter(ctx context.Context, serviceID string) bson.M {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	return bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnInstance, db.ColumnServiceID}): serviceID}
}

func GetInstancesByServiceID(ctx context.Context, serviceID string) ([]*pb.MicroServiceInstance, error) {
	var res []*pb.MicroServiceInstance
	var cacheUnavailable bool
	cacheInstances := sd.Store().Instance().Cache().GetIndexData(serviceID)
	for _, instID := range cacheInstances {
		inst, ok := sd.Store().Instance().Cache().Get(instID).(db.Instance)
		if !ok {
			cacheUnavailable = true
			break
		}
		res = append(res, inst.Instance)
	}
	if cacheUnavailable || len(res) == 0 {
		res, err := InstancesFilter(ctx, []string{serviceID})
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return res, nil
}

func InstancesFilter(ctx context.Context, serviceIDs []string) ([]*pb.MicroServiceInstance, error) {
	var instances []*pb.MicroServiceInstance
	if len(serviceIDs) == 0 {
		return instances, nil
	}
	resp, err := client.GetMongoClient().Find(ctx, db.CollectionInstance, bson.M{StringBuilder([]string{db.ColumnInstance, db.ColumnServiceID}): bson.M{"$in": serviceIDs}}, &options.FindOptions{
		Sort: bson.M{StringBuilder([]string{db.ColumnInstance, db.ColumnVersion}): -1}})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, datasource.ErrNoData
	}
	for resp.Next(ctx) {
		var instance db.Instance
		err := resp.Decode(&instance)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance.Instance)
	}
	return instances, nil
}

func GetAllInstancesOfOneService(ctx context.Context, serviceID string) ([]*pb.MicroServiceInstance, error) {
	filter := GeneratorServiceInstanceFilter(ctx, serviceID)
	res, err := client.GetMongoClient().Find(ctx, db.CollectionInstance, filter)
	if err != nil {
		return nil, err
	}
	var instances []*pb.MicroServiceInstance
	for res.Next(ctx) {
		var tmp db.Instance
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		instances = append(instances, tmp.Instance)
	}
	return instances, nil
}

func PreProcessRegisterInstance(ctx context.Context, instance *pb.MicroServiceInstance) *pb.Error {
	if len(instance.Status) == 0 {
		instance.Status = pb.MSI_UP
	}

	if len(instance.InstanceId) == 0 {
		instance.InstanceId = uuid.Generator().GetInstanceID(ctx)
	}

	instance.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	instance.ModTimestamp = instance.Timestamp

	// 这里应该根据租约计时
	renewalInterval := apt.RegistryDefaultLeaseRenewalinterval
	retryTimes := apt.RegistryDefaultLeaseRetrytimes
	if instance.HealthCheck == nil {
		instance.HealthCheck = &pb.HealthCheck{
			Mode:     pb.CHECK_BY_HEARTBEAT,
			Interval: renewalInterval,
			Times:    retryTimes,
		}
	} else {
		// Health check对象仅用于呈现服务健康检查逻辑，如果CHECK_BY_PLATFORM类型，表明由sidecar代发心跳，实例120s超时
		switch instance.HealthCheck.Mode {
		case pb.CHECK_BY_HEARTBEAT:
			d := instance.HealthCheck.Interval * (instance.HealthCheck.Times + 1)
			if d <= 0 {
				return pb.NewError(pb.ErrInvalidParams, "invalid 'healthCheck' settings in request body.")
			}
		case pb.CHECK_BY_PLATFORM:
			// 默认120s
			instance.HealthCheck.Interval = renewalInterval
			instance.HealthCheck.Times = retryTimes
		}
	}

	filter := GeneratorServiceFilter(ctx, instance.ServiceId)
	microservice, err := GetService(ctx, filter)
	if err != nil {
		return pb.NewError(pb.ErrServiceNotExists, "invalid 'serviceID' in request body.")
	}
	instance.Version = microservice.Service.Version
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

func KeepAliveLease(ctx context.Context, request *pb.HeartbeatRequest) *pb.Error {
	_, err := heartbeat.Instance().Heartbeat(ctx, request)
	if err != nil {
		return pb.NewError(pb.ErrInstanceNotExists, err.Error())
	}
	return nil
}

func GetHeartbeatFunc(ctx context.Context, domainProject string, instancesHbRst chan<- *pb.InstanceHbRst, element *pb.HeartbeatSetElement) func(context.Context) {
	return func(_ context.Context) {
		hbRst := &pb.InstanceHbRst{
			ServiceId:  element.ServiceId,
			InstanceId: element.InstanceId,
			ErrMessage: "",
		}

		req := &pb.HeartbeatRequest{
			InstanceId: element.InstanceId,
			ServiceId:  element.ServiceId,
		}

		err := KeepAliveLease(ctx, req)
		if err != nil {
			hbRst.ErrMessage = err.Error()
			log.Error(fmt.Sprintf("heartbeat set failed %s %s", element.ServiceId, element.InstanceId), err)
		}
		instancesHbRst <- hbRst
	}
}

func UpdateInstanceStatus(ctx context.Context, instance *pb.MicroServiceInstance) *pb.Error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnInstance, db.ColumnServiceID}):  instance.ServiceId,
		StringBuilder([]string{db.ColumnInstance, db.ColumnInstanceID}): instance.InstanceId}
	_, err := client.GetMongoClient().Update(ctx, db.CollectionInstance, filter, bson.M{"$set": bson.M{"instance.motTimestamp": strconv.FormatInt(time.Now().Unix(), 10), "instance.status": instance.Status}})
	if err != nil {
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	return nil
}

func UpdateInstanceProperties(ctx context.Context, instance *pb.MicroServiceInstance) *pb.Error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	filter := bson.M{
		db.ColumnDomain:  domain,
		db.ColumnProject: project,
		StringBuilder([]string{db.ColumnInstance, db.ColumnServiceID}):  instance.ServiceId,
		StringBuilder([]string{db.ColumnInstance, db.ColumnInstanceID}): instance.InstanceId}
	_, err := client.GetMongoClient().Update(ctx, db.CollectionInstance, filter, bson.M{"$set": bson.M{"instance.motTimestamp": strconv.FormatInt(time.Now().Unix(), 10), "instance.properties": instance.Properties}})
	if err != nil {
		return pb.NewError(pb.ErrUnavailableBackend, err.Error())
	}
	return nil
}

func FormatRevision(consumerServiceID string, instances []*pb.MicroServiceInstance) (string, error) {
	if instances == nil {
		return fmt.Sprintf("%x", sha1.Sum(util.StringToBytesWithNoCopy(consumerServiceID))), nil
	}
	copyInstance := make([]*pb.MicroServiceInstance, len(instances))
	copy(copyInstance, instances)
	sort.Sort(InstanceSlice(copyInstance))
	data, err := json.Marshal(copyInstance)
	if err != nil {
		log.Error("fail to marshal instance json", err)
		return "", err
	}
	s := fmt.Sprintf("%s.%x", consumerServiceID, sha1.Sum(data))
	return fmt.Sprintf("%x", sha1.Sum(util.StringToBytesWithNoCopy(s))), nil
}

func RegistryInstance(ctx context.Context, request *pb.RegisterInstanceRequest) (*pb.RegisterInstanceResponse, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	remoteIP := util.GetIPFromContext(ctx)
	instance := request.Instance
	instanceID := instance.InstanceId
	data := &db.Instance{
		Domain:      domain,
		Project:     project,
		RefreshTime: time.Now(),
		Instance:    instance,
	}

	instanceFlag := fmt.Sprintf("endpoints %v, host '%s', serviceID %s",
		instance.Endpoints, instance.HostName, instance.ServiceId)

	insertRes, err := client.GetMongoClient().Insert(ctx, db.CollectionInstance, data)
	if err != nil {
		log.Error(fmt.Sprintf("register instance failed %s instanceID %s operator %s", instanceFlag, instanceID, remoteIP), err)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponse(pb.ErrUnavailableBackend, err.Error()),
		}, err
	}

	log.Info(fmt.Sprintf("register instance %s, instanceID %s, operator %s",
		instanceFlag, insertRes.InsertedID, remoteIP))
	heartbeatRequest := pb.HeartbeatRequest{
		ServiceId:  instance.ServiceId,
		InstanceId: instance.InstanceId,
	}
	aliveErr := KeepAliveLease(ctx, &heartbeatRequest)
	if aliveErr != nil {
		log.Error(fmt.Sprintf("failed to send heartbeat after registering instance, instance %s operator %s", instanceFlag, remoteIP), err)
		return &pb.RegisterInstanceResponse{
			Response: pb.CreateResponseWithSCErr(aliveErr),
		}, err
	}
	return &pb.RegisterInstanceResponse{
		Response:   pb.CreateResponse(pb.ResponseSuccess, "Register service instance successfully."),
		InstanceId: instanceID,
	}, nil
}
