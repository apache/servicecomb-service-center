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

package service

import (
	"context"

	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func findServicesMapByDomainProject(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (map[string][]*pb.MicroService, error) {
	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter, opts...)
	if err != nil {
		log.Error("failed to find services", err)
		return nil, err
	}

	services := make(map[string][]*pb.MicroService)

	for findRes.Next(ctx) {
		var mongoService model.Service
		err := findRes.Decode(&mongoService)
		if err != nil {
			return nil, err
		}
		domainProject := mongoService.Domain + "/" + mongoService.Project
		services[domainProject] = append(services[domainProject], mongoService.Service)
	}
	return services, nil
}

func findService(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*model.Service, error) {
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	var svc *model.Service
	if result.Err() != nil {
		return nil, datasource.ErrNoData
	}
	err = result.Decode(&svc)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func findServices(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*model.Service, error) {
	result, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	var services []*model.Service
	for result.Next(ctx) {
		var tmp model.Service
		err := result.Decode(&tmp)
		if err != nil {
			log.Error("failed to decode service", err)
			return nil, err
		}
		services = append(services, &tmp)
	}
	return services, nil
}

func insertService(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) error {
	_, err := client.GetMongoClient().Insert(ctx, model.CollectionService, document, opts...)
	return err
}

func findMicroService(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*pb.MicroService, error) {
	result, err := client.GetMongoClient().FindOne(ctx, model.CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, datasource.ErrNoData
	}
	var service *model.Service
	err = result.Decode(&service)
	if err != nil {
		return nil, err
	}
	return service.Service, nil
}

func findMicroServices(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]*pb.MicroService, error) {
	res, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	var services []*pb.MicroService
	for res.Next(ctx) {
		var tmp model.Service
		err := res.Decode(&tmp)
		if err != nil {
			return nil, err
		}
		services = append(services, tmp.Service)
	}
	return services, nil
}

func exitService(ctx context.Context, filter interface{}) (bool, error) {
	return client.GetMongoClient().DocExist(ctx, model.CollectionService, filter)
}

func findServicesIDs(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (serviceIDs []string, err error) {
	result, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	for result.Next(ctx) {
		var tmp *model.Service
		err = result.Decode(&tmp)
		serviceIDs = append(serviceIDs, tmp.Service.ServiceId)
	}
	return
}

func findLatestServiceID(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (serviceIDs []string, err error) {
	result, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter, opts...)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	for result.Next(ctx) {
		var tmp *model.Service
		err = result.Decode(&tmp)
		serviceIDs = append(serviceIDs, tmp.Service.ServiceId)
		if serviceIDs != nil {
			return
		}
	}
	return
}

func updateService(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) error {
	res, err := client.GetMongoClient().FindOneAndUpdate(ctx, model.CollectionService, filter, update, opts...)
	if err != nil {
		return err
	}
	if res.Err() != nil {
		// means no doc find, if the operation is update,should return err
		return client.ErrNoDocuments
	}
	return nil
}

func getServicesVersions(ctx context.Context, filter interface{}) ([]string, error) {
	res, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter)
	if err != nil {
		return nil, nil
	}
	var versions []string
	for res.Next(ctx) {
		var tmp *model.Service
		err := res.Decode(&tmp)
		if err != nil {
			log.Error("failed to decode service", err)
			return nil, err
		}
		versions = append(versions, tmp.Service.Version)
	}
	return versions, nil
}

func batchUpdateServices(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) error {
	_, err := client.GetMongoClient().BatchUpdate(ctx, model.CollectionService, models, opts...)
	return err
}
