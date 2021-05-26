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
	"github.com/apache/servicecomb-service-center/server/service/validator"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/go-chassis/cari/discovery"
)

func (s *MicroServiceService) AddRule(ctx context.Context, in *pb.AddServiceRulesRequest) (*pb.AddServiceRulesResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Errorf(err, "add service[%s] rule failed, operator: %s", in.ServiceId, remoteIP)
		return &pb.AddServiceRulesResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().AddRule(ctx, in)
}

func (s *MicroServiceService) UpdateRule(ctx context.Context, in *pb.UpdateServiceRuleRequest) (*pb.UpdateServiceRuleResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Errorf(err, "update service rule[%s/%s] failed, operator: %s", in.ServiceId, in.RuleId, remoteIP)
		return &pb.UpdateServiceRuleResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().UpdateRule(ctx, in)
}

func (s *MicroServiceService) GetRule(ctx context.Context, in *pb.GetServiceRulesRequest) (*pb.GetServiceRulesResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		log.Errorf(err, "get service[%s] rule failed", in.ServiceId)
		return &pb.GetServiceRulesResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().GetRules(ctx, in)
}

func (s *MicroServiceService) DeleteRule(ctx context.Context, in *pb.DeleteServiceRulesRequest) (*pb.DeleteServiceRulesResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Errorf(err, "delete service[%s] rules %v failed, operator: %s", in.ServiceId, in.RuleIds, remoteIP)
		return &pb.DeleteServiceRulesResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().DeleteRule(ctx, in)
}
