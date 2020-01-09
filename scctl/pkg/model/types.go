// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/admin/model"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"strconv"
	"time"
)

func GetDomainProject(resouce interface{}) (domainProject string) {
	switch resouce.(type) {
	case *model.Microservice:
		_, domainProject = core.GetInfoFromSvcKV(
			util.StringToBytesWithNoCopy(resouce.(*model.Microservice).Key))
	case *model.Instance:
		_, _, domainProject = core.GetInfoFromInstKV(
			util.StringToBytesWithNoCopy(resouce.(*model.Instance).Key))
	}
	return
}

type Service struct {
	DomainProject string
	Environment   string
	AppId         string
	ServiceName   string
	Versions      []string
	Frameworks    []*proto.FrameWorkProperty
	Endpoints     []string
	Timestamp     int64 // the seconds from 0 to now
}

func (s *Service) AppendVersion(v string) {
	s.Versions = append(s.Versions, v)
}

func (s *Service) AppendFramework(property *proto.FrameWorkProperty) {
	if property == nil || property.Name == "" {
		return
	}
	for _, fw := range s.Frameworks {
		if fw.Name == property.Name && fw.Version == property.Version {
			return
		}
	}
	s.Frameworks = append(s.Frameworks, property)
}

func (s *Service) AppendEndpoints(endpoints []string) {
	s.Endpoints = append(s.Endpoints, endpoints...)
}

func (s *Service) UpdateTimestamp(t string) {
	d, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return
	}
	if s.Timestamp == 0 || s.Timestamp > d {
		s.Timestamp = d
	}
}

func (s *Service) Age() time.Duration {
	return time.Now().Sub(time.Unix(s.Timestamp, 0).Local())
}

type Instance struct {
	DomainProject string
	Host          string
	Endpoints     []string
	Environment   string
	AppId         string
	ServiceName   string
	Version       string
	Framework     *proto.FrameWorkProperty
	Lease         int64 // seconds
	Timestamp     int64 // the seconds from 0 to now
}

func (s *Instance) SetLease(hc *proto.HealthCheck) {
	if hc == nil {
		s.Lease = -1
		return
	}
	if hc.Mode == proto.CHECK_BY_PLATFORM {
		s.Lease = 0
		return
	}
	s.Lease = int64(hc.Interval * (hc.Times + 1))
	return
}

func (s *Instance) UpdateTimestamp(t string) {
	d, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return
	}
	if s.Timestamp == 0 || s.Timestamp > d {
		s.Timestamp = d
	}
}

func (s *Instance) Age() time.Duration {
	return time.Now().Sub(time.Unix(s.Timestamp, 0).Local())
}
