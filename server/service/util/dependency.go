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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend/store"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"strings"
)

func GetConsumersInCache(ctx context.Context, domainProject string, provider *pb.MicroService) ([]string, error) {
	// 查询所有consumer
	dr := NewProviderDependencyRelation(ctx, domainProject, provider)
	consumerIds, err := dr.GetDependencyConsumerIds()
	if err != nil {
		util.Logger().Errorf(err, "Get dependency consumerIds failed.%s", provider.ServiceId)
		return nil, err
	}
	return consumerIds, nil
}

func GetProvidersInCache(ctx context.Context, domainProject string, consumer *pb.MicroService) ([]string, error) {
	// 查询所有provider
	dr := NewConsumerDependencyRelation(ctx, domainProject, consumer)
	providerIds, err := dr.GetDependencyProviderIds()
	if err != nil {
		util.Logger().Errorf(err, "Get dependency providerIds failed.%s", consumer.ServiceId)
		return nil, err
	}
	return providerIds, nil
}

func GetConsumerIdsByProvider(ctx context.Context, domainProject string, provider *pb.MicroService) (allow []string, deny []string, _ error) {
	if provider == nil || len(provider.ServiceId) == 0 {
		return nil, nil, fmt.Errorf("invalid provider")
	}

	//todo 删除服务，最后实例推送有误差
	providerRules, err := GetRulesUtil(util.SetContext(util.CloneContext(ctx), "cacheOnly", "1"),
		domainProject, provider.ServiceId)
	if err != nil {
		return nil, nil, err
	}
	if len(providerRules) == 0 {
		return getConsumerIdsWithFilter(ctx, domainProject, provider, noFilter)
	}

	rf := RuleFilter{
		DomainProject: domainProject,
		Provider:      provider,
		ProviderRules: providerRules,
	}

	allow, deny, err = getConsumerIdsWithFilter(ctx, domainProject, provider, rf.Filter)
	if err != nil {
		return nil, nil, err
	}
	return allow, deny, nil
}

func getConsumerIdsWithFilter(ctx context.Context, domainProject string, provider *pb.MicroService,
	filter func(ctx context.Context, consumerId string) (bool, error)) (allow []string, deny []string, err error) {
	consumerIds, err := GetConsumersInCache(ctx, domainProject, provider)
	if err != nil {
		return nil, nil, err
	}
	return filterConsumerIds(ctx, consumerIds, filter)
}

func filterConsumerIds(ctx context.Context, consumerIds []string,
	filter func(ctx context.Context, consumerId string) (bool, error)) (allow []string, deny []string, err error) {
	l := len(consumerIds)
	if l == 0 {
		return nil, nil, nil
	}
	allowIdx, denyIdx := 0, l
	consumers := make([]string, l)
	for _, consumerId := range consumerIds {
		ok, err := filter(ctx, consumerId)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			consumers[allowIdx] = consumerId
			allowIdx++
		} else {
			denyIdx--
			consumers[denyIdx] = consumerId
		}
	}
	return consumers[:allowIdx], consumers[denyIdx:], nil
}

func noFilter(_ context.Context, _ string) (bool, error) {
	return true, nil
}

func GetProviderIdsByConsumer(ctx context.Context, domainProject string, service *pb.MicroService) (allow []string, deny []string, _ error) {
	providerIdsInCache, err := GetProvidersInCache(ctx, domainProject, service)
	if err != nil {
		return nil, nil, err
	}
	l := len(providerIdsInCache)
	rf := RuleFilter{
		DomainProject: domainProject,
	}
	allowIdx, denyIdx := 0, l
	providerIds := make([]string, l)
	copyCtx := util.SetContext(util.CloneContext(ctx), "cacheOnly", "1")
	for _, providerId := range providerIdsInCache {
		provider, err := GetService(ctx, domainProject, providerId)
		if provider == nil {
			continue
		}
		providerRules, err := GetRulesUtil(copyCtx, domainProject, provider.ServiceId)
		if err != nil {
			return nil, nil, err
		}
		if len(providerRules) == 0 {
			providerIds[allowIdx] = providerId
			allowIdx++
			continue
		}
		rf.Provider = provider
		rf.ProviderRules = providerRules
		ok, err := rf.Filter(ctx, service.ServiceId)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			providerIds[allowIdx] = providerId
			allowIdx++
		} else {
			denyIdx--
			providerIds[denyIdx] = providerId
		}
	}
	return providerIds[:allowIdx], providerIds[denyIdx:], nil
}

func ProviderDependencyRuleExist(ctx context.Context, provider *pb.MicroServiceKey, consumer *pb.MicroServiceKey) (bool, error) {
	targetDomainProject := provider.Tenant
	if len(targetDomainProject) == 0 {
		targetDomainProject = consumer.Tenant
	}
	providerKey := apt.GenerateProviderDependencyRuleKey(targetDomainProject, provider)
	consumers, err := TransferToMicroServiceDependency(ctx, providerKey)
	if err != nil {
		return false, err
	}
	if len(consumers.Dependency) != 0 {
		isEqual, err := containServiceDependency(consumers.Dependency, consumer)
		if err != nil {
			return false, err
		}
		if isEqual {
			//删除之前的依赖
			return true, nil
		}
	}
	return false, nil
}

func AddServiceVersionRule(ctx context.Context, domainProject string, consumer *pb.MicroService, provider *pb.MicroServiceKey) error {
	//创建依赖一致
	consumerKey := pb.MicroServiceToKey(domainProject, consumer)
	exist, err := ProviderDependencyRuleExist(ctx, provider, consumerKey)
	if exist || err != nil {
		return err
	}

	r := &pb.ConsumerDependency{
		Consumer:  consumerKey,
		Providers: []*pb.MicroServiceKey{provider},
		Override:  false,
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	id := util.StringJoin([]string{provider.AppId, provider.ServiceName}, "_")
	key := apt.GenerateConsumerDependencyQueueKey(domainProject, consumer.ServiceId, id)
	_, err = backend.Registry().Do(ctx, registry.PUT, registry.WithStrKey(key), registry.WithValue(data))
	if err != nil {
		return err
	}

	util.Logger().Infof("find request into dependency queue successfully, %s: %v", key, r)
	return nil
}

func TransferToMicroServiceDependency(ctx context.Context, key string) (*pb.MicroServiceDependency, error) {
	microServiceDependency := &pb.MicroServiceDependency{
		Dependency: []*pb.MicroServiceKey{},
	}

	opts := append(FromContext(ctx), registry.WithStrKey(key))
	res, err := store.Store().DependencyRule().Search(ctx, opts...)
	if err != nil {
		util.Logger().Errorf(nil, "Get dependency rule failed.")
		return nil, err
	}
	if len(res.Kvs) != 0 {
		err = json.Unmarshal(res.Kvs[0].Value, microServiceDependency)
		if err != nil {
			util.Logger().Errorf(nil, "Unmarshal res failed.")
			return nil, err
		}
	}
	return microServiceDependency, nil
}

func updateProviderDependencyRuleUtil(consumersOfProvideRules *pb.MicroServiceDependency, consumer *pb.MicroServiceKey, providerRuleKey string) (registry.PluginOp, error) {
	for key, consumerInner := range consumersOfProvideRules.Dependency {
		if ok := equalServiceDependency(consumerInner, consumer); ok {
			consumersOfProvideRules.Dependency = append(consumersOfProvideRules.Dependency[:key], consumersOfProvideRules.Dependency[key+1:]...)
			break
		}
	}
	if len(consumersOfProvideRules.Dependency) == 0 {
		util.Logger().Infof("delete dependency rule key is %s", providerRuleKey)
		return registry.OpDel(registry.WithStrKey(providerRuleKey)), nil
	} else {
		data, err := json.Marshal(consumersOfProvideRules)
		if err != nil {
			util.Logger().Errorf(nil, "update dependency key Marshal tmpValue failed.")
			return registry.PluginOp{}, err
		}
		util.Logger().Infof("put provider's dependency rule, key is  %s, value is %v.", providerRuleKey, consumersOfProvideRules)
		return registry.OpPut(registry.WithStrKey(providerRuleKey), registry.WithValue(data)), nil
	}
}

func equalServiceDependency(serviceA *pb.MicroServiceKey, serviceB *pb.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	if stringA == stringB {
		return true
	}
	return false
}

func diffServiceVersion(serviceA *pb.MicroServiceKey, serviceB *pb.MicroServiceKey) bool {
	stringA := toString(serviceA)
	stringB := toString(serviceB)
	if stringA != stringB &&
		stringA[:strings.LastIndex(stringA, "/")+1] == stringB[:strings.LastIndex(stringB, "/")+1] {
		return true
	}
	return false
}

func toString(in *pb.MicroServiceKey) string {
	return apt.GenerateProviderDependencyRuleKey(in.Tenant, in)
}

func parseAddOrUpdateRules(ctx context.Context, dep *Dependency) (newDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList []*pb.MicroServiceKey) {
	conKey := apt.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)

	oldProviderRules, err := TransferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		util.Logger().Errorf(err, "maintain dependency rule failed, consumer %s/%s/%s: get consumer depedency rule failed.",
			dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version)
		return
	}

	deleteDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	newDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	for _, tmpProviderRule := range dep.ProvidersRule {
		if ok, _ := containServiceDependency(oldProviderRules.Dependency, tmpProviderRule); ok {
			continue
		}

		if tmpProviderRule.ServiceName == "*" {
			newDependencyRuleList = append([]*pb.MicroServiceKey{}, tmpProviderRule)
			deleteDependencyRuleList = oldProviderRules.Dependency
			break
		}

		newDependencyRuleList = append(newDependencyRuleList, tmpProviderRule)
		old := isNeedUpdate(oldProviderRules.Dependency, tmpProviderRule)
		if old != nil {
			deleteDependencyRuleList = append(deleteDependencyRuleList, old)
		}
	}
	for _, oldProviderRule := range oldProviderRules.Dependency {
		if oldProviderRule.ServiceName == "*" {
			newDependencyRuleList = nil
			deleteDependencyRuleList = nil
			return
		}
		if ok, _ := containServiceDependency(deleteDependencyRuleList, oldProviderRule); !ok {
			existDependencyRuleList = append(existDependencyRuleList, oldProviderRule)
		}
	}

	dep.ProvidersRule = append(newDependencyRuleList, existDependencyRuleList...)
	return
}

func parseOverrideRules(ctx context.Context, dep *Dependency) (newDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList []*pb.MicroServiceKey) {
	conKey := apt.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)

	oldProviderRules, err := TransferToMicroServiceDependency(ctx, conKey)
	if err != nil {
		util.Logger().Errorf(err, "maintain dependency rule failed, consumer %s/%s/%s: get consumer depedency rule failed.",
			dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version)
		return
	}

	deleteDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	newDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(dep.ProvidersRule))
	existDependencyRuleList = make([]*pb.MicroServiceKey, 0, len(oldProviderRules.Dependency))
	for _, oldProviderRule := range oldProviderRules.Dependency {
		if ok, _ := containServiceDependency(dep.ProvidersRule, oldProviderRule); !ok {
			deleteDependencyRuleList = append(deleteDependencyRuleList, oldProviderRule)
		} else {
			existDependencyRuleList = append(existDependencyRuleList, oldProviderRule)
		}
	}
	for _, tmpProviderRule := range dep.ProvidersRule {
		if ok, _ := containServiceDependency(existDependencyRuleList, tmpProviderRule); !ok {
			newDependencyRuleList = append(newDependencyRuleList, tmpProviderRule)
		}
	}
	return
}

func syncDependencyRule(ctx context.Context, dep *Dependency, filter func(context.Context, *Dependency) (_, _, _ []*pb.MicroServiceKey)) error {
	//更新consumer的providers的值,consumer的版本是确定的
	consumerFlag := strings.Join([]string{dep.Consumer.AppId, dep.Consumer.ServiceName, dep.Consumer.Version}, "/")

	newDependencyRuleList, existDependencyRuleList, deleteDependencyRuleList := filter(ctx, dep)
	if len(newDependencyRuleList) == 0 && len(existDependencyRuleList) == 0 && len(deleteDependencyRuleList) == 0 {
		return nil
	}

	dep.err = make(chan error, 5)
	dep.chanNum = 0
	if len(deleteDependencyRuleList) != 0 {
		util.Logger().Infof("Delete dependency rule remove for consumer %s, %v, ", consumerFlag, deleteDependencyRuleList)
		dep.removedDependencyRuleList = deleteDependencyRuleList
		dep.RemoveConsumerOfProviderRule()
	}

	if len(newDependencyRuleList) != 0 {
		util.Logger().Infof("New dependency rule add for consumer %s, %v, ", consumerFlag, newDependencyRuleList)
		dep.NewDependencyRuleList = newDependencyRuleList
		dep.AddConsumerOfProviderRule()
	}

	conKey := apt.GenerateConsumerDependencyRuleKey(dep.DomainProject, dep.Consumer)
	err := dep.UpdateProvidersRuleOfConsumer(conKey)
	if err != nil {
		return err
	}

	if dep.chanNum != 0 {
		for tmpErr := range dep.err {
			dep.chanNum--
			if tmpErr != nil {
				return tmpErr
			}
			if 0 == dep.chanNum {
				close(dep.err)
			}
		}
	}
	return nil
}

func AddDependencyRule(ctx context.Context, dep *Dependency) error {
	return syncDependencyRule(ctx, dep, parseAddOrUpdateRules)
}

func CreateDependencyRule(ctx context.Context, dep *Dependency) error {
	return syncDependencyRule(ctx, dep, parseOverrideRules)
}

func isDependencyAll(dep *pb.MicroServiceDependency) bool {
	for _, serviceDep := range dep.Dependency {
		if serviceDep.ServiceName == "*" {
			return true
		}
	}
	return false
}

func isExist(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) bool {
	for _, tmp := range services {
		if equalServiceDependency(tmp, service) {
			return true
		}
	}
	return false
}

func deleteConsumerDepOfProviderRule(ctx context.Context, domainProject string, providerRule *pb.MicroServiceKey, deleteConsumer *pb.MicroServiceKey) (registry.PluginOp, error) {
	proKey := apt.GenerateProviderDependencyRuleKey(domainProject, providerRule)
	consumerDepRules, err := TransferToMicroServiceDependency(ctx, proKey)
	if err != nil {
		return registry.PluginOp{}, err
	}
	return deleteDepRuleUtil(proKey, consumerDepRules, deleteConsumer)
}

func deleteDepRuleUtil(key string, deps *pb.MicroServiceDependency, deleteDepRule *pb.MicroServiceKey) (registry.PluginOp, error) {
	for key, consumerDepRule := range deps.Dependency {
		if equalServiceDependency(consumerDepRule, deleteDepRule) {
			deps.Dependency = append(deps.Dependency[:key], deps.Dependency[key+1:]...)
			break
		}
	}
	data, err := json.Marshal(deps)
	if err != nil {
		return registry.PluginOp{}, err
	}
	return registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)), nil
}

func addDepRuleUtil(key string, deps *pb.MicroServiceDependency, updateDepRule *pb.MicroServiceKey) (registry.PluginOp, error) {
	deps.Dependency = append(deps.Dependency, updateDepRule)
	data, err := json.Marshal(deps)
	if err != nil {
		util.Logger().Errorf(err, "marshal consumerDepRules failed for delete consumer rule from provider rule's dep.")
		return registry.PluginOp{}, err
	}
	return registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)), nil
}

func updateDepRuleUtil(key string, deps *pb.MicroServiceDependency, updateDepRule *pb.MicroServiceKey) (string, registry.PluginOp, error) {
	oldRule := ""
	for _, serviceRule := range deps.Dependency {
		if diffServiceVersion(serviceRule, updateDepRule) {
			oldRule = serviceRule.Version
			serviceRule.Version = updateDepRule.Version
			break
		}
	}
	data, err := json.Marshal(deps)
	if err != nil {
		return oldRule, registry.PluginOp{}, err
	}
	return oldRule, registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)), nil
}

func isNeedUpdate(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) *pb.MicroServiceKey {
	for _, tmp := range services {
		if diffServiceVersion(tmp, service) {
			return tmp
		}
	}
	return nil
}

func containServiceDependency(services []*pb.MicroServiceKey, service *pb.MicroServiceKey) (bool, error) {
	if services == nil || service == nil {
		return false, errors.New("Invalid params input.")
	}
	for _, value := range services {
		rst := equalServiceDependency(service, value)
		if rst {
			return true, nil
		}
	}
	return false, nil
}

// fuzzyMatch: 是否使用模糊规则
func validateMicroServiceKey(in *pb.MicroServiceKey, fuzzyMatch bool) error {
	if fuzzyMatch {
		// provider的ServiceName, Version支持模糊规则
		return apt.ProviderMsValidator.Validate(in)
	} else {
		return apt.ConsumerMsValidator.Validate(in)
	}
}

func BadParamsResponse(detailErr string) *pb.CreateDependenciesResponse {
	util.Logger().Errorf(nil, "Request params is invalid.%s", detailErr)
	if len(detailErr) == 0 {
		detailErr = "Request params is invalid."
	}
	return &pb.CreateDependenciesResponse{
		Response: pb.CreateResponse(scerr.ErrInvalidParams, detailErr),
	}
}

func ParamsChecker(consumerInfo *pb.MicroServiceKey, providersInfo []*pb.MicroServiceKey) *pb.CreateDependenciesResponse {
	if err := validateMicroServiceKey(consumerInfo, false); err != nil {
		return BadParamsResponse(err.Error())
	}
	if providersInfo == nil {
		return BadParamsResponse("Invalid request body for provider info.")
	}
	flag := make(map[string]bool, len(providersInfo))
	for _, providerInfo := range providersInfo {
		//存在带*的情况，后面的数据就不校验了
		if providerInfo.ServiceName == "*" {
			util.Logger().Debugf("%s 's provider contains *.", consumerInfo.ServiceName)
			break
		}
		if len(providerInfo.AppId) == 0 {
			providerInfo.AppId = consumerInfo.AppId
		}
		if err := validateMicroServiceKey(providerInfo, true); err != nil {
			return BadParamsResponse(err.Error())
		}

		version := providerInfo.Version
		providerInfo.Version = ""
		if _, ok := flag[toString(providerInfo)]; ok {
			return BadParamsResponse("Invalid request body for provider info.Duplicate provider or (serviceName and appid is same).")
		} else {
			flag[toString(providerInfo)] = true
		}
		providerInfo.Version = version
	}
	return nil
}

type Dependency struct {
	ConsumerId                string
	DomainProject             string
	removedDependencyRuleList []*pb.MicroServiceKey
	NewDependencyRuleList     []*pb.MicroServiceKey
	err                       chan error
	chanNum                   int8
	Consumer                  *pb.MicroServiceKey
	ProvidersRule             []*pb.MicroServiceKey
}

func (dep *Dependency) RemoveConsumerOfProviderRule() {
	dep.chanNum++
	util.Go(dep.removeConsumerOfProviderRule)
}

func (dep *Dependency) removeConsumerOfProviderRule(ctx context.Context) {
	opts := make([]registry.PluginOp, 0, len(dep.removedDependencyRuleList))
	for _, providerRule := range dep.removedDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(providerRule.Tenant, providerRule)
		util.Logger().Debugf("This proProkey is %s.", proProkey)
		consumerValue, err := TransferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			dep.err <- err
			return
		}
		for key, tmp := range consumerValue.Dependency {
			if ok := equalServiceDependency(tmp, dep.Consumer); ok {
				consumerValue.Dependency = append(consumerValue.Dependency[:key], consumerValue.Dependency[key+1:]...)
				break
			}
			util.Logger().Debugf("tmp and dep.Consumer not equal, tmp %v, consumer %v", tmp, dep.Consumer)
		}
		//删除后，如果不存在依赖规则了，就删除该provider的依赖规则，如果有，则更新该依赖规则
		if len(consumerValue.Dependency) == 0 {
			opts = append(opts, registry.OpDel(registry.WithStrKey(proProkey)))
			continue
		}
		data, err := json.Marshal(consumerValue)
		if err != nil {
			util.Logger().Errorf(nil, "Marshal tmpValue failed.")
			dep.err <- err
			return
		}
		opts = append(opts, registry.OpPut(
			registry.WithStrKey(proProkey),
			registry.WithValue(data)))
	}
	if len(opts) != 0 {
		_, err := backend.Registry().Txn(ctx, opts)
		if err != nil {
			dep.err <- err
			return
		}
	}
	dep.err <- nil
}

func (dep *Dependency) AddConsumerOfProviderRule() {
	dep.chanNum++
	util.Go(dep.addConsumerOfProviderRule)
}

func (dep *Dependency) addConsumerOfProviderRule(ctx context.Context) {
	opts := []registry.PluginOp{}
	for _, providerRule := range dep.NewDependencyRuleList {
		proProkey := apt.GenerateProviderDependencyRuleKey(providerRule.Tenant, providerRule)
		tmpValue, err := TransferToMicroServiceDependency(ctx, proProkey)
		if err != nil {
			dep.err <- err
			return
		}
		tmpValue.Dependency = append(tmpValue.Dependency, dep.Consumer)

		data, errMarshal := json.Marshal(tmpValue)
		if errMarshal != nil {
			util.Logger().Errorf(nil, "Marshal tmpValue failed.")
			dep.err <- errors.New("Marshal tmpValue failed.")
			return
		}
		opts = append(opts, registry.OpPut(
			registry.WithStrKey(proProkey),
			registry.WithValue(data)))
		if providerRule.ServiceName == "*" {
			break
		}
	}
	if len(opts) != 0 {
		_, err := backend.Registry().Txn(ctx, opts)
		if err != nil {
			dep.err <- err
			return
		}
	}
	dep.err <- nil
}

func (dep *Dependency) UpdateProvidersRuleOfConsumer(conKey string) error {
	if len(dep.ProvidersRule) == 0 {
		_, err := backend.Registry().Do(context.TODO(),
			registry.DEL,
			registry.WithStrKey(conKey),
			)
		if err != nil {
			util.Logger().Errorf(nil, "Upload dependency rule failed.")
			return err
		}
		return nil
	}

	dependency := &pb.MicroServiceDependency{
		Dependency: dep.ProvidersRule,
	}
	data, err := json.Marshal(dependency)
	if err != nil {
		util.Logger().Errorf(nil, "Marshal tmpValue fialed.")
		return err
	}
	_, err = backend.Registry().Do(context.TODO(),
		registry.PUT,
		registry.WithStrKey(conKey),
		registry.WithValue(data))
	if err != nil {
		util.Logger().Errorf(nil, "Upload dependency rule failed.")
		return err
	}
	return nil
}

type DependencyRelation struct {
	ctx           context.Context
	domainProject string
	consumer      *pb.MicroService
	provider      *pb.MicroService
}

func NewProviderDependencyRelation(ctx context.Context, domainProject string, provider *pb.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, nil, provider)
}

func NewConsumerDependencyRelation(ctx context.Context, domainProject string, consumer *pb.MicroService) *DependencyRelation {
	return NewDependencyRelation(ctx, domainProject, consumer, nil)
}

func NewDependencyRelation(ctx context.Context, domainProject string, consumer *pb.MicroService, provider *pb.MicroService) *DependencyRelation {
	return &DependencyRelation{
		ctx:           ctx,
		domainProject: domainProject,
		consumer:      consumer,
		provider:      provider,
	}
}

func (dr *DependencyRelation) GetDependencyProviders(opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	keys, err := dr.getProviderKeys()
	if err != nil {
		return nil, err
	}
	services := make([]*pb.MicroService, 0, len(keys))
	op := toDependencyRelationFilterOpt(opts...)
	for _, key := range keys {
		if op.SameDomainProject && key.Tenant != dr.domainProject {
			continue
		}

		providerIds, err := dr.parseDependencyRule(key)
		if err != nil {
			return nil, err
		}

		if key.ServiceName == "*" {
			services = services[:0]
		}

		for _, providerId := range providerIds {
			provider, err := GetService(dr.ctx, key.Tenant, providerId)
			if err != nil {
				util.Logger().Warnf(nil, "Provider does not exist, %s/%s/%s",
					key.AppId, key.ServiceName, key.Version)
				continue
			}
			if op.NonSelf && providerId == dr.consumer.ServiceId {
				continue
			}
			services = append(services, provider)
		}

		if key.ServiceName == "*" {
			break
		}
	}
	return services, nil
}

func (dr *DependencyRelation) GetDependencyProviderIds() ([]string, error) {
	keys, err := dr.getProviderKeys()
	if err != nil {
		return nil, err
	}
	return dr.getDependencyProviderIds(keys)
}

func (dr *DependencyRelation) getProviderKeys() ([]*pb.MicroServiceKey, error) {
	if dr.consumer == nil {
		util.LOGGER.Infof("dr.consumer is nil ------->")
		return nil, fmt.Errorf("Invalid consumer")
	}
	consumerMicroServiceKey := pb.MicroServiceToKey(dr.domainProject, dr.consumer)

	conKey := apt.GenerateConsumerDependencyRuleKey(dr.domainProject, consumerMicroServiceKey)
	consumerDependency, err := TransferToMicroServiceDependency(dr.ctx, conKey)
	if err != nil {
		return nil, err
	}
	return consumerDependency.Dependency, nil
}

func (dr *DependencyRelation) getDependencyProviderIds(providerRules []*pb.MicroServiceKey) ([]string, error) {
	provideServiceIds := make([]string, 0, len(providerRules))
	for _, provider := range providerRules {
		serviceIds, err := dr.parseDependencyRule(provider)
		switch {
		case provider.ServiceName == "*":
			if err != nil {
				util.Logger().Errorf(err, "Add dependency failed, rely all service: get all services failed.")
				return provideServiceIds, err
			}
			return serviceIds, nil
		default:
			if err != nil {
				util.Logger().Errorf(err, "Get providerIds failed, service: %s/%s/%s",
					provider.AppId, provider.ServiceName, provider.Version)
				return provideServiceIds, err
			}
			if len(serviceIds) == 0 {
				util.Logger().Warnf(nil, "Get providerIds is empty, service: %s/%s/%s does not exist",
					provider.AppId, provider.ServiceName, provider.Version)
				continue
			}
			provideServiceIds = append(provideServiceIds, serviceIds...)
		}
	}
	return provideServiceIds, nil
}

func (dr *DependencyRelation) parseDependencyRule(dependencyRule *pb.MicroServiceKey) (serviceIds []string, err error) {
	opts := FromContext(dr.ctx)
	switch {
	case dependencyRule.ServiceName == "*":
		util.Logger().Infof("Rely all service, * type, consumerId %s", dr.consumer.ServiceId)
		splited := strings.Split(apt.GenerateServiceIndexKey(dependencyRule), "/")
		allServiceKey := util.StringJoin(splited[:len(splited)-3], "/") + "/"
		sopts := append(opts,
			registry.WithStrKey(allServiceKey),
			registry.WithPrefix())
		resp, err := store.Store().Service().Search(dr.ctx, sopts...)
		if err != nil {
			return nil, err
		}

		for _, kv := range resp.Kvs {
			serviceIds = append(serviceIds, util.BytesToStringWithNoCopy(kv.Value))
		}
	default:
		serviceIds, err = FindServiceIds(dr.ctx, dependencyRule.Version, dependencyRule)
	}
	return
}

func (dr *DependencyRelation) GetDependencyConsumers(opts ...DependencyRelationFilterOption) ([]*pb.MicroService, error) {
	consumerDependAllList, err := dr.getDependencyConsumersOfProvider()
	if err != nil {
		util.Logger().Errorf(err, "Get consumers of provider rule failed, %s", dr.provider.ServiceId)
		return nil, err
	}
	consumers := make([]*pb.MicroService, 0)
	op := toDependencyRelationFilterOpt(opts...)
	for _, consumer := range consumerDependAllList {
		if op.SameDomainProject && consumer.Tenant != dr.domainProject {
			continue
		}

		service, err := dr.getServiceByMicroServiceKey(consumer)
		if err != nil {
			return nil, err
		}
		if service == nil {
			util.Logger().Warnf(nil, "Consumer does not exist, %v", consumer)
			continue
		}

		if op.NonSelf && service.ServiceId == dr.provider.ServiceId {
			continue
		}

		consumers = append(consumers, service)
	}
	return consumers, nil
}

func (dr *DependencyRelation) getServiceByMicroServiceKey(service *pb.MicroServiceKey) (*pb.MicroService, error) {
	serviceId, err := GetServiceId(dr.ctx, service)
	if err != nil {
		return nil, err
	}
	if len(serviceId) == 0 {
		util.Logger().Warnf(nil, "Service not exist,%v", service)
		return nil, nil
	}
	return GetService(dr.ctx, service.Tenant, serviceId)
}

func (dr *DependencyRelation) GetDependencyConsumerIds() ([]string, error) {
	consumerDependAllList, err := dr.getDependencyConsumersOfProvider()
	if err != nil {
		return nil, err
	}
	consumerIds := make([]string, 0, len(consumerDependAllList))
	for _, consumer := range consumerDependAllList {
		consumerId, err := GetServiceId(dr.ctx, consumer)
		if err != nil {
			util.Logger().Errorf(err, "Get consumer failed, %v", consumer)
			return nil, err
		}
		if len(consumerId) == 0 {
			util.Logger().Warnf(nil, "Get consumer not exist, %v", consumer)
			continue
		}
		consumerIds = append(consumerIds, consumerId)
	}
	return consumerIds, nil

}

func (dr *DependencyRelation) getDependencyConsumersOfProvider() ([]*pb.MicroServiceKey, error) {
	if dr.provider == nil {
		util.LOGGER.Infof("dr.provider is nil ------->")
		return nil, fmt.Errorf("Invalid provider")
	}
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	consumerDependAllList, err := dr.getConsumerOfDependAllServices()
	if err != nil {
		util.Logger().Errorf(err, "Get consumer that depend on all services failed, %s", dr.provider.ServiceId)
		return nil, err
	}

	consumerDependList, err := dr.getConsumerOfSameServiceNameAndAppId(providerService)
	if err != nil {
		util.Logger().Errorf(err, "Get consumer that depend on same serviceName and appid rule failed, %s",
			dr.provider.ServiceId)
		return nil, err
	}
	consumerDependAllList = append(consumerDependAllList, consumerDependList...)
	return consumerDependAllList, nil
}

func (dr *DependencyRelation) getConsumerOfDependAllServices() ([]*pb.MicroServiceKey, error) {
	providerService := pb.MicroServiceToKey(dr.domainProject, dr.provider)
	providerService.ServiceName = "*"
	relyAllKey := apt.GenerateProviderDependencyRuleKey(dr.domainProject, providerService)
	opts := append(FromContext(dr.ctx), registry.WithStrKey(relyAllKey))
	rsp, err := store.Store().DependencyRule().Search(dr.ctx, opts...)
	if err != nil {
		util.Logger().Errorf(err, "get consumer that rely all service failed.")
		return nil, err
	}
	dependency := &pb.MicroServiceDependency{}
	if len(rsp.Kvs) != 0 {
		util.Logger().Infof("consumer that rely all service exist.ServiceName: %s.", dr.provider.ServiceName)
		err = json.Unmarshal(rsp.Kvs[0].Value, dependency)
		if err != nil {
			return nil, err
		}
		return dependency.Dependency, nil
	}
	return dependency.Dependency, nil
}

func (dr *DependencyRelation) getConsumerOfSameServiceNameAndAppId(provider *pb.MicroServiceKey) ([]*pb.MicroServiceKey, error) {
	providerVersion := provider.Version
	provider.Version = ""
	prefix := apt.GenerateProviderDependencyRuleKey(dr.domainProject, provider)
	provider.Version = providerVersion

	opts := append(FromContext(dr.ctx),
		registry.WithStrKey(prefix),
		registry.WithPrefix())
	rsp, err := store.Store().DependencyRule().Search(dr.ctx, opts...)
	if err != nil {
		util.Logger().Errorf(err, "get all dependency rule failed: provider rule key %v.", provider)
		return nil, err
	}

	allConsumers := make([]*pb.MicroServiceKey, 0, len(rsp.Kvs))
	var latestServiceId []string

	for _, kv := range rsp.Kvs {
		dependency := &pb.MicroServiceDependency{
			Dependency: []*pb.MicroServiceKey{},
		}
		providerVersionRuleArr := strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
		providerVersionRule := providerVersionRuleArr[len(providerVersionRuleArr)-1]
		if providerVersionRule == "latest" {
			if latestServiceId == nil {
				latestServiceId, err = FindServiceIds(dr.ctx, providerVersionRule, provider)
				if err != nil {
					util.Logger().Errorf(err, "Get latest service failed.")
					return nil, err
				}
			}
			if len(latestServiceId) == 0 {
				util.Logger().Infof("%s 's providerId is empty,no this service.", provider.ServiceName)
				continue
			}
			if dr.provider.ServiceId != latestServiceId[0] {
				continue
			}

		} else {
			if !VersionMatchRule(providerVersion, providerVersionRule) {
				continue
			}
		}

		util.Logger().Debugf("providerETCD is %s", providerVersionRuleArr)
		err = json.Unmarshal(kv.Value, dependency)
		if err != nil {
			util.Logger().Errorf(err, "Unmarshal consumers failed.")
			return nil, err
		}
		allConsumers = append(allConsumers, dependency.Dependency...)
	}
	return allConsumers, nil
}

func DeleteDependencyForDeleteService(domainProject string, serviceId string, service *pb.MicroServiceKey) (registry.PluginOp, error) {
	key := apt.GenerateConsumerDependencyQueueKey(domainProject, serviceId, "0")
	conDep := new(pb.ConsumerDependency)
	conDep.Consumer = service
	conDep.Providers = []*pb.MicroServiceKey{}
	conDep.Override = true
	data, err := json.Marshal(conDep)
	if err != nil {
		return registry.PluginOp{}, err
	}
	return registry.OpPut(registry.WithStrKey(key), registry.WithValue(data)), nil
}

type DependencyRelationFilterOpt struct {
	SameDomainProject bool
	NonSelf           bool
}

type DependencyRelationFilterOption func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt

func WithSameDomainProject() DependencyRelationFilterOption {
	return func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt {
		opt.SameDomainProject = true
		return opt
	}
}
func WithoutSelfDependency() DependencyRelationFilterOption {
	return func(opt DependencyRelationFilterOpt) DependencyRelationFilterOpt {
		opt.NonSelf = true
		return opt
	}
}

func toDependencyRelationFilterOpt(opts ...DependencyRelationFilterOption) (op DependencyRelationFilterOpt) {
	for _, opt := range opts {
		op = opt(op)
	}
	return
}
