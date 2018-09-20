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
	"encoding/json"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/quota"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
)

func (s *MicroServiceService) AddTags(ctx context.Context, in *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "add service tags failed, serviceId %s, tags %v: invalid parameters.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)
	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(nil, "add service tags failed, serviceId %s, tags %v: service not exist.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	addTags := in.GetTags()
	res := quota.NewApplyQuotaResource(quota.TagQuotaType, domainProject, in.ServiceId, int64(len(addTags)))
	rst := plugin.Plugins().Quota().Apply4Quotas(ctx, res)
	errQuota := rst.Err
	if errQuota != nil {
		log.Errorf(errQuota, "add tag info failed, check resource num failed, %s", in.ServiceId)
		response := &pb.AddServiceTagsResponse{
			Response: pb.CreateResponseWithSCErr(errQuota),
		}
		if errQuota.InternalError() {
			return response, errQuota
		}
		return response, nil
	}

	dataTags, err := serviceUtil.GetTagsUtils(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "add service tags failed, serviceId %s, tags %v: get existed tag failed.", in.ServiceId, in.Tags)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	for key, value := range dataTags {
		if _, ok := addTags[key]; ok {
			continue
		}
		addTags[key] = value
	}
	dataTags = addTags

	checkErr := serviceUtil.AddTagIntoETCD(ctx, domainProject, in.ServiceId, dataTags)
	if checkErr != nil {
		log.Errorf(checkErr, "add service tags failed, serviceId %s, tags %v", in.ServiceId, in.Tags)
		resp := &pb.AddServiceTagsResponse{
			Response: pb.CreateResponseWithSCErr(checkErr),
		}
		if checkErr.InternalError() {
			return resp, checkErr
		}
		return resp, nil
	}

	log.Infof("add service tags successful, serviceId %s, tags %v.", in.ServiceId, in.Tags)
	return &pb.AddServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Add service tags successfully."),
	}, nil
}

func (s *MicroServiceService) UpdateTag(ctx context.Context, in *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	tagFlag := util.StringJoin([]string{in.Key, in.Value}, "/")
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "update service tag failed, serviceId %s, tag %s: invalid params.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(err, "update service tag failed, serviceId %s, tag %s: service not exist.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "update service tag failed, serviceId %s, tag %s: get tag failed.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	//check tag 是否存在
	if _, ok := tags[in.Key]; !ok {
		log.Errorf(nil, "update service tag failed, serviceId %s, tag %s: tag not exist, please add first.", in.ServiceId, tagFlag)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(scerr.ErrTagNotExists, "Update tag for service failed for update tags not exist, please add first."),
		}, nil
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	copyTags[in.Key] = in.Value

	checkErr := serviceUtil.AddTagIntoETCD(ctx, domainProject, in.ServiceId, copyTags)
	if checkErr != nil {
		log.Errorf(checkErr, "update service tag failed, serviceId %s, tag %s.", in.ServiceId, tagFlag)
		resp := &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponseWithSCErr(checkErr),
		}
		if checkErr.InternalError() {
			return resp, checkErr
		}
		return resp, nil
	}

	log.Infof("update tag successful, serviceId %s, tag %s.", in.ServiceId, tagFlag)
	return &pb.UpdateServiceTagResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Update service tag success."),
	}, nil
}

func (s *MicroServiceService) DeleteTags(ctx context.Context, in *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "delete service tags failed, serviceId %s, tags %v: invalid params.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(nil, "delete service tags failed, serviceId %s, tags %v: service not exist.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "delete service tags failed, serviceId %s, tags %v: query service failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	for _, key := range in.Keys {
		if _, ok := copyTags[key]; !ok {
			log.Errorf(nil, "delete service tags failed, serviceId %s, tags %v: tag %s not exist.", in.ServiceId, in.Keys, key)
			return &pb.DeleteServiceTagsResponse{
				Response: pb.CreateResponse(scerr.ErrTagNotExists, "Delete tags failed for this key "+key+" does not exist."),
			}, nil
		}
		delete(copyTags, key)
	}

	// tags 可能size == 0
	data, err := json.Marshal(copyTags)
	if err != nil {
		log.Errorf(err, "delete service tags failed, serviceId %s, tags %v: marshall service tag failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	key := apt.GenerateServiceTagKey(domainProject, in.ServiceId)

	resp, err := backend.Registry().TxnWithCmp(ctx,
		[]registry.PluginOp{registry.OpPut(registry.WithStrKey(key), registry.WithValue(data))},
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, in.ServiceId))),
			registry.CMP_NOT_EQUAL, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "delete service tags failed, serviceId %s, tags %v: commit tag data into etcd failed.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "delete service properties failed, serviceId is %s, tags %v: service does not exist.", in.ServiceId, in.Keys)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("delete service tags successful: serviceId %s, tag %v.", in.ServiceId, in.Keys)
	return &pb.DeleteServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Delete service tags successfully."),
	}, nil
}

func (s *MicroServiceService) GetTags(ctx context.Context, in *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "get service tags failed, serviceId %s: invalid parameters.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(err, "get service tags failed, serviceId %s: service not exist.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "get service tags failed, serviceId %s: get tag failed.", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServiceTagsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Get service tags successfully."),
		Tags:     tags,
	}, nil
}
