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

package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/cache"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	simple "github.com/apache/servicecomb-service-center/pkg/time"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/apache/servicecomb-service-center/server/syncernotify"
)

const (
	Provider = "p"
	Consumer = "c"
)

// InstanceEventHandler is the handler to handle events
//as instance registry or instance delete, and notify syncer
type InstanceEventHandler struct {
}

func (h InstanceEventHandler) Type() string {
	return model.CollectionInstance
}

func (h InstanceEventHandler) OnEvent(evt sd.MongoEvent) {
	action := evt.Type
	if evt.Type == discovery.EVT_UPDATE {
		return
	}
	instance := evt.Value.(model.Instance)
	providerID := instance.Instance.ServiceId
	providerInstanceID := instance.Instance.InstanceId
	domainProject := instance.Domain + "/" + instance.Project
	ctx := util.SetDomainProject(context.Background(), instance.Domain, instance.Project)
	res, ok := cache.GetServiceByID(providerID)
	var err error
	if !ok {
		filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(providerID))
		res, err = getService(ctx, filter)
		if err != nil {
			log.Error(fmt.Sprintf("caught [%s] instance[%s/%s] event, endpoints %v, get provider's file failed from db\n",
				action, providerID, providerInstanceID, instance.Instance.Endpoints), err)
		}
	}
	if res == nil {
		return
	}
	microService := res.Service
	switch action {
	case discovery.EVT_INIT:
		metrics.ReportInstances(instance.Domain, increaseOne)
		frameworkName, frameworkVersion := getFramework(microService)
		metrics.ReportFramework(instance.Domain, instance.Project, frameworkName, frameworkVersion, increaseOne)
		return
	case discovery.EVT_CREATE:
		metrics.ReportInstances(instance.Domain, increaseOne)
	case discovery.EVT_DELETE:
		metrics.ReportInstances(instance.Domain, decreaseOne)
	}
	if !syncernotify.GetSyncerNotifyCenter().Closed() {
		NotifySyncerInstanceEvent(evt, microService)
	}
	consumerIDS, _, err := getAllConsumerIds(ctx, microService)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s][%s/%s/%s/%s]'s consumerIDs failed",
			providerID, microService.Environment, microService.AppId, microService.ServiceName, microService.Version), err)
		return
	}
	PublishInstanceEvent(evt, domainProject, discovery.MicroServiceToKey(domainProject, microService), consumerIDS)
}

func NewInstanceEventHandler() *InstanceEventHandler {
	return &InstanceEventHandler{}
}

func PublishInstanceEvent(evt sd.MongoEvent, domainProject string, serviceKey *discovery.MicroServiceKey, subscribers []string) {
	if len(subscribers) == 0 {
		return
	}
	response := &discovery.WatchInstanceResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "Watch instance successfully."),
		Action:   string(evt.Type),
		Key:      serviceKey,
		Instance: evt.Value.(model.Instance).Instance,
	}
	for _, consumerID := range subscribers {
		evt := event.NewInstanceEventWithTime(consumerID, domainProject, -1, simple.FromTime(time.Now()), response)
		err := event.Center().Fire(evt)
		if err != nil {
			log.Error(fmt.Sprintf("publish event[%v] into channel failed", evt), err)
		}
	}
}

func NotifySyncerInstanceEvent(event sd.MongoEvent, microService *discovery.MicroService) {
	instance := event.Value.(model.Instance).Instance
	log.Info(fmt.Sprintf("instanceId : %s and serviceId : %s in NotifySyncerInstanceEvent", instance.InstanceId, instance.ServiceId))
	instanceKey := util.StringJoin([]string{datasource.InstanceKeyPrefix, event.Value.(model.Instance).Domain,
		event.Value.(model.Instance).Project, instance.ServiceId, instance.InstanceId}, datasource.SPLIT)

	instanceKv := dump.KV{
		Key:   instanceKey,
		Value: instance,
	}

	dumpInstance := dump.Instance{
		KV:    &instanceKv,
		Value: instance,
	}
	serviceKey := util.StringJoin([]string{datasource.ServiceKeyPrefix, event.Value.(model.Instance).Domain,
		event.Value.(model.Instance).Project, instance.ServiceId}, datasource.SPLIT)
	serviceKv := dump.KV{
		Key:   serviceKey,
		Value: microService,
	}

	dumpService := dump.Microservice{
		KV:    &serviceKv,
		Value: microService,
	}

	instEvent := &dump.WatchInstanceChangedEvent{
		Action:   string(event.Type),
		Service:  &dumpService,
		Instance: &dumpInstance,
	}
	syncernotify.GetSyncerNotifyCenter().AddEvent(instEvent)

	log.Debug(fmt.Sprintf("success to add instance change event action [%s], instanceKey : %s to event queue", instEvent.Action, instanceKey))
}

func getAllConsumerIds(ctx context.Context, provider *discovery.MicroService) (allow []string, deny []string, err error) {
	if provider == nil || len(provider.ServiceId) == 0 {
		return nil, nil, fmt.Errorf("invalid provider")
	}

	//todo 删除服务，最后实例推送有误差
	providerRules, ok := cache.GetRulesByServiceID(provider.ServiceId)
	if !ok {
		filter := mutil.NewBasicFilter(ctx, mutil.ServiceID(provider.ServiceId))
		providerRules, err = getRules(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
	}

	allow, deny, err = getConsumerIDsWithFilter(ctx, provider, providerRules)
	if err != nil && !errors.Is(err, datasource.ErrNoData) {
		return nil, nil, err
	}
	return allow, deny, nil
}

func getConsumerIDsWithFilter(ctx context.Context, provider *discovery.MicroService, rules []*model.Rule) (allow []string, deny []string, err error) {
	serviceDeps, ok := cache.GetProviderServiceOfDeps(provider)
	if !ok {
		filter := mutil.NewFilter(
			mutil.DependencyRuleType(Provider),
			mutil.ServiceKeyTenant(util.ParseDomainProject(ctx)),
			mutil.ServiceKeyAppID(provider.AppId),
			mutil.ServiceKeyServiceName(provider.ServiceName),
			mutil.ServiceKeyServiceVersion(provider.Version),
		)
		serviceDeps, err = getMicroServiceDependency(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
	}
	consumerIDs := make([]string, len(serviceDeps.Dependency))
	for _, serviceKeys := range serviceDeps.Dependency {
		id, ok := cache.GetServiceID(ctx, serviceKeys)
		if !ok {
			id, err = findServiceID(ctx, serviceKeys)
			if err != nil {
				return nil, nil, err
			}
		}
		consumerIDs = append(consumerIDs, id)
	}
	return filterAllService(ctx, consumerIDs, rules)
}

func findServiceID(ctx context.Context, key *discovery.MicroServiceKey) (string, error) {
	filter := mutil.NewBasicFilter(
		ctx,
		mutil.ServiceEnv(key.Environment),
		mutil.ServiceAppID(key.AppId),
		mutil.ServiceServiceName(key.ServiceName),
		mutil.ServiceVersion(key.Version),
	)
	id, err := getServiceID(ctx, filter)
	if err != nil && !errors.Is(err, datasource.ErrNoData) {
		return "", err
	}
	if len(id) == 0 && len(key.Alias) != 0 {
		filter = mutil.NewBasicFilter(
			ctx,
			mutil.ServiceEnv(key.Environment),
			mutil.ServiceAppID(key.AppId),
			mutil.ServiceAlias(key.Alias),
			mutil.ServiceVersion(key.Version),
		)
		return getServiceID(ctx, filter)
	}
	return id, nil
}

func getServiceID(ctx context.Context, filter interface{}) (serviceID string, err error) {
	svc, err := getService(ctx, filter)
	if err != nil {
		return
	}
	if svc != nil {
		serviceID = svc.Service.ServiceId
		return
	}
	return
}

func filterAllService(ctx context.Context, consumerIDs []string, rules []*model.Rule) (allow []string, deny []string, err error) {
	l := len(consumerIDs)
	if l == 0 || len(rules) == 0 {
		return consumerIDs, nil, nil
	}

	allowIdx, denyIdx := 0, l
	consumers := make([]string, l)
	for _, consumerID := range consumerIDs {
		ok, err := filter(ctx, rules, consumerID)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			consumers[allowIdx] = consumerID
			allowIdx++
		} else {
			denyIdx--
			consumers[denyIdx] = consumerID
		}
	}
	return consumers[:allowIdx], consumers[denyIdx:], nil
}

func filter(ctx context.Context, rules []*model.Rule, consumerID string) (bool, error) {
	consumer, ok := cache.GetServiceByID(consumerID)
	if !ok {
		var err error
		filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(consumerID))
		consumer, err = getService(ctx, filter)
		if err != nil {
			return false, err
		}
	}

	if len(rules) == 0 {
		return true, nil
	}
	domain := util.ParseDomainProject(ctx)
	project := util.ParseProject(ctx)
	filter := mutil.NewDomainProjectFilter(domain, project, mutil.ServiceServiceID(consumerID))
	tags, err := getTags(ctx, filter)
	if err != nil {
		return false, err
	}
	matchErr := mutil.MatchRules(rules, consumer.Service, tags)
	if matchErr != nil {
		if matchErr.Code == discovery.ErrPermissionDeny {
			return false, nil
		}
		return false, matchErr
	}
	return true, nil
}
