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
	"encoding/json"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
)

func (s *MicroServiceService) AddTags(ctx context.Context, in *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "add service[%s]'s tags %v failed, operator: %s", in.ServiceId, in.Tags, remoteIP)
		return &pb.AddServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)
	// service id存在性校验
	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(nil, "add service[%s]'s tags %v failed, service does not exist, operator: %s",
			in.ServiceId, in.Tags, remoteIP)
		return &pb.AddServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	checkErr := serviceUtil.AddTagIntoETCD(ctx, domainProject, in.ServiceId, in.Tags)
	if checkErr != nil {
		log.Errorf(checkErr, "add service[%s]'s tags %v failed, operator: %s", in.ServiceId, in.Tags, remoteIP)
		resp := &pb.AddServiceTagsResponse{
			Response: proto.CreateResponseWithSCErr(checkErr),
		}
		if checkErr.InternalError() {
			return resp, checkErr
		}
		return resp, nil
	}

	log.Infof("add service[%s]'s tags %v successfully, operator: %s", in.ServiceId, in.Tags, remoteIP)
	return &pb.AddServiceTagsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Add service tags successfully."),
	}, nil
}

func (s *MicroServiceService) UpdateTag(ctx context.Context, in *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	tagFlag := util.StringJoin([]string{in.Key, in.Value}, "/")
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "update service[%s]'s tag[%s] failed, operator: %s", in.ServiceId, tagFlag, remoteIP)
		return &pb.UpdateServiceTagResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(err, "update service[%s]'s tag[%s] failed, service does not exist, operator: %s",
			in.ServiceId, tagFlag, remoteIP)
		return &pb.UpdateServiceTagResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "update service[%s]'s tag[%s] failed, get tag failed, operator: %s",
			in.ServiceId, tagFlag, remoteIP)
		return &pb.UpdateServiceTagResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}
	//check tag 是否存在
	if _, ok := tags[in.Key]; !ok {
		log.Errorf(nil, "update service[%s]'s tag[%s] failed, tag does not exist, operator: %s",
			in.ServiceId, tagFlag, remoteIP)
		return &pb.UpdateServiceTagResponse{
			Response: proto.CreateResponse(scerr.ErrTagNotExists, "Tag does not exist, please add one first."),
		}, nil
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	copyTags[in.Key] = in.Value

	checkErr := serviceUtil.AddTagIntoETCD(ctx, domainProject, in.ServiceId, copyTags)
	if checkErr != nil {
		log.Errorf(checkErr, "update service[%s]'s tag[%s] failed, operator: %s", in.ServiceId, tagFlag, remoteIP)
		resp := &pb.UpdateServiceTagResponse{
			Response: proto.CreateResponseWithSCErr(checkErr),
		}
		if checkErr.InternalError() {
			return resp, checkErr
		}
		return resp, nil
	}

	log.Infof("update service[%s]'s tag[%s] successfully, operator: %s", in.ServiceId, tagFlag, remoteIP)
	return &pb.UpdateServiceTagResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Update service tag success."),
	}, nil
}

func (s *MicroServiceService) DeleteTags(ctx context.Context, in *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "delete service[%s]'s tags %v failed, operator: %s", in.ServiceId, in.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(nil, "delete service[%s]'s tags %v failed, service does not exist, operator: %s",
			in.ServiceId, in.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "delete service[%s]'s tags %v failed, get service tags failed, operator: %s",
			in.ServiceId, in.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	copyTags := make(map[string]string, len(tags))
	for k, v := range tags {
		copyTags[k] = v
	}
	for _, key := range in.Keys {
		if _, ok := copyTags[key]; !ok {
			log.Errorf(nil, "delete service[%s]'s tags %v failed, tag[%s] does not exist, operator: %s",
				in.ServiceId, in.Keys, key, remoteIP)
			return &pb.DeleteServiceTagsResponse{
				Response: proto.CreateResponse(scerr.ErrTagNotExists, "Delete tags failed for this key "+key+" does not exist."),
			}, nil
		}
		delete(copyTags, key)
	}

	// tags 可能size == 0
	data, err := json.Marshal(copyTags)
	if err != nil {
		log.Errorf(err, "delete service[%s]'s tags %v failed, marshall service tags failed, operator: %s",
			in.ServiceId, in.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	key := apt.GenerateServiceTagKey(domainProject, in.ServiceId)

	resp, err := backend.Registry().TxnWithCmp(ctx,
		[]registry.PluginOp{registry.OpPut(registry.WithStrKey(key), registry.WithValue(data))},
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, in.ServiceId))),
			registry.CmpNotEqual, 0)},
		nil)
	if err != nil {
		log.Errorf(err, "delete service[%s]'s tags %v failed, operator: %s",
			in.ServiceId, in.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}
	if !resp.Succeeded {
		log.Errorf(err, "delete service[%s]'s tags %v failed, service does not exist, operator: %s",
			in.ServiceId, in.Keys, remoteIP)
		return &pb.DeleteServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	log.Infof("delete service[%s]'s tags %v successfully, operator: %s", in.ServiceId, in.Keys, remoteIP)
	return &pb.DeleteServiceTagsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Delete service tags successfully."),
	}, nil
}

func (s *MicroServiceService) GetTags(ctx context.Context, in *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "get service[%s]'s tags failed", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	domainProject := util.ParseDomainProject(ctx)

	if !serviceUtil.ServiceExist(ctx, domainProject, in.ServiceId) {
		log.Errorf(err, "get service[%s]'s tags failed, service does not exist", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrServiceNotExists, "Service does not exist."),
		}, nil
	}

	tags, err := serviceUtil.GetTagsUtils(ctx, domainProject, in.ServiceId)
	if err != nil {
		log.Errorf(err, "get service[%s]'s tags failed, get tags failed", in.ServiceId)
		return &pb.GetServiceTagsResponse{
			Response: proto.CreateResponse(scerr.ErrInternal, err.Error()),
		}, err
	}

	return &pb.GetServiceTagsResponse{
		Response: proto.CreateResponse(proto.Response_SUCCESS, "Get service tags successfully."),
		Tags:     tags,
	}, nil
}
