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
	"errors"
	"fmt"

	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func (ds *DataSource) AddTags(ctx context.Context, request *discovery.AddServiceTagsRequest) (*discovery.AddServiceTagsResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	setValue := mutil.NewFilter(
		mutil.Tags(request.Tags),
	)
	updateFilter := mutil.NewFilter(
		mutil.Set(setValue),
	)
	err := updateService(ctx, filter, updateFilter)
	if err == nil {
		return &discovery.AddServiceTagsResponse{
			Response: discovery.CreateResponse(discovery.ResponseSuccess, "add service tags successfully."),
		}, nil
	}
	log.Error(fmt.Sprintf("update service %s tags failed.", request.ServiceId), err)
	if err == client.ErrNoDocuments {
		return &discovery.AddServiceTagsResponse{
			Response: discovery.CreateResponse(discovery.ErrServiceNotExists, err.Error()),
		}, nil
	}
	return &discovery.AddServiceTagsResponse{
		Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
	}, nil
}

func (ds *DataSource) GetTags(ctx context.Context, request *discovery.GetServiceTagsRequest) (*discovery.GetServiceTagsResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	svc, err := findService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
			return &discovery.GetServiceTagsResponse{
				Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "service does not exist"),
			}, nil
		}
		log.Error(fmt.Sprintf("failed to get service %s tags", request.ServiceId), err)
		return &discovery.GetServiceTagsResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, nil
	}
	return &discovery.GetServiceTagsResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "get service tags successfully."),
		Tags:     svc.Tags,
	}, nil
}

func (ds *DataSource) UpdateTag(ctx context.Context, request *discovery.UpdateServiceTagRequest) (*discovery.UpdateServiceTagResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	svc, err := findService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
			return &discovery.UpdateServiceTagResponse{
				Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "service does not exist"),
			}, nil
		}
		log.Error(fmt.Sprintf("failed to get %s tags", request.ServiceId), err)
		return &discovery.UpdateServiceTagResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, nil
	}
	dataTags := svc.Tags
	if len(dataTags) > 0 {
		if _, ok := dataTags[request.Key]; !ok {
			return &discovery.UpdateServiceTagResponse{
				Response: discovery.CreateResponse(discovery.ErrTagNotExists, "tag does not exist"),
			}, nil
		}
	}
	newTags := make(map[string]string, len(dataTags))
	for k, v := range dataTags {
		newTags[k] = v
	}
	newTags[request.Key] = request.Value
	setValue := mutil.NewFilter(
		mutil.Tags(newTags),
	)
	updateFilter := mutil.NewFilter(
		mutil.Set(setValue),
	)
	err = updateService(ctx, filter, updateFilter)
	if err != nil {
		log.Error(fmt.Sprintf("update service %s tags failed", request.ServiceId), err)
		return &discovery.UpdateServiceTagResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, nil
	}
	return &discovery.UpdateServiceTagResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "update service tag success."),
	}, nil
}

func (ds *DataSource) DeleteTags(ctx context.Context, request *discovery.DeleteServiceTagsRequest) (*discovery.DeleteServiceTagsResponse, error) {
	filter := mutil.NewBasicFilter(ctx, mutil.ServiceServiceID(request.ServiceId))
	svc, err := findService(ctx, filter)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			log.Debug(fmt.Sprintf("service %s not exist in db", request.ServiceId))
			return &discovery.DeleteServiceTagsResponse{
				Response: discovery.CreateResponse(discovery.ErrServiceNotExists, "service does not exist"),
			}, nil
		}
		log.Error(fmt.Sprintf("failed to get service %s tags", request.ServiceId), err)
		return &discovery.DeleteServiceTagsResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, nil
	}
	dataTags := svc.Tags
	newTags := make(map[string]string, len(dataTags))
	for k, v := range dataTags {
		newTags[k] = v
	}
	if len(dataTags) > 0 {
		for _, key := range request.Keys {
			if _, ok := dataTags[key]; !ok {
				return &discovery.DeleteServiceTagsResponse{
					Response: discovery.CreateResponse(discovery.ErrTagNotExists, "tag does not exist"),
				}, nil
			}
			delete(newTags, key)
		}
	}
	setValue := mutil.NewFilter(
		mutil.Tags(newTags),
	)
	updateFilter := mutil.NewFilter(
		mutil.Set(setValue),
	)
	err = updateService(ctx, filter, updateFilter)
	if err != nil {
		log.Error(fmt.Sprintf("delete service %s tags failed", request.ServiceId), err)
		return &discovery.DeleteServiceTagsResponse{
			Response: discovery.CreateResponse(discovery.ErrInternal, err.Error()),
		}, nil
	}
	return &discovery.DeleteServiceTagsResponse{
		Response: discovery.CreateResponse(discovery.ResponseSuccess, "update service tag success."),
	}, nil
}
