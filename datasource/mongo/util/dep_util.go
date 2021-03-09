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
	"fmt"

	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/db"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetAllConsumerIds(ctx context.Context, provider *pb.MicroService) (allow []string, deny []string, _ error) {
	if provider == nil || len(provider.ServiceId) == 0 {
		return nil, nil, fmt.Errorf("invalid provider")
	}

	//todo 删除服务，最后实例推送有误差
	domain := util.ParseDomainProject(ctx)
	project := util.ParseProject(ctx)
	providerRules, err := GetRulesUtil(ctx, domain, project, provider.ServiceId)
	if err != nil {
		return nil, nil, err
	}

	allow, deny, err = GetConsumerIDsWithFilter(ctx, provider, providerRules)
	if err != nil {
		return nil, nil, err
	}
	return allow, deny, nil
}

func GetConsumerIDsWithFilter(ctx context.Context, provider *pb.MicroService, rules []*db.Rule) (allow []string, deny []string, err error) {
	domainProject := util.ParseDomainProject(ctx)
	dr := NewProviderDependencyRelation(ctx, domainProject, provider)
	consumerIDs, err := dr.GetDependencyConsumerIds()
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s consumerIds failed", provider.ServiceId), err)
		return nil, nil, err
	}
	return FilterAll(ctx, consumerIDs, rules)
}

func TransferToMicroServiceDependency(ctx context.Context, filter bson.M) (*pb.MicroServiceDependency, error) {
	microServiceDependency := &pb.MicroServiceDependency{
		Dependency: []*pb.MicroServiceKey{},
	}
	findRes, err := client.GetMongoClient().FindOne(context.TODO(), db.CollectionDep, filter)
	if err != nil {
		return nil, err
	}
	if findRes.Err() == nil {
		var depRule *db.DependencyRule
		err := findRes.Decode(&depRule)
		if err != nil {
			return nil, err
		}
		microServiceDependency.Dependency = append(microServiceDependency.Dependency, depRule.Dep.Dependency...)
		return microServiceDependency, nil
	}
	return microServiceDependency, nil
}

func SyncDependencyRule(ctx context.Context, domainProject string, r *pb.ConsumerDependency) error {

	consumerInfo := pb.DependenciesToKeys([]*pb.MicroServiceKey{r.Consumer}, domainProject)[0]
	providersInfo := pb.DependenciesToKeys(r.Providers, domainProject)

	var dep datasource.Dependency
	//var err error
	dep.DomainProject = domainProject
	dep.Consumer = consumerInfo
	dep.ProvidersRule = providersInfo
	// add mongo get dep here

	oldProviderRules, err := GetOldProviderRules(&dep)
	if err != nil {
		return err
	}

	if r.Override {
		datasource.ParseOverrideRules(ctx, &dep, oldProviderRules)
	} else {
		datasource.ParseAddOrUpdateRules(ctx, &dep, oldProviderRules)
	}
	return updateDeps(domainProject, &dep)
}

func updateDeps(domainProject string, dep *datasource.Dependency) error {
	var upsert = true
	for _, r := range dep.DeleteDependencyRuleList {
		filter := GenerateProviderDependencyRuleKey(domainProject, r)
		_, err := client.GetMongoClient().Update(context.TODO(), db.CollectionDep, filter, bson.M{"$pull": bson.M{StringBuilder([]string{db.ColumnDep, db.ColumnDependency}): dep.Consumer}})
		if err != nil {
			return err
		}
		if r.ServiceName == "*" {
			break
		}
	}
	for _, r := range dep.CreateDependencyRuleList {
		filter := GenerateProviderDependencyRuleKey(domainProject, r)
		data := bson.M{
			"$addToSet": bson.M{StringBuilder([]string{db.ColumnDep, db.ColumnDependency}): dep.Consumer},
		}
		_, err := client.GetMongoClient().Update(context.TODO(), db.CollectionDep, filter, data, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return err
		}
		if r.ServiceName == "*" {
			break
		}
	}
	filter := GenerateConsumerDependencyRuleKey(domainProject, dep.Consumer)
	if len(dep.ProvidersRule) == 0 {
		_, err := client.GetMongoClient().Delete(context.TODO(), db.CollectionDep, filter)
		if err != nil {
			return err
		}
	} else {
		updateData := bson.M{
			"$set": bson.M{StringBuilder([]string{db.ColumnDep, db.ColumnDependency}): dep.ProvidersRule},
		}
		_, err := client.GetMongoClient().Update(context.TODO(), db.CollectionDep, filter, updateData, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			return err
		}
	}

	err := CleanUpDepRules(context.TODO(), domainProject)
	if err != nil {
		return err
	}

	return nil
}

func CleanUpDepRules(ctx context.Context, domainProject string) error {
	if len(domainProject) == 0 {
		return ErrInvalidDomainProject
	}

	cache := make(map[*db.DelDepCacheKey]bool)
	err := removeProviderRuleOfConsumer(ctx, domainProject, cache)

	if err != nil {
		return err
	}

	return removeProviderRuleKeys(ctx, domainProject, cache)
}
func removeProviderRuleOfConsumer(ctx context.Context, domainProject string, cache map[*db.DelDepCacheKey]bool) error {
	filter := GenerateConsumerDependencyRuleKey(domainProject, nil)
	depRules, err := GetDepRules(ctx, filter)
	if err != nil {
		return err
	}
	for _, depRule := range depRules {
		err := removeConsumerDeps(ctx, depRule, cache)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeProviderRuleKeys(ctx context.Context, domainProject string, cache map[*db.DelDepCacheKey]bool) error {
	filter := GenerateProviderDependencyRuleKey(domainProject, nil)
	depRules, err := GetDepRules(ctx, filter)
	if err != nil {
		return err
	}
	for _, depRule := range depRules {
		err := removeProviderDeps(ctx, depRule, cache)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDepRules(ctx context.Context, filter bson.M) ([]*db.DependencyRule, error) {
	findRes, err := client.GetMongoClient().Find(ctx, db.CollectionDep, filter)
	if err != nil {
		return nil, err
	}

	var depRules []*db.DependencyRule
	for findRes.Next(ctx) {
		var depRule *db.DependencyRule
		err := findRes.Decode(&depRule)
		if err != nil {
			return nil, err
		}
		depRules = append(depRules, depRule)
	}
	return depRules, nil
}

func removeProviderDeps(ctx context.Context, depRule *db.DependencyRule, cache map[*db.DelDepCacheKey]bool) (err error) {
	id := &db.DelDepCacheKey{
		Key:  depRule.ServiceKey,
		Type: path.DepsConsumer,
	}
	exist, ok := cache[id]
	if !ok {
		if depRule.ServiceKey.ServiceName == "*" {
			return nil
		}

		_, exist, err = FindServiceIds(ctx, depRule.ServiceKey.Version, depRule.ServiceKey)
		if err != nil {
			return err
		}

		cache[id] = exist
	}

	filter := bson.M{
		db.ColumnServiceKey: depRule.ServiceKey,
	}
	if !exist {
		_, err = client.GetMongoClient().DocDelete(ctx, db.CollectionDep, filter)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeConsumerDeps(ctx context.Context, depRule *db.DependencyRule, cache map[*db.DelDepCacheKey]bool) (err error) {
	var left []*pb.MicroServiceKey
	for _, key := range depRule.Dep.Dependency {
		if key.ServiceName == "*" {
			left = append(left, key)
			continue
		}

		id := &db.DelDepCacheKey{
			Key:  key,
			Type: path.DepsProvider,
		}
		exist, ok := cache[id]
		if !ok {
			_, exist, err = FindServiceIds(ctx, key.Version, key)
			if err != nil {
				return err
			}
			cache[id] = exist
		}

		if exist {
			left = append(left, key)
		}
	}
	if len(depRule.Dep.Dependency) == len(left) {
		return nil
	}

	filter := bson.M{
		db.ColumnServiceKey: depRule.ServiceKey,
	}
	if len(left) == 0 {
		_, err = client.GetMongoClient().DocDelete(ctx, db.CollectionDep, filter)
	} else {
		updateData := bson.M{
			"$set": bson.M{StringBuilder([]string{db.ColumnDep, db.ColumnDependency}): left},
		}
		_, err = client.GetMongoClient().Update(ctx, db.CollectionDep, filter, updateData)
	}
	if err != nil {
		return err
	}
	return nil
}

func DeleteDependencyForDeleteService(domainProject string, serviceID string, service *pb.MicroServiceKey) error {
	conDep := new(pb.ConsumerDependency)
	conDep.Consumer = service
	conDep.Providers = []*pb.MicroServiceKey{}
	conDep.Override = true
	err := SyncDependencyRule(context.TODO(), domainProject, conDep)
	if err != nil {
		return err
	}
	return nil
}

func DependencyRuleExistUtil(ctx context.Context, key bson.M, target *pb.MicroServiceKey) (bool, error) {
	compareData, err := TransferToMicroServiceDependency(ctx, key)
	if err != nil {
		return false, err
	}

	if len(compareData.Dependency) != 0 {
		isEqual, err := datasource.ContainServiceDependency(compareData.Dependency, target)
		if err != nil {
			return false, err
		}
		if isEqual {
			return true, nil
		}
	}
	return false, nil
}

func DependencyRuleExist(ctx context.Context, provider *pb.MicroServiceKey, consumer *pb.MicroServiceKey) (bool, error) {
	targetDomainProject := provider.Tenant
	if len(targetDomainProject) == 0 {
		targetDomainProject = consumer.Tenant
	}
	consumerKey := GenerateConsumerDependencyRuleKey(consumer.Tenant, consumer)
	existed, err := DependencyRuleExistUtil(ctx, consumerKey, provider)
	if err != nil || existed {
		return existed, err
	}
	providerKey := GenerateProviderDependencyRuleKey(targetDomainProject, provider)
	return DependencyRuleExistUtil(ctx, providerKey, consumer)
}

func AddServiceVersionRule(ctx context.Context, domainProject string, consumer *pb.MicroService, provider *pb.MicroServiceKey) error {
	consumerKey := pb.MicroServiceToKey(domainProject, consumer)
	exist, err := DependencyRuleExist(ctx, provider, consumerKey)
	if exist || err != nil {
		return err
	}

	r := &pb.ConsumerDependency{
		Consumer:  consumerKey,
		Providers: []*pb.MicroServiceKey{provider},
		Override:  false,
	}
	err = SyncDependencyRule(ctx, domainProject, r)

	if err != nil {
		return err
	}

	return nil
}
