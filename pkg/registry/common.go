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

package registry

import (
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

const (
	ExistenceMicroservice string = "microservice"
	ExistenceSchema       string = "schema"

	PropAllowCrossApp = "allowCrossApp"

	ResponseSuccess int32 = 0

	AppID   = "default"
	VERSION = "0.0.1"
)

func CreateResponse(code int32, message string) *Response {
	resp := &Response{
		Code:    code,
		Message: message,
	}
	return resp
}

func CreateResponseWithSCErr(err *scerr.Error) *Response {
	return &Response{
		Code:    err.Code,
		Message: err.Detail,
	}
}

func DependenciesToKeys(in []*MicroServiceKey, domainProject string) []*MicroServiceKey {
	for _, value := range in {
		if len(value.Tenant) == 0 {
			value.Tenant = domainProject
		}
	}
	return in
}

func MicroServiceToKey(domainProject string, in *MicroService) *MicroServiceKey {
	return &MicroServiceKey{
		Tenant:      domainProject,
		Environment: in.Environment,
		AppId:       in.AppId,
		ServiceName: in.ServiceName,
		Alias:       in.Alias,
		Version:     in.Version,
	}
}
