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

package dao

import (
	"context"
	"errors"
	"strings"

	"github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/dao/util"
)

func GetServiceByID(ctx context.Context, serviceID string) (*Service, error) {
	return GetService(ctx, mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(serviceID)))
}

func GetService(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*Service, error) {
	result, err := client.GetMongoClient().FindOne(ctx, CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	var svc *Service
	if result.Err() != nil {
		return nil, datasource.ErrNoData
	}
	err = result.Decode(&svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func GetServiceID(ctx context.Context, key *discovery.MicroServiceKey) (string, error) {
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

func getServiceIDs(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (serviceIDs []string, err error) {
	result, err := client.GetMongoClient().Find(ctx, CollectionService, filter, opts...)
	if err != nil {
		return
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	for result.Next(ctx) {
		var service *Service
		err = result.Decode(&service)
		if err != nil {
			return
		}
		serviceIDs = append(serviceIDs, service.Service.ServiceId)
	}
	return
}

func getServiceID(ctx context.Context, filter bson.M) (serviceID string, err error) {
	svc, err := GetService(ctx, filter)
	if err != nil {
		return
	}
	if svc != nil {
		serviceID = svc.Service.ServiceId
		return
	}
	return
}

func GetServices(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*Service, error) {
	res, err := client.GetMongoClient().Find(ctx, CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	var services []*Service
	for res.Next(ctx) {
		var tmp *Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		services = append(services, tmp)
	}
	return services, nil
}

func GetMicroServices(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*discovery.MicroService, error) {
	res, err := client.GetMongoClient().Find(ctx, CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	var services []*discovery.MicroService
	for res.Next(ctx) {
		var tmp Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		services = append(services, tmp.Service)
	}
	return services, nil
}

func UpdateService(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) error {
	res, err := client.GetMongoClient().FindOneAndUpdate(ctx, CollectionService, filter, update, opts...)
	if err != nil {
		return err
	}
	if res.Err() != nil {
		// means no doc find, if the operation is update,should return err
		return mutil.ErrNoDocuments
	}
	return nil
}

func GetServicesVersions(ctx context.Context, filter interface{}) ([]string, error) {
	res, err := client.GetMongoClient().Find(ctx, CollectionService, filter)
	if err != nil {
		return nil, nil
	}
	var versions []string
	for res.Next(ctx) {
		var tmp Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		versions = append(versions, tmp.Service.Version)
	}
	return versions, nil
}

func getServiceIDsByVersion(ctx context.Context, filter bson.D, versionRule string) ([]string, error) {
	if len(versionRule) == 0 {
		return nil, nil
	}
	rangeIdx := strings.Index(versionRule, "-")
	switch {
	case versionRule == "latest":
		findOption := options.FindOptions{Sort: bson.M{mutil.ConnectWithDot([]string{ColumnService, ColumnVersion}): -1}}
		return getServiceIDs(ctx, filter, &findOption)
	case len(versionRule) > 0 && versionRule[len(versionRule)-1:] == "+":
		start := versionRule[:len(versionRule)-1]
		filter = append(filter, bson.E{Key: mutil.ConnectWithDot([]string{ColumnService, ColumnVersion}), Value: bson.M{"$gte": start}})
		return getServiceIDs(ctx, filter)
	case rangeIdx > 0:
		start := versionRule[:rangeIdx]
		end := versionRule[rangeIdx+1:]
		filter = append(filter, bson.E{Key: mutil.ConnectWithDot([]string{ColumnService, ColumnVersion}), Value: bson.M{"$gte": start, "$lte": end}})
		return getServiceIDs(ctx, filter)
	default:
		filter = append(filter, bson.E{Key: mutil.ConnectWithDot([]string{ColumnService, ColumnVersion}), Value: versionRule})
		return getServiceIDs(ctx, filter)
	}
}

func GetServiceIDs(ctx context.Context, versionRule string, key *discovery.MicroServiceKey) (serviceIDs []string, err error) {
	if len(versionRule) == 0 {
		return nil, nil
	}

	tenant := strings.Split(key.Tenant, "/")
	if len(tenant) != 2 {
		return nil, mutil.ErrInvalidDomainProject
	}
	baseFilter := bson.D{
		{Key: ColumnDomain, Value: tenant[0]},
		{Key: ColumnProject, Value: tenant[1]},
		{Key: mutil.ConnectWithDot([]string{ColumnService, ColumnEnv}), Value: key.Environment},
		{Key: mutil.ConnectWithDot([]string{ColumnService, ColumnAppID}), Value: key.AppId}}
	serviceIDs, err = getServiceIDsByVersion(ctx, baseFilter, versionRule)
	if err != nil {
		return nil, err
	}
	if len(serviceIDs) == 0 {
		if len(key.Alias) == 0 {
			return nil, nil
		}
		filter := append(baseFilter, bson.E{Key: mutil.ConnectWithDot([]string{ColumnService, ColumnAlias}), Value: key.Alias})
		serviceIDs, err = getServiceIDsByVersion(ctx, filter, versionRule)
		if err != nil {
			return nil, err
		}
	}
	return serviceIDs, nil
}
